package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type OkxResponse struct {
	Code string          `json:"code"`
	Data []OkxInstrument `json:"data"`
	Msg  string          `json:"msg"`
}

type OkxInstrument struct {
	Alias            string `json:"alias"`
	AuctionEndTime   string `json:"auctionEndTime"`
	BaseCcy          string `json:"baseCcy"`
	Category         string `json:"category"`
	ContTdSwTime     string `json:"contTdSwTime"`
	CtMult           string `json:"ctMult"`
	CtType           string `json:"ctType"`
	CtVal            string `json:"ctVal"`
	CtValCcy         string `json:"ctValCcy"`
	ExpTime          string `json:"expTime"`
	FutureSettlement bool   `json:"futureSettlement"`
	InstFamily       string `json:"instFamily"`
	InstId           string `json:"instId"`
	InstType         string `json:"instType"`
	Lever            string `json:"lever"`
	ListTime         string `json:"listTime"`
	LotSz            string `json:"lotSz"`
	MaxIcebergSz     string `json:"maxIcebergSz"`
	MaxLmtAmt        string `json:"maxLmtAmt"`
	MaxLmtSz         string `json:"maxLmtSz"`
	MaxMktAmt        string `json:"maxMktAmt"`
	MaxMktSz         string `json:"maxMktSz"`
	MaxStopSz        string `json:"maxStopSz"`
	MaxTriggerSz     string `json:"maxTriggerSz"`
	MaxTwapSz        string `json:"maxTwapSz"`
	MinSz            string `json:"minSz"`
	OpenType         string `json:"openType"`
	OptType          string `json:"optType"`
	QuoteCcy         string `json:"quoteCcy"`
	RuleType         string `json:"ruleType"`
	SettleCcy        string `json:"settleCcy"`
	State            string `json:"state"`
	Stk              string `json:"stk"`
	TickSz           string `json:"tickSz"`
	Uly              string `json:"uly"`
}

type TickerResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		InstId string `json:"instId"`
		Last   string `json:"last"`
	} `json:"data"`
}

func cmd_crypto(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		c.SendMsg(msg, "用法:\n1.crypto <币种> - 查询币种对USDT价格\n2.crypto <源币种> <目标币种> - 查询币种对目标货币价格\n例如: crypto BTC 或 crypto BTC USD")
		return
	}

	if args.Size == 2 {
		coin := strings.ToUpper(args.Contents[1])
		handleSingleCrypto(c, msg, coin)
		return
	}

	if args.Size == 3 {
		fromCoin := strings.ToUpper(args.Contents[1])
		toCurrency := strings.ToUpper(args.Contents[2])
		handleCryptoCurrencyPair(c, msg, fromCoin, toCurrency)
		return
	}

	c.SendMsg(msg, "参数数量错误")
}

func handleSingleCrypto(c *qbot.Client, msg *qbot.Message, coin string) {
	log.Printf("查询单个加密货币: %s", coin)
	price, err := getCryptoPrice(coin, "USDT")
	if err != nil {
		log.Printf("查询%s价格失败: %v", coin, err)
		c.SendMsg(msg, fmt.Sprintf("查询失败: %s", err.Error()))
		return
	}
	c.SendMsg(msg, fmt.Sprintf("%s 最新USDT价格: %s", coin, price))
}

func handleCryptoCurrencyPair(c *qbot.Client, msg *qbot.Message, fromCoin string, toCurrency string) {
	log.Printf("查询加密货币对: %s -> %s", fromCoin, toCurrency)

	usdPrice, err := getCryptoPrice(fromCoin, "USD")
	if err != nil {
		log.Printf("查询%s USD价格失败: %v", fromCoin, err)
		c.SendMsg(msg, fmt.Sprintf("查询%s价格失败: %s", fromCoin, err.Error()))
		return
	}

	usdPriceFloat, err := strconv.ParseFloat(usdPrice, 64)
	if err != nil {
		log.Printf("价格解析失败: %v", err)
		c.SendMsg(msg, fmt.Sprintf("价格解析失败: %s", err.Error()))
		return
	}

	if toCurrency == "USD" {
		c.SendMsg(msg, fmt.Sprintf("%s 最新USD价格: %.4f", fromCoin, usdPriceFloat))
		return
	}

	log.Printf("需要汇率换算: USD -> %s", toCurrency)
	exchangeRate, err := getExchangeRate("USD", toCurrency)
	if err != nil {
		log.Printf("获取汇率失败: %v", err)
		c.SendMsg(msg, fmt.Sprintf("获取汇率失败: %s", err.Error()))
		return
	}

	finalPrice := usdPriceFloat * exchangeRate
	log.Printf("换算完成: %s USD价格 %.4f, 汇率 %.4f, 最终价格 %.4f %s", fromCoin, usdPriceFloat, exchangeRate, finalPrice, toCurrency)
	c.SendMsg(msg, fmt.Sprintf("%s 最新%s价格: %.4f", fromCoin, toCurrency, finalPrice))
}

func getCryptoPrice(coin string, quoteCurrency string) (string, error) {
	instId := coin + "-" + quoteCurrency + "-SWAP"
	url := "https://bot-forward.lavacreeper.net/api/v5/market/ticker?instId=" + instId

	log.Printf("请求加密货币价格: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("请求创建失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Okx-Python-Client")
	req.Header.Set("X-API-Key", config.OkxMirrorAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("关闭响应体失败: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	var ticker TickerResp
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return "", fmt.Errorf("解析失败: %v", err)
	}

	if ticker.Code != "0" || len(ticker.Data) == 0 {
		return "", fmt.Errorf("API返回错误: %s", ticker.Msg)
	}

	log.Printf("获取到价格: %s = %s", instId, ticker.Data[0].Last)
	return ticker.Data[0].Last, nil
}

func getExchangeRate(baseCode string, targetCode string) (float64, error) {
	if config.ExchangeRateAPIKey == "" {
		return 0, fmt.Errorf("汇率API密钥未配置")
	}

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", config.ExchangeRateAPIKey, baseCode)

	log.Printf("请求汇率: %s", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("汇率请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("关闭响应体失败: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("汇率API HTTP错误: %d", resp.StatusCode)
	}

	var exchangeData ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeData); err != nil {
		return 0, fmt.Errorf("汇率数据解析失败: %v", err)
	}

	if exchangeData.Result != "success" {
		return 0, fmt.Errorf("汇率API返回错误: %s", exchangeData.Result)
	}

	rate, exists := exchangeData.ConversionRates[targetCode]
	if !exists {
		return 0, fmt.Errorf("不支持的货币: %s", targetCode)
	}

	log.Printf("获取到汇率: 1 %s = %f %s", baseCode, rate, targetCode)
	return rate, nil
}
