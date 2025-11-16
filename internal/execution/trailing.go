package execution

import "math"

type Side string

const (
	SideLong  Side = "LONG"
	SideShort Side = "SHORT"
)

type Trailing struct {
	Side   Side
	Entry  float64
	Offset float64
	Trail  float64
}

func NewTrailing(side Side, entryPrice float64, atrAtEntry float64, capPct float64) *Trailing {
	offset := atrAtEntry
	cap := entryPrice * capPct
	if offset > cap {
		offset = cap
	}
	t := &Trailing{Side: side, Entry: entryPrice, Offset: offset}
	if side == SideLong {
		t.Trail = entryPrice - offset
	} else {
		t.Trail = entryPrice + offset
	}
	return t
}

func (t *Trailing) Update(close float64) {
	if t.Side == SideLong {
		cand := close - t.Offset
		if cand > t.Trail {
			t.Trail = cand
		}
	} else {
		cand := close + t.Offset
		if cand < t.Trail {
			t.Trail = cand
		}
	}
}

func (t *Trailing) Hit(close float64) (bool, float64) {
	if t.Side == SideLong {
		if close <= t.Trail {
			return true, t.Trail
		}
	} else {
		if close >= t.Trail {
			return true, t.Trail
		}
	}
	return false, 0
}

// PercentTrailing implémente un trailing stop standard en pourcentage du prix.
// LONG : le stop suit le plus haut atteint en restant Pct en dessous.
// SHORT : le stop suit le plus bas atteint en restant Pct au-dessus.
type PercentTrailing struct {
	Side Side
	Pct  float64

	Trail float64
	max   float64 // plus haut atteint (LONG)
	min   float64 // plus bas atteint (SHORT)
	init  bool
}

// NewPercentTrailing crée un trailing stop en pourcentage autour du prix d'entrée.
// pct est un ratio (ex: 0.003 = 0.3%).
func NewPercentTrailing(side Side, entryPrice float64, pct float64) *PercentTrailing {
	tp := &PercentTrailing{Side: side, Pct: pct}
	if pct <= 0 || entryPrice <= 0 {
		return tp
	}
	if side == SideLong {
		tp.max = entryPrice
		tp.Trail = entryPrice * (1.0 - pct)
	} else {
		tp.min = entryPrice
		tp.Trail = entryPrice * (1.0 + pct)
	}
	tp.init = true
	return tp
}

// Update met à jour le trailing en fonction du dernier prix.
func (t *PercentTrailing) Update(price float64) {
	if !t.init || t.Pct <= 0 || price <= 0 {
		return
	}
	if t.Side == SideLong {
		if price > t.max {
			t.max = price
			cand := t.max * (1.0 - t.Pct)
			if cand > t.Trail {
				t.Trail = cand
			}
		}
	} else {
		if price < t.min || t.min == 0 {
			t.min = price
			cand := t.min * (1.0 + t.Pct)
			if t.Trail == 0 || cand < t.Trail {
				t.Trail = cand
			}
		}
	}
}

// Hit vérifie si la CLOSE a touché le trailing standard en %.
func (t *PercentTrailing) Hit(price float64) (bool, float64) {
	if !t.init || t.Pct <= 0 {
		return false, 0
	}
	if t.Side == SideLong {
		if price <= t.Trail {
			return true, t.Trail
		}
	} else {
		if price >= t.Trail {
			return true, t.Trail
		}
	}
	return false, 0
}

// VWMATrailing implémente un stop suiveur basé sur VWMA_fast
// Le stop suit la VWMA avec un offset proportionnel (OffsetPct),
// et ne se rapproche jamais du prix dans le mauvais sens.
type VWMATrailing struct {
	Side      Side
	OffsetPct float64
	Trail     float64
	init      bool
}

// NewVWMATrailing crée une instance de trailing VWMA avec un offset en pourcentage (ex: 0.002 = 0.2%)
func NewVWMATrailing(side Side, offsetPct float64) *VWMATrailing {
	return &VWMATrailing{Side: side, OffsetPct: offsetPct}
}

// Update met à jour le niveau de stop en fonction de la VWMA courante.
// Si vwma est NaN, la fonction ne fait rien.
func (t *VWMATrailing) Update(vwma float64) {
	if math.IsNaN(vwma) || t.OffsetPct <= 0 {
		return
	}
	if t.Side == SideLong {
		cand := vwma * (1.0 - t.OffsetPct)
		if !t.init || cand > t.Trail {
			t.Trail = cand
			t.init = true
		}
	} else {
		cand := vwma * (1.0 + t.OffsetPct)
		if !t.init || cand < t.Trail {
			t.Trail = cand
			t.init = true
		}
	}
}

// Hit vérifie si la CLOSE a touché le stop suiveur VWMA.
func (t *VWMATrailing) Hit(close float64) (bool, float64) {
	if !t.init {
		return false, 0
	}
	if t.Side == SideLong {
		if close <= t.Trail {
			return true, t.Trail
		}
	} else {
		if close >= t.Trail {
			return true, t.Trail
		}
	}
	return false, 0
}

type MultiStagePercentTrailing struct {
	Side  Side
	Entry float64

	Pct1 float64
	Pct2 float64
	Pct3 float64

	Thr2 float64
	Thr3 float64

	Trail float64
	max   float64
	min   float64
	init  bool
}

func NewMultiStagePercentTrailing(side Side, entryPrice float64, pct1, pct2, pct3, thr2, thr3 float64) *MultiStagePercentTrailing {
	t := &MultiStagePercentTrailing{
		Side:  side,
		Entry: entryPrice,
		Pct1:  pct1,
		Pct2:  pct2,
		Pct3:  pct3,
		Thr2:  thr2,
		Thr3:  thr3,
	}
	if entryPrice <= 0 || pct1 <= 0 {
		return t
	}
	if side == SideLong {
		t.max = entryPrice
		t.Trail = entryPrice * (1.0 - pct1)
	} else {
		t.min = entryPrice
		t.Trail = entryPrice * (1.0 + pct1)
	}
	t.init = true
	return t
}

func (t *MultiStagePercentTrailing) currentPct(profit float64) float64 {
	p := t.Pct1
	if t.Thr2 > 0 && profit >= t.Thr2 {
		p = t.Pct2
	}
	if t.Thr3 > 0 && profit >= t.Thr3 {
		p = t.Pct3
	}
	return p
}

func (t *MultiStagePercentTrailing) Update(price float64) {
	if !t.init || price <= 0 || t.Entry <= 0 {
		return
	}
	if t.Side == SideLong {
		if price > t.max {
			t.max = price
		}
		profit := (t.max - t.Entry) / t.Entry
		pct := t.currentPct(profit)
		if pct <= 0 {
			return
		}
		cand := t.max * (1.0 - pct)
		if cand > t.Trail {
			t.Trail = cand
		}
	} else {
		if price < t.min || t.min == 0 {
			t.min = price
		}
		profit := (t.Entry - t.min) / t.Entry
		pct := t.currentPct(profit)
		if pct <= 0 {
			return
		}
		cand := t.min * (1.0 + pct)
		if t.Trail == 0 || cand < t.Trail {
			t.Trail = cand
		}
	}
}

func (t *MultiStagePercentTrailing) Hit(price float64) (bool, float64) {
	if !t.init {
		return false, 0
	}
	if t.Side == SideLong {
		if price <= t.Trail {
			return true, t.Trail
		}
	} else {
		if price >= t.Trail {
			return true, t.Trail
		}
	}
	return false, 0
}
