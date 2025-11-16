package dmiresp

import (
	"math"
)

type Side string

const (
	SideLong  Side = "LONG"
	SideShort Side = "SHORT"
)

type Config struct {
	// Numeric gating
	UseMinDX bool
	MinDX    float64
	UseMinADX bool
	MinADX    float64
	UseMaxADX bool
	MaxADX    float64

	// Respiration gap threshold (seuil_resp)
	RespGap float64

	// Confirmation at i+1 if needed (case C)
	UseConfirm1 bool

	// Detection settings for respiration
	RequireDXDown bool
	RequireADXDown bool
	Eps           float64 // tolerance for "sur" equality at i-1

	// Logging level hint (0..3). Module itself does not print.
	LogLevel int
}

type Inputs struct {
	Index               int
	DIpPrev, DImPrev    float64
	DXPrev, ADXPrev     float64
	DIp, DIm, DX, ADX   float64
}

type Diagnostics struct {
	DIp, DIm, DX, ADX float64
	GapDXADX          float64
	Applied           struct {
		MinDX  bool
		MinADX bool
		MaxADX bool
	}
	Checks struct {
		DXDown      bool
		ADXDown     bool
		PrevOnADX   bool
		CurrBelow   bool
		ContextLong bool
		ContextShort bool
	}
}

type Event struct {
	Triggered bool
	Side      Side
	Lag       int // 0 for immediate, 1 for confirm at i+1
	Index     int // validation index
	CrossAt   int // respiration cross index
	Diag      Diagnostics
}

type pending struct {
	side       Side
	crossIndex int
	recheckAt  int // only i+1 supported
	lag        int // 1
}

type DmiResp struct {
	cfg      Config
	pendings []pending
}

func New(cfg Config) *DmiResp {
	return &DmiResp{cfg: cfg, pendings: make([]pending, 0, 4)}
}

func (d *DmiResp) Reset() { d.pendings = d.pendings[:0] }

func (d *DmiResp) Step(in Inputs) (Event, bool) {
	var out Event
	cfg := d.cfg

	// 1) Confirmation pendings at this bar (only i+1 supported)
	if len(d.pendings) > 0 {
		validatedCross := -1
		newPendings := d.pendings[:0]
		for _, p := range d.pendings {
			if p.recheckAt == in.Index {
				// Confirmation rule: DX > DI_inf and DX - DI_inf <= RespGap, plus numeric gating
				diSup, diInf := diSupInf(p.side, in.DIp, in.DIm)
				_ = diSup // not used directly in confirm
				if confirmOK(cfg, in.DX, in.ADX, diInf) {
					if !out.Triggered {
						out = Event{Triggered: true, Side: p.side, Lag: p.lag, Index: in.Index, CrossAt: p.crossIndex, Diag: diagFor(cfg, p.side, in)}
						validatedCross = p.crossIndex
					}
					continue // consume
				}
				// consume even if not validated; do not re-add
				continue
			}
			if validatedCross != -1 && p.crossIndex == validatedCross {
				continue
			}
			newPendings = append(newPendings, p)
		}
		d.pendings = newPendings
	}

	// 2) Detect respiration cross at this bar
	if in.Index > 0 {
		ctxLong := in.DIp > in.DIm
		ctxShort := in.DIm > in.DIp
		dxDown := !cfg.RequireDXDown || (in.DX < in.DXPrev)
		adxDown := !cfg.RequireADXDown || (in.ADX < in.ADXPrev)
		prevOn := math.Abs(in.DXPrev-in.ADXPrev) <= cfg.Eps
		currBelow := in.DX < in.ADX

		if (ctxLong || ctxShort) && dxDown && adxDown && prevOn && currBelow {
			var side Side
			if ctxLong { side = SideLong } else { side = SideShort }
			_, diInf := diSupInf(side, in.DIp, in.DIm)

			// Immediate validation cases A/B
			caseA := in.DX <= diInf
			caseB := in.DX > diInf && (in.DX-diInf) <= cfg.RespGap
			if (caseA || caseB) && numericOK(cfg, in.DX, in.ADX) && !out.Triggered {
				out = Event{Triggered: true, Side: side, Lag: 0, Index: in.Index, CrossAt: in.Index, Diag: diagFor(cfg, side, in)}
			} else if cfg.UseConfirm1 {
				// Schedule confirm at i+1
				d.pendings = append(d.pendings, pending{side: side, crossIndex: in.Index, recheckAt: in.Index + 1, lag: 1})
			}
		}
	}

	return out, out.Triggered
}

func diSupInf(side Side, DIp, DIm float64) (diSup, diInf float64) {
	if side == SideLong { return DIp, DIm }
	return DIm, DIp
}

func numericOK(cfg Config, DX, ADX float64) bool {
	if cfg.UseMinDX && !(DX >= cfg.MinDX) { return false }
	if cfg.UseMinADX && !(ADX >= cfg.MinADX) { return false }
	if cfg.UseMaxADX && !(ADX <= cfg.MaxADX) { return false }
	return true
}

func confirmOK(cfg Config, DX, ADX, diInf float64) bool {
	if !(DX > diInf) { return false }
	if (DX - diInf) > cfg.RespGap { return false }
	return numericOK(cfg, DX, ADX)
}

func diagFor(cfg Config, side Side, in Inputs) Diagnostics {
	d := Diagnostics{
		DIp: in.DIp, DIm: in.DIm, DX: in.DX, ADX: in.ADX, GapDXADX: math.Abs(in.DX-in.ADX),
	}
	d.Applied.MinDX = cfg.UseMinDX
	d.Applied.MinADX = cfg.UseMinADX
	d.Applied.MaxADX = cfg.UseMaxADX
	d.Checks.DXDown = cfg.RequireDXDown && (in.DX < in.DXPrev)
	d.Checks.ADXDown = cfg.RequireADXDown && (in.ADX < in.ADXPrev)
	d.Checks.PrevOnADX = math.Abs(in.DXPrev-in.ADXPrev) <= cfg.Eps
	d.Checks.CurrBelow = in.DX < in.ADX
	d.Checks.ContextLong = side == SideLong
	d.Checks.ContextShort = side == SideShort
	return d
}
