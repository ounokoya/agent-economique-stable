package stoch_mfi_cci

import (
	"math"
	
	"agent-economique/internal/indicators"
)

// generateSignals implements STOCH/MFI/CCI strategy signal generation
func generateSignals(results *indicators.IndicatorResults, config StrategyConfig) []indicators.StrategySignal {
	var signals []indicators.StrategySignal
	
	if results == nil || results.Stochastic == nil || results.MFI == nil || results.CCI == nil {
		return signals
	}
	
	stoch := results.Stochastic
	mfi := results.MFI
	cci := results.CCI
	
	// Check for LONG signal
	if longSignal := evaluateLONGSignal(stoch, mfi, cci, config); longSignal != nil {
		signals = append(signals, *longSignal)
	}
	
	// Check for SHORT signal  
	if shortSignal := evaluateSHORTSignal(stoch, mfi, cci, config); shortSignal != nil {
		signals = append(signals, *shortSignal)
	}
	
	return signals
}

// evaluateLONGSignal evaluates LONG signal conditions
func evaluateLONGSignal(stoch *indicators.StochasticValues, mfi *indicators.MFIValues, cci *indicators.CCIValues, config StrategyConfig) *indicators.StrategySignal {
	// STOCH must be in oversold zone with crossover up
	if stoch.Zone != indicators.StochOversold || stoch.CrossoverType != indicators.CrossUp {
		return nil
	}
	
	// At least one of MFI or CCI must be in oversold zone
	mfiOversold := mfi.Zone == indicators.MFIOversold
	cciOversold := cci.Zone == indicators.CCIOversold
	
	if !mfiOversold && !cciOversold {
		return nil
	}
	
	// Determine signal strength and confidence
	var confidence float64
	
	if mfiOversold && cciOversold {
		// Premium signal: all three indicators in favorable zones
		confidence = config.PremiumConfidence
	} else {
		// Minimal signal: STOCH + one other indicator
		confidence = config.MinConfidence
	}
	
	// Apply confidence bonuses
	confidence = applyConfidenceBonuses(confidence, stoch, mfi, cci, indicators.LongSignal)
	
	// Ensure confidence meets minimum threshold
	if confidence < config.MinConfidence {
		return nil
	}
	
	return &indicators.StrategySignal{
		Direction:  indicators.LongSignal,
		Type:       indicators.TrendSignal, // Will be refined by multi-timeframe
		Confidence: math.Min(confidence, 1.0),
		Timestamp:  0, // Will be set by caller
	}
}

// evaluateSHORTSignal evaluates SHORT signal conditions
func evaluateSHORTSignal(stoch *indicators.StochasticValues, mfi *indicators.MFIValues, cci *indicators.CCIValues, config StrategyConfig) *indicators.StrategySignal {
	// STOCH must be in overbought zone with crossover down
	if stoch.Zone != indicators.StochOverbought || stoch.CrossoverType != indicators.CrossDown {
		return nil
	}
	
	// At least one of MFI or CCI must be in overbought zone
	mfiOverbought := mfi.Zone == indicators.MFIOverbought
	cciOverbought := cci.Zone == indicators.CCIOverbought
	
	if !mfiOverbought && !cciOverbought {
		return nil
	}
	
	// Determine signal strength and confidence
	var confidence float64
	
	if mfiOverbought && cciOverbought {
		// Premium signal: all three indicators in favorable zones
		confidence = config.PremiumConfidence
	} else {
		// Minimal signal: STOCH + one other indicator
		confidence = config.MinConfidence
	}
	
	// Apply confidence bonuses
	confidence = applyConfidenceBonuses(confidence, stoch, mfi, cci, indicators.ShortSignal)
	
	// Ensure confidence meets minimum threshold
	if confidence < config.MinConfidence {
		return nil
	}
	
	return &indicators.StrategySignal{
		Direction:  indicators.ShortSignal,
		Type:       indicators.TrendSignal, // Will be refined by multi-timeframe
		Confidence: math.Min(confidence, 1.0),
		Timestamp:  0, // Will be set by caller
	}
}

// applyConfidenceBonuses applies confidence bonuses based on indicator alignment
func applyConfidenceBonuses(baseConfidence float64, stoch *indicators.StochasticValues, mfi *indicators.MFIValues, cci *indicators.CCIValues, direction indicators.SignalDirection) float64 {
	confidence := baseConfidence
	
	// Bonus for strong STOCH extreme values
	if direction == indicators.LongSignal && stoch.K < 10 {
		confidence += 0.05 // Very oversold bonus
	} else if direction == indicators.ShortSignal && stoch.K > 90 {
		confidence += 0.05 // Very overbought bonus
	}
	
	// Bonus for strong MFI extreme values
	if direction == indicators.LongSignal && mfi.Value < 10 {
		confidence += 0.03 // Very oversold MFI bonus
	} else if direction == indicators.ShortSignal && mfi.Value > 90 {
		confidence += 0.03 // Very overbought MFI bonus
	}
	
	// Bonus for strong CCI extreme values
	if direction == indicators.LongSignal && cci.Value < -150 {
		confidence += 0.03 // Very oversold CCI bonus
	} else if direction == indicators.ShortSignal && cci.Value > 150 {
		confidence += 0.03 // Very overbought CCI bonus
	}
	
	// Bonus for STOCH crossover strength (distance between K and D)
	kDDistance := math.Abs(stoch.K - stoch.D)
	if kDDistance > 5 {
		confidence += 0.02 // Strong crossover bonus
	}
	
	return confidence
}

// validateBarDirection validates that bar closed in the signal direction
func validateBarDirection(klines []indicators.Kline, direction indicators.SignalDirection) bool {
	if len(klines) == 0 {
		return false
	}
	
	lastBar := klines[len(klines)-1]
	
	switch direction {
	case indicators.LongSignal:
		return lastBar.Close > lastBar.Open // Bullish bar
	case indicators.ShortSignal:
		return lastBar.Close < lastBar.Open // Bearish bar
	default:
		return false
	}
}

// classifyMultiTimeframe classifies signal as trend or counter-trend based on higher timeframe
func classifyMultiTimeframe(signal *indicators.StrategySignal, higherTFResults *indicators.IndicatorResults) {
	if higherTFResults == nil || higherTFResults.Stochastic == nil {
		return // Keep default trend classification
	}
	
	htfStoch := higherTFResults.Stochastic
	
	// Classify based on higher timeframe STOCH alignment
	switch signal.Direction {
	case indicators.LongSignal:
		if htfStoch.Zone == indicators.StochOversold || htfStoch.Zone == indicators.StochNeutral {
			signal.Type = indicators.TrendSignal // Higher TF supports LONG
		} else {
			signal.Type = indicators.CounterTrendSignal // Higher TF is overbought
		}
	case indicators.ShortSignal:
		if htfStoch.Zone == indicators.StochOverbought || htfStoch.Zone == indicators.StochNeutral {
			signal.Type = indicators.TrendSignal // Higher TF supports SHORT
		} else {
			signal.Type = indicators.CounterTrendSignal // Higher TF is oversold
		}
	}
}

// GenerateStrategySignals is the main entry point for STOCH/MFI/CCI signal generation
func GenerateStrategySignals(results *indicators.IndicatorResults, klines []indicators.Kline, config StrategyConfig, higherTFResults *indicators.IndicatorResults) []indicators.StrategySignal {
	signals := generateSignals(results, config)
	
	// Set timestamp for all signals
	for i := range signals {
		signals[i].Timestamp = results.Timestamp
	}
	
	// Filter signals that don't meet bar confirmation requirement
	if config.RequireBarConfirmation {
		var validatedSignals []indicators.StrategySignal
		for _, signal := range signals {
			if validateBarDirection(klines, signal.Direction) {
				validatedSignals = append(validatedSignals, signal)
			}
		}
		signals = validatedSignals
	}
	
	// Apply multi-timeframe classification
	if config.EnableMultiTF && higherTFResults != nil {
		for i := range signals {
			classifyMultiTimeframe(&signals[i], higherTFResults)
		}
	}
	
	return signals
}

// EvaluateSignalConfidence evaluates signal confidence for display/logging
func EvaluateSignalConfidence(results *indicators.IndicatorResults, config StrategyConfig) float64 {
	signals := generateSignals(results, config)
	if len(signals) == 0 {
		return 0
	}
	
	// Return highest confidence signal
	maxConfidence := 0.0
	for _, signal := range signals {
		if signal.Confidence > maxConfidence {
			maxConfidence = signal.Confidence
		}
	}
	
	return maxConfidence
}

// GetSignalStrength determines signal strength for given indicator values
func GetSignalStrength(stoch *indicators.StochasticValues, mfi *indicators.MFIValues, cci *indicators.CCIValues, direction indicators.SignalDirection) SignalStrength {
	switch direction {
	case indicators.LongSignal:
		stochValid := stoch.Zone == indicators.StochOversold
		mfiValid := mfi.Zone == indicators.MFIOversold
		cciValid := cci.Zone == indicators.CCIOversold
		
		if stochValid && mfiValid && cciValid {
			return SignalPremium
		} else if stochValid && (mfiValid || cciValid) {
			return SignalMinimal
		}
		
	case indicators.ShortSignal:
		stochValid := stoch.Zone == indicators.StochOverbought
		mfiValid := mfi.Zone == indicators.MFIOverbought
		cciValid := cci.Zone == indicators.CCIOverbought
		
		if stochValid && mfiValid && cciValid {
			return SignalPremium
		} else if stochValid && (mfiValid || cciValid) {
			return SignalMinimal
		}
	}
	
	return SignalMinimal // Default fallback
}

// IsValidSTOCHCrossover validates STOCH crossover conditions
func IsValidSTOCHCrossover(stoch *indicators.StochasticValues, direction indicators.SignalDirection) bool {
	switch direction {
	case indicators.LongSignal:
		return stoch.Zone == indicators.StochOversold && stoch.CrossoverType == indicators.CrossUp
	case indicators.ShortSignal:
		return stoch.Zone == indicators.StochOverbought && stoch.CrossoverType == indicators.CrossDown
	default:
		return false
	}
}
