package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	dsbybit "agent-economique/internal/datasource/bybit"
	"agent-economique/internal/indicators"
)

// Defaults (alignés à la démo)
const (
	TF           = "1m"
	ATR_PERIOD   = 3
	BODY_PCT_MIN = 0.40
	BODY_ATR_MIN = 0.40
	EPS          = 1e-9
	BASE_ENABLED = false
	// Configuration du filtre VWMA4
	VWMA4_ENABLED = false
	VWMA4_PERIOD  = 2
	// Configuration du filtre VWMA10
	VWMA10_ENABLED = false
	VWMA10_PERIOD  = 6
	// Configuration du filtre VWMA50
	VWMA50_ENABLED = false
	VWMA50_PERIOD  = 30
	// Configuration du filtre VWMA200
	VWMA200_ENABLED = false
	VWMA200_PERIOD  = 200
	// Filtres composites (window-only): cassure du prix sur DEUX moyennes dans la fenêtre
	COMBO_VWMA4_10_ENABLED  = false
	COMBO_VWMA10_50_ENABLED = false
	// Croisements VWMA-VWMA (window-only) avec écart minimal en multiples d'ATR
	CROSS_VWMA4_10_ENABLED   = false
	CROSS_VWMA4_10_ATR_MULT  = 0.0
	CROSS_VWMA4_50_ENABLED   = true
	CROSS_VWMA4_50_ATR_MULT  = 0.5
	CROSS_VWMA10_50_ENABLED  = false
	CROSS_VWMA10_50_ATR_MULT = 0.0
	// DMI/DX/ADX configuration (spec)
	DI_ADX_PERIOD                    = 10
	ADX_SMOOTH_PERIOD                = 6
	DX_ADX_ENABLED                   = true
	DX_ADX_SCOPE                     = "both" // pre|post|both
	DX_ADX_DIRECTION                 = "up"   // up|down|any
	DX_ADX_GAP_MIN                   = 10.0
	DX_ADX_REQUIRE_UNDER_DI_INFERIOR = false
	DX_ADX_REQUIRE_UNDER_DI_SUPERIOR = true
	DX_REJECT_IF_ABOVE_BOTH_DI       = true
	POST_CROSS_SCORE_ENABLED         = false
	POST_CROSS_SCORE_MIN             = 0.5
	// Window mode
	WINDOW_MODE_ENABLED = true
	WINDOW_SIZE         = 5
	WINDOW_ANCHOR       = "cross_4_50" // base|vwma4|vwma10|vwma50|vwma200|cross_4_10|cross_4_50|cross_10_50|cross_di
)

type Kline struct {
	Ts     int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type BanEcoApp struct {
	symbol     string
	n          int
	bb         *dsbybit.Client
	useVWMA4   bool
	useVWMA10  bool
	useVWMA50  bool
	useVWMA200 bool
}

func NewBanEcoApp(symbol string, n int) *BanEcoApp {
	if n <= 0 {
		n = 1000
	}
	return &BanEcoApp{symbol: symbol, n: n, bb: dsbybit.NewClient(), useVWMA4: VWMA4_ENABLED, useVWMA10: VWMA10_ENABLED, useVWMA50: VWMA50_ENABLED, useVWMA200: VWMA200_ENABLED}
}

func (a *BanEcoApp) Run() error {
	kl, err := a.loadKlines()
	if err != nil {
		return err
	}
	if len(kl) == 0 {
		fmt.Println("Aucune kline")
		return nil
	}

	// 1) Ordonner chronologiquement (sécurité)
	sort.Slice(kl, func(i, j int) bool { return kl[i].Ts < kl[j].Ts })

	// 2) Contrôle des gaps temporels
	intervalMs := int64(60_000) // TF = 1m
	gaps := 0
	for i := 1; i < len(kl); i++ {
		delta := kl[i].Ts - kl[i-1].Ts
		if delta > intervalMs {
			missing := int(delta/intervalMs) - 1
			if missing > 0 {
				gaps += missing
			}
		}
	}

	highs := make([]float64, len(kl))
	lows := make([]float64, len(kl))
	closes := make([]float64, len(kl))
	volumes := make([]float64, len(kl))
	for i, k := range kl {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
		volumes[i] = k.Volume
	}

	atr := indicators.NewATRTVStandard(ATR_PERIOD).Calculate(highs, lows, closes)
	var vwma4 []float64
	if a.useVWMA4 || WINDOW_ANCHOR == "vwma4" || WINDOW_ANCHOR == "cross_4_10" || WINDOW_ANCHOR == "cross_4_50" || COMBO_VWMA4_10_ENABLED || CROSS_VWMA4_10_ENABLED || CROSS_VWMA4_50_ENABLED {
		vwma4 = indicators.NewVWMATVStandard(VWMA4_PERIOD).Calculate(closes, volumes)
	}
	var vwma10 []float64
	if a.useVWMA10 || WINDOW_ANCHOR == "vwma10" || WINDOW_ANCHOR == "cross_4_10" || WINDOW_ANCHOR == "cross_10_50" || COMBO_VWMA4_10_ENABLED || COMBO_VWMA10_50_ENABLED || CROSS_VWMA4_10_ENABLED || CROSS_VWMA10_50_ENABLED {
		vwma10 = indicators.NewVWMATVStandard(VWMA10_PERIOD).Calculate(closes, volumes)
	}
	var vwma50 []float64
	if a.useVWMA50 || WINDOW_ANCHOR == "vwma50" || WINDOW_ANCHOR == "cross_4_50" || WINDOW_ANCHOR == "cross_10_50" || COMBO_VWMA10_50_ENABLED || CROSS_VWMA4_50_ENABLED || CROSS_VWMA10_50_ENABLED {
		vwma50 = indicators.NewVWMATVStandard(VWMA50_PERIOD).Calculate(closes, volumes)
	}
	var vwma200 []float64
	if a.useVWMA200 || WINDOW_ANCHOR == "vwma200" {
		vwma200 = indicators.NewVWMATVStandard(VWMA200_PERIOD).Calculate(closes, volumes)
	}

	// Compute DMI (DI+/DI-, DX, ADX) if needed
	var diPlus, diMinus, adxDI, dxDI, adxTV []float64
	needDMI := WINDOW_ANCHOR == "cross_di" || DX_ADX_ENABLED || DX_ADX_REQUIRE_UNDER_DI_INFERIOR || DX_ADX_REQUIRE_UNDER_DI_SUPERIOR || DX_REJECT_IF_ABOVE_BOTH_DI
	if needDMI {
		dmi := indicators.NewDMITVStandard(DI_ADX_PERIOD)
		diPlus, diMinus, adxTV = dmi.Calculate(highs, lows, closes)
		dxDI = dmi.CalculateDX(diPlus, diMinus)
		adxDI = indicators.NewRMATVStandard(ADX_SMOOTH_PERIOD).Calculate(dxDI)
	}

	if WINDOW_MODE_ENABLED {
		countWin := 0
		half := WINDOW_SIZE / 2
		anchorIsCross := WINDOW_ANCHOR == "cross_4_10" || WINDOW_ANCHOR == "cross_4_50" || WINDOW_ANCHOR == "cross_10_50"
		for i := 1; i < len(kl)-1; i++ {
			sig := ""
			anchorOK := false
			// Track combos across the window
			c41_4hit, c41_10hit := false, false
			c1050_10hit, c1050_50hit := false, false
			// Track VWMA-to-VWMA validated gaps across the window
			x41Hit, x450Hit, x1050Hit := false, false, false
			// Track DX/ADX validated gap across the window
			dxadxHit := false
			// Index of first post-anchor bar where all conditions become true
			signalAtIdx := -1
			// Post-cross cumulative score (aligned vs contrary)
			posScore, negScore := 0.0, 0.0
			switch WINDOW_ANCHOR {
			case "base":
				if i < len(atr) && !math.IsNaN(atr[i]) {
					k := kl[i]
					rng := k.High - k.Low
					if rng > 0 {
						body := math.Abs(k.Close - k.Open)
						bodyPct := body / rng
						if bodyPct+EPS >= BODY_PCT_MIN && body+EPS >= BODY_ATR_MIN*atr[i] {
							if k.Close > k.Open {
								sig = "LONG"
							} else if k.Close < k.Open {
								sig = "SHORT"
							}
							anchorOK = sig != ""
						}
					}
				}
			case "vwma4":
				if len(vwma4) > 0 && !math.IsNaN(vwma4[i]) && i-1 >= 0 {
					prevClose := kl[i-1].Close
					prevV := vwma4[i-1]
					curV := vwma4[i]
					if !math.IsNaN(prevV) {
						if kl[i].Close > curV && prevClose <= prevV {
							sig = "LONG"
							anchorOK = true
						} else if kl[i].Close < curV && prevClose >= prevV {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "vwma10":
				if len(vwma10) > 0 && !math.IsNaN(vwma10[i]) && i-1 >= 0 {
					prevClose := kl[i-1].Close
					prevV := vwma10[i-1]
					curV := vwma10[i]
					if !math.IsNaN(prevV) {
						if kl[i].Close > curV && prevClose <= prevV {
							sig = "LONG"
							anchorOK = true
						} else if kl[i].Close < curV && prevClose >= prevV {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "vwma50":
				if len(vwma50) > 0 && !math.IsNaN(vwma50[i]) && i-1 >= 0 {
					prevClose := kl[i-1].Close
					prevV := vwma50[i-1]
					curV := vwma50[i]
					if !math.IsNaN(prevV) {
						if kl[i].Close > curV && prevClose <= prevV {
							sig = "LONG"
							anchorOK = true
						} else if kl[i].Close < curV && prevClose >= prevV {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "vwma200":
				if len(vwma200) > 0 && !math.IsNaN(vwma200[i]) && i-1 >= 0 {
					prevClose := kl[i-1].Close
					prevV := vwma200[i-1]
					curV := vwma200[i]
					if !math.IsNaN(prevV) {
						if kl[i].Close > curV && prevClose <= prevV {
							sig = "LONG"
							anchorOK = true
						} else if kl[i].Close < curV && prevClose >= prevV {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "cross_4_10":
				if len(vwma4) > 0 && len(vwma10) > 0 && i-1 >= 0 && !math.IsNaN(vwma4[i]) && !math.IsNaN(vwma10[i]) {
					pf, ps := vwma4[i-1], vwma10[i-1]
					cf, cs := vwma4[i], vwma10[i]
					if !math.IsNaN(pf) && !math.IsNaN(ps) {
						if pf <= ps && cf > cs {
							sig = "LONG"
							anchorOK = true
						} else if pf >= ps && cf < cs {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "cross_4_50":
				if len(vwma4) > 0 && len(vwma50) > 0 && i-1 >= 0 && !math.IsNaN(vwma4[i]) && !math.IsNaN(vwma50[i]) {
					pf, ps := vwma4[i-1], vwma50[i-1]
					cf, cs := vwma4[i], vwma50[i]
					if !math.IsNaN(pf) && !math.IsNaN(ps) {
						if pf <= ps && cf > cs {
							sig = "LONG"
							anchorOK = true
						} else if pf >= ps && cf < cs {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "cross_10_50":
				if len(vwma10) > 0 && len(vwma50) > 0 && i-1 >= 0 && !math.IsNaN(vwma10[i]) && !math.IsNaN(vwma50[i]) {
					pf, ps := vwma10[i-1], vwma50[i-1]
					cf, cs := vwma10[i], vwma50[i]
					if !math.IsNaN(pf) && !math.IsNaN(ps) {
						if pf <= ps && cf > cs {
							sig = "LONG"
							anchorOK = true
						} else if pf >= ps && cf < cs {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			case "cross_di":
				if needDMI && len(diPlus) > 0 && len(diMinus) > 0 && i-1 >= 0 && !math.IsNaN(diPlus[i]) && !math.IsNaN(diMinus[i]) {
					pf, ps := diPlus[i-1], diMinus[i-1]
					cf, cs := diPlus[i], diMinus[i]
					if !math.IsNaN(pf) && !math.IsNaN(ps) {
						if pf <= ps && cf > cs {
							sig = "LONG"
							anchorOK = true
						} else if pf >= ps && cf < cs {
							sig = "SHORT"
							anchorOK = true
						}
					}
				}
			}
			if !anchorOK {
				continue
			}

			need := map[string]bool{}
			if BASE_ENABLED {
				need["base"] = true
			}
			if a.useVWMA4 {
				need["vwma4"] = true
			}
			if a.useVWMA10 {
				need["vwma10"] = true
			}
			if a.useVWMA50 {
				need["vwma50"] = true
			}
			if a.useVWMA200 {
				need["vwma200"] = true
			}

			// credit anchor if it corresponds to a required condition
			switch WINDOW_ANCHOR {
			case "base":
				if need["base"] {
					delete(need, "base")
				}
			case "vwma4":
				if need["vwma4"] {
					delete(need, "vwma4")
				}
				if COMBO_VWMA4_10_ENABLED {
					c41_4hit = true
				}
			case "vwma10":
				if need["vwma10"] {
					delete(need, "vwma10")
				}
				if COMBO_VWMA4_10_ENABLED {
					c41_10hit = true
				}
				if COMBO_VWMA10_50_ENABLED {
					c1050_10hit = true
				}
			case "vwma50":
				if need["vwma50"] {
					delete(need, "vwma50")
				}
				if COMBO_VWMA10_50_ENABLED {
					c1050_50hit = true
				}
			case "vwma200":
				if need["vwma200"] {
					delete(need, "vwma200")
				}
				// no credit for cross_* at anchor; gap is validated within window
			}

			preStart := i - half
			if preStart < 0 {
				preStart = 0
			}
			postEnd := i + WINDOW_SIZE
			if postEnd >= len(kl) {
				postEnd = len(kl) - 1
			}

			// Prepare pair-specific anchor flags and cross indices for filters
			isAnchor41 := WINDOW_ANCHOR == "cross_4_10"
			isAnchor450 := WINDOW_ANCHOR == "cross_4_50"
			isAnchor1050 := WINDOW_ANCHOR == "cross_10_50"
			cross41Idx, cross450Idx, cross1050Idx := -1, -1, -1
			dxadxCrossIdx := -1

			// pre-window
			for j := preStart; j < i; j++ {
				// base
				if need["base"] && j < len(atr) && !math.IsNaN(atr[j]) {
					k := kl[j]
					rng := k.High - k.Low
					if rng > 0 {
						body := math.Abs(k.Close - k.Open)
						bodyPct := body / rng
						if bodyPct+EPS >= BODY_PCT_MIN && body+EPS >= BODY_ATR_MIN*atr[j] {
							if (sig == "LONG" && k.Close > k.Open) || (sig == "SHORT" && k.Close < k.Open) {
								delete(need, "base")
							}
						}
					}
				}
				// vwma4
				if need["vwma4"] && len(vwma4) > 0 && j-1 >= 0 && !math.IsNaN(vwma4[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma4[j-1] && kl[j].Close > vwma4[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma4[j-1] && kl[j].Close < vwma4[j]) {
						delete(need, "vwma4")
						c41_4hit = true
					}
				}
				if need["vwma10"] && len(vwma10) > 0 && j-1 >= 0 && !math.IsNaN(vwma10[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma10[j-1] && kl[j].Close > vwma10[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma10[j-1] && kl[j].Close < vwma10[j]) {
						delete(need, "vwma10")
						c41_10hit = true
						c1050_10hit = true
					}
				}
				if need["vwma50"] && len(vwma50) > 0 && j-1 >= 0 && !math.IsNaN(vwma50[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma50[j-1] && kl[j].Close > vwma50[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma50[j-1] && kl[j].Close < vwma50[j]) {
						delete(need, "vwma50")
						c1050_50hit = true
					}
				}
				// VWMA cross-as-filter gating in pre-window (pair must cross first, then validate gap after its cross). Anchor pair skipped in pre-window.
				// cross 4/10
				if CROSS_VWMA4_10_ENABLED && !x41Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma4) > 0 && len(vwma10) > 0 && !math.IsNaN(vwma4[j]) && !math.IsNaN(vwma10[j]) {
					// detect cross once, aligned with anchor direction
					if cross41Idx < 0 && j-1 >= 0 && !isAnchor41 {
						pf, ps := vwma4[j-1], vwma10[j-1]
						cf, cs := vwma4[j], vwma10[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross41Idx = j
						}
					}
					if !isAnchor41 && cross41Idx >= 0 && j >= cross41Idx {
						if sig == "LONG" {
							gap := vwma4[j] - vwma10[j]
							if gap+EPS >= CROSS_VWMA4_10_ATR_MULT*atr[j] {
								x41Hit = true
							}
						} else if sig == "SHORT" {
							gap := vwma10[j] - vwma4[j]
							if gap+EPS >= CROSS_VWMA4_10_ATR_MULT*atr[j] {
								x41Hit = true
							}
						}
					}
				}
				// cross 4/50
				if CROSS_VWMA4_50_ENABLED && !x450Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma4) > 0 && len(vwma50) > 0 && !math.IsNaN(vwma4[j]) && !math.IsNaN(vwma50[j]) {
					if cross450Idx < 0 && j-1 >= 0 && !isAnchor450 {
						pf, ps := vwma4[j-1], vwma50[j-1]
						cf, cs := vwma4[j], vwma50[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross450Idx = j
						}
					}
					if !isAnchor450 && cross450Idx >= 0 && j >= cross450Idx {
						if sig == "LONG" {
							gap := vwma4[j] - vwma50[j]
							if gap+EPS >= CROSS_VWMA4_50_ATR_MULT*atr[j] {
								x450Hit = true
							}
						} else if sig == "SHORT" {
							gap := vwma50[j] - vwma4[j]
							if gap+EPS >= CROSS_VWMA4_50_ATR_MULT*atr[j] {
								x450Hit = true
							}
						}
					}
				}
				// cross 10/50
				if CROSS_VWMA10_50_ENABLED && !x1050Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma10) > 0 && len(vwma50) > 0 && !math.IsNaN(vwma10[j]) && !math.IsNaN(vwma50[j]) {
					if cross1050Idx < 0 && j-1 >= 0 && !isAnchor1050 {
						pf, ps := vwma10[j-1], vwma50[j-1]
						cf, cs := vwma10[j], vwma50[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross1050Idx = j
						}
					}
					if !isAnchor1050 && cross1050Idx >= 0 && j >= cross1050Idx {
						if sig == "LONG" {
							gap := vwma10[j] - vwma50[j]
							if gap+EPS >= CROSS_VWMA10_50_ATR_MULT*atr[j] {
								x1050Hit = true
							}
						} else if sig == "SHORT" {
							gap := vwma50[j] - vwma10[j]
							if gap+EPS >= CROSS_VWMA10_50_ATR_MULT*atr[j] {
								x1050Hit = true
							}
						}
					}
				}

				// DX/ADX as filter: detect cross in pre-window (if scope allows), then validate gap after that cross
				if DX_ADX_ENABLED && needDMI && !dxadxHit && (DX_ADX_SCOPE == "pre" || DX_ADX_SCOPE == "both") && j < len(dxDI) && !math.IsNaN(dxDI[j]) && !math.IsNaN(adxDI[j]) {
					if dxadxCrossIdx < 0 && j-1 >= 0 && !math.IsNaN(dxDI[j-1]) && !math.IsNaN(adxDI[j-1]) {
						up := dxDI[j-1] <= adxDI[j-1] && dxDI[j] > adxDI[j]
						down := dxDI[j-1] >= adxDI[j-1] && dxDI[j] < adxDI[j]
						if (DX_ADX_DIRECTION == "up" && up) || (DX_ADX_DIRECTION == "down" && down) || (DX_ADX_DIRECTION == "any" && (up || down)) {
							ok := true
							if DX_ADX_REQUIRE_UNDER_DI_INFERIOR {
								diInf := math.Min(diPlus[j], diMinus[j])
								ok = ok && dxDI[j] <= diInf && adxDI[j] <= diInf
							}
							if DX_ADX_REQUIRE_UNDER_DI_SUPERIOR {
								diSup := math.Max(diPlus[j], diMinus[j])
								ok = ok && dxDI[j] <= diSup && adxDI[j] <= diSup
							}
							if ok {
								dxadxCrossIdx = j
							}
						}
					}
					if dxadxCrossIdx >= 0 && j >= dxadxCrossIdx {
						gap := math.Abs(dxDI[j] - adxDI[j])
						if gap+EPS >= DX_ADX_GAP_MIN {
							dxadxHit = true
						}
					}
				}
			}
			done := false
			for j := i + 1; j <= postEnd && !done; j++ {
				// accumulate post-cross score until validation, if enabled
				if anchorIsCross && POST_CROSS_SCORE_ENABLED && j < len(atr) && !math.IsNaN(atr[j]) {
					k := kl[j]
					rng := k.High - k.Low
					if rng > 0 && atr[j] > 0 {
						body := math.Abs(k.Close - k.Open)
						bodyATR := body / atr[j]
						if bodyATR > 0 {
							if (sig == "LONG" && k.Close > k.Open) || (sig == "SHORT" && k.Close < k.Open) {
								posScore += bodyATR
							} else if (sig == "LONG" && k.Close < k.Open) || (sig == "SHORT" && k.Close > k.Open) {
								negScore += bodyATR
							}
						}
					}
				}
				if need["base"] && j < len(atr) && !math.IsNaN(atr[j]) {
					k := kl[j]
					rng := k.High - k.Low
					if rng > 0 {
						body := math.Abs(k.Close - k.Open)
						bodyPct := body / rng
						if bodyPct+EPS >= BODY_PCT_MIN && body+EPS >= BODY_ATR_MIN*atr[j] {
							if (sig == "LONG" && k.Close > k.Open) || (sig == "SHORT" && k.Close < k.Open) {
								delete(need, "base")
							}
						}
					}
				}
				if need["vwma4"] && len(vwma4) > 0 && j-1 >= 0 && !math.IsNaN(vwma4[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma4[j-1] && kl[j].Close > vwma4[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma4[j-1] && kl[j].Close < vwma4[j]) {
						delete(need, "vwma4")
						c41_4hit = true
					}
				}
				if need["vwma10"] && len(vwma10) > 0 && j-1 >= 0 && !math.IsNaN(vwma10[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma10[j-1] && kl[j].Close > vwma10[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma10[j-1] && kl[j].Close < vwma10[j]) {
						delete(need, "vwma10")
						c41_10hit = true
						c1050_10hit = true
					}
				}
				if need["vwma50"] && len(vwma50) > 0 && j-1 >= 0 && !math.IsNaN(vwma50[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma50[j-1] && kl[j].Close > vwma50[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma50[j-1] && kl[j].Close < vwma50[j]) {
						delete(need, "vwma50")
						c1050_50hit = true
					}
				}
				if need["vwma200"] && len(vwma200) > 0 && j-1 >= 0 && !math.IsNaN(vwma200[j]) {
					if (sig == "LONG" && kl[j-1].Close <= vwma200[j-1] && kl[j].Close > vwma200[j]) || (sig == "SHORT" && kl[j-1].Close >= vwma200[j-1] && kl[j].Close < vwma200[j]) {
						delete(need, "vwma200")
					}
				}
				// VWMA gap checks (no re-cross required) in post-window
				if CROSS_VWMA4_10_ENABLED && !x41Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma4) > 0 && len(vwma10) > 0 && !math.IsNaN(vwma4[j]) && !math.IsNaN(vwma10[j]) {
					// detect cross for filter pair if not anchor
					if cross41Idx < 0 && j-1 >= 0 && !isAnchor41 {
						pf, ps := vwma4[j-1], vwma10[j-1]
						cf, cs := vwma4[j], vwma10[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross41Idx = j
						}
					}
					if sig == "LONG" {
						gap := vwma4[j] - vwma10[j]
						if gap+EPS >= CROSS_VWMA4_10_ATR_MULT*atr[j] {
							if !isAnchor41 {
								if cross41Idx >= 0 && j >= cross41Idx {
									x41Hit = true
								}
							} else {
								// anchor pair
								x41Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x41Hit = true
								}
							} else {
								x41Hit = true
							}
						}
					} else if sig == "SHORT" {
						gap := vwma10[j] - vwma4[j]
						if gap+EPS >= CROSS_VWMA4_10_ATR_MULT*atr[j] {
							if !isAnchor41 {
								if cross41Idx >= 0 && j >= cross41Idx {
									x41Hit = true
								}
							} else {
								x41Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x41Hit = true
								}
							} else {
								x41Hit = true
							}
						}
					}
				}
				if CROSS_VWMA4_50_ENABLED && !x450Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma4) > 0 && len(vwma50) > 0 && !math.IsNaN(vwma4[j]) && !math.IsNaN(vwma50[j]) {
					if cross450Idx < 0 && j-1 >= 0 && !isAnchor450 {
						pf, ps := vwma4[j-1], vwma50[j-1]
						cf, cs := vwma4[j], vwma50[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross450Idx = j
						}
					}
					if sig == "LONG" {
						gap := vwma4[j] - vwma50[j]
						if gap+EPS >= CROSS_VWMA4_50_ATR_MULT*atr[j] {
							if !isAnchor450 {
								if cross450Idx >= 0 && j >= cross450Idx {
									x450Hit = true
								}
							} else {
								x450Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x450Hit = true
								}
							} else {
								x450Hit = true
							}
						}
					} else if sig == "SHORT" {
						gap := vwma50[j] - vwma4[j]
						if gap+EPS >= CROSS_VWMA4_50_ATR_MULT*atr[j] {
							if !isAnchor450 {
								if cross450Idx >= 0 && j >= cross450Idx {
									x450Hit = true
								}
							} else {
								x450Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x450Hit = true
								}
							} else {
								x450Hit = true
							}
						}
					}
				}
				if CROSS_VWMA10_50_ENABLED && !x1050Hit && j < len(atr) && !math.IsNaN(atr[j]) && len(vwma10) > 0 && len(vwma50) > 0 && !math.IsNaN(vwma10[j]) && !math.IsNaN(vwma50[j]) {
					if cross1050Idx < 0 && j-1 >= 0 && !isAnchor1050 {
						pf, ps := vwma10[j-1], vwma50[j-1]
						cf, cs := vwma10[j], vwma50[j]
						bull := pf <= ps && cf > cs
						bear := pf >= ps && cf < cs
						if (sig == "LONG" && bull) || (sig == "SHORT" && bear) {
							cross1050Idx = j
						}
					}
					if sig == "LONG" {
						gap := vwma10[j] - vwma50[j]
						if gap+EPS >= CROSS_VWMA10_50_ATR_MULT*atr[j] {
							if !isAnchor1050 {
								if cross1050Idx >= 0 && j >= cross1050Idx {
									x1050Hit = true
								}
							} else {
								x1050Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x1050Hit = true
								}
							} else {
								x1050Hit = true
							}
						}
					} else if sig == "SHORT" {
						gap := vwma50[j] - vwma10[j]
						if gap+EPS >= CROSS_VWMA10_50_ATR_MULT*atr[j] {
							if !isAnchor1050 {
								if cross1050Idx >= 0 && j >= cross1050Idx {
									x1050Hit = true
								}
							} else {
								x1050Hit = true
							}
							if anchorIsCross && POST_CROSS_SCORE_ENABLED {
								if (posScore-negScore)+EPS >= POST_CROSS_SCORE_MIN {
									x1050Hit = true
								}
							} else {
								x1050Hit = true
							}
						}
					}
				}

				// DX/ADX as filter: detect cross in post-window (if scope allows), then validate gap after that cross
				if DX_ADX_ENABLED && needDMI && !dxadxHit && (DX_ADX_SCOPE == "post" || DX_ADX_SCOPE == "both") && j < len(dxDI) && !math.IsNaN(dxDI[j]) && !math.IsNaN(adxDI[j]) {
					if dxadxCrossIdx < 0 && j-1 >= 0 && !math.IsNaN(dxDI[j-1]) && !math.IsNaN(adxDI[j-1]) {
						up := dxDI[j-1] <= adxDI[j-1] && dxDI[j] > adxDI[j]
						down := dxDI[j-1] >= adxDI[j-1] && dxDI[j] < adxDI[j]
						if (DX_ADX_DIRECTION == "up" && up) || (DX_ADX_DIRECTION == "down" && down) || (DX_ADX_DIRECTION == "any" && (up || down)) {
							ok := true
							if DX_ADX_REQUIRE_UNDER_DI_INFERIOR {
								diInf := math.Min(diPlus[j], diMinus[j])
								ok = ok && dxDI[j] <= diInf && adxDI[j] <= diInf
							}
							if DX_ADX_REQUIRE_UNDER_DI_SUPERIOR {
								diSup := math.Max(diPlus[j], diMinus[j])
								ok = ok && dxDI[j] <= diSup && adxDI[j] <= diSup
							}
							if ok {
								dxadxCrossIdx = j
							}
						}
					}
					if dxadxCrossIdx >= 0 && j >= dxadxCrossIdx {
						gap := math.Abs(dxDI[j] - adxDI[j])
						if gap+EPS >= DX_ADX_GAP_MIN {
							dxadxHit = true
						}
					}
				}

				// Capture j* at the first moment where all conditions are satisfied
				if signalAtIdx < 0 {
					allSinglesOKNow := len(need) == 0
					combosOKNow := true
					if COMBO_VWMA4_10_ENABLED {
						combosOKNow = combosOKNow && (c41_4hit && c41_10hit)
					}
					if COMBO_VWMA10_50_ENABLED {
						combosOKNow = combosOKNow && (c1050_10hit && c1050_50hit)
					}
					crossesOKNow := true
					if CROSS_VWMA4_10_ENABLED {
						crossesOKNow = crossesOKNow && x41Hit
					}
					if CROSS_VWMA4_50_ENABLED {
						crossesOKNow = crossesOKNow && x450Hit
					}
					if CROSS_VWMA10_50_ENABLED {
						crossesOKNow = crossesOKNow && x1050Hit
					}
					if DX_ADX_ENABLED {
						crossesOKNow = crossesOKNow && dxadxHit
					}
					if allSinglesOKNow && combosOKNow && crossesOKNow {
						// Optional final DX guard: reject if DX > DI+ and DX > DI- at j*
						if DX_REJECT_IF_ABOVE_BOTH_DI && needDMI && j < len(dxDI) && j < len(diPlus) && j < len(diMinus) &&
							!math.IsNaN(dxDI[j]) && !math.IsNaN(diPlus[j]) && !math.IsNaN(diMinus[j]) {
							if dxDI[j] > diPlus[j] && dxDI[j] > diMinus[j] {
								// reject this j, continue scanning for a later j*
							} else {
								signalAtIdx = j
							}
						} else {
							signalAtIdx = j
						}
					}
				}
			}
			if signalAtIdx >= 0 {
				ts := time.Unix(0, kl[signalAtIdx].Ts*1e6).UTC()
				j := signalAtIdx
				dPlus, dMinus, dxv, adxv := math.NaN(), math.NaN(), math.NaN(), math.NaN()
				if j < len(diPlus) && j < len(diMinus) && j < len(dxDI) && j < len(adxDI) {
					dPlus = diPlus[j]
					dMinus = diMinus[j]
					dxv = dxDI[j]
					adxv = adxDI[j]
				}
				v4, v10, v50, v200 := math.NaN(), math.NaN(), math.NaN(), math.NaN()
				if j < len(vwma4) { v4 = vwma4[j] }
				if j < len(vwma10) { v10 = vwma10[j] }
				if j < len(vwma50) { v50 = vwma50[j] }
				if j < len(vwma200) { v200 = vwma200[j] }
				fmt.Printf("[WIN] anchor=%s | dir=%s | at=%s | window=%d | dmi: di+=%.2f di-=%.2f dx=%.2f adx=%.2f | vwma: v4=%.5f v10=%.5f v50=%.5f v200=%.5f | ok=ALL\n",
					WINDOW_ANCHOR, sig, ts.Format(time.RFC3339), WINDOW_SIZE, dPlus, dMinus, dxv, adxv, v4, v10, v50, v200)
				// Debug line: compare ADX (same period as DI) vs ADX decoupled, and verify kline spacing
				adxSame := math.NaN()
				if j < len(adxTV) { adxSame = adxTV[j] }
				var prevTs, nextTs int64
				if j > 0 { prevTs = kl[j-1].Ts }
				if j+1 < len(kl) { nextTs = kl[j+1].Ts }
				deltaPrev := int64(0)
				deltaNext := int64(0)
				if prevTs > 0 { deltaPrev = kl[j].Ts - prevTs }
				if nextTs > 0 { deltaNext = nextTs - kl[j].Ts }
				k := kl[j]
				fmt.Printf("[DMI-CHK] j=%d ts=%s | di+=%.2f di-=%.2f dx=%.2f adx=%.2f adx_tv=%.2f | ohlc=%.5f/%.5f/%.5f/%.5f | Δprev=%dms Δnext=%dms (expect=%dms) | periods: di=%d adx=%d\n",
					j, ts.Format(time.RFC3339), dPlus, dMinus, dxv, adxv, adxSame, k.Open, k.High, k.Low, k.Close, deltaPrev, deltaNext, intervalMs, DI_ADX_PERIOD, ADX_SMOOTH_PERIOD)
				countWin++
				i = i + half
			}
		}
		fmt.Printf("\nRésumé (window): %d signaux\n", countWin)
		return nil
	}

	count := 0
	warmupSkipped := 0
	zeroRangeSkipped := 0
	for i := range kl {
		if i >= len(atr) || math.IsNaN(atr[i]) {
			warmupSkipped++
			continue
		}
		k := kl[i]
		rng := k.High - k.Low
		if rng <= 0 {
			zeroRangeSkipped++
			continue
		}
		body := math.Abs(k.Close - k.Open)
		bodyPct := body / rng
		if bodyPct+EPS >= BODY_PCT_MIN && body+EPS >= BODY_ATR_MIN*atr[i] {
			// Déterminer la direction du signal selon Close vs Open
			sig := "FLAT"
			if k.Close > k.Open {
				sig = "LONG"
			} else if k.Close < k.Open {
				sig = "SHORT"
			}

			// Si filtre VWMA4 demandé: exiger une cassure par rapport à la bougie précédente
			if a.useVWMA4 {
				if i-1 < 0 || i >= len(vwma4) || math.IsNaN(vwma4[i]) {
					continue
				}
				prevClose := kl[i-1].Close
				prevV := vwma4[i-1]
				curV := vwma4[i]
				if math.IsNaN(prevV) || math.IsNaN(curV) {
					continue
				}
				if sig == "LONG" {
					// cassure haussière: close traverse au-dessus de VWMA4
					if !(prevClose <= prevV && k.Close > curV) {
						continue
					}
				} else if sig == "SHORT" {
					// cassure baissière: close traverse en-dessous de VWMA4
					if !(prevClose >= prevV && k.Close < curV) {
						continue
					}
				} else {
					continue
				}
			}

			// Si filtre VWMA10 demandé: exiger une cassure similaire
			if a.useVWMA10 {
				if i-1 < 0 || i >= len(vwma10) || math.IsNaN(vwma10[i]) {
					continue
				}
				prevClose := kl[i-1].Close
				prevV := vwma10[i-1]
				curV := vwma10[i]
				if math.IsNaN(prevV) || math.IsNaN(curV) {
					continue
				}
				if sig == "LONG" {
					if !(prevClose <= prevV && k.Close > curV) {
						continue
					}
				} else if sig == "SHORT" {
					if !(prevClose >= prevV && k.Close < curV) {
						continue
					}
				} else {
					continue
				}
			}

			// Si filtre VWMA50 demandé: exiger une cassure similaire
			if a.useVWMA50 {
				if i-1 < 0 || i >= len(vwma50) || math.IsNaN(vwma50[i]) {
					continue
				}
				prevClose := kl[i-1].Close
				prevV := vwma50[i-1]
				curV := vwma50[i]
				if math.IsNaN(prevV) || math.IsNaN(curV) {
					continue
				}
				if sig == "LONG" {
					if !(prevClose <= prevV && k.Close > curV) {
						continue
					}
				} else if sig == "SHORT" {
					if !(prevClose >= prevV && k.Close < curV) {
						continue
					}
				} else {
					continue
				}
			}

			// Si filtre VWMA200 demandé: exiger une cassure similaire
			if a.useVWMA200 {
				if i-1 < 0 || i >= len(vwma200) || math.IsNaN(vwma200[i]) {
					continue
				}
				prevClose := kl[i-1].Close
				prevV := vwma200[i-1]
				curV := vwma200[i]
				if math.IsNaN(prevV) || math.IsNaN(curV) {
					continue
				}
				if sig == "LONG" {
					if !(prevClose <= prevV && k.Close > curV) {
						continue
					}
				} else if sig == "SHORT" {
					if !(prevClose >= prevV && k.Close < curV) {
						continue
					}
				} else {
					continue
				}
			}

			count++
			ts := time.Unix(0, k.Ts*1e6).UTC()
			// Logging: construire dynamiquement la ligne avec les VWMA actives
			line := fmt.Sprintf("[BAN] %s | signal=%s | close=%.6f | body%%=%.2f | body/ATR=%.2f | atr=%.5f",
				ts.Format(time.RFC3339), sig, k.Close, bodyPct, body/atr[i], atr[i])
			if a.useVWMA4 {
				line += fmt.Sprintf(" | vwma4=%.5f", vwma4[i])
			}
			if a.useVWMA10 {
				line += fmt.Sprintf(" | vwma10=%.5f", vwma10[i])
			}
			if a.useVWMA50 {
				line += fmt.Sprintf(" | vwma50=%.5f", vwma50[i])
			}
			if a.useVWMA200 {
				line += fmt.Sprintf(" | vwma200=%.5f", vwma200[i])
			}
			fmt.Println(line)
		}
	}

	fmt.Printf("\nRésumé: %d bougies matchées sur %d | gaps détectés=%d | warmup_skipped=%d | zero_range=%d\n", count, len(kl), gaps, warmupSkipped, zeroRangeSkipped)
	return nil
}

func (a *BanEcoApp) loadKlines() ([]Kline, error) {
	gk, err := a.bb.GetKlines(context.Background(), a.symbol, TF, a.n)
	if err != nil {
		return nil, fmt.Errorf("Bybit SDK error: %w", err)
	}
	out := make([]Kline, len(gk))
	for i, k := range gk {
		out[i] = Kline{
			Ts:     k.OpenTime.UnixMilli(),
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
		}
	}
	return out, nil
}
