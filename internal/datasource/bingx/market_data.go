package bingx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MarketDataService provides market data operations for BingX
type MarketDataService struct {
	client *Client
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(client *Client) *MarketDataService {
	return &MarketDataService{
		client: client,
	}
}

// GetSymbols retrieves all available trading symbols
func (m *MarketDataService) GetSymbols(ctx context.Context) ([]Symbol, error) {
	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/spot/v1/common/symbols", nil, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	var symbols []Symbol
	if err := m.parseDataArray(resp.Data, &symbols); err != nil {
		return nil, fmt.Errorf("failed to parse symbols response: %w", err)
	}

	return symbols, nil
}

// GetPrice retrieves current price for a symbol
func (m *MarketDataService) GetPrice(ctx context.Context, symbol string) (*Ticker, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	params := map[string]string{
		"symbol": symbol,
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/spot/v1/ticker/price", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}

	var ticker Ticker
	if err := m.parseDataObject(resp.Data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	ticker.Symbol = symbol
	ticker.Timestamp = time.Now()

	return &ticker, nil
}

// GetTicker24hr retrieves 24hr ticker statistics for a symbol
func (m *MarketDataService) GetTicker24hr(ctx context.Context, symbol string) (*Ticker, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	params := map[string]string{
		"symbol": symbol,
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/spot/v1/ticker/24hr", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get 24hr ticker for %s: %w", symbol, err)
	}

	var ticker Ticker
	if err := m.parseDataObject(resp.Data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to parse ticker response: %w", err)
	}

	ticker.Timestamp = time.Now()

	return &ticker, nil
}

// GetKlines retrieves candlestick data for a symbol
func (m *MarketDataService) GetKlines(ctx context.Context, symbol, interval string, limit int, startTime, endTime *time.Time) ([]Kline, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	if interval == "" {
		return nil, fmt.Errorf("interval is required")
	}

	if !m.isValidInterval(interval) {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	params := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}

	if limit > 0 {
		if limit > 1500 {
			limit = 1500 // BingX maximum limit
		}
		params["limit"] = strconv.Itoa(limit)
	}

	if startTime != nil {
		params["startTime"] = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	if endTime != nil {
		params["endTime"] = strconv.FormatInt(endTime.UnixMilli(), 10)
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/spot/v1/market/kline", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines for %s: %w", symbol, err)
	}

	var rawKlines [][]interface{}
	if err := m.parseDataArray(resp.Data, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse klines response: %w", err)
	}

	return m.parseKlinesData(rawKlines)
}

// GetFuturesKlines retrieves candlestick data for futures symbols
func (m *MarketDataService) GetFuturesKlines(ctx context.Context, symbol, interval string, limit int, startTime, endTime *time.Time) ([]Kline, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	if interval == "" {
		return nil, fmt.Errorf("interval is required")
	}

	if !m.isValidInterval(interval) {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	params := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}

	if limit > 0 {
		if limit > 1500 {
			limit = 1500 // BingX maximum limit
		}
		params["limit"] = strconv.Itoa(limit)
	}

	if startTime != nil {
		params["startTime"] = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	if endTime != nil {
		params["endTime"] = strconv.FormatInt(endTime.UnixMilli(), 10)
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/swap/v2/quote/klines", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures klines for %s: %w", symbol, err)
	}

	var rawKlines [][]interface{}
	if err := m.parseDataArray(resp.Data, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse futures klines response: %w", err)
	}

	return m.parseKlinesData(rawKlines)
}

// GetFuturesPrice retrieves current price for futures symbol
func (m *MarketDataService) GetFuturesPrice(ctx context.Context, symbol string) (*Ticker, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	params := map[string]string{
		"symbol": symbol,
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/swap/v2/quote/price", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures price for %s: %w", symbol, err)
	}

	var ticker Ticker
	if err := m.parseDataObject(resp.Data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to parse futures price response: %w", err)
	}

	ticker.Symbol = symbol
	ticker.Timestamp = time.Now()

	return &ticker, nil
}

// GetFuturesTicker24hr retrieves 24hr ticker statistics for futures symbol
func (m *MarketDataService) GetFuturesTicker24hr(ctx context.Context, symbol string) (*Ticker, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	params := map[string]string{
		"symbol": symbol,
	}

	resp, err := m.client.DoRequest(ctx, http.MethodGet, "/openApi/swap/v2/quote/ticker", params, EndpointTypeMarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to get futures ticker for %s: %w", symbol, err)
	}

	var ticker Ticker
	if err := m.parseDataObject(resp.Data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to parse futures ticker response: %w", err)
	}

	ticker.Timestamp = time.Now()

	return &ticker, nil
}

// parseKlinesData converts raw API response to Kline structs
// BingX spot kline format: [timestamp, open, high, low, close, volume_base, end_timestamp, volume_quote]
func (m *MarketDataService) parseKlinesData(rawKlines [][]interface{}) ([]Kline, error) {
	klines := make([]Kline, len(rawKlines))

	for i, rawKline := range rawKlines {
		if len(rawKline) < 8 {
			return nil, fmt.Errorf("invalid kline data at index %d: insufficient fields", i)
		}

		// Parse timestamp
		openTimeMs, err := m.parseFloat64(rawKline[0])
		if err != nil {
			return nil, fmt.Errorf("invalid open time at index %d: %w", i, err)
		}
		openTime := time.Unix(int64(openTimeMs)/1000, (int64(openTimeMs)%1000)*1000000)

		closeTimeMs, err := m.parseFloat64(rawKline[6])
		if err != nil {
			return nil, fmt.Errorf("invalid close time at index %d: %w", i, err)
		}
		closeTime := time.Unix(int64(closeTimeMs)/1000, (int64(closeTimeMs)%1000)*1000000)

		// Parse prices (open, high, low, close)
		open, err := m.parseFloat64(rawKline[1])   // Open price
		if err != nil {
			return nil, fmt.Errorf("invalid open price at index %d: %w", i, err)
		}

		high, err := m.parseFloat64(rawKline[2])   // High price
		if err != nil {
			return nil, fmt.Errorf("invalid high price at index %d: %w", i, err)
		}

		low, err := m.parseFloat64(rawKline[3])    // Low price
		if err != nil {
			return nil, fmt.Errorf("invalid low price at index %d: %w", i, err)
		}

		close, err := m.parseFloat64(rawKline[4])  // Close price
		if err != nil {
			return nil, fmt.Errorf("invalid close price at index %d: %w", i, err)
		}

		volume, err := m.parseFloat64(rawKline[7])  // Volume USDT (champ 7)
		if err != nil {
			return nil, fmt.Errorf("invalid volume at index %d: %w", i, err)
		}

		klines[i] = Kline{
			OpenTime:  openTime,
			CloseTime: closeTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}
	}

	return klines, nil
}

// parseFloat64 safely converts interface{} to float64
func (m *MarketDataService) parseFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// parseDataObject parses API response data as a single object
func (m *MarketDataService) parseDataObject(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// parseDataArray parses API response data as an array
func (m *MarketDataService) parseDataArray(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// isValidInterval checks if the provided interval is valid for BingX
func (m *MarketDataService) isValidInterval(interval string) bool {
	validIntervals := []string{
		"1m", "3m", "5m", "15m", "30m",
		"1h", "2h", "4h", "6h", "8h", "12h",
		"1d", "3d", "1w", "1M",
	}

	for _, valid := range validIntervals {
		if interval == valid {
			return true
		}
	}

	return false
}

// ValidateSymbol validates if symbol format is correct for BingX
func (m *MarketDataService) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	// Basic validation - should contain letters and possibly dash
	if len(symbol) < 3 {
		return fmt.Errorf("symbol too short: %s", symbol)
	}

	// Check for valid characters (letters, numbers, dash)
	for _, char := range symbol {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || 
			 (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("invalid character in symbol: %c", char)
		}
	}

	return nil
}

// GetValidIntervals returns all valid intervals supported by BingX
func (m *MarketDataService) GetValidIntervals() []string {
	return []string{
		"1m", "3m", "5m", "15m", "30m",
		"1h", "2h", "4h", "6h", "8h", "12h",
		"1d", "3d", "1w", "1M",
	}
}

// FormatSymbol converts symbol to BingX format (e.g., "BTCUSDT" -> "BTC-USDT")
func (m *MarketDataService) FormatSymbol(symbol string) string {
	symbol = strings.ToUpper(symbol)
	
	// If already contains dash, return as-is
	if strings.Contains(symbol, "-") {
		return symbol
	}
	
	// Common USDT pairs - add dash before USDT
	if strings.HasSuffix(symbol, "USDT") && len(symbol) > 4 {
		base := symbol[:len(symbol)-4]
		return base + "-USDT"
	}
	
	// Common BTC pairs - add dash before BTC
	if strings.HasSuffix(symbol, "BTC") && len(symbol) > 3 {
		base := symbol[:len(symbol)-3]
		return base + "-BTC"
	}
	
	// Common ETH pairs - add dash before ETH
	if strings.HasSuffix(symbol, "ETH") && len(symbol) > 3 {
		base := symbol[:len(symbol)-3]
		return base + "-ETH"
	}
	
	// Return as-is if no common pattern found
	return symbol
}
