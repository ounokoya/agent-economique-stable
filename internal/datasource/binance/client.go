package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
)

// Client wraps Binance API client
type Client struct {
	client *binance.Client
}

// Kline represents Binance kline data (compatible avec format unifié)
type Kline struct {
	OpenTime         time.Time
	CloseTime        time.Time
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64 // Volume en asset (SOL pour SOLUSDT)
	QuoteAssetVolume float64 // Volume en quote asset (USDT pour SOLUSDT)
}

// Trade represents Binance trade data
type Trade struct {
	ID           int64     `json:"id"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	Time         time.Time `json:"time"`
	IsBuyerMaker bool      `json:"is_buyer_maker"`
}

// NewClient creates a new Binance client
func NewClient() *Client {
	// Use testnet for demo purposes (no API key required for market data)
	client := binance.NewClient("", "")
	// Set to testnet if needed: client.BaseURL = binance.BaseURLTestnet
	
	return &Client{
		client: client,
	}
}

// GetKlines retrieves klines from Binance
// symbol format: "SOLUSDT" (Binance uses no separator)
// interval: "5m", "15m", "1h", "4h"
func (c *Client) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	// Convert timeframe format
	binanceInterval := convertInterval(interval)
	if binanceInterval == "" {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// Call Binance API
	klines, err := c.client.NewKlinesService().
		Symbol(symbol).
		Interval(binanceInterval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %v", err)
	}

	// Convert to our Kline format
	result := make([]Kline, 0, len(klines))
	for _, kline := range klines {
		open, _ := strconv.ParseFloat(kline.Open, 64)
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)
		close, _ := strconv.ParseFloat(kline.Close, 64)
		volume, _ := strconv.ParseFloat(kline.Volume, 64)

		openTime := time.Unix(kline.OpenTime/1000, 0)
		closeTime := time.Unix(kline.CloseTime/1000, 0)

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

// GetAggTrades retrieves aggregated trades from Binance
// Returns trades within the last few hours (limited by API)
func (c *Client) GetAggTrades(ctx context.Context, symbol string, limit int) ([]Trade, error) {
	// Calculate start time (Binance only allows recent trades)
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Last 24 hours
	
	// Call Binance API for aggregated trades
	aggTrades, err := c.client.NewAggTradesService().
		Symbol(symbol).
		StartTime(startTime.UnixMilli()).
		EndTime(endTime.UnixMilli()).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated trades: %v", err)
	}

	// Convert to our Trade format
	result := make([]Trade, 0, len(aggTrades))
	for _, trade := range aggTrades {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Quantity, 64)
		tradeTime := time.Unix(trade.Timestamp/1000, 0)

		result = append(result, Trade{
			ID:           trade.AggTradeID,
			Price:        price,
			Quantity:     quantity,
			Time:         tradeTime,
			IsBuyerMaker: trade.IsBuyerMaker,
		})
	}

	return result, nil
}

// GetHistoricalAggTrades retrieves historical aggregated trades using kline data as base
// This is a workaround since Binance limits historical trade data access
func (c *Client) GetHistoricalAggTrades(ctx context.Context, symbol, interval string, limit int) ([]Trade, error) {
	// Get historical klines first
	klines, err := c.GetKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get base klines: %v", err)
	}
	
	// Generate realistic trades from klines (improved version)
	trades := make([]Trade, 0, len(klines)*8) // ~8 trades per kline for realism
	
	for i, kline := range klines {
		// Calculate kline duration
		duration := kline.CloseTime.Sub(kline.OpenTime)
		intervalMs := duration.Milliseconds() / 8 // 8 trades per kline
		
		baseTime := kline.OpenTime.UnixMilli()
		
		// Generate 8 realistic trades per kline with price evolution
		trades = append(trades, generateRealisticTrades(i, kline, baseTime, intervalMs)...)
	}
	
	return trades, nil
}

// generateRealisticTrades creates realistic trade sequence from a kline
func generateRealisticTrades(klIndex int, kline Kline, baseTime, intervalMs int64) []Trade {
	trades := make([]Trade, 8)
	
	// Price evolution: Open -> High -> Low -> Close with variations
	prices := []float64{
		kline.Open,                           // Start at open
		(kline.Open + kline.High) / 2,       // Move toward high
		kline.High,                          // Reach high
		(kline.High + kline.Low) / 2,        // Move toward low
		kline.Low,                           // Reach low
		(kline.Low + kline.Close) / 2,       // Move toward close
		(kline.Close + kline.Low) / 2,       // Small pullback
		kline.Close,                         // End at close
	}
	
	// Volume distribution (more volume at extremes)
	volumeDistribution := []float64{0.1, 0.15, 0.2, 0.1, 0.2, 0.1, 0.05, 0.1}
	
	for i := 0; i < 8; i++ {
		trades[i] = Trade{
			ID:           int64(klIndex*8 + i + 1), // Unique ID
			Price:        prices[i],
			Quantity:     kline.Volume * volumeDistribution[i],
			Time:         time.Unix((baseTime+int64(i)*intervalMs)/1000, 0),
			IsBuyerMaker: i%2 == 0, // Alternate buyer/seller for realism
		}
	}
	
	return trades
}

// convertInterval converts our interval format to Binance format
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
