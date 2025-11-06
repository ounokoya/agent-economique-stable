// Package binance provides statistics calculation functionality
package binance

import "agent-economique/internal/shared"

// Helper functions for statistics calculation

// calculatePriceRange returns price range for klines data
func (pdp *ParsedDataProcessor) calculatePriceRange(klines []shared.KlineData) map[string]float64 {
	if len(klines) == 0 {
		return map[string]float64{"min": 0, "max": 0}
	}

	min := klines[0].Low
	max := klines[0].High

	for _, kline := range klines {
		if kline.Low < min {
			min = kline.Low
		}
		if kline.High > max {
			max = kline.High
		}
	}

	return map[string]float64{"min": min, "max": max}
}

// calculateTotalVolume returns total volume for klines data
func (pdp *ParsedDataProcessor) calculateTotalVolume(klines []shared.KlineData) float64 {
	var total float64
	for _, kline := range klines {
		total += kline.Volume
	}
	return total
}

// calculateAvgTrades returns average number of trades per kline
func (pdp *ParsedDataProcessor) calculateAvgTrades(klines []shared.KlineData) float64 {
	if len(klines) == 0 {
		return 0
	}

	var total int64
	for _, kline := range klines {
		total += kline.NumberOfTrades
	}

	return float64(total) / float64(len(klines))
}

// calculateTradesPriceRange returns price range for trades data
func (pdp *ParsedDataProcessor) calculateTradesPriceRange(trades []shared.TradeData) map[string]float64 {
	if len(trades) == 0 {
		return map[string]float64{"min": 0, "max": 0}
	}

	min := trades[0].Price
	max := trades[0].Price

	for _, trade := range trades {
		if trade.Price < min {
			min = trade.Price
		}
		if trade.Price > max {
			max = trade.Price
		}
	}

	return map[string]float64{"min": min, "max": max}
}

// calculateTradesTotalVolume returns total volume for trades data
func (pdp *ParsedDataProcessor) calculateTradesTotalVolume(trades []shared.TradeData) float64 {
	var total float64
	for _, trade := range trades {
		total += trade.Quantity
	}
	return total
}

// calculateBuyerMakerRatio returns the ratio of buyer maker trades
func (pdp *ParsedDataProcessor) calculateBuyerMakerRatio(trades []shared.TradeData) float64 {
	if len(trades) == 0 {
		return 0
	}

	var buyerMakerCount int
	for _, trade := range trades {
		if trade.IsBuyerMaker {
			buyerMakerCount++
		}
	}

	return float64(buyerMakerCount) / float64(len(trades))
}
