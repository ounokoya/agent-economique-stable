package main

import (
	"context"
	"fmt"
	"log"
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

// Periodes VWMA
const (
	VWMA_RAPIDE = 6
	VWMA_LENT   = 24
	VWMA_STOP   = 72
)

// Calibrage
const (
	GAMMA_GAP      = 0.5 // 15% de ATR pour gap minimal
	ATR_PERIODE    = 30  // Periode ATR
	TAU_SLOPE      = 5.0 // Seuil variation pente
	VOLATILITE_MIN = 0.3 // 0.15% ATR% minimal
)

// Stop management
const (
	K_STOP = 1.2  // Multiplicateur ATR pour distance stop
	P_MIN  = 0.20 // Distance stop minimale (%)
	P_MAX  = 1.20 // Distance stop maximale (%)
)

// Fenetre validation
const (
	WINDOW_W              = 10 // Taille fenetre validation (non utilisé pour gamma)
	WINDOW_GAMMA_VALIDATE = 5  // Fenetre pour validation gamma différée (bougies)
)

// ============================================================================
// STRUCTURES
// ============================================================================

// Config contient tous les parametres configurables
type Config struct {
	// Symbole et timeframe
	Symbol    string
	Timeframe string
	NbCandles int

	// Periodes VWMA
	VWMARapide int
	VWMALent   int
	VWMAStop   int

	// Calibrage
	GammaGap      float64 // Ecart minimal pour croisement (ex: 0.15 * ATR)
	ATRPeriode    int
	TauSlope      float64 // Seuil pente
	VolatiliteMin float64 // ATR% minimal

	// Stop management
	KStop float64 // Multiplicateur ATR pour stop
	PMin  float64 // Distance stop min %
	PMax  float64 // Distance stop max %

	// Fenetre validation
	WindowW int // Taille fenetre validation
}

// VWMAStrategy gere la strategie VWMA
type VWMAStrategy struct {
	config Config

	// Indicateurs
	vwmaRapide *indicators.VWMATVStandard
	vwmaLent   *indicators.VWMATVStandard
	vwmaStop   *indicators.VWMATVStandard
	atr        *indicators.ATRTVStandard
}

// NewVWMAStrategy cree une nouvelle strategie
func NewVWMAStrategy(config Config) *VWMAStrategy {
	return &VWMAStrategy{
		config:     config,
		vwmaRapide: indicators.NewVWMATVStandard(config.VWMARapide),
		vwmaLent:   indicators.NewVWMATVStandard(config.VWMALent),
		vwmaStop:   indicators.NewVWMATVStandard(config.VWMAStop),
		atr:        indicators.NewATRTVStandard(config.ATRPeriode),
	}
}

// Signal represente un signal de trading
type Signal struct {
	Index           int
	Timestamp       time.Time
	Type            string // "LONG" ou "SHORT"
	Prix            float64
	VWMARapide      float64
	VWMALent        float64
	VWMAStop        float64
	Gap             float64
	GapValide       bool
	GapValideBougie int // 0=immédiat, >0=validé après N bougies, -1=jamais
	ATR             float64
	ATRPct          float64
	VolatiliteOK    bool
	Stop            float64
	DistanceStopPct float64
	Valide          bool   // Signal complet valide
	Raison          string // Raison de rejet si non valide
}

// NOTE: Toutes les operations analytiques sont desormais dans internal/indicators/helpers.go

// Calculer les indicateurs
func (s *VWMAStrategy) calculateIndicators(klines []gateio.Kline) ([]float64, []float64, []float64, []float64) {
	// Preparer les donnees
	n := len(klines)
	high := make([]float64, n)
	low := make([]float64, n)
	close := make([]float64, n)
	volume := make([]float64, n)

	for i := range klines {
		high[i] = klines[i].High
		low[i] = klines[i].Low
		close[i] = klines[i].Close
		volume[i] = klines[i].Volume
	}

	// Calculer les indicateurs
	vwmaRapideValues := s.vwmaRapide.Calculate(close, volume)
	vwmaLentValues := s.vwmaLent.Calculate(close, volume)
	vwmaStopValues := s.vwmaStop.Calculate(close, volume)
	atrValues := s.atr.Calculate(high, low, close)

	return vwmaRapideValues, vwmaLentValues, vwmaStopValues, atrValues
}

// Detecter les signaux VWMA (TOUS les croisements)
func (s *VWMAStrategy) detectSignals(klines []gateio.Kline, vwmaRapide, vwmaLent, vwmaStop, atr []float64) []Signal {
	var signals []Signal

	// Commencer apres periode de warmup
	startIdx := s.config.VWMALent + s.config.ATRPeriode + 10

	// IMPORTANT: On s'arrête à len(klines)-1 car en trading réel,
	// la dernière barre (klines[len-1]) est la barre COURANTE non fermée
	// On ne peut détecter des croisements que sur les barres FERMÉES
	for i := startIdx; i < len(klines)-1; i++ {
		// 1. Detection croisement entre n-2 et n-1 (barres fermées)
		cross, direction := indicators.DetecterCroisement(vwmaRapide, vwmaLent, i)

		if !cross {
			continue // Pas de croisement
		}

		// 2. Calcul gap au moment du croisement
		gapInitial := indicators.CalculerEcart(vwmaRapide[i], vwmaLent[i])

		// 3. Validation gap avec gamma ET fenetre de confirmation
		gammaGapValue := s.config.GammaGap * atr[i]
		gapValide := gapInitial >= gammaGapValue
		gapValideBougie := -1 // -1 = jamais validé

		if gapValide {
			// Validé immédiatement au croisement
			gapValideBougie = 0
		} else {
			// Pas validé immédiatement, vérifier dans la fenêtre (barres fermées uniquement)
			windowSize := WINDOW_GAMMA_VALIDATE
			for w := 1; w <= windowSize; w++ {
				futureIdx := i + w
				// On s'arrête avant la barre courante (len-1)
				if futureIdx >= len(klines)-1 {
					break // Fin des barres fermées
				}

				// Calculer gap et seuil gamma à chaque bougie future (fermée)
				gapFuture := indicators.CalculerEcart(vwmaRapide[futureIdx], vwmaLent[futureIdx])
				gammaFuture := s.config.GammaGap * atr[futureIdx]

				if gapFuture >= gammaFuture {
					// Gap validé après W bougies !
					gapValide = true
					gapValideBougie = w
					break
				}
			}
		}

		// 4. Validation volatilite
		prix := klines[i].Close
		atrPct := indicators.Normaliser(atr[i], prix)
		volatiliteOK := atrPct >= s.config.VolatiliteMin

		// 5. Calcul stop
		distanceStopPct := indicators.Clip(s.config.KStop*atrPct, s.config.PMin, s.config.PMax)

		var stop float64
		if direction == "HAUSSIER" {
			stop = vwmaStop[i] * (1 - distanceStopPct/100)
		} else {
			stop = vwmaStop[i] * (1 + distanceStopPct/100)
		}

		// 6. Determiner type signal
		typeSignal := "LONG"
		if direction == "BAISSIER" {
			typeSignal = "SHORT"
		}

		// 7. Determiner si valide et raison de rejet
		valide := true
		raison := "OK"

		if !gapValide {
			valide = false
			raison = fmt.Sprintf("Gap jamais validé (fenêtre=%d)", WINDOW_GAMMA_VALIDATE)
		} else if !volatiliteOK {
			valide = false
			raison = fmt.Sprintf("ATR%% %.2f%% < %.2f%%", atrPct, s.config.VolatiliteMin)
		} else if gapValideBougie > 0 {
			raison = fmt.Sprintf("OK (gamma validé +%d bougies)", gapValideBougie)
		}

		// Creer signal (valide ou non)
		signal := Signal{
			Index:           i,
			Timestamp:       klines[i].OpenTime,
			Type:            typeSignal,
			Prix:            prix,
			VWMARapide:      vwmaRapide[i],
			VWMALent:        vwmaLent[i],
			VWMAStop:        vwmaStop[i],
			Gap:             gapInitial,
			GapValide:       gapValide,
			GapValideBougie: gapValideBougie,
			ATR:             atr[i],
			ATRPct:          atrPct,
			VolatiliteOK:    volatiliteOK,
			Stop:            stop,
			DistanceStopPct: distanceStopPct,
			Valide:          valide,
			Raison:          raison,
		}

		signals = append(signals, signal)
	}

	return signals
}

// Afficher les signaux
func (s *VWMAStrategy) printSignals(signals []Signal) {
	fmt.Println("\n" + strings("=", 120))
	fmt.Println("TOUS LES CROISEMENTS VWMA DETECTES")
	fmt.Println(strings("=", 120))

	if len(signals) == 0 {
		fmt.Println("Aucun croisement detecte.")
		return
	}

	// Compter valides et rejetes
	nbValides := countValides(signals, true)
	nbRejetes := countValides(signals, false)

	// Format compact : une ligne par croisement
	fmt.Printf("%-4s | %-19s | %-5s | %-8s | %-8s | %-8s | %-6s | %-7s | %-7s | %-50s\n",
		"#", "Date/Heure (Open)", "Type", "Prix", "VWMA6", "Gap", "ATR%", "Gamma", "Status", "Raison")
	fmt.Println(strings("-", 140))

	for i, sig := range signals {
		status := "✓ OK"
		if !sig.Valide {
			status = "✗ REJET"
		}

		// Afficher quand le gap a été validé
		gammaInfo := "✗"
		if sig.GapValide {
			if sig.GapValideBougie == 0 {
				gammaInfo = "✓ imm"
			} else {
				gammaInfo = fmt.Sprintf("✓ +%d", sig.GapValideBougie)
			}
		}

		fmt.Printf("%-4d | %s | %-5s | %8.2f | %8.2f | %8.4f | %6.2f%% | %-7s | %-7s | %-50s\n",
			i+1,
			sig.Timestamp.Format("2006-01-02 15:04:05"),
			sig.Type,
			sig.Prix,
			sig.VWMARapide,
			sig.Gap,
			sig.ATRPct,
			gammaInfo,
			status,
			sig.Raison)
	}

	fmt.Println(strings("=", 120))
	fmt.Printf("Total croisements: %d | Valides: %d | Rejetes: %d\n",
		len(signals), nbValides, nbRejetes)
	fmt.Printf("Details valides: LONG=%d | SHORT=%d\n",
		countTypeValide(signals, "LONG", true),
		countTypeValide(signals, "SHORT", true))
	fmt.Println(strings("=", 120))
}

func countValides(signals []Signal, valide bool) int {
	count := 0
	for _, sig := range signals {
		if sig.Valide == valide {
			count++
		}
	}
	return count
}

func countTypeValide(signals []Signal, typeSignal string, valide bool) int {
	count := 0
	for _, sig := range signals {
		if sig.Type == typeSignal && sig.Valide == valide {
			count++
		}
	}
	return count
}

func countType(signals []Signal, typeSignal string) int {
	count := 0
	for _, sig := range signals {
		if sig.Type == typeSignal {
			count++
		}
	}
	return count
}

// Analyser les positions relatives
func (s *VWMAStrategy) analyzePositions(klines []gateio.Kline, vwmaRapide, vwmaLent []float64) {
	fmt.Println("\n" + strings("=", 80))
	fmt.Println("ANALYSE POSITIONS (10 dernieres bougies)")
	fmt.Println(strings("=", 80))

	startIdx := len(klines) - 10
	if startIdx < 0 {
		startIdx = 0
	}

	fmt.Printf("%-8s | %-7s | %-7s | %-10s | %-6s\n", "Heure", "VWMA6", "VWMA20", "Position", "Gap")
	fmt.Println(strings("-", 80))

	for i := startIdx; i < len(klines); i++ {
		position := indicators.PositionRelative(vwmaRapide[i], vwmaLent[i])
		gap := indicators.CalculerEcart(vwmaRapide[i], vwmaLent[i])

		timestamp := klines[i].OpenTime
		fmt.Printf("%s | %7.2f | %7.2f | %-10s | %6.4f\n",
			timestamp.Format("15:04:05"),
			vwmaRapide[i],
			vwmaLent[i],
			position, gap)
	}

	fmt.Println(strings("=", 80))
}

// Analyser les tendances maintenues
func (s *VWMAStrategy) analyzeTrends(klines []gateio.Kline, vwmaRapide, vwmaLent []float64) {
	fmt.Println("\n" + strings("=", 80))
	fmt.Println("ANALYSE TENDANCES")
	fmt.Println(strings("=", 80))

	periodes := []int{3, 5, 10}

	idx := len(klines) - 1
	timestamp := klines[idx].OpenTime

	fmt.Printf("Position actuelle (%s): VWMA6=%.2f | VWMA20=%.2f\n",
		timestamp.Format("15:04:05"),
		vwmaRapide[idx],
		vwmaLent[idx])
	fmt.Println(strings("-", 80))

	for _, p := range periodes {
		hausse := indicators.PositionMaintenue(vwmaRapide, vwmaLent, idx, p, "AU-DESSUS")
		baisse := indicators.PositionMaintenue(vwmaRapide, vwmaLent, idx, p, "EN-DESSOUS")

		var tendance string
		if hausse {
			tendance = "HAUSSIERE"
		} else if baisse {
			tendance = "BAISSIERE"
		} else {
			tendance = "INSTABLE"
		}

		fmt.Printf("Sur %2d bougies: %s\n", p, tendance)
	}

	fmt.Println(strings("=", 80))
}

// Afficher les statistiques
func (s *VWMAStrategy) printStats(signals []Signal) {
	fmt.Println("\n" + strings("=", 80))
	fmt.Println("STATISTIQUES (Signaux valides uniquement)")
	fmt.Println(strings("=", 80))

	// Filtrer uniquement les signaux valides
	var validSignals []Signal
	for _, sig := range signals {
		if sig.Valide {
			validSignals = append(validSignals, sig)
		}
	}

	if len(validSignals) == 0 {
		fmt.Println("Aucun signal valide pour statistiques.")
		fmt.Println(strings("=", 80))
		return
	}

	var nbLong, nbShort int
	var sumATRPct, sumGap, sumDistanceStop float64

	for _, sig := range validSignals {
		if sig.Type == "LONG" {
			nbLong++
		} else {
			nbShort++
		}
		sumATRPct += sig.ATRPct
		sumGap += sig.Gap
		sumDistanceStop += sig.DistanceStopPct
	}

	total := len(validSignals)

	fmt.Printf("Total: %d | LONG: %d (%.1f%%) | SHORT: %d (%.1f%%)\n",
		total,
		nbLong, float64(nbLong)/float64(total)*100,
		nbShort, float64(nbShort)/float64(total)*100)
	fmt.Printf("Moyennes: ATR%%=%.2f%% | Gap=%.4f | Stop=%.2f%%\n",
		sumATRPct/float64(total),
		sumGap/float64(total),
		sumDistanceStop/float64(total))

	fmt.Println(strings("=", 80))
}

func strings(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

func main() {
	// Configuration depuis les constantes globales
	config := Config{
		// Symbole et donnees
		Symbol:    SYMBOL,
		Timeframe: TIMEFRAME,
		NbCandles: NB_CANDLES,

		// Periodes VWMA
		VWMARapide: VWMA_RAPIDE,
		VWMALent:   VWMA_LENT,
		VWMAStop:   VWMA_STOP,

		// Calibrage
		GammaGap:      GAMMA_GAP,
		ATRPeriode:    ATR_PERIODE,
		TauSlope:      TAU_SLOPE,
		VolatiliteMin: VOLATILITE_MIN,

		// Stop management
		KStop: K_STOP,
		PMin:  P_MIN,
		PMax:  P_MAX,

		// Fenetre
		WindowW: WINDOW_W,
	}

	fmt.Println("=== DEMO STRATEGIE VWMA ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange        : Gate.io\n")
	fmt.Printf("Symbole         : %s\n", config.Symbol)
	fmt.Printf("Timeframe       : %s\n", config.Timeframe)
	fmt.Printf("VWMA rapide     : %d\n", config.VWMARapide)
	fmt.Printf("VWMA lent       : %d\n", config.VWMALent)
	fmt.Printf("VWMA stop       : %d\n", config.VWMAStop)
	fmt.Printf("Gamma gap       : %.2f x ATR\n", config.GammaGap)
	fmt.Printf("Fenetre gamma   : %d bougies (validation différée)\n", WINDOW_GAMMA_VALIDATE)
	fmt.Printf("ATR periode     : %d\n", config.ATRPeriode)
	fmt.Printf("Volatilite min  : %.2f%%\n", config.VolatiliteMin)
	fmt.Printf("K stop          : %.2f\n", config.KStop)
	fmt.Printf("Stop min/max    : %.2f%% / %.2f%%\n", config.PMin, config.PMax)

	// Creer la strategie
	strategy := NewVWMAStrategy(config)

	// Recuperer les donnees
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

	// Calculer les indicateurs
	fmt.Println("\nCalcul des indicateurs...")
	vwmaRapide, vwmaLent, vwmaStop, atr := strategy.calculateIndicators(klines)

	// Detecter les signaux
	fmt.Println("Detection des signaux...")
	signals := strategy.detectSignals(klines, vwmaRapide, vwmaLent, vwmaStop, atr)

	// Afficher les resultats
	strategy.printSignals(signals)
	strategy.analyzePositions(klines, vwmaRapide, vwmaLent)
	strategy.analyzeTrends(klines, vwmaRapide, vwmaLent)
	strategy.printStats(signals)

	fmt.Println("\n=== FIN DEMO ===")
}
