package signals

// Ce fichier contient des exemples d'intégration des générateurs
// dans scalping_live_bybit (ne pas compiler directement)

/*

// ============================================================================
// EXEMPLE 1 : INTÉGRATION DANS ScalpingLiveBybitApp
// ============================================================================

// Dans cmd/scalping_live_bybit/app_live.go

import (
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/direction"
	"agent-economique/internal/signals/trend"
)

// Ajouter champ dans ScalpingLiveBybitApp
type ScalpingLiveBybitApp struct {
	// ... existing fields ...
	
	// Générateur de signaux (nil = stratégie scalping classique)
	signalGenerator signals.Generator
	generatorType   string // "direction", "trend", ou "scalping"
}

// Modifier NewScalpingLiveBybitApp
func NewScalpingLiveBybitApp(config *shared.Config, mode string, generatorType string) *ScalpingLiveBybitApp {
	// ... existing code ...
	
	// Créer générateur selon le type
	var generator signals.Generator
	
	switch generatorType {
	case "direction":
		dirConfig := direction.Config{
			VWMAPeriod:          3,
			SlopePeriod:         2,
			KConfirmation:       2,
			UseDynamicThreshold: true,
			ATRPeriod:           14,
			ATRCoefficient:      1.0,
		}
		generator = direction.NewDirectionGenerator(dirConfig)
		fmt.Println("🎯 Générateur: DIRECTION (VWMA6 + ATR)")
		
	case "trend":
		trendConfig := trend.Config{
			VwmaRapide:          6,
			VwmaLent:            24,
			DmiPeriode:          14,
			DmiSmooth:           3,
			AtrPeriode:          30,
			GammaGapVWMA:        0.5,
			GammaGapDI:          5.0,
			GammaGapDX:          5.0,
			VolatiliteMin:       0.3,
			WindowGammaValidate: 5,
			WindowW:             10,
		}
		generator = trend.NewTrendGenerator(trendConfig)
		fmt.Println("🎯 Générateur: TREND (VWMA+DMI)")
		
	default:
		generator = nil
		fmt.Println("🎯 Générateur: SCALPING (classique)")
	}
	
	// Initialiser générateur si présent
	if generator != nil {
		genConfig := signals.GeneratorConfig{
			Symbol:      config.BinanceData.Symbols[0],
			Timeframe:   config.Strategy.ScalpingConfig.Timeframe,
			HistorySize: 300,
		}
		if err := generator.Initialize(genConfig); err != nil {
			log.Printf("⚠️  Erreur initialisation générateur: %v", err)
			generator = nil
		}
	}
	
	return &ScalpingLiveBybitApp{
		// ... existing fields ...
		signalGenerator: generator,
		generatorType:   generatorType,
	}
}

// Modifier processMarker pour utiliser le générateur
func (app *ScalpingLiveBybitApp) processMarker(candleTime time.Time, klines []Kline) error {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🕐 %s | TRAITEMENT CLÔTURE\n", candleTime.Format("15:04:05"))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Si générateur configuré, l'utiliser
	if app.signalGenerator != nil {
		return app.processWithGenerator(klines)
	}
	
	// Sinon, utiliser stratégie scalping classique
	return app.createStrategy(klines)
}

// Nouvelle fonction : processWithGenerator
func (app *ScalpingLiveBybitApp) processWithGenerator(klines []Kline) error {
	// 1️⃣ Convertir vers format unifié
	unifiedKlines := make([]signals.Kline, len(klines))
	for i, k := range klines {
		unifiedKlines[i] = signals.Kline{
			OpenTime: k.OpenTime,
			Open:     k.Open,
			High:     k.High,
			Low:      k.Low,
			Close:    k.Close,
			Volume:   k.Volume,
		}
	}
	
	// 2️⃣ Calculer indicateurs
	fmt.Println("📊 Calcul des indicateurs...")
	if err := app.signalGenerator.CalculateIndicators(unifiedKlines); err != nil {
		return fmt.Errorf("erreur calcul indicateurs: %w", err)
	}
	
	// 3️⃣ Détecter signaux
	fmt.Println("🔍 Détection des signaux...")
	newSignals, err := app.signalGenerator.DetectSignals(unifiedKlines)
	if err != nil {
		return fmt.Errorf("erreur détection signaux: %w", err)
	}
	
	// 4️⃣ Traiter les nouveaux signaux
	if len(newSignals) > 0 {
		fmt.Printf("   🎯 %d signal(aux) détecté(s)!\n\n", len(newSignals))
		
		symbol := app.config.BinanceData.Symbols[0]
		
		for _, sig := range newSignals {
			app.handleSignal(sig, symbol)
		}
		
		// 5️⃣ Sauvegarder avec lock (rapide)
		app.mu.Lock()
		// Convertir vers ancien format Signal si nécessaire
		// ou ajouter nouveau champ pour stocker signals.Signal
		app.mu.Unlock()
	} else {
		fmt.Println("   ℹ️  Aucun signal détecté\n")
	}
	
	// 6️⃣ Afficher métriques
	metrics := app.signalGenerator.GetMetrics()
	fmt.Printf("📊 Métriques %s:\n", app.signalGenerator.Name())
	fmt.Printf("   Total: %d signaux (%d ENTRY, %d EXIT)\n",
		metrics.TotalSignals, metrics.EntrySignals, metrics.ExitSignals)
	fmt.Printf("   LONG: %d | SHORT: %d\n",
		metrics.LongSignals, metrics.ShortSignals)
	fmt.Printf("   Confiance moy: %.1f%%\n", metrics.AvgConfidence*100)
	
	return nil
}

// Nouvelle fonction : handleSignal
func (app *ScalpingLiveBybitApp) handleSignal(sig signals.Signal, symbol string) {
	// Affichage selon type d'action
	if sig.Action == signals.SignalActionEntry {
		fmt.Printf("   🟢 ENTRY %s @ %.2f\n", sig.Type, sig.Price)
		fmt.Printf("      Confiance: %.0f%% | Time: %s\n",
			sig.Confidence*100, sig.Timestamp.Format("15:04:05"))
		
		// Métadonnées spécifiques
		if app.generatorType == "direction" {
			if startIdx, ok := sig.Metadata["start_index"].(int); ok {
				fmt.Printf("      Index: %d | VWMA6: %.2f\n",
					startIdx, sig.Metadata["vwma6"])
			}
		} else if app.generatorType == "trend" {
			if motif, ok := sig.Metadata["motif"].(string); ok {
				fmt.Printf("      Motif: %s | Distance: %d bars\n",
					motif, sig.Metadata["distance_bars"])
			}
		}
		
	} else if sig.Action == signals.SignalActionExit {
		variation := 0.0
		if sig.EntryPrice != nil {
			variation = (sig.Price - *sig.EntryPrice) / *sig.EntryPrice * 100
		}
		
		fmt.Printf("   🔴 EXIT %s @ %.2f\n", sig.Type, sig.Price)
		fmt.Printf("      Confiance: %.0f%% | Variation: %+.2f%%\n",
			sig.Confidence*100, variation)
		
		if sig.EntryTime != nil && sig.Metadata["duration_bars"] != nil {
			duration := sig.Metadata["duration_bars"].(int)
			fmt.Printf("      Durée: %d bars | Entrée: %s\n",
				duration, sig.EntryTime.Format("15:04:05"))
		}
	}
	
	// Envoyer notification ntfy
	app.sendGeneratorSignalNotification(sig, symbol)
}

// Nouvelle fonction : sendGeneratorSignalNotification
func (app *ScalpingLiveBybitApp) sendGeneratorSignalNotification(sig signals.Signal, symbol string) {
	emoji := "📈"
	action := "ENTRY"
	if sig.Action == signals.SignalActionExit {
		emoji = "📉"
		action = "EXIT"
	}
	
	direction := "LONG"
	if sig.Type == signals.SignalTypeShort {
		direction = "SHORT"
	}
	
	message := fmt.Sprintf("%s %s %s\n\n", emoji, action, direction)
	message += fmt.Sprintf("💰 Prix: %.2f\n", sig.Price)
	message += fmt.Sprintf("🎯 Confiance: %.0f%%\n", sig.Confidence*100)
	message += fmt.Sprintf("⏰ %s\n", sig.Timestamp.Format("15:04:05"))
	message += fmt.Sprintf("📊 Symbole: %s\n", symbol)
	message += fmt.Sprintf("🔧 Générateur: %s\n", app.generatorType)
	
	if sig.Action == signals.SignalActionExit && sig.EntryPrice != nil {
		variation := (sig.Price - *sig.EntryPrice) / *sig.EntryPrice * 100
		message += fmt.Sprintf("📈 Variation: %+.2f%%\n", variation)
	}
	
	if err := app.notifier.SendNotification(message); err != nil {
		fmt.Printf("      ⚠️  Notification échouée: %v\n", err)
	} else {
		fmt.Printf("      ✅ Notification envoyée\n")
	}
}

// ============================================================================
// EXEMPLE 2 : MODIFICATION DU MAIN
// ============================================================================

// Dans cmd/scalping_live_bybit/main.go

func main() {
	fmt.Println("🎯 SCALPING LIVE BYBIT - Trading Production")
	fmt.Println("===========================================")
	
	// Parser arguments CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin config")
	symbol := flag.String("symbol", "", "Symbole - override config")
	generatorType := flag.String("generator", "scalping", "Type générateur: direction, trend, scalping")
	flag.Parse()
	
	// Charger configuration
	config, err := shared.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Erreur config: %v", err)
	}
	
	// Override symbol si fourni
	if *symbol != "" {
		config.BinanceData.Symbols = []string{*symbol}
	}
	
	// Afficher paramètres
	fmt.Printf("\n📊 Paramètres Trading:\n")
	fmt.Printf("   - Mode: live\n")
	fmt.Printf("   - Générateur: %s\n", *generatorType)
	fmt.Printf("   - Symbole: %s\n", config.BinanceData.Symbols[0])
	fmt.Printf("   - Timeframe: %s\n", config.Strategy.ScalpingConfig.Timeframe)
	
	// Créer application avec générateur
	app := NewScalpingLiveBybitApp(config, "live", *generatorType)
	
	// Lancer trading
	if err := app.Run(ctx); err != nil {
		log.Fatalf("❌ Erreur: %v", err)
	}
}

// ============================================================================
// EXEMPLE 3 : UTILISATION CLI
// ============================================================================

// Direction sur 1m
$ go run cmd/scalping_live_bybit/*.go --generator=direction --symbol=SOLUSDT

// Sortie attendue:
// 🎯 SCALPING LIVE BYBIT - Trading Production
// 🎯 Générateur: DIRECTION (VWMA6 + ATR)
// 📊 Paramètres Trading:
//    - Mode: live
//    - Générateur: direction
//    - Symbole: SOLUSDT
//
// 🕐 09:30:00 | TRAITEMENT CLÔTURE
// 📊 Calcul des indicateurs...
// 🔍 Détection des signaux...
//    🎯 2 signal(aux) détectés!
//
//    🟢 ENTRY LONG @ 161.25
//       Confiance: 70% | Time: 09:30:00
//       Index: 285 | VWMA6: 161.18
//       ✅ Notification envoyée
//
//    🔴 EXIT SHORT @ 161.25
//       Confiance: 85% | Variation: -1.54%
//       Durée: 47 bars | Entrée: 09:06:00
//       ✅ Notification envoyée

// Trend sur 5m
$ go run cmd/scalping_live_bybit/*.go --generator=trend --symbol=SOLUSDT

// Sortie attendue:
// 🎯 SCALPING LIVE BYBIT - Trading Production
// 🎯 Générateur: TREND (VWMA+DMI)
// 📊 Paramètres Trading:
//    - Mode: live
//    - Générateur: trend
//    - Symbole: SOLUSDT
//
// 🕐 09:30:00 | TRAITEMENT CLÔTURE
// 📊 Calcul des indicateurs...
// 🔍 Détection des signaux...
//    🎯 1 signal(aux) détectés!
//
//    🟢 ENTRY LONG @ 155.20
//       Confiance: 85% | Time: 09:30:00
//       Motif: VWMA→DMI (+2 bars) | Distance: 2 bars
//       ✅ Notification envoyée

// Scalping classique
$ go run cmd/scalping_live_bybit/*.go --generator=scalping --symbol=SOLUSDT

// Sortie attendue:
// 🎯 SCALPING LIVE BYBIT - Trading Production
// 🎯 Générateur: SCALPING (classique)
// (Comportement actuel inchangé)

*/
