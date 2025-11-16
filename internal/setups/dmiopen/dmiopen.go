package dmiopen

import (
	"math"
)

// Side represents orientation for the setup
type Side string

const (
	SideLong  Side = "LONG"
	SideShort Side = "SHORT"
)

// Config controls gating rules and recheck plan
type Config struct {
	// Relative gating
	UseRelAdxUnderDiSup bool
	UseRelDxOverDiInf   bool

	// Numeric gating
	UseMinDX       bool
	MinDX          float64
	UseMinADX      bool
	MinADX         float64
	UseMaxADX      bool
	MaxADX         float64
	UseMinGapDXADX bool
	MinGapDXADX    float64

	// Rechecks (independent lags)
	UseRecheck1 bool
	UseRecheck2 bool
	UseRecheck3 bool
	UseRecheck4 bool
	UseRecheck5 bool
	UseRecheck6 bool

	// Recheck integrity requirements
	// If enabled, during recheck the DI orientation must still match the side
	// (LONG: DIp>DIm, SHORT: DIm>DIp)
	RecheckRequireContextSide   bool
	// If enabled, during recheck DX must remain above ADX
	RecheckRequireDXAboveADX    bool
	// If enabled, any flip of the above constraints cancels all remaining
	// rechecks for the same cross (siblings are purged)
	RecheckCancelSiblingsOnFlip bool

	// Detection settings
	RequireDXUp  bool
	RequireADXUp bool
	Eps          float64

	// Logging level hint (0..3). Module itself does not print.
	LogLevel int
}

// Inputs carries indicator values at i-1 and i
type Inputs struct {
	Index               int
	DIpPrev, DImPrev    float64
	DXPrev, ADXPrev     float64
	DIp, DIm, DX, ADX   float64
}

// Diagnostics provides values and applied checks for logging at caller level
type Diagnostics struct {
	DIp, DIm, DX, ADX float64
	GapDXADX          float64
	Applied           struct {
		RelAdxUnderDiSup bool
		RelDxOverDiInf   bool
		MinDX            bool
		MinADX           bool
		MaxADX           bool
		MinGapDXADX      bool
	}
	Checks struct {
		DXUp       bool
		ADXUp      bool
		CrossEdge  bool
		ContextLong  bool
		ContextShort bool
	}
}

// Event is emitted when a validation occurs
type Event struct {
	Triggered bool
	Side      Side
	Lag       int // 0 for immediate, 1..6 for rechecks
	Index     int // validation index
	CrossAt   int // crossover index
	Diag      Diagnostics
}

type pending struct {
	side       Side
	crossIndex int
	recheckAt  int
	lag        int
}

// DmiOpen is the stateful detector
type DmiOpen struct {
	cfg      Config
	pendings []pending
}

func New(cfg Config) *DmiOpen {
	return &DmiOpen{cfg: cfg, pendings: make([]pending, 0, 8)}
}

func (d *DmiOpen) Reset() { d.pendings = d.pendings[:0] }

func (d *DmiOpen) Step(in Inputs) (Event, bool) {
	var out Event
	cfg := d.cfg

	// 1) Process rechecks scheduled for this bar
	if len(d.pendings) > 0 {
		validatedCross := -1
		cancelCrossIndex := -1
		newPendings := d.pendings[:0]
		for _, p := range d.pendings {
			if p.recheckAt == in.Index {
				// During recheck, optionally enforce context-side and DX>ADX
				contextOK := true
				if cfg.RecheckRequireContextSide {
					if p.side == SideLong {
						contextOK = in.DIp > in.DIm
					} else {
						contextOK = in.DIm > in.DIp
					}
				}
				dxAboveOK := true
				if cfg.RecheckRequireDXAboveADX {
					dxAboveOK = in.DX > in.ADX
				}
				if (!contextOK || !dxAboveOK) && cfg.RecheckCancelSiblingsOnFlip {
					// Cancel all future rechecks for this cross; consume current
					cancelCrossIndex = p.crossIndex
					continue
				}
				// Evaluate gating only (no new cross/pente checks)
				diSup, diInf := diSupInf(p.side, in.DIp, in.DIm)
				pass, diag := passesGating(cfg, p.side, in.DX, in.ADX, diSup, diInf, in.DIp, in.DIm)
				if pass && !out.Triggered {
					out = Event{Triggered: true, Side: p.side, Lag: p.lag, Index: in.Index, CrossAt: p.crossIndex, Diag: diag}
					validatedCross = p.crossIndex
					continue // consume this pending
				}
				// consume even if not validated; do not re-add
				continue
			}
			// If we validated a cross this bar, purge its siblings
			if (validatedCross != -1 && p.crossIndex == validatedCross) || (cancelCrossIndex != -1 && p.crossIndex == cancelCrossIndex) {
				continue
			}
			newPendings = append(newPendings, p)
		}
		d.pendings = newPendings
	}

	// 2) Detect edge-trigger cross at this bar
	if in.Index > 0 {
		cross := in.DXPrev <= in.ADXPrev+cfg.Eps && in.DX > in.ADX
		dxUp := !cfg.RequireDXUp || (in.DX > in.DXPrev)
		adxUp := !cfg.RequireADXUp || (in.ADX > in.ADXPrev)
		ctxLong := in.DIp > in.DIm
		ctxShort := in.DIm > in.DIp

		if cross && dxUp && adxUp && (ctxLong || ctxShort) {
			var side Side
			if ctxLong {
				side = SideLong
			} else {
				side = SideShort
			}
			diSup, diInf := diSupInf(side, in.DIp, in.DIm)
			pass, diag := passesGating(cfg, side, in.DX, in.ADX, diSup, diInf, in.DIp, in.DIm)
			if pass && !out.Triggered {
				out = Event{Triggered: true, Side: side, Lag: 0, Index: in.Index, CrossAt: in.Index, Diag: diag}
			} else {
				// schedule rechecks according to config
				lags := []struct {
					use bool
					lag int
				}{
					{cfg.UseRecheck1, 1}, {cfg.UseRecheck2, 2}, {cfg.UseRecheck3, 3},
					{cfg.UseRecheck4, 4}, {cfg.UseRecheck5, 5}, {cfg.UseRecheck6, 6},
				}
				for _, lg := range lags {
					if lg.use {
						d.pendings = append(d.pendings, pending{side: side, crossIndex: in.Index, recheckAt: in.Index + lg.lag, lag: lg.lag})
					}
				}
			}
		}
	}

	return out, out.Triggered
}

func diSupInf(side Side, DIp, DIm float64) (diSup, diInf float64) {
	if side == SideLong {
		return DIp, DIm
	}
	return DIm, DIp
}

func passesGating(cfg Config, side Side, DX, ADX, diSup, diInf, DIp, DIm float64) (bool, Diagnostics) {
	diag := Diagnostics{
		DX:       DX,
		ADX:      ADX,
		DIp:      DIp,
		DIm:      DIm,
		GapDXADX: math.Abs(DX - ADX),
	}
	// DI values are provided in caller; not stored in diag beyond gate references

	relAdx := !cfg.UseRelAdxUnderDiSup || ADX < diSup
	relDx := !cfg.UseRelDxOverDiInf || DX > diInf
	minDX := !cfg.UseMinDX || DX >= cfg.MinDX
	minADX := !cfg.UseMinADX || ADX >= cfg.MinADX
	maxADX := !cfg.UseMaxADX || ADX <= cfg.MaxADX
	gapOK := !cfg.UseMinGapDXADX || math.Abs(DX-ADX) >= cfg.MinGapDXADX

	diag.Applied.RelAdxUnderDiSup = cfg.UseRelAdxUnderDiSup
	diag.Applied.RelDxOverDiInf = cfg.UseRelDxOverDiInf
	diag.Applied.MinDX = cfg.UseMinDX
	diag.Applied.MinADX = cfg.UseMinADX
	diag.Applied.MaxADX = cfg.UseMaxADX
	diag.Applied.MinGapDXADX = cfg.UseMinGapDXADX

	ok := relAdx && relDx && minDX && minADX && maxADX && gapOK
	return ok, diag
}
