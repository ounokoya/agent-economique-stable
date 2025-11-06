// Package indicators implements technical indicators calculations for trading strategy
package indicators

import (
	"errors"
	"fmt"
	"time"
)

// SignalType defines the type of trading signal
type SignalType int

const (
	TrendSignal SignalType = iota
	CounterTrendSignal
)

// String returns the string representation of SignalType
func (s SignalType) String() string {
	switch s {
	case TrendSignal:
		return "TREND"
	case CounterTrendSignal:
		return "COUNTER_TREND"
	default:
		return "UNKNOWN"
	}
}

// SignalDirection defines the direction of a trading signal
type SignalDirection int

const (
	NoSignal SignalDirection = iota
	LongSignal
	ShortSignal
)

// String returns the string representation of SignalDirection
func (d SignalDirection) String() string {
	switch d {
	case LongSignal:
		return "LONG"
	case ShortSignal:
		return "SHORT"
	default:
		return "NONE"
	}
}

// CCIZone defines CCI zone classifications
type CCIZone int

const (
	CCINormal CCIZone = iota
	CCIOversold
	CCIOverbought
)

// String returns the string representation of CCIZone
func (z CCIZone) String() string {
	switch z {
	case CCIOversold:
		return "OVERSOLD"
	case CCIOverbought:
		return "OVERBOUGHT"
	default:
		return "NORMAL"
	}
}

// CrossoverType defines the type of crossover
type CrossoverType int

const (
	NoCrossover CrossoverType = iota
	CrossUp
	CrossDown
)

// String returns the string representation of CrossoverType
func (c CrossoverType) String() string {
	switch c {
	case CrossUp:
		return "CROSS_UP"
	case CrossDown:
		return "CROSS_DOWN"
	default:
		return "NO_CROSS"
	}
}

// Kline represents candlestick data for calculations
type Kline struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// MACD represents MACD indicator values
type MACD struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// MACDResult contains MACD calculation results with history
type MACDResult struct {
	Current    MACD        `json:"current"`
	History    []MACD      `json:"history"`
	Crossovers []Crossover `json:"crossovers"`
}

// CCI represents CCI indicator values
type CCI struct {
	Value float64 `json:"value"`
	Zone  CCIZone `json:"zone"`
}

// CCIResult contains CCI calculation results
type CCIResult struct {
	Current     CCI         `json:"current"`
	History     []CCI       `json:"history"`
	ZoneChanges []ZoneEvent `json:"zone_changes"`
}

// DMI represents DMI/ADX indicator values
type DMI struct {
	DIPlus  float64 `json:"di_plus"`
	DIMinus float64 `json:"di_minus"`
	DX      float64 `json:"dx"`
	ADX     float64 `json:"adx"`
}

// DMIResult contains DMI calculation results
type DMIResult struct {
	Current    DMI         `json:"current"`
	History    []DMI       `json:"history"`
	Crossovers []Crossover `json:"crossovers"`
}

// Crossover represents a crossover event
type Crossover struct {
	Type      CrossoverType `json:"type"`
	Index     int           `json:"index"`
	Value1    float64       `json:"value1"`
	Value2    float64       `json:"value2"`
}

// TradingSignal represents a complete trading signal
type TradingSignal struct {
	Direction    SignalDirection `json:"direction"`
	Type         SignalType      `json:"type"`
	Confidence   float64         `json:"confidence"`
	Timestamp    int64           `json:"timestamp"`
	EntryPrice   float64         `json:"entry_price"`
	CCIZone      CCIZone         `json:"cci_zone"`
	Triggers     []string        `json:"triggers"`
	Filters      []string        `json:"filters"`
	Rejected     bool            `json:"rejected"`
	RejectReason string          `json:"reject_reason,omitempty"`
}

// IndicatorConfig holds configuration for all indicators
type IndicatorConfig struct {
	MACD   MACDConfig   `json:"macd"`
	CCI    CCIConfig    `json:"cci"`
	DMI    DMIConfig    `json:"dmi"`
	Signal SignalConfig `json:"signal"`
}

// MACDConfig defines MACD calculation parameters
type MACDConfig struct {
	FastPeriod   int `json:"fast_period"`
	SlowPeriod   int `json:"slow_period"`
	SignalPeriod int `json:"signal_period"`
}

// CCIConfig defines CCI calculation parameters
type CCIConfig struct {
	Period             int     `json:"period"`
	OversoldThreshold  float64 `json:"oversold_threshold"`
	OverboughtThreshold float64 `json:"overbought_threshold"`
}

// DMIConfig defines DMI/ADX calculation parameters
type DMIConfig struct {
	Period    int `json:"period"`
	ADXPeriod int `json:"adx_period"`
}

// SignalConfig defines signal generation parameters
type SignalConfig struct {
	// Signal type thresholds
	TrendCCIOversold      float64 `json:"trend_cci_oversold"`
	TrendCCIOverbought    float64 `json:"trend_cci_overbought"`
	CounterCCIOversold    float64 `json:"counter_cci_oversold"`
	CounterCCIOverbought  float64 `json:"counter_cci_overbought"`
	
	// Optional filters
	MACDSameSignFilter    bool    `json:"macd_same_sign_filter"`
	DXADXCrossFilter      bool    `json:"dx_adx_cross_filter"`
	
	// Confidence calculation
	MinConfidence         float64 `json:"min_confidence"`
	MACDWeakPenalty       float64 `json:"macd_weak_penalty"`
	CCIClosePenalty       float64 `json:"cci_close_penalty"`
	ADXLowPenalty         float64 `json:"adx_low_penalty"`
}

// REMOVED - OLD TYPES REPLACED BY NEW INTERFACE BELOW

// Common errors for indicator calculations
var (
	ErrInsufficientData    = errors.New("insufficient data for calculation")
	ErrInvalidPeriod       = errors.New("invalid period parameter")
	ErrInvalidConfig       = errors.New("invalid configuration")
	ErrCalculationFailed   = errors.New("calculation failed")
	ErrInvalidKlineData    = errors.New("invalid kline data")
)

// ValidateIndicatorConfig validates the indicator configuration
func ValidateIndicatorConfig(config IndicatorConfig) error {
	// Validate MACD config
	if config.MACD.FastPeriod <= 0 {
		return fmt.Errorf("MACD fast period must be positive: %d", config.MACD.FastPeriod)
	}
	if config.MACD.SlowPeriod <= config.MACD.FastPeriod {
		return fmt.Errorf("MACD slow period (%d) must be greater than fast period (%d)", 
			config.MACD.SlowPeriod, config.MACD.FastPeriod)
	}
	if config.MACD.SignalPeriod <= 0 {
		return fmt.Errorf("MACD signal period must be positive: %d", config.MACD.SignalPeriod)
	}
	
	// Validate CCI config
	if config.CCI.Period <= 0 {
		return fmt.Errorf("CCI period must be positive: %d", config.CCI.Period)
	}
	if config.CCI.OversoldThreshold >= 0 {
		return fmt.Errorf("CCI oversold threshold must be negative: %f", config.CCI.OversoldThreshold)
	}
	if config.CCI.OverboughtThreshold <= 0 {
		return fmt.Errorf("CCI overbought threshold must be positive: %f", config.CCI.OverboughtThreshold)
	}
	
	// Validate DMI config
	if config.DMI.Period <= 0 {
		return fmt.Errorf("DMI period must be positive: %d", config.DMI.Period)
	}
	if config.DMI.ADXPeriod <= 0 {
		return fmt.Errorf("ADX period must be positive: %d", config.DMI.ADXPeriod)
	}
	
	// Validate Signal config
	if config.Signal.MinConfidence < 0 || config.Signal.MinConfidence > 100 {
		return fmt.Errorf("min confidence must be between 0 and 100: %f", config.Signal.MinConfidence)
	}
	
	return nil
}

// DefaultIndicatorConfig returns default configuration for all indicators
func DefaultIndicatorConfig() IndicatorConfig {
	return IndicatorConfig{
		MACD: MACDConfig{
			FastPeriod:   12,
			SlowPeriod:   26,
			SignalPeriod: 9,
		},
		CCI: CCIConfig{
			Period:              14,
			OversoldThreshold:   -100.0,
			OverboughtThreshold: 100.0,
		},
		DMI: DMIConfig{
			Period:    14,
			ADXPeriod: 14,
		},
		Signal: SignalConfig{
			TrendCCIOversold:     -100.0,
			TrendCCIOverbought:   100.0,
			CounterCCIOversold:   -180.0,
			CounterCCIOverbought: 180.0,
			MACDSameSignFilter:   false,
			DXADXCrossFilter:     false,
			MinConfidence:        70.0,
			MACDWeakPenalty:      10.0,
			CCIClosePenalty:      15.0,
			ADXLowPenalty:        20.0,
		},
	}
}

// ValidateKlines validates kline data for calculations
func ValidateKlines(klines []Kline) error {
	if len(klines) == 0 {
		return fmt.Errorf("%w: no klines provided", ErrInsufficientData)
	}
	
	for i, kline := range klines {
		if kline.High < kline.Low {
			return fmt.Errorf("%w: high < low at index %d", ErrInvalidKlineData, i)
		}
		if kline.Close < 0 || kline.Open < 0 {
			return fmt.Errorf("%w: negative prices at index %d", ErrInvalidKlineData, i)
		}
		if kline.High < kline.Close || kline.High < kline.Open {
			return fmt.Errorf("%w: high < close/open at index %d", ErrInvalidKlineData, i)
		}
		if kline.Low > kline.Close || kline.Low > kline.Open {
			return fmt.Errorf("%w: low > close/open at index %d", ErrInvalidKlineData, i)
		}
	}
	
	return nil
}

// GetTypicalPrice calculates typical price (HLC/3) for a kline
func GetTypicalPrice(kline Kline) float64 {
	return (kline.High + kline.Low + kline.Close) / 3.0
}

// GetTrueRange calculates true range for DMI calculations
func GetTrueRange(current, previous Kline) float64 {
	tr1 := current.High - current.Low
	tr2 := abs(current.High - previous.Close)
	tr3 := abs(current.Low - previous.Close)
	
	return max(tr1, max(tr2, tr3))
}

// Communication Interface (Engine ↔ Indicateurs)

// CalculationRequest represents a request from Engine Temporel to calculate indicators
type CalculationRequest struct {
	// Context (Engine provides validated data)
	Symbol         string    `json:"symbol"`
	Timeframe      string    `json:"timeframe"`
	CurrentTime    int64     `json:"current_time"`
	CandleWindow   []Kline   `json:"candle_window"`    // Engine provides >= 35 validated candles
	
	// Position context for zone events detection
	PositionContext *PositionContext `json:"position_context,omitempty"`
	
	// Traceability
	RequestID      string    `json:"request_id"`
}

// PositionContext represents current position state (uses Engine types)
type PositionContext struct {
	IsOpen         bool      `json:"is_open"`
	Direction      string    `json:"direction"`        // "LONG" | "SHORT" 
	EntryPrice     float64   `json:"entry_price"`
	EntryTime      int64     `json:"entry_time"`
	EntryCCIZone   CCIZone   `json:"entry_cci_zone"`   // For inverse detection
	ProfitPercent  float64   `json:"profit_percent"`   // Current profit %
}

// CalculationResponse represents response from Indicateurs to Engine Temporel
type CalculationResponse struct {
	// Metadata
	RequestID       string        `json:"request_id"`
	Success         bool          `json:"success"`
	Error           error         `json:"error,omitempty"`
	CalculationTime time.Duration `json:"calculation_time"`
	
	// Results
	Results    *IndicatorResults `json:"results,omitempty"`
	Signals    []StrategySignal  `json:"signals,omitempty"`
	ZoneEvents []ZoneEvent       `json:"zone_events,omitempty"`
}

// MACDValues contains MACD indicator values
type MACDValues struct {
	MACD          float64       `json:"macd"`
	Signal        float64       `json:"signal"`
	Histogram     float64       `json:"histogram"`
	CrossoverType CrossoverType `json:"crossover_type"`
}

// CCIValues contains CCI indicator values
type CCIValues struct {
	Value float64 `json:"value"`
	Zone  CCIZone `json:"zone"`
}

// DMIValues contains DMI/ADX indicator values
type DMIValues struct {
	PlusDI  float64 `json:"plus_di"`
	MinusDI float64 `json:"minus_di"`
	DX      float64 `json:"dx"`
	ADX     float64 `json:"adx"`
}

// StochasticValues contains Stochastic %K and %D values
type StochasticValues struct {
	K             float64       `json:"k"`
	D             float64       `json:"d"`
	Zone          StochZone     `json:"zone"`
	CrossoverType CrossoverType `json:"crossover_type"`
	IsExtreme     bool          `json:"is_extreme"`
}

// MFIValues contains Money Flow Index values
type MFIValues struct {
	Value     float64 `json:"value"`
	Zone      MFIZone `json:"zone"`
	IsExtreme bool    `json:"is_extreme"`
}

// StochZone represents Stochastic oscillator zones
type StochZone int

const (
	StochOversold StochZone = iota
	StochNeutral
	StochOverbought
)

func (z StochZone) String() string {
	switch z {
	case StochOversold:
		return "OVERSOLD"
	case StochOverbought:
		return "OVERBOUGHT"
	default:
		return "NEUTRAL"
	}
}

// MFIZone represents Money Flow Index zones
type MFIZone int

const (
	MFIOversold MFIZone = iota
	MFINeutral
	MFIOverbought
)

func (z MFIZone) String() string {
	switch z {
	case MFIOversold:
		return "OVERSOLD"
	case MFIOverbought:
		return "OVERBOUGHT"
	default:
		return "NEUTRAL"
	}
}

// IndicatorResults contains calculated indicator values
type IndicatorResults struct {
	MACD       *MACDValues       `json:"macd"`
	CCI        *CCIValues        `json:"cci"`
	DMI        *DMIValues        `json:"dmi"`
	Stochastic *StochasticValues `json:"stochastic"`
	MFI        *MFIValues        `json:"mfi"`
	Timestamp  int64             `json:"timestamp"`
}

// StrategySignal represents a trading signal generated by strategy
type StrategySignal struct {
	Direction   SignalDirection `json:"direction"`
	Type        SignalType      `json:"type"`
	Confidence  float64         `json:"confidence"`
	Timestamp   int64           `json:"timestamp"`
}

// ZoneEvent represents zone monitoring event (compatible with Engine types)
type ZoneEvent struct {
	Type            string  `json:"type"`               // "ZONE_ACTIVATED" | "ZONE_DEACTIVATED"
	ZoneType        string  `json:"zone_type"`          // "CCI_INVERSE" | "MACD_INVERSE" | "DI_COUNTER"
	IsInverse       bool    `json:"is_inverse,omitempty"`
	RequiresProfit  bool    `json:"requires_profit,omitempty"`
	ProfitThreshold float64 `json:"profit_threshold,omitempty"`
	CurrentProfit   float64 `json:"current_profit,omitempty"`
	Timestamp       int64   `json:"timestamp"`
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
