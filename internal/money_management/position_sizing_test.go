package money_management

import (
	"path/filepath"
	"testing"
)

func TestNewPositionSizer(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	auditLogger, err := NewAuditLogger(filepath.Join(tempDir, "audit"))
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	// Verify initial configuration
	currentConfig := ps.GetConfiguration()
	if currentConfig.PositionSizing.Mode != PositionSizingFixed {
		t.Errorf("Expected default mode to be fixed, got %s", currentConfig.PositionSizing.Mode)
	}
}

func TestPositionSizingFixed(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.PositionSizing.SpotAmountUSDT = 1000.0
	config.PositionSizing.FuturesAmountUSDT = 500.0
	
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	testCases := []struct {
		name            string
		request         PositionSizingRequest
		expectValid     bool
		expectedQuantity float64
		expectedValue   float64
	}{
		{
			name: "Spot BTC Fixed",
			request: PositionSizingRequest{
				Symbol:           "BTC-USDT",
				Price:            50000.0,
				PositionType:     "spot",
				AvailableBalance: 2000.0,
				Leverage:         1,
			},
			expectValid:      true,
			expectedQuantity: 0.02,        // 1000 / 50000
			expectedValue:    1000.0,      // Fixed amount
		},
		{
			name: "Futures ETH Fixed",
			request: PositionSizingRequest{
				Symbol:           "ETH-USDT",
				Price:            3000.0,
				PositionType:     "futures",
				AvailableBalance: 100.0,
				Leverage:         10,
			},
			expectValid:      true,
			expectedQuantity: 1.666667, // (500 * 10) / 3000
			expectedValue:    5000.0,   // 500 * 10 leverage
		},
		{
			name: "Insufficient Balance",
			request: PositionSizingRequest{
				Symbol:           "BTC-USDT",
				Price:            50000.0,
				PositionType:     "spot",
				AvailableBalance: 500.0, // Less than required 1000
				Leverage:         1,
			},
			expectValid: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ps.CalculatePositionSize(tc.request)
			
			if result.IsValid != tc.expectValid {
				t.Errorf("Expected validity %t, got %t. Error: %s", tc.expectValid, result.IsValid, result.ValidationError)
			}
			
			if tc.expectValid {
				// Check quantity (with tolerance for floating point)
				tolerance := 0.01
				if abs(result.Quantity-tc.expectedQuantity) > tolerance {
					t.Errorf("Expected quantity ~%.6f, got %.6f", tc.expectedQuantity, result.Quantity)
				}
				
				// Check notional value
				if abs(result.NotionalValue-tc.expectedValue) > tolerance {
					t.Errorf("Expected notional value ~%.2f, got %.2f", tc.expectedValue, result.NotionalValue)
				}
			}
		})
	}
}

func TestPositionSizingPercentage(t *testing.T) {
	config := DefaultPercentageConfiguration(10000.0) // 10k USDT capital
	config.PositionSizing.SpotPercentage = 10.0      // 10% per spot trade
	config.PositionSizing.FuturesPercentage = 5.0    // 5% per futures trade
	
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	testCases := []struct {
		name             string
		request          PositionSizingRequest
		expectValid      bool
		expectedQuantity float64
		expectedValue    float64
	}{
		{
			name: "Spot BTC Percentage",
			request: PositionSizingRequest{
				Symbol:           "BTC-USDT",
				Price:            50000.0,
				PositionType:     "spot",
				AvailableBalance: 2000.0,
				Leverage:         1,
			},
			expectValid:      true,
			expectedQuantity: 0.02,   // (10000 * 0.10) / 50000 = 1000 / 50000 = 0.02
			expectedValue:    1000.0, // 10% of 10k capital
		},
		{
			name: "Futures ETH Percentage",
			request: PositionSizingRequest{
				Symbol:           "ETH-USDT",
				Price:            2000.0,
				PositionType:     "futures",
				AvailableBalance: 100.0,
				Leverage:         10,
			},
			expectValid:      true,
			expectedQuantity: 2.5,    // (10000 * 0.05 * 10) / 2000 = 5000 / 2000 = 2.5
			expectedValue:    5000.0, // 5% of capital * 10x leverage
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ps.CalculatePositionSize(tc.request)
			
			if result.IsValid != tc.expectValid {
				t.Errorf("Expected validity %t, got %t. Error: %s", tc.expectValid, result.IsValid, result.ValidationError)
			}
			
			if tc.expectValid {
				// Check quantity (with tolerance)
				tolerance := 0.01
				if abs(result.Quantity-tc.expectedQuantity) > tolerance {
					t.Errorf("Expected quantity ~%.6f, got %.6f", tc.expectedQuantity, result.Quantity)
				}
				
				// Check notional value
				if abs(result.NotionalValue-tc.expectedValue) > tolerance {
					t.Errorf("Expected notional value ~%.2f, got %.2f", tc.expectedValue, result.NotionalValue)
				}
			}
		})
	}
}

func TestPositionSizingLimits(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.PositionSizing.MaxPositionSize = 2000.0 // Max 2000 USDT
	config.PositionSizing.MinPositionSize = 100.0  // Min 100 USDT
	config.PositionSizing.SpotAmountUSDT = 3000.0  // Higher than max - should be scaled down
	
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	// Test max limit enforcement
	result := ps.CalculatePositionSize(PositionSizingRequest{
		Symbol:           "BTC-USDT",
		Price:            50000.0,
		PositionType:     "spot",
		AvailableBalance: 5000.0,
		Leverage:         1,
	})
	
	if !result.IsValid {
		t.Fatalf("Expected valid result, got error: %s", result.ValidationError)
	}
	
	// Should be scaled down to max position size
	if result.NotionalValue > config.PositionSizing.MaxPositionSize {
		t.Errorf("Position size %.2f exceeds max %.2f", result.NotionalValue, config.PositionSizing.MaxPositionSize)
	}
	
	// Test min limit enforcement
	config.PositionSizing.SpotAmountUSDT = 50.0 // Lower than min
	ps.UpdateConfiguration(config)
	
	result = ps.CalculatePositionSize(PositionSizingRequest{
		Symbol:           "BTC-USDT",
		Price:            50000.0,
		PositionType:     "spot",
		AvailableBalance: 5000.0,
		Leverage:         1,
	})
	
	if result.IsValid {
		t.Error("Expected invalid result due to min position size")
	}
}

func TestPositionCounters(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.MaxDailyPositions = 3
	config.MaxConcurrentTrades = 2
	
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	request := PositionSizingRequest{
		Symbol:           "BTC-USDT",
		Price:            50000.0,
		PositionType:     "spot",
		AvailableBalance: 5000.0,
		Leverage:         1,
	}
	
	// First two positions should succeed
	for i := 0; i < 2; i++ {
		result := ps.CalculatePositionSize(request)
		if !result.IsValid {
			t.Errorf("Position %d should be valid: %s", i+1, result.ValidationError)
		}
		ps.AddPosition()
	}
	
	// Third position should fail (concurrent limit)
	result := ps.CalculatePositionSize(request)
	if result.IsValid {
		t.Error("Position should be invalid due to concurrent limit")
	}
	
	// Remove one position
	ps.RemovePosition()
	
	// Now should succeed again
	result = ps.CalculatePositionSize(request)
	if !result.IsValid {
		t.Errorf("Position should be valid after removing one: %s", result.ValidationError)
	}
	ps.AddPosition()
	
	// Fourth position should fail (daily limit)
	result = ps.CalculatePositionSize(request)
	if result.IsValid {
		t.Error("Position should be invalid due to daily limit")
	}
}

func TestCapitalUpdate(t *testing.T) {
	config := DefaultPercentageConfiguration(10000.0)
	config.PositionSizing.SpotPercentage = 10.0 // 10%
	
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	request := PositionSizingRequest{
		Symbol:           "BTC-USDT",
		Price:            50000.0,
		PositionType:     "spot",
		AvailableBalance: 5000.0,
		Leverage:         1,
	}
	
	// Initial calculation with 10k capital
	result1 := ps.CalculatePositionSize(request)
	expectedQuantity1 := 0.02 // (10000 * 0.10) / 50000
	
	if abs(result1.Quantity-expectedQuantity1) > 0.01 {
		t.Errorf("Expected initial quantity ~%.6f, got %.6f", expectedQuantity1, result1.Quantity)
	}
	
	// Update capital to 20k
	err = ps.UpdateCapital(20000.0)
	if err != nil {
		t.Fatalf("Failed to update capital: %v", err)
	}
	
	// Recalculate with new capital
	result2 := ps.CalculatePositionSize(request)
	expectedQuantity2 := 0.04 // (20000 * 0.10) / 50000
	
	if abs(result2.Quantity-expectedQuantity2) > 0.01 {
		t.Errorf("Expected updated quantity ~%.6f, got %.6f", expectedQuantity2, result2.Quantity)
	}
}

func TestSymbolPrecision(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	ps, err := NewPositionSizer(config, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create position sizer: %v", err)
	}
	
	testCases := []struct {
		symbol            string
		expectedPrecision int
	}{
		{"BTC-USDT", 8},
		{"ETH-USDT", 6},
		{"SOL-USDT", 4},
		{"SUI-USDT", 3},
		{"UNKNOWN-USDT", 3}, // Default
	}
	
	for _, tc := range testCases {
		precision := ps.getSymbolPrecision(tc.symbol)
		if precision != tc.expectedPrecision {
			t.Errorf("Symbol %s: expected precision %d, got %d", tc.symbol, tc.expectedPrecision, precision)
		}
	}
}

func TestConfigurationValidation(t *testing.T) {
	tempDir := t.TempDir()
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	invalidConfigs := []BaseConfiguration{
		// Invalid position sizing mode
		func() BaseConfiguration {
			config := DefaultBaseConfiguration()
			config.PositionSizing.Mode = "invalid"
			return config
		}(),
		
		// Invalid percentages
		func() BaseConfiguration {
			config := DefaultPercentageConfiguration(10000)
			config.PositionSizing.SpotPercentage = 150.0 // > 100%
			return config
		}(),
		
		// Invalid limits
		func() BaseConfiguration {
			config := DefaultBaseConfiguration()
			config.PositionSizing.MaxPositionSize = 100.0
			config.PositionSizing.MinPositionSize = 200.0 // Min > Max
			return config
		}(),
	}
	
	for i, config := range invalidConfigs {
		_, err := NewPositionSizer(config, auditLogger)
		if err == nil {
			t.Errorf("Test %d: Expected error for invalid config, got nil", i)
		}
	}
}

// Helper function for float comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
