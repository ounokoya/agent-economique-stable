// Package bingx provides a complete SDK for BingX trading API
// Supporting Spot and Perpetual Futures with multi-account management
package bingx

import (
	"time"
)

// Environment represents the BingX API environment
type Environment string

const (
	// DemoEnvironment for testing with virtual funds
	DemoEnvironment Environment = "demo"
	// LiveEnvironment for real trading
	LiveEnvironment Environment = "live"
)

// URLs for different environments
const (
	DemoBaseURL = "https://open-api-vst.bingx.com"
	LiveBaseURL = "https://open-api.bingx.com"
)

// Rate limiting constants based on BingX documentation
const (
	MarketDataRateLimit = 10  // requests per second for market data
	TradingRateLimit    = 20  // requests per second for trading
	AccountRateLimit    = 10  // requests per second for account operations
	MaxWebSocketConns   = 10  // maximum WebSocket connections per IP
)

// APICredentials holds the authentication information
type APICredentials struct {
	APIKey    string
	SecretKey string
}

// ClientConfig represents the configuration for BingX client
type ClientConfig struct {
	Environment Environment
	Credentials APICredentials
	BaseURL     string
	Timeout     time.Duration
}

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents the type of an order
type OrderType string

const (
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

// PositionSide represents the position side for futures trading
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
	PositionSideBoth  PositionSide = "BOTH"
)

// MarginType represents the margin type for futures trading
type MarginType string

const (
	MarginTypeCross    MarginType = "CROSS"
	MarginTypeIsolated MarginType = "ISOLATED"
)

// Symbol represents a trading pair
type Symbol struct {
	Name     string
	BaseAsset string
	QuoteAsset string
	Status   string
}

// Kline represents a candlestick data point
type Kline struct {
	OpenTime  time.Time
	CloseTime time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// Ticker represents current price information
type Ticker struct {
	Symbol             string
	Price              float64
	PriceChangePercent float64
	Volume             float64
	Timestamp          time.Time
}

// Balance represents account balance information
type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

// Order represents a trading order
type Order struct {
	OrderID       string
	ClientOrderID string
	Symbol        string
	Side          OrderSide
	Type          OrderType
	Quantity      float64
	Price         float64
	Status        OrderStatus
	TimeInForce   string
	Timestamp     time.Time
}

// Position represents a futures position
type Position struct {
	Symbol           string
	PositionSide     PositionSide
	Size             float64
	EntryPrice       float64
	MarkPrice        float64
	UnrealizedPnL    float64
	Percentage       float64
	Leverage         int
	MarginType       MarginType
	LiquidationPrice float64
	Timestamp        time.Time
}

// SubAccount represents a sub-account information
type SubAccount struct {
	UID       int64
	Email     string
	Status    string
	CreatedAt time.Time
}

// Transfer represents an internal transfer between accounts
type Transfer struct {
	TransferID string
	FromUID    int64
	ToUID      int64
	Asset      string
	Amount     float64
	Status     string
	Timestamp  time.Time
}

// APIResponse represents the standard BingX API response structure
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Error implements the error interface for ErrorResponse
func (e ErrorResponse) Error() string {
	return e.Msg
}
