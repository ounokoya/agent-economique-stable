package indicators

import (
	"testing"
	"time"
)

func TestGenerateStrategySignals_LongTrendSignal(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{
			CrossoverType: CrossUp,
		},
		CCI: &CCIValues{
			Zone: CCIOversold,
		},
		DMI: &DMIValues{
			PlusDI:  25,
			MinusDI: 15, // DI+ > DI- (trend)
		},
		Timestamp: time.Now().UnixMilli(),
	}
	
	signals := generateStrategySignals(results)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	if signals[0].Direction != LongSignal {
		t.Errorf("Expected LongSignal, got %v", signals[0].Direction)
	}
	
	if signals[0].Type != TrendSignal {
		t.Errorf("Expected TrendSignal, got %v", signals[0].Type)
	}
}

func TestGenerateStrategySignals_LongCounterTrendSignal(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{
			CrossoverType: CrossUp,
		},
		CCI: &CCIValues{
			Zone: CCIOversold,
		},
		DMI: &DMIValues{
			PlusDI:  15,
			MinusDI: 25, // DI- > DI+ (counter-trend)
		},
		Timestamp: time.Now().UnixMilli(),
	}
	
	signals := generateStrategySignals(results)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	if signals[0].Direction != LongSignal {
		t.Errorf("Expected LongSignal, got %v", signals[0].Direction)
	}
	
	if signals[0].Type != CounterTrendSignal {
		t.Errorf("Expected CounterTrendSignal, got %v", signals[0].Type)
	}
}

func TestGenerateStrategySignals_ShortTrendSignal(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{
			CrossoverType: CrossDown,
		},
		CCI: &CCIValues{
			Zone: CCIOverbought,
		},
		DMI: &DMIValues{
			PlusDI:  15,
			MinusDI: 25, // DI+ < DI- (trend for SHORT)
		},
		Timestamp: time.Now().UnixMilli(),
	}
	
	signals := generateStrategySignals(results)
	
	if len(signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(signals))
	}
	
	if signals[0].Direction != ShortSignal {
		t.Errorf("Expected ShortSignal, got %v", signals[0].Direction)
	}
	
	if signals[0].Type != TrendSignal {
		t.Errorf("Expected TrendSignal, got %v", signals[0].Type)
	}
}

func TestGenerateStrategySignals_NoSignal_WrongCCIZone(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{
			CrossoverType: CrossUp,
		},
		CCI: &CCIValues{
			Zone: CCINormal, // Wrong zone for signal
		},
		DMI: &DMIValues{
			PlusDI:  25,
			MinusDI: 15,
		},
	}
	
	signals := generateStrategySignals(results)
	
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals, got %d", len(signals))
	}
}

func TestGenerateStrategySignals_NoSignal_NoCrossover(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{
			CrossoverType: NoCrossover, // No crossover
		},
		CCI: &CCIValues{
			Zone: CCIOversold,
		},
		DMI: &DMIValues{
			PlusDI:  25,
			MinusDI: 15,
		},
	}
	
	signals := generateStrategySignals(results)
	
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals, got %d", len(signals))
	}
}

func TestGenerateStrategySignals_NilInputs(t *testing.T) {
	// Test with nil results
	signals := generateStrategySignals(nil)
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals for nil input, got %d", len(signals))
	}
	
	// Test with nil MACD
	results := &IndicatorResults{
		MACD: nil,
		CCI: &CCIValues{Zone: CCIOversold},
		DMI: &DMIValues{PlusDI: 25, MinusDI: 15},
	}
	signals = generateStrategySignals(results)
	if len(signals) != 0 {
		t.Errorf("Expected 0 signals for nil MACD, got %d", len(signals))
	}
}

func TestCalculateConfidence_HighConfidence(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: -180}, // Very oversold
		MACD: &MACDValues{Histogram: 0.8}, // Strong momentum
		DMI: &DMIValues{ADX: 35, PlusDI: 30, MinusDI: 10}, // Strong trend + DI spread
	}
	
	confidence := calculateConfidence(results, TrendSignal, LongSignal)
	
	if confidence < 0.9 {
		t.Errorf("Expected high confidence (>0.9), got %.2f", confidence)
	}
	
	if confidence > 1.0 {
		t.Errorf("Confidence should be capped at 1.0, got %.2f", confidence)
	}
}

func TestCalculateConfidence_BaseConfidence(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: -110},
		MACD: &MACDValues{Histogram: 0.2},
		DMI: &DMIValues{ADX: 15, PlusDI: 22, MinusDI: 18},
	}
	
	confidence := calculateConfidence(results, TrendSignal, LongSignal)
	
	if confidence < 0.7 {
		t.Errorf("Expected base confidence (>=0.7), got %.2f", confidence)
	}
}

func TestCheckMACDSameSign(t *testing.T) {
	// Test same positive sign
	result := checkMACDSameSign(0.5, 0.3)
	if !result {
		t.Error("Expected true for same positive sign")
	}
	
	// Test same negative sign
	result = checkMACDSameSign(-0.5, -0.3)
	if !result {
		t.Error("Expected true for same negative sign")
	}
	
	// Test different signs
	result = checkMACDSameSign(0.5, -0.3)
	if result {
		t.Error("Expected false for different signs")
	}
}

func TestCheckDXADXFilter(t *testing.T) {
	// Test DX crossing above ADX
	result := checkDXADXFilter(25.0, 20.0, 15.0, 18.0)
	if !result {
		t.Error("Expected true for DX crossing above ADX")
	}
	
	// Test no crossing
	result = checkDXADXFilter(15.0, 20.0, 18.0, 18.0)
	if result {
		t.Error("Expected false for no crossing")
	}
}
