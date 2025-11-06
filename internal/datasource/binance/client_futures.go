// 🔧 CLIENT BINANCE FUTURES PERPÉTUELS
package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// FuturesClient wraps Binance Futures API client
type FuturesClient struct {
	client *futures.Client
}

// Kline represents Binance futures kline data (compatible avec format unifié)
type FuturesKline struct {
	OpenTime         time.Time
	CloseTime        time.Time
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64 // Volume en SOL (base asset)
	QuoteAssetVolume float64 // Volume en USDT (quote asset)
}

// NewFuturesClient creates a new Binance Futures client
func NewFuturesClient() *FuturesClient {
	// Client Futures perpétuels sans API key (market data)
	client := futures.NewClient("", "")
	
	return &FuturesClient{
		client: client,
	}
}

// GetKlines retrieves klines from Binance Futures
// symbol format: "SOLUSDT" (perpetual futures)
// interval: "5m", "15m", "1h", "4h"
func (c *FuturesClient) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]FuturesKline, error) {
	// Convert timeframe format
	binanceInterval := convertInterval(interval)
	if binanceInterval == "" {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// Call Binance Futures API
	klines, err := c.client.NewKlinesService().
		Symbol(symbol).
		Interval(binanceInterval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures klines: %v", err)
	}

	// Convert to our FuturesKline format
	result := make([]FuturesKline, 0, len(klines))
	for _, kline := range klines {
		open, _ := strconv.ParseFloat(kline.Open, 64)
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)
		close, _ := strconv.ParseFloat(kline.Close, 64)
		volume, _ := strconv.ParseFloat(kline.Volume, 64)
		quoteVolume, _ := strconv.ParseFloat(kline.QuoteAssetVolume, 64)

		openTime := time.Unix(kline.OpenTime/1000, 0)
		closeTime := time.Unix(kline.CloseTime/1000, 0)

		result = append(result, FuturesKline{
			OpenTime:         openTime,
			CloseTime:        closeTime,
			Open:             open,
			High:             high,
			Low:              low,
			Close:            close,
			Volume:           volume,      // Volume en SOL (base asset)
			QuoteAssetVolume: quoteVolume, // Volume en USDT (quote asset)
		})
	}

	return result, nil
}

// ConvertToStandardKline converts FuturesKline to standard Kline
func (c *FuturesClient) ConvertToStandardKline(fklines []FuturesKline) []Kline {
	result := make([]Kline, len(fklines))
	for i, fk := range fklines {
		result[i] = Kline{
			OpenTime:         fk.OpenTime,
			CloseTime:        fk.CloseTime,
			Open:             fk.Open,
			High:             fk.High,
			Low:              fk.Low,
			Close:            fk.Close,
			Volume:           fk.Volume,
			QuoteAssetVolume: fk.QuoteAssetVolume,
		}
	}
	return result
}
