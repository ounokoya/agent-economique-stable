package engine

import (
	"testing"
	"time"
)

func TestNewPositionManager(t *testing.T) {
	config := TrailingStopConfig{
		TrendPercent:        2.0,
		CounterTrendPercent: 1.5,
	}
	
	grid := []AdjustmentLevel{
		{ProfitMin: 0, ProfitMax: 5, TrailingPercent: 2.0},
		{ProfitMin: 5, ProfitMax: 10, TrailingPercent: 1.5},
	}
	
	pm := NewPositionManager(config, grid)
	
	if pm == nil {
		t.Fatal("expected non-nil position manager")
	}
	
	if pm.IsOpen() {
		t.Error("new position manager should not have open position")
	}
	
	if pm.totalPositions != 0 {
		t.Errorf("expected 0 total positions, got %d", pm.totalPositions)
	}
}

func TestPositionManager_OpenPosition(t *testing.T) {
	pm := NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, []AdjustmentLevel{})
	
	tests := []struct {
		name        string
		direction   PositionDirection
		entryPrice  float64
		expectError bool
	}{
		{
			name:        "Valid long position",
			direction:   PositionLong,
			entryPrice:  100.0,
			expectError: false,
		},
		{
			name:        "Invalid zero price",
			direction:   PositionLong,
			entryPrice:  0.0,
			expectError: true,
		},
		{
			name:        "Invalid negative price",
			direction:   PositionShort,
			entryPrice:  -50.0,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset position manager
			pm = NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, []AdjustmentLevel{})
			
			err := pm.OpenPosition(tt.direction, tt.entryPrice, time.Now().UnixMilli(), "OVERSOLD")
			
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %t, got error: %v", tt.expectError, err)
			}
			
			if !tt.expectError {
				if !pm.IsOpen() {
					t.Error("position should be open after successful OpenPosition")
				}
				
				position := pm.GetPosition()
				if position.Direction != tt.direction {
					t.Errorf("expected direction %v, got %v", tt.direction, position.Direction)
				}
				
				if position.EntryPrice != tt.entryPrice {
					t.Errorf("expected entry price %f, got %f", tt.entryPrice, position.EntryPrice)
				}
			}
		})
	}
}

func TestPositionManager_OpenPosition_AlreadyOpen(t *testing.T) {
	pm := NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, []AdjustmentLevel{})
	
	// Open first position
	err := pm.OpenPosition(PositionLong, 100.0, time.Now().UnixMilli(), "OVERSOLD")
	if err != nil {
		t.Fatalf("failed to open first position: %v", err)
	}
	
	// Try to open second position
	err = pm.OpenPosition(PositionShort, 200.0, time.Now().UnixMilli(), "OVERBOUGHT")
	if err == nil {
		t.Error("expected error when opening position while one is already open")
	}
	
	if err != ErrPositionAlreadyOpen {
		t.Errorf("expected ErrPositionAlreadyOpen, got %v", err)
	}
}

func TestPositionManager_ClosePosition(t *testing.T) {
	pm := NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, []AdjustmentLevel{})
	
	// Try to close when no position is open
	err := pm.ClosePosition(time.Now().UnixMilli(), 100.0, "test")
	if err == nil {
		t.Error("expected error when closing non-existent position")
	}
	
	// Open a position
	err = pm.OpenPosition(PositionLong, 100.0, time.Now().UnixMilli(), "OVERSOLD")
	if err != nil {
		t.Fatalf("failed to open position: %v", err)
	}
	
	// Close the position
	err = pm.ClosePosition(time.Now().UnixMilli(), 110.0, "profit_target")
	if err != nil {
		t.Errorf("failed to close position: %v", err)
	}
	
	if pm.IsOpen() {
		t.Error("position should be closed after ClosePosition")
	}
	
	// Verify statistics updated
	stats := pm.GetStatistics()
	if stats.TotalPositions != 1 {
		t.Errorf("expected 1 total position, got %d", stats.TotalPositions)
	}
}

func TestPositionManager_UpdateTrailingStop(t *testing.T) {
	pm := NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, []AdjustmentLevel{})
	
	// Try to update when no position is open
	err := pm.UpdateTrailingStop(100.0)
	if err == nil {
		t.Error("expected error when updating trailing stop with no open position")
	}
	
	// Open a long position
	err = pm.OpenPosition(PositionLong, 100.0, time.Now().UnixMilli(), "OVERSOLD")
	if err != nil {
		t.Fatalf("failed to open position: %v", err)
	}
	
	initialPosition := pm.GetPosition()
	initialStop := initialPosition.StopLoss
	
	// Update with higher price (should tighten stop for long position)
	err = pm.UpdateTrailingStop(110.0)
	if err != nil {
		t.Errorf("failed to update trailing stop: %v", err)
	}
	
	updatedPosition := pm.GetPosition()
	newStop := updatedPosition.StopLoss
	
	// For long position, new stop should be higher (tighter)
	if newStop <= initialStop {
		t.Errorf("expected new stop %f to be higher than initial stop %f for long position", newStop, initialStop)
	}
}

func TestPositionManager_GetAdjustmentForProfit(t *testing.T) {
	grid := []AdjustmentLevel{
		{ProfitMin: 0, ProfitMax: 5, TrailingPercent: 2.0},
		{ProfitMin: 5, ProfitMax: 10, TrailingPercent: 1.5},
		{ProfitMin: 10, ProfitMax: 20, TrailingPercent: 1.0},
	}
	
	pm := NewPositionManager(TrailingStopConfig{TrendPercent: 2.0}, grid)
	
	tests := []struct {
		name          string
		profitPercent float64
		expected      float64
	}{
		{"Low profit", 3.0, 2.0},
		{"Medium profit", 7.0, 1.5},
		{"High profit", 15.0, 1.0},
		{"Very high profit", 25.0, 1.0}, // Should use last level
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.GetAdjustmentForProfit(tt.profitPercent)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestValidateStopLevel(t *testing.T) {
	tests := []struct {
		name        string
		direction   PositionDirection
		entryPrice  float64
		stopPrice   float64
		expectError bool
	}{
		{
			name:        "Valid long stop",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   98.0,
			expectError: false,
		},
		{
			name:        "Valid short stop",
			direction:   PositionShort,
			entryPrice:  100.0,
			stopPrice:   102.0,
			expectError: false,
		},
		{
			name:        "Invalid long stop (above entry)",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   102.0,
			expectError: true,
		},
		{
			name:        "Invalid short stop (below entry)",
			direction:   PositionShort,
			entryPrice:  100.0,
			stopPrice:   98.0,
			expectError: true,
		},
		{
			name:        "Stop too wide",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   40.0, // 60% stop
			expectError: true,
		},
		{
			name:        "Stop too tight",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   99.99, // 0.01% stop
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStopLevel(tt.direction, tt.entryPrice, tt.stopPrice)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %t, got error: %v", tt.expectError, err)
			}
		})
	}
}

func TestCalculateRisk(t *testing.T) {
	tests := []struct {
		name       string
		direction  PositionDirection
		entryPrice float64
		stopPrice  float64
		expected   float64
	}{
		{
			name:       "Long position 2% risk",
			direction:  PositionLong,
			entryPrice: 100.0,
			stopPrice:  98.0,
			expected:   2.0,
		},
		{
			name:       "Short position 3% risk",
			direction:  PositionShort,
			entryPrice: 100.0,
			stopPrice:  103.0,
			expected:   3.0,
		},
		{
			name:       "Zero prices",
			direction:  PositionLong,
			entryPrice: 0.0,
			stopPrice:  0.0,
			expected:   0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRisk(tt.direction, tt.entryPrice, tt.stopPrice)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateRewardRatio(t *testing.T) {
	tests := []struct {
		name        string
		direction   PositionDirection
		entryPrice  float64
		stopPrice   float64
		targetPrice float64
		expected    float64
	}{
		{
			name:        "Long 1:2 ratio",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   98.0, // 2% risk
			targetPrice: 104.0, // 4% reward
			expected:    2.0,   // 4/2 = 2
		},
		{
			name:        "Short 1:1.5 ratio",
			direction:   PositionShort,
			entryPrice:  100.0,
			stopPrice:   102.0, // 2% risk
			targetPrice: 97.0,  // 3% reward
			expected:    1.5,   // 3/2 = 1.5
		},
		{
			name:        "No risk",
			direction:   PositionLong,
			entryPrice:  100.0,
			stopPrice:   100.0,
			targetPrice: 105.0,
			expected:    0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRewardRatio(tt.direction, tt.entryPrice, tt.stopPrice, tt.targetPrice)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}
