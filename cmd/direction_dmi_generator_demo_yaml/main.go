package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/direction_dmi"
	"agent-economique/internal/shared"
)

// Configuration depuis YAML
const (
	SYMBOL     = "SOL_USDT"
	NB_CANDLES = 500
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
	Config          direction_dmi.Config `json:"config"`
	Klines          []signals.Kline      `json:"klines"`
	Intervalles     []Intervalle         `json:"intervalles"`
	Stats           StatsSummary         `json:"stats"`
	TrackedSignals  []TrackedSignalExport `json:"tracked_signals"`
}

// TrackedSignalExport version exportable des signaux trackés avec dates
type TrackedSignalExport struct {
	Index           int     `json:"index"`
	Timestamp       string  `json:"timestamp"`
	Price           float64 `json:"price"`
	
	// État des conditions
	VWMAFound       bool    `json:"vwma_found"`
	DMIFound        bool    `json:"dmi_found"`
	DXADXFound      bool    `json:"dxadx_found"`
	
	// Fenêtre avec dates
	WindowStart     int     `json:"window_start"`
	WindowEnd       int     `json:"window_end"`
	WindowStartTime string  `json:"window_start_time"`
	WindowEndTime   string  `json:"window_end_time"`
	
	// Signaux constituants 
	VWMASignal      *SignalVWMAExport  `json:"vwma_signal,omitempty"`
	DMISignal       *SignalDMIExport   `json:"dmi_signal,omitempty"`
	DXADXSignal     *SignalDXADXExport `json:"dxadx_signal,omitempty"`
	
	// Statut
	IsValid         bool   `json:"is_valid"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	
	// Signal classifié (si valide)
	Action          string `json:"action,omitempty"`
	Type            string `json:"type,omitempty"`
	Mode            string `json:"mode,omitempty"`
}

// Structures pour signaux constituants export
type SignalVWMAExport struct {
	Index     int     `json:"index"`
	Timestamp string  `json:"timestamp"`
	Direction string  `json:"direction"`
	Slope     float64 `json:"slope"`
}

type SignalDMIExport struct {
	Index       int     `json:"index"`
	Timestamp   string  `json:"timestamp"`
	Direction   string  `json:"direction"`
	DIPlus      float64 `json:"di_plus"`
	DIMinus     float64 `json:"di_minus"`
	GapDI       float64 `json:"gap_di"`
	IsCrossover bool    `json:"is_crossover"`
}

type SignalDXADXExport struct {
	Index     int     `json:"index"`
	Timestamp string  `json:"timestamp"`
	Direction string  `json:"direction"`
	DX        float64 `json:"dx"`
	ADX       float64 `json:"adx"`
	GapDX     float64 `json:"gap_dx"`
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
	fmt.Println("  DIRECTION+DMI GENERATOR DEMO - Configuration YAML")
	fmt.Println("═══════════════════════════════════════════════════")

	// Charger configuration depuis YAML
	fmt.Println("\n📝 Chargement configuration YAML...")
	sharedConfig, err := shared.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ Erreur chargement config: %v", err)
	}

	// Extraire config Direction+DMI
	dmiConfigYAML := sharedConfig.Strategy.DirectionDMIConfig
	
	// Vérifier si config Direction+DMI est définie
	if dmiConfigYAML.VWMAPeriod == 0 {
		log.Fatalf("❌ Configuration Direction+DMI vide dans YAML. Vérifiez config/config.yaml section 'direction_dmi'")
	}

	// Convertir en config interne
	config := direction_dmi.Config{
		// Paramètres VWMA
		VWMAPeriod:          dmiConfigYAML.VWMAPeriod,
		SlopePeriod:         dmiConfigYAML.SlopePeriod,
		KConfirmation:       dmiConfigYAML.KConfirmation,
		UseDynamicThreshold: dmiConfigYAML.UseDynamicThreshold,
		ATRPeriod:           dmiConfigYAML.ATRPeriod,
		ATRCoefficient:      dmiConfigYAML.ATRCoefficient,
		FixedThreshold:      dmiConfigYAML.FixedThreshold,

		// Paramètres DMI
		DMIPeriod:           dmiConfigYAML.DMIPeriod,
		DMISmooth:           dmiConfigYAML.DMISmooth,
		GammaGapDI:          dmiConfigYAML.GammaGapDI,
		GammaGapDX:          dmiConfigYAML.GammaGapDX,
		WindowGammaValidate: dmiConfigYAML.WindowGammaValidate,
		WindowMatching:      dmiConfigYAML.WindowMatching,

		// Flags
		EnableEntryTrend:        dmiConfigYAML.EnableEntryTrend,
		EnableEntryCounterTrend: dmiConfigYAML.EnableEntryCounterTrend,
		EnableExitTrend:         dmiConfigYAML.EnableExitTrend,
		EnableExitCounterTrend:  dmiConfigYAML.EnableExitCounterTrend,
	}

	fmt.Println("\n📊 Configuration depuis YAML:")
	fmt.Printf("Symbole               : %s\n", SYMBOL)
	fmt.Printf("Timeframe             : %s\n", config.UseDynamicThreshold) // Utiliser depuis config
	fmt.Printf("Nb bougies            : %d\n", NB_CANDLES)
	fmt.Println("\n🎯 Paramètres VWMA (Direction):")
	fmt.Printf("VWMA période          : %d\n", config.VWMAPeriod)
	fmt.Printf("Slope période         : %d\n", config.SlopePeriod)
	fmt.Printf("K-confirmation        : %d\n", config.KConfirmation)
	fmt.Printf("Seuil dynamique       : %t\n", config.UseDynamicThreshold)
	fmt.Printf("ATR période           : %d\n", config.ATRPeriod)
	fmt.Printf("ATR coefficient       : %.2f\n", config.ATRCoefficient)
	if !config.UseDynamicThreshold {
		fmt.Printf("Seuil fixe            : %.2f\n", config.FixedThreshold)
	}
	fmt.Println("\n📡 Paramètres DMI:")
	fmt.Printf("DMI période           : %d\n", config.DMIPeriod)
	fmt.Printf("DMI smooth            : %d\n", config.DMISmooth)
	fmt.Printf("Gamma gap DI          : %.1f%%\n", config.GammaGapDI)
	fmt.Printf("Gamma gap DX          : %.1f%%\n", config.GammaGapDX)
	fmt.Printf("Window validation     : %d\n", config.WindowGammaValidate)
	fmt.Printf("Window matching       : %d\n", config.WindowMatching)
	fmt.Println("\n🎛️ Flags d'activation:")
	fmt.Printf("Entry Tendance        : %t\n", config.EnableEntryTrend)
	fmt.Printf("Entry Contre-T        : %t\n", config.EnableEntryCounterTrend)
	fmt.Printf("Exit Tendance         : %t\n", config.EnableExitTrend)
	fmt.Printf("Exit Contre-T         : %t\n", config.EnableExitCounterTrend)

	// Récupérer les données
	fmt.Println("\n📂 Récupération des klines depuis Gate.io...")
	gateioClient := gateio.NewClient()
	gateioKlines, err := gateioClient.GetKlines(context.Background(), SYMBOL, dmiConfigYAML.Timeframe, NB_CANDLES)
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

	// Affichage des 20 dernières valeurs DMI (pour comparaison)
	displayDMIValues(generator, klines)

	// Affichage détaillé des signaux
	displaySignalDetails(signaux, klines)
	
	// Affichage de tous les signaux trackés (debug)
	displayTrackedSignals(generator, klines)

	// Affichage des intervalles
	displayIntervalles(intervalles)

	// Affichage des statistiques
	stats := displayStatistics(intervalles)

	// Export JSON
	fmt.Println("\n💾 Export des résultats...")
	exportResults(config, klines, intervalles, stats, generator, "YAML")

	// Métriques du générateur
	metrics := generator.GetMetrics()
	fmt.Println("\n📈 Métriques générateur:")
	fmt.Printf("Signaux totaux        : %d\n", metrics.TotalSignals)
	
	// Debug positions ouvertes
	fmt.Printf("Positions ouvertes    : %d\n", generator.GetOpenPositionsCount())

	fmt.Println("\n✅ Analyse terminée avec configuration YAML!")
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
func displaySignalDetails(signaux []signals.Signal, klines []signals.Kline) {
	if len(signaux) == 0 {
		fmt.Println("\n⚠️  Aucun signal détecté")
		return
	}

	// Trier les signaux par timestamp pour ordre chronologique
	sort.Slice(signaux, func(i, j int) bool {
		return signaux[i].Timestamp.Before(signaux[j].Timestamp)
	})

	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("DÉTAIL DES SIGNAUX DÉTECTÉS (Configuration YAML)")
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
			windowStart := sig.Metadata["window_start"].(int)
			windowEnd := sig.Metadata["window_end"].(int)
			
			if windowStart < len(klines) && windowEnd < len(klines) {
				windowStartTime := klines[windowStart].OpenTime.Format("01-02 15:04")
				windowEndTime := klines[windowEnd].OpenTime.Format("01-02 15:04")
				fmt.Printf("   🪟 Fenêtre: [%s → %s] (#%d→#%d)\n",
					windowStartTime, windowEndTime, windowStart, windowEnd)
			} else {
				fmt.Printf("   🪟 Fenêtre: [%d → %d]\n", windowStart, windowEnd)
			}
		}

		if sig.Action == signals.SignalActionEntry {
			entryCount++
		} else {
			exitCount++
		}
	}

	fmt.Printf("\n📊 Résumé signaux: %d ENTRY | %d EXIT\n", entryCount, exitCount)
}

// displayTrackedSignals affiche tous les signaux trackés avec leur statut (debug)
func displayTrackedSignals(generator *direction_dmi.DirectionDMIGenerator, klines []signals.Kline) {
	tracked := generator.GetTrackedSignals()
	
	if len(tracked) == 0 {
		fmt.Println("\n⚠️  Aucun signal tracké")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("TOUS LES SIGNAUX TRACKÉS (Debug - Configuration YAML)")
	fmt.Println(strings.Repeat("=", 140))

	validCount := 0
	invalidCount := 0
	statusCounts := make(map[string]int)

	// Statistiques par statut
	for _, t := range tracked {
		statusCounts[t.Status]++
		if t.IsValid {
			validCount++
		} else {
			invalidCount++
		}
	}

	fmt.Printf("\n📊 Résumé: %d signaux analysés (%d valides, %d invalides)\n", len(tracked), validCount, invalidCount)
	fmt.Println("\n📋 Répartition par statut:")
	for status, count := range statusCounts {
		fmt.Printf("   • %s: %d\n", status, count)
	}

	// Affichage détaillé (limité aux 20 premiers pour éviter spam)
	fmt.Println("\n🔍 Détail des signaux (20 premiers):")
	displayCount := len(tracked)
	if displayCount > 20 {
		displayCount = 20
	}

	for i := 0; i < displayCount; i++ {
		t := tracked[i]
		
		statusIcon := "❌"
		if t.IsValid {
			statusIcon = "✅"
		}

		fmt.Printf("\n%s #%d [%d] %s - %s\n", 
			statusIcon, i+1, t.Index, 
			t.Timestamp.Format("01-02 15:04"), t.Status)
		
		// Conditions trouvées avec dates des fenêtres
		windowStartTime := klines[t.WindowStart].OpenTime.Format("01-02 15:04")
		windowEndTime := klines[t.WindowEnd].OpenTime.Format("01-02 15:04")
		fmt.Printf("   🎯 Conditions: VWMA=%t DMI=%t DX/ADX=%t | Fenêtre=[%s→%s] (#%d→#%d)\n",
			t.VWMAFound, t.DMIFound, t.DXADXFound, 
			windowStartTime, windowEndTime, t.WindowStart, t.WindowEnd)
		
		// Détails des signaux constituants si trouvés
		if t.VWMASignal != nil {
			fmt.Printf("   📊 VWMA: %.4f (%s) à #%d\n",
				t.VWMASignal.Slope, t.VWMASignal.Direction, t.VWMASignal.Index)
		}
		if t.DMISignal != nil {
			crossType := "Position"
			if t.DMISignal.IsCrossover {
				crossType = "Croisement"
			}
			fmt.Printf("   📡 DMI: DI+=%.2f DI-=%.2f (%s, gap=%.2f, %s) à #%d\n",
				t.DMISignal.DIPlus, t.DMISignal.DIMinus, 
				t.DMISignal.Direction, t.DMISignal.GapDI, crossType, t.DMISignal.Index)
		}
		if t.DXADXSignal != nil {
			fmt.Printf("   🎯 DX/ADX: %.2f/%.2f (%s, gap=%.2f) à #%d\n",
				t.DXADXSignal.DX, t.DXADXSignal.ADX,
				t.DXADXSignal.Direction, t.DXADXSignal.GapDX, t.DXADXSignal.Index)
		}
		
		// Signal classifié si valide
		if t.IsValid {
			fmt.Printf("   🚀 Signal: %s %s %s\n", t.Action, t.Type, t.Mode)
		}
		
		// Raison d'invalidation
		if !t.IsValid {
			fmt.Printf("   ❌ Raison: %s\n", t.Reason)
		}
	}

	if len(tracked) > 20 {
		fmt.Printf("\n... (%d signaux supplémentaires non affichés)\n", len(tracked)-20)
	}
}

// displayIntervalles affiche les intervalles
func displayIntervalles(intervalles []Intervalle) {
	if len(intervalles) == 0 {
		fmt.Println("\n⚠️  Aucun intervalle trouvé")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("INTERVALLES DÉTECTÉS (Configuration YAML)")
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
	fmt.Println("STATISTIQUES DIRECTION+DMI (Configuration YAML)")
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
func exportResults(config direction_dmi.Config, klines []signals.Kline, intervalles []Intervalle, stats StatsSummary, generator *direction_dmi.DirectionDMIGenerator, source string) {
	// Créer nom de dossier avec paramètres et source
	folderName := fmt.Sprintf("direction_dmi_%s_VWMA%d_Slope%d_K%d_ATR%d_Coef%.2f_DMI%d_GapDI%.1f_GapDX%.1f_WM%d_ET%t_ECT%t_XT%t_XCT%t",
		source,
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

	// Convertir signaux trackés pour export  
	tracked := generator.GetTrackedSignals()
	trackedExport := make([]TrackedSignalExport, len(tracked))
	
	for i, t := range tracked {
		exp := TrackedSignalExport{
			Index:       t.Index,
			Timestamp:   t.Timestamp.Format("2006-01-02T15:04:05Z"),
			Price:       t.Price,
			VWMAFound:   t.VWMAFound,
			DMIFound:    t.DMIFound,
			DXADXFound:  t.DXADXFound,
			WindowStart: t.WindowStart,
			WindowEnd:   t.WindowEnd,
			WindowStartTime: klines[t.WindowStart].OpenTime.Format("2006-01-02T15:04:05Z"),
			WindowEndTime:   klines[t.WindowEnd].OpenTime.Format("2006-01-02T15:04:05Z"),
			IsValid:     t.IsValid,
			Status:      t.Status,
			Reason:      t.Reason,
			Action:      t.Action,
			Type:        t.Type,
			Mode:        t.Mode,
		}
		
		// Convertir signaux constituants 
		if t.VWMASignal != nil {
			exp.VWMASignal = &SignalVWMAExport{
				Index:     t.VWMASignal.Index,
				Timestamp: t.VWMASignal.Timestamp.Format("2006-01-02T15:04:05Z"),
				Direction: t.VWMASignal.Direction,
				Slope:     t.VWMASignal.Slope,
			}
		}
		
		if t.DMISignal != nil {
			exp.DMISignal = &SignalDMIExport{
				Index:       t.DMISignal.Index,
				Timestamp:   t.DMISignal.Timestamp.Format("2006-01-02T15:04:05Z"),
				Direction:   t.DMISignal.Direction,
				DIPlus:      t.DMISignal.DIPlus,
				DIMinus:     t.DMISignal.DIMinus,
				GapDI:       t.DMISignal.GapDI,
				IsCrossover: t.DMISignal.IsCrossover,
			}
		}
		
		if t.DXADXSignal != nil {
			exp.DXADXSignal = &SignalDXADXExport{
				Index:     t.DXADXSignal.Index,
				Timestamp: t.DXADXSignal.Timestamp.Format("2006-01-02T15:04:05Z"),
				Direction: t.DXADXSignal.Direction,
				DX:        t.DXADXSignal.DX,
				ADX:       t.DXADXSignal.ADX,
				GapDX:     t.DXADXSignal.GapDX,
			}
		}
		
		trackedExport[i] = exp
	}

	// Préparer données export
	exportData := ExportData{
		Config:         config,
		Klines:         klines,
		Intervalles:    intervalles,
		Stats:          stats,
		TrackedSignals: trackedExport,
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

	// Export signaux trackés (nouveau)
	trackedFile := filepath.Join(outputDir, "tracked_signals.json")
	if err := saveJSON(trackedFile, trackedExport); err != nil {
		fmt.Printf("⚠️  Erreur export signaux trackés: %v\n", err)
	} else {
		fmt.Printf("✅ Signaux trackés exportés: %s (%d signaux)\n", trackedFile, len(trackedExport))
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

// displayDMIValues affiche les 20 dernières valeurs DMI pour comparaison avec validateur
func displayDMIValues(generator *direction_dmi.DirectionDMIGenerator, klines []signals.Kline) {
	diPlus, diMinus, adx := generator.GetDMIValues()
	
	if len(diPlus) == 0 || len(diMinus) == 0 || len(adx) == 0 {
		fmt.Println("\n⚠️  Aucune valeur DMI calculée")
		return
	}

	fmt.Println("\n📊 DMI TV STANDARD - DIRECTION DMI (20 dernières valeurs):")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Printf("%-12s %-10s %-10s %-10s %-10s %-10s %-12s %-10s\n", 
		"TIME", "CLOSE", "+DI", "-DI", "DX", "ADX", "TREND", "SIGNAL")
	fmt.Println(strings.Repeat("-", 95))

	// Calculer DX manuellement car pas exposé directement
	start := len(klines) - 20
	if start < 14 { start = 14 } // DMI needs at least 14 periods

	for i := start; i < len(klines) && i < len(diPlus) && i < len(diMinus) && i < len(adx); i++ {
		// Calculer DX = |DI+ - DI-| / (DI+ + DI-) * 100
		dx := 0.0
		if diPlus[i]+diMinus[i] != 0 {
			dx = math.Abs(diPlus[i]-diMinus[i]) / (diPlus[i]+diMinus[i]) * 100
		}

		// Déterminer trend et signal
		trend := "Faible"
		if adx[i] > 25 {
			trend = "Fort"
		}
		if diMinus[i] > diPlus[i] {
			trend += " ↓"
		} else {
			trend += " ↑"
		}

		signal := "NEUTRE"
		if adx[i] > 25 {
			if diPlus[i] > diMinus[i] {
				signal = "ACHAT"
			} else {
				signal = "VENTE"
			}
		}

		fmt.Printf("%-12s %-10.2f %-10.2f %-10.2f %-10.2f %-10.2f %-12s %-10s\n",
			klines[i].OpenTime.Format("15:04"),
			klines[i].Close,
			diPlus[i],
			diMinus[i], 
			dx,
			adx[i],
			trend,
			signal)
	}
	fmt.Println(strings.Repeat("=", 95))
}
