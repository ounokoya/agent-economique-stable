package kucoin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Kucoin/kucoin-go-sdk"
)

// Client wraps KuCoin API client
type Client struct {
	client *kucoin.ApiService
}

// Kline represents KuCoin kline data (compatible avec format unifié)
type Kline struct {
	OpenTime   time.Time
	CloseTime  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
}

// NewClient creates a new KuCoin client
func NewClient() *Client {
	// Use sandbox for demo purposes (no API key required for market data)
	s := kucoin.NewApiService(
		kucoin.ApiBaseURIOption("https://api.kucoin.com"), // Production
		kucoin.ApiKeyOption(""),
		kucoin.ApiSecretOption(""),
		kucoin.ApiPassPhraseOption(""),
	)
	
	return &Client{
		client: s,
	}
}

// GetKlines retrieves klines from KuCoin
// symbol format: "SOL-USDT" (KuCoin uses dash separator)
// interval: "5min", "15min", "1hour", "4hour"
func (c *Client) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	// Convert timeframe format
	kucoinInterval := convertInterval(interval)
	if kucoinInterval == "" {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// Calculate time range
	endAt := time.Now().Unix()
	startAt := endAt - int64(limit*getIntervalSeconds(interval))

	// Call KuCoin API
	rsp, err := c.client.KLines(ctx, symbol, kucoinInterval, startAt, endAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %v", err)
	}

	// Parse response as interface{} and handle the actual structure
	var rawData interface{}
	if err := rsp.ReadData(&rawData); err != nil {
		return nil, fmt.Errorf("failed to parse klines: %v", err)
	}

	// KuCoin returns an array of arrays like: [["timestamp", "open", "close", "high", "low", "volume", "turnover"], ...]
	klinesArray, ok := rawData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected KuCoin data format")
	}

	// Convert to our Kline format
	result := make([]Kline, 0, len(klinesArray))
	for _, item := range klinesArray {
		klineArray, ok := item.([]interface{})
		if !ok || len(klineArray) < 6 {
			continue
		}

		timestamp, _ := strconv.ParseInt(fmt.Sprintf("%v", klineArray[0]), 10, 64)
		open, _ := strconv.ParseFloat(fmt.Sprintf("%v", klineArray[1]), 64)
		close, _ := strconv.ParseFloat(fmt.Sprintf("%v", klineArray[2]), 64)
		high, _ := strconv.ParseFloat(fmt.Sprintf("%v", klineArray[3]), 64)
		low, _ := strconv.ParseFloat(fmt.Sprintf("%v", klineArray[4]), 64)
		volume, _ := strconv.ParseFloat(fmt.Sprintf("%v", klineArray[5]), 64)

		openTime := time.Unix(timestamp, 0)
		closeTime := openTime.Add(time.Duration(getIntervalSeconds(interval)) * time.Second)

		result = append(result, Kline{
			OpenTime:  openTime,
			CloseTime: closeTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return result, nil
}

// convertInterval converts our interval format to KuCoin format
func convertInterval(interval string) string {
	switch interval {
	case "1m":
		return "1min"
	case "5m":
		return "5min"
	case "15m":
		return "15min"
	case "30m":
		return "30min"
	case "1h":
		return "1hour"
	case "2h":
		return "2hour"
	case "4h":
		return "4hour"
	case "8h":
		return "8hour"
	case "12h":
		return "12hour"
	case "1d":
		return "1day"
	default:
		return ""
	}
}

// getIntervalSeconds returns interval duration in seconds
func getIntervalSeconds(interval string) int {
	switch interval {
	case "1m":
		return 60
	case "5m":
		return 300
	case "15m":
		return 900
	case "30m":
		return 1800
	case "1h":
		return 3600
	case "2h":
		return 7200
	case "4h":
		return 14400
	case "8h":
		return 28800
	case "12h":
		return 43200
	case "1d":
		return 86400
	default:
		return 300 // default 5m
	}
}
