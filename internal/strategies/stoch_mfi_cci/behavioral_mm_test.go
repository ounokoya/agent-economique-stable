package stoch_mfi_cci

import (
	"testing"
	
	"agent-economique/internal/indicators"
)

func TestNewBehavioralMoneyManager(t *testing.T) {
	config := DefaultStrategyConfig()
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

func TestBehavioralMMPositionOpening(t *testing.T) {
	config := DefaultStrategyConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Create test indicator results
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	err := bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
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
	
	if status.PositionState.SignalStrength != SignalPremium {
		t.Errorf("Expected premium signal strength, got %s", status.PositionState.SignalStrength)
	}
}

func TestSTOCHInverseDetection(t *testing.T) {
	config := DefaultStrategyConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open LONG position in oversold
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	
	// Update with STOCH in overbought (inverse)
	inverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought, // Inverse zone
		},
		MFI: &indicators.MFIValues{
			Value: 50.0,
			Zone:  indicators.MFINeutral,
		},
		CCI: &indicators.CCIValues{
			Value: 50.0,
			Zone:  indicators.CCINormal,
		},
	}
	
	err := bmm.OnTickUpdate(51000.0, inverseResults)
	if err != nil {
		t.Errorf("Failed to process tick update: %v", err)
	}
	
	status := bmm.GetStatus()
	if status.MonitoringState != StateSTOCHInverse {
		t.Errorf("Expected STOCH_INVERSE state, got %s", status.MonitoringState)
	}
}

func TestTripleInverseDetection(t *testing.T) {
	config := DefaultStrategyConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Open LONG position with all indicators oversold
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	
	// Update with all indicators overbought (triple inverse)
	tripleInverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 85.0,
			Zone:  indicators.MFIOverbought,
		},
		CCI: &indicators.CCIValues{
			Value: 120.0,
			Zone:  indicators.CCIOverbought,
		},
	}
	
	err := bmm.OnTickUpdate(52000.0, tripleInverseResults)
	if err != nil {
		t.Errorf("Failed to process tick update: %v", err)
	}
	
	status := bmm.GetStatus()
	if status.MonitoringState != StateTripleInverse {
		t.Errorf("Expected TRIPLE_INVERSE state, got %s", status.MonitoringState)
	}
}

func TestTrailingAdjustments(t *testing.T) {
	config := DefaultStrategyConfig()
	config.EnableDynamicAdjustments = true
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Track adjustments
	adjustmentCount := 0
	bmm.SetCallbacks(
		func(adj TrailingAdjustment) {
			adjustmentCount++
		},
		nil,
		nil,
	)
	
	// Open position
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	initialStop := bmm.GetCurrentStopPrice()
	
	// Trigger STOCH inverse
	inverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 85.0,
			Zone:  indicators.MFIOverbought,
		},
		CCI: &indicators.CCIValues{
			Value: 120.0,
			Zone:  indicators.CCIOverbought,
		},
	}
	
	err := bmm.OnTickUpdate(52000.0, inverseResults)
	if err != nil {
		t.Errorf("Failed to process tick update: %v", err)
	}
	
	finalStop := bmm.GetCurrentStopPrice()
	
	// Stop should be tighter (higher for LONG position)
	if finalStop <= initialStop {
		t.Errorf("Expected tighter stop, got initial: %.2f, final: %.2f", initialStop, finalStop)
	}
	
	if adjustmentCount == 0 {
		t.Error("Expected at least one adjustment to be made")
	}
}

func TestEarlyExitDecision(t *testing.T) {
	config := DefaultStrategyConfig()
	config.EnableEarlyExit = true
	config.TripleInverseEarlyExit = true
	config.MinProfitForEarlyExit = 1.0
	
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Track early exit calls
	earlyExitCalled := false
	bmm.SetCallbacks(
		nil,
		func(decision EarlyExitDecision) error {
			earlyExitCalled = decision.ShouldExit
			return nil
		},
		nil,
	)
	
	// Open position
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	
	// Update with profitable price and triple inverse
	profitablePrice := 51000.0 // 2% profit
	tripleInverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 85.0,
			Zone:  indicators.MFIOverbought,
		},
		CCI: &indicators.CCIValues{
			Value: 120.0,
			Zone:  indicators.CCIOverbought,
		},
	}
	
	err := bmm.OnTickUpdate(profitablePrice, tripleInverseResults)
	if err != nil {
		t.Errorf("Failed to process tick update: %v", err)
	}
	
	// Early exit should be triggered due to triple inverse + profit
	if !earlyExitCalled {
		t.Error("Expected early exit to be triggered")
	}
}

func TestMonitoringStateTransitions(t *testing.T) {
	config := DefaultStrategyConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Track state changes
	var stateChanges []MonitoringState
	bmm.SetCallbacks(
		nil,
		nil,
		func(newState MonitoringState) {
			stateChanges = append(stateChanges, newState)
		},
	)
	
	// Open position (should start in NORMAL state)
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	
	// Trigger STOCH inverse (should change to STOCH_INVERSE)
	stochInverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 50.0,
			Zone:  indicators.MFINeutral,
		},
		CCI: &indicators.CCIValues{
			Value: 50.0,
			Zone:  indicators.CCINormal,
		},
	}
	
	bmm.OnTickUpdate(51000.0, stochInverseResults)
	
	// Trigger triple inverse (should change to TRIPLE_INVERSE)
	tripleInverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 85.0,
			Zone:  indicators.MFIOverbought,
		},
		CCI: &indicators.CCIValues{
			Value: 120.0,
			Zone:  indicators.CCIOverbought,
		},
	}
	
	bmm.OnTickUpdate(52000.0, tripleInverseResults)
	
	// Should have recorded state transitions
	expectedStates := []MonitoringState{StateSTOCHInverse, StateTripleInverse}
	if len(stateChanges) != len(expectedStates) {
		t.Errorf("Expected %d state changes, got %d", len(expectedStates), len(stateChanges))
	}
	
	for i, expected := range expectedStates {
		if i < len(stateChanges) && stateChanges[i] != expected {
			t.Errorf("Expected state change %d to be %s, got %s", i, expected, stateChanges[i])
		}
	}
}

func TestIsMonitoringActive(t *testing.T) {
	config := DefaultStrategyConfig()
	bmm, _ := NewBehavioralMoneyManager(config)
	
	// Should not be active initially
	if bmm.IsMonitoringActive() {
		t.Error("Monitoring should not be active initially")
	}
	
	// Open position
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    15.0,
			D:    18.0,
			Zone: indicators.StochOversold,
		},
		MFI: &indicators.MFIValues{
			Value: 15.0,
			Zone:  indicators.MFIOversold,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0,
			Zone:  indicators.CCIOversold,
		},
	}
	
	bmm.OnPositionOpened("LONG", 50000.0, SignalPremium, results, "TREND")
	
	// Still should not be active (normal state)
	if bmm.IsMonitoringActive() {
		t.Error("Monitoring should not be active in normal state")
	}
	
	// Trigger STOCH inverse
	inverseResults := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:    85.0,
			D:    82.0,
			Zone: indicators.StochOverbought,
		},
		MFI: &indicators.MFIValues{
			Value: 50.0,
			Zone:  indicators.MFINeutral,
		},
		CCI: &indicators.CCIValues{
			Value: 50.0,
			Zone:  indicators.CCINormal,
		},
	}
	
	bmm.OnTickUpdate(51000.0, inverseResults)
	
	// Now should be active
	if !bmm.IsMonitoringActive() {
		t.Error("Monitoring should be active after STOCH inverse")
	}
}
