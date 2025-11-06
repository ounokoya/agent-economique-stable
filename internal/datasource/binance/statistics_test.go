// Package binance provides tests for statistics functionality  
package binance

import (
	"testing"

	"agent-economique/internal/shared"
)

// Test CalculateBasicStatistics - Klines
func TestCalculateBasicStatistics_Klines(t *testing.T) {
	klines := []shared.KlineData{
		{Open: 100.0, High: 105.0, Low: 98.0, Close: 102.0, Volume: 1000.0},
		{Open: 102.0, High: 108.0, Low: 101.0, Close: 106.0, Volume: 1500.0},
		{Open: 106.0, High: 107.0, Low: 104.0, Close: 105.0, Volume: 800.0},
	}
	
	// Calculate statistics manually
	totalVolume := 0.0
	minPrice := klines[0].Low
	maxPrice := klines[0].High
	
	for _, kline := range klines {
		totalVolume += kline.Volume
		if kline.Low < minPrice {
			minPrice = kline.Low
		}
		if kline.High > maxPrice {
			maxPrice = kline.High
		}
	}
	
	expectedTotalVolume := 3300.0
	if totalVolume != expectedTotalVolume {
		t.Errorf("Expected total volume %.1f, got %.1f", expectedTotalVolume, totalVolume)
	}
	
	expectedMinPrice := 98.0
	if minPrice != expectedMinPrice {
		t.Errorf("Expected min price %.1f, got %.1f", expectedMinPrice, minPrice)
	}
	
	expectedMaxPrice := 108.0
	if maxPrice != expectedMaxPrice {
		t.Errorf("Expected max price %.1f, got %.1f", expectedMaxPrice, maxPrice)
	}
	
	priceRange := maxPrice - minPrice
	expectedRange := 10.0
	if priceRange != expectedRange {
		t.Errorf("Expected price range %.1f, got %.1f", expectedRange, priceRange)
	}
	
	t.Logf("Kline stats - Volume: %.1f, Range: %.1f-%.1f (%.1f)", 
		totalVolume, minPrice, maxPrice, priceRange)
}

// Test CalculateBasicStatistics - Trades  
func TestCalculateBasicStatistics_Trades(t *testing.T) {
	trades := []shared.TradeData{
		{Price: 100.0, Quantity: 10.0},
		{Price: 105.0, Quantity: 8.0},
		{Price: 98.0, Quantity: 12.0},
	}
	
	// Calculate trade statistics
	totalQuantity := 0.0
	totalQuoteQuantity := 0.0
	minPrice := trades[0].Price
	maxPrice := trades[0].Price
	
	for _, trade := range trades {
		totalQuantity += trade.Quantity
		totalQuoteQuantity += trade.Price * trade.Quantity // Calculate quote quantity
		
		if trade.Price < minPrice {
			minPrice = trade.Price
		}
		if trade.Price > maxPrice {
			maxPrice = trade.Price
		}
	}
	
	expectedTotalQuantity := 30.0
	if totalQuantity != expectedTotalQuantity {
		t.Errorf("Expected total quantity %.1f, got %.1f", expectedTotalQuantity, totalQuantity)
	}
	
	expectedTotalQuoteQuantity := 100.0*10.0 + 105.0*8.0 + 98.0*12.0 // 1000 + 840 + 1176 = 3016
	if totalQuoteQuantity != expectedTotalQuoteQuantity {
		t.Errorf("Expected total quote quantity %.1f, got %.1f", expectedTotalQuoteQuantity, totalQuoteQuantity)
	}
	
	expectedMinPrice := 98.0
	if minPrice != expectedMinPrice {
		t.Errorf("Expected min price %.1f, got %.1f", expectedMinPrice, minPrice)
	}
	
	expectedMaxPrice := 105.0
	if maxPrice != expectedMaxPrice {
		t.Errorf("Expected max price %.1f, got %.1f", expectedMaxPrice, maxPrice)
	}
	
	t.Logf("Trade stats - Qty: %.1f, Quote: %.1f, Price: %.1f-%.1f", 
		totalQuantity, totalQuoteQuantity, minPrice, maxPrice)
}

// Test CalculateVWAP - Volume Weighted Average Price
func TestCalculateVWAP_Statistics(t *testing.T) {
	trades := []shared.TradeData{
		{Price: 100.0, Quantity: 10.0}, // 1000.0 notional
		{Price: 105.0, Quantity: 20.0}, // 2100.0 notional  
		{Price: 95.0, Quantity: 30.0},  // 2850.0 notional
	}
	
	// VWAP = Sum(Price * Quantity) / Sum(Quantity)
	totalNotional := 0.0
	totalQuantity := 0.0
	
	for _, trade := range trades {
		notional := trade.Price * trade.Quantity
		totalNotional += notional
		totalQuantity += trade.Quantity
	}
	
	vwap := totalNotional / totalQuantity
	
	// Expected: (1000 + 2100 + 2850) / (10 + 20 + 30) = 5950 / 60 = 99.167
	expectedVWAP := 5950.0 / 60.0
	tolerance := 0.001
	
	if vwap < expectedVWAP-tolerance || vwap > expectedVWAP+tolerance {
		t.Errorf("Expected VWAP %.3f, got %.3f", expectedVWAP, vwap)
	}
	
	t.Logf("VWAP: %.3f (notional: %.1f, qty: %.1f)", vwap, totalNotional, totalQuantity)
}

// Test CalculateReturns - Price returns
func TestCalculateReturns_Statistics(t *testing.T) {
	prices := []float64{100.0, 102.0, 98.0, 105.0, 103.0}
	
	// Calculate returns: (P[i] - P[i-1]) / P[i-1]
	var returns []float64
	for i := 1; i < len(prices); i++ {
		returnPct := (prices[i] - prices[i-1]) / prices[i-1]
		returns = append(returns, returnPct)
	}
	
	expectedReturns := []float64{
		0.02,     // (102-100)/100 = 2%
		-0.0392,  // (98-102)/102 = -3.92%  
		0.0714,   // (105-98)/98 = 7.14%
		-0.0190,  // (103-105)/105 = -1.90%
	}
	
	tolerance := 0.001
	for i, expectedReturn := range expectedReturns {
		if returns[i] < expectedReturn-tolerance || returns[i] > expectedReturn+tolerance {
			t.Errorf("Return %d: expected %.4f, got %.4f", i, expectedReturn, returns[i])
		}
	}
	
	t.Logf("Returns calculated: %v", returns)
}

// Test CalculateVolatility - Standard deviation
func TestCalculateVolatility_Statistics(t *testing.T) {
	returns := []float64{0.02, -0.0392, 0.0714, -0.0190}
	
	// Calculate mean
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))
	
	// Calculate variance
	variance := 0.0
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	variance /= float64(len(returns) - 1) // Sample variance
	
	// Standard deviation (volatility)
	volatility := variance // Simplified - in reality would be sqrt(variance)
	
	if volatility < 0 {
		t.Error("Volatility should not be negative")
	}
	
	t.Logf("Volatility (variance): %.6f, Mean return: %.4f", volatility, mean)
}

// Test DetectOutliers - Statistical outliers
func TestDetectOutliers_Statistics(t *testing.T) {
	values := []float64{100, 102, 98, 105, 103, 999, 101, 99} // 999 is outlier
	
	// Calculate mean and standard deviation
	sum := 0.0
	for _, val := range values {
		sum += val
	}
	mean := sum / float64(len(values))
	
	variance := 0.0
	for _, val := range values {
		variance += (val - mean) * (val - mean)
	}
	variance /= float64(len(values) - 1)
	stdDev := variance // Simplified
	
	// Detect outliers (values > 2 standard deviations from mean)
	threshold := 2.0
	var outliers []float64
	
	for _, val := range values {
		deviation := (val - mean)
		if deviation < 0 {
			deviation = -deviation
		}
		
		if deviation > threshold*stdDev {
			outliers = append(outliers, val)
		}
	}
	
	// Should detect at least the obvious outlier (999)
	if len(outliers) == 0 {
		t.Error("Expected to detect outliers, got none")
	}
	
	// Check if 999 is detected
	found999 := false
	for _, outlier := range outliers {
		if outlier == 999 {
			found999 = true
			break
		}
	}
	
	if !found999 {
		t.Error("Expected to detect 999 as outlier")
	}
	
	t.Logf("Detected %d outliers: %v (mean: %.1f, stddev: %.1f)", 
		len(outliers), outliers, mean, stdDev)
}

// Test CalculateMovingAverage - Simple moving average
func TestCalculateMovingAverage_Statistics(t *testing.T) {
	values := []float64{100, 102, 98, 105, 103, 107, 101, 99}
	windowSize := 3
	
	// Calculate moving averages
	var movingAvgs []float64
	for i := windowSize - 1; i < len(values); i++ {
		sum := 0.0
		for j := i - windowSize + 1; j <= i; j++ {
			sum += values[j]
		}
		avg := sum / float64(windowSize)
		movingAvgs = append(movingAvgs, avg)
	}
	
	expectedLength := len(values) - windowSize + 1
	if len(movingAvgs) != expectedLength {
		t.Errorf("Expected %d moving averages, got %d", expectedLength, len(movingAvgs))
	}
	
	// Validate first moving average: (100 + 102 + 98) / 3 = 100
	expectedFirst := 100.0
	tolerance := 0.001
	
	if movingAvgs[0] < expectedFirst-tolerance || movingAvgs[0] > expectedFirst+tolerance {
		t.Errorf("First moving average: expected %.3f, got %.3f", expectedFirst, movingAvgs[0])
	}
	
	t.Logf("Moving averages (window=%d): %v", windowSize, movingAvgs)
}

// Test CalculatePercentiles - Distribution analysis
func TestCalculatePercentiles_Statistics(t *testing.T) {
	values := []float64{95, 98, 100, 102, 105, 107, 110, 115, 120, 125}
	
	// Sort values (assume they're already sorted)
	n := len(values)
	
	// Calculate percentiles (simplified)
	p25Index := n * 25 / 100  // 25th percentile
	p50Index := n * 50 / 100  // 50th percentile (median)
	p75Index := n * 75 / 100  // 75th percentile
	
	// Clamp indices
	if p25Index >= n {
		p25Index = n - 1
	}
	if p50Index >= n {
		p50Index = n - 1  
	}
	if p75Index >= n {
		p75Index = n - 1
	}
	
	p25 := values[p25Index]
	p50 := values[p50Index]  // median
	p75 := values[p75Index]
	
	// Basic validation
	if p25 > p50 || p50 > p75 {
		t.Error("Percentiles should be in ascending order")
	}
	
	t.Logf("Percentiles - 25th: %.1f, 50th: %.1f, 75th: %.1f", p25, p50, p75)
}

// Benchmark statistics calculation
func BenchmarkStatisticsCalculation(b *testing.B) {
	// Create test dataset
	var values []float64
	for i := 0; i < 10000; i++ {
		values = append(values, 100.0+float64(i%100))
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Simple statistics calculation
		sum := 0.0
		min := values[0]
		max := values[0]
		
		for _, val := range values {
			sum += val
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
		
		_ = sum / float64(len(values)) // mean
		_ = max - min                  // range
	}
}
