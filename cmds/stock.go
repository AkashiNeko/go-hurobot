package cmds

import (
	"context"
	"errors"
	"fmt"
	"go-hurobot/qbot"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/longportapp/openapi-go/config"
	"github.com/longportapp/openapi-go/quote"
	"github.com/shopspring/decimal"
)

type KlineData struct {
	Symbol    string          `json:"symbol"`
	Timestamp time.Time       `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    int64           `json:"volume"`
	Turnover  decimal.Decimal `json:"turnover"`
	Date      string          `json:"date"`
}

type HistoryKlineResult struct {
	Symbol       string      `json:"symbol"`
	Period       string      `json:"period"`
	StartDate    string      `json:"start_date"`
	EndDate      string      `json:"end_date"`
	Count        int         `json:"count"`
	Candlesticks []KlineData `json:"candlesticks"`
}

func GetHistoryDailyKlineByOffset(ctx context.Context, quoteCtx *quote.QuoteContext, symbol, baseDate string, count int, forward bool, adjustType int32) (*HistoryKlineResult, error) {

	baseDateParsed, err := time.Parse("2006-01-02", baseDate)
	if err != nil {
		return nil, err
	}

	candlesticks, err := quoteCtx.HistoryCandlesticksByOffset(ctx, symbol, quote.PeriodDay, quote.AdjustType(adjustType), forward, &baseDateParsed, int32(count))
	if err != nil {
		return nil, err
	}

	var klineDataList []KlineData
	for _, candle := range candlesticks {
		klineData := KlineData{
			Symbol:    symbol,
			Timestamp: time.Unix(candle.Timestamp, 0),
			Volume:    candle.Volume,
		}

		if candle.Open != nil {
			klineData.Open = *candle.Open
		}
		if candle.High != nil {
			klineData.High = *candle.High
		}
		if candle.Low != nil {
			klineData.Low = *candle.Low
		}
		if candle.Close != nil {
			klineData.Close = *candle.Close
		}
		if candle.Turnover != nil {
			klineData.Turnover = *candle.Turnover
		}

		klineData.Date = klineData.Timestamp.Format("2006-01-02")

		klineDataList = append(klineDataList, klineData)
	}

	result := &HistoryKlineResult{
		Symbol:       symbol,
		Period:       "1Day",
		StartDate:    baseDate,
		EndDate:      baseDate,
		Count:        len(klineDataList),
		Candlesticks: klineDataList,
	}

	return result, nil
}

func GetHistoryDailyKlineByDateRange(ctx context.Context, quoteCtx *quote.QuoteContext, symbol, startDate, endDate string, adjustType int32) (*HistoryKlineResult, error) {

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	candlesticks, err := quoteCtx.HistoryCandlesticksByDate(ctx, symbol, quote.PeriodDay, quote.AdjustType(adjustType), &startTime, &endTime)
	if err != nil {
		return nil, err
	}

	var klineDataList []KlineData
	for _, candle := range candlesticks {
		klineData := KlineData{
			Symbol:    symbol,
			Timestamp: time.Unix(candle.Timestamp, 0),
			Volume:    candle.Volume,
		}

		if candle.Open != nil {
			klineData.Open = *candle.Open
		}
		if candle.High != nil {
			klineData.High = *candle.High
		}
		if candle.Low != nil {
			klineData.Low = *candle.Low
		}
		if candle.Close != nil {
			klineData.Close = *candle.Close
		}
		if candle.Turnover != nil {
			klineData.Turnover = *candle.Turnover
		}

		klineData.Date = klineData.Timestamp.Format("2006-01-02")

		klineDataList = append(klineDataList, klineData)
	}

	result := &HistoryKlineResult{
		Symbol:       symbol,
		Period:       "1Day",
		StartDate:    startDate,
		EndDate:      endDate,
		Count:        len(klineDataList),
		Candlesticks: klineDataList,
	}

	return result, nil
}

func GenerateStockChart(symbol string, days int) (string, error) {

	conf, err := config.New()
	if err != nil {
		return "", err
	}

	quoteCtx, err := quote.NewFromCfg(conf)
	if err != nil {
		return "", err
	}
	defer quoteCtx.Close()

	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	result, err := GetHistoryDailyKlineByOffset(ctx, quoteCtx, symbol, today, days, false, 1)
	if err != nil {
		return "", err
	}

	if result == nil || len(result.Candlesticks) == 0 {
		return "", errors.New("no data")
	}

	resDir := "res"
	if err := os.MkdirAll(resDir, 0755); err != nil {
		return "", err
	}

	if err := drawKlineToPNG(result, symbol, symbol, resDir); err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://akashi.top/res/%s.png", symbol)
	return url, nil
}

func DrawKlineChart(result *HistoryKlineResult, filename, title string) error {
	if result == nil || len(result.Candlesticks) == 0 {
		return errors.New("no data")
	}

	dates := make([]string, 0, len(result.Candlesticks))
	klineData := make([]opts.KlineData, 0, len(result.Candlesticks))
	volumes := make([]opts.BarData, 0, len(result.Candlesticks))

	var minPrice, maxPrice float64
	for i, candle := range result.Candlesticks {
		dates = append(dates, candle.Date)

		klineData = append(klineData, opts.KlineData{
			Value: [4]float64{
				candle.Open.InexactFloat64(),
				candle.Close.InexactFloat64(),
				candle.Low.InexactFloat64(),
				candle.High.InexactFloat64(),
			},
		})

		volumes = append(volumes, opts.BarData{
			Value: float64(candle.Volume),
		})

		high := candle.High.InexactFloat64()
		low := candle.Low.InexactFloat64()

		if i == 0 {
			minPrice = low
			maxPrice = high
		} else {
			if low < minPrice {
				minPrice = low
			}
			if high > maxPrice {
				maxPrice = high
			}
		}
	}

	priceRange := maxPrice - minPrice
	yAxisMin := minPrice - priceRange*0.03
	yAxisMax := maxPrice + priceRange*0.03

	if yAxisMin < 0 {
		yAxisMin = minPrice * 0.97
	}

	kline := charts.NewKLine()
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: fmt.Sprintf("数据范围: %s 到 %s (共%d条)", result.StartDate, result.EndDate, result.Count),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			Min:  yAxisMin,
			Max:  yAxisMax,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Data: []string{"K线", "成交量"},
		}),
	)

	kline.SetXAxis(dates).AddSeries("K线", klineData)

	volume := charts.NewBar()
	volume.SetGlobalOptions(
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
	)
	volume.SetXAxis(dates).AddSeries("成交量", volumes)

	page := components.NewPage()
	page.AddCharts(kline, volume)

	outputDir := "charts"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(outputDir, filename+".html")

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	return nil
}

func DrawSimpleKlineChart(result *HistoryKlineResult, symbol, title string, outputFormat string) error {
	if result == nil || len(result.Candlesticks) == 0 {
		return errors.New("no data")
	}

	kline := charts.NewKLine()

	xAxis := []string{}
	klineData := []opts.KlineData{}

	var minPrice, maxPrice decimal.Decimal
	for i, candle := range result.Candlesticks {
		if i == 0 {
			minPrice = candle.Low
			maxPrice = candle.High
		} else {
			if candle.Low.LessThan(minPrice) {
				minPrice = candle.Low
			}
			if candle.High.GreaterThan(maxPrice) {
				maxPrice = candle.High
			}
		}
	}

	priceRange := maxPrice.Sub(minPrice)
	yAxisMin := minPrice.Sub(priceRange.Mul(decimal.NewFromFloat(0.03)))
	yAxisMax := maxPrice.Add(priceRange.Mul(decimal.NewFromFloat(0.03)))

	if yAxisMin.LessThan(decimal.Zero) {
		yAxisMin = minPrice.Mul(decimal.NewFromFloat(0.97))
	}

	for _, candle := range result.Candlesticks {
		xAxis = append(xAxis, candle.Date)
		klineData = append(klineData, opts.KlineData{
			Value: [4]interface{}{
				candle.Open.InexactFloat64(),
				candle.Close.InexactFloat64(),
				candle.Low.InexactFloat64(),
				candle.High.InexactFloat64(),
			},
		})
	}

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(true),
			Min:   yAxisMin.InexactFloat64(),
			Max:   yAxisMax.InexactFloat64(),
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
	)

	kline.SetXAxis(xAxis).AddSeries("K线", klineData)

	chartsDir := "charts"
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		return err
	}

	if outputFormat == "png" {
		return drawKlineToPNG(result, symbol, title, chartsDir)
	} else {

		filename := filepath.Join(chartsDir, fmt.Sprintf("%s.html", symbol))
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		return kline.Render(file)
	}
}

func DrawCompareChart(results []*HistoryKlineResult, filename, title string) error {
	if len(results) == 0 {
		return errors.New("no data")
	}

	firstResult := results[0]
	dates := make([]string, 0, len(firstResult.Candlesticks))
	for _, candle := range firstResult.Candlesticks {
		dates = append(dates, candle.Date)
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: "收盘价对比图",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{}),
	)

	var minPrice, maxPrice float64
	firstStock := true

	for _, result := range results {
		closeData := make([]opts.LineData, 0, len(result.Candlesticks))
		for _, candle := range result.Candlesticks {
			closePrice := candle.Close.InexactFloat64()
			closeData = append(closeData, opts.LineData{
				Value: closePrice,
			})

			if firstStock {
				minPrice = closePrice
				maxPrice = closePrice
				firstStock = false
			} else {
				if closePrice < minPrice {
					minPrice = closePrice
				}
				if closePrice > maxPrice {
					maxPrice = closePrice
				}
			}
		}
		line.SetXAxis(dates).AddSeries(result.Symbol, closeData)
	}

	priceRange := maxPrice - minPrice
	yAxisMin := minPrice - priceRange*0.03
	yAxisMax := maxPrice + priceRange*0.03

	if yAxisMin < 0 {
		yAxisMin = minPrice * 0.97
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: "收盘价对比图",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			Min:  yAxisMin,
			Max:  yAxisMax,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{}),
	)

	outputDir := "charts"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(outputDir, filename+".html")

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := line.Render(io.MultiWriter(f)); err != nil {
		return err
	}
	return nil
}

func ConvertHTMLToPNG(symbol string) error {

	htmlPath := filepath.Join("charts", fmt.Sprintf("%s.html", symbol))

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return errors.New("file not found")
	}

	pngPath := filepath.Join("charts", fmt.Sprintf("%s.png", symbol))

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	absHTMLPath, err := filepath.Abs(htmlPath)
	if err != nil {
		return err
	}

	fileURL := "file://" + absHTMLPath

	var buf []byte
	if err := chromedp.Run(ctx, screenshot(fileURL, &buf)); err != nil {
		return err
	}

	if err := os.WriteFile(pngPath, buf, 0644); err != nil {
		return err
	}

	return nil
}

func screenshot(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible("canvas", chromedp.ByQuery),
		chromedp.Sleep(2 * time.Second),
		chromedp.Screenshot("body", res, chromedp.NodeVisible),
	}
}

func drawKlineToPNG(result *HistoryKlineResult, symbol, title string, chartsDir string) error {

	width, height := 1200, 800
	margin := 80

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)

	plotX := margin
	plotY := margin
	plotWidth := width - 2*margin
	plotHeight := height - 2*margin

	var minPrice, maxPrice float64
	for i, candle := range result.Candlesticks {
		high := candle.High.InexactFloat64()
		low := candle.Low.InexactFloat64()

		if i == 0 {
			minPrice = low
			maxPrice = high
		} else {
			if low < minPrice {
				minPrice = low
			}
			if high > maxPrice {
				maxPrice = high
			}
		}
	}

	priceRange := maxPrice - minPrice
	minPrice -= priceRange * 0.03
	maxPrice += priceRange * 0.03

	candleWidth := plotWidth / len(result.Candlesticks)
	for i, candle := range result.Candlesticks {
		x := plotX + i*candleWidth + candleWidth/2

		open := candle.Open.InexactFloat64()
		close := candle.Close.InexactFloat64()
		high := candle.High.InexactFloat64()
		low := candle.Low.InexactFloat64()

		highY := plotY + int(float64(plotHeight)*(maxPrice-high)/(maxPrice-minPrice))
		lowY := plotY + int(float64(plotHeight)*(maxPrice-low)/(maxPrice-minPrice))
		openY := plotY + int(float64(plotHeight)*(maxPrice-open)/(maxPrice-minPrice))
		closeY := plotY + int(float64(plotHeight)*(maxPrice-close)/(maxPrice-minPrice))

		var candleColor color.RGBA
		if close >= open {
			candleColor = color.RGBA{0, 255, 0, 255}
		} else {
			candleColor = color.RGBA{255, 0, 0, 255}
		}

		drawLine(img, x, highY, x, lowY, candleColor)

		bodyTop := int(math.Min(float64(openY), float64(closeY)))
		bodyBottom := int(math.Max(float64(openY), float64(closeY)))
		bodyHeight := bodyBottom - bodyTop
		if bodyHeight < 1 {
			bodyHeight = 1
		}

		candleRect := image.Rect(x-candleWidth/4, bodyTop, x+candleWidth/4, bodyBottom)
		draw.Draw(img, candleRect, &image.Uniform{candleColor}, image.Point{}, draw.Src)
	}

	drawRect(img, plotX, plotY, plotWidth, plotHeight, color.RGBA{255, 255, 255, 255})

	pngPath := filepath.Join(chartsDir, symbol+".png")
	file, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return err
	}

	return nil
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	if x1 == x2 {

		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for y := y1; y <= y2; y++ {
			if x1 >= 0 && x1 < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x1, y, c)
			}
		}
	}
}

func drawRect(img *image.RGBA, x, y, width, height int, c color.RGBA) {

	for i := range width {
		if x+i >= 0 && x+i < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x+i, y, c)
		}
	}

	for i := range width {
		if x+i >= 0 && x+i < img.Bounds().Dx() && y+height >= 0 && y+height < img.Bounds().Dy() {
			img.Set(x+i, y+height, c)
		}
	}

	for i := range height {
		if x >= 0 && x < img.Bounds().Dx() && y+i >= 0 && y+i < img.Bounds().Dy() {
			img.Set(x, y+i, c)
		}
	}

	for i := range height {
		if x+width >= 0 && x+width < img.Bounds().Dx() && y+i >= 0 && y+i < img.Bounds().Dy() {
			img.Set(x+width, y+i, c)
		}
	}
}

func extractAPIErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	if idx := strings.Index(errStr, "message:"); idx != -1 {
		message := strings.TrimSpace(errStr[idx+8:])
		if message != "" {
			return message
		}
	}
	return errStr
}

func cmd_stock(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		c.SendReplyMsg(msg, "Usage: stock <code> [days]")
		return
	}

	symbol := args.Contents[1]
	days := 60

	if args.Size >= 3 {
		if parsedDays, err := strconv.Atoi(args.Contents[2]); err == nil {
			if parsedDays > 0 && parsedDays <= 365 {
				days = parsedDays
			} else {
				c.SendReplyMsg(msg, "1 <= days <= 365")
				return
			}
		} else {
			c.SendReplyMsg(msg, "<days> is invalid")
			return
		}
	}

	imageURL, err := GenerateStockChart(symbol, days)
	if err != nil {
		c.SendReplyMsg(msg, extractAPIErrorMessage(err))
		return
	}

	c.SendImage(msg, imageURL)
}
