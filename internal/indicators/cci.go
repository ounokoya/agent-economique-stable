package indicators

import (
	"math"
	"sort"
)

// CCIFromSeries computes a TradingView-like CCI using SMA for both basis and MAD.
// src: price source series (e.g., HLC3 or close)
// length: window size
// Returns a vector of length N using ALL available data (no exclusion of last candle).
// Warm-up and strict NaN/Inf propagation are preserved.
func CCIFromSeries(src []float64, length int) []float64 {
	n := len(src)
	cci := make([]float64, n)
	for i := range cci { cci[i] = math.NaN() }
	if n == 0 || length <= 0 { return cci }
	// Use all available data, no exclusion
	if length > n { return cci }
	// Compute basis on full series
	basis := SMA(src, length)
	for i := 0; i < n; i++ {
		if i < length-1 { cci[i] = math.NaN(); continue }
		b := basis[i]
		v := src[i]
		if math.IsNaN(b) || math.IsInf(b, 0) || math.IsNaN(v) || math.IsInf(v, 0) { cci[i] = math.NaN(); continue }
		// MAD around basis[i] over inclusive window [i-length+1..i]
		start := i - length + 1
		if start < 0 { start = 0 }
		sumAbs := 0.0
		bad := false
		for j := start; j <= i; j++ {
			vj := src[j]
			if math.IsNaN(vj) || math.IsInf(vj, 0) { bad = true; break }
			sumAbs += math.Abs(vj - b)
		}
		if bad { cci[i] = math.NaN(); continue }
		mad := sumAbs / float64(length)
		if mad == 0 || math.IsNaN(mad) || math.IsInf(mad, 0) {
			cci[i] = math.NaN()
			continue
		}
		cci[i] = (v - b) / (0.015 * mad)
	}
	return cci
}

// CCIFromKlines builds source from candles based on mode ("hlc3"|"close") and computes CCI with SMA.
func CCIFromKlines(klines []Kline, mode string, length int) []float64 {
	// Ensure chronological order
	klines2 := make([]Kline, len(klines))
	copy(klines2, klines)
	sort.SliceStable(klines2, func(i, j int) bool { return klines2[i].Timestamp < klines2[j].Timestamp })

	n := len(klines2)
	src := make([]float64, n)
	useClose := (mode == "close")
	for i := range klines2 {
		if useClose { src[i] = klines2[i].Close } else { src[i] = (klines2[i].High + klines2[i].Low + klines2[i].Close) / 3.0 }
	}
	return CCIFromSeries(src, length)
}
