package stoch_mfi_cci

import (
	"errors"
	"time"
	
	"agent-economique/internal/indicators"
)

// Errors for STOCH/MFI/CCI strategy
var (
	ErrInvalidConfig       = errors.New("invalid strategy configuration")
	ErrInsufficientData    = errors.New("insufficient data for signal generation")
	ErrNoPositionOpen      = errors.New("no open position for monitoring")
	ErrInvalidTimeframe    = errors.New("invalid timeframe configuration")
)

// SignalStrength defines signal quality levels
type SignalStrength string

const (
	SignalMinimal SignalStrength = "MINIMAL" // STOCH + (MFI OR CCI) extreme
	SignalPremium SignalStrength = "PREMIUM" // STOCH + MFI + CCI all extreme
)

// MonitoringState defines current monitoring state
type MonitoringState string

const (
	StateNormal        MonitoringState = "NORMAL"         // Normal marker event processing
	StateSTOCHInverse  MonitoringState = "STOCH_INVERSE"  // Tick-by-tick monitoring active
	StateTripleInverse MonitoringState = "TRIPLE_INVERSE" // Maximum protection mode
)

// StrategyConfig holds configuration for STOCH/MFI/CCI strategy
type StrategyConfig struct {
	// Indicator Parameters
	StochPeriodK    int     `json:"stoch_period_k"`     // Default: 14
	StochSmoothK    int     `json:"stoch_smooth_k"`     // Default: 3
	StochPeriodD    int     `json:"stoch_period_d"`     // Default: 3
	StochOversold   float64 `json:"stoch_oversold"`     // Default: 20
	StochOverbought float64 `json:"stoch_overbought"`   // Default: 80
	
	MFIPeriod       int     `json:"mfi_period"`         // Default: 14
	MFIOversold     float64 `json:"mfi_oversold"`       // Default: 20
	MFIOverbought   float64 `json:"mfi_overbought"`     // Default: 80
	
	CCIThreshold    float64 `json:"cci_threshold"`      // Default: 100
	
	// Signal Generation
	MinConfidence         float64 `json:"min_confidence"`          // Default: 0.7
	PremiumConfidence     float64 `json:"premium_confidence"`      // Default: 0.9
	RequireBarConfirmation bool   `json:"require_bar_confirmation"` // Default: true
	
	// Multi-Timeframe
	HigherTimeframe string `json:"higher_timeframe"`    // Default: auto-detect
	EnableMultiTF   bool   `json:"enable_multi_tf"`     // Default: true
	
	// Trailing Management
	BaseTrailingPercent      float64 `json:"base_trailing_percent"`       // Default: 2.0
	TrendTrailingPercent     float64 `json:"trend_trailing_percent"`      // Default: 2.5
	CounterTrendTrailing     float64 `json:"counter_trend_trailing"`      // Default: 1.5
	
	// Dynamic Adjustments (tick-by-tick)
	EnableDynamicAdjustments bool    `json:"enable_dynamic_adjustments"`  // Default: true
	STOCHInverseAdjust       float64 `json:"stoch_inverse_adjust"`        // Default: 0.2
	MFIInverseAdjust         float64 `json:"mfi_inverse_adjust"`          // Default: 0.3
	CCIInverseAdjust         float64 `json:"cci_inverse_adjust"`          // Default: 0.4
	TripleInverseAdjust      float64 `json:"triple_inverse_adjust"`       // Default: 0.9
	
	// Safety Limits
	MaxCumulativeAdjust      float64 `json:"max_cumulative_adjust"`       // Default: 1.0
	MinTrailingPercent       float64 `json:"min_trailing_percent"`        // Default: 0.3
	MaxTrailingPercent       float64 `json:"max_trailing_percent"`        // Default: 5.0
	
	// Early Exit
	EnableEarlyExit          bool    `json:"enable_early_exit"`           // Default: true
	MinProfitForEarlyExit    float64 `json:"min_profit_for_early_exit"`   // Default: 0.5
	TripleInverseEarlyExit   bool    `json:"triple_inverse_early_exit"`   // Default: true
}

// PositionState tracks current position for STOCH/MFI/CCI strategy
type PositionState struct {
	IsOpen         bool      `json:"is_open"`
	EntryTime      time.Time `json:"entry_time"`
	EntryPrice     float64   `json:"entry_price"`
	Direction      string    `json:"direction"`              // "LONG" or "SHORT"
	SignalStrength SignalStrength `json:"signal_strength"`   // MINIMAL or PREMIUM
	
	// Entry Conditions
	EntrySTOCHK       float64 `json:"entry_stoch_k"`
	EntrySTOCHD       float64 `json:"entry_stoch_d"`
	EntrySTOCHZone    string  `json:"entry_stoch_zone"`      // "OVERSOLD", "OVERBOUGHT", "NEUTRAL"
	EntryMFIValue     float64 `json:"entry_mfi_value"`
	EntryMFIZone      string  `json:"entry_mfi_zone"`
	EntryCCIValue     float64 `json:"entry_cci_value"`
	EntryCCIZone      string  `json:"entry_cci_zone"`
	
	// Multi-Timeframe Context
	TrendClassification string  `json:"trend_classification"`   // "TREND" or "COUNTER_TREND"
	HigherTFData        string  `json:"higher_tf_data"`        // Serialized higher TF context
	
	// Trailing State
	InitialTrailing     float64 `json:"initial_trailing"`
	CurrentTrailing     float64 `json:"current_trailing"`
	CurrentStopPrice    float64 `json:"current_stop_price"`
	
	// Monitoring State
	MonitoringState     MonitoringState `json:"monitoring_state"`
	LastAdjustmentTime  time.Time       `json:"last_adjustment_time"`
	TotalAdjustments    int             `json:"total_adjustments"`
	CumulativeAdjust    float64         `json:"cumulative_adjust"`       // Total % adjusted
}

// IndicatorSnapshot holds current indicator values for comparison
type IndicatorSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	STOCHValue   float64   `json:"stoch_k"`
	STOCHSignal  float64   `json:"stoch_d"`
	STOCHZone    string    `json:"stoch_zone"`
	MFIValue     float64   `json:"mfi_value"`
	MFIZone      string    `json:"mfi_zone"`
	CCIValue     float64   `json:"cci_value"`
	CCIZone      string    `json:"cci_zone"`
}

// AdjustmentReason defines why a trailing adjustment was made
type AdjustmentReason string

const (
	AdjustmentSTOCHInverse  AdjustmentReason = "STOCH_INVERSE"
	AdjustmentMFIInverse    AdjustmentReason = "MFI_INVERSE"
	AdjustmentCCIInverse    AdjustmentReason = "CCI_INVERSE"
	AdjustmentTripleInverse AdjustmentReason = "TRIPLE_INVERSE"
	AdjustmentNormalUpdate  AdjustmentReason = "NORMAL_UPDATE"
)

// TrailingAdjustment represents a trailing stop adjustment event
type TrailingAdjustment struct {
	Timestamp             time.Time        `json:"timestamp"`
	Reason                AdjustmentReason `json:"reason"`
	OldTrailingPercent    float64          `json:"old_trailing_percent"`
	NewTrailingPercent    float64          `json:"new_trailing_percent"`
	OldStopPrice          float64          `json:"old_stop_price"`
	NewStopPrice          float64          `json:"new_stop_price"`
	CurrentPrice          float64          `json:"current_price"`
	CurrentProfitPercent  float64          `json:"current_profit_percent"`
	IndicatorSnapshot     IndicatorSnapshot `json:"indicator_snapshot"`
	CumulativeAdjustment  float64          `json:"cumulative_adjustment"`
	Success               bool             `json:"success"`
	ErrorMessage          string           `json:"error_message,omitempty"`
}

// EarlyExitDecision represents an early exit decision
type EarlyExitDecision struct {
	Timestamp         time.Time         `json:"timestamp"`
	ShouldExit        bool              `json:"should_exit"`
	Reason            string            `json:"reason"`
	CurrentPrice      float64           `json:"current_price"`
	ProfitPercent     float64           `json:"profit_percent"`
	IndicatorSnapshot IndicatorSnapshot `json:"indicator_snapshot"`
	TriggerCondition  string            `json:"trigger_condition"`    // "TRIPLE_INVERSE", "STOCH_EXTREME", etc.
}

// StrategyStatus provides current strategy status
type StrategyStatus struct {
	IsActive              bool                 `json:"is_active"`
	PositionState         PositionState        `json:"position_state"`
	MonitoringState       MonitoringState      `json:"monitoring_state"`
	RecentAdjustments     []TrailingAdjustment `json:"recent_adjustments"`
	LastEarlyExitDecision *EarlyExitDecision   `json:"last_early_exit_decision,omitempty"`
	TotalSignalsGenerated int                  `json:"total_signals_generated"`
	PremiumSignalsCount   int                  `json:"premium_signals_count"`
	AverageConfidence     float64              `json:"average_confidence"`
	TickProcessingCount   int64                `json:"tick_processing_count"`
}

// DefaultStrategyConfig returns default configuration for STOCH/MFI/CCI strategy
func DefaultStrategyConfig() StrategyConfig {
	return StrategyConfig{
		// Indicator Parameters
		StochPeriodK:    14,
		StochSmoothK:    3,
		StochPeriodD:    3,
		StochOversold:   20.0,
		StochOverbought: 80.0,
		
		MFIPeriod:       14,
		MFIOversold:     20.0,
		MFIOverbought:   80.0,
		
		CCIThreshold:    100.0,
		
		// Signal Generation
		MinConfidence:         0.7,
		PremiumConfidence:     0.9,
		RequireBarConfirmation: true,
		
		// Multi-Timeframe
		HigherTimeframe: "auto",  // Auto-detect based on base timeframe
		EnableMultiTF:   true,
		
		// Trailing Management
		BaseTrailingPercent:      2.0,
		TrendTrailingPercent:     2.5,
		CounterTrendTrailing:     1.5,
		
		// Dynamic Adjustments
		EnableDynamicAdjustments: true,
		STOCHInverseAdjust:       0.2,
		MFIInverseAdjust:         0.3,
		CCIInverseAdjust:         0.4,
		TripleInverseAdjust:      0.9,
		
		// Safety Limits
		MaxCumulativeAdjust:      1.0,
		MinTrailingPercent:       0.3,
		MaxTrailingPercent:       5.0,
		
		// Early Exit
		EnableEarlyExit:          true,
		MinProfitForEarlyExit:    0.5,
		TripleInverseEarlyExit:   true,
	}
}

// Validate validates strategy configuration
func (sc StrategyConfig) Validate() error {
	if sc.StochPeriodK <= 0 || sc.StochPeriodK > 100 {
		return errors.New("stoch period K must be between 1 and 100")
	}
	if sc.StochOversold <= 0 || sc.StochOversold >= sc.StochOverbought {
		return errors.New("invalid stoch oversold/overbought levels")
	}
	if sc.MFIPeriod <= 0 || sc.MFIPeriod > 100 {
		return errors.New("MFI period must be between 1 and 100")
	}
	if sc.MFIOversold <= 0 || sc.MFIOversold >= sc.MFIOverbought {
		return errors.New("invalid MFI oversold/overbought levels")
	}
	if sc.MinConfidence <= 0 || sc.MinConfidence > 1.0 {
		return errors.New("min confidence must be between 0.1 and 1.0")
	}
	if sc.PremiumConfidence <= sc.MinConfidence || sc.PremiumConfidence > 1.0 {
		return errors.New("premium confidence must be higher than min confidence and <= 1.0")
	}
	if sc.MaxCumulativeAdjust <= 0 || sc.MaxCumulativeAdjust > 5.0 {
		return errors.New("max cumulative adjust must be between 0.1 and 5.0")
	}
	if sc.MinTrailingPercent >= sc.MaxTrailingPercent {
		return errors.New("min trailing must be less than max trailing")
	}
	
	return nil
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

// ConvertIndicatorResults converts engine indicator results to strategy snapshot
func ConvertIndicatorResults(results *indicators.IndicatorResults) IndicatorSnapshot {
	snapshot := IndicatorSnapshot{
		Timestamp: time.Now().UTC(),
	}
	
	// Extract STOCH values
	if results.Stochastic != nil {
		snapshot.STOCHValue = results.Stochastic.K
		snapshot.STOCHSignal = results.Stochastic.D
		snapshot.STOCHZone = results.Stochastic.Zone.String()
	}
	
	// Extract MFI values
	if results.MFI != nil {
		snapshot.MFIValue = results.MFI.Value
		snapshot.MFIZone = results.MFI.Zone.String()
	}
	
	// Extract CCI values
	if results.CCI != nil {
		snapshot.CCIValue = results.CCI.Value
		snapshot.CCIZone = results.CCI.Zone.String()
	}
	
	return snapshot
}

// DetectSTOCHInverse checks if STOCH moved to inverse zone
func DetectSTOCHInverse(entryZone, currentZone string, direction string) bool {
	switch direction {
	case "LONG":
		return entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		return entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	default:
		return false
	}
}

// DetectTripleInverse checks if all three indicators are in inverse zones
func DetectTripleInverse(entry, current IndicatorSnapshot, direction string) bool {
	stochInverse := DetectSTOCHInverse(entry.STOCHZone, current.STOCHZone, direction)
	
	var mfiInverse, cciInverse bool
	
	switch direction {
	case "LONG":
		mfiInverse = entry.MFIZone == "OVERSOLD" && current.MFIZone == "OVERBOUGHT"
		cciInverse = entry.CCIZone == "OVERSOLD" && current.CCIZone == "OVERBOUGHT"
	case "SHORT":
		mfiInverse = entry.MFIZone == "OVERBOUGHT" && current.MFIZone == "OVERSOLD"
		cciInverse = entry.CCIZone == "OVERBOUGHT" && current.CCIZone == "OVERSOLD"
	}
	
	return stochInverse && mfiInverse && cciInverse
}
