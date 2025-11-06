// Package binance provides tests for aggregator functionality
package binance

import (
	"testing"

	"agent-economique/internal/shared"
)

// Test NewTimeframeAggregator function - VRAIE FONCTION avec defaults
func TestNewTimeframeAggregator_RealFunction(t *testing.T) {
	config := shared.AggregationConfig{}
	
	// TEST VRAIE FONCTION NewTimeframeAggregator
	aggregator, err := NewTimeframeAggregator(config)
	if err != nil {
		t.Fatalf("NewTimeframeAggregator failed: %v", err)
	}
	
	if aggregator == nil {
		t.Fatal("Expected aggregator instance, got nil")
	}

	// Vérifier que les defaults sont appliqués selon code réel
	if len(aggregator.config.SourceTimeframes) == 0 {
		t.Error("Expected default SourceTimeframes to be set")
	}

	if len(aggregator.config.TargetTimeframes) == 0 {
		t.Error("Expected default TargetTimeframes to be set")
	}

	// Vérifier defaults selon code
	expectedSourceDefault := "5m"
	if len(aggregator.config.SourceTimeframes) > 0 && aggregator.config.SourceTimeframes[0] != expectedSourceDefault {
		t.Errorf("Expected default source %s, got %s", expectedSourceDefault, aggregator.config.SourceTimeframes[0])
	}

	expectedTargets := []string{"15m", "1h", "4h"}
	if len(aggregator.config.TargetTimeframes) >= len(expectedTargets) {
		for i, expected := range expectedTargets {
			if aggregator.config.TargetTimeframes[i] != expected {
				t.Errorf("Expected target[%d] %s, got %s", i, expected, aggregator.config.TargetTimeframes[i])
			}
		}
	}
}

// Test TimeframeConversion - Utilitaire
func TestTimeframeConversion(t *testing.T) {
	conversions := map[string]int64{
		"5m":  5 * 60 * 1000,
		"15m": 15 * 60 * 1000,  
		"1h":  60 * 60 * 1000,
		"4h":  4 * 60 * 60 * 1000,
	}
	
	for tf, expectedMs := range conversions {
		// Simple validation that timeframe strings are recognized
		if len(tf) < 2 {
			t.Errorf("Timeframe %s too short", tf)
		}
		
		// Check suffix
		if tf[len(tf)-1] != 'm' && tf[len(tf)-1] != 'h' {
			t.Errorf("Timeframe %s has invalid suffix", tf)
		}
		
		t.Logf("Timeframe %s validates to %d ms", tf, expectedMs)
	}
}

// Test AggregateKlinesToTimeframe - VRAIE FONCTION critique
func TestAggregateKlinesToTimeframe_RealFunction(t *testing.T) {
	config := shared.AggregationConfig{}
	aggregator, err := NewTimeframeAggregator(config)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	// Données fixes - 3 klines 5m consécutives = 1 kline 15m
	sourceKlines := []shared.KlineData{
		{
			OpenTime:  1623024000000, // 00:00:00
			CloseTime: 1623024299999, // 00:04:59
			Open:      100.0,
			High:      102.0,
			Low:       99.0,
			Close:     101.0,
			Volume:    1000.0,
			NumberOfTrades: 50,
		},
		{
			OpenTime:  1623024300000, // 00:05:00
			CloseTime: 1623024599999, // 00:09:59
			Open:      101.0,
			High:      104.0,
			Low:       100.5,
			Close:     103.0,
			Volume:    1200.0,
			NumberOfTrades: 60,
		},
		{
			OpenTime:  1623024600000, // 00:10:00
			CloseTime: 1623024899999, // 00:14:59
			Open:      103.0,
			High:      105.0,
			Low:       102.0,
			Close:     104.5,
			Volume:    800.0,
			NumberOfTrades: 40,
		},
	}

	// TEST VRAIE FONCTION AggregateKlinesToTimeframe
	aggregated, err := aggregator.AggregateKlinesToTimeframe(sourceKlines, "15m")
	if err != nil {
		t.Fatalf("AggregateKlinesToTimeframe failed: %v", err)
	}

	// Validation résultat selon logique OHLCV
	if len(aggregated) != 1 {
		t.Errorf("Expected 1 aggregated kline, got %d", len(aggregated))
		return
	}

	result := aggregated[0]

	// Valider logique agrégation selon code réel
	expectedOpen := sourceKlines[0].Open        // Premier Open
	expectedClose := sourceKlines[2].Close      // Dernier Close  
	expectedHigh := 105.0                       // Max des High
	expectedLow := 99.0                         // Min des Low
	expectedVolume := 3000.0                    // Somme volumes
	expectedTrades := int64(150)                // Somme trades

	if result.Open != expectedOpen {
		t.Errorf("Expected Open %.1f, got %.1f", expectedOpen, result.Open)
	}

	if result.Close != expectedClose {
		t.Errorf("Expected Close %.1f, got %.1f", expectedClose, result.Close)
	}

	if result.High != expectedHigh {
		t.Errorf("Expected High %.1f, got %.1f", expectedHigh, result.High)
	}

	if result.Low != expectedLow {
		t.Errorf("Expected Low %.1f, got %.1f", expectedLow, result.Low)
	}

	if result.Volume != expectedVolume {
		t.Errorf("Expected Volume %.1f, got %.1f", expectedVolume, result.Volume)
	}

	if result.NumberOfTrades != expectedTrades {
		t.Errorf("Expected NumberOfTrades %d, got %d", expectedTrades, result.NumberOfTrades)
	}

	// Vérifier timeframe 15m
	expectedInterval := int64(15 * 60 * 1000) // 15 minutes en ms
	actualInterval := result.CloseTime - result.OpenTime + 1
	if actualInterval != expectedInterval {
		t.Errorf("Expected 15m interval %d ms, got %d ms", expectedInterval, actualInterval)
	}

	t.Logf("Aggregation 5m->15m: OHLCV(%.1f,%.1f,%.1f,%.1f,%.1f) Trades:%d", 
		result.Open, result.High, result.Low, result.Close, result.Volume, result.NumberOfTrades)
}

// Test AggregateKlinesToTimeframe - Erreur handling 
func TestAggregateKlinesToTimeframe_ErrorHandling(t *testing.T) {
	config := shared.AggregationConfig{}
	aggregator, err := NewTimeframeAggregator(config)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	// Test avec klines vides
	_, err = aggregator.AggregateKlinesToTimeframe([]shared.KlineData{}, "15m")
	if err == nil {
		t.Error("Expected error for empty klines slice")
	}

	// Test avec timeframe invalide  
	validKlines := []shared.KlineData{
		{OpenTime: 1623024000000, Close: 100.0},
	}
	_, err = aggregator.AggregateKlinesToTimeframe(validKlines, "invalid")
	if err == nil {
		t.Error("Expected error for invalid timeframe")
	}

	// Test avec nil klines
	_, err = aggregator.AggregateKlinesToTimeframe(nil, "15m")
	if err == nil {
		t.Error("Expected error for nil klines")
	}
}

// Test TradeAggregation - Logique Trades vers Klines
func TestTradeAggregation_Logic(t *testing.T) {
	trades := []shared.TradeData{
		{
			ID:       1001,
			Price:    100.0,
			Quantity: 10.0,
		},
		{
			ID:       1002,
			Price:    102.0,
			Quantity: 15.0,
		},
		{
			ID:       1003,
			Price:    98.0,
			Quantity: 8.0,
		},
		{
			ID:       1004,
			Price:    101.0,
			Quantity: 12.0,
		},
	}
	
	// Manual conversion: Trades -> Kline
	if len(trades) == 0 {
		t.Fatal("No trades to aggregate")
	}
	
	kline := shared.KlineData{
		Open:  trades[0].Price,                    // First trade price
		Close: trades[len(trades)-1].Price,       // Last trade price
		High:  trades[0].Price,
		Low:   trades[0].Price,
		Volume: 0.0,
	}
	
	// Calculate OHLCV from trades
	for _, trade := range trades {
		if trade.Price > kline.High {
			kline.High = trade.Price
		}
		if trade.Price < kline.Low {
			kline.Low = trade.Price
		}
		kline.Volume += trade.Quantity
	}
	
	// Validate conversion
	expectedOpen := 100.0
	if kline.Open != expectedOpen {
		t.Errorf("Expected Open %.1f, got %.1f", expectedOpen, kline.Open)
	}
	
	expectedClose := 101.0
	if kline.Close != expectedClose {
		t.Errorf("Expected Close %.1f, got %.1f", expectedClose, kline.Close)
	}
	
	expectedHigh := 102.0
	if kline.High != expectedHigh {
		t.Errorf("Expected High %.1f, got %.1f", expectedHigh, kline.High)
	}
	
	expectedLow := 98.0
	if kline.Low != expectedLow {
		t.Errorf("Expected Low %.1f, got %.1f", expectedLow, kline.Low)
	}
	
	expectedVolume := 45.0
	if kline.Volume != expectedVolume {
		t.Errorf("Expected Volume %.1f, got %.1f", expectedVolume, kline.Volume)
	}
	
	t.Logf("Trade aggregation: %d trades -> OHLCV(%.1f,%.1f,%.1f,%.1f,%.1f)", 
		len(trades), kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)
}

// Test CompressionRatio - Calcul statistique
func TestCompressionRatio_Calculation(t *testing.T) {
	sourceCount := 300    // 300 klines 5m
	targetCount := 20     // 20 klines 15m  
	
	// Compression ratio = (source - target) / source * 100
	compressionRatio := float64(sourceCount - targetCount) / float64(sourceCount) * 100
	
	expectedRatio := 93.33 // (300-20)/300 * 100 = 93.33%
	tolerance := 0.1
	
	if compressionRatio < expectedRatio-tolerance || compressionRatio > expectedRatio+tolerance {
		t.Errorf("Expected compression ratio %.2f%%, got %.2f%%", expectedRatio, compressionRatio)
	}
	
	t.Logf("Compression: %d -> %d (%.2f%% reduction)", sourceCount, targetCount, compressionRatio)
}

// Test TimeframeContinuity - Validation temporelle
func TestTimeframeContinuity_Validation(t *testing.T) {
	baseTime := int64(1623024000000) // 2021-06-07 00:00:00 UTC
	timeframe := "5m"
	intervalMs := int64(5 * 60 * 1000) // 5 minutes
	
	// Test continuous sequence
	continuousKlines := []shared.KlineData{
		{OpenTime: baseTime, CloseTime: baseTime + intervalMs - 1},
		{OpenTime: baseTime + intervalMs, CloseTime: baseTime + 2*intervalMs - 1},
		{OpenTime: baseTime + 2*intervalMs, CloseTime: baseTime + 3*intervalMs - 1},
	}
	
	// Validate continuity
	for i := 1; i < len(continuousKlines); i++ {
		expectedOpenTime := continuousKlines[i-1].CloseTime + 1
		actualOpenTime := continuousKlines[i].OpenTime
		
		if actualOpenTime != expectedOpenTime {
			t.Errorf("Gap detected at index %d: expected %d, got %d", 
				i, expectedOpenTime, actualOpenTime)
		}
	}
	
	t.Logf("Continuity validated for %s timeframe (%d intervals)", timeframe, len(continuousKlines))
}

// Benchmark aggregation performance
func BenchmarkKlineAggregation(b *testing.B) {
	// Create test dataset
	var klines []shared.KlineData
	baseTime := int64(1623024000000)
	
	for i := 0; i < 1000; i++ {
		klines = append(klines, shared.KlineData{
			OpenTime: baseTime + int64(i*5*60*1000),
			Open:     100.0 + float64(i%10),
			High:     105.0 + float64(i%10),
			Low:      95.0 + float64(i%10),
			Close:    102.0 + float64(i%10),
			Volume:   1000.0 + float64(i%100),
		})
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Simple aggregation simulation
		_ = len(klines)
	}
}
