package indicators

import (
    "math"
    "sort"
)

// MFI computes a TradingView-like Money Flow Index.
// Rules:
// - Typical Price TP = (H+L+C)/3; Raw Money Flow MF = TP*V
// - Classification vs previous TP (i>=1):
//   * TP[i] > TP[i-1] => posMF[i] = MF[i], negMF[i] = 0
//   * TP[i] < TP[i-1] => posMF[i] = 0,     negMF[i] = MF[i]
//   * TP[i] == TP[i-1] => posMF[i] = 0,    negMF[i] = 0
//   * If any of TP[i], TP[i-1], MF[i] invalid => posMF[i] = negMF[i] = NaN
// - Sliding window sums over `length` bars, window [i-length+1 .. i]
// - Warm-up: first defined index is i = length (since i=0 is invalid for classification)
// - NaN/Inf propagation: if any NaN within the window for pos or neg sums, MFI = NaN
// - Zero-division conventions:
//   * sumPos > 0 && sumNeg = 0 => 100
//   * sumPos = 0 && sumNeg > 0 => 0
//   * sumPos = 0 && sumNeg = 0 => 50
// - Exclude last bar: do not compute MFI for the in-flight last bar; force out[n-1] = NaN
// - Full-length output aligned 1:1 with inputs
func MFI(h, l, c, v []float64, length int) []float64 {
    n := len(h)
    out := make([]float64, n)
    for i := range out { out[i] = math.NaN() }
    if length <= 0 || n == 0 || len(l) != n || len(c) != n || len(v) != n || length > n {
        return out
    }

    isBad := func(x float64) bool { return math.IsNaN(x) || math.IsInf(x, 0) }

    tp := make([]float64, n)
    mf := make([]float64, n)
    for i := 0; i < n; i++ {
        H, L, C, V := h[i], l[i], c[i], v[i]
        // Invalidate bar if non-finite inputs
        if isBad(H) || isBad(L) || isBad(C) || isBad(V) {
            tp[i] = math.NaN()
            mf[i] = math.NaN()
            continue
        }
        // Normalize H/L instead of invalidating when H < L (TV tolerance)
        if H < L {
            H, L = L, H
        }
        t := (H + L + C) / 3.0
        tp[i] = t
        // Volume rules: V==0 -> MF=0 valid; V<0 -> clamp to 0 (MF=0)
        if V < 0 {
            V = 0
        }
        mf[i] = t * V
    }

    pos := make([]float64, n)
    neg := make([]float64, n)
    for i := range pos { pos[i], neg[i] = math.NaN(), math.NaN() }

    // classification
    for i := 0; i < n; i++ {
        if i == 0 {
            // no previous TP available
            pos[i], neg[i] = math.NaN(), math.NaN()
            continue
        }
        t, tPrev := tp[i], tp[i-1]
        f := mf[i]
        if isBad(t) || isBad(tPrev) || isBad(f) {
            pos[i], neg[i] = math.NaN(), math.NaN()
            continue
        }
        if t > tPrev {
            pos[i], neg[i] = f, 0
        } else if t < tPrev {
            pos[i], neg[i] = 0, f
        } else {
            // tie => zero flows
            pos[i], neg[i] = 0, 0
        }
    }

    // sliding sums with NaN propagation per window
    var sumP, sumN float64
    badP, badN := 0, 0

    // Exclude last bar computations: end is n-1; we won't emit output for index n-1
    end := n - 1
    for i := 0; i < n; i++ {
        // enter current
        if isBad(pos[i]) { badP++ } else { sumP += pos[i] }
        if isBad(neg[i]) { badN++ } else { sumN += neg[i] }

        // remove leaving when window exceeds length
        if i >= length {
            leave := i - length
            if isBad(pos[leave]) { badP-- } else { sumP -= pos[leave] }
            if isBad(neg[leave]) { badN-- } else { sumN -= neg[leave] }
        }

        // warm-up: first defined index is i >= length
        if i < length { continue }

        // Do not compute the in-flight last bar
        if i >= end { continue }

        if badP > 0 || badN > 0 {
            out[i] = math.NaN()
            continue
        }
        denom := sumP + sumN
        var val float64
        if denom == 0 {
            val = 50.0
        } else {
            val = 100.0 * (sumP / denom)
        }
        // Clamp to [0,100] and normalize signed zeros
        if val < 0 {
            val = 0
        } else if val > 100 {
            val = 100
        }
        if val == 0 { // normalize -0.0 -> 0.0
            val = 0
        }
        out[i] = val
    }

    // Force last bar to NaN explicitly
    if n > 0 {
        out[n-1] = math.NaN()
    }

    return out
}

// MFIFromKlines computes MFI from candles, defensively sorting by timestamp ascending.
func MFIFromKlines(klines []Kline, length int) []float64 {
    if len(klines) > 1 {
        cpy := make([]Kline, len(klines))
        copy(cpy, klines)
        // Stable ascending sort by timestamp to match TV-like deterministic ordering
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
    return MFI(h, l, c, v, length)
}
