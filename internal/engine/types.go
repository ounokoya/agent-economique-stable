// Package engine implements the temporal engine for trading strategy execution
package engine

import (
	"errors"
	"time"
)

// ExecutionMode defines the execution mode of the temporal engine
type ExecutionMode int

const (
	BacktestMode ExecutionMode = iota
	PaperMode
	LiveMode
)

// String returns the string representation of ExecutionMode
func (m ExecutionMode) String() string {
	switch m {
	case BacktestMode:
		return "backtest"
	case PaperMode:
		return "paper"
	case LiveMode:
		return "live"
	default:
		return "unknown"
	}
}

// PositionDirection defines the direction of a trading position
type PositionDirection int

const (
	PositionNone PositionDirection = iota
	PositionLong
	PositionShort
)

// String returns the string representation of PositionDirection
func (d PositionDirection) String() string {
	switch d {
	case PositionLong:
		return "LONG"
	case PositionShort:
		return "SHORT"
	default:
		return "NONE"
	}
}

// PositionState defines the state of a trading position
type PositionState int

const (
	PositionClosed PositionState = iota
	PositionOpen
)

// ZoneType defines the type of active zone monitoring
type ZoneType int

const (
	ZoneCCIInverse ZoneType = iota
	ZoneMACDInverse
	ZoneDICounter
)

// String returns the string representation of ZoneType
func (z ZoneType) String() string {
	switch z {
	case ZoneCCIInverse:
		return "CCI_INVERSE"
	case ZoneMACDInverse:
		return "MACD_INVERSE"
	case ZoneDICounter:
		return "DI_COUNTER"
	default:
		return "UNKNOWN"
	}
}

// Trade represents a single trade data point
type Trade struct {
	Timestamp    int64   `json:"timestamp"`
	Price        float64 `json:"price"`
	Quantity     float64 `json:"quantity"`
	IsBuyerMaker bool    `json:"is_buyer_maker"`
}

// Kline represents candlestick data
type Kline struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// Position represents a trading position
type Position struct {
	IsOpen      bool              `json:"is_open"`
	Direction   PositionDirection `json:"direction"`
	EntryPrice  float64           `json:"entry_price"`
	EntryTime   int64             `json:"entry_time"`
	StopLoss    float64           `json:"stop_loss"`
	EntryCCIZone string           `json:"entry_cci_zone"`
}

// CalculateProfitPercent calculates current profit percentage
func (p Position) CalculateProfitPercent(currentPrice float64) float64 {
	if !p.IsOpen || p.EntryPrice == 0 {
		return 0.0
	}
	
	var profit float64
	switch p.Direction {
	case PositionLong:
		profit = (currentPrice - p.EntryPrice) / p.EntryPrice * 100.0
	case PositionShort:
		profit = (p.EntryPrice - currentPrice) / p.EntryPrice * 100.0
	default:
		profit = 0.0
	}
	
	return profit
}

// IsStopHit checks if the stop loss has been hit
func (p Position) IsStopHit(currentPrice float64) bool {
	if !p.IsOpen {
		return false
	}
	
	switch p.Direction {
	case PositionLong:
		return currentPrice <= p.StopLoss
	case PositionShort:
		return currentPrice >= p.StopLoss
	default:
		return false
	}
}

// ActiveZone represents a zone being monitored
type ActiveZone struct {
	Type         ZoneType `json:"type"`
	Active       bool     `json:"active"`
	EntryTime    int64    `json:"entry_time"`
	TriggerTime  int64    `json:"trigger_time,omitempty"`
	LastAdjustment int64  `json:"last_adjustment,omitempty"`
}

// ZoneEvent represents an event from zone monitoring
type ZoneEvent struct {
	Type            string  `json:"type"`
	ZoneType        ZoneType `json:"zone_type"`
	IsInverse       bool    `json:"is_inverse,omitempty"`
	RequiresProfit  bool    `json:"requires_profit,omitempty"`
	ProfitThreshold float64 `json:"profit_threshold,omitempty"`
	CurrentProfit   float64 `json:"current_profit,omitempty"`
	Timestamp       int64   `json:"timestamp"`
}

// EngineConfig holds configuration for the temporal engine
type EngineConfig struct {
	WindowSize       int                    `json:"window_size"`
	AntiLookAhead    bool                  `json:"anti_lookahead"`
	TrailingStop     TrailingStopConfig    `json:"trailing_stop"`
	AdjustmentGrid   []AdjustmentLevel     `json:"adjustment_grid"`
	Zones            ZoneConfig            `json:"zones"`
}

// TrailingStopConfig defines trailing stop configuration
type TrailingStopConfig struct {
	TrendPercent        float64 `json:"trend_percent"`
	CounterTrendPercent float64 `json:"counter_trend_percent"`
}

// AdjustmentLevel defines profit-based stop adjustments
type AdjustmentLevel struct {
	ProfitMin       float64 `json:"profit_min"`
	ProfitMax       float64 `json:"profit_max"`
	TrailingPercent float64 `json:"trailing_percent"`
}

// ZoneConfig defines zone monitoring configuration
type ZoneConfig struct {
	CCIInverse   ZoneSettings `json:"cci_inverse"`
	MACDInverse  ZoneSettings `json:"macd_inverse"`
	DICounter    ZoneSettings `json:"di_counter"`
}

// ZoneSettings defines settings for a specific zone type
type ZoneSettings struct {
	Enabled         bool    `json:"enabled"`
	Monitoring      string  `json:"monitoring"` // "continuous" or "event"
	ProfitThreshold float64 `json:"profit_threshold,omitempty"`
}

// PerformanceMetrics holds performance metrics for the engine
type PerformanceMetrics struct {
	CyclesExecuted          int     `json:"cycles_executed"`
	PositionsOpened         int     `json:"positions_opened"`
	PositionsClosed         int     `json:"positions_closed"`
	ZoneActivations         int     `json:"zone_activations"`
	StopAdjustments         int     `json:"stop_adjustments"`
	AntiLookAheadViolations int     `json:"anti_lookahead_violations"`
	AverageLatencyMs        float64 `json:"average_latency_ms"`
	MaxLatencyMs            float64 `json:"max_latency_ms"`
	MemoryUsageMB           float64 `json:"memory_usage_mb"`
}

// Common errors for the engine
var (
	ErrLookAheadDetected     = errors.New("look-ahead detected: accessing future data")
	ErrInsufficientData      = errors.New("insufficient data for calculation")
	ErrPositionAlreadyOpen   = errors.New("position already open")
	ErrNoOpenPosition        = errors.New("no open position")
	ErrInvalidConfiguration  = errors.New("invalid configuration")
	ErrInvalidTimestamp      = errors.New("invalid timestamp")
	ErrCalculationTimeout    = errors.New("calculation timeout")
)

// ValidateConfiguration validates engine configuration
func ValidateConfiguration(config EngineConfig) error {
	if config.WindowSize <= 0 {
		return errors.New("window_size must be positive")
	}
	
	if config.TrailingStop.TrendPercent <= 0 || config.TrailingStop.TrendPercent > 100 {
		return errors.New("trend_percent must be between 0 and 100")
	}
	
	if config.TrailingStop.CounterTrendPercent <= 0 || config.TrailingStop.CounterTrendPercent > 100 {
		return errors.New("counter_trend_percent must be between 0 and 100")
	}
	
	// Validate adjustment grid is properly ordered (allow touching boundaries)
	for i := 1; i < len(config.AdjustmentGrid); i++ {
		if config.AdjustmentGrid[i].ProfitMin < config.AdjustmentGrid[i-1].ProfitMax {
			return errors.New("adjustment_grid levels must not overlap")
		}
	}
	
	return nil
}

// IsMarkerTimestamp checks if timestamp represents a 5m candle marker
// Markers occur at 00:00, 00:05, 00:10, 00:15, etc. (every 5 minutes with seconds = 0)
func IsMarkerTimestamp(timestamp int64) bool {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1000000).UTC()
	
	// Must be at exact minute start (seconds = 0)
	if t.Second() != 0 {
		return false
	}
	
	// Must be on 5-minute boundary (00, 05, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55)
	return t.Minute()%5 == 0
}

// GetTimeframeAlignment checks alignment with specific timeframes
func GetTimeframeAlignment(timestamp int64, timeframes []string) []string {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1000000).UTC()
	
	if t.Second() != 0 {
		return nil
	}
	
	var aligned []string
	minute := t.Minute()
	hour := t.Hour()
	
	for _, tf := range timeframes {
		switch tf {
		case "5m":
			if minute%5 == 0 {
				aligned = append(aligned, tf)
			}
		case "15m":
			if minute%15 == 0 {
				aligned = append(aligned, tf)
			}
		case "1h":
			if minute == 0 {
				aligned = append(aligned, tf)
			}
		case "4h":
			if minute == 0 && hour%4 == 0 {
				aligned = append(aligned, tf)
			}
		}
	}
	
	return aligned
}

// DefaultEngineConfig returns a default configuration
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		WindowSize:    300,
		AntiLookAhead: true,
		TrailingStop: TrailingStopConfig{
			TrendPercent:        2.0,
			CounterTrendPercent: 1.5,
		},
		AdjustmentGrid: []AdjustmentLevel{
			{ProfitMin: 0, ProfitMax: 5, TrailingPercent: 2.0},
			{ProfitMin: 5, ProfitMax: 10, TrailingPercent: 1.5},
			{ProfitMin: 10, ProfitMax: 20, TrailingPercent: 1.0},
			{ProfitMin: 20, ProfitMax: 100, TrailingPercent: 0.5},
		},
		Zones: ZoneConfig{
			CCIInverse: ZoneSettings{
				Enabled:    true,
				Monitoring: "continuous",
			},
			MACDInverse: ZoneSettings{
				Enabled:         true,
				Monitoring:      "event",
				ProfitThreshold: 0.5,
			},
			DICounter: ZoneSettings{
				Enabled:         true,
				Monitoring:      "event",
				ProfitThreshold: 1.0,
			},
		},
	}
}
