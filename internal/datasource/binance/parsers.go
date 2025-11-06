// Package binance provides data parsing functionality for Binance Vision data
package binance

import (
	"fmt"
	"math"
	"time"

	"agent-economique/internal/shared"
)

// ParsedDataProcessor processes batches of parsed data
type ParsedDataProcessor struct {
	cache     *CacheManager
	streaming *StreamingReader
	config    shared.AggregationConfig
}

// NewParsedDataProcessor creates a new ParsedDataProcessor instance
func NewParsedDataProcessor(cache *CacheManager, streaming *StreamingReader, config shared.AggregationConfig) (*ParsedDataProcessor, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}
	if streaming == nil {
		return nil, fmt.Errorf("streaming reader cannot be nil")
	}

	// Set default validation rules if not provided
	if config.ValidationRules.MaxPriceDeviation == 0 {
		config.ValidationRules.MaxPriceDeviation = 50.0 // 50% max price change
	}
	if config.ValidationRules.MaxVolumeDeviation == 0 {
		config.ValidationRules.MaxVolumeDeviation = 1000.0 // 1000% max volume change
	}
	if config.ValidationRules.MaxTimeGap == 0 {
		config.ValidationRules.MaxTimeGap = 300000 // 5 minutes in milliseconds
	}

	return &ParsedDataProcessor{
		cache:     cache,
		streaming: streaming,
		config:    config,
	}, nil
}

// ParseKlinesBatch processes klines data from a file into a parsed batch
func (pdp *ParsedDataProcessor) ParseKlinesBatch(filePath, symbol, timeframe, date string) (*shared.ParsedDataBatch, error) {
	if filePath == "" || symbol == "" || timeframe == "" || date == "" {
		return nil, fmt.Errorf("all parameters (filePath, symbol, timeframe, date) are required")
	}

	batch := &shared.ParsedDataBatch{
		Symbol:      symbol,
		DataType:    "klines",
		Timeframe:   timeframe,
		Date:        date,
		KlinesData:  make([]shared.KlineData, 0),
		ProcessedAt: time.Now(),
	}

	// Stream klines data from file
	err := pdp.streaming.StreamKlines(filePath, func(kline shared.KlineData) error {
		// Add to batch
		batch.KlinesData = append(batch.KlinesData, kline)
		batch.RecordCount++

		// Update time boundaries
		if batch.StartTime == 0 || kline.OpenTime < batch.StartTime {
			batch.StartTime = kline.OpenTime
		}
		if batch.EndTime == 0 || kline.CloseTime > batch.EndTime {
			batch.EndTime = kline.CloseTime
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse klines: %w", err)
	}

	return batch, nil
}

// ParseTradesBatch processes trades data from a file into a parsed batch
func (pdp *ParsedDataProcessor) ParseTradesBatch(filePath, symbol, date string) (*shared.ParsedDataBatch, error) {
	if filePath == "" || symbol == "" || date == "" {
		return nil, fmt.Errorf("all parameters (filePath, symbol, date) are required")
	}

	batch := &shared.ParsedDataBatch{
		Symbol:      symbol,
		DataType:    "trades",
		Date:        date,
		TradesData:  make([]shared.TradeData, 0),
		ProcessedAt: time.Now(),
	}

	// Stream trades data from file
	err := pdp.streaming.StreamTrades(filePath, func(trade shared.TradeData) error {
		// Add to batch
		batch.TradesData = append(batch.TradesData, trade)
		batch.RecordCount++

		// Update time boundaries
		if batch.StartTime == 0 || trade.Time < batch.StartTime {
			batch.StartTime = trade.Time
		}
		if batch.EndTime == 0 || trade.Time > batch.EndTime {
			batch.EndTime = trade.Time
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse trades: %w", err)
	}

	return batch, nil
}

// ValidateDataBatch validates a parsed data batch according to configuration rules
func (pdp *ParsedDataProcessor) ValidateDataBatch(batch *shared.ParsedDataBatch) (*shared.DataValidationResult, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	result := &shared.DataValidationResult{
		IsValid:     true,
		ProcessedAt: time.Now(),
	}

	// Validate based on data type
	if batch.DataType == "klines" {
		pdp.validateKlinesBatch(batch, result)
	} else if batch.DataType == "trades" {
		pdp.validateTradesBatch(batch, result)
	} else {
		result.IsValid = false
		result.ErrorCount++
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported data type: %s", batch.DataType))
	}

	// Set overall validity
	result.IsValid = result.ErrorCount == 0

	return result, nil
}

// ProcessAndValidateBatch combines parsing and validation in one step
func (pdp *ParsedDataProcessor) ProcessAndValidateBatch(filePath, symbol, dataType, date string, timeframe ...string) (*shared.ParsedDataBatch, *shared.DataValidationResult, error) {
	var batch *shared.ParsedDataBatch
	var err error

	// Parse based on data type
	if dataType == "klines" {
		if len(timeframe) == 0 {
			return nil, nil, fmt.Errorf("timeframe is required for klines data")
		}
		batch, err = pdp.ParseKlinesBatch(filePath, symbol, timeframe[0], date)
	} else if dataType == "trades" {
		batch, err = pdp.ParseTradesBatch(filePath, symbol, date)
	} else {
		return nil, nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Validate the batch
	validation, err := pdp.ValidateDataBatch(batch)
	if err != nil {
		return batch, nil, fmt.Errorf("validation failed: %w", err)
	}

	return batch, validation, nil
}

// GetBatchStatistics returns statistics for a parsed data batch
func (pdp *ParsedDataProcessor) GetBatchStatistics(batch *shared.ParsedDataBatch) (map[string]interface{}, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	stats := make(map[string]interface{})
	stats["symbol"] = batch.Symbol
	stats["data_type"] = batch.DataType
	stats["timeframe"] = batch.Timeframe
	stats["date"] = batch.Date
	stats["record_count"] = batch.RecordCount
	stats["start_time"] = batch.StartTime
	stats["end_time"] = batch.EndTime
	stats["processed_at"] = batch.ProcessedAt

	if batch.DataType == "klines" && len(batch.KlinesData) > 0 {
		stats["price_range"] = pdp.calculatePriceRange(batch.KlinesData)
		stats["total_volume"] = pdp.calculateTotalVolume(batch.KlinesData)
		stats["avg_trades_per_kline"] = pdp.calculateAvgTrades(batch.KlinesData)
	} else if batch.DataType == "trades" && len(batch.TradesData) > 0 {
		stats["price_range"] = pdp.calculateTradesPriceRange(batch.TradesData)
		stats["total_volume"] = pdp.calculateTradesTotalVolume(batch.TradesData)
		stats["buyer_maker_ratio"] = pdp.calculateBuyerMakerRatio(batch.TradesData)
	}

	return stats, nil
}

// Helper functions for validation

// validateKlinesBatch validates klines data according to rules
func (pdp *ParsedDataProcessor) validateKlinesBatch(batch *shared.ParsedDataBatch, result *shared.DataValidationResult) {
	if len(batch.KlinesData) == 0 {
		result.ErrorCount++
		result.Errors = append(result.Errors, "no klines data found")
		return
	}

	var prevKline *shared.KlineData
	for i, kline := range batch.KlinesData {
		// Validate individual kline
		if kline.OpenTime >= kline.CloseTime {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("kline %d: invalid time range", i))
		}

		if kline.High < kline.Low {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("kline %d: high < low", i))
		}

		if kline.Open <= 0 || kline.High <= 0 || kline.Low <= 0 || kline.Close <= 0 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("kline %d: invalid prices (must be > 0)", i))
		}

		if kline.Volume < 0 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("kline %d: negative volume", i))
		}

		// Validate sequence if we have a previous kline
		if prevKline != nil {
			// Check monotonic time if required
			if pdp.config.ValidationRules.RequireMonotonicTime && kline.OpenTime <= prevKline.OpenTime {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("kline %d: non-monotonic time", i))
			}

			// Check time gap
			timeGap := kline.OpenTime - prevKline.CloseTime
			if timeGap > pdp.config.ValidationRules.MaxTimeGap {
				result.WarningCount++
				result.Warnings = append(result.Warnings, fmt.Sprintf("kline %d: large time gap (%d ms)", i, timeGap))
			}

			// Check price deviation
			priceChange := math.Abs(kline.Open-prevKline.Close) / prevKline.Close * 100
			if priceChange > pdp.config.ValidationRules.MaxPriceDeviation {
				result.WarningCount++
				result.Warnings = append(result.Warnings, fmt.Sprintf("kline %d: large price change (%.2f%%)", i, priceChange))
			}
		}

		prevKline = &kline
	}
}

// validateTradesBatch validates trades data according to rules
func (pdp *ParsedDataProcessor) validateTradesBatch(batch *shared.ParsedDataBatch, result *shared.DataValidationResult) {
	if len(batch.TradesData) == 0 {
		result.ErrorCount++
		result.Errors = append(result.Errors, "no trades data found")
		return
	}

	var prevTrade *shared.TradeData
	for i, trade := range batch.TradesData {
		// Validate individual trade
		if trade.Price <= 0 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("trade %d: invalid price (must be > 0)", i))
		}

		if trade.Quantity <= 0 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("trade %d: invalid quantity (must be > 0)", i))
		}

		if trade.Time <= 0 {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("trade %d: invalid timestamp", i))
		}

		// Validate sequence if we have a previous trade
		if prevTrade != nil {
			// Check monotonic time if required
			if pdp.config.ValidationRules.RequireMonotonicTime && trade.Time < prevTrade.Time {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("trade %d: non-monotonic time", i))
			}

			// Check price deviation
			if prevTrade.Price > 0 {
				priceChange := math.Abs(trade.Price-prevTrade.Price) / prevTrade.Price * 100
				if priceChange > pdp.config.ValidationRules.MaxPriceDeviation {
					result.WarningCount++
					result.Warnings = append(result.Warnings, fmt.Sprintf("trade %d: large price change (%.2f%%)", i, priceChange))
				}
			}
		}

		prevTrade = &trade
	}
}
