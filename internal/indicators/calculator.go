package indicators

import (
	"fmt"
	"time"
)

// Calculate orchestrates indicator calculations for Engine Temporel
// Uses existing indicator functions without data validation (Engine already validated)
func Calculate(request *CalculationRequest) *CalculationResponse {
	startTime := time.Now()
	
	response := &CalculationResponse{
		RequestID: request.RequestID,
		Success:   true,
	}
	
	// Engine provides already validated data (>= 35 candles), no validation needed
	klines := request.CandleWindow
	if len(klines) == 0 {
		response.Success = false
		response.Error = fmt.Errorf("empty candle window")
		return response
	}
	
	// Calculate indicators using existing functions
	results, err := calculateIndicators(klines, request.CurrentTime)
	if err != nil {
		response.Success = false
		response.Error = err
		return response
	}
	
	// Generate strategy signals using pure logic
	signals := generateStrategySignals(results)
	
	// Detect zone events if position context provided
	var zoneEvents []ZoneEvent
	if request.PositionContext != nil {
		zoneEvents = detectZoneEvents(results, request.PositionContext)
	}
	
	// Assemble response
	response.Results = results
	response.Signals = signals
	response.ZoneEvents = zoneEvents
	response.CalculationTime = time.Since(startTime)
	
	return response
}

// calculateIndicators performs pure calculations using existing indicator functions
func calculateIndicators(klines []Kline, currentTime int64) (*IndicatorResults, error) {
	// Use existing MACD function - no validation needed (Engine validated)
	macdLine, signalLine, histogram := MACDFromKlines(klines, 12, 26, 9, func(k Kline) float64 { 
		return k.Close 
	})
	
	// Use existing CCI function 
	cciValues := CCIFromKlines(klines, "hlc3", 14)
	
	// Use existing DMI function
	plusDI, minusDI, dx, adx := DMIFromKlines(klines, 14)
	
	// Use existing Stochastic function
	stochK, stochD := StochasticFromKlines(klines, 14, 3, 3)
	
	// Use existing MFI function
	mfiValues := MFIFromKlines(klines, 14)
	
	// Get latest values (last calculated point)
	lastIdx := len(klines) - 1
	if lastIdx < 0 {
		return nil, fmt.Errorf("insufficient data")
	}
	
	// Detect MACD crossover
	macdCrossover := detectMACDCrossover(macdLine, signalLine)
	
	// Detect Stochastic crossover
	stochCrossover := detectStochasticCrossover(stochK, stochD)
	
	// Classify zones
	cciZone := classifyCCIZone(cciValues[lastIdx])
	stochZone := classifyStochZone(stochK[lastIdx])
	mfiZone := classifyMFIZone(mfiValues[lastIdx])
	
	return &IndicatorResults{
		MACD: &MACDValues{
			MACD:          macdLine[lastIdx],
			Signal:        signalLine[lastIdx],
			Histogram:     histogram[lastIdx],
			CrossoverType: macdCrossover,
		},
		CCI: &CCIValues{
			Value: cciValues[lastIdx],
			Zone:  cciZone,
		},
		DMI: &DMIValues{
			PlusDI:  plusDI[lastIdx],
			MinusDI: minusDI[lastIdx],
			DX:      dx[lastIdx],
			ADX:     adx[lastIdx],
		},
		Stochastic: &StochasticValues{
			K:             stochK[lastIdx],
			D:             stochD[lastIdx],
			Zone:          stochZone,
			CrossoverType: stochCrossover,
			IsExtreme:     stochZone == StochOversold || stochZone == StochOverbought,
		},
		MFI: &MFIValues{
			Value:     mfiValues[lastIdx],
			Zone:      mfiZone,
			IsExtreme: mfiZone == MFIOversold || mfiZone == MFIOverbought,
		},
		Timestamp: currentTime,
	}, nil
}

// detectMACDCrossover detects crossover between MACD and Signal lines
func detectMACDCrossover(macdLine, signalLine []float64) CrossoverType {
	if len(macdLine) < 2 || len(signalLine) < 2 {
		return NoCrossover
	}
	
	lastIdx := len(macdLine) - 1
	prevIdx := lastIdx - 1
	
	// Current and previous values
	macdCurr := macdLine[lastIdx]
	macdPrev := macdLine[prevIdx]
	signalCurr := signalLine[lastIdx]
	signalPrev := signalLine[prevIdx]
	
	// Check for crossover
	if macdPrev <= signalPrev && macdCurr > signalCurr {
		return CrossUp
	} else if macdPrev >= signalPrev && macdCurr < signalCurr {
		return CrossDown
	}
	
	return NoCrossover
}

// detectStochasticCrossover detects crossover between %K and %D lines
func detectStochasticCrossover(kLine, dLine []float64) CrossoverType {
	if len(kLine) < 2 || len(dLine) < 2 {
		return NoCrossover
	}
	
	lastIdx := len(kLine) - 1
	prevIdx := lastIdx - 1
	
	// Current and previous values
	kCurr := kLine[lastIdx]
	kPrev := kLine[prevIdx]
	dCurr := dLine[lastIdx]
	dPrev := dLine[prevIdx]
	
	// Check for crossover
	if kPrev <= dPrev && kCurr > dCurr {
		return CrossUp
	} else if kPrev >= dPrev && kCurr < dCurr {
		return CrossDown
	}
	
	return NoCrossover
}

// classifyCCIZone classifies CCI value into zones
func classifyCCIZone(cciValue float64) CCIZone {
	if cciValue < -100 {
		return CCIOversold
	} else if cciValue > 100 {
		return CCIOverbought
	}
	return CCINormal
}

// classifyStochZone classifies Stochastic %K value into zones
func classifyStochZone(kValue float64) StochZone {
	if kValue < 20 {
		return StochOversold
	} else if kValue > 80 {
		return StochOverbought
	}
	return StochNeutral
}

// classifyMFIZone classifies MFI value into zones
func classifyMFIZone(mfiValue float64) MFIZone {
	if mfiValue < 20 {
		return MFIOversold
	} else if mfiValue > 80 {
		return MFIOverbought
	}
	return MFINeutral
}
