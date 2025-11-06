package money_management

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	
	auditLogger, err := NewAuditLogger(filepath.Join(tempDir, "audit"))
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Verify initial state
	if cb.IsActive() {
		t.Error("Circuit breaker should not be active initially")
	}
	
	state := cb.GetState()
	if state.TotalBreaches != 0 {
		t.Errorf("Expected 0 total breaches, got %d", state.TotalBreaches)
	}
}

func TestCircuitBreakerInvalidConfig(t *testing.T) {
	invalidConfigs := []BaseConfiguration{
		{DailyLimitPercent: 0},      // Invalid daily limit
		{DailyLimitPercent: 5, MonthlyLimitPercent: 0},  // Invalid monthly limit
		func() BaseConfiguration {   // Invalid spot amount
			config := DefaultBaseConfiguration()
			config.DailyLimitPercent = 5
			config.MonthlyLimitPercent = 15
			config.PositionSizing.SpotAmountUSDT = -100
			return config
		}(),
	}
	
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "test_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	for i, config := range invalidConfigs {
		_, err := NewCircuitBreaker(config, stateFile, auditLogger)
		if err == nil {
			t.Errorf("Test %d: Expected error for invalid config, got nil", i)
		}
	}
}

func TestDailyLimitBreach(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.DailyLimitPercent = 5.0  // -5% daily limit
	config.StartingCapital = 10000.0
	
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Mock emergency stop callback
	emergencyStopCalled := false
	cb.SetEmergencyStopCallback(func() error {
		emergencyStopCalled = true
		return nil
	})
	
	// Test cases
	testCases := []struct {
		name        string
		currentPnL  float64
		expectError bool
		expectStop  bool
	}{
		{"No breach at -4%", -400.0, false, false},
		{"Breach at -5%", -500.0, true, true},
		{"Breach at -6%", -600.0, true, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset for each test
			emergencyStopCalled = false
			cb.state.DailyBreakerActive = false
			
			err := cb.CheckDailyLimit(tc.currentPnL)
			
			if tc.expectError && err == nil {
				t.Error("Expected error for daily limit breach")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectStop && !emergencyStopCalled {
				t.Error("Expected emergency stop to be called")
			}
			if !tc.expectStop && emergencyStopCalled {
				t.Error("Emergency stop should not be called")
			}
		})
	}
}

func TestMonthlyLimitBreach(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.MonthlyLimitPercent = 15.0  // -15% monthly limit
	config.StartingCapital = 10000.0
	
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Mock emergency stop callback
	emergencyStopCalled := false
	cb.SetEmergencyStopCallback(func() error {
		emergencyStopCalled = true
		return nil
	})
	
	// Test cases
	testCases := []struct {
		name        string
		currentPnL  float64
		expectError bool
		expectStop  bool
	}{
		{"No breach at -14%", -1400.0, false, false},
		{"Breach at -15%", -1500.0, true, true},
		{"Breach at -20%", -2000.0, true, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset for each test
			emergencyStopCalled = false
			cb.state.MonthlyBreakerActive = false
			
			err := cb.CheckMonthlyLimit(tc.currentPnL)
			
			if tc.expectError && err == nil {
				t.Error("Expected error for monthly limit breach")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.expectStop && !emergencyStopCalled {
				t.Error("Expected emergency stop to be called")
			}
			if !tc.expectStop && emergencyStopCalled {
				t.Error("Emergency stop should not be called")
			}
		})
	}
}

func TestCircuitBreakerPersistence(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	// Create first circuit breaker and trigger breach
	cb1, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create first circuit breaker: %v", err)
	}
	
	cb1.SetEmergencyStopCallback(func() error { return nil })
	
	// Trigger daily breach
	err = cb1.CheckDailyLimit(-500.0) // -5% breach
	if err == nil {
		t.Error("Expected daily limit breach error")
	}
	
	// Verify state is active
	if !cb1.IsActive() {
		t.Error("Circuit breaker should be active after breach")
	}
	
	// Create second circuit breaker (simulating restart)
	cb2, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create second circuit breaker: %v", err)
	}
	
	// Verify state persisted
	if !cb2.IsActive() {
		t.Error("Circuit breaker state should persist after restart")
	}
	
	state := cb2.GetState()
	if state.TotalBreaches == 0 {
		t.Error("Total breaches should persist after restart")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Manually set daily breaker as active with past reset time
	cb.state.DailyBreakerActive = true
	cb.state.DailyResetTime = time.Now().UTC().Add(-time.Hour) // Past time
	
	// Trigger reset check
	if !cb.state.ShouldResetDaily() {
		t.Error("Should indicate daily reset is needed")
	}
	
	// Perform reset
	cb.resetDaily()
	
	// Verify reset
	if cb.IsActive() {
		t.Error("Circuit breaker should not be active after reset")
	}
}

func TestEmergencyStopCallback(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Test without callback - should handle gracefully
	err = cb.CheckDailyLimit(-500.0) // Trigger breach
	if err == nil {
		t.Error("Expected error when no emergency callback is set")
	}
	
	// Test with successful callback
	cb.SetEmergencyStopCallback(func() error { return nil })
	cb.state.DailyBreakerActive = false // Reset state
	
	err = cb.CheckDailyLimit(-500.0)
	if err != ErrDailyLimitBreached {
		t.Errorf("Expected ErrDailyLimitBreached, got %v", err)
	}
}

func TestConfigurationUpdate(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	// Update configuration
	newConfig := config
	newConfig.DailyLimitPercent = 3.0 // Stricter limit
	
	err = cb.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("Failed to update configuration: %v", err)
	}
	
	// Verify configuration updated
	if cb.config.DailyLimitPercent != 3.0 {
		t.Errorf("Expected daily limit 3.0, got %f", cb.config.DailyLimitPercent)
	}
	
	// Test invalid configuration
	invalidConfig := newConfig
	invalidConfig.DailyLimitPercent = -1.0
	
	err = cb.UpdateConfiguration(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid configuration")
	}
}

// Benchmark circuit breaker performance
func BenchmarkCircuitBreakerCheck(b *testing.B) {
	config := DefaultBaseConfiguration()
	tempDir := b.TempDir()
	stateFile := filepath.Join(tempDir, "circuit_breaker_state.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	cb, err := NewCircuitBreaker(config, stateFile, auditLogger)
	if err != nil {
		b.Fatalf("Failed to create circuit breaker: %v", err)
	}
	
	cb.SetEmergencyStopCallback(func() error { return nil })
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Check limit without triggering breach
		cb.CheckDailyLimit(-400.0) // -4%, below -5% limit
	}
}
