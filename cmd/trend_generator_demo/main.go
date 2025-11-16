package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/trend"
)

// Configuration (paramètres du trend_demo original)
const (
	SYMBOL     = "SOL_USDT"
	TIMEFRAME  = "5m"
	NB_CANDLES = 1000

	// Paramètres VWMA (du trend_demo)
	VWMA_RAPIDE = 5
	VWMA_LENT   = 15
	VWMA_STOP   = 72

	// Paramètres DMI
	DMI_PERIODE = 5
	DMI_SMOOTH  = 3

	// Calibrage ATR et validation (RÉDUIT pour plus de signaux)
	ATR_PERIODE           = 3
	GAMMA_GAP_VWMA        = 0.5 // 50% de ATR pour gap VWMA
	GAMMA_GAP_DI          = 2.0 // Gap minimal DI+ vs DI- (réduit de 5.0 à 2.0)
	GAMMA_GAP_DX          = 2.0 // Gap minimal DX vs ADX (réduit de 5.0 à 2.0)
	VOLATILITE_MIN        = 0.3 // 0.3% ATR% minimal
	WINDOW_GAMMA_VALIDATE = 5   // Fenetre validation gamma différée
	WINDOW_W              = 10  // Fenetre matching VWMA + DMI

	// Filtres de bougie initiale (setup de base)
	BODY_PCT_MIN             = 0.60
	BODY_ATR_MIN             = 0.60
	ENFORCE_CANDLE_DIRECTION = true

	// Sorties indépendantes
	EXIT_VWMA          = true
	EXIT_TRAILING      = false
	TRAILING_ATR_COEFF = 1.0
	TRAILING_CAP_PCT   = 0.003
)

// Intervalle pour statistiques
type Intervalle struct {
	Numero          int
	Type            signals.SignalType
	DateDebut       time.Time
	DateFin         time.Time
	PrixDebut       float64
	PrixFin         float64
	NbBougies       int
	VariationCaptee float64
}

func main() {
	fmt.Println("=== DEMO GENERATEUR TREND (VWMA + DMI) ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange           : Gate.io\n")
	fmt.Printf("Symbole            : %s\n", SYMBOL)
	fmt.Printf("Timeframe          : %s\n", TIMEFRAME)
	fmt.Printf("Nb klines          : %d\n", NB_CANDLES)
	fmt.Printf("VWMA rapide/lent   : %d / %d\n", VWMA_RAPIDE, VWMA_LENT)
	fmt.Printf("DMI periode/smooth : %d / %d\n", DMI_PERIODE, DMI_SMOOTH)
	fmt.Printf("Gamma gap VWMA     : %.2f x ATR\n", GAMMA_GAP_VWMA)
	fmt.Printf("Gamma gap DI       : %.1f\n", GAMMA_GAP_DI)
	fmt.Printf("Gamma gap DX       : %.1f\n", GAMMA_GAP_DX)
	fmt.Printf("Fenetre gamma      : %d bougies\n", WINDOW_GAMMA_VALIDATE)
	fmt.Printf("Fenetre W (match)  : %d bougies\n", WINDOW_W)
	fmt.Printf("Volatilite min     : %.2f%%\n", VOLATILITE_MIN)
	fmt.Printf("Body%% min           : %.2f\n", BODY_PCT_MIN)
	fmt.Printf("Body/ATR min        : %.2f\n", BODY_ATR_MIN)
	fmt.Printf("Bougie cohérente    : %v\n", ENFORCE_CANDLE_DIRECTION)
	fmt.Printf("Exit VWMA           : %v\n", EXIT_VWMA)
	fmt.Printf("Exit Trailing       : %v (ATR coeff=%.2f, cap=%.3f)\n", EXIT_TRAILING, TRAILING_ATR_COEFF, TRAILING_CAP_PCT)

	// Récupération des données
	fmt.Printf("\nRecuperation des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	gateioKlines, err := client.GetKlines(ctx, SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}

	// Tri chronologique
	sort.Slice(gateioKlines, func(i, j int) bool {
		return gateioKlines[i].OpenTime.Before(gateioKlines[j].OpenTime)
	})

	fmt.Printf("Klines recuperees: %d\n", len(gateioKlines))

	// LOG: Période des klines après tri
	if len(gateioKlines) > 0 {
		premier := gateioKlines[0].OpenTime
		dernier := gateioKlines[len(gateioKlines)-1].OpenTime
		fmt.Printf("📅 Période des klines:\n")
		fmt.Printf("  - Premier : %s\n", premier.Format("2006-01-02 15:04:05"))
		fmt.Printf("  - Dernier : %s\n", dernier.Format("2006-01-02 15:04:05"))
		fmt.Printf("  - Durée   : %.1f heures\n", dernier.Sub(premier).Hours())
	}

	// Convertir en format signals.Kline
	signalKlines := make([]signals.Kline, len(gateioKlines))
	for i, k := range gateioKlines {
		signalKlines[i] = signals.Kline{
			OpenTime: k.OpenTime,
			Open:     k.Open,
			High:     k.High,
			Low:      k.Low,
			Close:    k.Close,
			Volume:   k.Volume,
		}
	}

	// Créer le générateur avec configuration du trend_demo
	config := trend.Config{
		VwmaRapide:          VWMA_RAPIDE,
		VwmaLent:            VWMA_LENT,
		DmiPeriode:          DMI_PERIODE,
		DmiSmooth:           DMI_SMOOTH,
		AtrPeriode:          ATR_PERIODE,
		GammaGapVWMA:        GAMMA_GAP_VWMA,
		GammaGapDI:          GAMMA_GAP_DI,
		GammaGapDX:          GAMMA_GAP_DX,
		VolatiliteMin:       VOLATILITE_MIN,
		WindowGammaValidate: WINDOW_GAMMA_VALIDATE,
		WindowW:             WINDOW_W,
		// Filtres de bougie initiale
		BodyPctMin:             BODY_PCT_MIN,
		BodyATRMin:             BODY_ATR_MIN,
		EnforceCandleDirection: ENFORCE_CANDLE_DIRECTION,
		// New exit controls
		EnableExitVWMA:     EXIT_VWMA,
		EnableExitTrailing: EXIT_TRAILING,
		TrailingATRCoeff:   TRAILING_ATR_COEFF,
		TrailingCapPct:     TRAILING_CAP_PCT,
	}

	generator := trend.NewTrendGenerator(config)

	// Initialiser
	genConfig := signals.GeneratorConfig{
		Symbol:    SYMBOL,
		Timeframe: TIMEFRAME,
	}
	if err := generator.Initialize(genConfig); err != nil {
		log.Fatalf("Erreur initialisation: %v", err)
	}

	// Calculer indicateurs
	fmt.Println("\nCalcul des indicateurs...")
	if err := generator.CalculateIndicators(signalKlines); err != nil {
		log.Fatalf("Erreur calcul indicateurs: %v", err)
	}

	// Détecter signaux (FORCER traitement complet en mode demo)
	fmt.Println("Detection des signaux...")
	// Note: En mode demo, on veut traiter TOUTES les klines, pas seulement les nouvelles
	allSignals, err := generator.DetectSignals(signalKlines)
	if err != nil {
		log.Fatalf("Erreur detection signaux: %v", err)
	}

	fmt.Printf("Signaux detectes: %d\n", len(allSignals))

	// Regrouper en intervalles
	intervalles := buildIntervalles(allSignals, signalKlines)

	// Affichage
	displaySignals(allSignals)
	displayIntervalles(intervalles)
	displayStatistics(intervalles)

	// Métriques générateur
	metrics := generator.GetMetrics()
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("METRIQUES GENERATEUR TREND")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Total signaux      : %d\n", metrics.TotalSignals)
	fmt.Printf("  - Entry          : %d\n", metrics.EntrySignals)
	fmt.Printf("  - Exit           : %d\n", metrics.ExitSignals)
	fmt.Printf("  - Long           : %d\n", metrics.LongSignals)
	fmt.Printf("  - Short          : %d\n", metrics.ShortSignals)
	fmt.Printf("Confiance moyenne  : %.2f\n", metrics.AvgConfidence)
	fmt.Println(strings.Repeat("=", 100))

	// Exporter résultats
	if err := exportResults(signalKlines, intervalles, allSignals); err != nil {
		log.Printf("⚠️  Erreur export: %v", err)
	}

	fmt.Println("\n=== FIN DEMO TREND GENERATOR ===")
}

func buildIntervalles(sigs []signals.Signal, klines []signals.Kline) []Intervalle {
	var intervalles []Intervalle

	// IMPORTANT: Trier les signaux chronologiquement AVANT de les traiter
	sort.Slice(sigs, func(i, j int) bool {
		return sigs[i].Timestamp.Before(sigs[j].Timestamp)
	})

	var currentEntry *signals.Signal
	intervalNum := 0

	for i := 0; i < len(sigs); i++ {
		sig := sigs[i]

		if sig.Action == "ENTRY" {
			currentEntry = &sig
		} else if sig.Action == "EXIT" && currentEntry != nil {
			intervalNum++

			intervalle := Intervalle{
				Numero:    intervalNum,
				Type:      currentEntry.Type,
				DateDebut: currentEntry.Timestamp,
				DateFin:   sig.Timestamp,
				PrixDebut: currentEntry.Price,
				PrixFin:   sig.Price,
			}

			// Calculer variation selon type
			if currentEntry.Type == "LONG" {
				intervalle.VariationCaptee = ((sig.Price - currentEntry.Price) / currentEntry.Price) * 100
			} else {
				intervalle.VariationCaptee = ((currentEntry.Price - sig.Price) / currentEntry.Price) * 100
			}

			// Calculer durée en bougies
			for j, kline := range klines {
				if kline.OpenTime.Equal(currentEntry.Timestamp) {
					for k, kline2 := range klines[j:] {
						if kline2.OpenTime.Equal(sig.Timestamp) {
							intervalle.NbBougies = k + 1
							break
						}
					}
					break
				}
			}

			intervalles = append(intervalles, intervalle)
			currentEntry = nil
		}
	}

	return intervalles
}

func displaySignals(signals []signals.Signal) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("SIGNAUX TREND DETECTES")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s | %-19s | %-6s | %-8s | %-8s | %-10s | %-50s\n",
		"#", "Date/Heure", "Action", "Type", "Prix", "Confiance", "Raison")
	fmt.Println(strings.Repeat("-", 120))

	// Trier les signaux par timestamp chronologique
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Timestamp.Before(signals[j].Timestamp)
	})

	for i, sig := range signals {
		actionStr := "ENTRY"
		if sig.Action == "EXIT" {
			actionStr = "EXIT"
		}

		typeStr := "LONG"
		if sig.Type == "SHORT" {
			typeStr = "SHORT"
		}

		reason := ""
		if reasonVal, ok := sig.Metadata["reason"]; ok {
			reason = reasonVal.(string)
		}

		fmt.Printf("%-4d | %s | %-6s | %-8s | %8.2f | %10.2f | %-50s\n",
			i+1, sig.Timestamp.Format("2006-01-02 15:04:05"), actionStr, typeStr,
			sig.Price, sig.Confidence, reason)
	}

	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("Total signaux: %d\n", len(signals))
}

func displayIntervalles(intervalles []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("INTERVALLES TREND")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s | %-8s | %-19s | %-19s | %-8s | %-8s | %-7s | %-10s\n",
		"#", "Type", "Debut", "Fin", "P.Debut", "P.Fin", "Bougies", "Variation")
	fmt.Println(strings.Repeat("-", 120))

	// Trier par date
	sort.Slice(intervalles, func(i, j int) bool {
		return intervalles[i].DateDebut.Before(intervalles[j].DateDebut)
	})

	for _, inter := range intervalles {
		typeStr := "LONG"
		if inter.Type == "SHORT" {
			typeStr = "SHORT"
		}

		variationStr := fmt.Sprintf("%.2f%%", inter.VariationCaptee)
		if inter.VariationCaptee > 0 {
			variationStr = fmt.Sprintf("+%.2f%%", inter.VariationCaptee)
		}

		fmt.Printf("%-4d | %-8s | %s | %s | %8.2f | %8.2f | %-7d | %-10s\n",
			inter.Numero, typeStr,
			inter.DateDebut.Format("2006-01-02 15:04:05"),
			inter.DateFin.Format("2006-01-02 15:04:05"),
			inter.PrixDebut, inter.PrixFin, inter.NbBougies, variationStr)
	}
	fmt.Println(strings.Repeat("=", 120))
}

func displayStatistics(intervalles []Intervalle) {
	if len(intervalles) == 0 {
		fmt.Println("\n⚠️  Aucun intervalle complet pour calculer les statistiques")
		return
	}

	// Statistiques globales
	var totalVariation, totalPositive, totalNegative float64
	var nbPositives, nbNegatives int
	var nbLong, nbShort int

	for _, inter := range intervalles {
		totalVariation += inter.VariationCaptee

		if inter.VariationCaptee > 0 {
			totalPositive += inter.VariationCaptee
			nbPositives++
		} else {
			totalNegative += inter.VariationCaptee
			nbNegatives++
		}

		if inter.Type == "LONG" {
			nbLong++
		} else {
			nbShort++
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("STATISTIQUES TREND")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Nombre intervalles     : %d\n", len(intervalles))
	fmt.Printf("  - Long               : %d (%.1f%%)\n", nbLong, float64(nbLong)/float64(len(intervalles))*100)
	fmt.Printf("  - Short              : %d (%.1f%%)\n", nbShort, float64(nbShort)/float64(len(intervalles))*100)
	fmt.Printf("\nVariations:\n")
	fmt.Printf("  - Variation totale   : %.2f%%\n", totalVariation)
	fmt.Printf("  - Variation moyenne  : %.2f%%\n", totalVariation/float64(len(intervalles)))
	fmt.Printf("  - Trades positifs    : %d (%.1f%%)\n", nbPositives, float64(nbPositives)/float64(len(intervalles))*100)
	fmt.Printf("  - Trades négatifs    : %d (%.1f%%)\n", nbNegatives, float64(nbNegatives)/float64(len(intervalles))*100)

	if nbPositives > 0 {
		fmt.Printf("  - Gain moyen positif : %.2f%%\n", totalPositive/float64(nbPositives))
	}
	if nbNegatives > 0 {
		fmt.Printf("  - Perte moyenne      : %.2f%%\n", totalNegative/float64(nbNegatives))
	}

	winRate := float64(nbPositives) / float64(len(intervalles)) * 100
	fmt.Printf("\nWin Rate               : %.1f%%\n", winRate)
	fmt.Println(strings.Repeat("=", 80))
}

func exportResults(klines []signals.Kline, intervalles []Intervalle, signals []signals.Signal) error {
	// Créer dossier output
	outputDir := "out/trend_generator_demo_" + time.Now().Format("20060102_150405")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("création dossier: %w", err)
	}

	// Export klines
	klinesFile := filepath.Join(outputDir, "klines.json")
	if err := saveJSON(klinesFile, klines); err != nil {
		return fmt.Errorf("export klines: %w", err)
	}

	// Export intervalles
	intervallesFile := filepath.Join(outputDir, "intervalles.json")
	if err := saveJSON(intervallesFile, intervalles); err != nil {
		return fmt.Errorf("export intervalles: %w", err)
	}

	// Export signaux
	signalsFile := filepath.Join(outputDir, "signals.json")
	if err := saveJSON(signalsFile, signals); err != nil {
		return fmt.Errorf("export signals: %w", err)
	}

	fmt.Printf("\n✅ Resultats exportes dans: %s\n", outputDir)
	return nil
}

func saveJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
