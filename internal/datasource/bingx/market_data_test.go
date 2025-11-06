package bingx

import (
	"context"
	"testing"
)

func TestNewMarketDataService(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	if service == nil {
		t.Fatal("NewMarketDataService should not return nil")
	}
	
	if service.client != client {
		t.Error("Service should reference the provided client")
	}
}

func TestMarketDataServiceValidateSymbol(t *testing.T) {
	service := &MarketDataService{}
	
	tests := []struct {
		name        string
		symbol      string
		expectError bool
	}{
		{
			name:        "Valid symbol BTC-USDT",
			symbol:      "BTC-USDT",
			expectError: false,
		},
		{
			name:        "Valid symbol ETHUSDT",
			symbol:      "ETHUSDT",
			expectError: false,
		},
		{
			name:        "Valid symbol with numbers",
			symbol:      "SOL1-USDT",
			expectError: false,
		},
		{
			name:        "Empty symbol",
			symbol:      "",
			expectError: true,
		},
		{
			name:        "Too short symbol",
			symbol:      "BT",
			expectError: true,
		},
		{
			name:        "Invalid character",
			symbol:      "BTC@USDT",
			expectError: true,
		},
		{
			name:        "Invalid character space",
			symbol:      "BTC USDT",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSymbol(tt.symbol)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestMarketDataServiceIsValidInterval(t *testing.T) {
	service := &MarketDataService{}
	
	validIntervals := []string{
		"1m", "3m", "5m", "15m", "30m",
		"1h", "2h", "4h", "6h", "8h", "12h",
		"1d", "3d", "1w", "1M",
	}
	
	for _, interval := range validIntervals {
		if !service.isValidInterval(interval) {
			t.Errorf("Interval %s should be valid", interval)
		}
	}
	
	invalidIntervals := []string{
		"", "2m", "10m", "45m", "5h", "2d", "1y", "invalid",
	}
	
	for _, interval := range invalidIntervals {
		if service.isValidInterval(interval) {
			t.Errorf("Interval %s should be invalid", interval)
		}
	}
}

func TestMarketDataServiceGetValidIntervals(t *testing.T) {
	service := &MarketDataService{}
	
	intervals := service.GetValidIntervals()
	
	expectedCount := 15 // Total number of valid intervals
	if len(intervals) != expectedCount {
		t.Errorf("Expected %d intervals, got %d", expectedCount, len(intervals))
	}
	
	// Check some key intervals exist
	expectedIntervals := []string{"5m", "15m", "1h", "4h", "1d"}
	for _, expected := range expectedIntervals {
		found := false
		for _, interval := range intervals {
			if interval == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected interval %s not found", expected)
		}
	}
}

func TestMarketDataServiceFormatSymbol(t *testing.T) {
	service := &MarketDataService{}
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BTCUSDT to BTC-USDT",
			input:    "BTCUSDT",
			expected: "BTC-USDT",
		},
		{
			name:     "ethusdt to ETH-USDT",
			input:    "ethusdt",
			expected: "ETH-USDT",
		},
		{
			name:     "SOLUSDT to SOL-USDT",
			input:    "SOLUSDT",
			expected: "SOL-USDT",
		},
		{
			name:     "Already formatted BTC-USDT",
			input:    "BTC-USDT",
			expected: "BTC-USDT",
		},
		{
			name:     "ETHBTC to ETH-BTC",
			input:    "ETHBTC",
			expected: "ETH-BTC",
		},
		{
			name:     "USDCETH to USDC-ETH",
			input:    "USDCETH",
			expected: "USDC-ETH",
		},
		{
			name:     "Unknown format stays same",
			input:    "UNKNOWN",
			expected: "UNKNOWN",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FormatSymbol(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMarketDataServiceParseFloat64(t *testing.T) {
	service := &MarketDataService{}
	
	tests := []struct {
		name        string
		input       interface{}
		expected    float64
		expectError bool
	}{
		{
			name:        "Float64 input",
			input:       123.456,
			expected:    123.456,
			expectError: false,
		},
		{
			name:        "Int input",
			input:       42,
			expected:    42.0,
			expectError: false,
		},
		{
			name:        "String input valid",
			input:       "98.765",
			expected:    98.765,
			expectError: false,
		},
		{
			name:        "String input invalid",
			input:       "not_a_number",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Bool input invalid",
			input:       true,
			expected:    0,
			expectError: true,
		},
		{
			name:        "Nil input invalid",
			input:       nil,
			expected:    0,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseFloat64(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}

func TestMarketDataServiceParseKlinesData(t *testing.T) {
	service := &MarketDataService{}
	
	// Valid klines data
	rawKlines := [][]interface{}{
		{
			1640995200000.0, // open time
			"45000.00",      // open
			"45100.00",      // high
			"44900.00",      // low
			"45050.00",      // close
			"100.5",         // volume
			1640995259999.0, // close time
			"extra_field",   // additional field (ignored)
		},
	}
	
	klines, err := service.parseKlinesData(rawKlines)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	if len(klines) != 1 {
		t.Fatalf("Expected 1 kline, got %d", len(klines))
	}
	
	kline := klines[0]
	expectedOpen := 45000.00
	if kline.Open != expectedOpen {
		t.Errorf("Expected open %f, got %f", expectedOpen, kline.Open)
	}
	
	expectedHigh := 45100.00
	if kline.High != expectedHigh {
		t.Errorf("Expected high %f, got %f", expectedHigh, kline.High)
	}
	
	expectedLow := 44900.00
	if kline.Low != expectedLow {
		t.Errorf("Expected low %f, got %f", expectedLow, kline.Low)
	}
	
	expectedClose := 45050.00
	if kline.Close != expectedClose {
		t.Errorf("Expected close %f, got %f", expectedClose, kline.Close)
	}
	
	expectedVolume := 100.5
	if kline.Volume != expectedVolume {
		t.Errorf("Expected volume %f, got %f", expectedVolume, kline.Volume)
	}
}

func TestMarketDataServiceParseKlinesDataInvalid(t *testing.T) {
	service := &MarketDataService{}
	
	tests := []struct {
		name      string
		rawKlines [][]interface{}
	}{
		{
			name: "Insufficient fields",
			rawKlines: [][]interface{}{
				{1640995200000.0, "45000.00", "45100.00"}, // Only 3 fields
			},
		},
		{
			name: "Invalid open time",
			rawKlines: [][]interface{}{
				{"invalid_time", "45000.00", "45100.00", "44900.00", "45050.00", "100.5", 1640995259999.0},
			},
		},
		{
			name: "Invalid close time",
			rawKlines: [][]interface{}{
				{1640995200000.0, "45000.00", "45100.00", "44900.00", "45050.00", "100.5", "invalid_time"},
			},
		},
		{
			name: "Invalid open price",
			rawKlines: [][]interface{}{
				{1640995200000.0, "invalid_price", "45100.00", "44900.00", "45050.00", "100.5", 1640995259999.0},
			},
		},
		{
			name: "Invalid volume",
			rawKlines: [][]interface{}{
				{1640995200000.0, "45000.00", "45100.00", "44900.00", "45050.00", "invalid_volume", 1640995259999.0},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.parseKlinesData(tt.rawKlines)
			if err == nil {
				t.Error("Expected error for invalid klines data")
			}
		})
	}
}

func TestMarketDataServiceParseDataObject(t *testing.T) {
	service := &MarketDataService{}
	
	// Test data that can be marshaled/unmarshaled
	inputData := map[string]interface{}{
		"symbol": "BTC-USDT",
		"price":  "45000.00",
	}
	
	var result map[string]interface{}
	err := service.parseDataObject(inputData, &result)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	if result["symbol"] != "BTC-USDT" {
		t.Errorf("Expected symbol BTC-USDT, got %v", result["symbol"])
	}
	
	if result["price"] != "45000.00" {
		t.Errorf("Expected price 45000.00, got %v", result["price"])
	}
}

func TestMarketDataServiceParseDataArray(t *testing.T) {
	service := &MarketDataService{}
	
	// Test array data
	inputData := []interface{}{
		map[string]interface{}{"symbol": "BTC-USDT"},
		map[string]interface{}{"symbol": "ETH-USDT"},
	}
	
	var result []map[string]interface{}
	err := service.parseDataArray(inputData, &result)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}
	
	if result[0]["symbol"] != "BTC-USDT" {
		t.Errorf("Expected first symbol BTC-USDT, got %v", result[0]["symbol"])
	}
	
	if result[1]["symbol"] != "ETH-USDT" {
		t.Errorf("Expected second symbol ETH-USDT, got %v", result[1]["symbol"])
	}
}

// Test GetKlines parameter validation
func TestMarketDataServiceGetKlinesValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetKlines(ctx, "", "5m", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// Test empty interval
	_, err = service.GetKlines(ctx, "BTC-USDT", "", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for empty interval")
	}
	
	// Test invalid interval
	_, err = service.GetKlines(ctx, "BTC-USDT", "invalid", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for invalid interval")
	}
}

// Test GetFuturesKlines parameter validation
func TestMarketDataServiceGetFuturesKlinesValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetFuturesKlines(ctx, "", "5m", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// Test empty interval
	_, err = service.GetFuturesKlines(ctx, "SOL-USDT", "", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for empty interval")
	}
	
	// Test invalid interval
	_, err = service.GetFuturesKlines(ctx, "SOL-USDT", "invalid", 100, nil, nil)
	if err == nil {
		t.Error("Expected error for invalid interval")
	}
}

// Test GetPrice parameter validation
func TestMarketDataServiceGetPriceValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetPrice(ctx, "")
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}

// Test GetFuturesPrice parameter validation
func TestMarketDataServiceGetFuturesPriceValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetFuturesPrice(ctx, "")
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}

// Test GetTicker24hr parameter validation
func TestMarketDataServiceGetTicker24hrValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetTicker24hr(ctx, "")
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}

// Test GetFuturesTicker24hr parameter validation
func TestMarketDataServiceGetFuturesTicker24hrValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	ctx := context.Background()
	
	// Test empty symbol
	_, err = service.GetFuturesTicker24hr(ctx, "")
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
}
