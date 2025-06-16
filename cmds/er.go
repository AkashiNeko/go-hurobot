package cmds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	BaseCode        string             `json:"base_code"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

func cmd_er(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 3 {
		c.SendMsg(msg, "用法: fx <源币种> <目标币种>\n例如: fx CNY HKD")
		return
	}

	if config.ExchangeRateAPIKey == "" {
		return
	}

	fromCurrency := strings.ToUpper(args.Contents[1])
	toCurrency := strings.ToUpper(args.Contents[2])

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", config.ExchangeRateAPIKey, fromCurrency)

	log.Println(url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.SendMsg(msg, fmt.Sprintf("%d", resp.StatusCode))
		return
	}

	var exchangeData ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeData); err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	if exchangeData.Result != "success" {
		c.SendMsg(msg, fmt.Sprintf("%v", exchangeData.Result))
		return
	}

	toRate, exists := exchangeData.ConversionRates[toCurrency]
	if !exists {
		c.SendMsg(msg, fmt.Sprintf("Unsupported %s", toCurrency))
		return
	}

	fromRate, exists := exchangeData.ConversionRates[fromCurrency]
	if !exists {
		c.SendMsg(msg, fmt.Sprintf("Unsupported %s", fromCurrency))
		return
	}

	rate1to2 := toRate / fromRate
	rate2to1 := fromRate / toRate

	result := fmt.Sprintf("1 %s = %.4f %s\n1 %s = %.4f %s",
		fromCurrency, rate1to2, toCurrency,
		toCurrency, rate2to1, fromCurrency)

	c.SendMsg(msg, result)
}
