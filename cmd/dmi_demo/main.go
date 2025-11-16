package main

import (
	"context"
	"fmt"
	"log"
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
	NB_CANDLES = 500
)

// Periodes DMI
const (
	DMI_PERIODE = 14 // Période DI+, DI-, DX
	DMI_SMOOTH  = 6  // Lissage ADX
	ATR_PERIODE = 14 // Période ATR
)

// Calibrage validation
const (
	GAMMA_GAP_DI          = 5.0 // Écart minimal DI+ vs DI- (scalping: 5-8, invest: 8-12)
	GAMMA_GAP_DX          = 5.0 // Écart minimal DX vs ADX
	WINDOW_GAMMA_VALIDATE = 5   // Fenetre validation gamma différée
)

// Thresholds
const (
	DI_MIN = 20.0 // DI minimal pour dominance
	DX_MIN = 25.0 // DX minimal pour force tendance
)

// ============================================================================
// STRUCTURES
// ============================================================================

type Config struct {
	Symbol              string
	Timeframe           string
	NbCandles           int
	DmiPeriode          int
	DmiSmooth           int
	AtrPeriode          int
	GammaGapDI          float64
	GammaGapDX          float64
	WindowGammaValidate int
	DiMin               float64
	DxMin               float64
}

type SignalDMI struct {
	Index     int
	Timestamp time.Time
	Type      string // "TENDANCE_HAUSSIER", "TENDANCE_BAISSIER", "MOMENTUM"
	Direction string // "LONG", "SHORT"
	DiPlus    float64
	DiMinus   float64
	Dx        float64
	Adx       float64

	// Validation croisement DI
	CrossDI           bool
	GapDI             float64
	GapDIValide       bool
	GapDIValideBougie int

	// Validation croisement DX/ADX
	CrossDXADX           bool
	GapDXADX             float64
	GapDXADXValide       bool
	GapDXADXValideBougie int

	// Status global
	Valide bool
	Raison string
}

// ============================================================================
// DETECTION SIGNAUX DMI
// ============================================================================

type DMIDetector struct {
	config Config
}

func NewDMIDetector(config Config) *DMIDetector {
	return &DMIDetector{config: config}
}

func (d *DMIDetector) detectSignals(klines []gateio.Kline, diPlus, diMinus, dx, adx []float64) []SignalDMI {
	signals := []SignalDMI{}

	// Index de départ (besoin de données suffisantes)
	startIdx := d.config.DmiPeriode + d.config.DmiSmooth + 5

	// IMPORTANT: On s'arrête à len(klines)-1 car la dernière barre est courante (non fermée)
	for i := startIdx; i < len(klines)-1; i++ {

		// ========================================
		// 1. DETECTION CROISEMENT DI+ / DI-
		// ========================================
		crossDI, directionDI := indicators.DetecterCroisement(diPlus, diMinus, i)

		if crossDI {
			signal := SignalDMI{
				Index:     i,
				Timestamp: klines[i].OpenTime,
				CrossDI:   true,
				DiPlus:    diPlus[i],
				DiMinus:   diMinus[i],
				Dx:        dx[i],
				Adx:       adx[i],
			}

			// Déterminer direction
			if directionDI == "HAUSSIER" {
				signal.Direction = "LONG"
				signal.Type = "TENDANCE_HAUSSIER"
			} else {
				signal.Direction = "SHORT"
				signal.Type = "TENDANCE_BAISSIER"
			}

			// ========================================
			// 2. VALIDATION GAMMA GAP DI (avec fenêtre différée)
			// ========================================
			gapDIInitial := indicators.CalculerEcart(diPlus[i], diMinus[i])
			signal.GapDI = gapDIInitial
			signal.GapDIValide = gapDIInitial >= d.config.GammaGapDI
			signal.GapDIValideBougie = -1

			if signal.GapDIValide {
				// Validé immédiatement
				signal.GapDIValideBougie = 0
			} else {
				// Fenêtre de validation différée
				for w := 1; w <= d.config.WindowGammaValidate; w++ {
					futureIdx := i + w
					if futureIdx >= len(klines)-1 {
						break
					}

					gapDIFuture := indicators.CalculerEcart(diPlus[futureIdx], diMinus[futureIdx])
					if gapDIFuture >= d.config.GammaGapDI {
						signal.GapDIValide = true
						signal.GapDIValideBougie = w
						break
					}
				}
			}

			// ========================================
			// 3. VALIDATION CROISEMENT DX / ADX (dans fenêtre)
			// ========================================
			signal.CrossDXADX = false
			signal.GapDXADXValide = false
			signal.GapDXADXValideBougie = -1

			// Chercher croisement DX/ADX dans la fenêtre suivant le croisement DI
			for w := 0; w <= d.config.WindowGammaValidate; w++ {
				futureIdx := i + w
				if futureIdx >= len(klines)-1 {
					break
				}

				crossDXADX, directionDXADX := indicators.DetecterCroisement(dx, adx, futureIdx)

				if crossDXADX && directionDXADX == "HAUSSIER" {
					signal.CrossDXADX = true
					gapDXADX := indicators.CalculerEcart(dx[futureIdx], adx[futureIdx])
					signal.GapDXADX = gapDXADX

					// Validation gamma gap DX/ADX avec fenêtre différée
					if gapDXADX >= d.config.GammaGapDX {
						signal.GapDXADXValide = true
						signal.GapDXADXValideBougie = w
						break
					} else {
						// Fenêtre de validation différée pour DX/ADX
						for w2 := 1; w2 <= d.config.WindowGammaValidate; w2++ {
							futureIdx2 := futureIdx + w2
							if futureIdx2 >= len(klines)-1 {
								break
							}

							gapDXADXFuture := indicators.CalculerEcart(dx[futureIdx2], adx[futureIdx2])
							if gapDXADXFuture >= d.config.GammaGapDX {
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

			// ========================================
			// 4. VALIDATION FINALE
			// ========================================
			signal.Valide = signal.GapDIValide && signal.GapDXADXValide

			if !signal.GapDIValide {
				signal.Raison = fmt.Sprintf("Gap DI jamais validé (fenêtre=%d)", d.config.WindowGammaValidate)
			} else if !signal.GapDXADXValide {
				signal.Raison = "DX/ADX: Croisement non trouvé ou gap jamais validé"
			} else {
				if signal.GapDIValideBougie == 0 && signal.GapDXADXValideBougie == 0 {
					signal.Raison = "OK (validé immédiatement)"
				} else {
					signal.Raison = fmt.Sprintf("OK (DI validé +%d, DX/ADX validé +%d bougies)",
						signal.GapDIValideBougie, signal.GapDXADXValideBougie)
				}
			}

			signals = append(signals, signal)
		}

		// ========================================
		// 5. DETECTION MODE MOMENTUM (alternative)
		// ========================================
		// DX ou ADX croise au-dessus d'un DI
		// Pour l'instant on se concentre sur le mode tendance
	}

	return signals
}

// ============================================================================
// AFFICHAGE
// ============================================================================

func displaySignals(signals []SignalDMI) {
	fmt.Println("\n" + strings.Repeat("=", 160))
	fmt.Println("TOUS LES SIGNAUX DMI DETECTES (MODE TENDANCE)")
	fmt.Println(strings.Repeat("=", 160))
	fmt.Printf("%-4s | %-19s | %-6s | %-8s | %-7s | %-7s | %-7s | %-7s | %-12s | %-12s | %-10s | %-50s\n",
		"#", "Date/Heure (Open)", "Dir", "Type", "DI+", "DI-", "DX", "ADX", "GapDI", "GapDX/ADX", "Status", "Raison")
	fmt.Println(strings.Repeat("-", 160))

	validCount := 0
	rejectCount := 0
	longCount := 0
	shortCount := 0

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
			if sig.Direction == "LONG" {
				longCount++
			} else {
				shortCount++
			}
		} else {
			rejectCount++
		}

		typeShort := "TEND"
		if sig.Type == "MOMENTUM" {
			typeShort = "MOM"
		}

		fmt.Printf("%-4d | %s | %-6s | %-8s | %7.2f | %7.2f | %7.2f | %7.2f | %-12s | %-12s | %-10s | %-50s\n",
			idx+1,
			sig.Timestamp.Format("2006-01-02 15:04:05"),
			sig.Direction,
			typeShort,
			sig.DiPlus,
			sig.DiMinus,
			sig.Dx,
			sig.Adx,
			statusDI,
			statusDXADX,
			status,
			sig.Raison)
	}

	fmt.Println(strings.Repeat("=", 160))
	fmt.Printf("Total signaux: %d | Valides: %d | Rejetes: %d\n", len(signals), validCount, rejectCount)
	fmt.Printf("Details valides: LONG=%d | SHORT=%d\n", longCount, shortCount)
	fmt.Println(strings.Repeat("=", 160))
}

func analyzeCurrentPosition(klines []gateio.Kline, diPlus, diMinus, dx, adx []float64) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("ANALYSE POSITION ACTUELLE (10 dernieres bougies)")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-8s | %-7s | %-7s | %-7s | %-7s | %-15s | %-12s\n",
		"Heure", "DI+", "DI-", "DX", "ADX", "Dominance", "Position DX/ADX")
	fmt.Println(strings.Repeat("-", 100))

	startIdx := len(klines) - 11
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(klines)-1; i++ {
		dominance := "NEUTRE"
		if diPlus[i] > diMinus[i] {
			dominance = "HAUSSIERE (DI+)"
		} else if diMinus[i] > diPlus[i] {
			dominance = "BAISSIERE (DI-)"
		}

		positionDX := "DX EN-DESSOUS"
		if dx[i] > adx[i] {
			positionDX = "DX AU-DESSUS"
		}

		fmt.Printf("%s | %7.2f | %7.2f | %7.2f | %7.2f | %-15s | %-12s\n",
			klines[i].OpenTime.Format("15:04:05"),
			diPlus[i],
			diMinus[i],
			dx[i],
			adx[i],
			dominance,
			positionDX)
	}
	fmt.Println(strings.Repeat("=", 100))
}

func displayStatistics(signals []SignalDMI) {
	validSignals := []SignalDMI{}
	for _, sig := range signals {
		if sig.Valide {
			validSignals = append(validSignals, sig)
		}
	}

	if len(validSignals) == 0 {
		fmt.Println("\n=== Aucun signal valide ===")
		return
	}

	// Calcul moyennes
	sumGapDI := 0.0
	sumGapDXADX := 0.0
	longCount := 0
	shortCount := 0

	for _, sig := range validSignals {
		sumGapDI += sig.GapDI
		sumGapDXADX += sig.GapDXADX
		if sig.Direction == "LONG" {
			longCount++
		} else {
			shortCount++
		}
	}

	avgGapDI := sumGapDI / float64(len(validSignals))
	avgGapDXADX := sumGapDXADX / float64(len(validSignals))

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES (Signaux valides uniquement)")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Total: %d | LONG: %d (%.1f%%) | SHORT: %d (%.1f%%)\n",
		len(validSignals),
		longCount, float64(longCount)/float64(len(validSignals))*100,
		shortCount, float64(shortCount)/float64(len(validSignals))*100)
	fmt.Printf("Moyennes: GapDI=%.2f | GapDX/ADX=%.2f\n", avgGapDI, avgGapDXADX)
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
		DmiPeriode:          DMI_PERIODE,
		DmiSmooth:           DMI_SMOOTH,
		AtrPeriode:          ATR_PERIODE,
		GammaGapDI:          GAMMA_GAP_DI,
		GammaGapDX:          GAMMA_GAP_DX,
		WindowGammaValidate: WINDOW_GAMMA_VALIDATE,
		DiMin:               DI_MIN,
		DxMin:               DX_MIN,
	}

	fmt.Println("=== DEMO STRATEGIE DMI ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange        : Gate.io\n")
	fmt.Printf("Symbole         : %s\n", config.Symbol)
	fmt.Printf("Timeframe       : %s\n", config.Timeframe)
	fmt.Printf("DMI periode     : %d\n", config.DmiPeriode)
	fmt.Printf("DMI smooth      : %d (ADX)\n", config.DmiSmooth)
	fmt.Printf("Gamma gap DI    : %.1f\n", config.GammaGapDI)
	fmt.Printf("Gamma gap DX    : %.1f\n", config.GammaGapDX)
	fmt.Printf("Fenetre gamma   : %d bougies\n", config.WindowGammaValidate)
	fmt.Printf("DI min          : %.1f\n", config.DiMin)
	fmt.Printf("DX min          : %.1f\n", config.DxMin)

	// Récupération des données
	fmt.Printf("\nRecuperation des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, config.Symbol, config.Timeframe, config.NbCandles)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}

	// Trier chronologiquement (Gate.io retourne dans l'ordre inverse)
	for i := 0; i < len(klines); i++ {
		for j := i + 1; j < len(klines); j++ {
			if klines[j].OpenTime.Before(klines[i].OpenTime) {
				klines[i], klines[j] = klines[j], klines[i]
			}
		}
	}

	fmt.Printf("Klines recuperees depuis Gate.io: %d\n", len(klines))

	// Préparation données pour calcul DMI
	fmt.Println("\nCalcul des indicateurs DMI...")

	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	closes := make([]float64, len(klines))

	for i, k := range klines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
	}

	// Calcul DMI
	dmi := indicators.NewDMITVStandard(config.DmiPeriode)
	diPlus, diMinus, adx := dmi.Calculate(highs, lows, closes)

	// Calcul DX (à partir de DI+ et DI-)
	dx := make([]float64, len(diPlus))
	for i := range diPlus {
		sum := diPlus[i] + diMinus[i]
		if sum != 0 {
			dx[i] = indicators.CalculerEcart(diPlus[i], diMinus[i]) / sum * 100
		}
	}

	// Détection signaux
	fmt.Println("Detection des signaux DMI...\n")
	detector := NewDMIDetector(config)
	signals := detector.detectSignals(klines, diPlus, diMinus, dx, adx)

	// Affichage
	displaySignals(signals)
	analyzeCurrentPosition(klines, diPlus, diMinus, dx, adx)
	displayStatistics(signals)

	fmt.Println("\n=== FIN DEMO ===")
}
