package scalping_momentium

import (
	"fmt"
	"math"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

type Generator struct {
	config signals.GeneratorConfig

	atrPeriod        int
	bodyPctMin       float64
	bodyATRMin       float64
	stochKPeriod     int
	stochKSmooth     int
	stochDPeriod     int
	stochKLongMax    float64
	stochKShortMin   float64
	enableStochCross bool

	atrValues  []float64
	stochK     []float64
	stochD     []float64

	// Optional indicators
	enableDMICross   bool
	dmiPeriod        int
	diPlus           []float64
	diMinus          []float64

	enableMFIFilter  bool
	mfiPeriod        int
	mfiOversold      float64
	mfiOverbought    float64
	mfiValues        []float64

	enableCCIFilter  bool
	cciPeriod        int
	cciOversold      float64
	cciOverbought    float64
	cciValues        []float64

	lastProcessedIdx int
	metrics          signals.GeneratorMetrics
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
	EnableStochCross bool
	// Optional filters
	EnableDMICross   bool
	DMIPeriod        int
	EnableMFIFilter  bool
	MFIPeriod        int
	MFIOversold      float64
	MFIOverbought    float64
	EnableCCIFilter  bool
	CCIPeriod        int
	CCIOversold      float64
	CCIOverbought    float64
}

func NewGenerator(cfg Config) *Generator {
	return &Generator{
		atrPeriod:      cfg.ATRPeriod,
		bodyPctMin:     cfg.BodyPctMin,
		bodyATRMin:     cfg.BodyATRMin,
		stochKPeriod:   cfg.StochKPeriod,
		stochKSmooth:   cfg.StochKSmooth,
		stochDPeriod:   cfg.StochDPeriod,
		stochKLongMax:  cfg.StochKLongMax,
		stochKShortMin: cfg.StochKShortMin,
		enableStochCross: cfg.EnableStochCross,
		enableDMICross:  cfg.EnableDMICross,
		dmiPeriod:       cfg.DMIPeriod,
		enableMFIFilter: cfg.EnableMFIFilter,
		mfiPeriod:       cfg.MFIPeriod,
		mfiOversold:     cfg.MFIOversold,
		mfiOverbought:   cfg.MFIOverbought,
		enableCCIFilter: cfg.EnableCCIFilter,
		cciPeriod:       cfg.CCIPeriod,
		cciOversold:     cfg.CCIOversold,
		cciOverbought:   cfg.CCIOverbought,
		lastProcessedIdx: -1,
	}
}

func (g *Generator) Name() string { return "ScalpingMomentium" }

func (g *Generator) Initialize(config signals.GeneratorConfig) error {
	g.config = config
	if g.atrPeriod < 1 { return fmt.Errorf("ATRPeriod doit être > 0") }
	if g.stochKPeriod < 1 { return fmt.Errorf("StochKPeriod doit être > 0") }
	if g.stochKSmooth < 1 { return fmt.Errorf("StochKSmooth doit être > 0") }
	if g.stochDPeriod < 1 { return fmt.Errorf("StochDPeriod doit être > 0") }
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
	kVals, dVals := stoch.Calculate(highs, lows, closes)
	g.stochK = kVals
	g.stochD = dVals

	// Optional indicators
	if g.enableDMICross {
		dmiInd := indicators.NewDMITVStandard(g.dmiPeriod)
		plus, minus, _ := dmiInd.Calculate(highs, lows, closes)
		g.diPlus = plus
		g.diMinus = minus
	}
	if g.enableMFIFilter {
		mfi := indicators.NewMFITVStandard(g.mfiPeriod)
		g.mfiValues = mfi.Calculate(highs, lows, closes, volumes)
	}
	if g.enableCCIFilter {
		cci := indicators.NewCCITVStandard(g.cciPeriod)
		g.cciValues = cci.Calculate(highs, lows, closes)
	}
	return nil
}

func (g *Generator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var out []signals.Signal
	if len(klines) < 3 { return out, nil }
	lastClosedIdx := len(klines) - 2
	startIdx := g.lastProcessedIdx + 1
	warmup := max3(g.atrPeriod, g.stochKPeriod+g.stochKSmooth+g.stochDPeriod, 2)
	if g.enableDMICross && g.dmiPeriod > warmup { warmup = g.dmiPeriod }
	if g.enableMFIFilter && g.mfiPeriod > warmup { warmup = g.mfiPeriod }
	if g.enableCCIFilter && g.cciPeriod > warmup { warmup = g.cciPeriod }
	if startIdx < warmup { startIdx = warmup }
	if startIdx > lastClosedIdx { return out, nil }

	for i := startIdx; i <= lastClosedIdx; i++ {
		// Indicateurs valides
		if i >= len(g.atrValues) || math.IsNaN(g.atrValues[i]) { continue }
		if i >= len(g.stochK) || math.IsNaN(g.stochK[i]) { continue }
		if g.enableStochCross {
			if i-1 < 0 || i >= len(g.stochD) || math.IsNaN(g.stochD[i]) || math.IsNaN(g.stochD[i-1]) { continue }
		}
		if g.enableDMICross {
			if i-1 < 0 || i >= len(g.diPlus) || i >= len(g.diMinus) || math.IsNaN(g.diPlus[i]) || math.IsNaN(g.diMinus[i]) || math.IsNaN(g.diPlus[i-1]) || math.IsNaN(g.diMinus[i-1]) { continue }
		}

		k := klines[i]
		rangeHL := k.High - k.Low
		if rangeHL <= 0 { continue }
		body := math.Abs(k.Close - k.Open)
		bodyPct := body / rangeHL
		if bodyPct < g.bodyPctMin { continue }
		if body < g.bodyATRMin * g.atrValues[i] { continue }

		// Direction
		var sigType signals.SignalType
		if k.Close > k.Open {
			sigType = signals.SignalTypeLong
		} else if k.Close < k.Open {
			sigType = signals.SignalTypeShort
		} else {
			continue
		}

		// Filtre Stoch valeur seuils
		if sigType == signals.SignalTypeLong {
			if !(g.stochK[i] < g.stochKLongMax) { continue }
		} else {
			if !(g.stochK[i] > g.stochKShortMin) { continue }
		}

		// Filtre croisement Stoch K/D bar-à-barre (optionnel)
		if g.enableStochCross {
			prevK := g.stochK[i-1]
			prevD := g.stochD[i-1]
			curK := g.stochK[i]
			curD := g.stochD[i]
			if sigType == signals.SignalTypeLong {
				// K traverse au-dessus de D (prev K<D et cur K>D)
				if !(prevK < prevD && curK > curD) { continue }
			} else {
				// SHORT: K traverse sous D (prev K>D et cur K<D)
				if !(prevK > prevD && curK < curD) { continue }
			}
		}

		// Filtre croisement DMI (+DI / -DI) (optionnel)
		if g.enableDMICross {
			prevPlus := g.diPlus[i-1]; prevMinus := g.diMinus[i-1]
			curPlus := g.diPlus[i];   curMinus := g.diMinus[i]
			if sigType == signals.SignalTypeLong {
				if !(prevPlus < prevMinus && curPlus > curMinus) { continue }
			} else {
				if !(prevMinus < prevPlus && curMinus > curPlus) { continue }
			}
		}

		// Filtre MFI (optionnel): éviter LONG en surachat, éviter SHORT en survente
		if g.enableMFIFilter {
			if i >= len(g.mfiValues) || math.IsNaN(g.mfiValues[i]) { continue }
			mv := g.mfiValues[i]
			if sigType == signals.SignalTypeLong {
				if mv >= g.mfiOverbought { continue }
			} else {
				if mv <= g.mfiOversold { continue }
			}
		}

		// Filtre CCI (optionnel): éviter LONG en surachat, éviter SHORT en survente
		if g.enableCCIFilter {
			if i >= len(g.cciValues) || math.IsNaN(g.cciValues[i]) { continue }
			cv := g.cciValues[i]
			if sigType == signals.SignalTypeLong {
				if cv >= g.cciOverbought { continue }
			} else {
				if cv <= g.cciOversold { continue }
			}
		}

		// Label ENTRY/EXIT via références n-1/n-2
		ref1 := refForIndex(klines[i-1], sigType)
		ref2 := refForIndex(klines[i-2], sigType)
		var action signals.SignalAction
		if sigType == signals.SignalTypeLong {
			if k.Close >= maxf(ref1, ref2) { action = signals.SignalActionEntry } else { action = signals.SignalActionExit }
		} else {
			if k.Close <= minf(ref1, ref2) { action = signals.SignalActionEntry } else { action = signals.SignalActionExit }
		}

		conf := confidence(bodyPct, body/g.atrValues[i])
		        // Prepare optional indicator values for metadata
        diPlusVal := math.NaN()
        diMinusVal := math.NaN()
        mfiVal := math.NaN()
        cciVal := math.NaN()
        if i < len(g.diPlus) { diPlusVal = g.diPlus[i] }
        if i < len(g.diMinus) { diMinusVal = g.diMinus[i] }
        if i < len(g.mfiValues) { mfiVal = g.mfiValues[i] }
        if i < len(g.cciValues) { cciVal = g.cciValues[i] }

        out = append(out, signals.Signal{
            Timestamp:  k.OpenTime,
            Action:     action,
            Type:       sigType,
            Price:      k.Close,
            Confidence: conf,
            Metadata: map[string]interface{}{
                "generator":  "scalping_momentium",
                "body":       body,
                "range":      rangeHL,
                "body_pct":   bodyPct,
                "atr":        g.atrValues[i],
                "body_to_atr": body / g.atrValues[i],
                "stoch_k":    g.stochK[i],
                "stoch_d":    g.stochD[i],
                "di_plus":    diPlusVal,
                "di_minus":   diMinusVal,
                "mfi":        mfiVal,
                "cci":        cciVal,
            },
        })
	}

	if len(out) > 0 {
		g.metrics.TotalSignals += len(out)
		entry, exit, l, s := 0,0,0,0
		accConf := 0.0
		for _, sig := range out {
			if sig.Action == signals.SignalActionEntry { entry++ } else { exit++ }
			if sig.Type == signals.SignalTypeLong { l++ } else { s++ }
			accConf += sig.Confidence
		}
		g.metrics.EntrySignals += entry
		g.metrics.ExitSignals += exit
		g.metrics.LongSignals += l
		g.metrics.ShortSignals += s
		g.metrics.AvgConfidence = accConf / float64(len(out))
		g.metrics.LastSignalTime = out[len(out)-1].Timestamp
	}

	g.lastProcessedIdx = lastClosedIdx
	return out, nil
}

func (g *Generator) GetMetrics() signals.GeneratorMetrics { return g.metrics }

func refForIndex(k signals.Kline, sigType signals.SignalType) float64 {
    if sigType == signals.SignalTypeLong {
        // LONG: si bougie rouge -> utiliser Open, si verte -> utiliser Close
        if k.Close < k.Open { return k.Open }
        return k.Close
    }
    // SHORT: si bougie rouge -> utiliser Close, si verte -> utiliser Open
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
