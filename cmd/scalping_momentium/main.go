package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/indicators"
)

// ============================================================================
// PARAMETRES CONFIGURABLES
// ============================================================================

const (
	SYMBOL            = "SOL_USDT" // Format Gate.io
	TIMEFRAME         = "5m"
	NB_CANDLES        = 1000
	ATR_PERIOD        = 3
	BODY_PCT_MIN      = 0.60 // Corps >= 60% de la bougie (High-Low)
	BODY_ATR_MIN      = 0.60 // Corps >= 90% de l'ATR(3)
	STOCH_K_LONG_MAX  = 50.0 // LONG valide si K < 40
	STOCH_K_SHORT_MIN = 50.0 // SHORT valide si K > 60
	STOCH_K_PERIOD    = 14
	STOCH_K_SMOOTH    = 3
	STOCH_D_PERIOD    = 3
)

// SignalMomentum représente un signal détecté
type SignalMomentum struct {
	Index      int
	Timestamp  time.Time
	Type       string // LONG | SHORT
	Label      string // ENTRY | EXIT
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Body       float64 // |Close-Open|
	Range      float64 // High-Low
	BodyPctBar float64 // Body/Range
	ATR        float64 // ATR(3)
	BodyToATR  float64 // Body/ATR
	StochK     float64
}

func main() {
	fmt.Println("=== DEMO STRATEGIE SCALPING MOMENTIUM (Gate.io) ===")
	fmt.Printf("Symbole     : %s\n", SYMBOL)
	fmt.Printf("Timeframe   : %s\n", TIMEFRAME)
	fmt.Printf("Bougie type : Corps >= %.0f%% du range ET Corps >= %.0f%% de l'ATR(%d)\n",
		BODY_PCT_MIN*100, BODY_ATR_MIN*100, ATR_PERIOD)

	// 1) Récupération des données Gate.io
	fmt.Println("\nRécupération des klines...")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("Erreur récupération klines: %v", err)
	}
	fmt.Printf("Klines récupérées: %d\n", len(klines))

	if len(klines) == 0 {
		fmt.Println("Aucune donnée.")
		return
	}

	// 2) Préparer tableaux prix
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	closes := make([]float64, len(klines))
	opens := make([]float64, len(klines))
	for i, k := range klines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
		opens[i] = k.Open
	}

	// 3) Calcul ATR(3)
	atrInd := indicators.NewATRTVStandard(ATR_PERIOD)
	atr := atrInd.Calculate(highs, lows, closes)

	// 3bis) Calcul Stochastic (14,1,3)
	stochTV := indicators.NewStochTVStandard(STOCH_K_PERIOD, STOCH_K_SMOOTH, STOCH_D_PERIOD)
	stochK, _ := stochTV.Calculate(highs, lows, closes)

	// 4) Détection signaux
	fmt.Println("\nDétection des signaux (corps>60% & corps>0.9*ATR)...")
	signals := detectMomentumSignals(klines, opens, highs, lows, closes, atr, stochK)

	// 5) Affichage
	displaySignals(signals)

	fmt.Println("\n=== FIN DEMO ===")
}

func detectMomentumSignals(klines []gateio.Kline, opens, highs, lows, closes, atr, stochK []float64) []SignalMomentum {
	res := []SignalMomentum{}
	start := ATR_PERIOD // ATR seed à partir de period-1; on démarre prudemment à period
	if start < 1 {
		start = 1
	}
	for i := start; i < len(klines); i++ {
		if i >= len(atr) || math.IsNaN(atr[i]) {
			continue
		}

		rangeHL := highs[i] - lows[i]
		if rangeHL <= 0 {
			continue
		}
		body := math.Abs(closes[i] - opens[i])
		bodyPct := body / rangeHL
		if bodyPct < BODY_PCT_MIN {
			continue
		}

		if body < BODY_ATR_MIN*atr[i] {
			continue
		}

		dir := ""
		if closes[i] > opens[i] {
			dir = "LONG"
		} else if closes[i] < opens[i] {
			dir = "SHORT"
		} else {
			continue // doji (pas de signal)
		}

		// Filtre Stoch K: LONG valide si K<40, SHORT valide si K>60
		if i >= len(stochK) || math.IsNaN(stochK[i]) {
			continue
		}
		if dir == "LONG" {
			if !(stochK[i] < STOCH_K_LONG_MAX) {
				continue
			}
		} else if dir == "SHORT" {
			if !(stochK[i] > STOCH_K_SHORT_MIN) {
				continue
			}
		}

		// Déterminer l'étiquette ENTRY/EXIT selon le second filtre
		label := "EXIT"
		if i >= 2 {
			c0 := closes[i]
			if dir == "LONG" {
				// Close actuel doit être le sommet des 3 dernières clôtures
				r1 := math.Max(opens[i-1], closes[i-1])
				r2 := math.Max(opens[i-2], closes[i-2])
				if c0 >= r1 && c0 >= r2 {
					label = "ENTRY"
				}
			} else if dir == "SHORT" {
				// Close actuel doit être le creux des 3 dernières clôtures
				r1 := math.Min(opens[i-1], closes[i-1])
				r2 := math.Min(opens[i-2], closes[i-2])
				if c0 <= r1 && c0 <= r2 {
					label = "ENTRY"
				}
			}
		}

		s := SignalMomentum{
			Index:      i,
			Timestamp:  klines[i].OpenTime,
			Type:       dir,
			Label:      label,
			Open:       opens[i],
			High:       highs[i],
			Low:        lows[i],
			Close:      closes[i],
			Body:       body,
			Range:      rangeHL,
			BodyPctBar: bodyPct,
			ATR:        atr[i],
			BodyToATR:  safeDiv(body, atr[i]),
			StochK:     stochK[i],
		}
		res = append(res, s)
	}
	return res
}

func displaySignals(signals []SignalMomentum) {
	fmt.Println("\n" + strings.Repeat("=", 150))
	fmt.Println("SIGNAUX MOMENTUM (Corps >= 60% du range & Corps >= 90% ATR(3))")
	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("%-4s | %-19s | %-6s | %-5s | %8s | %7s | %8s | %6s | %8s | %8s | %7s\n",
		"#", "Date/Heure", "Type", "Tag", "Close", "StochK", "Body", "Body%", "ATR(3)", "Body/ATR", "Capté")
	fmt.Println(strings.Repeat("-", 150))

	longCount := 0
	shortCount := 0
	sumLong := 0.0
	sumShort := 0.0
	openType := ""
	entryPrice := 0.0
	atrEntry := 0.0
	trail := 0.0
	for idx, s := range signals {
		if s.Type == "LONG" {
			longCount++
		} else {
			shortCount++
		}
		pctStr := ""
		if openType == "" {
			if s.Label == "ENTRY" {
				openType = s.Type
				entryPrice = s.Close
				atrEntry = math.Min(s.ATR, entryPrice*0.005)
				if openType == "LONG" {
					trail = entryPrice - atrEntry
				} else {
					trail = entryPrice + atrEntry
				}
			}
		} else {
			// Update trailing stop with fixed ATR offset from entry
			if openType == "LONG" {
				// trail cannot go down; follow price up with fixed offset = atrEntry
				candTrail := s.Close - atrEntry
				if candTrail > trail {
					trail = candTrail
				}
			} else { // SHORT
				// trail cannot go up; follow price down with fixed offset = atrEntry
				candTrail := s.Close + atrEntry
				if candTrail < trail {
					trail = candTrail
				}
			}

			// Check stop hit on this row's close
			stopped := false
			stopPrice := trail
			if openType == "LONG" {
				if s.Close <= trail {
					stopped = true
				}
			} else { // SHORT
				if s.Close >= trail {
					stopped = true
				}
			}

			if stopped {
				diff := stopPrice - entryPrice
				pctStr = fmt.Sprintf("%6.2f", diff)
				if openType == "LONG" {
					sumLong += diff
				} else {
					sumShort += diff
				}
				// flat now
				openType = ""
				// If this row is also an ENTRY, immediately open a new position using current values
				if s.Label == "ENTRY" {
					openType = s.Type
					entryPrice = s.Close
					atrEntry = math.Min(s.ATR, entryPrice*0.005)
					if openType == "LONG" {
						trail = entryPrice - atrEntry
					} else {
						trail = entryPrice + atrEntry
					}
				}
			} else if s.Label == "EXIT" && s.Type == openType {
				diff := s.Close - entryPrice
				pctStr = fmt.Sprintf("%6.2f", diff)
				if openType == "LONG" {
					sumLong += diff
				} else {
					sumShort += diff
				}
				openType = ""
			} else if s.Label == "ENTRY" && s.Type != openType {
				diff := s.Close - entryPrice
				pctStr = fmt.Sprintf("%6.2f", diff)
				if openType == "LONG" {
					sumLong += diff
				} else {
					sumShort += diff
				}
				// Ré-ouvre immédiatement avec le nouveau type (toujours sur ENTRY quand flat)
				openType = s.Type
				entryPrice = s.Close
				atrEntry = math.Min(s.ATR, entryPrice*0.005)
				if openType == "LONG" {
					trail = entryPrice - atrEntry
				} else {
					trail = entryPrice + atrEntry
				}
			}
		}
		fmt.Printf("%-4d | %s | %-6s | %-5s | %8.2f | %7.2f | %8.2f | %5.1f%% | %8.2f | %8.2f | %7s\n",
			idx+1,
			s.Timestamp.Format("2006-01-02 15:04:05"),
			s.Type,
			s.Label,
			s.Close,
			s.StochK,
			s.Body,
			s.BodyPctBar*100,
			s.ATR,
			s.BodyToATR,
			pctStr,
		)
	}

	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("Total signaux: %d | LONG=%d | SHORT=%d\n", len(signals), longCount, shortCount)
	fmt.Printf("Somme LONG: %.2f | Somme SHORT: %.2f | TOTAL directionnel: %.2f\n", sumLong, sumShort, sumLong+(-1.0*sumShort))
	fmt.Println(strings.Repeat("=", 150))
}

func safeDiv(a, b float64) float64 {
	if b == 0 || math.IsNaN(a) || math.IsNaN(b) {
		return math.NaN()
	}
	return a / b
}
