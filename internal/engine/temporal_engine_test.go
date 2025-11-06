package engine

import (
	"testing"
	"time"

	"agent-economique/internal/indicators"
)

func TestTemporalEngine_CalculateIndicators(t *testing.T) {
	// Create engine with minimal config
	config := EngineConfig{
		WindowSize:   50,
		AntiLookAhead: true,
		TrailingStop: TrailingStopConfig{
			TrendPercent:        2.0,
			CounterTrendPercent: 3.0,
		},
		AdjustmentGrid: []AdjustmentLevel{
			{ProfitMin: 0.5, ProfitMax: 2.0, TrailingPercent: 1.5},
		},
		Zones: ZoneConfig{
			CCIInverse: ZoneSettings{Enabled: true, Monitoring: "continuous"},
		},
	}
	
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Add test klines
	testKlines := generateEngineTestKlines(50)
	engine.historicalKlines = testKlines
	engine.currentTimestamp = time.Now().UnixMilli()
	
	// Test calculateIndicators
	response, err := engine.calculateIndicators()
	
	if err != nil {
		t.Errorf("calculateIndicators failed: %v", err)
	}
	
	if !response.Success {
		t.Errorf("Expected success, got: %v", response.Error)
	}
	
	if response.Results == nil {
		t.Error("Expected results not nil")
	}
}

func TestTemporalEngine_GetPositionContext(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test with no open position
	ctx := engine.getPositionContext()
	if ctx != nil {
		t.Error("Expected nil context for no open position")
	}
	
	// Open a position
	err = engine.positionManager.OpenPosition(PositionLong, 100.0, time.Now().UnixMilli(), "OVERSOLD")
	if err != nil {
		t.Fatalf("Failed to open position: %v", err)
	}
	
	// Add test klines for getCurrentPrice
	engine.historicalKlines = []Kline{
		{Close: 105.0, Timestamp: time.Now().UnixMilli()},
	}
	
	// Test with open position
	ctx = engine.getPositionContext()
	if ctx == nil {
		t.Error("Expected non-nil context for open position")
	}
	
	if !ctx.IsOpen {
		t.Error("Expected IsOpen=true")
	}
	
	if ctx.Direction != "LONG" {
		t.Errorf("Expected Direction=LONG, got %s", ctx.Direction)
	}
}

func TestTemporalEngine_GetCandleWindow(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Add test klines
	testKlines := generateEngineTestKlines(30)
	engine.historicalKlines = testKlines
	
	// Test conversion
	window := engine.getCandleWindow()
	
	if len(window) != len(testKlines) {
		t.Errorf("Expected %d klines, got %d", len(testKlines), len(window))
	}
	
	if len(window) > 0 {
		// Check first kline conversion
		if window[0].Timestamp != testKlines[0].Timestamp {
			t.Error("Timestamp conversion failed")
		}
		if window[0].Close != testKlines[0].Close {
			t.Error("Close price conversion failed")
		}
	}
}

func TestTemporalEngine_ConvertCCIZone(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test conversions
	tests := []struct {
		input    string
		expected indicators.CCIZone
	}{
		{"OVERSOLD", indicators.CCIOversold},
		{"OVERBOUGHT", indicators.CCIOverbought},
		{"NORMAL", indicators.CCINormal},
		{"UNKNOWN", indicators.CCINormal}, // Default
	}
	
	for _, tt := range tests {
		result := engine.convertCCIZone(tt.input)
		if result != tt.expected {
			t.Errorf("convertCCIZone(%s): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestTemporalEngine_GetCurrentPrice(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test with no data
	price := engine.getCurrentPrice()
	if price != 0.0 {
		t.Errorf("Expected 0.0 for no data, got %f", price)
	}
	
	// Test with klines
	engine.historicalKlines = []Kline{
		{Close: 100.5, Timestamp: time.Now().UnixMilli()},
	}
	price = engine.getCurrentPrice()
	if price != 100.5 {
		t.Errorf("Expected 100.5 from klines, got %f", price)
	}
	
	// Test with trades (should prefer klines)
	engine.historicalTrades = []Trade{
		{Price: 99.0, Timestamp: time.Now().UnixMilli()},
	}
	price = engine.getCurrentPrice()
	if price != 100.5 {
		t.Errorf("Expected 100.5 (klines priority), got %f", price)
	}
	
	// Test with only trades
	engine.historicalKlines = []Kline{}
	price = engine.getCurrentPrice()
	if price != 99.0 {
		t.Errorf("Expected 99.0 from trades, got %f", price)
	}
}

func TestTemporalEngine_ProcessStrategySignals(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Add test klines for price
	engine.historicalKlines = []Kline{
		{Close: 100.0, Timestamp: time.Now().UnixMilli()},
	}
	engine.currentTimestamp = time.Now().UnixMilli()
	
	// Test with high confidence signal
	signals := []indicators.StrategySignal{
		{
			Direction:  indicators.LongSignal,
			Type:       indicators.TrendSignal,
			Confidence: 0.8,
			Timestamp:  time.Now().UnixMilli(),
		},
	}
	
	// Should open position
	engine.processStrategySignals(signals)
	
	if !engine.positionManager.IsOpen() {
		t.Error("Expected position to be opened")
	}
	
	// Test with low confidence signal (new engine)
	engine2, _ := NewTemporalEngine(BacktestMode, config)
	engine2.historicalKlines = engine.historicalKlines
	engine2.currentTimestamp = engine.currentTimestamp
	
	lowConfidenceSignal := []indicators.StrategySignal{
		{
			Direction:  indicators.LongSignal,
			Confidence: 0.5, // Below 0.7 threshold
			Timestamp:  time.Now().UnixMilli(),
		},
	}
	
	engine2.processStrategySignals(lowConfidenceSignal)
	
	if engine2.positionManager.IsOpen() {
		t.Error("Expected position NOT to be opened for low confidence")
	}
}

func TestTemporalEngine_ProcessZoneEvents(t *testing.T) {
	config := createMinimalConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test with no open position
	events := []indicators.ZoneEvent{
		{Type: "ZONE_ACTIVATED", ZoneType: "CCI_INVERSE"},
	}
	
	// Should not crash
	engine.processZoneEvents(events)
	
	// Open position for zone events
	engine.historicalKlines = []Kline{{Close: 100.0}}
	err = engine.positionManager.OpenPosition(PositionLong, 100.0, time.Now().UnixMilli(), "OVERSOLD")
	if err != nil {
		t.Fatalf("Failed to open position: %v", err)
	}
	
	// Test CCI inverse event
	engine.processZoneEvents(events)
	
	// Should complete without error (adjustment tested in position_manager_test.go)
}

// Helper functions
func createMinimalConfig() EngineConfig {
	return EngineConfig{
		WindowSize:   50,
		AntiLookAhead: true,
		TrailingStop: TrailingStopConfig{
			TrendPercent:        2.0,
			CounterTrendPercent: 3.0,
		},
		AdjustmentGrid: []AdjustmentLevel{
			{ProfitMin: 0.5, ProfitMax: 2.0, TrailingPercent: 1.5},
		},
		Zones: ZoneConfig{
			CCIInverse: ZoneSettings{Enabled: true, Monitoring: "continuous"},
		},
	}
}

func generateEngineTestKlines(count int) []Kline {
	klines := make([]Kline, count)
	baseTime := time.Now().UnixMilli()
	basePrice := 100.0
	
	for i := 0; i < count; i++ {
		price := basePrice + float64(i%10-5)
		klines[i] = Kline{
			Timestamp: baseTime + int64(i*60000),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price + 0.1,
			Volume:    1000 + float64(i%100),
		}
	}
	
	return klines
}
