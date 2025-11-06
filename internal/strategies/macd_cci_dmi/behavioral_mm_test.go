package macd_cci_dmi

import (
	"testing"
	"time"
)

func TestNewBehavioralMoneyManager(t *testing.T) {
	config := DefaultBehavioralConfig()
	bmm, err := NewBehavioralMoneyManager(config)
	if err != nil {
		t.Fatalf("Failed to create behavioral MM: %v", err)
	}
	
	if bmm == nil {
		t.Fatal("Behavioral MM should not be nil")
	}
	
	status := bmm.GetStatus()
	if status.IsActive {
		t.Error("Behavioral MM should not be active initially")
	}
}

func TestBehavioralMMInvalidConfig(t *testing.T) {
	invalidConfigs := []BehavioralConfig{
		func() BehavioralConfig {
			config := DefaultBehavioralConfig()
			config.TrendTrailingPercent = -1.0 // Invalid
			return config
		}(),
		func() BehavioralConfig {
			config := DefaultBehavioralConfig()
			config.MaxAdjustmentPercent = 10.0 // Too high
			return config
		}(),
		func() BehavioralConfig {
			config := DefaultBehavioralConfig()
			config.MinTrailingPercent = 5.0
			config.MaxTrailingPercent = 2.0 // Min > Max
			return config
		}(),
	}
	
	for i, config := range invalidConfigs {
		_, err := NewBehavioralMoneyManager(config)
		if err == nil {
			t.Errorf("Test %d: Expected error for invalid config", i)
		}
	}
}

func TestPositionOpening(t *testing.T) {
	config := DefaultBehavioralConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	indicators := IndicatorSnapshot{
		Timestamp:  time.Now().UTC(),
		MACDValue:  0.5,
		MACDSignal: 0.3,
		CCIValue:   -150, // Oversold
		DIPlus:     25,
		DIMinus:    15,   // Trend
		ADXValue:   20,
	}
	
	err := bmm.OnPositionOpened("LONG", 50000.0, indicators)
	if err != nil {
		t.Errorf("Failed to open position: %v", err)
	}
	
	status := bmm.GetStatus()
	if !status.IsActive {
		t.Error("Behavioral MM should be active after position opening")
	}
	
	if status.PositionState.Direction != "LONG" {
		t.Errorf("Expected LONG direction, got %s", status.PositionState.Direction)
	}
	
	if status.PositionState.EntryPrice != 50000.0 {
		t.Errorf("Expected entry price 50000, got %.2f", status.PositionState.EntryPrice)
	}
	
	// Should use trend trailing (2%) since DI+ > DI-
	expectedTrailing := config.TrendTrailingPercent
	if status.PositionState.CurrentTrailingPercent != expectedTrailing {
		t.Errorf("Expected %.2f%% trailing, got %.2f%%", 
			expectedTrailing, status.PositionState.CurrentTrailingPercent)
	}
}

func TestTrailingStopModes(t *testing.T) {
	testCases := []struct {
		name           string
		mode           TrailingStopMode
		diPlus         float64
		diMinus        float64
		expectedBase   float64
	}{
		{
			name:         "Fixed mode",
			mode:         TrailingModeFixed,
			diPlus:       25,
			diMinus:      15,
			expectedBase: 2.0, // Always trend percent
		},
		{
			name:         "Adaptive mode - Trend",
			mode:         TrailingModeAdaptive,
			diPlus:       25,
			diMinus:      15,
			expectedBase: 2.0, // Trend percent
		},
		{
			name:         "Adaptive mode - Counter-trend",
			mode:         TrailingModeAdaptive,
			diPlus:       15,
			diMinus:      25,
			expectedBase: 1.5, // Counter-trend percent
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultBehavioralConfig()
			config.TrailingMode = tc.mode
			
			bmm, _ := NewBehavioralMoneyManager(config)
			
			indicators := IndicatorSnapshot{
				DIPlus:  tc.diPlus,
				DIMinus: tc.diMinus,
				ADXValue: 20,
			}
			
			trailing := bmm.calculateInitialTrailingStop(indicators)
			if abs(trailing-tc.expectedBase) > 0.01 {
				t.Errorf("Expected %.2f%% trailing, got %.2f%%", tc.expectedBase, trailing)
			}
		})
	}
}

func TestCCIZoneInverseDetection(t *testing.T) {
	config := DefaultBehavioralConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open LONG position in oversold CCI
	indicators := IndicatorSnapshot{
		CCIValue: -150, // Oversold
		DIPlus:   25,
		DIMinus:  15,
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, indicators)
	
	// Update with CCI in overbought (inverse zone)
	currentIndicators := IndicatorSnapshot{
		CCIValue: 150, // Overbought
	}
	
	// Test detection
	detected := bmm.detectCCIZoneInverse(currentIndicators, indicators)
	if !detected {
		t.Error("Should detect CCI zone inverse from oversold to overbought")
	}
	
	// Test that same zone doesn't trigger
	sameZoneIndicators := IndicatorSnapshot{
		CCIValue: -120, // Still oversold
	}
	
	detected = bmm.detectCCIZoneInverse(sameZoneIndicators, indicators)
	if detected {
		t.Error("Should not detect CCI zone inverse for same zone")
	}
}

func TestMACDInverseDetection(t *testing.T) {
	config := DefaultBehavioralConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open LONG position
	bmm.OnPositionOpened("LONG", 50000.0, IndicatorSnapshot{})
	
	// Test MACD cross down (inverse for LONG)
	previous := IndicatorSnapshot{
		MACDValue:  0.5,
		MACDSignal: 0.3, // MACD above signal
	}
	
	current := IndicatorSnapshot{
		MACDValue:  0.2,
		MACDSignal: 0.3, // MACD below signal (cross down)
	}
	
	detected := bmm.detectMACDInverse(current, previous)
	if !detected {
		t.Error("Should detect MACD inverse (cross down) for LONG position")
	}
	
	// Test no cross
	noCross := IndicatorSnapshot{
		MACDValue:  0.6,
		MACDSignal: 0.3, // Still above
	}
	
	detected = bmm.detectMACDInverse(noCross, previous)
	if detected {
		t.Error("Should not detect MACD inverse when no cross occurred")
	}
}

func TestProfitGridAdjustments(t *testing.T) {
	config := DefaultBehavioralConfig()
	config.ProfitAdjustmentLevels = []ProfitAdjustmentLevel{
		{ProfitThresholdPercent: 1.0, AdjustmentPercent: 0.2},
		{ProfitThresholdPercent: 2.0, AdjustmentPercent: 0.3},
	}
	
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open position
	entryPrice := 50000.0
	bmm.OnPositionOpened("LONG", entryPrice, IndicatorSnapshot{})
	
	// Test profit level adjustments
	testPrices := []struct {
		price          float64
		expectedAdjust bool
		description    string
	}{
		{50250.0, false, "0.5% profit - below first threshold"},
		{50500.0, true, "1% profit - first threshold triggered"},
		{51000.0, true, "2% profit - second threshold triggered"},
	}
	
	for _, test := range testPrices {
		t.Run(test.description, func(t *testing.T) {
			initialAdjustments := bmm.positionState.TotalAdjustments
			
			bmm.processProfitGridAdjustments(test.price)
			
			finalAdjustments := bmm.positionState.TotalAdjustments
			adjustmentMade := finalAdjustments > initialAdjustments
			
			if adjustmentMade != test.expectedAdjust {
				t.Errorf("Expected adjustment: %t, got: %t for price %.2f", 
					test.expectedAdjust, adjustmentMade, test.price)
			}
		})
	}
}

func TestEarlyExitDecision(t *testing.T) {
	config := DefaultBehavioralConfig()
	config.EnableEarlyExit = true
	config.EarlyExitOnMACDInverse = true
	config.MinProfitForEarlyExit = 0.5
	
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open position
	bmm.OnPositionOpened("LONG", 50000.0, IndicatorSnapshot{})
	
	// Test early exit with insufficient profit
	currentPrice := 50100.0 // 0.2% profit, below minimum 0.5%
	previous := IndicatorSnapshot{MACDValue: 0.5, MACDSignal: 0.3}
	current := IndicatorSnapshot{MACDValue: 0.2, MACDSignal: 0.3} // MACD cross down
	
	decision := bmm.evaluateEarlyExit(currentPrice, current, previous)
	if decision.ShouldExit {
		t.Error("Should not exit with insufficient profit")
	}
	
	// Test early exit with sufficient profit and MACD inverse
	currentPrice = 50300.0 // 0.6% profit, above minimum
	decision = bmm.evaluateEarlyExit(currentPrice, current, previous)
	if !decision.ShouldExit {
		t.Error("Should exit with sufficient profit and MACD inverse")
	}
}

func TestStopTrigger(t *testing.T) {
	config := DefaultBehavioralConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open LONG position
	entryPrice := 50000.0
	bmm.OnPositionOpened("LONG", entryPrice, IndicatorSnapshot{})
	
	// Get initial stop price
	initialStop := bmm.GetCurrentStopPrice()
	
	// Test prices
	aboveStop := initialStop + 100
	belowStop := initialStop - 100
	
	// Price above stop should not trigger
	if bmm.ShouldTriggerStop(aboveStop) {
		t.Error("Price above stop should not trigger stop loss")
	}
	
	// Price below stop should trigger
	if !bmm.ShouldTriggerStop(belowStop) {
		t.Error("Price below stop should trigger stop loss")
	}
}

func TestDynamicAdjustments(t *testing.T) {
	config := DefaultBehavioralConfig()
	config.EnableDynamicAdjustments = true
	
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Track adjustment calls
	adjustmentCalled := false
	bmm.SetCallbacks(
		func(adj TrailingAdjustment) {
			adjustmentCalled = true
		},
		nil,
	)
	
	// Open position
	bmm.OnPositionOpened("LONG", 50000.0, IndicatorSnapshot{
		CCIValue: -150, // Oversold entry
	})
	
	// Update with profitable position and CCI inverse
	currentPrice := 50500.0 // 1% profit
	current := IndicatorSnapshot{
		CCIValue: 150, // Overbought (inverse)
	}
	previous := IndicatorSnapshot{
		CCIValue: -150, // Oversold
	}
	
	bmm.processDynamicAdjustments(currentPrice, current, previous)
	
	if !adjustmentCalled {
		t.Error("Expected adjustment callback to be called for CCI zone inverse")
	}
}

// Helper function for float comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
