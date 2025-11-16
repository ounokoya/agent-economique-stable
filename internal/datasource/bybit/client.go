package bybit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	bybit "github.com/bybit-exchange/bybit.go.api"
)

// Client wraps Bybit API client
type Client struct {
	client *bybit.Client
}

// Kline represents Bybit kline data
type Kline struct {
	OpenTime  time.Time
	CloseTime time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64 // Volume in base asset (SOL)
}

// NewClient creates a new Bybit client
func NewClient() *Client {
	// Production endpoint
	client := bybit.NewBybitHttpClient("", "", bybit.WithBaseURL(bybit.MAINNET))

	return &Client{
		client: client,
	}
}

// GetKlines retrieves klines from Bybit Futures
// symbol format: "SOLUSDT" (no underscore)
// interval: "5", "15", "60", "240" (minutes as string)
func (c *Client) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	// Convert interval to Bybit format
	bybitInterval := convertInterval(interval)
	if bybitInterval == "" {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// Calculate time range
	now := time.Now()
	intervalSeconds := getIntervalSeconds(interval)
	startTime := now.Add(-time.Duration(limit*intervalSeconds) * time.Second)

	// Bybit uses milliseconds for timestamps
	startMs := startTime.UnixMilli()
	endMs := now.UnixMilli()

	// Get klines from Bybit Linear (USDT perpetual)
	params := map[string]interface{}{
		"category": "linear",
		"symbol":   symbol,
		"interval": bybitInterval,
		"start":    fmt.Sprintf("%d", startMs),
		"end":      fmt.Sprintf("%d", endMs),
		"limit":    fmt.Sprintf("%d", limit),
	}

	result, err := c.client.NewUtaBybitServiceWithParams(params).GetMarketKline(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}

	// Parse result
	resultMap, ok := result.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	// Check return code
	if result.RetCode != 0 {
		return nil, fmt.Errorf("bybit API error: %s (code: %d)", result.RetMsg, result.RetCode)
	}

	// Extract data
	listData, ok := resultMap["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no list data")
	}

	// Convert to our Kline format
	klines := make([]Kline, 0, len(listData))
	for _, item := range listData {
		klineData, ok := item.([]interface{})
		if !ok || len(klineData) < 7 {
			continue
		}

		// Bybit kline format: [startTime, open, high, low, close, volume, turnover]
		// All values are strings
		timestampStr, _ := klineData[0].(string)
		openStr, _ := klineData[1].(string)
		highStr, _ := klineData[2].(string)
		lowStr, _ := klineData[3].(string)
		closeStr, _ := klineData[4].(string)
		volumeStr, _ := klineData[5].(string)

		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		open, _ := strconv.ParseFloat(openStr, 64)
		high, _ := strconv.ParseFloat(highStr, 64)
		low, _ := strconv.ParseFloat(lowStr, 64)
		close, _ := strconv.ParseFloat(closeStr, 64)
		volume, _ := strconv.ParseFloat(volumeStr, 64)

		openTime := time.UnixMilli(timestamp)
		closeTime := openTime.Add(time.Duration(intervalSeconds) * time.Second)

		klines = append(klines, Kline{
			OpenTime:  openTime,
			CloseTime: closeTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume, // Volume in base asset (SOL)
		})
	}

	// IMPORTANT: Bybit retourne les klines en ordre INVERSE (plus récent → plus ancien)
	// On doit inverser pour avoir l'ordre chronologique (plus ancien → plus récent)
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}

	return klines, nil
}

// convertInterval converts our interval format to Bybit format
func convertInterval(interval string) string {
	switch interval {
	case "1m":
		return "1"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "30m":
		return "30"
	case "1h":
		return "60"
	case "2h":
		return "120"
	case "4h":
		return "240"
	case "1d":
		return "D"
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
	case "1d":
		return 86400
	default:
		return 300 // default 5m
	}
}
