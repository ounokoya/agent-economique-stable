package indicators

import (
    "math"
    "sort"
)

// DirectionalMovementFromKlines computes TradingView-like raw directional movements (+DM_raw, -DM_raw).
// Rules per Wilder/TV:
// - upMove = H[i] - H[i-1]
// - downMove = L[i-1] - L[i]
// - +DM_raw[i] = upMove if upMove > downMove && upMove > 0 else 0
// - -DM_raw[i] = downMove if downMove > upMove && downMove > 0 else 0
// - i=0 is NaN
// Defensive: sorts a copy of candles by timestamp ascending before computation.
func DirectionalMovementFromKlines(klines []Kline) (plusRaw, minusRaw []float64) {
    n := len(klines)
    plusRaw = make([]float64, n)
    minusRaw = make([]float64, n)
    for i := 0; i < n; i++ { plusRaw[i], minusRaw[i] = math.NaN(), math.NaN() }
    if n == 0 { return }

    if n > 1 {
        cpy := make([]Kline, n)
        copy(cpy, klines)
        sort.Slice(cpy, func(i, j int) bool { return cpy[i].Timestamp < cpy[j].Timestamp })
        klines = cpy
    }

    plusRaw[0], minusRaw[0] = math.NaN(), math.NaN()
    for i := 1; i < n; i++ {
        h := klines[i].High
        l := klines[i].Low
        ph := klines[i-1].High
        pl := klines[i-1].Low
        // Validate inputs
        if math.IsNaN(h) || math.IsNaN(l) || math.IsNaN(ph) || math.IsNaN(pl) || math.IsInf(h,0) || math.IsInf(l,0) || math.IsInf(ph,0) || math.IsInf(pl,0) {
            plusRaw[i], minusRaw[i] = math.NaN(), math.NaN()
            continue
        }
        up := h - ph
        dn := pl - l
        var p, m float64
        if up > dn && up > 0 {
            p = up
            m = 0
        } else if dn > up && dn > 0 {
            p = 0
            m = dn
        } else {
            p = 0
            m = 0
        }
        plusRaw[i] = p
        minusRaw[i] = m
    }
    return
}
