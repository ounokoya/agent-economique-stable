package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/indicators"
)

// ============================================================================
// PARAMETRES CONFIGURABLES
// ============================================================================

// Symbole et donnees
const (
	SYMBOL     = "SOL_USDT" // Format Gate.io
	TIMEFRAME  = "5m"
	NB_CANDLES = 1000
)

// Periodes VWMA
const (
	VWMA_RAPIDE = 12
	VWMA_LENT   = 36
	VWMA_STOP   = 72
)

// Periodes DMI
const (
	DMI_PERIODE = 10
	DMI_SMOOTH  = 3
)

// Calibrage ATR et validation
const (
	ATR_PERIODE           = 3
	GAMMA_GAP_VWMA        = 0.2 // 50% de ATR pour gap VWMA
	GAMMA_GAP_DI          = 2.0 // Gap minimal DI+ vs DI-
	GAMMA_GAP_DX          = 2.0 // Gap minimal DX vs ADX
	VOLATILITE_MIN        = 0.1 // 0.3% ATR% minimal
	WINDOW_GAMMA_VALIDATE = 6   // Fenetre validation gamma différée
	WINDOW_W              = 10  // Fenetre matching VWMA + DMI
)

// ============================================================================
// STRUCTURES
// ============================================================================

type Config struct {
	Symbol              string
	Timeframe           string
	NbCandles           int
	VwmaRapide          int
	VwmaLent            int
	VwmaStop            int
	DmiPeriode          int
	DmiSmooth           int
	AtrPeriode          int
	GammaGapVWMA        float64
	GammaGapDI          float64
	GammaGapDX          float64
	VolatiliteMin       float64
	WindowGammaValidate int
	WindowW             int
}

type SignalVWMA struct {
	Index           int
	Timestamp       time.Time
	Direction       string
	Prix            float64
	Vwma6           float64
	Vwma24          float64
	Gap             float64
	AtrPct          float64
	GapValide       bool
	GapValideBougie int
	VolatiliteOK    bool
	Valide          bool
	Raison          string
}

type SignalDMI struct {
	Index                int
	Timestamp            time.Time
	Direction            string
	DiPlus               float64
	DiMinus              float64
	Dx                   float64
	Adx                  float64
	GapDI                float64
	GapDIValide          bool
	GapDIValideBougie    int
	GapDXADX             float64
	GapDXADXValide       bool
	GapDXADXValideBougie int
	Valide               bool
	Raison               string
}

type SignalTrend struct {
	Index        int
	Timestamp    time.Time
	Direction    string
	VWMAIndex    int // Index du signal VWMA
	DMIIndex     int // Index du signal DMI
	HasVWMA      bool
	HasDMI       bool
	Matched      bool
	DistanceBars int // Distance en barres entre VWMA et DMI
	Motif        string
	Prix         float64
	Vwma6        float64
	DiPlus       float64
	DiMinus      float64
	Valide       bool
	Raison       string
}

// ============================================================================
// DETECTION SIGNAUX VWMA
// ============================================================================

func detectVWMASignals(config Config, klines []gateio.Kline, vwmaRapide, vwmaLent, atr []float64) []SignalVWMA {
	signals := []SignalVWMA{}
	startIdx := config.VwmaLent + 5

	for i := startIdx; i < len(klines)-1; i++ {
		cross, direction := indicators.DetecterCroisement(vwmaRapide, vwmaLent, i)
		if !cross {
			continue
		}

		signal := SignalVWMA{
			Index:     i,
			Timestamp: klines[i].OpenTime,
			Direction: direction,
			Prix:      klines[i].Close,
			Vwma6:     vwmaRapide[i],
			Vwma24:    vwmaLent[i],
		}

		// Validation gamma gap avec fenêtre différée
		gapInitial := indicators.CalculerEcart(vwmaRapide[i], vwmaLent[i])
		signal.Gap = gapInitial
		gammaGapValue := config.GammaGapVWMA * atr[i]
		signal.GapValide = gapInitial >= gammaGapValue
		signal.GapValideBougie = -1

		if signal.GapValide {
			signal.GapValideBougie = 0
		} else {
			for w := 1; w <= config.WindowGammaValidate; w++ {
				futureIdx := i + w
				if futureIdx >= len(klines)-1 {
					break
				}
				gapFuture := indicators.CalculerEcart(vwmaRapide[futureIdx], vwmaLent[futureIdx])
				gammaFuture := config.GammaGapVWMA * atr[futureIdx]
				if gapFuture >= gammaFuture {
					signal.GapValide = true
					signal.GapValideBougie = w
					break
				}
			}
		}

		// Validation volatilité
		atrPct := indicators.Normaliser(atr[i], signal.Prix)
		signal.AtrPct = atrPct
		signal.VolatiliteOK = atrPct >= config.VolatiliteMin

		// Status global
		signal.Valide = signal.GapValide && signal.VolatiliteOK

		if !signal.GapValide {
			signal.Raison = fmt.Sprintf("Gap jamais validé (fenêtre=%d)", config.WindowGammaValidate)
		} else if !signal.VolatiliteOK {
			signal.Raison = fmt.Sprintf("ATR%% %.2f%% < %.2f%%", atrPct, config.VolatiliteMin)
		} else {
			if signal.GapValideBougie == 0 {
				signal.Raison = "OK"
			} else {
				signal.Raison = fmt.Sprintf("OK (gamma validé +%d bougies)", signal.GapValideBougie)
			}
		}

		signals = append(signals, signal)
	}

	return signals
}

// ============================================================================
// DETECTION SIGNAUX DMI
// ============================================================================

func detectDMISignals(config Config, klines []gateio.Kline, diPlus, diMinus, dx, adx []float64) []SignalDMI {
	signals := []SignalDMI{}
	startIdx := config.DmiPeriode + config.DmiSmooth + 5

	for i := startIdx; i < len(klines)-1; i++ {
		crossDI, directionDI := indicators.DetecterCroisement(diPlus, diMinus, i)
		if !crossDI {
			continue
		}

		signal := SignalDMI{
			Index:     i,
			Timestamp: klines[i].OpenTime,
			Direction: directionDI,
			DiPlus:    diPlus[i],
			DiMinus:   diMinus[i],
			Dx:        dx[i],
			Adx:       adx[i],
		}

		// Validation gamma gap DI avec fenêtre différée
		gapDIInitial := indicators.CalculerEcart(diPlus[i], diMinus[i])
		signal.GapDI = gapDIInitial
		signal.GapDIValide = gapDIInitial >= config.GammaGapDI
		signal.GapDIValideBougie = -1

		if signal.GapDIValide {
			signal.GapDIValideBougie = 0
		} else {
			for w := 1; w <= config.WindowGammaValidate; w++ {
				futureIdx := i + w
				if futureIdx >= len(klines)-1 {
					break
				}
				gapDIFuture := indicators.CalculerEcart(diPlus[futureIdx], diMinus[futureIdx])
				if gapDIFuture >= config.GammaGapDI {
					signal.GapDIValide = true
					signal.GapDIValideBougie = w
					break
				}
			}
		}

		// Validation croisement DX/ADX dans fenêtre
		signal.GapDXADXValide = false
		signal.GapDXADXValideBougie = -1

		for w := 0; w <= config.WindowGammaValidate; w++ {
			futureIdx := i + w
			if futureIdx >= len(klines)-1 {
				break
			}

			crossDXADX, directionDXADX := indicators.DetecterCroisement(dx, adx, futureIdx)
			if crossDXADX && directionDXADX == "HAUSSIER" {
				gapDXADX := indicators.CalculerEcart(dx[futureIdx], adx[futureIdx])
				signal.GapDXADX = gapDXADX

				if gapDXADX >= config.GammaGapDX {
					signal.GapDXADXValide = true
					signal.GapDXADXValideBougie = w
					break
				} else {
					for w2 := 1; w2 <= config.WindowGammaValidate; w2++ {
						futureIdx2 := futureIdx + w2
						if futureIdx2 >= len(klines)-1 {
							break
						}
						gapDXADXFuture := indicators.CalculerEcart(dx[futureIdx2], adx[futureIdx2])
						if gapDXADXFuture >= config.GammaGapDX {
							signal.GapDXADXValide = true
							signal.GapDXADXValideBougie = w + w2
							signal.GapDXADX = gapDXADXFuture
							break
						}
					}
					if signal.GapDXADXValide {
						break
					}
				}
			}
		}

		// Status global
		signal.Valide = signal.GapDIValide && signal.GapDXADXValide

		if !signal.GapDIValide {
			signal.Raison = fmt.Sprintf("Gap DI jamais validé (fenêtre=%d)", config.WindowGammaValidate)
		} else if !signal.GapDXADXValide {
			signal.Raison = "DX/ADX: Croisement non trouvé ou gap jamais validé"
		} else {
			signal.Raison = fmt.Sprintf("OK (DI validé +%d, DX/ADX validé +%d bougies)",
				signal.GapDIValideBougie, signal.GapDXADXValideBougie)
		}

		signals = append(signals, signal)
	}

	return signals
}

// ============================================================================
// MATCHING VWMA + DMI
// ============================================================================

func matchSignals(config Config, klines []gateio.Kline, vwmaSignals []SignalVWMA, dmiSignals []SignalDMI) []SignalTrend {
	trendSignals := []SignalTrend{}

	// Pour chaque signal VWMA valide, chercher un signal DMI dans la fenêtre W
	// Assurer un matching 1-1: chaque DMI ne peut être utilisé qu'une fois
	usedDMI := make([]bool, len(dmiSignals))
	for vIdx, vwma := range vwmaSignals {
		if !vwma.Valide {
			continue
		}

		matched := false
		var matchedDMI *SignalDMI
		var dIdx int
		minDistance := config.WindowW + 1

		// Chercher DMI dans fenêtre W (avant ou après VWMA)
		for dmiIdx, dmi := range dmiSignals {
			if usedDMI[dmiIdx] {
				continue
			}
			if !dmi.Valide {
				continue
			}

			// Même direction requise
			if vwma.Direction != dmi.Direction {
				continue
			}

			// Distance en barres
			distance := 0
			if dmi.Index >= vwma.Index {
				distance = dmi.Index - vwma.Index
			} else {
				distance = vwma.Index - dmi.Index
			}

			// Dans fenêtre W ?
			if distance <= config.WindowW && distance < minDistance {
				matched = true
				matchedDMI = &dmi
				dIdx = dmiIdx
				minDistance = distance
			}
		}

		if matched && matchedDMI != nil {
			// Marquer le DMI comme utilisé pour éviter les doublons dans les logs
			usedDMI[dIdx] = true
			// Déterminer qui est arrivé en premier
			motif := ""
			sigTimestamp := vwma.Timestamp
			sigIndex := vwma.Index

			if vwma.Index < matchedDMI.Index {
				motif = fmt.Sprintf("VWMA→DMI (+%d bars)", matchedDMI.Index-vwma.Index)
			} else if vwma.Index > matchedDMI.Index {
				motif = fmt.Sprintf("DMI→VWMA (+%d bars)", vwma.Index-matchedDMI.Index)
				sigTimestamp = matchedDMI.Timestamp
				sigIndex = matchedDMI.Index
			} else {
				motif = "VWMA+DMI simultané"
			}

			trendSignals = append(trendSignals, SignalTrend{
				Index:        sigIndex,
				Timestamp:    sigTimestamp,
				Direction:    vwma.Direction,
				VWMAIndex:    vIdx,
				DMIIndex:     dIdx,
				HasVWMA:      true,
				HasDMI:       true,
				Matched:      true,
				DistanceBars: minDistance,
				Motif:        motif,
				Prix:         vwma.Prix,
				Vwma6:        vwma.Vwma6,
				DiPlus:       matchedDMI.DiPlus,
				DiMinus:      matchedDMI.DiMinus,
				Valide:       true,
				Raison:       fmt.Sprintf("Matché: %s", motif),
			})
		}
	}

	// Ordonner chronologiquement par Timestamp
	sort.Slice(trendSignals, func(i, j int) bool {
		return trendSignals[i].Timestamp.Before(trendSignals[j].Timestamp)
	})
	return trendSignals
}

// ============================================================================
// AFFICHAGE
// ============================================================================

func displayVWMASignals(signals []SignalVWMA) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("SIGNAUX VWMA DETECTES")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-19s | %-6s | %-8s | %-8s | %-8s | %-7s | %-10s | %-10s | %-50s\n",
		"#", "Date/Heure", "Dir", "Prix", "VWMA6", "Gap", "ATR%", "Gamma", "Status", "Raison")
	fmt.Println(strings.Repeat("-", 140))

	validCount := 0
	for idx, sig := range signals {
		statusGamma := "✗"
		if sig.GapValide {
			if sig.GapValideBougie == 0 {
				statusGamma = "✓ imm"
			} else {
				statusGamma = fmt.Sprintf("✓ +%d", sig.GapValideBougie)
			}
		}

		status := "✗ REJET"
		if sig.Valide {
			status = "✓ OK"
			validCount++
		}

		fmt.Printf("%-4d | %s | %-6s | %8.2f | %8.2f | %8.4f | %6.2f%% | %-10s | %-10s | %-50s\n",
			idx+1, sig.Timestamp.Format("2006-01-02 15:04:05"), sig.Direction,
			sig.Prix, sig.Vwma6, sig.Gap, sig.AtrPct, statusGamma, status, sig.Raison)
	}

	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("Total VWMA: %d | Valides: %d | Rejetes: %d\n", len(signals), validCount, len(signals)-validCount)
	fmt.Println(strings.Repeat("=", 140))
}

func displayDMISignals(signals []SignalDMI) {
	fmt.Println("\n" + strings.Repeat("=", 150))
	fmt.Println("SIGNAUX DMI DETECTES")
	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("%-4s | %-19s | %-6s | %-7s | %-7s | %-12s | %-12s | %-10s | %-50s\n",
		"#", "Date/Heure", "Dir", "DI+", "DI-", "GapDI", "GapDX/ADX", "Status", "Raison")
	fmt.Println(strings.Repeat("-", 150))

	validCount := 0
	for idx, sig := range signals {
		statusDI := "✗"
		if sig.GapDIValide {
			if sig.GapDIValideBougie == 0 {
				statusDI = "✓ imm"
			} else {
				statusDI = fmt.Sprintf("✓ +%d", sig.GapDIValideBougie)
			}
		}

		statusDXADX := "✗"
		if sig.GapDXADXValide {
			if sig.GapDXADXValideBougie == 0 {
				statusDXADX = "✓ imm"
			} else {
				statusDXADX = fmt.Sprintf("✓ +%d", sig.GapDXADXValideBougie)
			}
		}

		status := "✗ REJET"
		if sig.Valide {
			status = "✓ OK"
			validCount++
		}

		fmt.Printf("%-4d | %s | %-6s | %7.2f | %7.2f | %-12s | %-12s | %-10s | %-50s\n",
			idx+1, sig.Timestamp.Format("2006-01-02 15:04:05"), sig.Direction,
			sig.DiPlus, sig.DiMinus, statusDI, statusDXADX, status, sig.Raison)
	}

	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("Total DMI: %d | Valides: %d | Rejetes: %d\n", len(signals), validCount, len(signals)-validCount)
	fmt.Println(strings.Repeat("=", 150))
}

func displayTrendSignals(signals []SignalTrend) {
	fmt.Println("\n" + strings.Repeat("=", 150))
	fmt.Println("SIGNAUX TREND (MATCHING VWMA + DMI)")
	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("%-4s | %-19s | %-6s | %-8s | %-8s | %-7s | %-7s | %-8s | %-40s | %-10s\n",
		"#", "Date/Heure", "Dir", "Prix", "VWMA6", "DI+", "DI-", "Distance", "Motif", "Status")
	fmt.Println(strings.Repeat("-", 150))

	for idx, sig := range signals {
		status := "✓ MATCHE"

		fmt.Printf("%-4d | %s | %-6s | %8.2f | %8.2f | %7.2f | %7.2f | %8d | %-40s | %-10s\n",
			idx+1, sig.Timestamp.Format("2006-01-02 15:04:05"), sig.Direction,
			sig.Prix, sig.Vwma6, sig.DiPlus, sig.DiMinus, sig.DistanceBars, sig.Motif, status)
	}

	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("Total Signaux TREND: %d\n", len(signals))

	// Comptage par direction
	longCount := 0
	shortCount := 0
	for _, sig := range signals {
		if sig.Direction == "HAUSSIER" {
			longCount++
		} else {
			shortCount++
		}
	}
	fmt.Printf("Details: LONG=%d | SHORT=%d\n", longCount, shortCount)

	// Analyse motifs
	vwmaFirst := 0
	dmiFirst := 0
	simultane := 0
	for _, sig := range signals {
		if strings.Contains(sig.Motif, "VWMA→DMI") {
			vwmaFirst++
		} else if strings.Contains(sig.Motif, "DMI→VWMA") {
			dmiFirst++
		} else {
			simultane++
		}
	}
	fmt.Printf("Motifs: VWMA→DMI=%d | DMI→VWMA=%d | Simultané=%d\n", vwmaFirst, dmiFirst, simultane)
	fmt.Println(strings.Repeat("=", 150))
}

func displayStatistics(vwmaSignals []SignalVWMA, dmiSignals []SignalDMI, trendSignals []SignalTrend) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES GLOBALES")
	fmt.Println(strings.Repeat("=", 100))

	vwmaValid := 0
	dmiValid := 0
	for _, s := range vwmaSignals {
		if s.Valide {
			vwmaValid++
		}
	}
	for _, s := range dmiSignals {
		if s.Valide {
			dmiValid++
		}
	}

	fmt.Printf("Croisements VWMA détectés  : %d (valides: %d, rejetés: %d)\n",
		len(vwmaSignals), vwmaValid, len(vwmaSignals)-vwmaValid)
	fmt.Printf("Croisements DMI détectés   : %d (valides: %d, rejetés: %d)\n",
		len(dmiSignals), dmiValid, len(dmiSignals)-dmiValid)
	fmt.Printf("Signaux TREND matchés      : %d\n", len(trendSignals))

	if vwmaValid > 0 && dmiValid > 0 {
		matchRate := float64(len(trendSignals)) / float64(vwmaValid) * 100
		fmt.Printf("Taux de matching           : %.1f%% (sur VWMA valides)\n", matchRate)
	}

	fmt.Println(strings.Repeat("=", 100))
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	config := Config{
		Symbol:              SYMBOL,
		Timeframe:           TIMEFRAME,
		NbCandles:           NB_CANDLES,
		VwmaRapide:          VWMA_RAPIDE,
		VwmaLent:            VWMA_LENT,
		VwmaStop:            VWMA_STOP,
		DmiPeriode:          DMI_PERIODE,
		DmiSmooth:           DMI_SMOOTH,
		AtrPeriode:          ATR_PERIODE,
		GammaGapVWMA:        GAMMA_GAP_VWMA,
		GammaGapDI:          GAMMA_GAP_DI,
		GammaGapDX:          GAMMA_GAP_DX,
		VolatiliteMin:       VOLATILITE_MIN,
		WindowGammaValidate: WINDOW_GAMMA_VALIDATE,
		WindowW:             WINDOW_W,
	}

	fmt.Println("=== DEMO STRATEGIE TREND (VWMA + DMI) ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange           : Gate.io\n")
	fmt.Printf("Symbole            : %s\n", config.Symbol)
	fmt.Printf("Timeframe          : %s\n", config.Timeframe)
	fmt.Printf("VWMA rapide/lent   : %d / %d\n", config.VwmaRapide, config.VwmaLent)
	fmt.Printf("DMI periode/smooth : %d / %d\n", config.DmiPeriode, config.DmiSmooth)
	fmt.Printf("Gamma gap VWMA     : %.2f x ATR\n", config.GammaGapVWMA)
	fmt.Printf("Gamma gap DI       : %.1f\n", config.GammaGapDI)
	fmt.Printf("Gamma gap DX       : %.1f\n", config.GammaGapDX)
	fmt.Printf("Fenetre gamma      : %d bougies\n", config.WindowGammaValidate)
	fmt.Printf("Fenetre W (match)  : %d bougies\n", config.WindowW)
	fmt.Printf("Volatilite min     : %.2f%%\n", config.VolatiliteMin)

	// Récupération des données
	fmt.Printf("\nRecuperation des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, config.Symbol, config.Timeframe, config.NbCandles)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}

	// Trier chronologiquement
	for i := 0; i < len(klines); i++ {
		for j := i + 1; j < len(klines); j++ {
			if klines[j].OpenTime.Before(klines[i].OpenTime) {
				klines[i], klines[j] = klines[j], klines[i]
			}
		}
	}

	fmt.Printf("Klines recuperees: %d\n", len(klines))

	// Préparation données
	fmt.Println("\nCalcul des indicateurs...")

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

	// Calcul VWMA
	vwmaRapideIndicator := indicators.NewVWMATVStandard(config.VwmaRapide)
	vwmaLentIndicator := indicators.NewVWMATVStandard(config.VwmaLent)
	vwmaRapide := vwmaRapideIndicator.Calculate(closes, volumes)
	vwmaLent := vwmaLentIndicator.Calculate(closes, volumes)

	// Calcul ATR
	atrIndicator := indicators.NewATRTVStandard(config.AtrPeriode)
	atr := atrIndicator.Calculate(highs, lows, closes)

	// Calcul DMI
	dmi := indicators.NewDMITVStandard(config.DmiPeriode)
	diPlus, diMinus, adx := dmi.Calculate(highs, lows, closes)

	// Calcul DX
	dx := make([]float64, len(diPlus))
	for i := range diPlus {
		sum := diPlus[i] + diMinus[i]
		if sum != 0 {
			dx[i] = indicators.CalculerEcart(diPlus[i], diMinus[i]) / sum * 100
		}
	}

	// Détection signaux
	fmt.Println("Detection des signaux VWMA...")
	vwmaSignals := detectVWMASignals(config, klines, vwmaRapide, vwmaLent, atr)

	fmt.Println("Detection des signaux DMI...")
	dmiSignals := detectDMISignals(config, klines, diPlus, diMinus, dx, adx)

	fmt.Println("Matching VWMA + DMI...")
	trendSignals := matchSignals(config, klines, vwmaSignals, dmiSignals)

	// Affichage
	displayVWMASignals(vwmaSignals)
	displayDMISignals(dmiSignals)
	displayTrendSignals(trendSignals)
	displayStatistics(vwmaSignals, dmiSignals, trendSignals)

	fmt.Println("\n=== FIN DEMO ===")
}
