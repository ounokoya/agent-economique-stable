package bingx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// TradingService provides trading operations for BingX Spot and Futures
type TradingService struct {
	client *Client
}

// NewTradingService creates a new trading service
func NewTradingService(client *Client) *TradingService {
	return &TradingService{
		client: client,
	}
}

// SpotBuy executes a spot buy order (acheter)
func (t *TradingService) SpotBuy(ctx context.Context, symbol string, quantity, price float64, orderType OrderType) (*Order, error) {
	return t.placeSpotOrder(ctx, symbol, OrderSideBuy, quantity, price, orderType)
}

// SpotSell executes a spot sell order (vendre)  
func (t *TradingService) SpotSell(ctx context.Context, symbol string, quantity, price float64, orderType OrderType) (*Order, error) {
	return t.placeSpotOrder(ctx, symbol, OrderSideSell, quantity, price, orderType)
}

// SpotBuyMarket executes a market buy order with USDT amount
func (t *TradingService) SpotBuyMarket(ctx context.Context, symbol string, quoteAmount float64) (*Order, error) {
	params := map[string]string{
		"symbol":         symbol,
		"side":           string(OrderSideBuy),
		"type":           string(OrderTypeMarket),
		"quoteOrderQty":  strconv.FormatFloat(quoteAmount, 'f', -1, 64),
	}

	return t.executeSpotOrder(ctx, params)
}

// SpotSellMarket executes a market sell order with crypto quantity
func (t *TradingService) SpotSellMarket(ctx context.Context, symbol string, quantity float64) (*Order, error) {
	params := map[string]string{
		"symbol":   symbol,
		"side":     string(OrderSideSell),
		"type":     string(OrderTypeMarket),
		"quantity": strconv.FormatFloat(quantity, 'f', -1, 64),
	}

	return t.executeSpotOrder(ctx, params)
}

// placeSpotOrder is a helper function for spot orders
func (t *TradingService) placeSpotOrder(ctx context.Context, symbol string, side OrderSide, quantity, price float64, orderType OrderType) (*Order, error) {
	if err := t.validateSpotOrderParams(symbol, side, quantity, orderType); err != nil {
		return nil, err
	}

	params := map[string]string{
		"symbol":   symbol,
		"side":     string(side),
		"type":     string(orderType),
		"quantity": strconv.FormatFloat(quantity, 'f', -1, 64),
	}

	if orderType == OrderTypeLimit && price > 0 {
		params["price"] = strconv.FormatFloat(price, 'f', -1, 64)
		params["timeInForce"] = "GTC"
	}

	return t.executeSpotOrder(ctx, params)
}

// executeSpotOrder executes the spot order with BingX API
func (t *TradingService) executeSpotOrder(ctx context.Context, params map[string]string) (*Order, error) {
	resp, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/spot/v1/trade/order", params, EndpointTypeTrading)
	if err != nil {
		return nil, fmt.Errorf("failed to place spot order: %w", err)
	}

	var order Order
	if err := t.parseOrderResponse(resp.Data, &order); err != nil {
		return nil, fmt.Errorf("failed to parse spot order response: %w", err)
	}

	return &order, nil
}

// FuturesLong opens a long position (ouvrir position long)
func (t *TradingService) FuturesLong(ctx context.Context, symbol string, quantity float64, leverage int, marginType MarginType, orderType OrderType, price float64) (*Order, error) {
	return t.placeFuturesOrder(ctx, symbol, OrderSideBuy, PositionSideLong, quantity, leverage, marginType, orderType, price)
}

// FuturesShort opens a short position (ouvrir position short)
func (t *TradingService) FuturesShort(ctx context.Context, symbol string, quantity float64, leverage int, marginType MarginType, orderType OrderType, price float64) (*Order, error) {
	return t.placeFuturesOrder(ctx, symbol, OrderSideSell, PositionSideShort, quantity, leverage, marginType, orderType, price)
}

// CloseFuturesLong closes a long position (fermer position long)
func (t *TradingService) CloseFuturesLong(ctx context.Context, symbol string, quantity float64, orderType OrderType, price float64) (*Order, error) {
	return t.placeFuturesOrder(ctx, symbol, OrderSideSell, PositionSideLong, quantity, 0, "", orderType, price)
}

// CloseFuturesShort closes a short position (fermer position short)
func (t *TradingService) CloseFuturesShort(ctx context.Context, symbol string, quantity float64, orderType OrderType, price float64) (*Order, error) {
	return t.placeFuturesOrder(ctx, symbol, OrderSideBuy, PositionSideShort, quantity, 0, "", orderType, price)
}

// placeFuturesOrder is a helper function for futures orders
func (t *TradingService) placeFuturesOrder(ctx context.Context, symbol string, side OrderSide, positionSide PositionSide, quantity float64, leverage int, marginType MarginType, orderType OrderType, price float64) (*Order, error) {
	if err := t.validateFuturesOrderParams(symbol, side, positionSide, quantity, orderType); err != nil {
		return nil, err
	}

	// Set leverage if specified
	if leverage > 0 {
		if err := t.SetLeverage(ctx, symbol, leverage); err != nil {
			return nil, fmt.Errorf("failed to set leverage: %w", err)
		}
	}

	// Set margin type if specified
	if marginType != "" {
		if err := t.SetMarginType(ctx, symbol, marginType); err != nil {
			return nil, fmt.Errorf("failed to set margin type: %w", err)
		}
	}

	params := map[string]string{
		"symbol":       symbol,
		"side":         string(side),
		"positionSide": string(positionSide),
		"type":         string(orderType),
		"quantity":     strconv.FormatFloat(quantity, 'f', -1, 64),
	}

	if orderType == OrderTypeLimit && price > 0 {
		params["price"] = strconv.FormatFloat(price, 'f', -1, 64)
		params["timeInForce"] = "GTC"
	}

	return t.executeFuturesOrder(ctx, params)
}

// executeFuturesOrder executes the futures order with BingX API
func (t *TradingService) executeFuturesOrder(ctx context.Context, params map[string]string) (*Order, error) {
	resp, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/swap/v2/trade/order", params, EndpointTypeTrading)
	if err != nil {
		return nil, fmt.Errorf("failed to place futures order: %w", err)
	}

	var order Order
	if err := t.parseOrderResponse(resp.Data, &order); err != nil {
		return nil, fmt.Errorf("failed to parse futures order response: %w", err)
	}

	return &order, nil
}

// SetLeverage configures leverage for a futures symbol
func (t *TradingService) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	if leverage < 1 || leverage > 125 {
		return fmt.Errorf("invalid leverage: %d (must be between 1 and 125)", leverage)
	}

	params := map[string]string{
		"symbol":   symbol,
		"leverage": strconv.Itoa(leverage),
	}

	_, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/swap/v2/trade/leverage", params, EndpointTypeTrading)
	if err != nil {
		return fmt.Errorf("failed to set leverage for %s: %w", symbol, err)
	}

	return nil
}

// SetMarginType configures margin type for a futures symbol
func (t *TradingService) SetMarginType(ctx context.Context, symbol string, marginType MarginType) error {
	if marginType != MarginTypeCross && marginType != MarginTypeIsolated {
		return fmt.Errorf("invalid margin type: %s", marginType)
	}

	params := map[string]string{
		"symbol":     symbol,
		"marginType": string(marginType),
	}

	_, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/swap/v2/trade/marginType", params, EndpointTypeTrading)
	if err != nil {
		return fmt.Errorf("failed to set margin type for %s: %w", symbol, err)
	}

	return nil
}

// CancelOrder cancels an existing order
func (t *TradingService) CancelOrder(ctx context.Context, symbol, orderID string, isSpot bool) error {
	params := map[string]string{
		"symbol":  symbol,
		"orderId": orderID,
	}

	endpoint := "/openApi/swap/v2/trade/cancel"
	if isSpot {
		endpoint = "/openApi/spot/v1/trade/cancel"
	}

	_, err := t.client.DoRequest(ctx, http.MethodDelete, endpoint, params, EndpointTypeTrading)
	if err != nil {
		return fmt.Errorf("failed to cancel order %s: %w", orderID, err)
	}

	return nil
}

// GetOrderStatus retrieves the status of an order
func (t *TradingService) GetOrderStatus(ctx context.Context, symbol, orderID string, isSpot bool) (*Order, error) {
	params := map[string]string{
		"symbol":  symbol,
		"orderId": orderID,
	}

	endpoint := "/openApi/swap/v2/trade/query"
	if isSpot {
		endpoint = "/openApi/spot/v1/trade/query"
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, endpoint, params, EndpointTypeTrading)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status: %w", err)
	}

	var order Order
	if err := t.parseOrderResponse(resp.Data, &order); err != nil {
		return nil, fmt.Errorf("failed to parse order status response: %w", err)
	}

	return &order, nil
}

// GetOpenOrders retrieves all open orders
func (t *TradingService) GetOpenOrders(ctx context.Context, symbol string, isSpot bool) ([]Order, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = symbol
	}

	endpoint := "/openApi/swap/v2/trade/openOrders"
	if isSpot {
		endpoint = "/openApi/spot/v1/trade/openOrders"
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, endpoint, params, EndpointTypeTrading)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	var orders []Order
	if err := t.parseOrdersResponse(resp.Data, &orders); err != nil {
		return nil, fmt.Errorf("failed to parse open orders response: %w", err)
	}

	return orders, nil
}

// GetPositions retrieves all open futures positions
func (t *TradingService) GetPositions(ctx context.Context, symbol string) ([]Position, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = symbol
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, "/openApi/swap/v2/user/positions", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var positions []Position
	if err := t.parsePositionsResponse(resp.Data, &positions); err != nil {
		return nil, fmt.Errorf("failed to parse positions response: %w", err)
	}

	return positions, nil
}

// GetAccountBalance retrieves account balance
func (t *TradingService) GetAccountBalance(ctx context.Context, isSpot bool) ([]Balance, error) {
	endpoint := "/openApi/swap/v2/user/balance"
	if isSpot {
		endpoint = "/openApi/spot/v1/account/balance"
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, endpoint, nil, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}

	var balances []Balance
	if err := t.parseBalancesResponse(resp.Data, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse balance response: %w", err)
	}

	return balances, nil
}

// validateSpotOrderParams validates parameters for spot orders
func (t *TradingService) validateSpotOrderParams(symbol string, side OrderSide, quantity float64, orderType OrderType) error {
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if side != OrderSideBuy && side != OrderSideSell {
		return fmt.Errorf("invalid order side: %s", side)
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %f", quantity)
	}

	if orderType != OrderTypeMarket && orderType != OrderTypeLimit {
		return fmt.Errorf("invalid order type for spot: %s", orderType)
	}

	return nil
}

// validateFuturesOrderParams validates parameters for futures orders
func (t *TradingService) validateFuturesOrderParams(symbol string, side OrderSide, positionSide PositionSide, quantity float64, orderType OrderType) error {
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if side != OrderSideBuy && side != OrderSideSell {
		return fmt.Errorf("invalid order side: %s", side)
	}

	if positionSide != PositionSideLong && positionSide != PositionSideShort {
		return fmt.Errorf("invalid position side: %s", positionSide)
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %f", quantity)
	}

	validOrderTypes := []OrderType{OrderTypeMarket, OrderTypeLimit, OrderTypeTrailingStopMarket}
	isValid := false
	for _, validType := range validOrderTypes {
		if orderType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid order type for futures: %s", orderType)
	}

	return nil
}

// parseOrderResponse parses order response from BingX API
func (t *TradingService) parseOrderResponse(data interface{}, order *Order) error {
	return t.parseResponseData(data, order)
}

// parseOrdersResponse parses multiple orders response from BingX API
func (t *TradingService) parseOrdersResponse(data interface{}, orders *[]Order) error {
	return t.parseResponseData(data, orders)
}

// parsePositionsResponse parses positions response from BingX API
func (t *TradingService) parsePositionsResponse(data interface{}, positions *[]Position) error {
	return t.parseResponseData(data, positions)
}

// parseBalancesResponse parses balances response from BingX API
func (t *TradingService) parseBalancesResponse(data interface{}, balances *[]Balance) error {
	return t.parseResponseData(data, balances)
}

// parseResponseData is a generic response parser
func (t *TradingService) parseResponseData(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}
