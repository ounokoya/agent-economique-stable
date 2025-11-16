package ban_fin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

// Config holds BAN_FIN generator parameters (defaults aligned with SPEC_BAN_FIN.md)
type Config struct {
	VWMAShortPeriod int
	VWMALongPeriod  int
	DMIPeriod       int
	DMISmooth       int // kept for compatibility; DMITVStandard uses one period
	WindowMatching  int

	ATRPeriod        int
	GapATRMultiplier float64
	GapBasis         string // "vwma_spread" | "price_vs_vwma_short" | "price_vs_vwma_long"
	EnableGapGating  bool

	EnableSlopeVWMAShort bool
	EnableSlopeVWMALong  bool
	SlopeVWMAShortMin    float64
	SlopeVWMALongMin     float64
	SlopeBasisVWMA       string // "delta_1_bougie" | "spread_vwma"

	EnableSlopeDX bool
	EnableSlopeADX bool
	SlopeDXMin     float64
	SlopeADXMin    float64

	CCIPeriod         int
	EnableCCIExtremes bool
	CCIOverbought     float64
	CCIOversold       float64
	EnableSlopeCCI    bool
	SlopeCCIMin       float64

	EnableDXADXSpread            bool
	DXADXSpreadMin               float64
	DXADXRequiredDirectionalCross bool // true => LONG needs UP, SHORT needs DOWN

	EdgeTrigger bool // emit only on first valid bar after last VWMA cross
}

func DefaultConfig() Config {
	return Config{
		VWMAShortPeriod: 10,
		VWMALongPeriod:  20,
		DMIPeriod:       14,
		DMISmooth:       14,
		WindowMatching:  5,
		ATRPeriod:       3,
		GapATRMultiplier: 1.0,
		GapBasis:        "vwma_spread",
		EnableGapGating: true,
		EnableSlopeVWMAShort: true,
		EnableSlopeVWMALong:  true,
		SlopeVWMAShortMin:    0.0,
		SlopeVWMALongMin:     0.0,
		SlopeBasisVWMA:       "delta_1_bougie",
		EnableSlopeDX:        false,
		EnableSlopeADX:       false,
		SlopeDXMin:           0.0,
		SlopeADXMin:          0.0,
		CCIPeriod:            20,
		EnableCCIExtremes:    true,
		CCIOverbought:        100,
		CCIOversold:          -100,
		EnableSlopeCCI:       false,
		SlopeCCIMin:          0.0,
		EnableDXADXSpread:    false,
		DXADXSpreadMin:       0.0,
		DXADXRequiredDirectionalCross: true,
		EdgeTrigger: true,
	}
}

// Generator evaluates signal presence on the last closed candle
type Generator struct {
	cfg Config
}

func NewGenerator(cfg Config) *Generator { return &Generator{cfg: cfg} }

// EvaluateLast decides if there is a BAN_FIN signal on the last candle of klines.
// It returns a signals.Signal when i* is the FIRST valid bar after last VWMA cross (edge-trigger), otherwise nil.
func (g *Generator) EvaluateLast(klines []signals.Kline) (*signals.Signal, error) {
	if len(klines) < max3(g.cfg.VWMALongPeriod, g.cfg.DMIPeriod, g.cfg.ATRPeriod)+2 {
		return nil, fmt.Errorf("insufficient klines: %d", len(klines))
	}
	ks := ensureSortedUnique(klines)
	if len(ks) < 2 {
		return nil, errors.New("need at least 2 closed candles")
	}
	last := len(ks) - 1 // assume all are closed

	// Build arrays
	high := make([]float64, len(ks))
	low := make([]float64, len(ks))
	closeArr := make([]float64, len(ks))
	vol := make([]float64, len(ks))
	for i, k := range ks {
		high[i] = k.High
		low[i] = k.Low
		closeArr[i] = k.Close
		vol[i] = k.Volume
	}

	// VWMA short/long
	vwmaS := indicators.NewVWMATVStandard(g.cfg.VWMAShortPeriod).Calculate(closeArr, vol)
	vwmaL := indicators.NewVWMATVStandard(g.cfg.VWMALongPeriod).Calculate(closeArr, vol)

	// DMI: +DI, -DI, ADX ; DX separately
	dmi := indicators.NewDMITVStandard(g.cfg.DMIPeriod)
	plusDI, minusDI, adx := dmi.Calculate(high, low, closeArr)
	dx := dmi.CalculateDX(plusDI, minusDI)

	// ATR(3)
	atr := indicators.NewATRTVStandard(g.cfg.ATRPeriod).Calculate(high, low, closeArr)

	// CCI
	cciVals := indicators.NewCCITVStandard(g.cfg.CCIPeriod).Calculate(high, low, closeArr)

	W := g.cfg.WindowMatching
	winStart := last - W + 1
	if winStart < 1 { // need prev bar for crosses
		winStart = 1
	}

	// 1) Find last VWMA cross in W
	hasVwmaCross, vwmaCrossIndex, vwmaUp := lastVwmaCross(vwmaS, vwmaL, winStart, last)
	if !hasVwmaCross {
		return nil, nil // no signal
	}
	targetLong := vwmaUp

	// 2) Bases at candidate i* (per-candle evaluation uses sliding window semantics)
	if !diAlignedAt(plusDI[last], minusDI[last], targetLong) {
		return nil, nil
	}
	// Last DX/ADX cross in W must be directional and coherent
	hasDxCross, _, dxUp := lastDxAdxCross(dx, adx, winStart, last)
	if !hasDxCross {
		return nil, nil
	}
	if g.cfg.DXADXRequiredDirectionalCross {
		if targetLong && !dxUp { // LONG requires UP
			return nil, nil
		}
		if !targetLong && dxUp { // SHORT requires DOWN
			return nil, nil
		}
	}

	// 3) Gating at i*
	if !g.gatingOKAt(last, targetLong, vwmaS, vwmaL, dx, adx, cciVals, atr, closeArr) {
		return nil, nil
	}

	// 4) Edge-trigger: j0 = first index in [vwmaCrossIndex .. last] satisfying bases+gating with its own sliding window
	j0 := -1
	for j := vwmaCrossIndex; j <= last; j++ {
		// window for j
		ws := j - W + 1
		if ws < 1 { ws = 1 }
		// bases for j
		if !diAlignedAt(plusDI[j], minusDI[j], targetLong) {
			continue
		}
		hasDxJ, _, dxUpJ := lastDxAdxCross(dx, adx, ws, j)
		if !hasDxJ { continue }
		if g.cfg.DXADXRequiredDirectionalCross {
			if targetLong && !dxUpJ { continue }
			if !targetLong && dxUpJ { continue }
		}
		// gating at j
		if !g.gatingOKAt(j, targetLong, vwmaS, vwmaL, dx, adx, cciVals, atr, closeArr) {
			continue
		}
		j0 = j
		break
	}
	if j0 != last {
		return nil, nil
	}

	// Build signal
	var sigType signals.SignalType
	if targetLong { sigType = signals.SignalTypeLong } else { sigType = signals.SignalTypeShort }
	meta := map[string]interface{}{
		"generator": "ban_fin",
		"mode":      "TREND",
		"vwma_cross_index": vwmaCrossIndex,
		"window_matching":  W,
		"dxadx_required_direction": map[string]bool{"UP_for_LONG": g.cfg.DXADXRequiredDirectionalCross},
	}
	// enrich current values
	meta["vwma_short"] = vwmaS[last]
	meta["vwma_long"] = vwmaL[last]
	meta["di_plus"] = plusDI[last]
	meta["di_minus"] = minusDI[last]
	meta["dx"] = dx[last]
	meta["adx"] = adx[last]
	meta["atr3"] = atr[last]
	meta["cci"] = cciVals[last]

	// Confidence simple heuristic
	confidence := 0.8
	// Encode gating snapshot for traceability
	if snap, _ := json.Marshal(g.gatingSnapshot(last, targetLong, vwmaS, vwmaL, dx, adx, cciVals, atr, closeArr)); len(snap) > 0 {
		meta["gating_snapshot"] = string(snap)
	}

	return &signals.Signal{
		Timestamp: ks[last].OpenTime,
		Action:    signals.SignalActionEntry, // contextless demo; engine decides ENTRY/EXIT
		Type:      sigType,
		Price:     ks[last].Close,
		Confidence: confidence,
		Metadata:  meta,
	}, nil
}

func (g *Generator) gatingOKAt(i int, targetLong bool, vwmaS, vwmaL, dx, adx, cci, atr, closeArr []float64) bool {
	// Slopes
	if g.cfg.EnableSlopeVWMAShort {
		d := vwmaS[i] - vwmaS[i-1]
		if targetLong && d < g.cfg.SlopeVWMAShortMin { return false }
		if !targetLong && -d < g.cfg.SlopeVWMAShortMin { return false }
	}
	if g.cfg.EnableSlopeVWMALong {
		d := vwmaL[i] - vwmaL[i-1]
		if targetLong && d < g.cfg.SlopeVWMALongMin { return false }
		if !targetLong && -d < g.cfg.SlopeVWMALongMin { return false }
	}
	if g.cfg.EnableSlopeDX {
		d := dx[i] - dx[i-1]
		if targetLong && d < g.cfg.SlopeDXMin { return false }
		if !targetLong && -d < g.cfg.SlopeDXMin { return false }
	}
	if g.cfg.EnableSlopeADX {
		d := adx[i] - adx[i-1]
		if targetLong && d < g.cfg.SlopeADXMin { return false }
		if !targetLong && -d < g.cfg.SlopeADXMin { return false }
	}
	if g.cfg.EnableSlopeCCI {
		d := cci[i] - cci[i-1]
		if targetLong && d < g.cfg.SlopeCCIMin { return false }
		if !targetLong && -d < g.cfg.SlopeCCIMin { return false }
	}
	// CCI extremes
	if g.cfg.EnableCCIExtremes {
		if targetLong && cci[i] >= g.cfg.CCIOverbought { return false }
		if !targetLong && cci[i] <= g.cfg.CCIOversold { return false }
	}
	// Gap vs ATR
	if g.cfg.EnableGapGating {
		gap := 0.0
		switch g.cfg.GapBasis {
		case "price_vs_vwma_short":
			gap = math.Abs(closeArr[i] - vwmaS[i])
		case "price_vs_vwma_long":
			gap = math.Abs(closeArr[i] - vwmaL[i])
		default:
			gap = math.Abs(vwmaS[i] - vwmaL[i])
		}
		if atr[i] == 0 || math.IsNaN(atr[i]) { return false }
		if gap < g.cfg.GapATRMultiplier*atr[i] { return false }
	}
	// DX/ADX dominance & spread
	if g.cfg.EnableDXADXSpread {
		sp := math.Abs(dx[i]-adx[i])
		if sp < g.cfg.DXADXSpreadMin { return false }
		if g.cfg.DXADXRequiredDirectionalCross {
			if targetLong && !(dx[i] > adx[i]) { return false }
			if !targetLong && !(adx[i] > dx[i]) { return false }
		}
	}
	return true
}

func (g *Generator) gatingSnapshot(i int, targetLong bool, vwmaS, vwmaL, dx, adx, cci, atr, closeArr []float64) map[string]interface{} {
	return map[string]interface{}{
		"i": i,
		"target_long": targetLong,
		"slope_vwma_s": vwmaS[i] - vwmaS[i-1],
		"slope_vwma_l": vwmaL[i] - vwmaL[i-1],
		"slope_dx":     dx[i] - dx[i-1],
		"slope_adx":    adx[i] - adx[i-1],
		"slope_cci":    cci[i] - cci[i-1],
		"cci":          cci[i],
		"atr3":         atr[i],
		"gap_spread_vwma": math.Abs(vwmaS[i]-vwmaL[i]),
		"gap_price_vwma_s": math.Abs(closeArr[i]-vwmaS[i]),
		"gap_price_vwma_l": math.Abs(closeArr[i]-vwmaL[i]),
	}
}

func lastVwmaCross(vwmaS, vwmaL []float64, start, end int) (bool, int, bool) {
	found := false
	idx := -1
	up := false
	for i := start; i <= end; i++ {
		if i <= 0 { continue }
		prevUp := vwmaS[i-1] > vwmaL[i-1]
		currUp := vwmaS[i] > vwmaL[i]
		if !prevUp && currUp {
			found, idx, up = true, i, true
		}
		if prevUp && !currUp {
			found, idx, up = true, i, false
		}
	}
	return found, idx, up
}

func lastDxAdxCross(dx, adx []float64, start, end int) (bool, int, bool) {
	found := false
	idx := -1
	up := false
	for i := start; i <= end; i++ {
		if i <= 0 { continue }
		prevDXAbove := dx[i-1] > adx[i-1]
		currDXAbove := dx[i] > adx[i]
		if !prevDXAbove && currDXAbove {
			found, idx, up = true, i, true
		}
		if prevDXAbove && !currDXAbove {
			found, idx, up = true, i, false
		}
	}
	return found, idx, up
}

func diAlignedAt(plus, minus float64, targetLong bool) bool {
	if targetLong {
		return plus > minus
	}
	return minus > plus
}

func ensureSortedUnique(in []signals.Kline) []signals.Kline {
	cp := make([]signals.Kline, len(in))
	copy(cp, in)
	sort.Slice(cp, func(i, j int) bool { return cp[i].OpenTime.Before(cp[j].OpenTime) })
	// dedup by OpenTime
	out := make([]signals.Kline, 0, len(cp))
	var lastTs time.Time
	for _, k := range cp {
		if k.OpenTime.Equal(lastTs) { continue }
		out = append(out, k)
		lastTs = k.OpenTime
	}
	return out
}

func max3(a, b, c int) int {
	m := a
	if b > m { m = b }
	if c > m { m = c }
	return m
}
