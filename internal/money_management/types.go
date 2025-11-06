package money_management

import (
	"errors"
	"fmt"
	"time"
)

// Errors for Money Management BASE
var (
	ErrDailyLimitBreached     = errors.New("daily loss limit breached")
	ErrMonthlyLimitBreached   = errors.New("monthly loss limit breached")
	ErrInsufficientBalance    = errors.New("insufficient balance for position sizing")
	ErrInvalidConfiguration   = errors.New("invalid money management configuration")
	ErrCircuitBreakerActive   = errors.New("circuit breaker is active - trading halted")
	ErrInvalidPositionAmount  = errors.New("invalid position amount")
	ErrInvalidPriceData      = errors.New("invalid price data for calculations")
)

// CircuitBreakerState represents the current state of circuit breakers
type CircuitBreakerState struct {
	DailyBreakerActive   bool      `json:"daily_breaker_active"`
	MonthlyBreakerActive bool      `json:"monthly_breaker_active"`
	LastDailyBreach      time.Time `json:"last_daily_breach"`
	LastMonthlyBreach    time.Time `json:"last_monthly_breach"`
	DailyResetTime       time.Time `json:"daily_reset_time"`     // Next 00:00 UTC
	MonthlyRetryTime     time.Time `json:"monthly_retry_time"`   // Next 00:00 UTC after monthly breach
	TotalBreaches        int       `json:"total_breaches"`
}

// PositionSizingMode defines how position sizes are calculated
type PositionSizingMode string

const (
	PositionSizingFixed      PositionSizingMode = "fixed"      // Fixed USDT amounts
	PositionSizingPercentage PositionSizingMode = "percentage" // Percentage of capital
)

// PositionSizingConfig holds position sizing configuration
type PositionSizingConfig struct {
	Mode PositionSizingMode `json:"mode"` // "fixed" or "percentage"
	
	// For Fixed Mode
	SpotAmountUSDT    float64 `json:"spot_amount_usdt"`    // Fixed USDT per spot trade
	FuturesAmountUSDT float64 `json:"futures_amount_usdt"` // Fixed USDT per futures trade
	
	// For Percentage Mode  
	SpotPercentage    float64 `json:"spot_percentage"`    // % of capital per spot trade
	FuturesPercentage float64 `json:"futures_percentage"` // % of capital per futures trade
	
	// Common
	DefaultLeverage int `json:"default_leverage"` // Default leverage for futures
	MaxPositionSize float64 `json:"max_position_size"` // Maximum position size (USDT)
	MinPositionSize float64 `json:"min_position_size"` // Minimum position size (USDT)
}

// BaseConfiguration holds core MM configuration (invariant across strategies)
type BaseConfiguration struct {
	// Circuit Breaker Limits (configurable)
	DailyLimitPercent   float64 `json:"daily_limit_percent"`   // % daily loss limit
	MonthlyLimitPercent float64 `json:"monthly_limit_percent"` // % monthly loss limit
	
	// Position Sizing (flexible configuration)
	PositionSizing PositionSizingConfig `json:"position_sizing"`
	
	// Capital Management
	StartingCapital float64 `json:"starting_capital"` // Initial capital for percentage calculations
	CurrentCapital  float64 `json:"current_capital"`  // Current capital (for percentage mode)
	
	// Risk Controls
	MaxDailyPositions   int     `json:"max_daily_positions"`   // Max positions per day
	MaxConcurrentTrades int     `json:"max_concurrent_trades"` // Max concurrent positions
	MaxRiskPerTrade     float64 `json:"max_risk_per_trade"`    // Max % risk per single trade
	
	// Monitoring
	MetricsUpdateIntervalSeconds int `json:"metrics_update_interval_seconds"` // Update frequency
	ReportGenerationHour         int `json:"report_generation_hour"`          // Report generation hour UTC
	
	// Emergency Settings
	EmergencyStopEnabled      bool    `json:"emergency_stop_enabled"`       // Enable emergency stops
	ForceStopLossPercent      float64 `json:"force_stop_loss_percent"`      // Force stop loss %
	TradingHaltDurationHours  int     `json:"trading_halt_duration_hours"`  // Hours to halt after breach
}

// GlobalMetrics represents cross-strategy performance metrics
type GlobalMetrics struct {
	// Time tracking
	LastUpdateTime time.Time `json:"last_update_time"`
	DayStartTime   time.Time `json:"day_start_time"`
	
	// Performance metrics  
	DailyPnL     float64 `json:"daily_pnl"`      // PnL since day start
	MonthlyPnL   float64 `json:"monthly_pnl"`    // PnL last 30 days
	TotalPnL     float64 `json:"total_pnl"`      // All-time PnL
	
	// Position statistics
	TotalPositions       int `json:"total_positions"`
	WinningPositions     int `json:"winning_positions"`
	LosingPositions      int `json:"losing_positions"`
	CurrentOpenPositions int `json:"current_open_positions"`
	
	// Performance ratios
	WinRate      float64 `json:"win_rate"`       // % winning positions
	ProfitFactor float64 `json:"profit_factor"`  // Total wins / Total losses
	
	// Risk metrics
	MaxDrawdown        float64 `json:"max_drawdown"`         // Worst peak-to-trough %
	DailyLossPercent   float64 `json:"daily_loss_percent"`   // Current daily loss %
	MonthlyLossPercent float64 `json:"monthly_loss_percent"` // Current monthly loss %
	
	// Strategy breakdown
	StrategyMetrics map[string]StrategyMetrics `json:"strategy_metrics"`
}

// StrategyMetrics represents performance metrics for individual strategy
type StrategyMetrics struct {
	StrategyName string `json:"strategy_name"`
	
	// Performance
	PnL              float64 `json:"pnl"`
	Positions        int     `json:"positions"`
	WinningPositions int     `json:"winning_positions"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	
	// Risk
	MaxDrawdown float64 `json:"max_drawdown"`
	
	// Activity
	LastTradeTime time.Time `json:"last_trade_time"`
	IsActive      bool      `json:"is_active"`
}

// PositionSizingResult holds results of position sizing calculations
type PositionSizingResult struct {
	Quantity          float64 `json:"quantity"`           // Calculated quantity
	NotionalValue     float64 `json:"notional_value"`     // Total position value
	RequiredBalance   float64 `json:"required_balance"`   // Balance needed
	Precision         int     `json:"precision"`          // Decimal places for symbol
	MinimumQuantity   float64 `json:"minimum_quantity"`   // Exchange minimum
	IsValid           bool    `json:"is_valid"`           // Whether calculation is valid
	ValidationError   string  `json:"validation_error"`   // Error if invalid
}

// PositionSizingRequest holds parameters for position sizing
type PositionSizingRequest struct {
	Symbol        string  `json:"symbol"`         // Trading symbol (BTC-USDT)
	Price         float64 `json:"price"`          // Current market price
	PositionType  string  `json:"position_type"`  // "spot" or "futures"
	Leverage      int     `json:"leverage"`       // For futures (ignored for spot)
	AvailableBalance float64 `json:"available_balance"` // Current available balance
}

// AuditLogEntry represents a single audit trail entry
type AuditLogEntry struct {
	Timestamp   time.Time                `json:"timestamp"`
	EventType   string                   `json:"event_type"`   // "CIRCUIT_BREAKER", "POSITION_SIZE", "CONFIG_CHANGE"
	EventData   map[string]interface{}   `json:"event_data"`
	Success     bool                     `json:"success"`
	ErrorMsg    string                   `json:"error_msg,omitempty"`
	Impact      string                   `json:"impact"`       // Description of impact
}

// DailyReport represents end-of-day summary
type DailyReport struct {
	Date                 time.Time                      `json:"date"`
	GlobalMetrics        GlobalMetrics                  `json:"global_metrics"`
	CircuitBreakerState  CircuitBreakerState           `json:"circuit_breaker_state"`
	StrategyPerformance  map[string]StrategyMetrics    `json:"strategy_performance"`
	RiskAnalysis         RiskAnalysis                   `json:"risk_analysis"`
	Recommendations      []string                       `json:"recommendations"`
}

// RiskAnalysis provides risk assessment
type RiskAnalysis struct {
	DistanceToDailyLimit   float64 `json:"distance_to_daily_limit"`   // % away from -5%
	DistanceToMonthlyLimit float64 `json:"distance_to_monthly_limit"` // % away from -15%
	RiskLevel              string  `json:"risk_level"`                // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	AlertsGenerated        []string `json:"alerts_generated"`
}

// DefaultBaseConfiguration returns sensible defaults for BASE MM
func DefaultBaseConfiguration() BaseConfiguration {
	return BaseConfiguration{
		// Circuit Breaker defaults
		DailyLimitPercent:   5.0,  // -5% daily limit
		MonthlyLimitPercent: 15.0, // -15% monthly limit
		
		// Position Sizing defaults (Fixed mode)
		PositionSizing: PositionSizingConfig{
			Mode:              PositionSizingFixed,
			SpotAmountUSDT:    1000.0, // 1000 USDT per spot trade
			FuturesAmountUSDT: 500.0,  // 500 USDT per futures trade
			SpotPercentage:    10.0,   // 10% of capital per spot trade (if percentage mode)
			FuturesPercentage: 5.0,    // 5% of capital per futures trade (if percentage mode)
			DefaultLeverage:   10,     // 10x leverage default
			MaxPositionSize:   5000.0, // Max 5000 USDT per position
			MinPositionSize:   10.0,   // Min 10 USDT per position
		},
		
		// Capital defaults
		StartingCapital: 10000.0, // Default starting capital
		CurrentCapital:  10000.0, // Initialize current = starting
		
		// Risk Control defaults
		MaxDailyPositions:   50,   // Max 50 positions per day
		MaxConcurrentTrades: 10,   // Max 10 concurrent positions
		MaxRiskPerTrade:     2.0,  // Max 2% risk per trade
		
		// Monitoring defaults
		MetricsUpdateIntervalSeconds: 1,  // Update every second
		ReportGenerationHour:         23, // Generate reports at 23:59 UTC
		
		// Emergency defaults
		EmergencyStopEnabled:     true, // Enable emergency stops
		ForceStopLossPercent:     10.0, // Force stop at -10% position loss
		TradingHaltDurationHours: 24,   // Halt trading for 24h after breach
	}
}

// DefaultPercentageConfiguration returns configuration using percentage mode
func DefaultPercentageConfiguration(startingCapital float64) BaseConfiguration {
	config := DefaultBaseConfiguration()
	config.PositionSizing.Mode = PositionSizingPercentage
	config.StartingCapital = startingCapital
	config.CurrentCapital = startingCapital
	return config
}

// IsCircuitBreakerActive checks if any circuit breaker is currently active
func (cbs CircuitBreakerState) IsActive() bool {
	return cbs.DailyBreakerActive || cbs.MonthlyBreakerActive
}

// ShouldResetDaily checks if daily circuit breaker should be reset (new day)
func (cbs CircuitBreakerState) ShouldResetDaily() bool {
	return time.Now().UTC().After(cbs.DailyResetTime)
}

// ShouldRetryMonthly checks if monthly circuit breaker should retry (next day after breach)
func (cbs CircuitBreakerState) ShouldRetryMonthly() bool {
	return cbs.MonthlyBreakerActive && time.Now().UTC().After(cbs.MonthlyRetryTime)
}

// Validate validates position sizing configuration
func (psc PositionSizingConfig) Validate() error {
	if psc.Mode != PositionSizingFixed && psc.Mode != PositionSizingPercentage {
		return errors.New("position sizing mode must be 'fixed' or 'percentage'")
	}
	
	if psc.Mode == PositionSizingFixed {
		if psc.SpotAmountUSDT <= 0 {
			return errors.New("spot amount USDT must be positive in fixed mode")
		}
		if psc.FuturesAmountUSDT <= 0 {
			return errors.New("futures amount USDT must be positive in fixed mode")
		}
	}
	
	if psc.Mode == PositionSizingPercentage {
		if psc.SpotPercentage <= 0 || psc.SpotPercentage > 100 {
			return errors.New("spot percentage must be between 0.1 and 100.0")
		}
		if psc.FuturesPercentage <= 0 || psc.FuturesPercentage > 100 {
			return errors.New("futures percentage must be between 0.1 and 100.0")
		}
	}
	
	if psc.DefaultLeverage < 1 || psc.DefaultLeverage > 125 {
		return errors.New("default leverage must be between 1 and 125")
	}
	if psc.MaxPositionSize <= psc.MinPositionSize {
		return errors.New("max position size must be greater than min position size")
	}
	if psc.MinPositionSize <= 0 {
		return errors.New("min position size must be positive")
	}
	
	return nil
}

// Validate validates base configuration
func (bc BaseConfiguration) Validate() error {
	// Circuit Breaker validation
	if bc.DailyLimitPercent <= 0 || bc.DailyLimitPercent > 50 {
		return errors.New("daily limit percent must be between 0.1 and 50.0")
	}
	if bc.MonthlyLimitPercent <= 0 || bc.MonthlyLimitPercent > 90 {
		return errors.New("monthly limit percent must be between 1.0 and 90.0")
	}
	
	// Position Sizing validation
	if err := bc.PositionSizing.Validate(); err != nil {
		return fmt.Errorf("position sizing validation failed: %w", err)
	}
	
	// Capital validation
	if bc.StartingCapital <= 0 {
		return errors.New("starting capital must be positive")
	}
	if bc.CurrentCapital <= 0 {
		return errors.New("current capital must be positive")
	}
	
	// Risk Controls validation
	if bc.MaxDailyPositions <= 0 {
		return errors.New("max daily positions must be positive")
	}
	if bc.MaxConcurrentTrades <= 0 {
		return errors.New("max concurrent trades must be positive")
	}
	if bc.MaxRiskPerTrade <= 0 || bc.MaxRiskPerTrade > 50 {
		return errors.New("max risk per trade must be between 0.1 and 50.0")
	}
	
	// Monitoring validation
	if bc.MetricsUpdateIntervalSeconds <= 0 {
		return errors.New("metrics update interval must be positive")
	}
	if bc.ReportGenerationHour < 0 || bc.ReportGenerationHour > 23 {
		return errors.New("report generation hour must be between 0 and 23")
	}
	
	// Emergency validation
	if bc.ForceStopLossPercent <= 0 || bc.ForceStopLossPercent > 100 {
		return errors.New("force stop loss percent must be between 0.1 and 100.0")
	}
	if bc.TradingHaltDurationHours <= 0 {
		return errors.New("trading halt duration must be positive")
	}
	
	return nil
}
