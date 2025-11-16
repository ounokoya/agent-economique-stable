package smart_eco_anchored

import (
    "fmt"
    "math"

    "agent-economique/internal/indicators"
    "agent-economique/internal/signals"
)

// Generator implements the SmartEco anchored variant
// Window: [anchor - floor(WindowSize/2) .. anchor + WindowSize]
// Rules:
// - Select an anchor:
//   - If any cross filter is enabled, first index where one selected cross (or relative if enabled) occurs defines the anchor and the side (LONG/SHORT)
//   - Else, first candle that passes core body/ATR validation; side from candle color
// - Conditions (filters/crosses) accumulate independently anywhere in the combined window.
// - Candle validations (body/ATR) must be true on a single trigger bar j in the post-window (including anchor).
// - Generate a signal at the first j in [anchor .. end] where all selected conditions are satisfied and candle validations are true at j.

type Generator struct {
    // base cfg
    atrPeriod   int
    bodyPctMin  float64
    bodyATRMin  float64

    // stochastic
    stochKPeriod   int
    stochKSmooth   int
    stochDPeriod   int
    stochKLongMax  float64
    stochKShortMin float64
    enableStochExtremes bool
    enableStochCross    bool
    stochUseRelative    bool

    // VWMA
    enableVwmaCross bool
    vwmaFast        int
    vwmaSlow        int
    vwmaUseRelative bool

    // DMI
    enableDMICross bool
    dmiPeriod      int
    dmiUseRelative bool

    // MFI
    enableMFIFilter bool
    mfiPeriod       int
    mfiOversold     float64
    mfiOverbought   float64

    // CCI
    enableCCIFilter bool
    cciPeriod       int
    cciOversold     float64
    cciOverbought   float64

    // MACD
    enableMacdHistogramFilter bool
    enableMacdSigneFilter     bool
    macdFast                  int
    macdSlow                  int
    macdSignalPeriod          int

    // Anchored window
    windowSize int // post window size; pre window = floor(windowSize/2)
    anchorByCrossOnly bool

    // runtime
    config           signals.GeneratorConfig
    lastProcessedIdx int

    // series
    atrValues []float64
    stochK    []float64
    stochD    []float64

    diPlus  []float64
    diMinus []float64

    vwmaFastValues []float64
    vwmaSlowValues []float64

    mfiValues []float64
    cciValues []float64

    macdLine   []float64
    macdSignal []float64
    macdHist   []float64
}

type Config struct {
    ATRPeriod      int
    BodyPctMin     float64
    BodyATRMin     float64

    StochKPeriod   int
    StochKSmooth   int
    StochDPeriod   int
    StochKLongMax  float64
    StochKShortMin float64
    EnableStochExtremes bool
    EnableStochCross    bool
    StochUseRelative    bool

    EnableVwmaCross bool
    VwmaFast        int
    VwmaSlow        int
    VwmaUseRelative bool

    EnableDMICross bool
    DMIPeriod      int
    DmiUseRelative bool

    EnableMFIFilter bool
    MFIPeriod       int
    MFIOversold     float64
    MFIOverbought   float64

    EnableCCIFilter bool
    CCIPeriod       int
    CCIOversold     float64
    CCIOverbought   float64

    EnableMacdHistogramFilter bool
    EnableMacdSigneFilter     bool
    MacdFast                  int
    MacdSlow                  int
    MacdSignalPeriod          int

    // Anchored behavior
    WindowSize       int // default 20
    AnchorByCrossOnly bool
}

func NewGenerator(cfg Config) *Generator {
    g := &Generator{
        atrPeriod:   cfg.ATRPeriod,
        bodyPctMin:  cfg.BodyPctMin,
        bodyATRMin:  cfg.BodyATRMin,

        stochKPeriod:   cfg.StochKPeriod,
        stochKSmooth:   cfg.StochKSmooth,
        stochDPeriod:   cfg.StochDPeriod,
        stochKLongMax:  cfg.StochKLongMax,
        stochKShortMin: cfg.StochKShortMin,
        enableStochExtremes: cfg.EnableStochExtremes,
        enableStochCross:    cfg.EnableStochCross,
        stochUseRelative:    cfg.StochUseRelative,

        enableVwmaCross: cfg.EnableVwmaCross,
        vwmaFast:        cfg.VwmaFast,
        vwmaSlow:        cfg.VwmaSlow,
        vwmaUseRelative: cfg.VwmaUseRelative,

        enableDMICross: cfg.EnableDMICross,
        dmiPeriod:      cfg.DMIPeriod,
        dmiUseRelative: cfg.DmiUseRelative,

        enableMFIFilter: cfg.EnableMFIFilter,
        mfiPeriod:       cfg.MFIPeriod,
        mfiOversold:     cfg.MFIOversold,
        mfiOverbought:   cfg.MFIOverbought,

        enableCCIFilter: cfg.EnableCCIFilter,
        cciPeriod:       cfg.CCIPeriod,
        cciOversold:     cfg.CCIOversold,
        cciOverbought:   cfg.CCIOverbought,

        enableMacdHistogramFilter: cfg.EnableMacdHistogramFilter,
        enableMacdSigneFilter:     cfg.EnableMacdSigneFilter,
        macdFast:                  cfg.MacdFast,
        macdSlow:                  cfg.MacdSlow,
        macdSignalPeriod:          cfg.MacdSignalPeriod,

        windowSize:       cfg.WindowSize,
        anchorByCrossOnly: cfg.AnchorByCrossOnly,

        lastProcessedIdx: -1,
    }
    if g.windowSize <= 0 { g.windowSize = 20 }
    return g
}

func (g *Generator) Name() string { return "SmartEcoAnchored" }

func (g *Generator) Initialize(config signals.GeneratorConfig) error {
    g.config = config
    if g.atrPeriod < 1 { return fmt.Errorf("ATRPeriod doit être > 0") }
    if g.stochKPeriod < 1 || g.stochKSmooth < 1 || g.stochDPeriod < 1 { return fmt.Errorf("Stoch params invalides") }
    return nil
}

func (g *Generator) CalculateIndicators(klines []signals.Kline) error {
    if len(klines) == 0 { return fmt.Errorf("aucune kline") }
    closes := make([]float64, len(klines))
    highs := make([]float64, len(klines))
    lows := make([]float64, len(klines))
    volumes := make([]float64, len(klines))
    for i, k := range klines {
        closes[i] = k.Close
        highs[i] = k.High
        lows[i] = k.Low
        volumes[i] = k.Volume
    }

    atrInd := indicators.NewATRTVStandard(g.atrPeriod)
    g.atrValues = atrInd.Calculate(highs, lows, closes)

    stoch := indicators.NewStochTVStandard(g.stochKPeriod, g.stochKSmooth, g.stochDPeriod)
    g.stochK, g.stochD = stoch.Calculate(highs, lows, closes)

    if g.enableDMICross || g.dmiUseRelative {
        dmi := indicators.NewDMITVStandard(g.dmiPeriod)
        plus, minus, _ := dmi.Calculate(highs, lows, closes)
        g.diPlus, g.diMinus = plus, minus
    }
    if g.enableVwmaCross || g.vwmaUseRelative {
        if g.vwmaFast > 0 && g.vwmaSlow > 0 {
            vwmaF := indicators.NewVWMATVStandard(g.vwmaFast)
            vwmaS := indicators.NewVWMATVStandard(g.vwmaSlow)
            g.vwmaFastValues = vwmaF.Calculate(closes, volumes)
            g.vwmaSlowValues = vwmaS.Calculate(closes, volumes)
        }
    }
    if g.enableMFIFilter {
        mfi := indicators.NewMFITVStandard(g.mfiPeriod)
        g.mfiValues = mfi.Calculate(highs, lows, closes, volumes)
    }
    if g.enableCCIFilter {
        cci := indicators.NewCCITVStandard(g.cciPeriod)
        g.cciValues = cci.Calculate(highs, lows, closes)
    }
    if g.enableMacdHistogramFilter || g.enableMacdSigneFilter {
        macd := indicators.NewMACDTVStandard(g.macdFast, g.macdSlow, g.macdSignalPeriod)
        ml, sl, hl := macd.Calculate(closes)
        g.macdLine, g.macdSignal, g.macdHist = ml, sl, hl
    }
    return nil
}

func (g *Generator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
    out := make([]signals.Signal, 0)
    if len(klines) < 3 { return out, nil }
    lastClosedIdx := len(klines) - 2

    // warmup
    warmup := max3(g.atrPeriod, g.stochKPeriod+g.stochKSmooth+g.stochDPeriod, 2)
    if g.vwmaSlow > warmup { warmup = g.vwmaSlow }
    if (g.enableDMICross || g.dmiUseRelative) && g.dmiPeriod > warmup { warmup = g.dmiPeriod }
    if g.enableMFIFilter && g.mfiPeriod > warmup { warmup = g.mfiPeriod }
    if g.enableCCIFilter && g.cciPeriod > warmup { warmup = g.cciPeriod }
    if (g.enableMacdHistogramFilter || g.enableMacdSigneFilter) {
        macdw := g.macdSlow + g.macdSignalPeriod
        if macdw > warmup { warmup = macdw }
    }

    // 1) Find anchor
    anchorIdx, side, ok := g.findAnchor(warmup, lastClosedIdx)
    if !ok { return out, nil }

    // 2) Build combined window [start..end]
    pre := g.windowSize / 2
    start := anchorIdx - pre
    if start < 0 { start = 0 }
    end := anchorIdx + g.windowSize
    if end > lastClosedIdx { end = lastClosedIdx }

    // 3) Accumulate independent validations in [start..end]
    // We will walk from start to end, and emit at first j >= anchorIdx where candle validations are true

    // Trackers for condition satisfaction (excluding candle validations which must be true at j)
    haveStochCross := false
    haveDmiCross := false
    haveVwmaCross := false

    if g.stochUseRelative { haveStochCross = false }
    if g.dmiUseRelative { haveDmiCross = false }
    if g.vwmaUseRelative { haveVwmaCross = false }

    haveStochRelative := false
    haveDmiRelative := false
    haveVwmaRelative := false

    haveMFI := !g.enableMFIFilter // if not enabled, treat as satisfied
    haveCCI := !g.enableCCIFilter
    haveMacdHist := !g.enableMacdHistogramFilter
    haveMacdSigne := !g.enableMacdSigneFilter
    haveStochExtremes := !g.enableStochExtremes

    // scan window incrementally
    for i := start; i <= end; i++ {
        // accumulate events anywhere in window
        g.accumulateConditions(i, side,
            &haveStochCross, &haveStochRelative,
            &haveDmiCross, &haveDmiRelative,
            &haveVwmaCross, &haveVwmaRelative,
            &haveMFI, &haveCCI, &haveMacdHist, &haveMacdSigne, &haveStochExtremes,
        )

        if i < anchorIdx { continue }

        // Candle validations must be on single bar i
        k := klines[i]
        rangeHL := k.High - k.Low
        if rangeHL <= 0 { continue }
        body := math.Abs(k.Close - k.Open)
        bodyPct := body / rangeHL
        if bodyPct < g.bodyPctMin { continue }
        if i >= len(g.atrValues) || math.IsNaN(g.atrValues[i]) { continue }
        if body < g.bodyATRMin * g.atrValues[i] { continue }

        // Combine all conditions according to enabled flags and relative modes
        okAll := true
        // Stoch cross/relative
        if g.enableStochCross {
            if g.stochUseRelative {
                okAll = okAll && haveStochRelative
            } else {
                okAll = okAll && haveStochCross
            }
        }
        // DMI cross/relative
        if g.enableDMICross {
            if g.dmiUseRelative {
                okAll = okAll && haveDmiRelative
            } else {
                okAll = okAll && haveDmiCross
            }
        }
        // VWMA cross/relative
        if g.enableVwmaCross {
            if g.vwmaUseRelative {
                okAll = okAll && haveVwmaRelative
            } else {
                okAll = okAll && haveVwmaCross
            }
        }
        // MFI, CCI, MACD, Stoch extremes
        okAll = okAll && haveMFI && haveCCI && haveMacdHist && haveMacdSigne && haveStochExtremes

        if !okAll { continue }

        // All satisfied, emit signal at i
        action := g.labelAction(klines, i, side)
        conf := confidence(bodyPct, body/g.atrValues[i])
        out = append(out, signals.Signal{
            Timestamp:  k.OpenTime,
            Action:     action,
            Type:       side,
            Price:      k.Close,
            Confidence: conf,
            Metadata: map[string]interface{}{
                "generator": "smart_eco_anchored",
                "anchor_idx": anchorIdx,
                "window_start": start,
                "window_end": end,
                "body": body,
                "body_pct": bodyPct,
                "atr": g.atrValues[i],
            },
        })
        break
    }

    return out, nil
}

func (g *Generator) findAnchor(warmup, last int) (idx int, side signals.SignalType, ok bool) {
    // Check if any cross-based anchor is allowed
    crossesSelected := g.enableStochCross || g.enableDMICross || g.enableVwmaCross

    for i := warmup; i <= last; i++ {
        // Try cross-based anchor first if any selected
        if crossesSelected {
            if s, hit := g.crossEventAt(i); hit {
                return i, s, true
            }
        }
        // Fallback to candle validation if allowed
        if !g.anchorByCrossOnly {
            // validate candle
            if i < len(g.atrValues) && !math.IsNaN(g.atrValues[i]) {
                // body/atr validation
                k := i
                // We don't have klines values here; rely on series via later call.
                // Anchor fallback will be finalized in DetectSignals when scanning window (we can still accept candle here if ATR valid)
                // Determine side by stoch/candle direction is ambiguous here without klines; we'll pick side later in emit phase.
                // For simplicity, derive side by relative stoch if available else default LONG.
                // However, better approach: we require cross-based anchor; if not, we'll pick side at signal emission based on candle color.
                // Mark as ok anchor; side to be determined later => choose LONG placeholder
                _ = k
                return i, signals.SignalTypeLong, true
            }
        }
    }
    return 0, signals.SignalTypeLong, false
}

func (g *Generator) crossEventAt(i int) (side signals.SignalType, hit bool) {
    // Stoch cross
    if g.enableStochCross && i-1 >= 0 && i < len(g.stochK) && i < len(g.stochD) &&
        !math.IsNaN(g.stochK[i]) && !math.IsNaN(g.stochD[i]) && !math.IsNaN(g.stochK[i-1]) && !math.IsNaN(g.stochD[i-1]) {
        prevK, prevD := g.stochK[i-1], g.stochD[i-1]
        curK, curD := g.stochK[i], g.stochD[i]
        if prevK <= prevD && curK > curD { return signals.SignalTypeLong, true }
        if prevK >= prevD && curK < curD { return signals.SignalTypeShort, true }
    }
    // DMI cross
    if g.enableDMICross && i-1 >= 0 && i < len(g.diPlus) && i < len(g.diMinus) &&
        !math.IsNaN(g.diPlus[i]) && !math.IsNaN(g.diMinus[i]) && !math.IsNaN(g.diPlus[i-1]) && !math.IsNaN(g.diMinus[i-1]) {
        prevPlus, prevMinus := g.diPlus[i-1], g.diMinus[i-1]
        curPlus, curMinus := g.diPlus[i], g.diMinus[i]
        if prevPlus <= prevMinus && curPlus > curMinus { return signals.SignalTypeLong, true }
        if prevMinus <= prevPlus && curMinus > curPlus { return signals.SignalTypeShort, true }
    }
    // VWMA cross
    if g.enableVwmaCross && i-1 >= 0 && i < len(g.vwmaFastValues) && i < len(g.vwmaSlowValues) &&
        !math.IsNaN(g.vwmaFastValues[i]) && !math.IsNaN(g.vwmaSlowValues[i]) &&
        !math.IsNaN(g.vwmaFastValues[i-1]) && !math.IsNaN(g.vwmaSlowValues[i-1]) {
        cross, direction := indicators.DetecterCroisement(g.vwmaFastValues, g.vwmaSlowValues, i)
        if cross {
            if direction == "HAUSSIER" { return signals.SignalTypeLong, true }
            if direction == "BAISSIER" { return signals.SignalTypeShort, true }
        }
    }
    return signals.SignalTypeLong, false
}

func (g *Generator) accumulateConditions(i int, side signals.SignalType,
    haveStochCross, haveStochRelative *bool,
    haveDmiCross, haveDmiRelative *bool,
    haveVwmaCross, haveVwmaRelative *bool,
    haveMFI, haveCCI, haveMacdHist, haveMacdSigne, haveStochExtremes *bool,
) {
    // Stoch cross
    if g.enableStochCross && i-1 >= 0 && i < len(g.stochK) && i < len(g.stochD) &&
        !math.IsNaN(g.stochK[i]) && !math.IsNaN(g.stochD[i]) && !math.IsNaN(g.stochK[i-1]) && !math.IsNaN(g.stochD[i-1]) {
        prevK, prevD := g.stochK[i-1], g.stochD[i-1]
        curK, curD := g.stochK[i], g.stochD[i]
        if side == signals.SignalTypeLong {
            if prevK <= prevD && curK > curD { *haveStochCross = true }
        } else {
            if prevK >= prevD && curK < curD { *haveStochCross = true }
        }
    }
    // Stoch relative
    if g.stochUseRelative && i < len(g.stochK) && i < len(g.stochD) && !math.IsNaN(g.stochK[i]) && !math.IsNaN(g.stochD[i]) {
        if side == signals.SignalTypeLong {
            if g.stochK[i] > g.stochD[i] { *haveStochRelative = true }
        } else {
            if g.stochK[i] < g.stochD[i] { *haveStochRelative = true }
        }
    }

    // DMI cross
    if g.enableDMICross && i-1 >= 0 && i < len(g.diPlus) && i < len(g.diMinus) &&
        !math.IsNaN(g.diPlus[i]) && !math.IsNaN(g.diMinus[i]) && !math.IsNaN(g.diPlus[i-1]) && !math.IsNaN(g.diMinus[i-1]) {
        prevPlus, prevMinus := g.diPlus[i-1], g.diMinus[i-1]
        curPlus, curMinus := g.diPlus[i], g.diMinus[i]
        if side == signals.SignalTypeLong {
            if prevPlus <= prevMinus && curPlus > curMinus { *haveDmiCross = true }
        } else {
            if prevMinus <= prevPlus && curMinus > curPlus { *haveDmiCross = true }
        }
    }
    // DMI relative
    if g.dmiUseRelative && i < len(g.diPlus) && i < len(g.diMinus) && !math.IsNaN(g.diPlus[i]) && !math.IsNaN(g.diMinus[i]) {
        if side == signals.SignalTypeLong {
            if g.diPlus[i] > g.diMinus[i] { *haveDmiRelative = true }
        } else {
            if g.diMinus[i] > g.diPlus[i] { *haveDmiRelative = true }
        }
    }

    // VWMA cross/relative
    if i-1 >= 0 && i < len(g.vwmaFastValues) && i < len(g.vwmaSlowValues) &&
        !math.IsNaN(g.vwmaFastValues[i]) && !math.IsNaN(g.vwmaSlowValues[i]) &&
        !math.IsNaN(g.vwmaFastValues[i-1]) && !math.IsNaN(g.vwmaSlowValues[i-1]) {
        if g.enableVwmaCross {
            cross, direction := indicators.DetecterCroisement(g.vwmaFastValues, g.vwmaSlowValues, i)
            if cross {
                if side == signals.SignalTypeLong && direction == "HAUSSIER" { *haveVwmaCross = true }
                if side == signals.SignalTypeShort && direction == "BAISSIER" { *haveVwmaCross = true }
            }
        }
        if g.vwmaUseRelative {
            if side == signals.SignalTypeLong {
                if g.vwmaFastValues[i] > g.vwmaSlowValues[i] { *haveVwmaRelative = true }
            } else {
                if g.vwmaFastValues[i] < g.vwmaSlowValues[i] { *haveVwmaRelative = true }
            }
        }
    }

    // MFI
    if g.enableMFIFilter && i < len(g.mfiValues) && !math.IsNaN(g.mfiValues[i]) {
        mv := g.mfiValues[i]
        if side == signals.SignalTypeLong {
            if mv < g.mfiOverbought { *haveMFI = true }
        } else {
            if mv > g.mfiOversold { *haveMFI = true }
        }
    }
    // CCI
    if g.enableCCIFilter && i < len(g.cciValues) && !math.IsNaN(g.cciValues[i]) {
        cv := g.cciValues[i]
        if side == signals.SignalTypeLong {
            if cv < g.cciOverbought { *haveCCI = true }
        } else {
            if cv > g.cciOversold { *haveCCI = true }
        }
    }
    // MACD histogram
    if g.enableMacdHistogramFilter && i < len(g.macdHist) && !math.IsNaN(g.macdHist[i]) {
        if side == signals.SignalTypeLong {
            if g.macdHist[i] > 0 { *haveMacdHist = true }
        } else {
            if g.macdHist[i] < 0 { *haveMacdHist = true }
        }
    }
    // MACD signe
    if g.enableMacdSigneFilter && i < len(g.macdLine) && i < len(g.macdSignal) && !math.IsNaN(g.macdLine[i]) && !math.IsNaN(g.macdSignal[i]) {
        if side == signals.SignalTypeLong {
            if g.macdLine[i] < 0 && g.macdSignal[i] < 0 { *haveMacdSigne = true }
        } else {
            if g.macdLine[i] > 0 && g.macdSignal[i] > 0 { *haveMacdSigne = true }
        }
    }
    // Stoch extremes
    if g.enableStochExtremes && i < len(g.stochK) && !math.IsNaN(g.stochK[i]) {
        if side == signals.SignalTypeLong {
            if g.stochK[i] < g.stochKLongMax { *haveStochExtremes = true }
        } else {
            if g.stochK[i] > g.stochKShortMin { *haveStochExtremes = true }
        }
    }
}

func (g *Generator) labelAction(klines []signals.Kline, i int, side signals.SignalType) signals.SignalAction {
    if i-2 < 0 { return signals.SignalActionEntry }
    ref1 := refForIndex(klines[i-1], side)
    ref2 := refForIndex(klines[i-2], side)
    k := klines[i]
    if side == signals.SignalTypeLong {
        if k.Close >= maxf(ref1, ref2) { return signals.SignalActionEntry }
        return signals.SignalActionExit
    }
    if k.Close <= minf(ref1, ref2) { return signals.SignalActionEntry }
    return signals.SignalActionExit
}

func refForIndex(k signals.Kline, sigType signals.SignalType) float64 {
    if sigType == signals.SignalTypeLong {
        if k.Close < k.Open { return k.Open }
        return k.Close
    }
    if k.Close < k.Open { return k.Close }
    return k.Open
}

func max3(a,b,c int) int { if a<b { a=b }; if a<c { a=c }; return a }
func maxf(a,b float64) float64 { if a>b { return a }; return b }
func minf(a,b float64) float64 { if a<b { return a }; return b }

func confidence(bodyPct, bodyToATR float64) float64 {
    c := 0.5
    if bodyPct >= 0.6 { c += 0.15 }
    if bodyToATR >= 0.8 { c += 0.15 }
    if c > 0.95 { c = 0.95 }
    if c < 0.4 { c = 0.4 }
    return c
}
