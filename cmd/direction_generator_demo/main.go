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
	"agent-economique/internal/signals/direction"
)

// Configuration
const (
	SYMBOL     = "SOL_USDT"
	TIMEFRAME  = "1m"
	NB_CANDLES = 500

	VWMA_RAPIDE           = 3 //3
	PERIODE_PENTE         = 2 //5
	SEUIL_PENTE_VWMA      = 0.1
	K_CONFIRMATION        = 1 // 2
	USE_DYNAMIC_THRESHOLD = true
	ATR_PERIODE           = 8   //4
	ATR_COEFFICIENT       = 0.1 //0.8
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
	fmt.Println("=== DEMO GENERATEUR DIRECTION (Production) ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange           : Gate.io\n")
	fmt.Printf("Symbole            : %s\n", SYMBOL)
	fmt.Printf("Timeframe          : %s\n", TIMEFRAME)
	fmt.Printf("VWMA rapide        : %d\n", VWMA_RAPIDE)
	fmt.Printf("Periode pente      : %d bougies\n", PERIODE_PENTE)
	if USE_DYNAMIC_THRESHOLD {
		fmt.Printf("Seuil pente        : DYNAMIQUE (ATR × %.2f)\n", ATR_COEFFICIENT)
		fmt.Printf("ATR periode        : %d\n", ATR_PERIODE)
	} else {
		fmt.Printf("Seuil pente        : %.2f%% (fixe)\n", SEUIL_PENTE_VWMA)
	}
	fmt.Printf("K confirmation     : %d bougies\n", K_CONFIRMATION)

	// Récupération données
	fmt.Printf("\nRecuperation des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}

	// Tri chronologique
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime.Before(klines[j].OpenTime)
	})

	fmt.Printf("Klines recuperees: %d\n", len(klines))

	// Convertir en format signals.Kline
	signalKlines := make([]signals.Kline, len(klines))
	for i, k := range klines {
		signalKlines[i] = signals.Kline{
			OpenTime: k.OpenTime,
			Open:     k.Open,
			High:     k.High,
			Low:      k.Low,
			Close:    k.Close,
			Volume:   k.Volume,
		}
	}

	// Créer le générateur
	config := direction.Config{
		VWMAPeriod:          VWMA_RAPIDE,
		SlopePeriod:         PERIODE_PENTE,
		KConfirmation:       K_CONFIRMATION,
		UseDynamicThreshold: USE_DYNAMIC_THRESHOLD,
		ATRPeriod:           ATR_PERIODE,
		ATRCoefficient:      ATR_COEFFICIENT,
		FixedThreshold:      SEUIL_PENTE_VWMA,
	}

	generator := direction.NewDirectionGenerator(config)

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

	// Détecter signaux
	fmt.Println("Detection des signaux...")
	allSignals, err := generator.DetectSignals(signalKlines)
	if err != nil {
		log.Fatalf("Erreur detection signaux: %v", err)
	}

	fmt.Printf("Signaux detectes: %d\n", len(allSignals))

	// Regrouper en intervalles
	intervalles := buildIntervalles(allSignals, signalKlines)

	// Affichage
	displayIntervalles(intervalles)
	displayStatistics(intervalles)

	// Métriques générateur
	metrics := generator.GetMetrics()
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("METRIQUES GENERATEUR")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Total signaux      : %d\n", metrics.TotalSignals)
	fmt.Printf("  - Entry          : %d\n", metrics.EntrySignals)
	fmt.Printf("  - Exit           : %d\n", metrics.ExitSignals)
	fmt.Printf("  - Long           : %d\n", metrics.LongSignals)
	fmt.Printf("  - Short          : %d\n", metrics.ShortSignals)
	fmt.Printf("Confiance moyenne  : %.2f\n", metrics.AvgConfidence)
	fmt.Println(strings.Repeat("=", 100))

	// Exporter résultats
	if err := exportResults(signalKlines, intervalles); err != nil {
		log.Printf("⚠️  Erreur export: %v", err)
	}

	fmt.Println("\n=== FIN DEMO ===")
}

func buildIntervalles(sigs []signals.Signal, klines []signals.Kline) []Intervalle {
	var intervalles []Intervalle

	var currentEntry *signals.Signal
	intervalNum := 0

	for i := 0; i < len(sigs); i++ {
		sig := sigs[i]

		if sig.Action == signals.SignalActionEntry {
			currentEntry = &sig
		} else if sig.Action == signals.SignalActionExit && currentEntry != nil {
			intervalNum++

			// Calculer nombre de bougies
			nbBougies := 0
			for j := 0; j < len(klines); j++ {
				if klines[j].OpenTime.Equal(currentEntry.Timestamp) || klines[j].OpenTime.After(currentEntry.Timestamp) {
					if klines[j].OpenTime.Before(sig.Timestamp) || klines[j].OpenTime.Equal(sig.Timestamp) {
						nbBougies++
					}
				}
			}

			variation := 0.0
			if currentEntry.Price != 0 {
				variation = (sig.Price - currentEntry.Price) / currentEntry.Price * 100
			}

			intervalle := Intervalle{
				Numero:          intervalNum,
				Type:            currentEntry.Type,
				DateDebut:       currentEntry.Timestamp,
				DateFin:         sig.Timestamp,
				PrixDebut:       currentEntry.Price,
				PrixFin:         sig.Price,
				NbBougies:       nbBougies,
				VariationCaptee: variation,
			}

			intervalles = append(intervalles, intervalle)
			currentEntry = nil
		}
	}

	return intervalles
}

func displayIntervalles(intervalles []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("INTERVALLES DIRECTION (via Generateur)")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-9s | %-14s | %-19s | %-19s | %-8s | %-12s | %-12s\n",
		"#", "Var %", "Type", "Date Début", "Date Fin", "Bougies", "Prix Début", "Prix Fin")
	fmt.Println(strings.Repeat("-", 140))

	for _, inter := range intervalles {
		typeDisplay := "↗ LONG"
		if inter.Type == signals.SignalTypeShort {
			typeDisplay = "↘ SHORT"
		}

		fmt.Printf("%-4d | %+8.2f%% | %-14s | %-19s | %-19s | %8d | %12.2f | %12.2f\n",
			inter.Numero,
			inter.VariationCaptee,
			typeDisplay,
			inter.DateDebut.Format("2006-01-02 15:04:05"),
			inter.DateFin.Format("2006-01-02 15:04:05"),
			inter.NbBougies,
			inter.PrixDebut,
			inter.PrixFin)
	}

	fmt.Println(strings.Repeat("=", 140))
}

func displayStatistics(intervalles []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES INTERVALLES")
	fmt.Println(strings.Repeat("=", 100))

	totalLong := 0
	totalShort := 0
	bougiesLong := 0
	bougiesShort := 0
	variationLong := 0.0
	variationShort := 0.0

	for _, inter := range intervalles {
		if inter.Type == signals.SignalTypeLong {
			totalLong++
			bougiesLong += inter.NbBougies
			variationLong += inter.VariationCaptee
		} else {
			totalShort++
			bougiesShort += inter.NbBougies
			variationShort += inter.VariationCaptee
		}
	}

	fmt.Println("\nINTERVALLES:")
	fmt.Printf("  Total intervalles    : %d\n", len(intervalles))
	fmt.Printf("  - Long (↗)           : %d intervalles (%d bougies)\n", totalLong, bougiesLong)
	fmt.Printf("  - Short (↘)          : %d intervalles (%d bougies)\n", totalShort, bougiesShort)

	fmt.Println("\nVARIATIONS CAPTÉES:")
	if totalLong > 0 {
		fmt.Printf("  - Long (↗)           : %+.2f%% total, %+.2f%% moyen par intervalle\n",
			variationLong, variationLong/float64(totalLong))
	}
	if totalShort > 0 {
		fmt.Printf("  - Short (↘)          : %+.2f%% total, %+.2f%% moyen par intervalle\n",
			variationShort, variationShort/float64(totalShort))
	}

	// Total capté = LONG + (SHORT × -1)
	// Les variations SHORT profitables sont négatives, donc on inverse
	totalCapte := variationLong - variationShort

	fmt.Printf("  - TOTAL CAPTÉ        : %.2f%% (bidirectionnel LONG+SHORT)\n",
		totalCapte)

	fmt.Println(strings.Repeat("=", 100))
}

// exportResults exporte les klines et intervalles en JSON
func exportResults(klines []signals.Kline, intervalles []Intervalle) error {
	// Créer nom dossier avec paramètres
	config := fmt.Sprintf("vwma%d_slope%d_k%d_atr%d_coef%.2f",
		VWMA_RAPIDE, PERIODE_PENTE, K_CONFIRMATION, ATR_PERIODE, ATR_COEFFICIENT)
	outDir := filepath.Join("out", fmt.Sprintf("direction_demo_%s_%s", TIMEFRAME, config))

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("erreur création dossier: %w", err)
	}

	// Export klines
	klinesFile := filepath.Join(outDir, "klines.json")
	kf, err := os.Create(klinesFile)
	if err != nil {
		return err
	}
	defer kf.Close()

	encoder := json.NewEncoder(kf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(klines); err != nil {
		return err
	}

	// Export intervalles
	intervallesFile := filepath.Join(outDir, "intervalles.json")
	intf, err := os.Create(intervallesFile)
	if err != nil {
		return err
	}
	defer intf.Close()

	encoder = json.NewEncoder(intf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(intervalles); err != nil {
		return err
	}

	fmt.Printf("\n💾 Données exportées dans: %s/\n", outDir)
	return nil
}
