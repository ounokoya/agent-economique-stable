package ban_fin_momentium

import (
	"fmt"
	"math"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

type Config struct {
	ATRPeriod         int
	BodyATRMultiplier float64

	// Volume gating
	VolumeSMAPeriod int
	VolumeCoeff     float64 // volume_normalized > VolumeCoeff * SMA(volume)

	// VWMA fast/slow (trend gate)
	VWMAFastPeriod      int
	VWMASlowPeriod      int
	EnableVWMATrendGate bool // if true, require fast>slow for LONG and fast<slow for SHORT
	EnableVWMACross     bool

	// Aggregation
	Aggregate3    bool // if true, use 3-candle aggregation for body/volume/color
	EnableBarGate bool

	// Setup toggles (enable/disable independently)
	EnableOpenLong   bool
	EnableOpenShort  bool
	EnableCloseLong  bool
	EnableCloseShort bool

	// Stochastic
	StochKPeriod int
	StochKSmooth int
	StochDPeriod int
	// Filter toggles
	EnableStochCross bool // optional cross K vs D in direction
	// Per-direction Stoch K extremes (range-style)
	StochKOversoldLong    float64
	StochKOverboughtLong  float64
	StochKOversoldShort   float64
	StochKOverboughtShort float64
	// Independent toggles per Stoch extreme
	EnableStochKOversoldLongGate    bool
	EnableStochKOverboughtLongGate  bool
	EnableStochKOversoldShortGate   bool
	EnableStochKOverboughtShortGate bool

	// MFI (per-direction extremes, both directions)
	MFIPeriod          int
	MFIOverboughtLong  float64 // used for OPEN LONG gating ("non en sur-achat")
	MFIOversoldShort   float64 // used for OPEN SHORT gating ("non en sur-vente")
	MFIOverboughtShort float64 // used for OPEN SHORT gating upper bound
	MFIOversoldLong    float64 // used for OPEN LONG gating lower bound

	// CCI (per-direction extremes, both directions)
	CCIPeriod          int
	CCIOverboughtLong  float64 // used for OPEN LONG gating ("non en sur-achat")
	CCIOversoldShort   float64 // used for OPEN SHORT gating ("non en sur-vente")
	CCIOverboughtShort float64 // used for OPEN SHORT gating upper bound
	CCIOversoldLong    float64 // used for OPEN LONG gating lower bound

	// MACD (line, signal, histogram)
	MACDFastPeriod      int
	MACDSlowPeriod      int
	MACDSignalPeriod    int
	EnableMACDCrossGate bool
	EnableMACDSignGate  bool
	EnableMACDHistGate  bool

	// Filter toggles
	EnableMFIGate bool
	EnableCCIGate bool
	// Independent toggles per MFI extreme
	EnableMFIOversoldLongGate    bool
	EnableMFIOverboughtLongGate  bool
	EnableMFIOversoldShortGate   bool
	EnableMFIOverboughtShortGate bool
	// Independent toggles per CCI extreme
	EnableCCIOversoldLongGate    bool
	EnableCCIOverboughtLongGate  bool
	EnableCCIOversoldShortGate   bool
	EnableCCIOverboughtShortGate bool
}

// EvaluateLast evaluates only the last closed candle and returns at most one signal
// Priority order: OPEN LONG, OPEN SHORT, CLOSE LONG, CLOSE SHORT
func (g *Generator) EvaluateLast(klines []signals.Kline) (*signals.Signal, error) {
	if len(klines) < 3 {
		return nil, nil
	}
	i := len(klines) - 1 // assume all are closed in this slice, like ban_fin
	// Validate indicators are ready (only those needed)
	if i >= len(g.atr) || math.IsNaN(g.atr[i]) {
		return nil, nil
	}
	if i >= len(g.volSMA) || math.IsNaN(g.volSMA[i]) {
		return nil, nil
	}
	// Stoch n'est requis que si le cross ou les bornes K sont utilisés.
	stochExtGatesEnabled := g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
		g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate
	needStoch := g.cfg.EnableStochCross || stochExtGatesEnabled
	needMFI := g.cfg.EnableMFIGate
	needCCI := g.cfg.EnableCCIGate
	needVWMA := g.cfg.EnableVWMATrendGate || g.cfg.EnableVWMACross
	needMACD := g.cfg.EnableMACDCrossGate || g.cfg.EnableMACDSignGate || g.cfg.EnableMACDHistGate
	// Si aucun gate n'est actif (bar, Stoch, MFI, CCI, VWMA, MACD), ne pas produire de signaux.
	hasAnyGate := g.cfg.EnableBarGate || g.cfg.EnableStochCross || stochExtGatesEnabled || needMFI || needCCI || needVWMA || needMACD
	if !hasAnyGate {
		return nil, nil
	}
	if needStoch {
		if i >= len(g.stochK) || math.IsNaN(g.stochK[i]) {
			return nil, nil
		}
		if g.cfg.EnableStochCross {
			if i >= len(g.stochD) || math.IsNaN(g.stochD[i]) {
				return nil, nil
			}
		}
	}
	if needMFI {
		if i >= len(g.mfi) || math.IsNaN(g.mfi[i]) {
			return nil, nil
		}
	}
	if needCCI {
		if i >= len(g.cci) || math.IsNaN(g.cci[i]) {
			return nil, nil
		}
	}
	if needVWMA {
		if i >= len(g.vwmaFast) || math.IsNaN(g.vwmaFast[i]) {
			return nil, nil
		}
		if i >= len(g.vwmaSlow) || math.IsNaN(g.vwmaSlow[i]) {
			return nil, nil
		}
	}
	if needMACD {
		if i >= len(g.macdLine) || math.IsNaN(g.macdLine[i]) {
			return nil, nil
		}
		if i >= len(g.macdSignal) || math.IsNaN(g.macdSignal[i]) {
			return nil, nil
		}
		if i >= len(g.macdHist) || math.IsNaN(g.macdHist[i]) {
			return nil, nil
		}
	}

	body, isGreen, isRed, volNorm := g.bodyVolumeColorAt(klines, i)
	b2atr := body / g.atr[i]
	v2sma := volNorm / g.volSMA[i]
	crossUp, crossDown := false, false
	if g.cfg.EnableStochCross {
		up, down := g.stochCrossAt(i)
		crossUp, crossDown = up, down
	}
	vwUp, vwDown := false, false
	if g.cfg.EnableVWMACross {
		vwUp, vwDown = g.vwmaCrossAt(i)
	}

	// Trend gating via optional VWMA combine
	trendOKLong := true
	trendOKShort := true
	if g.cfg.EnableVWMATrendGate {
		trendOKLong = g.vwmaFast[i] > g.vwmaSlow[i]
		trendOKShort = g.vwmaFast[i] < g.vwmaSlow[i]
	}

	// Step1: body/ATR + volume/SMA + couleur comme filtre de base
	barOKLong, barOKShort := true, true
	if g.cfg.EnableBarGate {
		barOKLong = isGreen && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
		barOKShort = isRed && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
	}

	// CCI court: extrême inverse de la tendance (filtre optionnel via EnableCCIGate)
	cciOKLong, cciOKShort := true, true
	if g.cfg.EnableCCIGate {
		cciVal := g.cci[i]
		if g.cfg.EnableCCIOversoldLongGate {
			cciOKLong = cciOKLong && cciVal >= g.cfg.CCIOversoldLong
		}
		if g.cfg.EnableCCIOverboughtLongGate {
			cciOKLong = cciOKLong && cciVal <= g.cfg.CCIOverboughtLong
		}
		if g.cfg.EnableCCIOversoldShortGate {
			cciOKShort = cciOKShort && cciVal >= g.cfg.CCIOversoldShort
		}
		if g.cfg.EnableCCIOverboughtShortGate {
			cciOKShort = cciOKShort && cciVal <= g.cfg.CCIOverboughtShort
		}
	}

	// MFI: extrême inverse de la tendance (filtre optionnel via EnableMFIGate)
	mfiOKLong, mfiOKShort := true, true
	if g.cfg.EnableMFIGate {
		mfiVal := g.mfi[i]
		if g.cfg.EnableMFIOversoldLongGate {
			mfiOKLong = mfiOKLong && mfiVal >= g.cfg.MFIOversoldLong
		}
		if g.cfg.EnableMFIOverboughtLongGate {
			mfiOKLong = mfiOKLong && mfiVal <= g.cfg.MFIOverboughtLong
		}
		if g.cfg.EnableMFIOversoldShortGate {
			mfiOKShort = mfiOKShort && mfiVal >= g.cfg.MFIOversoldShort
		}
		if g.cfg.EnableMFIOverboughtShortGate {
			mfiOKShort = mfiOKShort && mfiVal <= g.cfg.MFIOverboughtShort
		}
	}

	// Cross Stoch obligatoire si EnableStochCross, sinon pas de filtre de croisement
	longCrossOK := true
	shortCrossOK := true
	if g.cfg.EnableStochCross {
		longCrossOK = crossUp
		shortCrossOK = crossDown
	}
	stochKOKLong := true
	stochKOKShort := true
	if g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
		g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate {
		kVal := g.stochK[i]
		if g.cfg.EnableStochKOversoldLongGate {
			// LONG: oversold_long borne haute d'acceptation (ex: K <= 20)
			stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOversoldLong
		}
		if g.cfg.EnableStochKOverboughtLongGate {
			// LONG: overbought_long borne haute d'acceptation (ex: K <= 80)
			stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOverboughtLong
		}
		if g.cfg.EnableStochKOversoldShortGate {
			// SHORT: oversold_short borne basse d'acceptation (ex: K >= 20)
			stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOversoldShort
		}
		if g.cfg.EnableStochKOverboughtShortGate {
			// SHORT: overbought_short borne basse d'acceptation (ex: K >= 80)
			stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOverboughtShort
		}
	}
	vwCrossOKLong := true
	vwCrossOKShort := true
	if g.cfg.EnableVWMACross {
		vwCrossOKLong = vwUp
		vwCrossOKShort = vwDown
	}
	macdCrossOKLong := true
	macdCrossOKShort := true
	if g.cfg.EnableMACDCrossGate {
		up, down := g.macdCrossAt(i)
		macdCrossOKLong = up
		macdCrossOKShort = down
	}
	macdSignOKLong := true
	macdSignOKShort := true
	if g.cfg.EnableMACDSignGate {
		m := g.macdLine[i]
		s := g.macdSignal[i]
		macdSignOKLong = m < 0 && s < 0
		macdSignOKShort = m > 0 && s > 0
	}
	macdHistOKLong := true
	macdHistOKShort := true
	if g.cfg.EnableMACDHistGate {
		h := g.macdHist[i]
		macdHistOKLong = h > 0
		macdHistOKShort = h < 0
	}

	// Signal directionnel LONG
	if g.cfg.EnableOpenLong && trendOKLong && longCrossOK && stochKOKLong && vwCrossOKLong && macdCrossOKLong && macdSignOKLong && macdHistOKLong && barOKLong && cciOKLong && mfiOKLong {
		return g.build(signals.SignalActionEntry, signals.SignalTypeLong, klines, i, b2atr, v2sma, isGreen, isRed), nil
	}

	// Signal directionnel SHORT
	if g.cfg.EnableOpenShort && trendOKShort && shortCrossOK && stochKOKShort && vwCrossOKShort && macdCrossOKShort && macdSignOKShort && macdHistOKShort && barOKShort && cciOKShort && mfiOKShort {
		return g.build(signals.SignalActionEntry, signals.SignalTypeShort, klines, i, b2atr, v2sma, isGreen, isRed), nil
	}

	return nil, nil
}

func (g *Generator) build(act signals.SignalAction, typ signals.SignalType, klines []signals.Kline, i int, b2atr, v2sma float64, isGreen, isRed bool) *signals.Signal {
	bodyVal, _, _, volNorm := g.bodyVolumeColorAt(klines, i)
	meta := map[string]interface{}{
		"generator":   "ban_fin_momentium",
		"mode":        "MOMENTUM",
		"agg3":        g.cfg.Aggregate3,
		"body":        bodyVal,
		"atr10":       g.atr[i],
		"body_to_atr": b2atr,
		"volume":      volNorm,
		"vol_sma":     g.volSMA[i],
		"vol_to_sma":  v2sma,
		"stoch_k":     g.stochK[i],
		"stoch_d":     g.stochD[i],
		"cci":         g.cci[i],
	}
	addIfFinite(meta, "mfi", g.mfi[i])
	if len(g.vwmaFast) > i && len(g.vwmaSlow) > i {
		addIfFinite(meta, "vwma_fast", g.vwmaFast[i])
		addIfFinite(meta, "vwma_slow", g.vwmaSlow[i])
	}
	if snap := g.gatingSnapshot(i, b2atr, v2sma, isGreen, isRed); snap != nil {
		meta["gating_snapshot"] = snap
	}
	conf := confidence(b2atr, v2sma)
	return &signals.Signal{Timestamp: klines[i].OpenTime, Action: act, Type: typ, Price: klines[i].Close, Confidence: conf, Metadata: meta}
}

func (g *Generator) gatingSnapshot(i int, b2atr, v2sma float64, isGreen, isRed bool) map[string]interface{} {
	snap := map[string]interface{}{
		"i":        i,
		"is_green": isGreen,
		"is_red":   isRed,
	}
	addIfFinite(snap, "body_to_atr", b2atr)
	addIfFinite(snap, "vol_to_sma", v2sma)
	addIfFinite(snap, "vwma_fast", safeAt(g.vwmaFast, i))
	addIfFinite(snap, "vwma_slow", safeAt(g.vwmaSlow, i))
	addIfFinite(snap, "stoch_k", g.stochK[i])
	addIfFinite(snap, "stoch_d", g.stochD[i])
	addIfFinite(snap, "mfi", g.mfi[i])
	addIfFinite(snap, "cci", g.cci[i])
	return snap
}

type Generator struct {
	cfg    Config
	config signals.GeneratorConfig

	atr        []float64
	vwmaFast   []float64
	vwmaSlow   []float64
	stochK     []float64
	stochD     []float64
	mfi        []float64
	cci        []float64
	volSMA     []float64
	macdLine   []float64
	macdSignal []float64
	macdHist   []float64

	lastProcessedIdx int
	metrics          signals.GeneratorMetrics
}

func NewGenerator(cfg Config) *Generator { return &Generator{cfg: cfg, lastProcessedIdx: -1} }

func (g *Generator) Name() string { return "BAN_FIN_MOMENTIUM" }

func (g *Generator) Initialize(config signals.GeneratorConfig) error {
	g.config = config
	if g.cfg.ATRPeriod <= 0 {
		return fmt.Errorf("ATRPeriod must be >0")
	}
	if g.cfg.BodyATRMultiplier <= 0 {
		return fmt.Errorf("BodyATRMultiplier must be >0")
	}
	if g.cfg.VolumeSMAPeriod <= 0 {
		return fmt.Errorf("VolumeSMAPeriod must be >0")
	}
	if g.cfg.EnableVWMATrendGate {
		if g.cfg.VWMAFastPeriod <= 0 || g.cfg.VWMASlowPeriod <= 0 {
			return fmt.Errorf("VWMA periods must be >0 when VWMATrendGate enabled")
		}
	}
	if g.cfg.StochKPeriod <= 0 || g.cfg.StochKSmooth <= 0 || g.cfg.StochDPeriod <= 0 {
		return fmt.Errorf("stoch periods must be >0")
	}
	if g.cfg.MFIPeriod <= 0 {
		return fmt.Errorf("MFIPeriod must be >0")
	}
	if g.cfg.CCIPeriod <= 0 {
		return fmt.Errorf("CCIPeriod must be >0")
	}
	needMACD := g.cfg.EnableMACDCrossGate || g.cfg.EnableMACDSignGate || g.cfg.EnableMACDHistGate
	if needMACD {
		if g.cfg.MACDFastPeriod <= 0 || g.cfg.MACDSlowPeriod <= 0 || g.cfg.MACDSignalPeriod <= 0 {
			return fmt.Errorf("MACD periods must be >0 when MACD gates are enabled")
		}
	}
	return nil
}

func (g *Generator) CalculateIndicators(klines []signals.Kline) error {
	if len(klines) == 0 {
		return fmt.Errorf("no klines")
	}
	N := len(klines)
	highs := make([]float64, N)
	lows := make([]float64, N)
	closes := make([]float64, N)
	volumes := make([]float64, N)
	for i, k := range klines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
		volumes[i] = k.Volume
	}

	atrInd := indicators.NewATRTVStandard(g.cfg.ATRPeriod)
	g.atr = atrInd.Calculate(highs, lows, closes)

	// VWMA fast/slow
	if g.cfg.VWMAFastPeriod > 0 {
		vwf := indicators.NewVWMATVStandard(g.cfg.VWMAFastPeriod)
		g.vwmaFast = vwf.Calculate(closes, volumes)
	}
	if g.cfg.VWMASlowPeriod > 0 {
		vws := indicators.NewVWMATVStandard(g.cfg.VWMASlowPeriod)
		g.vwmaSlow = vws.Calculate(closes, volumes)
	}

	stoch := indicators.NewStochTVStandard(g.cfg.StochKPeriod, g.cfg.StochKSmooth, g.cfg.StochDPeriod)
	kVals, dVals := stoch.Calculate(highs, lows, closes)
	g.stochK, g.stochD = kVals, dVals

	mfi := indicators.NewMFITVStandard(g.cfg.MFIPeriod)
	g.mfi = mfi.Calculate(highs, lows, closes, volumes)

	cci := indicators.NewCCITVStandard(g.cfg.CCIPeriod)
	g.cci = cci.Calculate(highs, lows, closes)

	volSMA := indicators.NewSMATVStandard(g.cfg.VolumeSMAPeriod)
	g.volSMA = volSMA.Calculate(volumes)
	if g.cfg.MACDFastPeriod > 0 && g.cfg.MACDSlowPeriod > 0 && g.cfg.MACDSignalPeriod > 0 {
		macdInd := indicators.NewMACDTVStandard(g.cfg.MACDFastPeriod, g.cfg.MACDSlowPeriod, g.cfg.MACDSignalPeriod)
		macdLine, macdSignal, macdHist := macdInd.Calculate(closes)
		g.macdLine = macdLine
		g.macdSignal = macdSignal
		g.macdHist = macdHist
	}
	return nil
}

func (g *Generator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var out []signals.Signal
	if len(klines) < 3 {
		return out, nil
	}
	lastClosedIdx := len(klines) - 2
	startIdx := g.lastProcessedIdx + 1
	warmup := max5(g.cfg.ATRPeriod, g.cfg.StochKPeriod+g.cfg.StochKSmooth+g.cfg.StochDPeriod, g.cfg.MFIPeriod, g.cfg.CCIPeriod, g.cfg.VolumeSMAPeriod)
	if g.cfg.Aggregate3 && warmup < 2 {
		warmup = 2
	}
	if startIdx < warmup {
		startIdx = warmup
	}
	if startIdx > lastClosedIdx {
		return out, nil
	}

	for i := startIdx; i <= lastClosedIdx; i++ {
		if i >= len(g.atr) || math.IsNaN(g.atr[i]) {
			continue
		}
		if i >= len(g.volSMA) || math.IsNaN(g.volSMA[i]) {
			continue
		}
		// Stoch n'est requis que si le cross ou les bornes K sont utilisés.
		stochExtGatesEnabled := g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
			g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate
		needStoch := g.cfg.EnableStochCross || stochExtGatesEnabled
		needMFI := g.cfg.EnableMFIGate
		needCCI := g.cfg.EnableCCIGate
		needVWMA := g.cfg.EnableVWMATrendGate || g.cfg.EnableVWMACross
		needMACD := g.cfg.EnableMACDCrossGate || g.cfg.EnableMACDSignGate || g.cfg.EnableMACDHistGate
		// Si aucun gate n'est actif, ne pas produire de signaux pour cette bougie.
		hasAnyGate := g.cfg.EnableBarGate || g.cfg.EnableStochCross || stochExtGatesEnabled || needMFI || needCCI || needVWMA || needMACD
		if !hasAnyGate {
			continue
		}
		if needStoch {
			if i >= len(g.stochK) || math.IsNaN(g.stochK[i]) {
				continue
			}
			if g.cfg.EnableStochCross {
				if i >= len(g.stochD) || math.IsNaN(g.stochD[i]) {
					continue
				}
			}
		}
		if needMFI {
			if i >= len(g.mfi) || math.IsNaN(g.mfi[i]) {
				continue
			}
		}
		if needCCI {
			if i >= len(g.cci) || math.IsNaN(g.cci[i]) {
				continue
			}
		}
		if needVWMA {
			if i >= len(g.vwmaFast) || math.IsNaN(g.vwmaFast[i]) {
				continue
			}
			if i >= len(g.vwmaSlow) || math.IsNaN(g.vwmaSlow[i]) {
				continue
			}
		}
		if needMACD {
			if i >= len(g.macdLine) || math.IsNaN(g.macdLine[i]) {
				continue
			}
			if i >= len(g.macdSignal) || math.IsNaN(g.macdSignal[i]) {
				continue
			}
			if i >= len(g.macdHist) || math.IsNaN(g.macdHist[i]) {
				continue
			}
		}

		body, isGreen, isRed, volNorm := g.bodyVolumeColorAt(klines, i)
		b2atr := body / g.atr[i]
		v2sma := volNorm / g.volSMA[i]

		// Trend via optional VWMA
		trendOKLong := true
		trendOKShort := true
		if g.cfg.EnableVWMATrendGate {
			trendOKLong = g.vwmaFast[i] > g.vwmaSlow[i]
			trendOKShort = g.vwmaFast[i] < g.vwmaSlow[i]
		}
		crossUp, crossDown := false, false
		if g.cfg.EnableStochCross {
			up, down := g.stochCrossAt(i)
			crossUp, crossDown = up, down
		}
		vwUp, vwDown := false, false
		if g.cfg.EnableVWMACross {
			vwUp, vwDown = g.vwmaCrossAt(i)
		}

		// Step1: body/ATR + volume/SMA + couleur comme filtre de base
		barOKLong, barOKShort := true, true
		if g.cfg.EnableBarGate {
			barOKLong = isGreen && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
			barOKShort = isRed && b2atr >= g.cfg.BodyATRMultiplier && v2sma >= g.cfg.VolumeCoeff
		}

		// CCI court: extrême inverse de la tendance (filtre optionnel via EnableCCIGate)
		cciOKLong, cciOKShort := true, true
		if g.cfg.EnableCCIGate {
			cciVal := g.cci[i]
			if g.cfg.EnableCCIOversoldLongGate {
				cciOKLong = cciOKLong && cciVal <= g.cfg.CCIOversoldLong
			}
			if g.cfg.EnableCCIOverboughtLongGate {
				cciOKLong = cciOKLong && cciVal <= g.cfg.CCIOverboughtLong
			}
			if g.cfg.EnableCCIOversoldShortGate {
				cciOKShort = cciOKShort && cciVal >= g.cfg.CCIOversoldShort
			}
			if g.cfg.EnableCCIOverboughtShortGate {
				cciOKShort = cciOKShort && cciVal >= g.cfg.CCIOverboughtShort
			}
		}

		// MFI: extrême inverse de la tendance (filtre optionnel via EnableMFIGate)
		mfiOKLong, mfiOKShort := true, true
		if g.cfg.EnableMFIGate {
			mfiVal := g.mfi[i]
			if g.cfg.EnableMFIOversoldLongGate {
				mfiOKLong = mfiOKLong && mfiVal <= g.cfg.MFIOversoldLong
			}
			if g.cfg.EnableMFIOverboughtLongGate {
				mfiOKLong = mfiOKLong && mfiVal <= g.cfg.MFIOverboughtLong
			}
			if g.cfg.EnableMFIOversoldShortGate {
				mfiOKShort = mfiOKShort && mfiVal >= g.cfg.MFIOversoldShort
			}
			if g.cfg.EnableMFIOverboughtShortGate {
				mfiOKShort = mfiOKShort && mfiVal >= g.cfg.MFIOverboughtShort
			}
		}

		// Cross Stoch obligatoire si EnableStochCross, sinon pas de filtre de croisement
		longCrossOK := true
		shortCrossOK := true
		if g.cfg.EnableStochCross {
			longCrossOK = crossUp
			shortCrossOK = crossDown
		}
		stochKOKLong := true
		stochKOKShort := true
		if g.cfg.EnableStochKOversoldLongGate || g.cfg.EnableStochKOverboughtLongGate ||
			g.cfg.EnableStochKOversoldShortGate || g.cfg.EnableStochKOverboughtShortGate {
			kVal := g.stochK[i]
			if g.cfg.EnableStochKOversoldLongGate {
				stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOversoldLong
			}
			if g.cfg.EnableStochKOverboughtLongGate {
				stochKOKLong = stochKOKLong && kVal <= g.cfg.StochKOverboughtLong
			}
			if g.cfg.EnableStochKOversoldShortGate {
				stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOversoldShort
			}
			if g.cfg.EnableStochKOverboughtShortGate {
				stochKOKShort = stochKOKShort && kVal >= g.cfg.StochKOverboughtShort
			}
		}
		vwCrossOKLong := true
		vwCrossOKShort := true
		if g.cfg.EnableVWMACross {
			vwCrossOKLong = vwUp
			vwCrossOKShort = vwDown
		}
		macdCrossOKLong := true
		macdCrossOKShort := true
		if g.cfg.EnableMACDCrossGate {
			up, down := g.macdCrossAt(i)
			macdCrossOKLong = up
			macdCrossOKShort = down
		}
		macdSignOKLong := true
		macdSignOKShort := true
		if g.cfg.EnableMACDSignGate {
			m := g.macdLine[i]
			s := g.macdSignal[i]
			macdSignOKLong = m < 0 && s < 0
			macdSignOKShort = m > 0 && s > 0
		}
		macdHistOKLong := true
		macdHistOKShort := true
		if g.cfg.EnableMACDHistGate {
			h := g.macdHist[i]
			macdHistOKLong = h > 0
			macdHistOKShort = h < 0
		}

		// Signal directionnel LONG
		if g.cfg.EnableOpenLong && trendOKLong && longCrossOK && stochKOKLong && vwCrossOKLong && macdCrossOKLong && macdSignOKLong && macdHistOKLong && barOKLong && cciOKLong && mfiOKLong {
			conf := confidence(b2atr, v2sma)
			meta := map[string]interface{}{
				"generator":   "ban_fin_momentium",
				"agg3":        g.cfg.Aggregate3,
				"body":        body,
				"atr10":       g.atr[i],
				"body_to_atr": b2atr,
				"volume":      volNorm,
				"vol_sma":     g.volSMA[i],
				"vol_to_sma":  v2sma,
				"stoch_k":     g.stochK[i],
				"stoch_d":     g.stochD[i],
				"cci":         g.cci[i],
			}
			addIfFinite(meta, "mfi", g.mfi[i])
			out = append(out, signals.Signal{
				Timestamp:  klines[i].OpenTime,
				Action:     signals.SignalActionEntry,
				Type:       signals.SignalTypeLong,
				Price:      klines[i].Close,
				Confidence: conf,
				Metadata:   meta,
			})
		}

		// Signal directionnel SHORT
		if g.cfg.EnableOpenShort && trendOKShort && shortCrossOK && stochKOKShort && vwCrossOKShort && macdCrossOKShort && macdSignOKShort && macdHistOKShort && barOKShort && cciOKShort && mfiOKShort {
			conf := confidence(b2atr, v2sma)
			meta := map[string]interface{}{
				"generator":   "ban_fin_momentium",
				"agg3":        g.cfg.Aggregate3,
				"body":        body,
				"atr10":       g.atr[i],
				"body_to_atr": b2atr,
				"volume":      volNorm,
				"vol_sma":     g.volSMA[i],
				"vol_to_sma":  v2sma,
				"stoch_k":     g.stochK[i],
				"stoch_d":     g.stochD[i],
				"cci":         g.cci[i],
			}
			addIfFinite(meta, "mfi", g.mfi[i])
			out = append(out, signals.Signal{
				Timestamp:  klines[i].OpenTime,
				Action:     signals.SignalActionEntry,
				Type:       signals.SignalTypeShort,
				Price:      klines[i].Close,
				Confidence: conf,
				Metadata:   meta,
			})
		}
	}

	if len(out) > 0 {
		g.metrics.TotalSignals += len(out)
		entry, exit, l, s := 0, 0, 0, 0
		acc := 0.0
		for _, sig := range out {
			if sig.Action == signals.SignalActionEntry {
				entry++
			} else {
				exit++
			}
			if sig.Type == signals.SignalTypeLong {
				l++
			} else {
				s++
			}
			acc += sig.Confidence
		}
		g.metrics.EntrySignals += entry
		g.metrics.ExitSignals += exit
		g.metrics.LongSignals += l
		g.metrics.ShortSignals += s
		g.metrics.AvgConfidence = acc / float64(len(out))
		g.metrics.LastSignalTime = out[len(out)-1].Timestamp
	}
	g.lastProcessedIdx = lastClosedIdx
	return out, nil
}

func (g *Generator) GetMetrics() signals.GeneratorMetrics { return g.metrics }

func (g *Generator) bodyVolumeColorAt(klines []signals.Kline, i int) (body float64, isGreen, isRed bool, volNorm float64) {
	if g.cfg.Aggregate3 {
		if i < 2 {
			return 0, false, false, 0
		}
		o := klines[i-2].Open
		c := klines[i].Close
		body = math.Abs(c - o)
		isGreen = c > o
		isRed = c < o
		vol := klines[i-2].Volume + klines[i-1].Volume + klines[i].Volume
		volNorm = vol / 3.0
		return
	}
	o := klines[i].Open
	c := klines[i].Close
	body = math.Abs(c - o)
	isGreen = c > o
	isRed = c < o
	volNorm = klines[i].Volume
	return
}

func (g *Generator) stochCrossAt(i int) (up, down bool) {
	if i-1 < 0 || i >= len(g.stochD) {
		return false, false
	}
	pk, pd := g.stochK[i-1], g.stochD[i-1]
	ck, cd := g.stochK[i], g.stochD[i]
	if math.IsNaN(pk) || math.IsNaN(pd) || math.IsNaN(ck) || math.IsNaN(cd) {
		return false, false
	}
	up = pk < pd && ck > cd
	down = pk > pd && ck < cd
	return
}

func (g *Generator) vwmaCrossAt(i int) (up, down bool) {
	if i-1 < 0 || i >= len(g.vwmaFast) || i >= len(g.vwmaSlow) {
		return false, false
	}
	pf, ps := g.vwmaFast[i-1], g.vwmaSlow[i-1]
	cf, cs := g.vwmaFast[i], g.vwmaSlow[i]
	if math.IsNaN(pf) || math.IsNaN(ps) || math.IsNaN(cf) || math.IsNaN(cs) {
		return false, false
	}
	up = pf <= ps && cf > cs
	down = pf >= ps && cf < cs
	return
}

func (g *Generator) macdCrossAt(i int) (up, down bool) {
	if i-1 < 0 || i >= len(g.macdLine) || i >= len(g.macdSignal) {
		return false, false
	}
	pm, ps := g.macdLine[i-1], g.macdSignal[i-1]
	cm, cs := g.macdLine[i], g.macdSignal[i]
	if math.IsNaN(pm) || math.IsNaN(ps) || math.IsNaN(cm) || math.IsNaN(cs) {
		return false, false
	}
	up = pm <= ps && cm > cs
	down = pm >= ps && cm < cs
	return
}

func max5(a, b, c, d, e int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	if e > m {
		m = e
	}
	return m
}

func confidence(b2atr, v2sma float64) float64 {
	c := 0.5
	if b2atr >= 1.0 {
		c += 0.2
	}
	if v2sma >= 1.2 {
		c += 0.15
	}
	if c > 0.95 {
		c = 0.95
	}
	if c < 0.4 {
		c = 0.4
	}
	return c
}

func safeAt(arr []float64, i int) float64 {
	if i >= 0 && i < len(arr) {
		if !math.IsNaN(arr[i]) {
			return arr[i]
		}
	}
	return math.NaN()
}

func addIfFinite(m map[string]interface{}, key string, v float64) {
	if !math.IsNaN(v) && !math.IsInf(v, 0) {
		m[key] = v
	}
}
