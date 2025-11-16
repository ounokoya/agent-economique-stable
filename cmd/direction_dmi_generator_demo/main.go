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
	"agent-economique/internal/signals/direction_dmi"
)

// Configuration
// Config par défaut = Paramètres optimaux Direction (VWMA=20, Slope=6, ATR=8, Coef=0.25)
// + Paramètres DMI standards
const (
	SYMBOL     = "SOL_USDT"
	TIMEFRAME  = "5m"
	NB_CANDLES = 500

	// Paramètres VWMA (optimaux identifiés)
	VWMA_PERIOD           = 20   // Optimal filtrage bruit
	SLOPE_PERIOD          = 6    // Pente stable
	K_CONFIRMATION        = 2    // Standard
	USE_DYNAMIC_THRESHOLD = true
	ATR_PERIOD            = 8    // Adapté moyen terme
	ATR_COEFFICIENT       = 0.25 // Sensibilité optimale
	FIXED_THRESHOLD       = 0.1  // Ignoré si dynamic

	// Paramètres DMI
	DMI_PERIOD             = 14  // Standard
	DMI_SMOOTH             = 14  // Standard
	GAMMA_GAP_DI           = 2.0 // % minimum gap DI+ vs DI-
	GAMMA_GAP_DX           = 2.0 // % minimum gap DX vs ADX
	WINDOW_GAMMA_VALIDATE  = 5   // Fenêtre validation gap
	WINDOW_MATCHING        = 5   // Fenêtre matching 3 conditions

	// Flags d'activation (mode AGRESSIF pour démo)
	ENABLE_ENTRY_TREND         = true  // Entrées tendance
	ENABLE_ENTRY_COUNTER_TREND = true  // Entrées contre-tendance
	ENABLE_EXIT_TREND          = true  // Sorties tendance
	ENABLE_EXIT_COUNTER_TREND  = true  // Sorties contre-tendance
)

// Intervalle pour statistiques
type Intervalle struct {
	Numero          int
	Type            signals.SignalType
	Mode            string // "TREND" ou "COUNTER_TREND"
	DateDebut       time.Time
	DateFin         time.Time
	PrixDebut       float64
	PrixFin         float64
	NbBougies       int
	VariationCaptee float64
}

// ExportData pour sauvegarde JSON
type ExportData struct {
	Config     direction_dmi.Config `json:"config"`
	Klines     []signals.Kline      `json:"klines"`
	Intervalles []Intervalle        `json:"intervalles"`
	Stats      StatsSummary         `json:"stats"`
}

// StatsSummary résumé des statistiques
type StatsSummary struct {
	TotalIntervalles    int     `json:"total_intervalles"`
	LongTendance        int     `json:"long_tendance"`
	LongContreTendance  int     `json:"long_contre_tendance"`
	ShortTendance       int     `json:"short_tendance"`
	ShortContreTendance int     `json:"short_contre_tendance"`
	VariationLongTrend  float64 `json:"variation_long_trend"`
	VariationLongCT     float64 `json:"variation_long_ct"`
	VariationShortTrend float64 `json:"variation_short_trend"`
	VariationShortCT    float64 `json:"variation_short_ct"`
	TotalCapte          float64 `json:"total_capte"`
}

func main() {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  DIRECTION+DMI GENERATOR DEMO")
	fmt.Println("═══════════════════════════════════════════════════")

	// Configuration
	config := direction_dmi.Config{
		// Paramètres VWMA
		VWMAPeriod:          VWMA_PERIOD,
		SlopePeriod:         SLOPE_PERIOD,
		KConfirmation:       K_CONFIRMATION,
		UseDynamicThreshold: USE_DYNAMIC_THRESHOLD,
		ATRPeriod:           ATR_PERIOD,
		ATRCoefficient:      ATR_COEFFICIENT,
		FixedThreshold:      FIXED_THRESHOLD,

		// Paramètres DMI
		DMIPeriod:           DMI_PERIOD,
		DMISmooth:           DMI_SMOOTH,
		GammaGapDI:          GAMMA_GAP_DI,
		GammaGapDX:          GAMMA_GAP_DX,
		WindowGammaValidate: WINDOW_GAMMA_VALIDATE,
		WindowMatching:      WINDOW_MATCHING,

		// Flags
		EnableEntryTrend:        ENABLE_ENTRY_TREND,
		EnableEntryCounterTrend: ENABLE_ENTRY_COUNTER_TREND,
		EnableExitTrend:         ENABLE_EXIT_TREND,
		EnableExitCounterTrend:  ENABLE_EXIT_COUNTER_TREND,
	}

	fmt.Println("\n📊 Configuration:")
	fmt.Printf("Symbole            : %s\n", SYMBOL)
	fmt.Printf("Timeframe          : %s\n", TIMEFRAME)
	fmt.Printf("Nb bougies         : %d\n", NB_CANDLES)
	fmt.Println("\n🎯 Paramètres VWMA (Direction):")
	fmt.Printf("VWMA période       : %d\n", config.VWMAPeriod)
	fmt.Printf("Slope période      : %d\n", config.SlopePeriod)
	fmt.Printf("K-confirmation     : %d\n", config.KConfirmation)
	fmt.Printf("Seuil dynamique    : %t\n", config.UseDynamicThreshold)
	fmt.Printf("ATR période        : %d\n", config.ATRPeriod)
	fmt.Printf("ATR coefficient    : %.2f\n", config.ATRCoefficient)
	fmt.Println("\n📡 Paramètres DMI:")
	fmt.Printf("DMI période        : %d\n", config.DMIPeriod)
	fmt.Printf("DMI smooth         : %d\n", config.DMISmooth)
	fmt.Printf("Gamma gap DI       : %.1f%%\n", config.GammaGapDI)
	fmt.Printf("Gamma gap DX       : %.1f%%\n", config.GammaGapDX)
	fmt.Printf("Window validation  : %d\n", config.WindowGammaValidate)
	fmt.Printf("Window matching    : %d\n", config.WindowMatching)
	fmt.Println("\n🎛️ Flags d'activation:")
	fmt.Printf("Entry Tendance     : %t\n", config.EnableEntryTrend)
	fmt.Printf("Entry Contre-T     : %t\n", config.EnableEntryCounterTrend)
	fmt.Printf("Exit Tendance      : %t\n", config.EnableExitTrend)
	fmt.Printf("Exit Contre-T      : %t\n", config.EnableExitCounterTrend)

	// Récupérer les données
	fmt.Println("\n📂 Récupération des klines depuis Gate.io...")
	gateioClient := gateio.NewClient()
	gateioKlines, err := gateioClient.GetKlines(context.Background(), SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("❌ Erreur récupération klines: %v", err)
	}
	
	// Tri chronologique (comme Direction Demo et Validateur DMI)
	sort.Slice(gateioKlines, func(i, j int) bool {
		return gateioKlines[i].OpenTime.Before(gateioKlines[j].OpenTime)
	})
	
	fmt.Printf("✅ %d klines récupérées et triées chronologiquement\n", len(gateioKlines))

	// Convertir en signals.Kline
	klines := make([]signals.Kline, len(gateioKlines))
	for i, gk := range gateioKlines {
		klines[i] = signals.Kline{
			OpenTime: gk.OpenTime,
			Open:     gk.Open,
			High:     gk.High,
			Low:      gk.Low,
			Close:    gk.Close,
			Volume:   gk.Volume,
		}
	}

	// Créer le générateur
	fmt.Println("\n⚙️  Initialisation générateur Direction+DMI...")
	genConfig := signals.GeneratorConfig{}
	generator := direction_dmi.NewDirectionDMIGenerator(genConfig, config)

	if err := generator.Initialize(); err != nil {
		log.Fatalf("❌ Erreur initialisation: %v", err)
	}

	// Calculer indicateurs
	fmt.Println("📊 Calcul indicateurs (VWMA, pente, ATR, DMI, DX, ADX)...")
	if err := generator.CalculateIndicators(klines); err != nil {
		log.Fatalf("❌ Erreur calcul indicateurs: %v", err)
	}

	// Détecter signaux
	fmt.Println("🔍 Détection signaux avec fenêtre de matching...")
	signaux, err := generator.DetectSignals(klines)
	if err != nil {
		log.Fatalf("❌ Erreur détection signaux: %v", err)
	}

	fmt.Printf("✅ %d signaux détectés\n", len(signaux))

	// Construire intervalles
	fmt.Println("\n🔄 Construction des intervalles...")
	intervalles := buildIntervalles(signaux, klines)
	fmt.Printf("✅ %d intervalles construits\n", len(intervalles))

	// Affichage détaillé des signaux
	displaySignalDetails(signaux)

	// Affichage des intervalles
	displayIntervalles(intervalles)

	// Affichage des statistiques
	stats := displayStatistics(intervalles)

	// Export JSON
	fmt.Println("\n💾 Export des résultats...")
	exportResults(config, klines, intervalles, stats)

	// Métriques du générateur
	metrics := generator.GetMetrics()
	fmt.Println("\n📈 Métriques générateur:")
	fmt.Printf("Signaux totaux     : %d\n", metrics.TotalSignals)

	fmt.Println("\n✅ Analyse terminée!")
}

// buildIntervalles construit les intervalles à partir des signaux
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

			// Extraire le mode depuis les métadonnées
			mode := "UNKNOWN"
			if modeVal, exists := currentEntry.Metadata["mode"]; exists {
				mode = modeVal.(string)
			}

			intervalle := Intervalle{
				Numero:          intervalNum,
				Type:            currentEntry.Type,
				Mode:            mode,
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

// displaySignalDetails affiche les détails des signaux
func displaySignalDetails(signaux []signals.Signal) {
	if len(signaux) == 0 {
		fmt.Println("\n⚠️  Aucun signal détecté")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("DÉTAIL DES SIGNAUX DÉTECTÉS")
	fmt.Println(strings.Repeat("=", 120))

	entryCount := 0
	exitCount := 0

	for i, sig := range signaux {
		fmt.Printf("\n🔸 Signal %d: %s %s %s (Conf: %.1f)\n",
			i+1, sig.Action, sig.Type, 
			sig.Metadata["mode"], sig.Confidence)
		fmt.Printf("   📅 %s | 💰 $%.4f\n",
			sig.Timestamp.Format("2006-01-02 15:04"), sig.Price)
		
		if sig.Metadata["vwma_direction"] != nil {
			fmt.Printf("   📊 VWMA: %.4f (pente %s)\n",
				sig.Metadata["vwma_slope"], sig.Metadata["vwma_direction"])
		}
		if sig.Metadata["dmi_di_plus"] != nil {
			fmt.Printf("   📡 DMI: DI+=%.2f DI-=%.2f (gap=%.2f)\n",
				sig.Metadata["dmi_di_plus"], sig.Metadata["dmi_di_minus"],
				sig.Metadata["dmi_gap"])
		}
		if sig.Metadata["dx"] != nil {
			fmt.Printf("   🎯 DX/ADX: %.2f/%.2f (gap=%.2f)\n",
				sig.Metadata["dx"], sig.Metadata["adx"],
				sig.Metadata["dx_gap"])
		}
		if sig.Metadata["window_start"] != nil {
			fmt.Printf("   🪟 Fenêtre: [%d → %d]\n",
				sig.Metadata["window_start"], sig.Metadata["window_end"])
		}

		if sig.Action == signals.SignalActionEntry {
			entryCount++
		} else {
			exitCount++
		}
	}

	fmt.Printf("\n📊 Résumé signaux: %d ENTRY | %d EXIT\n", entryCount, exitCount)
}

// displayIntervalles affiche les intervalles
func displayIntervalles(intervalles []Intervalle) {
	if len(intervalles) == 0 {
		fmt.Println("\n⚠️  Aucun intervalle trouvé")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("INTERVALLES DÉTECTÉS")
	fmt.Println(strings.Repeat("=", 100))

	// Trier par date
	sort.Slice(intervalles, func(i, j int) bool {
		return intervalles[i].DateDebut.Before(intervalles[j].DateDebut)
	})

	for _, inter := range intervalles {
		typeIcon := "📈"
		if inter.Type == signals.SignalTypeShort {
			typeIcon = "📉"
		}

		modeIcon := "🎯"
		if inter.Mode == "COUNTER_TREND" {
			modeIcon = "🔄"
		}

		fmt.Printf("%s %s #%d (%s %s): $%.4f → $%.4f = %+.2f%% (%d bougies)\n",
			typeIcon, modeIcon, inter.Numero,
			inter.Type, inter.Mode,
			inter.PrixDebut, inter.PrixFin,
			inter.VariationCaptee, inter.NbBougies)
		fmt.Printf("     📅 %s → %s (durée: %v)\n",
			inter.DateDebut.Format("01-02 15:04"),
			inter.DateFin.Format("01-02 15:04"),
			inter.DateFin.Sub(inter.DateDebut))
	}
}

// displayStatistics affiche les statistiques et retourne le résumé
func displayStatistics(intervalles []Intervalle) StatsSummary {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES DIRECTION+DMI")
	fmt.Println(strings.Repeat("=", 100))

	// Compteurs par type et mode
	longTrend := 0
	longCT := 0
	shortTrend := 0
	shortCT := 0
	bougiesLongTrend := 0
	bougiesLongCT := 0
	bougiesShortTrend := 0
	bougiesShortCT := 0
	varLongTrend := 0.0
	varLongCT := 0.0
	varShortTrend := 0.0
	varShortCT := 0.0

	for _, inter := range intervalles {
		if inter.Type == signals.SignalTypeLong {
			if inter.Mode == "TREND" {
				longTrend++
				bougiesLongTrend += inter.NbBougies
				varLongTrend += inter.VariationCaptee
			} else {
				longCT++
				bougiesLongCT += inter.NbBougies
				varLongCT += inter.VariationCaptee
			}
		} else {
			if inter.Mode == "TREND" {
				shortTrend++
				bougiesShortTrend += inter.NbBougies
				varShortTrend += inter.VariationCaptee
			} else {
				shortCT++
				bougiesShortCT += inter.NbBougies
				varShortCT += inter.VariationCaptee
			}
		}
	}

	fmt.Println("\n📊 RÉPARTITION PAR TYPE ET MODE:")
	fmt.Printf("  Total intervalles     : %d\n", len(intervalles))
	fmt.Printf("  - LONG Tendance (🎯)  : %d intervalles (%d bougies)\n", longTrend, bougiesLongTrend)
	fmt.Printf("  - LONG Contre-T (🔄)  : %d intervalles (%d bougies)\n", longCT, bougiesLongCT)
	fmt.Printf("  - SHORT Tendance (🎯) : %d intervalles (%d bougies)\n", shortTrend, bougiesShortTrend)
	fmt.Printf("  - SHORT Contre-T (🔄) : %d intervalles (%d bougies)\n", shortCT, bougiesShortCT)

	fmt.Println("\n💰 VARIATIONS CAPTÉES:")
	if longTrend > 0 {
		fmt.Printf("  - LONG Tendance       : %+.2f%% total, %+.2f%% moyen\n",
			varLongTrend, varLongTrend/float64(longTrend))
	}
	if longCT > 0 {
		fmt.Printf("  - LONG Contre-T       : %+.2f%% total, %+.2f%% moyen\n",
			varLongCT, varLongCT/float64(longCT))
	}
	if shortTrend > 0 {
		fmt.Printf("  - SHORT Tendance      : %+.2f%% total, %+.2f%% moyen\n",
			varShortTrend, varShortTrend/float64(shortTrend))
	}
	if shortCT > 0 {
		fmt.Printf("  - SHORT Contre-T      : %+.2f%% total, %+.2f%% moyen\n",
			varShortCT, varShortCT/float64(shortCT))
	}

	// Total capté = LONG - SHORT (car SHORT profitable = variation négative)
	totalCapte := (varLongTrend + varLongCT) - (varShortTrend + varShortCT)

	fmt.Printf("\n🎯 TOTAL CAPTÉ         : %.2f%% (bidirectionnel)\n", totalCapte)
	fmt.Printf("🏆 Performance         : %.2f%% sur %d trades\n", 
		totalCapte, len(intervalles))

	fmt.Println(strings.Repeat("=", 100))

	return StatsSummary{
		TotalIntervalles:    len(intervalles),
		LongTendance:        longTrend,
		LongContreTendance:  longCT,
		ShortTendance:       shortTrend,
		ShortContreTendance: shortCT,
		VariationLongTrend:  varLongTrend,
		VariationLongCT:     varLongCT,
		VariationShortTrend: varShortTrend,
		VariationShortCT:    varShortCT,
		TotalCapte:          totalCapte,
	}
}

// exportResults exporte les résultats au format JSON
func exportResults(config direction_dmi.Config, klines []signals.Kline, intervalles []Intervalle, stats StatsSummary) {
	// Créer nom de dossier avec paramètres
	folderName := fmt.Sprintf("direction_dmi_VWMA%d_Slope%d_K%d_ATR%d_Coef%.2f_DMI%d_GapDI%.1f_GapDX%.1f_WM%d_ET%t_ECT%t_XT%t_XCT%t",
		config.VWMAPeriod, config.SlopePeriod, config.KConfirmation,
		config.ATRPeriod, config.ATRCoefficient, config.DMIPeriod,
		config.GammaGapDI, config.GammaGapDX, config.WindowMatching,
		config.EnableEntryTrend, config.EnableEntryCounterTrend,
		config.EnableExitTrend, config.EnableExitCounterTrend)

	outputDir := filepath.Join("out", folderName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("⚠️  Erreur création dossier: %v\n", err)
		return
	}

	// Préparer données export
	exportData := ExportData{
		Config:      config,
		Klines:      klines,
		Intervalles: intervalles,
		Stats:       stats,
	}

	// Export klines
	klinesFile := filepath.Join(outputDir, "klines.json")
	if err := saveJSON(klinesFile, klines); err != nil {
		fmt.Printf("⚠️  Erreur export klines: %v\n", err)
	} else {
		fmt.Printf("✅ Klines exportées: %s\n", klinesFile)
	}

	// Export intervalles
	intervallesFile := filepath.Join(outputDir, "intervalles.json")
	if err := saveJSON(intervallesFile, intervalles); err != nil {
		fmt.Printf("⚠️  Erreur export intervalles: %v\n", err)
	} else {
		fmt.Printf("✅ Intervalles exportés: %s\n", intervallesFile)
	}

	// Export données complètes
	fullFile := filepath.Join(outputDir, "full_data.json")
	if err := saveJSON(fullFile, exportData); err != nil {
		fmt.Printf("⚠️  Erreur export complet: %v\n", err)
	} else {
		fmt.Printf("✅ Données complètes: %s\n", fullFile)
	}
}

// saveJSON sauvegarde données au format JSON
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
