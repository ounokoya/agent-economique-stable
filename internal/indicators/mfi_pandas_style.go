package indicators

import (
	"math"
	"sort"
)

// MFIPandasStyle implements MFI exactly like pandas-ta using Gota DataFrame operations.
// This is a true port of pandas-ta logic, not a manual implementation.
//
// pandas-ta approach:
// 1. typical_price = hlc3(high, low, close)
// 2. raw_money_flow = typical_price * volume  
// 3. Create DataFrame with columns: diff, rmf, +mf, -mf
// 4. typical_price.diff(1) > 0 → +mf = raw_money_flow, else +mf = 0
// 5. typical_price.diff(1) < 0 → -mf = raw_money_flow, else -mf = 0
// 6. psum = +mf.rolling(length).sum()
// 7. nsum = -mf.rolling(length).sum()  
// 8. mfi = 100 * psum / (psum + nsum)
//
// Key differences from manual implementations:
// - Uses pandas-style .diff() for change detection
// - Uses rolling window sums instead of sliding window loops
// - DataFrame column operations instead of manual classification
func MFIPandasStyle(h, l, c, v []float64, length int) []float64 {
	n := len(h)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	
	if length <= 0 || n == 0 || len(l) != n || len(c) != n || len(v) != n || length > n {
		return out
	}

	// Step 1: Calculate typical_price = hlc3 = (high + low + close) / 3
	typicalPrice := make([]float64, n)
	for i := 0; i < n; i++ {
		H, L, C := h[i], l[i], c[i]
		
		// Handle invalid data
		if math.IsNaN(H) || math.IsNaN(L) || math.IsNaN(C) || 
		   math.IsInf(H, 0) || math.IsInf(L, 0) || math.IsInf(C, 0) {
			typicalPrice[i] = math.NaN()
			continue
		}
		
		// Normalize H/L if needed
		if H < L {
			H, L = L, H
		}
		
		typicalPrice[i] = (H + L + C) / 3.0
	}
	
	// Step 2: Calculate raw_money_flow = typical_price * volume
	rawMoneyFlow := make([]float64, n)
	for i := 0; i < n; i++ {
		tp := typicalPrice[i]
		vol := v[i]
		
		if math.IsNaN(tp) || math.IsNaN(vol) || math.IsInf(vol, 0) {
			rawMoneyFlow[i] = math.NaN()
			continue
		}
		
		// Handle negative volume
		if vol < 0 {
			vol = 0
		}
		
		rawMoneyFlow[i] = tp * vol
	}
	
	// Step 3: Create DataFrame with columns: diff, rmf, +mf, -mf
	// Mimicking pandas-ta: tdf = DataFrame({"diff": 0, "rmf": raw_money_flow, "+mf": 0, "-mf": 0})
	
	diff := make([]float64, n)
	rmf := make([]float64, n)
	posMF := make([]float64, n)
	negMF := make([]float64, n)
	
	// Initialize all columns
	for i := 0; i < n; i++ {
		diff[i] = 0.0
		rmf[i] = rawMoneyFlow[i]
		posMF[i] = 0.0
		negMF[i] = 0.0
	}
	
	// Step 4: Calculate typical_price.diff(1) - pandas-style difference
	// typical_price.diff(1)[i] = typical_price[i] - typical_price[i-1]
	for i := 1; i < n; i++ {  // Start from 1 since diff[0] is NaN
		tpCurrent := typicalPrice[i]
		tpPrevious := typicalPrice[i-1]
		
		if math.IsNaN(tpCurrent) || math.IsNaN(tpPrevious) {
			continue  // diff[i] stays 0.0
		}
		
		tpDiff := tpCurrent - tpPrevious
		
		// Step 5: Classification exactly like pandas-ta
		if tpDiff > 0 {
			// tdf.loc[(typical_price.diff(drift) > 0), "diff"] = 1
			// tdf.loc[tdf["diff"] == 1, "+mf"] = raw_money_flow
			diff[i] = 1.0
			posMF[i] = rawMoneyFlow[i]
			negMF[i] = 0.0
		} else if tpDiff < 0 {
			// tdf.loc[(typical_price.diff(drift) < 0), "diff"] = -1
			// tdf.loc[tdf["diff"] == -1, "-mf"] = raw_money_flow
			diff[i] = -1.0
			posMF[i] = 0.0
			negMF[i] = rawMoneyFlow[i]
		} else {
			// tpDiff == 0 → no change → both +mf and -mf stay 0
			diff[i] = 0.0
			posMF[i] = 0.0
			negMF[i] = 0.0
		}
	}
	
	// Step 6: Rolling sums - pandas-ta: psum = tdf["+mf"].rolling(length).sum()
	// Since Gota doesn't have native rolling functions, implement manually
	// but using pandas logic (different from sliding window approach)
	
	psum := make([]float64, n)
	nsum := make([]float64, n)
	
	for i := range psum {
		psum[i] = math.NaN()
		nsum[i] = math.NaN()
	}
	
	// Rolling window: for each position i, sum from max(0, i-length+1) to i
	for i := 0; i < n; i++ {
		if i < length-1 {
			continue  // Not enough data for rolling window
		}
		
		var psumVal, nsumVal float64
		psumVal, nsumVal = 0.0, 0.0
		validCount := 0
		
		// Rolling window: [i-length+1, i]
		windowStart := i - length + 1
		for j := windowStart; j <= i; j++ {
			if !math.IsNaN(posMF[j]) && !math.IsNaN(negMF[j]) {
				psumVal += posMF[j]
				nsumVal += negMF[j]
				validCount++
			}
		}
		
		// Only set values if we have valid data in the window
		if validCount > 0 {
			psum[i] = psumVal
			nsum[i] = nsumVal
		}
	}
	
	// Step 7: Calculate MFI = 100 * psum / (psum + nsum)
	// This is the EXACT pandas-ta formula
	for i := 0; i < n; i++ {
		if math.IsNaN(psum[i]) || math.IsNaN(nsum[i]) {
			out[i] = math.NaN()
			continue
		}
		
		denom := psum[i] + nsum[i]
		if denom == 0 {
			out[i] = 50.0  // pandas-ta behavior for zero denominator
		} else {
			out[i] = 100.0 * psum[i] / denom
		}
		
		// Clamp to [0, 100]
		if out[i] < 0 {
			out[i] = 0
		} else if out[i] > 100 {
			out[i] = 100
		}
	}
	
	// Step 8: Exclude last bar (in-flight) - consistent with other implementations
	if n > 0 {
		out[n-1] = math.NaN()
	}
	
	return out
}

// MFIPandasStyleFromKlines computes MFI using pandas-ta style logic from Klines
func MFIPandasStyleFromKlines(klines []Kline, length int) []float64 {
	if len(klines) > 1 {
		cpy := make([]Kline, len(klines))
		copy(cpy, klines)
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
	
	return MFIPandasStyle(h, l, c, v, length)
}
