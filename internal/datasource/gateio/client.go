package gateio

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v6"
)

// Client wraps Gate.io API client
type Client struct {
	client *gateapi.APIClient
}

// Kline represents Gate.io kline data (compatible avec BingX format)
type Kline struct {
	OpenTime   time.Time
	CloseTime  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
}

// NewClient creates a new Gate.io client
func NewClient() *Client {
	config := gateapi.NewConfiguration()
	// Futures live trading API endpoint
	config.BasePath = "https://fx-api.gateio.ws/api/v4"
	
	client := gateapi.NewAPIClient(config)
	
	return &Client{
		client: client,
	}
}

// GetKlines retrieves klines from Gate.io
// symbol format: "SOL_USDT" (Gate.io uses underscore)
// interval: "5m", "15m", "1h", "4h"
func (c *Client) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	// Convert timeframe format
	gateInterval := convertInterval(interval)
	if gateInterval == "" {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// Set time range (last 'limit' periods)
	to := time.Now().Unix()
	
	// Calculate 'from' based on interval and limit
	intervalSeconds := getIntervalSeconds(interval)
	from := to - int64(limit*intervalSeconds)

	// Call Gate.io FUTURES API (perpétuels comme demandé)
	// NOTE: Ne pas utiliser limit avec from/to en même temps
	opts := &gateapi.ListFuturesCandlesticksOpts{
		From:     optional.NewInt64(from),
		To:       optional.NewInt64(to),
		Interval: optional.NewString(gateInterval),
	}

	candlesticks, _, err := c.client.FuturesApi.ListFuturesCandlesticks(ctx, "usdt", symbol, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures candlesticks: %v", err)
	}

	// Convert to our Kline format
	klines := make([]Kline, 0, len(candlesticks))
	for _, candle := range candlesticks {
		// Gate.io FUTURES candlestick struct avec champs nommés
		// T = timestamp, V = volume SOL (base asset), C = close, H = high, L = low, O = open
		
		timestamp := int64(candle.T)
		
		// Parse les prix
		close, _ := strconv.ParseFloat(candle.C, 64)
		high, _ := strconv.ParseFloat(candle.H, 64)
		low, _ := strconv.ParseFloat(candle.L, 64)
		open, _ := strconv.ParseFloat(candle.O, 64)
		
		// Volume SOL (base asset)
		volumeSOL := float64(candle.V)

		openTime := time.Unix(timestamp, 0)
		closeTime := openTime.Add(time.Duration(intervalSeconds) * time.Second)

		klines = append(klines, Kline{
			OpenTime:  openTime,
			CloseTime: closeTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volumeSOL,  // Volume SOL (base asset)
		})
	}

	// Tri défensif : s'assurer que les klines sont en ordre chronologique
	// (du plus ancien au plus récent)
	for i := 0; i < len(klines)-1; i++ {
		if klines[i].OpenTime.After(klines[i+1].OpenTime) {
			// Ordre inversé détecté, on inverse tout
			for j, k := 0, len(klines)-1; j < k; j, k = j+1, k-1 {
				klines[j], klines[k] = klines[k], klines[j]
			}
			break
		}
	}

	return klines, nil
}

// convertInterval converts our interval format to Gate.io format
func convertInterval(interval string) string {
	switch interval {
	case "1m":
		return "1m"
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "30m":
		return "30m"
	case "1h":
		return "1h"
	case "2h":
		return "2h"
	case "4h":
		return "4h"
	case "8h":
		return "8h"
	case "12h":
		return "12h"
	case "1d":
		return "1d"
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
