package indicators

import "math"

// RMA computes Wilder's RMA (alpha = 1/length) in a TradingView-like way.
// - Seed: first defined value at index (length-1) equals SMA(src, length)[length-1]
// - Warm-up: indices < length-1 => NaN
// - Recurrence (Wilder): out[i] = out[i-1] + (src[i] - out[i-1]) / length
//   which is equivalent to (out[i-1]*(length-1) + src[i]) / length
// - length == 1: return src as-is (rma(x,1) == x in Pine/TV)
// - NaN/Inf propagation: if src[i] is invalid, out[i] = NaN; if out[i-1] is invalid, we re-seed
//   at the first next index where SMA is defined (lazy seed/reseed), matching TV behavior.
func RMA(src []float64, length int) []float64 {
    n := len(src)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if length <= 0 || n == 0 || length > n { return out }
    if length == 1 {
        // Return the input series directly (propagate NaN/Inf as-is)
        for i := 0; i < n; i++ { out[i] = src[i] }
        return out
    }

    sma := SMA(src, length)
    for i := 0; i < n; i++ {
        x := src[i]
        // If input is invalid on this bar, output is NaN
        if math.IsNaN(x) || math.IsInf(x, 0) {
            out[i] = math.NaN()
            continue
        }
        // First bar cannot compute (needs prior)
        if i == 0 {
            out[i] = math.NaN()
            continue
        }
        prev := out[i-1]
        if math.IsNaN(prev) || math.IsInf(prev, 0) {
            // Seed on this bar using SMA at i (lazy seed/reseed)
            out[i] = sma[i]
            continue
        }
        out[i] = (prev*float64(length-1) + x) / float64(length)
    }
    return out
}
