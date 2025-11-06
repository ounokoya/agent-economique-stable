// Package testutils provides specialized assertions for trading data
package testutils

import (
	"fmt"
	"math"
	"testing"
	"time"

	"agent-economique/internal/shared"
)

// AssertKlineDataValid validates that kline data is well-formed and consistent
func AssertKlineDataValid(t *testing.T, kline shared.KlineData, msgAndArgs ...interface{}) {
	t.Helper()

	// Time validation
	if kline.OpenTime <= 0 {
		t.Errorf("OpenTime must be positive, got %d %v", kline.OpenTime, msgAndArgs)
	}
	if kline.CloseTime <= kline.OpenTime {
		t.Errorf("CloseTime (%d) must be greater than OpenTime (%d) %v", kline.CloseTime, kline.OpenTime, msgAndArgs)
	}

	// Price validation (OHLC)
	if kline.Open <= 0 {
		t.Errorf("Open price must be positive, got %f %v", kline.Open, msgAndArgs)
	}
	if kline.High <= 0 {
		t.Errorf("High price must be positive, got %f %v", kline.High, msgAndArgs)
	}
	if kline.Low <= 0 {
		t.Errorf("Low price must be positive, got %f %v", kline.Low, msgAndArgs)
	}
	if kline.Close <= 0 {
		t.Errorf("Close price must be positive, got %f %v", kline.Close, msgAndArgs)
	}

	// OHLC consistency
	if kline.High < kline.Low {
		t.Errorf("High (%f) cannot be less than Low (%f) %v", kline.High, kline.Low, msgAndArgs)
	}
	if kline.High < kline.Open {
		t.Errorf("High (%f) cannot be less than Open (%f) %v", kline.High, kline.Open, msgAndArgs)
	}
	if kline.High < kline.Close {
		t.Errorf("High (%f) cannot be less than Close (%f) %v", kline.High, kline.Close, msgAndArgs)
	}
	if kline.Low > kline.Open {
		t.Errorf("Low (%f) cannot be greater than Open (%f) %v", kline.Low, kline.Open, msgAndArgs)
	}
	if kline.Low > kline.Close {
		t.Errorf("Low (%f) cannot be greater than Close (%f) %v", kline.Low, kline.Close, msgAndArgs)
	}

	// Volume validation
	if kline.Volume < 0 {
		t.Errorf("Volume cannot be negative, got %f %v", kline.Volume, msgAndArgs)
	}
	if kline.QuoteAssetVolume < 0 {
		t.Errorf("QuoteAssetVolume cannot be negative, got %f %v", kline.QuoteAssetVolume, msgAndArgs)
	}
	if kline.TakerBuyBaseAssetVolume < 0 {
		t.Errorf("TakerBuyBaseAssetVolume cannot be negative, got %f %v", kline.TakerBuyBaseAssetVolume, msgAndArgs)
	}
	if kline.TakerBuyQuoteAssetVolume < 0 {
		t.Errorf("TakerBuyQuoteAssetVolume cannot be negative, got %f %v", kline.TakerBuyQuoteAssetVolume, msgAndArgs)
	}

	// Volume consistency
	if kline.TakerBuyBaseAssetVolume > kline.Volume {
		t.Errorf("TakerBuyBaseAssetVolume (%f) cannot exceed Volume (%f) %v", 
			kline.TakerBuyBaseAssetVolume, kline.Volume, msgAndArgs)
	}
	if kline.TakerBuyQuoteAssetVolume > kline.QuoteAssetVolume {
		t.Errorf("TakerBuyQuoteAssetVolume (%f) cannot exceed QuoteAssetVolume (%f) %v", 
			kline.TakerBuyQuoteAssetVolume, kline.QuoteAssetVolume, msgAndArgs)
	}

	// Trade count validation
	if kline.NumberOfTrades < 0 {
		t.Errorf("NumberOfTrades cannot be negative, got %d %v", kline.NumberOfTrades, msgAndArgs)
	}
}

// AssertTradeDataValid validates that trade data is well-formed and consistent
func AssertTradeDataValid(t *testing.T, trade shared.TradeData, msgAndArgs ...interface{}) {
	t.Helper()

	// ID validation
	if trade.ID <= 0 {
		t.Errorf("Trade ID must be positive, got %d %v", trade.ID, msgAndArgs)
	}

	// Price validation
	if trade.Price <= 0 {
		t.Errorf("Trade price must be positive, got %f %v", trade.Price, msgAndArgs)
	}

	// Quantity validation
	if trade.Quantity <= 0 {
		t.Errorf("Trade quantity must be positive, got %f %v", trade.Quantity, msgAndArgs)
	}

	// Quote quantity validation
	if trade.QuoteQty < 0 {
		t.Errorf("QuoteQty cannot be negative, got %f %v", trade.QuoteQty, msgAndArgs)
	}

	// Consistency check: QuoteQty should approximately equal Price * Quantity
	expectedQuoteQty := trade.Price * trade.Quantity
	tolerance := expectedQuoteQty * 0.001 // 0.1% tolerance for floating point precision
	if math.Abs(trade.QuoteQty-expectedQuoteQty) > tolerance {
		t.Errorf("QuoteQty (%f) inconsistent with Price*Quantity (%f), diff: %f %v", 
			trade.QuoteQty, expectedQuoteQty, math.Abs(trade.QuoteQty-expectedQuoteQty), msgAndArgs)
	}

	// Time validation
	if trade.Time <= 0 {
		t.Errorf("Trade time must be positive, got %d %v", trade.Time, msgAndArgs)
	}
}

// AssertKlineSequenceValid validates that a sequence of klines is temporally consistent
func AssertKlineSequenceValid(t *testing.T, klines []shared.KlineData, timeframe string, msgAndArgs ...interface{}) {
	t.Helper()

	if len(klines) < 2 {
		return // Cannot validate sequence with less than 2 klines
	}

	// Get expected interval for timeframe
	intervalMs, err := parseTimeframeInterval(timeframe)
	if err != nil {
		t.Errorf("Invalid timeframe '%s': %v %v", timeframe, err, msgAndArgs)
		return
	}

	for i := 1; i < len(klines); i++ {
		prev := klines[i-1]
		curr := klines[i]

		// Validate individual klines
		AssertKlineDataValid(t, prev, msgAndArgs...)
		AssertKlineDataValid(t, curr, msgAndArgs...)

		// Time sequence validation
		if curr.OpenTime <= prev.OpenTime {
			t.Errorf("Kline %d OpenTime (%d) must be greater than previous kline OpenTime (%d) %v", 
				i, curr.OpenTime, prev.OpenTime, msgAndArgs)
		}

		// Check for expected time interval (with some tolerance)
		expectedNextTime := prev.OpenTime + intervalMs
		timeDiff := curr.OpenTime - expectedNextTime
		if timeDiff != 0 && math.Abs(float64(timeDiff)) > float64(intervalMs)*0.1 { // 10% tolerance
			t.Errorf("Kline %d time gap unexpected: expected %d, got %d (diff: %d ms) %v", 
				i, expectedNextTime, curr.OpenTime, timeDiff, msgAndArgs)
		}

		// Price continuity check (gaps should be reasonable)
		priceGap := math.Abs(curr.Open - prev.Close)
		maxReasonableGap := (prev.Close + curr.Open) / 2 * 0.1 // 10% of average price
		if priceGap > maxReasonableGap {
			t.Logf("Warning: Large price gap between klines %d-%d: %f -> %f (gap: %f) %v", 
				i-1, i, prev.Close, curr.Open, priceGap, msgAndArgs)
		}
	}
}

// AssertTradeSequenceValid validates that a sequence of trades is temporally consistent
func AssertTradeSequenceValid(t *testing.T, trades []shared.TradeData, msgAndArgs ...interface{}) {
	t.Helper()

	if len(trades) < 2 {
		return // Cannot validate sequence with less than 2 trades
	}

	for i := 1; i < len(trades); i++ {
		prev := trades[i-1]
		curr := trades[i]

		// Validate individual trades
		AssertTradeDataValid(t, prev, msgAndArgs...)
		AssertTradeDataValid(t, curr, msgAndArgs...)

		// Time sequence validation
		if curr.Time < prev.Time {
			t.Errorf("Trade %d time (%d) cannot be earlier than previous trade time (%d) %v", 
				i, curr.Time, prev.Time, msgAndArgs)
		}

		// ID sequence validation (should be increasing)
		if curr.ID <= prev.ID {
			t.Errorf("Trade %d ID (%d) should be greater than previous trade ID (%d) %v", 
				i, curr.ID, prev.ID, msgAndArgs)
		}

		// Price reasonableness check
		priceChange := math.Abs(curr.Price - prev.Price)
		maxReasonableChange := (prev.Price + curr.Price) / 2 * 0.05 // 5% of average price
		if priceChange > maxReasonableChange {
			t.Logf("Warning: Large price change between trades %d-%d: %f -> %f (change: %f) %v", 
				i-1, i, prev.Price, curr.Price, priceChange, msgAndArgs)
		}
	}
}

// AssertParsedDataBatchValid validates a complete parsed data batch
func AssertParsedDataBatchValid(t *testing.T, batch *shared.ParsedDataBatch, msgAndArgs ...interface{}) {
	t.Helper()

	if batch == nil {
		t.Errorf("ParsedDataBatch cannot be nil %v", msgAndArgs)
		return
	}

	// Basic field validation
	if batch.Symbol == "" {
		t.Errorf("Symbol cannot be empty %v", msgAndArgs)
	}
	if batch.DataType == "" {
		t.Errorf("DataType cannot be empty %v", msgAndArgs)
	}
	if batch.Date == "" {
		t.Errorf("Date cannot be empty %v", msgAndArgs)
	}

	// Record count consistency
	if batch.DataType == "klines" {
		if len(batch.KlinesData) != batch.RecordCount {
			t.Errorf("KlinesData length (%d) doesn't match RecordCount (%d) %v", 
				len(batch.KlinesData), batch.RecordCount, msgAndArgs)
		}
		if batch.Timeframe == "" {
			t.Errorf("Timeframe cannot be empty for klines data %v", msgAndArgs)
		}
		if len(batch.KlinesData) > 0 {
			AssertKlineSequenceValid(t, batch.KlinesData, batch.Timeframe, msgAndArgs...)
		}
	} else if batch.DataType == "trades" {
		if len(batch.TradesData) != batch.RecordCount {
			t.Errorf("TradesData length (%d) doesn't match RecordCount (%d) %v", 
				len(batch.TradesData), batch.RecordCount, msgAndArgs)
		}
		if len(batch.TradesData) > 0 {
			AssertTradeSequenceValid(t, batch.TradesData, msgAndArgs...)
		}
	}

	// Time boundaries validation
	if batch.StartTime > batch.EndTime {
		t.Errorf("StartTime (%d) cannot be greater than EndTime (%d) %v", 
			batch.StartTime, batch.EndTime, msgAndArgs)
	}

	// ProcessedAt should be recent
	if time.Since(batch.ProcessedAt) > time.Hour {
		t.Errorf("ProcessedAt timestamp seems too old: %v %v", batch.ProcessedAt, msgAndArgs)
	}
}

// AssertValidationResultValid validates a data validation result
func AssertValidationResultValid(t *testing.T, result *shared.DataValidationResult, msgAndArgs ...interface{}) {
	t.Helper()

	if result == nil {
		t.Errorf("DataValidationResult cannot be nil %v", msgAndArgs)
		return
	}

	// Consistency checks
	if result.ErrorCount != len(result.Errors) {
		t.Errorf("ErrorCount (%d) doesn't match Errors length (%d) %v", 
			result.ErrorCount, len(result.Errors), msgAndArgs)
	}
	if result.WarningCount != len(result.Warnings) {
		t.Errorf("WarningCount (%d) doesn't match Warnings length (%d) %v", 
			result.WarningCount, len(result.Warnings), msgAndArgs)
	}

	// Validity logic
	if result.IsValid && result.ErrorCount > 0 {
		t.Errorf("Result cannot be valid with %d errors %v", result.ErrorCount, msgAndArgs)
	}
	if !result.IsValid && result.ErrorCount == 0 {
		t.Errorf("Result cannot be invalid with 0 errors %v", msgAndArgs)
	}

	// ProcessedAt should be recent
	if time.Since(result.ProcessedAt) > time.Hour {
		t.Errorf("ProcessedAt timestamp seems too old: %v %v", result.ProcessedAt, msgAndArgs)
	}
}

// AssertPerformanceWithin validates that execution time is within expected bounds
func AssertPerformanceWithin(t *testing.T, duration time.Duration, maxExpected time.Duration, operation string, msgAndArgs ...interface{}) {
	t.Helper()

	if duration > maxExpected {
		t.Errorf("%s took %v, expected ≤ %v %v", operation, duration, maxExpected, msgAndArgs)
	}

	// Also log if performance is surprisingly good (might indicate test issues)
	if duration < maxExpected/100 {
		t.Logf("Note: %s completed very quickly (%v), verify test is running properly %v", 
			operation, duration, msgAndArgs)
	}
}

// AssertMemoryUsageReasonable validates memory usage metrics
func AssertMemoryUsageReasonable(t *testing.T, metrics *shared.MemoryMetrics, maxMB float64, msgAndArgs ...interface{}) {
	t.Helper()

	if metrics == nil {
		t.Errorf("MemoryMetrics cannot be nil %v", msgAndArgs)
		return
	}

	// Current usage validation
	if metrics.CurrentUsageMB < 0 {
		t.Errorf("CurrentUsageMB cannot be negative, got %f %v", metrics.CurrentUsageMB, msgAndArgs)
	}
	if metrics.CurrentUsageMB > maxMB {
		t.Errorf("CurrentUsageMB (%f) exceeds limit (%f) %v", metrics.CurrentUsageMB, maxMB, msgAndArgs)
	}

	// Peak usage validation
	if metrics.PeakUsageMB < metrics.CurrentUsageMB {
		t.Errorf("PeakUsageMB (%f) cannot be less than CurrentUsageMB (%f) %v", 
			metrics.PeakUsageMB, metrics.CurrentUsageMB, msgAndArgs)
	}

	// Buffer count validation
	if metrics.BuffersActive < 0 {
		t.Errorf("BuffersActive cannot be negative, got %d %v", metrics.BuffersActive, msgAndArgs)
	}

	// Timestamp validation
	if time.Since(metrics.LastUpdateTime) > time.Hour {
		t.Errorf("LastUpdateTime seems too old: %v %v", metrics.LastUpdateTime, msgAndArgs)
	}
}

// AssertCacheStatsReasonable validates cache statistics
func AssertCacheStatsReasonable(t *testing.T, stats *shared.CacheStatistics, msgAndArgs ...interface{}) {
	t.Helper()

	if stats == nil {
		t.Errorf("CacheStatistics cannot be nil %v", msgAndArgs)
		return
	}

	// File count validation
	if stats.TotalFiles < 0 {
		t.Errorf("TotalFiles cannot be negative, got %d %v", stats.TotalFiles, msgAndArgs)
	}

	// Size validation
	if stats.TotalSizeMB < 0 {
		t.Errorf("TotalSizeMB cannot be negative, got %f %v", stats.TotalSizeMB, msgAndArgs)
	}

	// Hit rate validation
	if stats.HitRate < 0 || stats.HitRate > 1 {
		t.Errorf("HitRate must be between 0 and 1, got %f %v", stats.HitRate, msgAndArgs)
	}

	// Corrupted files validation
	if stats.CorruptedFiles < 0 {
		t.Errorf("CorruptedFiles cannot be negative, got %d %v", stats.CorruptedFiles, msgAndArgs)
	}
	if stats.CorruptedFiles > stats.TotalFiles {
		t.Errorf("CorruptedFiles (%d) cannot exceed TotalFiles (%d) %v", 
			stats.CorruptedFiles, stats.TotalFiles, msgAndArgs)
	}

	// Timestamp validation
	if time.Since(stats.LastUpdateTime) > time.Hour {
		t.Errorf("LastUpdateTime seems too old: %v %v", stats.LastUpdateTime, msgAndArgs)
	}
}

// Helper function to parse timeframe intervals
func parseTimeframeInterval(timeframe string) (int64, error) {
	switch timeframe {
	case "1m":
		return 60 * 1000, nil
	case "5m":
		return 5 * 60 * 1000, nil
	case "15m":
		return 15 * 60 * 1000, nil
	case "30m":
		return 30 * 60 * 1000, nil
	case "1h":
		return 60 * 60 * 1000, nil
	case "2h":
		return 2 * 60 * 60 * 1000, nil
	case "4h":
		return 4 * 60 * 60 * 1000, nil
	case "6h":
		return 6 * 60 * 60 * 1000, nil
	case "8h":
		return 8 * 60 * 60 * 1000, nil
	case "12h":
		return 12 * 60 * 60 * 1000, nil
	case "1d":
		return 24 * 60 * 60 * 1000, nil
	default:
		return 0, fmt.Errorf("unsupported timeframe: %s", timeframe)
	}
}

// AssertFloatEquals compares floating point numbers with tolerance
func AssertFloatEquals(t *testing.T, expected, actual, tolerance float64, msgAndArgs ...interface{}) {
	t.Helper()

	if math.Abs(expected-actual) > tolerance {
		t.Errorf("Expected %f, got %f (tolerance: %f) %v", expected, actual, tolerance, msgAndArgs)
	}
}

// AssertStringNotEmpty validates that a string is not empty
func AssertStringNotEmpty(t *testing.T, str, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()

	if str == "" {
		t.Errorf("%s cannot be empty %v", fieldName, msgAndArgs)
	}
}

// AssertPositiveInt64 validates that an int64 value is positive
func AssertPositiveInt64(t *testing.T, value int64, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()

	if value <= 0 {
		t.Errorf("%s must be positive, got %d %v", fieldName, value, msgAndArgs)
	}
}

// AssertPositiveFloat64 validates that a float64 value is positive
func AssertPositiveFloat64(t *testing.T, value float64, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()

	if value <= 0 {
		t.Errorf("%s must be positive, got %f %v", fieldName, value, msgAndArgs)
	}
}
