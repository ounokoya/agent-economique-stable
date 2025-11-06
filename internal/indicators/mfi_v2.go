package indicators

import (
	"math"
	"sort"
)

// MFIV2 computes Money Flow Index using the standard TA-Lib/TradingView formula.
// This version corrects the calculation to match industry standards:
// MFI = 100 - (100 / (1 + Money_Ratio)) where Money_Ratio = PMF / NMF
//
// Differences from original MFI():
// - Uses standard formula: 100 - (100/(1 + PMF/NMF)) instead of 100 * (PMF/(PMF+NMF))
// - Handles PMF=NMF=0 case as undefined (NaN) instead of 50
// - Maintains same robustness for edge cases and NaN handling
//
// Algorithm:
// 1. Typical Price: TP = (H + L + C) / 3
// 2. Money Flow: MF = TP * Volume  
// 3. Classification vs previous TP:
//    - TP[i] > TP[i-1] → Positive Money Flow
//    - TP[i] < TP[i-1] → Negative Money Flow
//    - TP[i] = TP[i-1] → Neutral (ignored)
// 4. Sum PMF and NMF over sliding window of `length` periods
// 5. Money Ratio = PMF / NMF
// 6. MFI = 100 - (100 / (1 + Money_Ratio))
//
// Edge cases:
// - PMF > 0, NMF = 0 → MFI = 100
// - PMF = 0, NMF > 0 → MFI = 0  
// - PMF = 0, NMF = 0 → MFI = NaN (undefined)
func MFIV2(h, l, c, v []float64, length int) []float64 {
	n := len(h)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	if length <= 0 || n == 0 || len(l) != n || len(c) != n || len(v) != n || length > n {
		return out
	}

	isBad := func(x float64) bool { return math.IsNaN(x) || math.IsInf(x, 0) }

	// Step 1: Calculate Typical Price and Money Flow
	tp := make([]float64, n)
	mf := make([]float64, n)
	for i := 0; i < n; i++ {
		H, L, C, V := h[i], l[i], c[i], v[i]
		// Invalidate bar if non-finite inputs
		if isBad(H) || isBad(L) || isBad(C) || isBad(V) {
			tp[i] = math.NaN()
			mf[i] = math.NaN()
			continue
		}
		// Normalize H/L instead of invalidating when H < L (TV tolerance)
		if H < L {
			H, L = L, H
		}
		t := (H + L + C) / 3.0
		tp[i] = t
		// Volume rules: V==0 -> MF=0 valid; V<0 -> clamp to 0 (MF=0)
		if V < 0 {
			V = 0
		}
		mf[i] = t * V
	}

	// Step 2: Classify Money Flows (positive/negative)
	pos := make([]float64, n)
	neg := make([]float64, n)
	for i := range pos {
		pos[i], neg[i] = math.NaN(), math.NaN()
	}

	for i := 0; i < n; i++ {
		if i == 0 {
			// No previous TP available for first bar
			pos[i], neg[i] = math.NaN(), math.NaN()
			continue
		}
		t, tPrev := tp[i], tp[i-1]
		f := mf[i]
		if isBad(t) || isBad(tPrev) || isBad(f) {
			pos[i], neg[i] = math.NaN(), math.NaN()
			continue
		}
		if t > tPrev {
			pos[i], neg[i] = f, 0
		} else if t < tPrev {
			pos[i], neg[i] = 0, f
		} else {
			// TP equal → neutral (no money flow)
			pos[i], neg[i] = 0, 0
		}
	}

	// Step 3: Sliding window sums with NaN propagation
	var sumP, sumN float64
	badP, badN := 0, 0

	// Exclude last bar: end is n-1; we won't emit output for index n-1
	end := n - 1
	for i := 0; i < n; i++ {
		// Add current values to window
		if isBad(pos[i]) {
			badP++
		} else {
			sumP += pos[i]
		}
		if isBad(neg[i]) {
			badN++
		} else {
			sumN += neg[i]
		}

		// Remove values leaving the window
		if i >= length {
			leave := i - length
			if isBad(pos[leave]) {
				badP--
			} else {
				sumP -= pos[leave]
			}
			if isBad(neg[leave]) {
				badN--
			} else {
				sumN -= neg[leave]
			}
		}

		// Wait for warm-up period
		if i < length {
			continue
		}

		// Don't compute in-flight last bar
		if i >= end {
			continue
		}

		// Check for NaN contamination in window
		if badP > 0 || badN > 0 {
			out[i] = math.NaN()
			continue
		}

		// Step 4: Calculate MFI using standard formula
		var val float64
		if sumN == 0 && sumP == 0 {
			// Both PMF and NMF are zero → undefined case
			val = math.NaN()
		} else if sumN == 0 {
			// Only positive money flow → MFI = 100
			val = 100.0
		} else if sumP == 0 {
			// Only negative money flow → MFI = 0
			val = 0.0
		} else {
			// Standard TA-Lib formula: MFI = 100 - (100 / (1 + PMF/NMF))
			moneyRatio := sumP / sumN
			val = 100.0 - (100.0 / (1.0 + moneyRatio))
		}

		// Clamp to [0,100] and normalize signed zeros
		if val < 0 {
			val = 0
		} else if val > 100 {
			val = 100
		}
		if val == 0 { // normalize -0.0 -> 0.0
			val = 0
		}
		out[i] = val
	}

	// Force last bar to NaN explicitly (in-flight bar)
	if n > 0 {
		out[n-1] = math.NaN()
	}

	return out
}

// MFIV2FromKlines computes MFI v2 from candles using standard TA-Lib formula
func MFIV2FromKlines(klines []Kline, length int) []float64 {
	if len(klines) > 1 {
		cpy := make([]Kline, len(klines))
		copy(cpy, klines)
		// Stable ascending sort by timestamp for deterministic ordering
		sort.SliceStable(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
		klines = cpy
	}
	n := len(klines)
	h := make([]float64, n)
	l := make([]float64, n)
	c := make([]float64, n)
	v := make([]float64, n)
	for i := range klines {
		h[i] = klines[i].High
		l[i] = klines[i].Low
		c[i] = klines[i].Close
		v[i] = klines[i].Volume
	}
	return MFIV2(h, l, c, v, length)
}
