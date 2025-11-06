package indicators

import (
    "math"
    "sort"
)

// MACDCalculate computes a TradingView-like MACD triplet from a price series.
// Returns three full-length slices (macd, signal, hist), with NaNs during warm-up.
// Defaults (if used in demos): fast=12, slow=26, signal=9.
// Behavior:
// - macd = EMA(src, fast) - EMA(src, slow)
// - signal = EMA(macd, signalLen)
// - hist = macd - signal
// Warm-up indices:
// - EMA_fast first at fast-1; EMA_slow first at slow-1
// - macd first at slow-1
// - signal and hist first at slow+signalLen-2
func MACDCalculate(src []float64, fast, slow, signalLen int) (macd, signal, hist []float64) {
    n := len(src)
    macd = make([]float64, n)
    signal = make([]float64, n)
    hist = make([]float64, n)
    for i := 0; i < n; i++ { macd[i], signal[i], hist[i] = math.NaN(), math.NaN(), math.NaN() }
    if n == 0 || fast <= 0 || slow <= 0 || signalLen <= 0 || fast >= slow { return }

    // Use full data length instead of n-1
    if n <= 0 { return }
    if slow > n || signalLen > n { return }

    emaF := EMA(src, fast)
    emaS := EMA(src, slow)

    macdCore := make([]float64, n)
    for i := 0; i < n; i++ {
        if math.IsNaN(emaF[i]) || math.IsNaN(emaS[i]) {
            macdCore[i] = math.NaN()
            continue
        }
        macdCore[i] = emaF[i] - emaS[i]
        macd[i] = macdCore[i]
    }

    sigCore := EMA(macdCore, signalLen)
    for i := 0; i < n; i++ {
        signal[i] = sigCore[i]
        if !math.IsNaN(macd[i]) && !math.IsNaN(signal[i]) {
            hist[i] = macd[i] - signal[i]
        }
    }

    return
}

// MACDFromKlines computes MACD from candles with a custom source accessor.
// Defensive: sorts a copy by timestamp ascending before computation.
func MACDFromKlines(klines []Kline, fast, slow, signalLen int, src func(Kline) float64) (macd, signal, hist []float64) {
    if len(klines) > 1 {
        cpy := make([]Kline, len(klines))
        copy(cpy, klines)
        sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
        klines = cpy
    }
    vals := make([]float64, len(klines))
    for i := range klines { vals[i] = src(klines[i]) }
    return MACDCalculate(vals, fast, slow, signalLen)
}
