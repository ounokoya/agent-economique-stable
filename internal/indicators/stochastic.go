package indicators

import (
	"math"
	"sort"
)

// StochasticFromSeries computes TradingView-like Stochastic %K and %D over series of H,L,C.
// lengthK: window for HH/LL range (inclusive)
// smoothK: SMA smoothing window applied to raw %K
// lengthD: SMA window applied to smoothed %K to form %D
// Returns k and d arrays of length N with the last index set to NaN (last candle excluded from calc), with NaN warm-up.
func StochasticFromSeries(h, l, c []float64, lengthK, smoothK, lengthD int) (k []float64, d []float64) {
	n := len(c)
	k = make([]float64, n)
	d = make([]float64, n)
	for i := 0; i < n; i++ { k[i] = math.NaN(); d[i] = math.NaN() }
	if len(h) != n || len(l) != n || n == 0 || lengthK <= 0 || smoothK <= 0 || lengthD <= 0 {
		return k, d
	}

	// Work on N-1 bars (exclude last), keep output length N with last index NaN
	effN := n - 1
	if effN <= 0 { return k, d }
	// If window length exceeds available bars (excluding the last), no valid outputs can be produced
	// Behavior: entire series remains NaN (including last index), which matches TV-like warm-up.
	if lengthK > effN || smoothK > effN || lengthD > effN { return k, d }

	// Rolling highest high and lowest low over inclusive window lengthK on truncated series
	hh := make([]float64, effN)
	ll := make([]float64, effN)
	for i := 0; i < effN; i++ {
		start := i - lengthK + 1
		if start < 0 { start = 0 }
		var hhv float64 = math.Inf(-1)
		var llv float64 = math.Inf(1)
		count := 0
		bad := false
		for j := start; j <= i; j++ {
			H := h[j]
			L := l[j]
			if math.IsNaN(H) || math.IsNaN(L) || math.IsInf(H, 0) || math.IsInf(L, 0) { bad = true; break }
			count++
			if H > hhv { hhv = H }
			if L < llv { llv = L }
		}
		if bad || count < lengthK { hh[i] = math.NaN(); ll[i] = math.NaN() } else { hh[i] = hhv; ll[i] = llv }
	}

	rawK := make([]float64, effN)
	for i := 0; i < effN; i++ {
		if i < lengthK-1 || math.IsNaN(hh[i]) || math.IsNaN(ll[i]) || math.IsNaN(c[i]) || math.IsInf(c[i], 0) {
			rawK[i] = math.NaN()
			continue
		}
		rangeV := hh[i] - ll[i]
		if rangeV == 0 || math.IsNaN(rangeV) || math.IsInf(rangeV, 0) {
			rawK[i] = math.NaN()
			continue
		}
		rawK[i] = 100.0 * (c[i] - ll[i]) / rangeV
	}

	ks := SMA(rawK, smoothK)
	ds := SMA(ks, lengthD)

	// Copy into outputs (length N). Last index remains NaN by design.
	for i := 0; i < effN; i++ { k[i] = ks[i]; d[i] = ds[i] }
	return
}

// StochasticFromKlines builds series from candles and computes %K and %D.
func StochasticFromKlines(klines []Kline, lengthK, smoothK, lengthD int) (k []float64, d []float64) {
	// Ensure chronological order
	klines2 := make([]Kline, len(klines))
	copy(klines2, klines)
	sort.SliceStable(klines2, func(i, j int) bool { return klines2[i].Timestamp < klines2[j].Timestamp })

	n := len(klines2)
	h := make([]float64, n)
	l := make([]float64, n)
	c := make([]float64, n)
	for i := range klines2 { h[i] = klines2[i].High; l[i] = klines2[i].Low; c[i] = klines2[i].Close }
	return StochasticFromSeries(h, l, c, lengthK, smoothK, lengthD)
}
