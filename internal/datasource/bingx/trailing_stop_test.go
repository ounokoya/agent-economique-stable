package bingx

import (
	"context"
	"testing"
	"time"
)

func TestNewTrailingStopManager(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	if tsm == nil {
		t.Fatal("NewTrailingStopManager should not return nil")
	}
	
	if tsm.client != client {
		t.Error("TrailingStopManager should reference the provided client")
	}
	
	if tsm.trading != trading {
		t.Error("TrailingStopManager should reference the provided trading service")
	}
	
	if tsm.positions == nil {
		t.Error("Positions map should be initialized")
	}
	
	if tsm.conditions == nil {
		t.Error("Conditions slice should be initialized")
	}
	
	if tsm.monitoring {
		t.Error("Monitoring should be false initially")
	}
}

func TestTrailingStopManagerAddCondition(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	condition := TrailingStopCondition{
		Indicator:     "CCI",
		Trigger:       "inverse_zone",
		Action:        "tighten",
		AdjustmentPct: 0.1,
		MinProfit:     0.02,
	}
	
	tsm.AddCondition(condition)
	
	if len(tsm.conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(tsm.conditions))
	}
	
	if tsm.conditions[0].Indicator != "CCI" {
		t.Errorf("Expected indicator CCI, got %s", tsm.conditions[0].Indicator)
	}
	
	if tsm.conditions[0].Trigger != "inverse_zone" {
		t.Errorf("Expected trigger inverse_zone, got %s", tsm.conditions[0].Trigger)
	}
	
	if tsm.conditions[0].Action != "tighten" {
		t.Errorf("Expected action tighten, got %s", tsm.conditions[0].Action)
	}
	
	if tsm.conditions[0].AdjustmentPct != 0.1 {
		t.Errorf("Expected adjustment 0.1, got %f", tsm.conditions[0].AdjustmentPct)
	}
}

func TestTrailingStopManagerAddMultipleConditions(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	conditions := []TrailingStopCondition{
		{
			Indicator:     "CCI",
			Trigger:       "inverse_zone",
			Action:        "tighten",
			AdjustmentPct: 0.1,
			MinProfit:     0.02,
		},
		{
			Indicator:     "MACD",
			Trigger:       "signal_cross",
			Action:        "close",
			AdjustmentPct: 0.0,
			MinProfit:     0.01,
		},
		{
			Indicator:     "DMI",
			Trigger:       "counter_trend",
			Action:        "loosen",
			AdjustmentPct: 0.05,
			MinProfit:     0.03,
		},
	}
	
	for _, condition := range conditions {
		tsm.AddCondition(condition)
	}
	
	if len(tsm.conditions) != 3 {
		t.Errorf("Expected 3 conditions, got %d", len(tsm.conditions))
	}
}

func TestTrailingStopManagerGetClosingSide(t *testing.T) {
	tsm := &TrailingStopManager{}
	
	tests := []struct {
		positionSide PositionSide
		expected     string
	}{
		{PositionSideLong, string(OrderSideSell)},
		{PositionSideShort, string(OrderSideBuy)},
	}
	
	for _, tt := range tests {
		result := tsm.getClosingSide(tt.positionSide)
		if result != tt.expected {
			t.Errorf("Expected closing side %s for position %s, got %s", tt.expected, tt.positionSide, result)
		}
	}
}

func TestTrailingStopManagerCalculatePositionProfit(t *testing.T) {
	tsm := &TrailingStopManager{}
	
	tests := []struct {
		name         string
		position     *Position
		expected     float64
	}{
		{
			name: "Long position with profit",
			position: &Position{
				PositionSide: PositionSideLong,
				EntryPrice:   45000.0,
				MarkPrice:    46000.0,
			},
			expected: (46000.0 - 45000.0) / 45000.0, // ~0.0222
		},
		{
			name: "Long position with loss",
			position: &Position{
				PositionSide: PositionSideLong,
				EntryPrice:   45000.0,
				MarkPrice:    44000.0,
			},
			expected: (44000.0 - 45000.0) / 45000.0, // ~-0.0222
		},
		{
			name: "Short position with profit",
			position: &Position{
				PositionSide: PositionSideShort,
				EntryPrice:   45000.0,
				MarkPrice:    44000.0,
			},
			expected: (45000.0 - 44000.0) / 45000.0, // ~0.0222
		},
		{
			name: "Short position with loss",
			position: &Position{
				PositionSide: PositionSideShort,
				EntryPrice:   45000.0,
				MarkPrice:    46000.0,
			},
			expected: (45000.0 - 46000.0) / 45000.0, // ~-0.0222
		},
		{
			name: "Zero entry price",
			position: &Position{
				PositionSide: PositionSideLong,
				EntryPrice:   0.0,
				MarkPrice:    46000.0,
			},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tsm.calculatePositionProfit(tt.position)
			
			// Use a small epsilon for float comparison
			epsilon := 0.0001
			if result-tt.expected > epsilon || tt.expected-result > epsilon {
				t.Errorf("Expected profit %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestTrailingStopManagerEvaluateCondition(t *testing.T) {
	tsm := &TrailingStopManager{}
	
	tests := []struct {
		name         string
		condition    TrailingStopCondition
		signal       string
		value        float64
		managedPos   *ManagedPosition
		expected     bool
	}{
		{
			name: "CCI inverse zone - Long position oversold",
			condition: TrailingStopCondition{
				Indicator: "CCI",
				Trigger:   "inverse_zone",
			},
			signal: "oversold",
			value:  -120.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideLong},
			},
			expected: true,
		},
		{
			name: "CCI inverse zone - Short position overbought",
			condition: TrailingStopCondition{
				Indicator: "CCI",
				Trigger:   "inverse_zone",
			},
			signal: "overbought",
			value:  120.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideShort},
			},
			expected: true,
		},
		{
			name: "CCI inverse zone - Long position not oversold",
			condition: TrailingStopCondition{
				Indicator: "CCI",
				Trigger:   "inverse_zone",
			},
			signal: "normal",
			value:  -80.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideLong},
			},
			expected: false,
		},
		{
			name: "MACD signal cross - Long position bearish cross",
			condition: TrailingStopCondition{
				Indicator: "MACD",
				Trigger:   "signal_cross",
			},
			signal: "bearish_cross",
			value:  0.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideLong},
			},
			expected: true,
		},
		{
			name: "MACD signal cross - Short position bullish cross",
			condition: TrailingStopCondition{
				Indicator: "MACD",
				Trigger:   "signal_cross",
			},
			signal: "bullish_cross",
			value:  0.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideShort},
			},
			expected: true,
		},
		{
			name: "DMI counter trend",
			condition: TrailingStopCondition{
				Indicator: "DMI",
				Trigger:   "counter_trend",
			},
			signal: "counter_trend",
			value:  0.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideLong},
			},
			expected: true,
		},
		{
			name: "Unknown trigger",
			condition: TrailingStopCondition{
				Indicator: "UNKNOWN",
				Trigger:   "unknown_trigger",
			},
			signal: "unknown",
			value:  0.0,
			managedPos: &ManagedPosition{
				Position: &Position{PositionSide: PositionSideLong},
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tsm.evaluateCondition(tt.condition, tt.signal, tt.value, tt.managedPos)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestTrailingStopManagerGetManagedPositions(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	// Add some test positions
	tsm.positions["BTC-USDT_LONG"] = &ManagedPosition{
		Position: &Position{
			Symbol:       "BTC-USDT",
			PositionSide: PositionSideLong,
		},
		IsActive: true,
	}
	
	tsm.positions["ETH-USDT_SHORT"] = &ManagedPosition{
		Position: &Position{
			Symbol:       "ETH-USDT",
			PositionSide: PositionSideShort,
		},
		IsActive: true,
	}
	
	tsm.positions["SOL-USDT_LONG"] = &ManagedPosition{
		Position: &Position{
			Symbol:       "SOL-USDT",
			PositionSide: PositionSideLong,
		},
		IsActive: false, // This should not be included
	}
	
	result := tsm.GetManagedPositions()
	
	if len(result) != 2 {
		t.Errorf("Expected 2 active positions, got %d", len(result))
	}
	
	if _, exists := result["BTC-USDT_LONG"]; !exists {
		t.Error("BTC-USDT_LONG should be in managed positions")
	}
	
	if _, exists := result["ETH-USDT_SHORT"]; !exists {
		t.Error("ETH-USDT_SHORT should be in managed positions")
	}
	
	if _, exists := result["SOL-USDT_LONG"]; exists {
		t.Error("SOL-USDT_LONG should not be in managed positions (inactive)")
	}
}

func TestTrailingStopManagerRemovePosition(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	// Add test position
	tsm.positions["BTC-USDT_LONG"] = &ManagedPosition{
		Position: &Position{
			Symbol:       "BTC-USDT",
			PositionSide: PositionSideLong,
		},
		IsActive: true,
	}
	
	// Verify position exists
	if len(tsm.positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(tsm.positions))
	}
	
	// Remove position
	tsm.RemovePosition("BTC-USDT", PositionSideLong)
	
	// Verify position removed
	if len(tsm.positions) != 0 {
		t.Errorf("Expected 0 positions after removal, got %d", len(tsm.positions))
	}
}

func TestTrailingStopManagerStartStopMonitoring(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	
	ctx := context.Background()
	
	// Test start monitoring
	if tsm.monitoring {
		t.Error("Monitoring should be false initially")
	}
	
	tsm.StartMonitoring(ctx, 100*time.Millisecond)
	
	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)
	
	if !tsm.monitoring {
		t.Error("Monitoring should be true after starting")
	}
	
	// Test start monitoring when already monitoring (should not crash)
	tsm.StartMonitoring(ctx, 100*time.Millisecond)
	
	// Test stop monitoring
	tsm.StopMonitoring()
	
	// Give it a moment to stop
	time.Sleep(150 * time.Millisecond)
	
	if tsm.monitoring {
		t.Error("Monitoring should be false after stopping")
	}
	
	// Test stop monitoring when not monitoring (should not crash)
	tsm.StopMonitoring()
}

func TestTrailingStopManagerValidationErrors(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	trading := NewTradingService(client)
	tsm := NewTrailingStopManager(client, trading)
	ctx := context.Background()
	
	// Test AddTrailingStop with invalid trailing rate
	err = tsm.AddTrailingStop(ctx, "BTC-USDT", PositionSideLong, 0, 45000)
	if err == nil {
		t.Error("AddTrailingStop should return error for zero trailing rate")
	}
	
	err = tsm.AddTrailingStop(ctx, "BTC-USDT", PositionSideLong, -0.1, 45000)
	if err == nil {
		t.Error("AddTrailingStop should return error for negative trailing rate")
	}
	
	err = tsm.AddTrailingStop(ctx, "BTC-USDT", PositionSideLong, 1.5, 45000)
	if err == nil {
		t.Error("AddTrailingStop should return error for trailing rate > 1")
	}
	
	// Test UpdateTrailingStop with invalid rate
	err = tsm.UpdateTrailingStop(ctx, "BTC-USDT", PositionSideLong, 0, "test")
	if err == nil {
		t.Error("UpdateTrailingStop should return error for zero rate")
	}
	
	err = tsm.UpdateTrailingStop(ctx, "BTC-USDT", PositionSideLong, -0.1, "test")
	if err == nil {
		t.Error("UpdateTrailingStop should return error for negative rate")
	}
	
	err = tsm.UpdateTrailingStop(ctx, "BTC-USDT", PositionSideLong, 1.5, "test")
	if err == nil {
		t.Error("UpdateTrailingStop should return error for rate > 1")
	}
	
	// Test UpdateTrailingStop with non-existent position
	err = tsm.UpdateTrailingStop(ctx, "NONEXISTENT", PositionSideLong, 0.1, "test")
	if err == nil {
		t.Error("UpdateTrailingStop should return error for non-existent position")
	}
}

func TestManagedPositionStruct(t *testing.T) {
	// Test ManagedPosition struct creation and fields
	position := &Position{
		Symbol:       "BTC-USDT",
		PositionSide: PositionSideLong,
		Size:         100.0,
		EntryPrice:   45000.0,
		MarkPrice:    46000.0,
	}
	
	conditions := []TrailingStopCondition{
		{
			Indicator:     "CCI",
			Trigger:       "inverse_zone",
			Action:        "tighten",
			AdjustmentPct: 0.1,
			MinProfit:     0.02,
		},
	}
	
	managedPos := &ManagedPosition{
		Position:         position,
		TrailingStopRate: 0.005,
		ActivationPrice:  45500.0,
		CurrentStopPrice: 45000.0,
		LastUpdateTime:   time.Now(),
		Conditions:       conditions,
		IsActive:         true,
		ProfitThreshold:  0.01,
	}
	
	if managedPos.Position.Symbol != "BTC-USDT" {
		t.Error("Position symbol should be preserved")
	}
	
	if managedPos.TrailingStopRate != 0.005 {
		t.Error("Trailing stop rate should be preserved")
	}
	
	if !managedPos.IsActive {
		t.Error("Position should be active")
	}
	
	if len(managedPos.Conditions) != 1 {
		t.Error("Conditions should be preserved")
	}
}

func TestTrailingStopConditionStruct(t *testing.T) {
	// Test TrailingStopCondition struct creation and fields
	condition := TrailingStopCondition{
		Indicator:     "MACD",
		Trigger:       "signal_cross",
		Action:        "close",
		AdjustmentPct: 0.0,
		MinProfit:     0.015,
	}
	
	if condition.Indicator != "MACD" {
		t.Error("Indicator should be preserved")
	}
	
	if condition.Trigger != "signal_cross" {
		t.Error("Trigger should be preserved")
	}
	
	if condition.Action != "close" {
		t.Error("Action should be preserved")
	}
	
	if condition.AdjustmentPct != 0.0 {
		t.Error("AdjustmentPct should be preserved")
	}
	
	if condition.MinProfit != 0.015 {
		t.Error("MinProfit should be preserved")
	}
}
