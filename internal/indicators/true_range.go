package indicators

import (
    "math"
    "sort"
)

// TrueRangeFromKlines computes the True Range series in a TradingView-like manner.
// TR[i] = max(H-L, abs(H-prevClose), abs(L-prevClose)) for i>=1; TR[0]=NaN.
// Defensive: sorts a copy of candles by timestamp ascending before computation.
func TrueRangeFromKlines(klines []Kline) []float64 {
    n := len(klines)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if n == 0 { return out }

    if n > 1 {
        cpy := make([]Kline, n)
        copy(cpy, klines)
        sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
        klines = cpy
    }

    out[0] = math.NaN()
    for i := 1; i < n; i++ {
        h := klines[i].High
        l := klines[i].Low
        pc := klines[i-1].Close
        if math.IsNaN(h) || math.IsNaN(l) || math.IsNaN(pc) || math.IsInf(h, 0) || math.IsInf(l, 0) || math.IsInf(pc, 0) {
            out[i] = math.NaN()
            continue
        }
        r1 := h - l
        if r1 < 0 { r1 = 0 } // TV does not NaN; ensure non-negative
        r2 := math.Abs(h - pc)
        r3 := math.Abs(l - pc)
        out[i] = math.Max(r1, math.Max(r2, r3))
    }
    return out
}
