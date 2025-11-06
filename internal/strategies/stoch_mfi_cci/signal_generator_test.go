package stoch_mfi_cci

import (
	"testing"
	
	"agent-economique/internal/indicators"
)

func TestGenerateSignalsLONG(t *testing.T) {
	config := DefaultStrategyConfig()
	
	// Create test indicator results for LONG signal
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:             15.0, // Oversold
			D:             18.0,
			Zone:          indicators.StochOversold,
			CrossoverType: indicators.CrossUp, // K crosses above D
			IsExtreme:     true,
		},
		MFI: &indicators.MFIValues{
			Value:     15.0, // Oversold
			Zone:      indicators.MFIOversold,
			IsExtreme: true,
		},
		CCI: &indicators.CCIValues{
			Value: -120.0, // Oversold
			Zone:  indicators.CCIOversold,
		},
		Timestamp: 1699000000000,
	}
	
	signals := generateSignals(results, config)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	signal := signals[0]
	if signal.Direction != indicators.LongSignal {
		t.Errorf("Expected LONG signal, got %s", signal.Direction)
	}
	
	if signal.Confidence < config.PremiumConfidence {
		t.Errorf("Expected premium confidence (%.2f), got %.2f", config.PremiumConfidence, signal.Confidence)
	}
}

func TestGenerateSignalsSHORT(t *testing.T) {
	config := DefaultStrategyConfig()
	
	// Create test indicator results for SHORT signal
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:             85.0, // Overbought
			D:             82.0,
			Zone:          indicators.StochOverbought,
			CrossoverType: indicators.CrossDown, // K crosses below D
			IsExtreme:     true,
		},
		MFI: &indicators.MFIValues{
			Value:     85.0, // Overbought
			Zone:      indicators.MFIOverbought,
			IsExtreme: true,
		},
		CCI: &indicators.CCIValues{
			Value: 120.0, // Overbought
			Zone:  indicators.CCIOverbought,
		},
		Timestamp: 1699000000000,
	}
	
	signals := generateSignals(results, config)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	signal := signals[0]
	if signal.Direction != indicators.ShortSignal {
		t.Errorf("Expected SHORT signal, got %s", signal.Direction)
	}
	
	if signal.Confidence < config.PremiumConfidence {
		t.Errorf("Expected premium confidence (%.2f), got %.2f", config.PremiumConfidence, signal.Confidence)
	}
}

func TestGenerateSignalsMinimal(t *testing.T) {
	config := DefaultStrategyConfig()
	
	// Create test for minimal signal (STOCH + MFI only)
	results := &indicators.IndicatorResults{
		Stochastic: &indicators.StochasticValues{
			K:             18.0, // Oversold
			D:             20.0,
			Zone:          indicators.StochOversold,
			CrossoverType: indicators.CrossUp,
			IsExtreme:     true,
		},
		MFI: &indicators.MFIValues{
			Value:     18.0, // Oversold
			Zone:      indicators.MFIOversold,
			IsExtreme: true,
		},
		CCI: &indicators.CCIValues{
			Value: -50.0, // Neutral (not extreme)
			Zone:  indicators.CCINormal,
		},
		Timestamp: 1699000000000,
	}
	
	signals := generateSignals(results, config)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	signal := signals[0]
	if signal.Direction != indicators.LongSignal {
		t.Errorf("Expected LONG signal, got %s", signal.Direction)
	}
	
	if signal.Confidence >= config.PremiumConfidence {
		t.Errorf("Expected minimal confidence (< %.2f), got %.2f", config.PremiumConfidence, signal.Confidence)
	}
	
	if signal.Confidence < config.MinConfidence {
		t.Errorf("Expected confidence >= %.2f, got %.2f", config.MinConfidence, signal.Confidence)
	}
}

func TestGenerateSignalsNoSignal(t *testing.T) {
	config := DefaultStrategyConfig()
	
	testCases := []struct {
		name    string
		results *indicators.IndicatorResults
	}{
		{
			name: "No STOCH crossover",
			results: &indicators.IndicatorResults{
				Stochastic: &indicators.StochasticValues{
					K:             15.0,
					D:             18.0,
					Zone:          indicators.StochOversold,
					CrossoverType: indicators.NoCrossover, // No crossover
					IsExtreme:     true,
				},
				MFI: &indicators.MFIValues{
					Value:     15.0,
					Zone:      indicators.MFIOversold,
					IsExtreme: true,
				},
				CCI: &indicators.CCIValues{
					Value: -120.0,
					Zone:  indicators.CCIOversold,
				},
			},
		},
		{
			name: "STOCH not extreme",
			results: &indicators.IndicatorResults{
				Stochastic: &indicators.StochasticValues{
					K:             50.0, // Neutral
					D:             48.0,
					Zone:          indicators.StochNeutral,
					CrossoverType: indicators.CrossUp,
					IsExtreme:     false,
				},
				MFI: &indicators.MFIValues{
					Value:     15.0,
					Zone:      indicators.MFIOversold,
					IsExtreme: true,
				},
				CCI: &indicators.CCIValues{
					Value: -120.0,
					Zone:  indicators.CCIOversold,
				},
			},
		},
		{
			name: "No supporting indicators",
			results: &indicators.IndicatorResults{
				Stochastic: &indicators.StochasticValues{
					K:             15.0,
					D:             18.0,
					Zone:          indicators.StochOversold,
					CrossoverType: indicators.CrossUp,
					IsExtreme:     true,
				},
				MFI: &indicators.MFIValues{
					Value:     50.0, // Neutral
					Zone:      indicators.MFINeutral,
					IsExtreme: false,
				},
				CCI: &indicators.CCIValues{
					Value: -50.0, // Neutral
					Zone:  indicators.CCINormal,
				},
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.results.Timestamp = 1699000000000
			signals := generateSignals(tc.results, config)
			
			if len(signals) != 0 {
				t.Errorf("Expected 0 signals, got %d", len(signals))
			}
		})
	}
}

func TestValidateBarDirection(t *testing.T) {
	testCases := []struct {
		name      string
		klines    []indicators.Kline
		direction indicators.SignalDirection
		expected  bool
	}{
		{
			name: "Bullish bar for LONG signal",
			klines: []indicators.Kline{
				{Open: 100.0, Close: 105.0}, // Bullish
			},
			direction: indicators.LongSignal,
			expected:  true,
		},
		{
			name: "Bearish bar for LONG signal",
			klines: []indicators.Kline{
				{Open: 100.0, Close: 95.0}, // Bearish
			},
			direction: indicators.LongSignal,
			expected:  false,
		},
		{
			name: "Bearish bar for SHORT signal",
			klines: []indicators.Kline{
				{Open: 100.0, Close: 95.0}, // Bearish
			},
			direction: indicators.ShortSignal,
			expected:  true,
		},
		{
			name: "Bullish bar for SHORT signal",
			klines: []indicators.Kline{
				{Open: 100.0, Close: 105.0}, // Bullish
			},
			direction: indicators.ShortSignal,
			expected:  false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateBarDirection(tc.klines, tc.direction)
			if result != tc.expected {
				t.Errorf("Expected %t, got %t", tc.expected, result)
			}
		})
	}
}

func TestGetSignalStrength(t *testing.T) {
	// Premium signal components
	stochPremium := &indicators.StochasticValues{
		Zone: indicators.StochOversold,
	}
	mfiPremium := &indicators.MFIValues{
		Zone: indicators.MFIOversold,
	}
	cciPremium := &indicators.CCIValues{
		Zone: indicators.CCIOversold,
	}
	
	// Test premium signal
	strength := GetSignalStrength(stochPremium, mfiPremium, cciPremium, indicators.LongSignal)
	if strength != SignalPremium {
		t.Errorf("Expected premium signal, got %s", strength)
	}
	
	// Test minimal signal (missing CCI)
	cciNeutral := &indicators.CCIValues{
		Zone: indicators.CCINormal,
	}
	strength = GetSignalStrength(stochPremium, mfiPremium, cciNeutral, indicators.LongSignal)
	if strength != SignalMinimal {
		t.Errorf("Expected minimal signal, got %s", strength)
	}
}

func TestIsValidSTOCHCrossover(t *testing.T) {
	testCases := []struct {
		name      string
		stoch     *indicators.StochasticValues
		direction indicators.SignalDirection
		expected  bool
	}{
		{
			name: "Valid LONG crossover",
			stoch: &indicators.StochasticValues{
				Zone:          indicators.StochOversold,
				CrossoverType: indicators.CrossUp,
			},
			direction: indicators.LongSignal,
			expected:  true,
		},
		{
			name: "Invalid LONG crossover (wrong zone)",
			stoch: &indicators.StochasticValues{
				Zone:          indicators.StochOverbought,
				CrossoverType: indicators.CrossUp,
			},
			direction: indicators.LongSignal,
			expected:  false,
		},
		{
			name: "Invalid LONG crossover (wrong crossover)",
			stoch: &indicators.StochasticValues{
				Zone:          indicators.StochOversold,
				CrossoverType: indicators.CrossDown,
			},
			direction: indicators.LongSignal,
			expected:  false,
		},
		{
			name: "Valid SHORT crossover",
			stoch: &indicators.StochasticValues{
				Zone:          indicators.StochOverbought,
				CrossoverType: indicators.CrossDown,
			},
			direction: indicators.ShortSignal,
			expected:  true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidSTOCHCrossover(tc.stoch, tc.direction)
			if result != tc.expected {
				t.Errorf("Expected %t, got %t", tc.expected, result)
			}
		})
	}
}

func TestApplyConfidenceBonuses(t *testing.T) {
	baseConfidence := 0.7
	
	// Create indicators with extreme values for bonuses
	stoch := &indicators.StochasticValues{
		K: 5.0, // Very oversold (bonus)
		D: 8.0,
	}
	mfi := &indicators.MFIValues{
		Value: 8.0, // Very oversold (bonus)
	}
	cci := &indicators.CCIValues{
		Value: -180.0, // Very oversold (bonus)
	}
	
	confidence := applyConfidenceBonuses(baseConfidence, stoch, mfi, cci, indicators.LongSignal)
	
	// Should have bonuses applied
	if confidence <= baseConfidence {
		t.Errorf("Expected confidence > %.2f with bonuses, got %.2f", baseConfidence, confidence)
	}
	
	// Test with normal values (no bonuses)
	stochNormal := &indicators.StochasticValues{K: 15.0, D: 18.0}
	mfiNormal := &indicators.MFIValues{Value: 18.0}
	cciNormal := &indicators.CCIValues{Value: -120.0}
	
	confidenceNormal := applyConfidenceBonuses(baseConfidence, stochNormal, mfiNormal, cciNormal, indicators.LongSignal)
	
	if confidenceNormal < baseConfidence {
		t.Errorf("Expected confidence >= %.2f, got %.2f", baseConfidence, confidenceNormal)
	}
}
