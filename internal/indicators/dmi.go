package indicators

import "math"

// DMIFromKlines computes TradingView-like DI+, DI-, DX, and ADX from candles using Wilder's RMA.
// Uses helper primitives: TrueRangeFromKlines, DirectionalMovementFromKlines, RMA.
// Returns full-length slices aligned to the input, with NaN during warm-ups.
func DMIFromKlines(klines []Kline, length int) (plusDI, minusDI, dx, adx []float64) {
    n := len(klines)
    plusDI = make([]float64, n)
    minusDI = make([]float64, n)
    dx = make([]float64, n)
    adx = make([]float64, n)
    for i := 0; i < n; i++ { plusDI[i], minusDI[i], dx[i], adx[i] = math.NaN(), math.NaN(), math.NaN(), math.NaN() }
    if n == 0 || length <= 0 { return }

    tr := TrueRangeFromKlines(klines)                     // TR: NaN at i=0, valid from i>=1
    plusRaw, minusRaw := DirectionalMovementFromKlines(klines)   // Raw DM: NaN at i=0, valid from i>=1

    atr := RMA(tr, length)          // ATR via Wilder RMA
    pDM := RMA(plusRaw, length)     // +DM smoothed
    mDM := RMA(minusRaw, length)    // -DM smoothed

    // DI+ and DI-
    for i := 0; i < n; i++ {
        a := atr[i]
        p := pDM[i]
        m := mDM[i]
        if math.IsNaN(a) || math.IsNaN(p) || math.IsNaN(m) { continue }
        if a == 0 {
            plusDI[i] = 0
            minusDI[i] = 0
            dx[i] = 0
            continue
        }
        plus := 100.0 * (p / a)
        minus := 100.0 * (m / a)
        plusDI[i] = plus
        minusDI[i] = minus
        denom := plus + minus
        if denom == 0 {
            dx[i] = 0
        } else {
            d := 100.0 * math.Abs(plus - minus) / denom
            dx[i] = d
        }
    }

    // ADX = RMA(DX, length)
    adx = RMA(dx, length)
    return
}
