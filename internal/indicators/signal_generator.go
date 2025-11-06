package indicators

import "math"

// generateStrategySignals implements pure MACD/CCI/DMI strategy from user memory
func generateStrategySignals(results *IndicatorResults) []StrategySignal {
	var signals []StrategySignal
	
	if results == nil || results.MACD == nil || results.CCI == nil || results.DMI == nil {
		return signals
	}
	
	// Extract indicator states
	macdCross := results.MACD.CrossoverType
	cciZone := results.CCI.Zone
	diPlus := results.DMI.PlusDI
	diMinus := results.DMI.MinusDI
	
	// Strategy from user memory:
	// LONG: MACD croise à la hausse + CCI en survente + (DI+ > DI- pour tendance OU DI- > DI+ pour contre-tendance)
	if macdCross == CrossUp && cciZone == CCIOversold {
		if diPlus > diMinus {
			// LONG Tendance: MACD↗ + CCI survente + DI+ > DI-
			signals = append(signals, StrategySignal{
				Direction:  LongSignal,
				Type:       TrendSignal,
				Confidence: calculateConfidence(results, TrendSignal, LongSignal),
				Timestamp:  results.Timestamp,
			})
		} else if diMinus > diPlus {
			// LONG Contre-tendance: MACD↗ + CCI survente + DI- > DI+
			signals = append(signals, StrategySignal{
				Direction:  LongSignal,
				Type:       CounterTrendSignal,
				Confidence: calculateConfidence(results, CounterTrendSignal, LongSignal),
				Timestamp:  results.Timestamp,
			})
		}
	}
	
	// SHORT: MACD croise à la baisse + CCI en surachat + (DI+ < DI- pour tendance OU DI+ > DI- pour contre-tendance)
	if macdCross == CrossDown && cciZone == CCIOverbought {
		if diPlus < diMinus {
			// SHORT Tendance: MACD↘ + CCI surachat + DI+ < DI-
			signals = append(signals, StrategySignal{
				Direction:  ShortSignal,
				Type:       TrendSignal,
				Confidence: calculateConfidence(results, TrendSignal, ShortSignal),
				Timestamp:  results.Timestamp,
			})
		} else if diPlus > diMinus {
			// SHORT Contre-tendance: MACD↘ + CCI surachat + DI+ > DI-
			signals = append(signals, StrategySignal{
				Direction:  ShortSignal,
				Type:       CounterTrendSignal,
				Confidence: calculateConfidence(results, CounterTrendSignal, ShortSignal),
				Timestamp:  results.Timestamp,
			})
		}
	}
	
	return signals
}

// calculateConfidence computes signal confidence based on indicator alignment
func calculateConfidence(results *IndicatorResults, sigType SignalType, direction SignalDirection) float64 {
	baseConfidence := 0.7 // Base confidence for strategy match
	
	// Bonus for strong indicator values
	var bonuses float64
	
	// CCI strength bonus
	cciValue := results.CCI.Value
	if direction == LongSignal && cciValue < -150 {
		bonuses += 0.1 // Very oversold
	} else if direction == ShortSignal && cciValue > 150 {
		bonuses += 0.1 // Very overbought
	}
	
	// MACD histogram strength bonus
	histValue := math.Abs(results.MACD.Histogram)
	if histValue > 0.5 {
		bonuses += 0.1 // Strong momentum
	}
	
	// ADX strength bonus (trend strength)
	adxValue := results.DMI.ADX
	if sigType == TrendSignal && adxValue > 25 {
		bonuses += 0.1 // Strong trend
	}
	
	// DI spread bonus
	diSpread := math.Abs(results.DMI.PlusDI - results.DMI.MinusDI)
	if diSpread > 10 {
		bonuses += 0.05 // Clear directional bias
	}
	
	confidence := baseConfidence + bonuses
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// Optional filters from user memory (for future enhancement)

// checkMACDSameSign checks if MACD and Signal have same sign at crossover
func checkMACDSameSign(macd, signal float64) bool {
	return (macd >= 0 && signal >= 0) || (macd < 0 && signal < 0)
}

// checkDXADXFilter checks DX/ADX crossover filter (advanced filter)
func checkDXADXFilter(dx, adx float64, prevDX, prevADX float64) bool {
	// DX crossing above ADX (strengthening trend)
	return prevDX <= prevADX && dx > adx
}
