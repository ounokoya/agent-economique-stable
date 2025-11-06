package indicators

import (
    "math"
    "sort"
)

// EMA computes a TradingView-like Exponential Moving Average over src.
// Rules:
// - alpha = 2/(length+1)
// - Warm-up: indices < length-1 => NaN
// - Lazy seed: the first defined EMA occurs at the first index i >= length-1 where
//   SMA(src, length)[i] is defined; EMA[i] = SMA[i].
// - Reseed after gaps: if a gap (NaN/Inf) breaks continuity and EMA[i-1] is NaN,
//   wait until the next index k where SMA[k] is defined again, then set EMA[k] = SMA[k].
// - Strict NaN/Inf propagation: if src[i] is NaN/Inf then EMA[i] = NaN and continuity breaks
//   until a future reseed point.
func EMA(src []float64, length int) []float64 {
    n := len(src)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if length <= 0 || n == 0 || length > n { return out }

    // Precompute TV-like SMA as seed series for lazy (re)seeding
    sma := SMA(src, length)
    alpha := 2.0 / (float64(length) + 1.0)

    // Lazy seed + reseed forward pass starting from length-1
    seeded := false
    for i := length - 1; i < n; i++ {
        if !seeded {
            // Attempt to seed at i using SMA if available
            if !math.IsNaN(sma[i]) && !math.IsInf(sma[i], 0) {
                out[i] = sma[i]
                seeded = true
            } else {
                out[i] = math.NaN()
            }
            continue
        }

        prev := out[i-1]
        v := src[i]
        // If continuity is broken (prev invalid) or current src invalid, emit NaN and drop seed
        if math.IsNaN(prev) || math.IsInf(prev, 0) || math.IsNaN(v) || math.IsInf(v, 0) {
            out[i] = math.NaN()
            seeded = false
            // On next iterations, we'll wait for the next SMA window to reseed
            continue
        }
        out[i] = alpha*v + (1.0-alpha)*prev
    }
    return out
}

// EMAFromKlines computes EMA from candles using a custom source accessor.
// Defensive: sorts a copy of candles by Timestamp ascending before computation.
func EMAFromKlines(klines []Kline, length int, src func(Kline) float64) []float64 {
    if len(klines) > 1 {
        cpy := make([]Kline, len(klines))
        copy(cpy, klines)
        sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
        klines = cpy
    }
    vals := make([]float64, len(klines))
    for i := range klines { vals[i] = src(klines[i]) }
    return EMA(vals, length)
}
