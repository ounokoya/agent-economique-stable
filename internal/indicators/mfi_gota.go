package indicators

import (
	"fmt"
	"math"
	"sort"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// MFIGota computes Money Flow Index using Gota DataFrame (pandas-like approach).
// This demonstrates how to implement technical indicators using a pandas-style
// approach in Go, providing better readability and data manipulation capabilities.
//
// Algorithm (same as MFI/MFIV2 but using DataFrame operations):
// 1. Create DataFrame from OHLCV arrays
// 2. Calculate Typical Price: TP = (H + L + C) / 3
// 3. Calculate Money Flow: MF = TP * Volume
// 4. Calculate TP shift for comparison: TP_prev = TP.shift(1)
// 5. Classify Money Flows based on TP direction
// 6. Calculate rolling sums of PMF and NMF over period
// 7. Apply MFI formula: 100 - (100 / (1 + PMF/NMF))
//
// Benefits of Gota approach:
// - Cleaner, more readable code
// - Built-in data manipulation functions
// - Easy to add new calculated columns
// - Pandas-like syntax familiar to Python developers
//
// Trade-offs:
// - Slightly more memory usage than pure Go arrays
// - Additional dependency
// - Conversion overhead between Go types and DataFrame
func MFIGota(h, l, c, v []float64, length int) []float64 {
	n := len(h)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	
	if length <= 0 || n == 0 || len(l) != n || len(c) != n || len(v) != n || length > n {
		return out
	}

	// Step 1: Create DataFrame from OHLCV data
	// Convert arrays to series for DataFrame creation
	highSeries := series.New(h, series.Float, "high")
	lowSeries := series.New(l, series.Float, "low") 
	closeSeries := series.New(c, series.Float, "close")
	volumeSeries := series.New(v, series.Float, "volume")
	
	df := dataframe.New(highSeries, lowSeries, closeSeries, volumeSeries)
	
	// Step 2: Calculate Typical Price using DataFrame operations
	// TP = (H + L + C) / 3
	tpValues := make([]float64, n)
	for i := 0; i < n; i++ {
		H := df.Elem(i, 0).Float()  // high
		L := df.Elem(i, 1).Float()  // low
		C := df.Elem(i, 2).Float()  // close
		
		// Handle invalid data
		if math.IsNaN(H) || math.IsNaN(L) || math.IsNaN(C) || math.IsInf(H, 0) || math.IsInf(L, 0) || math.IsInf(C, 0) {
			tpValues[i] = math.NaN()
			continue
		}
		
		// Normalize H/L if needed (same as original implementation)
		if H < L {
			H, L = L, H
		}
		
		tpValues[i] = (H + L + C) / 3.0
	}
	
	// Add TP column to DataFrame
	tpSeries := series.New(tpValues, series.Float, "tp")
	df = df.CBind(dataframe.New(tpSeries))
	
	// Step 3: Calculate Money Flow = TP * Volume
	mfValues := make([]float64, n)
	for i := 0; i < n; i++ {
		tp := df.Elem(i, 4).Float()  // tp column
		vol := df.Elem(i, 3).Float() // volume column
		
		if math.IsNaN(tp) || math.IsNaN(vol) || math.IsInf(vol, 0) {
			mfValues[i] = math.NaN()
			continue
		}
		
		// Handle negative volume (clamp to 0)
		if vol < 0 {
			vol = 0
		}
		
		mfValues[i] = tp * vol
	}
	
	// Add MF column to DataFrame
	mfSeries := series.New(mfValues, series.Float, "mf")
	df = df.CBind(dataframe.New(mfSeries))
	
	// Step 4: Create shifted TP for comparison (TP_prev)
	tpPrevValues := make([]float64, n)
	tpPrevValues[0] = math.NaN() // First element has no previous
	for i := 1; i < n; i++ {
		tpPrevValues[i] = tpValues[i-1]
	}
	
	tpPrevSeries := series.New(tpPrevValues, series.Float, "tp_prev")
	df = df.CBind(dataframe.New(tpPrevSeries))
	
	// Step 5: Classify Money Flows (Positive/Negative)
	posMfValues := make([]float64, n)
	negMfValues := make([]float64, n)
	
	for i := 0; i < n; i++ {
		if i == 0 {
			// First bar has no previous TP
			posMfValues[i] = math.NaN()
			negMfValues[i] = math.NaN()
			continue
		}
		
		tp := df.Elem(i, 4).Float()      // tp
		tpPrev := df.Elem(i, 6).Float()  // tp_prev  
		mf := df.Elem(i, 5).Float()      // mf
		
		if math.IsNaN(tp) || math.IsNaN(tpPrev) || math.IsNaN(mf) {
			posMfValues[i] = math.NaN()
			negMfValues[i] = math.NaN()
			continue
		}
		
		if tp > tpPrev {
			// Positive money flow
			posMfValues[i] = mf
			negMfValues[i] = 0.0
		} else if tp < tpPrev {
			// Negative money flow
			posMfValues[i] = 0.0
			negMfValues[i] = mf
		} else {
			// TP unchanged = neutral
			posMfValues[i] = 0.0
			negMfValues[i] = 0.0
		}
	}
	
	// Add PMF and NMF columns to DataFrame
	posMfSeries := series.New(posMfValues, series.Float, "pmf")
	negMfSeries := series.New(negMfValues, series.Float, "nmf")
	df = df.CBind(dataframe.New(posMfSeries)).CBind(dataframe.New(negMfSeries))
	
	// Step 6: Calculate rolling sums and MFI
	// Unfortunately, Gota doesn't have built-in rolling window functions yet,
	// so we'll implement the sliding window manually (same logic as original)
	var sumP, sumN float64
	badP, badN := 0, 0
	
	// Exclude last bar (in-flight bar)
	end := n - 1
	
	for i := 0; i < n; i++ {
		// Get current PMF and NMF values
		pmf := df.Elem(i, 7).Float()  // pmf column
		nmf := df.Elem(i, 8).Float()  // nmf column
		
		// Add current values to window
		if math.IsNaN(pmf) {
			badP++
		} else {
			sumP += pmf
		}
		if math.IsNaN(nmf) {
			badN++
		} else {
			sumN += nmf
		}
		
		// Remove values leaving the window
		if i >= length {
			leave := i - length
			leavePmf := df.Elem(leave, 7).Float()
			leaveNmf := df.Elem(leave, 8).Float()
			
			if math.IsNaN(leavePmf) {
				badP--
			} else {
				sumP -= leavePmf
			}
			if math.IsNaN(leaveNmf) {
				badN--
			} else {
				sumN -= leaveNmf
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
		
		// Check for NaN contamination
		if badP > 0 || badN > 0 {
			out[i] = math.NaN()
			continue
		}
		
		// Calculate MFI using standard formula
		var val float64
		if sumN == 0 && sumP == 0 {
			val = math.NaN()  // Both zero = undefined
		} else if sumN == 0 {
			val = 100.0  // Only positive flow
		} else if sumP == 0 {
			val = 0.0    // Only negative flow
		} else {
			// Standard TA-Lib formula
			moneyRatio := sumP / sumN
			val = 100.0 - (100.0 / (1.0 + moneyRatio))
		}
		
		// Clamp to [0,100]
		if val < 0 {
			val = 0
		} else if val > 100 {
			val = 100
		}
		if val == 0 {
			val = 0  // normalize -0.0
		}
		
		out[i] = val
	}
	
	// Force last bar to NaN (in-flight)
	if n > 0 {
		out[n-1] = math.NaN()
	}
	
	return out
}

// MFIGotaFromKlines computes MFI using Gota DataFrame from Klines
func MFIGotaFromKlines(klines []Kline, length int) []float64 {
	if len(klines) > 1 {
		cpy := make([]Kline, len(klines))
		copy(cpy, klines)
		// Sort by timestamp for consistency
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
	
	return MFIGota(h, l, c, v, length)
}

// PrintMFIDataFrame utility function to inspect DataFrame during debugging
func PrintMFIDataFrame(klines []Kline, length int) {
	if len(klines) < 5 {
		return
	}
	
	// Take only first 5 rows for debugging
	sample := klines[:5]
	
	n := len(sample)
	h := make([]float64, n)
	l := make([]float64, n)
	c := make([]float64, n)
	v := make([]float64, n)
	
	for i := range sample {
		h[i] = sample[i].High
		l[i] = sample[i].Low
		c[i] = sample[i].Close
		v[i] = sample[i].Volume
	}
	
	// Create DataFrame for inspection
	highSeries := series.New(h, series.Float, "high")
	lowSeries := series.New(l, series.Float, "low")
	closeSeries := series.New(c, series.Float, "close")
	volumeSeries := series.New(v, series.Float, "volume")
	
	df := dataframe.New(highSeries, lowSeries, closeSeries, volumeSeries)
	
	fmt.Println("🔍 MFI DataFrame Debug (first 5 rows):")
	fmt.Println(df)
}
