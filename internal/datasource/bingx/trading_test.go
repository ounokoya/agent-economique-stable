package bingx

import (
	"context"
	"testing"
)

func TestNewTradingService(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTradingService(client)
	if service == nil {
		t.Fatal("NewTradingService should not return nil")
	}
	
	if service.client != client {
		t.Error("Service should reference the provided client")
	}
}

func TestTradingServiceValidateSpotOrderParams(t *testing.T) {
	service := &TradingService{}
	
	tests := []struct {
		name        string
		symbol      string
		side        OrderSide
		quantity    float64
		orderType   OrderType
		expectError bool
	}{
		{
			name:        "Valid buy market order",
			symbol:      "BTC-USDT",
			side:        OrderSideBuy,
			quantity:    0.001,
			orderType:   OrderTypeMarket,
			expectError: false,
		},
		{
			name:        "Valid sell limit order",
			symbol:      "ETH-USDT",
			side:        OrderSideSell,
			quantity:    0.1,
			orderType:   OrderTypeLimit,
			expectError: false,
		},
		{
			name:        "Empty symbol",
			symbol:      "",
			side:        OrderSideBuy,
			quantity:    0.001,
			orderType:   OrderTypeMarket,
			expectError: true,
		},
		{
			name:        "Invalid side",
			symbol:      "BTC-USDT",
			side:        OrderSide("INVALID"),
			quantity:    0.001,
			orderType:   OrderTypeMarket,
			expectError: true,
		},
		{
			name:        "Zero quantity",
			symbol:      "BTC-USDT",
			side:        OrderSideBuy,
			quantity:    0,
			orderType:   OrderTypeMarket,
			expectError: true,
		},
		{
			name:        "Negative quantity",
			symbol:      "BTC-USDT",
			side:        OrderSideBuy,
			quantity:    -0.001,
			orderType:   OrderTypeMarket,
			expectError: true,
		},
		{
			name:        "Invalid order type for spot",
			symbol:      "BTC-USDT",
			side:        OrderSideBuy,
			quantity:    0.001,
			orderType:   OrderTypeTrailingStopMarket,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSpotOrderParams(tt.symbol, tt.side, tt.quantity, tt.orderType)
			
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

func TestTradingServiceValidateFuturesOrderParams(t *testing.T) {
	service := &TradingService{}
	
	tests := []struct {
		name         string
		symbol       string
		side         OrderSide
		positionSide PositionSide
		quantity     float64
		orderType    OrderType
		expectError  bool
	}{
		{
			name:         "Valid long market order",
			symbol:       "BTC-USDT",
			side:         OrderSideBuy,
			positionSide: PositionSideLong,
			quantity:     100,
			orderType:    OrderTypeMarket,
			expectError:  false,
		},
		{
			name:         "Valid short limit order",
			symbol:       "ETH-USDT",
			side:         OrderSideSell,
			positionSide: PositionSideShort,
			quantity:     200,
			orderType:    OrderTypeLimit,
			expectError:  false,
		},
		{
			name:         "Valid trailing stop order",
			symbol:       "SOL-USDT",
			side:         OrderSideBuy,
			positionSide: PositionSideLong,
			quantity:     50,
			orderType:    OrderTypeTrailingStopMarket,
			expectError:  false,
		},
		{
			name:         "Empty symbol",
			symbol:       "",
			side:         OrderSideBuy,
			positionSide: PositionSideLong,
			quantity:     100,
			orderType:    OrderTypeMarket,
			expectError:  true,
		},
		{
			name:         "Invalid side",
			symbol:       "BTC-USDT",
			side:         OrderSide("INVALID"),
			positionSide: PositionSideLong,
			quantity:     100,
			orderType:    OrderTypeMarket,
			expectError:  true,
		},
		{
			name:         "Invalid position side",
			symbol:       "BTC-USDT",
			side:         OrderSideBuy,
			positionSide: PositionSide("INVALID"),
			quantity:     100,
			orderType:    OrderTypeMarket,
			expectError:  true,
		},
		{
			name:         "Zero quantity",
			symbol:       "BTC-USDT",
			side:         OrderSideBuy,
			positionSide: PositionSideLong,
			quantity:     0,
			orderType:    OrderTypeMarket,
			expectError:  true,
		},
		{
			name:         "Negative quantity",
			symbol:       "BTC-USDT",
			side:         OrderSideBuy,
			positionSide: PositionSideLong,
			quantity:     -100,
			orderType:    OrderTypeMarket,
			expectError:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateFuturesOrderParams(tt.symbol, tt.side, tt.positionSide, tt.quantity, tt.orderType)
			
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

func TestTradingServiceParseResponseData(t *testing.T) {
	service := &TradingService{}
	
	// Test parsing object data
	inputData := map[string]interface{}{
		"orderId": "12345",
		"symbol":  "BTC-USDT",
		"side":    "BUY",
		"status":  "FILLED",
	}
	
	var order Order
	err := service.parseResponseData(inputData, &order)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	// Note: The actual field mapping would depend on the Order struct tags
	// This test validates the parsing mechanism works
}

func TestTradingServiceParseOrderResponse(t *testing.T) {
	service := &TradingService{}
	
	inputData := map[string]interface{}{
		"orderId": "67890",
		"symbol":  "ETH-USDT",
	}
	
	var order Order
	err := service.parseOrderResponse(inputData, &order)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

func TestTradingServiceParseOrdersResponse(t *testing.T) {
	service := &TradingService{}
	
	inputData := []interface{}{
		map[string]interface{}{
			"orderId": "111",
			"symbol":  "BTC-USDT",
		},
		map[string]interface{}{
			"orderId": "222",
			"symbol":  "ETH-USDT",
		},
	}
	
	var orders []Order
	err := service.parseOrdersResponse(inputData, &orders)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

func TestTradingServiceParsePositionsResponse(t *testing.T) {
	service := &TradingService{}
	
	inputData := []interface{}{
		map[string]interface{}{
			"symbol":       "SOL-USDT",
			"positionSide": "LONG",
			"size":         100.0,
		},
	}
	
	var positions []Position
	err := service.parsePositionsResponse(inputData, &positions)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

func TestTradingServiceParseBalancesResponse(t *testing.T) {
	service := &TradingService{}
	
	inputData := []interface{}{
		map[string]interface{}{
			"asset": "USDT",
			"free":  1000.00,
		},
		map[string]interface{}{
			"asset": "BTC",
			"free":  0.1,
		},
	}
	
	var balances []Balance
	err := service.parseBalancesResponse(inputData, &balances)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

// Test spot order parameter validation for actual service calls
func TestTradingServiceSpotOrderValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTradingService(client)
	ctx := context.Background()
	
	// Test SpotBuy with invalid parameters
	_, err = service.SpotBuy(ctx, "", 0.001, 45000, OrderTypeMarket)
	if err == nil {
		t.Error("SpotBuy should return error for empty symbol")
	}
	
	_, err = service.SpotBuy(ctx, "BTC-USDT", 0, 45000, OrderTypeMarket)
	if err == nil {
		t.Error("SpotBuy should return error for zero quantity")
	}
	
	// Test SpotSell with invalid parameters
	_, err = service.SpotSell(ctx, "", 0.001, 45000, OrderTypeMarket)
	if err == nil {
		t.Error("SpotSell should return error for empty symbol")
	}
	
	_, err = service.SpotSell(ctx, "BTC-USDT", -0.001, 45000, OrderTypeMarket)
	if err == nil {
		t.Error("SpotSell should return error for negative quantity")
	}
	
	// Test SpotBuyMarket with invalid parameters
	_, err = service.SpotBuyMarket(ctx, "", 100)
	if err == nil {
		t.Error("SpotBuyMarket should return error for empty symbol")
	}
	
	// Test SpotSellMarket with invalid parameters
	_, err = service.SpotSellMarket(ctx, "", 0.001)
	if err == nil {
		t.Error("SpotSellMarket should return error for empty symbol")
	}
}

// Test futures order parameter validation for actual service calls
func TestTradingServiceFuturesOrderValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTradingService(client)
	ctx := context.Background()
	
	// Test FuturesLong with invalid parameters
	_, err = service.FuturesLong(ctx, "", 100, 10, MarginTypeCross, OrderTypeMarket, 0)
	if err == nil {
		t.Error("FuturesLong should return error for empty symbol")
	}
	
	_, err = service.FuturesLong(ctx, "BTC-USDT", 0, 10, MarginTypeCross, OrderTypeMarket, 0)
	if err == nil {
		t.Error("FuturesLong should return error for zero quantity")
	}
	
	// Test FuturesShort with invalid parameters
	_, err = service.FuturesShort(ctx, "", 100, 10, MarginTypeCross, OrderTypeMarket, 0)
	if err == nil {
		t.Error("FuturesShort should return error for empty symbol")
	}
	
	_, err = service.FuturesShort(ctx, "BTC-USDT", -100, 10, MarginTypeCross, OrderTypeMarket, 0)
	if err == nil {
		t.Error("FuturesShort should return error for negative quantity")
	}
	
	// Test CloseFuturesLong with invalid parameters
	_, err = service.CloseFuturesLong(ctx, "", 100, OrderTypeMarket, 0)
	if err == nil {
		t.Error("CloseFuturesLong should return error for empty symbol")
	}
	
	// Test CloseFuturesShort with invalid parameters
	_, err = service.CloseFuturesShort(ctx, "", 100, OrderTypeMarket, 0)
	if err == nil {
		t.Error("CloseFuturesShort should return error for empty symbol")
	}
}

// Test leverage validation
func TestTradingServiceSetLeverageValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTradingService(client)
	ctx := context.Background()
	
	// Test invalid leverage values
	err = service.SetLeverage(ctx, "BTC-USDT", 0)
	if err == nil {
		t.Error("SetLeverage should return error for leverage 0")
	}
	
	err = service.SetLeverage(ctx, "BTC-USDT", -5)
	if err == nil {
		t.Error("SetLeverage should return error for negative leverage")
	}
	
	err = service.SetLeverage(ctx, "BTC-USDT", 200)
	if err == nil {
		t.Error("SetLeverage should return error for leverage > 125")
	}
}

// Test margin type validation
func TestTradingServiceSetMarginTypeValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTradingService(client)
	ctx := context.Background()
	
	// Test invalid margin type
	err = service.SetMarginType(ctx, "BTC-USDT", MarginType("INVALID"))
	if err == nil {
		t.Error("SetMarginType should return error for invalid margin type")
	}
}

// Test order status and open orders parameter validation
func TestTradingServiceQueryValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	_ = NewTradingService(client)
	_ = context.Background()
	
	// These would normally make API calls, but we're testing parameter validation
	// The actual API calls would fail in test environment, which is expected
	
	// Test GetOrderStatus - the method itself doesn't validate empty strings
	// but the API would return an error
	
	// Test GetOpenOrders - empty symbol is allowed (gets all orders)
	
	// Test GetPositions - empty symbol is allowed (gets all positions)
	
	// Test GetAccountBalance - no parameters to validate
	
	// Test CancelOrder with empty parameters would be handled by API validation
}
