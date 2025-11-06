package indicators

import (
	"math"
)

// MFIEnhanced implements multiple enhanced MFI variants for improved precision
type MFIEnhanced struct {
	Standard   []float64 // MFI standard
	Smoothed   []float64 // MFI avec lissage exponentiel
	Stochastic []float64 // Stochastic MFI
	Enhanced   []float64 // MFI Enhanced combiné
}

// MFIEnhancedFromKlines calculates all enhanced MFI variants
func MFIEnhancedFromKlines(klines []Kline, period int) *MFIEnhanced {
	n := len(klines)
	if n < period+1 {
		return &MFIEnhanced{
			Standard:   make([]float64, n),
			Smoothed:   make([]float64, n),
			Stochastic: make([]float64, n),
			Enhanced:   make([]float64, n),
		}
	}

	// Extract data
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	volume := make([]float64, n)

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volume[i] = k.Volume
	}

	// 1. Calculate Standard MFI
	standardMFI := calculateStandardMFI(high, low, close, volume, period)

	// 2. Calculate Smoothed MFI (avec EMA)
	smoothedMFI := calculateSmoothedMFI(high, low, close, volume, period, 0.2) // Alpha = 0.2

	// 3. Calculate Stochastic MFI
	stochasticMFI := calculateStochasticMFI(standardMFI, period)

	// 4. Calculate Enhanced MFI (volume weighted + adaptive)
	enhancedMFI := calculateEnhancedMFI(high, low, close, volume, period)

	return &MFIEnhanced{
		Standard:   standardMFI,
		Smoothed:   smoothedMFI,
		Stochastic: stochasticMFI,
		Enhanced:   enhancedMFI,
	}
}

// calculateStandardMFI implements the classic MFI calculation
func calculateStandardMFI(high, low, close, volume []float64, period int) []float64 {
	n := len(high)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}

	if period <= 0 || n < period+1 {
		return out
	}

	// Calculate typical prices and raw money flow
	typicalPrice := make([]float64, n)
	rawMoneyFlow := make([]float64, n)
	
	for i := 0; i < n; i++ {
		typicalPrice[i] = (high[i] + low[i] + close[i]) / 3.0
		rawMoneyFlow[i] = typicalPrice[i] * volume[i]
	}

	// Calculate MFI
	for i := period; i < n; i++ {
		posFlow := 0.0
		negFlow := 0.0

		for j := i - period + 1; j <= i; j++ {
			if j > 0 {
				if typicalPrice[j] > typicalPrice[j-1] {
					posFlow += rawMoneyFlow[j]
				} else if typicalPrice[j] < typicalPrice[j-1] {
					negFlow += rawMoneyFlow[j]
				}
			}
		}

		if negFlow != 0 {
			ratio := posFlow / negFlow
			out[i] = 100.0 - (100.0 / (1.0 + ratio))
		} else {
			out[i] = 100.0
		}
	}

	return out
}

// calculateSmoothedMFI implements MFI with exponential smoothing
func calculateSmoothedMFI(high, low, close, volume []float64, period int, alpha float64) []float64 {
	n := len(high)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}

	if period <= 0 || n < period+1 {
		return out
	}

	// Calculate typical prices and raw money flow
	typicalPrice := make([]float64, n)
	rawMoneyFlow := make([]float64, n)
	
	for i := 0; i < n; i++ {
		typicalPrice[i] = (high[i] + low[i] + close[i]) / 3.0
		rawMoneyFlow[i] = typicalPrice[i] * volume[i]
	}

	// Initialize EMA variables
	posFlowEMA := 0.0
	negFlowEMA := 0.0
	initialized := false

	for i := 1; i < n; i++ {
		posFlow := 0.0
		negFlow := 0.0

		if typicalPrice[i] > typicalPrice[i-1] {
			posFlow = rawMoneyFlow[i]
		} else if typicalPrice[i] < typicalPrice[i-1] {
			negFlow = rawMoneyFlow[i]
		}

		if !initialized && i >= period {
			// Initialize with SMA for first value
			posSum := 0.0
			negSum := 0.0
			for j := i - period + 1; j <= i; j++ {
				if j > 0 {
					if typicalPrice[j] > typicalPrice[j-1] {
						posSum += rawMoneyFlow[j]
					} else if typicalPrice[j] < typicalPrice[j-1] {
						negSum += rawMoneyFlow[j]
					}
				}
			}
			posFlowEMA = posSum / float64(period)
			negFlowEMA = negSum / float64(period)
			initialized = true
		} else if initialized {
			// Apply exponential smoothing
			posFlowEMA = alpha*posFlow + (1-alpha)*posFlowEMA
			negFlowEMA = alpha*negFlow + (1-alpha)*negFlowEMA
		}

		if initialized && negFlowEMA != 0 {
			ratio := posFlowEMA / negFlowEMA
			out[i] = 100.0 - (100.0 / (1.0 + ratio))
		} else if initialized {
			out[i] = 100.0
		}
	}

	return out
}

// calculateStochasticMFI implements Stochastic MFI for enhanced sensitivity
func calculateStochasticMFI(mfi []float64, period int) []float64 {
	n := len(mfi)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}

	if period <= 0 || n < period {
		return out
	}

	for i := period - 1; i < n; i++ {
		if math.IsNaN(mfi[i]) {
			continue
		}

		// Find min and max MFI over period
		minMFI := math.Inf(1)
		maxMFI := math.Inf(-1)

		for j := i - period + 1; j <= i; j++ {
			if !math.IsNaN(mfi[j]) {
				if mfi[j] < minMFI {
					minMFI = mfi[j]
				}
				if mfi[j] > maxMFI {
					maxMFI = mfi[j]
				}
			}
		}

		// Calculate Stochastic MFI
		if maxMFI != minMFI {
			out[i] = (mfi[i] - minMFI) / (maxMFI - minMFI) * 100.0
		} else {
			out[i] = 50.0
		}
	}

	return out
}

// calculateEnhancedMFI implements volume-weighted MFI with adaptive features
func calculateEnhancedMFI(high, low, close, volume []float64, period int) []float64 {
	n := len(high)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}

	if period <= 0 || n < period+1 {
		return out
	}

	// Calculate average volume for weighting
	avgVolume := make([]float64, n)
	for i := period - 1; i < n; i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += volume[j]
		}
		avgVolume[i] = sum / float64(period)
	}

	// Calculate typical prices and enhanced raw money flow
	typicalPrice := make([]float64, n)
	for i := 0; i < n; i++ {
		typicalPrice[i] = (high[i] + low[i] + close[i]) / 3.0
	}

	// Enhanced MFI calculation with volume weighting
	for i := period; i < n; i++ {
		posFlow := 0.0
		negFlow := 0.0

		for j := i - period + 1; j <= i; j++ {
			if j > 0 && avgVolume[i] > 0 {
				// Volume weight: higher for above-average volume
				volumeWeight := math.Min(volume[j]/avgVolume[i], 3.0) // Cap at 3x weight
				
				enhancedFlow := typicalPrice[j] * volume[j] * volumeWeight
				
				if typicalPrice[j] > typicalPrice[j-1] {
					posFlow += enhancedFlow
				} else if typicalPrice[j] < typicalPrice[j-1] {
					negFlow += enhancedFlow
				}
			}
		}

		if negFlow != 0 {
			ratio := posFlow / negFlow
			out[i] = 100.0 - (100.0 / (1.0 + ratio))
		} else {
			out[i] = 100.0
		}
	}

	return out
}

// MFISmoothedFromKlines calculates smoothed MFI only (convenience function)
func MFISmoothedFromKlines(klines []Kline, period int) []float64 {
	enhanced := MFIEnhancedFromKlines(klines, period)
	return enhanced.Smoothed
}

// MFIStochasticFromKlines calculates Stochastic MFI only (convenience function)  
func MFIStochasticFromKlines(klines []Kline, period int) []float64 {
	enhanced := MFIEnhancedFromKlines(klines, period)
	return enhanced.Stochastic
}

// MFIEnhancedOnlyFromKlines calculates Enhanced MFI only (convenience function)
func MFIEnhancedOnlyFromKlines(klines []Kline, period int) []float64 {
	enhanced := MFIEnhancedFromKlines(klines, period)
	return enhanced.Enhanced
}
