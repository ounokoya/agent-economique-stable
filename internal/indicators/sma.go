package indicators

import "math"

// SMA computes a TradingView-like Simple Moving Average over a float64 series.
// Rules:
// - Warm-up: indices < length-1 => NaN
// - NaN/Inf propagation: if any NaN or Inf is present in the window, output is NaN
// - Full-length output aligned 1:1 with input
// - Double precision; no internal rounding
func SMA(src []float64, length int) []float64 {
    n := len(src)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if n == 0 || length <= 0 || length > n { return out }

    // Helper to mark invalid values (treat +/-Inf as invalid like NaN)
    isBad := func(x float64) bool { return math.IsNaN(x) || math.IsInf(x, 0) }

    var sum float64
    badCount := 0

    // Prime first window [0..length-1]
    for i := 0; i < length; i++ {
        v := src[i]
        if isBad(v) { badCount++ } else { sum += v }
    }

    // First output at index length-1
    i := length - 1
    if badCount == 0 && !math.IsInf(sum, 0) && !math.IsNaN(sum) {
        out[i] = sum / float64(length)
    } else {
        out[i] = math.NaN()
    }

    // Sliding window
    // Optional re-anchor to curb drift: every K steps, recompute exact sum and badCount
    const reanchorK = 4096
    steps := 0
    for i = length; i < n; i++ {
        enter := src[i]
        leave := src[i-length]
        // update leaving
        if isBad(leave) { badCount-- } else { sum -= leave }
        // update entering
        if isBad(enter) { badCount++ } else { sum += enter }

        steps++
        // Re-anchor if needed
        if steps >= reanchorK {
            steps = 0
            sum = 0
            badCount = 0
            start := i - length + 1
            for k := start; k <= i; k++ {
                v := src[k]
                if isBad(v) { badCount++ } else { sum += v }
            }
        }

        if badCount == 0 && !math.IsInf(sum, 0) && !math.IsNaN(sum) {
            out[i] = sum / float64(length)
        } else {
            out[i] = math.NaN()
        }
    }

    return out
}

// SMAFromKlines computes a TradingView-like SMA over candles using a custom source accessor.
// Example src: func(k Kline) float64 { return k.Close }
func SMAFromKlines(klines []Kline, length int, src func(Kline) float64) []float64 {
    n := len(klines)
    vals := make([]float64, n)
    for i := range klines { vals[i] = src(klines[i]) }
    return SMA(vals, length)
}
