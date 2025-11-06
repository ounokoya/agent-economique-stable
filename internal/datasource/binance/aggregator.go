// Package binance provides multi-timeframe aggregation functionality
package binance

import (
	"fmt"
	"math"
	"sort"

	"agent-economique/internal/shared"
)

// TimeframeAggregator handles aggregation between different timeframes
type TimeframeAggregator struct {
	config shared.AggregationConfig
}

// NewTimeframeAggregator creates a new TimeframeAggregator instance
func NewTimeframeAggregator(config shared.AggregationConfig) (*TimeframeAggregator, error) {
	// Set default source and target timeframes if not provided
	if len(config.SourceTimeframes) == 0 {
		config.SourceTimeframes = []string{"5m"}
	}
	if len(config.TargetTimeframes) == 0 {
		config.TargetTimeframes = []string{"15m", "1h", "4h"}
	}

	return &TimeframeAggregator{
		config: config,
	}, nil
}

// AggregateKlinesToTimeframe aggregates 5m klines to higher timeframes (15m, 1h, 4h)
func (ta *TimeframeAggregator) AggregateKlinesToTimeframe(sourceKlines []shared.KlineData, targetTimeframe string) ([]shared.KlineData, error) {
	if len(sourceKlines) == 0 {
		return nil, fmt.Errorf("source klines cannot be empty")
	}

	// Get aggregation interval in milliseconds
	intervalMs, err := ta.getTimeframeInterval(targetTimeframe)
	if err != nil {
		return nil, fmt.Errorf("invalid target timeframe: %w", err)
	}

	// Sort klines by time to ensure proper order
	sort.Slice(sourceKlines, func(i, j int) bool {
		return sourceKlines[i].OpenTime < sourceKlines[j].OpenTime
	})

	var aggregatedKlines []shared.KlineData
	var currentAgg *shared.KlineData

	for _, kline := range sourceKlines {
		// Calculate the bucket start time for this kline
		bucketStart := ta.getBucketStartTime(kline.OpenTime, intervalMs)

		// If we don't have a current aggregation or this kline belongs to a new bucket
		if currentAgg == nil || currentAgg.OpenTime != bucketStart {
			// Save previous aggregation if it exists
			if currentAgg != nil {
				aggregatedKlines = append(aggregatedKlines, *currentAgg)
			}

			// Start new aggregation
			currentAgg = &shared.KlineData{
				OpenTime:                 bucketStart,
				CloseTime:                bucketStart + intervalMs - 1,
				Open:                     kline.Open,
				High:                     kline.High,
				Low:                      kline.Low,
				Close:                    kline.Close,
				Volume:                   kline.Volume,
				QuoteAssetVolume:        kline.QuoteAssetVolume,
				NumberOfTrades:          kline.NumberOfTrades,
				TakerBuyBaseAssetVolume: kline.TakerBuyBaseAssetVolume,
				TakerBuyQuoteAssetVolume: kline.TakerBuyQuoteAssetVolume,
				Ignore:                   "0",
			}
		} else {
			// Aggregate with existing bucket
			if kline.High > currentAgg.High {
				currentAgg.High = kline.High
			}
			if kline.Low < currentAgg.Low {
				currentAgg.Low = kline.Low
			}
			currentAgg.Close = kline.Close // Last close becomes the aggregated close
			currentAgg.Volume += kline.Volume
			currentAgg.QuoteAssetVolume += kline.QuoteAssetVolume
			currentAgg.NumberOfTrades += kline.NumberOfTrades
			currentAgg.TakerBuyBaseAssetVolume += kline.TakerBuyBaseAssetVolume
			currentAgg.TakerBuyQuoteAssetVolume += kline.TakerBuyQuoteAssetVolume
		}
	}

	// Don't forget the last aggregation
	if currentAgg != nil {
		aggregatedKlines = append(aggregatedKlines, *currentAgg)
	}

	return aggregatedKlines, nil
}

// AggregateTradestoKlines converts trades data into klines for a specific timeframe
func (ta *TimeframeAggregator) AggregateTradestoKlines(trades []shared.TradeData, timeframe string) ([]shared.KlineData, error) {
	if len(trades) == 0 {
		return nil, fmt.Errorf("trades cannot be empty")
	}

	// Get aggregation interval in milliseconds
	intervalMs, err := ta.getTimeframeInterval(timeframe)
	if err != nil {
		return nil, fmt.Errorf("invalid timeframe: %w", err)
	}

	// Sort trades by time
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Time < trades[j].Time
	})

	var klines []shared.KlineData
	var currentKline *shared.KlineData

	for _, trade := range trades {
		// Calculate the bucket start time for this trade
		bucketStart := ta.getBucketStartTime(trade.Time, intervalMs)

		// If we don't have a current kline or this trade belongs to a new bucket
		if currentKline == nil || currentKline.OpenTime != bucketStart {
			// Save previous kline if it exists
			if currentKline != nil {
				klines = append(klines, *currentKline)
			}

			// Start new kline from this trade
			currentKline = &shared.KlineData{
				OpenTime:                 bucketStart,
				CloseTime:                bucketStart + intervalMs - 1,
				Open:                     trade.Price,
				High:                     trade.Price,
				Low:                      trade.Price,
				Close:                    trade.Price,
				Volume:                   trade.Quantity,
				QuoteAssetVolume:        trade.QuoteQty,
				NumberOfTrades:          1,
				TakerBuyBaseAssetVolume: 0,
				TakerBuyQuoteAssetVolume: 0,
				Ignore:                   "0",
			}

			// Add to taker buy volumes if this is a buyer maker trade
			if trade.IsBuyerMaker {
				currentKline.TakerBuyBaseAssetVolume = trade.Quantity
				currentKline.TakerBuyQuoteAssetVolume = trade.QuoteQty
			}
		} else {
			// Aggregate with existing kline
			if trade.Price > currentKline.High {
				currentKline.High = trade.Price
			}
			if trade.Price < currentKline.Low {
				currentKline.Low = trade.Price
			}
			currentKline.Close = trade.Price // Last price becomes close
			currentKline.Volume += trade.Quantity
			currentKline.QuoteAssetVolume += trade.QuoteQty
			currentKline.NumberOfTrades++

			// Add to taker buy volumes if this is a buyer maker trade
			if trade.IsBuyerMaker {
				currentKline.TakerBuyBaseAssetVolume += trade.Quantity
				currentKline.TakerBuyQuoteAssetVolume += trade.QuoteQty
			}
		}
	}

	// Don't forget the last kline
	if currentKline != nil {
		klines = append(klines, *currentKline)
	}

	return klines, nil
}

// ValidateTimeframeContinuity checks if klines have proper time continuity
func (ta *TimeframeAggregator) ValidateTimeframeContinuity(klines []shared.KlineData, timeframe string) error {
	if len(klines) < 2 {
		return nil // Not enough data to validate continuity
	}

	intervalMs, err := ta.getTimeframeInterval(timeframe)
	if err != nil {
		return fmt.Errorf("invalid timeframe: %w", err)
	}

	// Sort by time first
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime < klines[j].OpenTime
	})

	for i := 1; i < len(klines); i++ {
		expectedOpenTime := klines[i-1].OpenTime + intervalMs
		actualOpenTime := klines[i].OpenTime

		// Allow some tolerance for small gaps, but detect major issues
		if actualOpenTime != expectedOpenTime {
			gap := actualOpenTime - expectedOpenTime
			if gap < 0 || gap > intervalMs {
				return fmt.Errorf("time continuity broken between klines %d and %d: expected %d, got %d (gap: %d ms)",
					i-1, i, expectedOpenTime, actualOpenTime, gap)
			}
		}
	}

	return nil
}

// SynchronizeMultiTimeframes ensures all timeframes have aligned data
func (ta *TimeframeAggregator) SynchronizeMultiTimeframes(klinesMap map[string][]shared.KlineData) (map[string][]shared.KlineData, error) {
	if len(klinesMap) == 0 {
		return nil, fmt.Errorf("klines map cannot be empty")
	}

	synchronized := make(map[string][]shared.KlineData)

	// Find the common time range across all timeframes
	var minStartTime, maxEndTime int64
	first := true

	for _, klines := range klinesMap {
		if len(klines) == 0 {
			continue
		}

		// Sort klines by time
		sort.Slice(klines, func(i, j int) bool {
			return klines[i].OpenTime < klines[j].OpenTime
		})

		startTime := klines[0].OpenTime
		endTime := klines[len(klines)-1].CloseTime

		if first {
			minStartTime = startTime
			maxEndTime = endTime
			first = false
		} else {
			if startTime > minStartTime {
				minStartTime = startTime
			}
			if endTime < maxEndTime {
				maxEndTime = endTime
			}
		}
	}

	// Filter each timeframe to the common range
	for timeframe, klines := range klinesMap {
		var filteredKlines []shared.KlineData

		for _, kline := range klines {
			if kline.OpenTime >= minStartTime && kline.CloseTime <= maxEndTime {
				filteredKlines = append(filteredKlines, kline)
			}
		}

		synchronized[timeframe] = filteredKlines
	}

	return synchronized, nil
}

// GetAggregationStatistics returns statistics about the aggregation process
func (ta *TimeframeAggregator) GetAggregationStatistics(originalCount, aggregatedCount int, timeframe string) map[string]interface{} {
	stats := make(map[string]interface{})
	stats["original_count"] = originalCount
	stats["aggregated_count"] = aggregatedCount
	stats["target_timeframe"] = timeframe
	stats["compression_ratio"] = float64(originalCount) / math.Max(float64(aggregatedCount), 1)
	
	intervalMs, _ := ta.getTimeframeInterval(timeframe)
	stats["interval_ms"] = intervalMs
	stats["interval_minutes"] = intervalMs / (60 * 1000)
	
	return stats
}

// Helper functions

// getTimeframeInterval converts timeframe string to milliseconds
func (ta *TimeframeAggregator) getTimeframeInterval(timeframe string) (int64, error) {
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

// getBucketStartTime calculates the start time of the bucket for a given timestamp
func (ta *TimeframeAggregator) getBucketStartTime(timestamp, intervalMs int64) int64 {
	// Round down to the nearest interval boundary
	return (timestamp / intervalMs) * intervalMs
}
