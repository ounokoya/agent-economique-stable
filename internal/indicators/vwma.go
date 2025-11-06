package indicators

import (
    "math"
    "sort"
)

// VWMA computes a TradingView-like Volume-Weighted Moving Average using our SMA.
// Rules inherited from SMA:
// - Warm-up: indices < length-1 => NaN
// - NaN/Inf propagation: if any NaN/Inf present in the window for numerator or denominator, output is NaN
// - If denominator (SMA of volume) is zero -> output NaN (avoid silent zero-division)
// - Full-length output aligned 1:1 with input
func VWMA(src, vol []float64, length int) []float64 {
    n := len(src)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if length <= 0 || n == 0 || len(vol) != n || length > n { return out }

    // Build src*vol series; let NaN/Inf propagate so SMA flags the window as invalid.
    cv := make([]float64, n)
    for i := 0; i < n; i++ {
        cv[i] = src[i] * vol[i]
    }

    num := SMA(cv, length)
    den := SMA(vol, length)

    for i := 0; i < n; i++ {
        if i < length-1 { continue }
        if math.IsNaN(num[i]) || math.IsNaN(den[i]) { continue }
        if den[i] == 0 { continue }
        out[i] = num[i] / den[i]
    }
    return out
}

// VWMAFromKlines computes VWMA from candles using custom accessors for src and volume.
func VWMAFromKlines(klines []Kline, length int, src func(Kline) float64, vol func(Kline) float64) []float64 {
    // Defensive: sort a copy by timestamp ascending to ensure stable ordering
    if len(klines) > 1 {
        cpy := make([]Kline, len(klines))
        copy(cpy, klines)
        sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
        klines = cpy
    }
    n := len(klines)
    vals := make([]float64, n)
    vols := make([]float64, n)
    for i := range klines {
        vals[i] = src(klines[i])
        vols[i] = vol(klines[i])
    }
    return VWMA(vals, vols, length)
}

// VWMABase computes VWMA using base volume (k.Volume) and the provided src accessor (e.g., close).
func VWMABase(klines []Kline, length int, src func(Kline) float64) []float64 {
    return VWMAFromKlines(klines, length, src, func(k Kline) float64 { return k.Volume })
}

// VWMAQuote computes VWMA using quote volume (close * baseVolume) and the provided src accessor (e.g., close).
func VWMAQuote(klines []Kline, length int, src func(Kline) float64) []float64 {
    return VWMAFromKlines(klines, length, src, func(k Kline) float64 { return k.Close * k.Volume })
}
