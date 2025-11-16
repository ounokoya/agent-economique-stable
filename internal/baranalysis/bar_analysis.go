package baranalysis

import (
	"math"
	"time"
)

type BarAnalysis struct {
	Time              time.Time
	Open              float64
	High              float64
	Low               float64
	Close             float64
	Type              string
	Range             float64
	Body              float64
	NonBody           float64
	ATR               float64
	OccupationPercent float64
	BodyATRRatio      float64
	NonBodyATRRatio   float64
}

func AnalyzeBar(open, high, low, close, atr float64, t time.Time) BarAnalysis {
	rng := high - low
	if rng < 0 {
		rng = 0
	}
	body := math.Abs(close - open)
	nonBody := rng - body
	if nonBody < 0 {
		nonBody = 0
	}

	kind := "NEUTRE"
	if close > open {
		kind = "VERT"
	} else if close < open {
		kind = "ROUGE"
	}

	occ := 0.0
	if rng > 0 {
		occ = (body / rng) * 100.0
	}

	bodyATR := math.NaN()
	nonBodyATR := math.NaN()
	if atr > 0 && !math.IsNaN(atr) {
		bodyATR = body / atr
		nonBodyATR = nonBody / atr
	}

	return BarAnalysis{
		Time:              t,
		Open:              open,
		High:              high,
		Low:               low,
		Close:             close,
		Type:              kind,
		Range:             rng,
		Body:              body,
		NonBody:           nonBody,
		ATR:               atr,
		OccupationPercent: occ,
		BodyATRRatio:      bodyATR,
		NonBodyATRRatio:   nonBodyATR,
	}
}

type AggregatedBarStats struct {
	Count              int
	CountVert          int
	CountRouge         int
	SumOccupationPct   float64
	AvgOccupationPct   float64
	SumBodyATRRatio    float64
	SumNonBodyATRRatio float64
	AvgATR             float64

	// Signed aggregation (per user rule)
	SumSignedBody              float64
	SumSignedNonBody           float64
	AvgSignedBody              float64
	AvgSignedNonBody           float64
	SumSignedBodyATRRatio      float64
	SumSignedNonBodyATRRatio   float64
	AvgSignedBodyATRRatio      float64
	AvgSignedNonBodyATRRatio   float64
}

func AggregateAnalyses(list []BarAnalysis) AggregatedBarStats {
	var s AggregatedBarStats
	var sumATR float64
	var atrCount int
	var ratioCount int
	for _, a := range list {
		s.Count++
		if a.Type == "VERT" {
			s.CountVert++
		}
		if a.Type == "ROUGE" {
			s.CountRouge++
		}
		s.SumOccupationPct += a.OccupationPercent
		if !math.IsNaN(a.BodyATRRatio) {
			s.SumBodyATRRatio += a.BodyATRRatio
		}
		if !math.IsNaN(a.NonBodyATRRatio) {
			s.SumNonBodyATRRatio += a.NonBodyATRRatio
		}
		if !math.IsNaN(a.ATR) && a.ATR > 0 {
			sumATR += a.ATR
			atrCount++
		}

		// Signed contributions per user rule
		// VERT: body +, nonBody -
		// ROUGE: body -, nonBody +
		// NEUTRE: both 0
		var sb, snb float64
		switch a.Type {
		case "VERT":
			sb = a.Body
			snb = -a.NonBody
		case "ROUGE":
			sb = -a.Body
			snb = a.NonBody
		default:
			sb = 0
			snb = 0
		}
		s.SumSignedBody += sb
		s.SumSignedNonBody += snb

		// Signed ATR ratios if valid
		if !math.IsNaN(a.BodyATRRatio) && !math.IsNaN(a.NonBodyATRRatio) {
			ratioCount++
			if a.Type == "VERT" {
				s.SumSignedBodyATRRatio += a.BodyATRRatio
				s.SumSignedNonBodyATRRatio += -a.NonBodyATRRatio
			} else if a.Type == "ROUGE" {
				s.SumSignedBodyATRRatio += -a.BodyATRRatio
				s.SumSignedNonBodyATRRatio += a.NonBodyATRRatio
			}
		}
	}
	if s.Count > 0 {
		s.AvgOccupationPct = s.SumOccupationPct / float64(s.Count)
		s.AvgSignedBody = s.SumSignedBody / float64(s.Count)
		s.AvgSignedNonBody = s.SumSignedNonBody / float64(s.Count)
	}
	if atrCount > 0 {
		s.AvgATR = sumATR / float64(atrCount)
	}
	if ratioCount > 0 {
		s.AvgSignedBodyATRRatio = s.SumSignedBodyATRRatio / float64(ratioCount)
		s.AvgSignedNonBodyATRRatio = s.SumSignedNonBodyATRRatio / float64(ratioCount)
	}
	return s
}
