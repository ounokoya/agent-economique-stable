package macd_cci_dmi

import (
	"errors"
	"time"
)

// Errors for Behavioral Money Management
var (
	ErrInvalidTrailingConfig = errors.New("invalid trailing stop configuration")
	ErrInvalidProfitGrid     = errors.New("invalid profit adjustment grid")
	ErrPositionNotOpen       = errors.New("no open position for behavioral MM")
	ErrInvalidIndicatorEvent = errors.New("invalid indicator event for adjustment")
)

// TrailingStopMode defines how trailing stops are calculated
type TrailingStopMode string

const (
	TrailingModeFixed    TrailingStopMode = "fixed"    // Fixed percentage
	TrailingModeAdaptive TrailingStopMode = "adaptive" // Based on DMI trend detection
	TrailingModeDynamic  TrailingStopMode = "dynamic"  // Real-time adjustment based on events
)

// BehavioralConfig holds behavioral money management configuration for MACD/CCI/DMI strategy
type BehavioralConfig struct {
	// Trailing Stop Configuration
	TrailingMode                TrailingStopMode `json:"trailing_mode"`
	TrendTrailingPercent        float64          `json:"trend_trailing_percent"`         // Default: 2.0% for trend
	CounterTrendTrailingPercent float64          `json:"counter_trend_trailing_percent"` // Default: 1.5% for counter-trend

	// Dynamic Adjustments
	EnableDynamicAdjustments     bool    `json:"enable_dynamic_adjustments"`
	CCIZoneAdjustmentPercent     float64 `json:"cci_zone_adjustment_percent"`     // Default: 0.5% tighter when CCI inverse
	MACDInverseAdjustmentPercent float64 `json:"macd_inverse_adjustment_percent"` // Default: 0.3% tighter on MACD inverse
	DMICounterAdjustmentPercent  float64 `json:"dmi_counter_adjustment_percent"`  // Default: 0.4% tighter on DI counter

	// Early Exit Configuration
	EnableEarlyExit        bool    `json:"enable_early_exit"`
	EarlyExitOnMACDInverse bool    `json:"early_exit_on_macd_inverse"`
	MinProfitForEarlyExit  float64 `json:"min_profit_for_early_exit"` // Minimum profit % before early exit

	// Profit Adjustment Grid
	EnableProfitGrid       bool                    `json:"enable_profit_grid"`
	ProfitAdjustmentLevels []ProfitAdjustmentLevel `json:"profit_adjustment_levels"`

	// Event Monitoring
	EnableCCIMonitoring    bool    `json:"enable_cci_monitoring"`
	CCIMonitoringThreshold float64 `json:"cci_monitoring_threshold"` // Default: 100 (CCI threshold)

	// Risk Controls
	MaxAdjustmentPercent float64 `json:"max_adjustment_percent"` // Max adjustment in one go
	MinTrailingPercent   float64 `json:"min_trailing_percent"`   // Minimum trailing stop
	MaxTrailingPercent   float64 `json:"max_trailing_percent"`   // Maximum trailing stop
}

// ProfitAdjustmentLevel defines profit-based trailing stop adjustments
type ProfitAdjustmentLevel struct {
	ProfitThresholdPercent float64 `json:"profit_threshold_percent"` // % profit threshold
	AdjustmentPercent      float64 `json:"adjustment_percent"`       // % to tighten trailing stop
	Description            string  `json:"description"`              // Human readable description
}

// PositionState tracks current position for behavioral management
type PositionState struct {
	IsOpen                 bool      `json:"is_open"`
	EntryTime              time.Time `json:"entry_time"`
	EntryPrice             float64   `json:"entry_price"`
	Direction              string    `json:"direction"` // "LONG" or "SHORT"
	InitialTrailingPercent float64   `json:"initial_trailing_percent"`
	CurrentTrailingPercent float64   `json:"current_trailing_percent"`
	CurrentStopPrice       float64   `json:"current_stop_price"`

	// Entry conditions tracking
	EntryMACDValue  float64 `json:"entry_macd_value"`
	EntryMACDSignal float64 `json:"entry_macd_signal"`
	EntryCCIValue   float64 `json:"entry_cci_value"`
	EntryCCIZone    string  `json:"entry_cci_zone"`  // "OVERSOLD", "OVERBOUGHT", "NEUTRAL"
	EntryDMITrend   string  `json:"entry_dmi_trend"` // "TREND", "COUNTER_TREND"

	// Dynamic state
	LastAdjustmentTime  time.Time         `json:"last_adjustment_time"`
	TotalAdjustments    int               `json:"total_adjustments"`
	LastIndicatorValues IndicatorSnapshot `json:"last_indicator_values"`
}

// IndicatorSnapshot holds current indicator values for comparison
type IndicatorSnapshot struct {
	Timestamp  time.Time `json:"timestamp"`
	MACDValue  float64   `json:"macd_value"`
	MACDSignal float64   `json:"macd_signal"`
	CCIValue   float64   `json:"cci_value"`
	DIPlus     float64   `json:"di_plus"`
	DIMinus    float64   `json:"di_minus"`
	ADXValue   float64   `json:"adx_value"`
}

// AdjustmentReason defines why a trailing stop adjustment was made
type AdjustmentReason string

const (
	AdjustmentCCIInverse  AdjustmentReason = "CCI_ZONE_INVERSE"
	AdjustmentMACDInverse AdjustmentReason = "MACD_INVERSE"
	AdjustmentDMICounter  AdjustmentReason = "DMI_COUNTER_TREND"
	AdjustmentProfitGrid  AdjustmentReason = "PROFIT_GRID_LEVEL"
	AdjustmentVolatility  AdjustmentReason = "VOLATILITY_ADJUSTMENT"
)

// TrailingAdjustment represents a trailing stop adjustment event
type TrailingAdjustment struct {
	Timestamp            time.Time         `json:"timestamp"`
	Reason               AdjustmentReason  `json:"reason"`
	OldTrailingPercent   float64           `json:"old_trailing_percent"`
	NewTrailingPercent   float64           `json:"new_trailing_percent"`
	OldStopPrice         float64           `json:"old_stop_price"`
	NewStopPrice         float64           `json:"new_stop_price"`
	CurrentPrice         float64           `json:"current_price"`
	CurrentProfitPercent float64           `json:"current_profit_percent"`
	IndicatorSnapshot    IndicatorSnapshot `json:"indicator_snapshot"`
	Success              bool              `json:"success"`
	ErrorMessage         string            `json:"error_message,omitempty"`
}

// EarlyExitDecision represents an early exit decision
type EarlyExitDecision struct {
	Timestamp         time.Time         `json:"timestamp"`
	ShouldExit        bool              `json:"should_exit"`
	Reason            string            `json:"reason"`
	CurrentPrice      float64           `json:"current_price"`
	ProfitPercent     float64           `json:"profit_percent"`
	IndicatorSnapshot IndicatorSnapshot `json:"indicator_snapshot"`
}

// BehavioralMMStatus provides status information
type BehavioralMMStatus struct {
	IsActive               bool                 `json:"is_active"`
	PositionState          PositionState        `json:"position_state"`
	RecentAdjustments      []TrailingAdjustment `json:"recent_adjustments"`
	LastEarlyExitDecision  *EarlyExitDecision   `json:"last_early_exit_decision,omitempty"`
	TotalAdjustmentsMade   int                  `json:"total_adjustments_made"`
	AverageTrailingPercent float64              `json:"average_trailing_percent"`
}

// DefaultBehavioralConfig returns default configuration for MACD/CCI/DMI behavioral MM
func DefaultBehavioralConfig() BehavioralConfig {
	return BehavioralConfig{
		// Trailing Stop Configuration
		TrailingMode:                TrailingModeAdaptive,
		TrendTrailingPercent:        2.0, // 2% for trend following
		CounterTrendTrailingPercent: 1.5, // 1.5% for counter-trend

		// Dynamic Adjustments
		EnableDynamicAdjustments:     true,
		CCIZoneAdjustmentPercent:     0.5, // 0.5% tighter when CCI in inverse zone
		MACDInverseAdjustmentPercent: 0.3, // 0.3% tighter on MACD inverse signal
		DMICounterAdjustmentPercent:  0.4, // 0.4% tighter on DMI counter signal

		// Early Exit Configuration
		EnableEarlyExit:        true,
		EarlyExitOnMACDInverse: true,
		MinProfitForEarlyExit:  0.5, // Minimum 0.5% profit before early exit

		// Profit Adjustment Grid
		EnableProfitGrid: true,
		ProfitAdjustmentLevels: []ProfitAdjustmentLevel{
			{ProfitThresholdPercent: 1.0, AdjustmentPercent: 0.2, Description: "First profit level - slight tightening"},
			{ProfitThresholdPercent: 2.0, AdjustmentPercent: 0.3, Description: "Second profit level - moderate tightening"},
			{ProfitThresholdPercent: 3.0, AdjustmentPercent: 0.4, Description: "Third profit level - aggressive tightening"},
			{ProfitThresholdPercent: 5.0, AdjustmentPercent: 0.5, Description: "High profit level - maximum protection"},
		},

		// Event Monitoring
		EnableCCIMonitoring:    true,
		CCIMonitoringThreshold: 100.0,

		// Risk Controls
		MaxAdjustmentPercent: 1.0, // Max 1% adjustment in one go
		MinTrailingPercent:   0.5, // Minimum 0.5% trailing stop
		MaxTrailingPercent:   5.0, // Maximum 5% trailing stop
	}
}

// Validate validates behavioral configuration
func (bc BehavioralConfig) Validate() error {
	if bc.TrendTrailingPercent <= 0 || bc.TrendTrailingPercent > 10 {
		return errors.New("trend trailing percent must be between 0.1 and 10.0")
	}
	if bc.CounterTrendTrailingPercent <= 0 || bc.CounterTrendTrailingPercent > 10 {
		return errors.New("counter-trend trailing percent must be between 0.1 and 10.0")
	}
	if bc.MaxAdjustmentPercent <= 0 || bc.MaxAdjustmentPercent > 5 {
		return errors.New("max adjustment percent must be between 0.1 and 5.0")
	}
	if bc.MinTrailingPercent >= bc.MaxTrailingPercent {
		return errors.New("min trailing percent must be less than max trailing percent")
	}
	if bc.MinProfitForEarlyExit < 0 {
		return errors.New("min profit for early exit cannot be negative")
	}

	// Validate profit grid
	if bc.EnableProfitGrid {
		for i, level := range bc.ProfitAdjustmentLevels {
			if level.ProfitThresholdPercent <= 0 {
				return errors.New("profit threshold must be positive")
			}
			if level.AdjustmentPercent <= 0 || level.AdjustmentPercent > bc.MaxAdjustmentPercent {
				return errors.New("adjustment percent must be positive and within max adjustment")
			}
			if i > 0 && level.ProfitThresholdPercent <= bc.ProfitAdjustmentLevels[i-1].ProfitThresholdPercent {
				return errors.New("profit thresholds must be in ascending order")
			}
		}
	}

	return nil
}

// DetermineTrendType determines if current DMI indicates trend or counter-trend
func DetermineTrendType(diPlus, diMinus float64) string {
	if diPlus > diMinus {
		return "TREND"
	}
	return "COUNTER_TREND"
}

// DetermineCCIZone determines CCI zone based on value
func DetermineCCIZone(cciValue, threshold float64) string {
	if cciValue < -threshold {
		return "OVERSOLD"
	}
	if cciValue > threshold {
		return "OVERBOUGHT"
	}
	return "NEUTRAL"
}

// CalculateCurrentProfit calculates current profit percentage
func (ps PositionState) CalculateCurrentProfit(currentPrice float64) float64 {
	if !ps.IsOpen || ps.EntryPrice == 0 {
		return 0
	}

	switch ps.Direction {
	case "LONG":
		return ((currentPrice - ps.EntryPrice) / ps.EntryPrice) * 100
	case "SHORT":
		return ((ps.EntryPrice - currentPrice) / ps.EntryPrice) * 100
	default:
		return 0
	}
}

// ShouldTriggerStop checks if current price should trigger stop loss
func (ps PositionState) ShouldTriggerStop(currentPrice float64) bool {
	if !ps.IsOpen {
		return false
	}

	switch ps.Direction {
	case "LONG":
		return currentPrice <= ps.CurrentStopPrice
	case "SHORT":
		return currentPrice >= ps.CurrentStopPrice
	default:
		return false
	}
}
