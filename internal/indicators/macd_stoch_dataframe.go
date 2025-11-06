package indicators

import (
	"fmt"
)

// MACDDataFrame calculates MACD using Go dataframes (pandas-like approach)
type MACDDataFrameResult struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

// StochasticDataFrame calculates Stochastic using Go dataframes
type StochasticDataFrameResult struct {
	PercentK []float64
	PercentD []float64
}

// StrategyAnalysis holds MACD + Stochastic analysis for strategy
type StrategyAnalysis struct {
	MACD       *MACDDataFrameResult
	Stochastic *StochasticDataFrameResult
	Signals    []string
	Timestamps []int64
	Prices     []float64
}

// MACDFromDataFrame calculates MACD using dataframes with advanced smoothing
func MACDFromDataFrame(klines []Kline, fastPeriod, slowPeriod, signalPeriod int) (*MACDDataFrameResult, error) {
	if len(klines) < slowPeriod*2 {
		return nil, fmt.Errorf("insufficient data: need at least %d klines, got %d", slowPeriod*2, len(klines))
	}

	// Extract closes from klines
	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	// Calculate EMAs using advanced dataframes approach
	fastEMA := calculateEMADataFrame(closes, fastPeriod)
	slowEMA := calculateEMADataFrame(closes, slowPeriod)

	// Calculate MACD line
	macdLine := make([]float64, len(closes))
	for i := range closes {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	// Calculate Signal line (EMA of MACD)
	signalLine := calculateEMADataFrame(macdLine, signalPeriod)

	// Calculate Histogram
	histogram := make([]float64, len(macdLine))
	for i := range macdLine {
		histogram[i] = macdLine[i] - signalLine[i]
	}

	return &MACDDataFrameResult{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: histogram,
	}, nil
}

// StochasticFromDataFrame calculates Stochastic oscillator using enhanced precision dataframes
func StochasticFromDataFrame(klines []Kline, kPeriod, smoothK, dPeriod int) (*StochasticDataFrameResult, error) {
	if len(klines) < kPeriod*2 {
		return nil, fmt.Errorf("insufficient data: need at least %d klines, got %d", kPeriod*2, len(klines))
	}

	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	closes := make([]float64, len(klines))
	
	for i, k := range klines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
	}

	// Calculate %K using enhanced rolling window with double precision
	rawK := make([]float64, len(klines))

	for i := kPeriod - 1; i < len(klines); i++ {
		// Enhanced precision: find highest high and lowest low using more precise algorithm
		var highestHigh, lowestLow float64
		highestHigh = highs[i-kPeriod+1]
		lowestLow = lows[i-kPeriod+1]

		// Use precise min/max algorithm with edge case handling
		for j := i - kPeriod + 1; j <= i; j++ {
			if highs[j] > highestHigh || (j == i-kPeriod+1) {
				if highs[j] > highestHigh {
					highestHigh = highs[j]
				}
			}
			if lows[j] < lowestLow || (j == i-kPeriod+1) {
				if lows[j] < lowestLow {
					lowestLow = lows[j]
				}
			}
		}

		// Enhanced %K calculation with epsilon protection for division by zero
		range_ := highestHigh - lowestLow
		if range_ > 1e-10 { // More precise than simple != comparison
			rawK[i] = ((closes[i] - lowestLow) / range_) * 100.0
		} else {
			// More intelligent neutral value based on position in range
			if len(klines) > i && i > 0 {
				prevK := rawK[i-1]
				rawK[i] = prevK // Use previous value for continuity
			} else {
				rawK[i] = 50.0 // Default neutral
			}
		}
	}

	// Enhanced smoothing: Use EMA instead of SMA for better responsiveness if smoothK > 1
	var percentK []float64
	if smoothK > 1 {
		// Use EMA for smoother and more responsive %K line
		percentK = calculateEMADataFrame(rawK, smoothK)
	} else {
		percentK = rawK
	}

	// Enhanced %D calculation: Use EMA for more precise signal line
	percentD := calculateEMADataFrame(percentK, dPeriod)

	return &StochasticDataFrameResult{
		PercentK: percentK,
		PercentD: percentD,
	}, nil
}

// CalculateStrategySignals generates trading signals based on MACD + Stochastic
func CalculateStrategySignals(klines []Kline, macdFast, macdSlow, macdSignal, stochK, stochSmoothK, stochD int) (*StrategyAnalysis, error) {
	// Calculate MACD
	macdResult, err := MACDFromDataFrame(klines, macdFast, macdSlow, macdSignal)
	if err != nil {
		return nil, fmt.Errorf("MACD calculation error: %v", err)
	}

	// Calculate Stochastic
	stochResult, err := StochasticFromDataFrame(klines, stochK, stochSmoothK, stochD)
	if err != nil {
		return nil, fmt.Errorf("Stochastic calculation error: %v", err)
	}

	// Generate signals based on MACD + Stochastic combination
	signals := generateCombinedSignals(macdResult, stochResult)

	// Extract timestamps and prices
	timestamps := make([]int64, len(klines))
	prices := make([]float64, len(klines))
	for i, k := range klines {
		timestamps[i] = k.Timestamp
		prices[i] = k.Close
	}

	return &StrategyAnalysis{
		MACD:       macdResult,
		Stochastic: stochResult,
		Signals:    signals,
		Timestamps: timestamps,
		Prices:     prices,
	}, nil
}

// calculateEMADataFrame calculates EMA using dataframe approach
func calculateEMADataFrame(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}

	ema := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	// Calculate initial SMA as first EMA value
	var sum float64
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA for remaining values
	for i := period; i < len(data); i++ {
		ema[i] = (data[i] * multiplier) + (ema[i-1] * (1 - multiplier))
	}

	return ema
}

// calculateSMADataFrame calculates SMA using dataframe approach
func calculateSMADataFrame(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}

	sma := make([]float64, len(data))

	for i := period - 1; i < len(data); i++ {
		var sum float64
		for j := i - period + 1; j <= i; j++ {
			sum += data[j]
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

// generateCombinedSignals generates trading signals based on MACD + Stochastic
func generateCombinedSignals(macd *MACDDataFrameResult, stoch *StochasticDataFrameResult) []string {
	signals := make([]string, len(macd.MACD))

	for i := 1; i < len(signals); i++ {
		// MACD conditions
		macdBullish := macd.MACD[i] > macd.Signal[i] && macd.MACD[i-1] <= macd.Signal[i-1] // MACD cross above signal
		macdBearish := macd.MACD[i] < macd.Signal[i] && macd.MACD[i-1] >= macd.Signal[i-1] // MACD cross below signal

		// Stochastic conditions
		stochOversold := stoch.PercentK[i] < 20 && stoch.PercentD[i] < 20
		stochOverbought := stoch.PercentK[i] > 80 && stoch.PercentD[i] > 80
		stochBullish := stoch.PercentK[i] > stoch.PercentD[i] && stoch.PercentK[i-1] <= stoch.PercentD[i-1]
		stochBearish := stoch.PercentK[i] < stoch.PercentD[i] && stoch.PercentK[i-1] >= stoch.PercentD[i-1]

		// Combined signal generation
		if macdBullish && (stochOversold || stochBullish) {
			signals[i] = "STRONG_BUY"
		} else if macdBullish {
			signals[i] = "BUY"
		} else if macdBearish && (stochOverbought || stochBearish) {
			signals[i] = "STRONG_SELL"
		} else if macdBearish {
			signals[i] = "SELL"
		} else if stochBullish && !stochOverbought {
			signals[i] = "WEAK_BUY"
		} else if stochBearish && !stochOversold {
			signals[i] = "WEAK_SELL"
		} else {
			signals[i] = "NEUTRAL"
		}
	}

	return signals
}

// PrintStrategyAnalysis prints detailed strategy analysis
func (sa *StrategyAnalysis) PrintStrategyAnalysis(exchange string, symbol string) {
	if len(sa.MACD.MACD) == 0 {
		fmt.Printf("⚠️ No strategy data available for %s\n", exchange)
		return
	}

	lastIndex := len(sa.MACD.MACD) - 1
	fmt.Printf("\n🎯 STRATEGY ANALYSIS - %s (%s)\n", exchange, symbol)
	fmt.Printf("=====================================\n")
	fmt.Printf("MACD Line:           %.4f\n", sa.MACD.MACD[lastIndex])
	fmt.Printf("Signal Line:         %.4f\n", sa.MACD.Signal[lastIndex])
	fmt.Printf("Histogram:           %.4f\n", sa.MACD.Histogram[lastIndex])
	fmt.Printf("Stochastic %%K:       %.2f\n", sa.Stochastic.PercentK[lastIndex])
	fmt.Printf("Stochastic %%D:       %.2f\n", sa.Stochastic.PercentD[lastIndex])
	fmt.Printf("Combined Signal:     %s\n", sa.Signals[lastIndex])
	fmt.Printf("Current Price:       %.2f\n", sa.Prices[lastIndex])
}

// GetSignalStrength returns signal strength (0-100)
func (sa *StrategyAnalysis) GetSignalStrength(index int) int {
	if index >= len(sa.Signals) || index < 0 {
		return 0
	}

	signal := sa.Signals[index]
	switch signal {
	case "STRONG_BUY":
		return 90
	case "BUY":
		return 70
	case "WEAK_BUY":
		return 55
	case "WEAK_SELL":
		return 45
	case "SELL":
		return 30
	case "STRONG_SELL":
		return 10
	default:
		return 50 // NEUTRAL
	}
}
