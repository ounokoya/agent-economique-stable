package engine

import (
	"testing"
	"time"
)

func TestPositionDirection_String(t *testing.T) {
	tests := []struct {
		name     string
		direction PositionDirection
		expected string
	}{
		{"Long position", PositionLong, "LONG"},
		{"Short position", PositionShort, "SHORT"},
		{"No position", PositionNone, "NONE"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.direction.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		name     string
		mode     ExecutionMode
		expected string
	}{
		{"Backtest mode", BacktestMode, "backtest"},
		{"Paper mode", PaperMode, "paper"},
		{"Live mode", LiveMode, "live"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mode.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestZoneType_String(t *testing.T) {
	tests := []struct {
		name     string
		zoneType ZoneType
		expected string
	}{
		{"CCI Inverse", ZoneCCIInverse, "CCI_INVERSE"},
		{"MACD Inverse", ZoneMACDInverse, "MACD_INVERSE"},
		{"DI Counter", ZoneDICounter, "DI_COUNTER"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.zoneType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPosition_CalculateProfitPercent(t *testing.T) {
	tests := []struct {
		name         string
		position     Position
		currentPrice float64
		expected     float64
	}{
		{
			name: "Long position profit",
			position: Position{
				IsOpen:     true,
				Direction:  PositionLong,
				EntryPrice: 100.0,
			},
			currentPrice: 110.0,
			expected:     10.0,
		},
		{
			name: "Long position loss",
			position: Position{
				IsOpen:     true,
				Direction:  PositionLong,
				EntryPrice: 100.0,
			},
			currentPrice: 95.0,
			expected:     -5.0,
		},
		{
			name: "Short position profit",
			position: Position{
				IsOpen:     true,
				Direction:  PositionShort,
				EntryPrice: 100.0,
			},
			currentPrice: 90.0,
			expected:     10.0,
		},
		{
			name: "Closed position",
			position: Position{
				IsOpen:     false,
				Direction:  PositionLong,
				EntryPrice: 100.0,
			},
			currentPrice: 110.0,
			expected:     0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.position.CalculateProfitPercent(tt.currentPrice)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestPosition_IsStopHit(t *testing.T) {
	tests := []struct {
		name         string
		position     Position
		currentPrice float64
		expected     bool
	}{
		{
			name: "Long stop hit",
			position: Position{
				IsOpen:    true,
				Direction: PositionLong,
				StopLoss:  95.0,
			},
			currentPrice: 94.0,
			expected:     true,
		},
		{
			name: "Long stop not hit",
			position: Position{
				IsOpen:    true,
				Direction: PositionLong,
				StopLoss:  95.0,
			},
			currentPrice: 96.0,
			expected:     false,
		},
		{
			name: "Short stop hit",
			position: Position{
				IsOpen:    true,
				Direction: PositionShort,
				StopLoss:  105.0,
			},
			currentPrice: 106.0,
			expected:     true,
		},
		{
			name: "Closed position",
			position: Position{
				IsOpen:    false,
				Direction: PositionLong,
				StopLoss:  95.0,
			},
			currentPrice: 90.0,
			expected:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.position.IsStopHit(tt.currentPrice)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    EngineConfig
		expectErr bool
	}{
		{
			name:      "Valid config",
			config:    DefaultEngineConfig(),
			expectErr: false,
		},
		{
			name: "Invalid window size",
			config: EngineConfig{
				WindowSize: 0,
				TrailingStop: TrailingStopConfig{
					TrendPercent:        2.0,
					CounterTrendPercent: 1.5,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid trend percent",
			config: EngineConfig{
				WindowSize: 300,
				TrailingStop: TrailingStopConfig{
					TrendPercent:        0,
					CounterTrendPercent: 1.5,
				},
			},
			expectErr: true,
		},
		{
			name: "Overlapping adjustment grid",
			config: EngineConfig{
				WindowSize: 300,
				TrailingStop: TrailingStopConfig{
					TrendPercent:        2.0,
					CounterTrendPercent: 1.5,
				},
				AdjustmentGrid: []AdjustmentLevel{
					{ProfitMin: 0, ProfitMax: 10, TrailingPercent: 2.0},
					{ProfitMin: 5, ProfitMax: 15, TrailingPercent: 1.5},
				},
			},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfiguration(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %t, got error: %v", tt.expectErr, err)
			}
		})
	}
}

func TestIsMarkerTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		expected  bool
	}{
		{
			name:      "5m marker - 00:00:00",
			timestamp: 1640995200000, // 2022-01-01 00:00:00 UTC
			expected:  true,
		},
		{
			name:      "5m marker - 00:05:00", 
			timestamp: 1640995500000, // 2022-01-01 00:05:00 UTC
			expected:  true,
		},
		{
			name:      "5m marker - 00:10:00",
			timestamp: 1640995800000, // 2022-01-01 00:10:00 UTC  
			expected:  true,
		},
		{
			name:      "non-marker - 00:03:00 (not 5m boundary)",
			timestamp: 1640995380000, // 2022-01-01 00:03:00 UTC
			expected:  false,
		},
		{
			name:      "non-marker - 00:05:30 (has seconds)",
			timestamp: 1640995530000, // 2022-01-01 00:05:30 UTC
			expected:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMarkerTimestamp(tt.timestamp)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestGetTimeframeAlignment(t *testing.T) {
	// 12:00:00 UTC should align with all timeframes (12h is divisible by 4)
	timestamp := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	timeframes := []string{"5m", "15m", "1h", "4h"}
	
	result := GetTimeframeAlignment(timestamp, timeframes)
	
	expectedCount := 4 // Should align with all timeframes
	if len(result) != expectedCount {
		t.Errorf("expected %d alignments, got %d", expectedCount, len(result))
	}
	
	// Test 5-minute only alignment
	timestamp5m := time.Date(2023, 6, 1, 10, 5, 0, 0, time.UTC).UnixMilli()
	result5m := GetTimeframeAlignment(timestamp5m, timeframes)
	
	expected5m := []string{"5m"}
	if len(result5m) != len(expected5m) {
		t.Errorf("expected %v, got %v", expected5m, result5m)
	}
}

func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()
	
	// Validate the default config is valid
	err := ValidateConfiguration(config)
	if err != nil {
		t.Errorf("default config should be valid, got error: %v", err)
	}
	
	// Check some expected values
	if config.WindowSize != 300 {
		t.Errorf("expected WindowSize 300, got %d", config.WindowSize)
	}
	
	if !config.AntiLookAhead {
		t.Error("expected AntiLookAhead to be true")
	}
	
	if len(config.AdjustmentGrid) == 0 {
		t.Error("expected non-empty AdjustmentGrid")
	}
}
