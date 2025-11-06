package engine

import (
	"testing"
	"time"
)

func TestMMIntegrationBasic(t *testing.T) {
	// Test que les fonctions d'intégration existent et fonctionnent
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test emergency stop callback
	err = engine.executeEmergencyStopAll()
	if err != nil {
		t.Errorf("Emergency stop failed: %v", err)
	}
}

func TestMMPositionValidation(t *testing.T) {
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test position validation
	err = engine.validatePositionWithMM(PositionLong, 50000.0)
	if err != nil {
		t.Errorf("Position validation failed: %v", err)
	}
}

func TestMMStatusFunctions(t *testing.T) {
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test status functions don't crash
	status := engine.getMMStatus()
	if status.IsActive != true {
		t.Errorf("Expected MM to be active initially")
	}
	
	// Test status summary printing (shouldn't crash)
	engine.printMMStatusSummary()
}

func TestTradeRecording(t *testing.T) {
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Create a mock position
	position := Position{
		Direction:  PositionLong,
		EntryPrice: 50000.0,
		EntryTime:  time.Now().Unix(),
	}
	
	// Test position notifications
	engine.notifyPositionOpened(1000.0)
	engine.notifyPositionClosed(position, 50500.0, "TAKE_PROFIT")
	
	// Verify trade was recorded
	if engine.moneyManager != nil {
		metrics := engine.moneyManager.GetCurrentMetrics()
		if metrics.TotalPositions != 1 {
			t.Errorf("Expected 1 recorded trade, got %d", metrics.TotalPositions)
		}
		expectedPnL := 500.0 // (50500 - 50000) * 1.0
		if metrics.TotalPnL != expectedPnL {
			t.Errorf("Expected %.2f total PnL, got %.2f", expectedPnL, metrics.TotalPnL)
		}
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test real-time PnL update with normal values
	err = engine.updateMMWithRealTimePnL()
	if err != nil {
		t.Errorf("Normal PnL update should not trigger circuit breaker: %v", err)
	}
	
	// Test that MM is still active
	if engine.moneyManager != nil && !engine.moneyManager.IsActive() {
		t.Error("Money Manager should be active after normal PnL update")
	}
}

func TestMMShutdown(t *testing.T) {
	config := DefaultEngineConfig()
	engine, err := NewTemporalEngine(BacktestMode, config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test graceful shutdown
	err = engine.shutdownMoneyManager()
	if err != nil {
		t.Errorf("MM shutdown failed: %v", err)
	}
}
