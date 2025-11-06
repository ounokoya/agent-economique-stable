package indicators

import (
	"math"
	"sort"
)

// KVOFromKlines computes Klinger Volume Oscillator (KVO) and its signal line in a
// TradingView-like manner from candles. It returns three series of length N (same as input):
// kvo, signal, hist. The last index n-1 is set to NaN to exclude the in-flight bar, and
// strict NaN/Inf propagation is applied.
//
// TV-aligned spec implemented (popular TradingView variant):
// - Signed Volume sv based on change(hlc3):
//     * typical (hlc3) = (H+L+C)/3
//     * chg = typical[i] - typical[i-1]
//     * sv[i] = volume[i] if chg >= 0; else sv[i] = -volume[i]
//       For i=0, set sv[0] = 0 (or NaN); using 0 starts EMA later without breaking continuity.
// - KVO = EMA_fast(sv) - EMA_slow(sv) with defaults fast=34, slow=55
// - Signal = EMA_trig(KVO) with trig=13
// - EMA behavior: EMA with lazy seed and reseed after gaps; strict NaN propagation.
//   Note: we exclude the last bar from computation; EMA tolerates trailing NaNs (no resume
//   attempted at the end) and we explicitly force KVO/Signal/Hist at index n-1 to NaN as well
//   to follow TV-like conventions.
//
// Sanity expectations (assuming valid continuous data):
// - First non-NaN index of KVO is slow-1
// - First non-NaN index of Signal is slow-1 + (trig-1)
// - Hist = KVO - Signal at the same indices
// - Warm-ups expected (assuming valid continuous data):
//     * EMA(sv, slow) becomes defined around index slow-1
//     * KVO defined from index slow-1
//     * Signal defined from slow-1 + (trig-1)
//
func KVOFromKlines(klines []Kline, fast, slow, trig int) (kvo, signal, hist []float64) {
	if len(klines) > 1 {
		cpy := make([]Kline, len(klines))
		copy(cpy, klines)
		sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
		klines = cpy
	}
	
	n := len(klines)
	kvo = make([]float64, n)
	signal = make([]float64, n)
	hist = make([]float64, n)
	for i := 0; i < n; i++ {
		kvo[i] = math.NaN()
		signal[i] = math.NaN()
		hist[i] = math.NaN()
	}
	if n == 0 {
		return
	}
	// Exclude last bar convention: we'll compute up to n-2 and leave index n-1 as NaN
	end := n - 1
	if end < 0 {
		end = 0
	}

	// Build signed volume series sv from change(hlc3)
	sv := make([]float64, n)
	for i := 0; i < n; i++ { sv[i] = math.NaN() }
	typ := make([]float64, n)
	for i := 0; i < n; i++ { typ[i] = math.NaN() }

	// Compute typical price and signed volume up to n-2 (exclude the last bar)
	for i := 0; i < end; i++ {
		h := klines[i].High
		l := klines[i].Low
		c := klines[i].Close
		v := klines[i].Volume
		if math.IsNaN(h) || math.IsNaN(l) || math.IsNaN(c) || math.IsNaN(v) ||
			math.IsInf(h, 0) || math.IsInf(l, 0) || math.IsInf(c, 0) || math.IsInf(v, 0) {
			// strict propagation
			typ[i] = math.NaN()
			sv[i] = math.NaN()
			continue
		}
		typ[i] = (h + l + c) / 3.0
		if i == 0 {
			// First bar: define sv[0] = 0 (could be NaN as well); EMA will warm up later
			sv[i] = 0.0
			continue
		}
		if math.IsNaN(typ[i-1]) || math.IsInf(typ[i-1], 0) {
			sv[i] = math.NaN()
			continue
		}
		chg := typ[i] - typ[i-1]
		if chg >= 0 {
			sv[i] = v
		} else {
			sv[i] = -v
		}
	}

	// KVO and signal via TV-like EMA on signed volume
	fastE := EMA(sv, fast)
	slowE := EMA(sv, slow)
	for i := 0; i < n; i++ { // compute diff; last index will be overwritten to NaN below
		if math.IsNaN(fastE[i]) || math.IsNaN(slowE[i]) || math.IsInf(fastE[i], 0) || math.IsInf(slowE[i], 0) {
			kvo[i] = math.NaN()
		} else {
			kvo[i] = fastE[i] - slowE[i]
		}
	}
	sig := EMA(kvo, trig)
	for i := 0; i < n; i++ {
		signal[i] = sig[i]
		if math.IsNaN(kvo[i]) || math.IsNaN(signal[i]) || math.IsInf(kvo[i], 0) || math.IsInf(signal[i], 0) {
			hist[i] = math.NaN()
		} else {
			hist[i] = kvo[i] - signal[i]
		}
	}

	// Exclude last bar by forcing NaN at n-1
	kvo[n-1] = math.NaN()
	signal[n-1] = math.NaN()
	hist[n-1] = math.NaN()
	return
}
