package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"

	"agent-economique/internal/indicators"
	"agent-economique/internal/notifications"
	"agent-economique/internal/shared"
)

// ScalpingPaperApp manages the scalping strategy in paper/live mode
type ScalpingPaperApp struct {
	config *shared.Config
	mode   string // "paper" ou "live"

	// Données
	klines []Kline

	// Dernière bougie connue (pour détecter nouvelles clôtures)
	lastKnownTimestamp int64

	// Stratégie scalping (recréée à chaque marqueur)
	strategy *ScalpingStrategy

	// Signaux détectés
	signals []Signal

	// Analyses en cours (persistent entre marqueurs)
	pendingAnalyses []PendingAnalysis

	// Métriques
	startTime time.Time

	// Client SDK Binance
	binanceClient *binance.Client

	// Client de notification ntfy
	notifier *notifications.NtfyClient

	// Mutex pour accès concurrents (goroutines async)
	mu sync.Mutex
}

// Kline représente une bougie OHLCV
type Kline struct {
	Timestamp        int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	QuoteAssetVolume float64
}

// Signal représente un signal de trading détecté
type Signal struct {
	GlobalN2Index       int
	GlobalN1Index       int
	GlobalIndex         int
	GlobalCrossingIndex int
	Type                string
	Timestamp           int64
	Price               float64
	Volume              float64
	QuoteVolume         float64
	CCI                 float64
	MFI                 float64
	StochK              float64
	StochD              float64
}

// PendingAnalysis représente une analyse en attente de validation
type PendingAnalysis struct {
	GlobalN2Index        int
	GlobalN1Index        int
	SignalType           string
	GlobalWindowEndIndex int
}

// ScalpingStrategy implémente la stratégie de scalping
type ScalpingStrategy struct {
	config           ScalpingConfig
	klines           []Kline
	cciValues        []float64
	mfiValues        []float64
	stochKValues     []float64
	stochDValues     []float64
	windowStartIndex int
	pendingAnalyses  []PendingAnalysis
}

// ScalpingConfig holds strategy parameters
type ScalpingConfig struct {
	CCISurachat      float64
	CCISurvente      float64
	MFISurachat      float64
	MFISurvente      float64
	StochSurachat    float64
	StochSurvente    float64
	ValidationWindow int
	VolumeEnabled    bool
	VolumeThreshold  float64
	VolumePeriod     int
	VolumeMaxExt     int
}

// NewScalpingPaperApp creates a new scalping paper/live application
func NewScalpingPaperApp(config *shared.Config, mode string) *ScalpingPaperApp {
	// Déterminer le topic ntfy selon le mode
	ntfyTopic := "scalping-paper"
	if mode == "live" {
		ntfyTopic = "scalping-live"
	}

	// Créer client ntfy
	notifier := notifications.NewNtfyClient("https://notifications.koyad.com", ntfyTopic)

	// Créer client Binance SDK (Spot API)
	var binanceClient *binance.Client
	if mode == "paper" {
		// Testnet Spot (lecture publique)
		binance.UseTestnet = true
		binanceClient = binance.NewClient("", "")
	} else {
		// Production Spot
		binance.UseTestnet = false
		binanceClient = binance.NewClient("", "")
	}

	return &ScalpingPaperApp{
		config:        config,
		mode:          mode,
		klines:        make([]Kline, 0, 300),
		signals:       make([]Signal, 0),
		binanceClient: binanceClient,
		notifier:      notifier,
	}
}

// Run executes the scalping engine in paper/live mode
func (app *ScalpingPaperApp) Run(ctx context.Context) error {
	app.startTime = time.Now()

	// Notification de démarrage (asynchrone pour ne pas bloquer)
	symbol := app.config.BinanceData.Symbols[0]
	timeframe := app.config.Strategy.ScalpingConfig.Timeframe
	startMsg := fmt.Sprintf("🚀 Démarrage Scalping %s\n\n📊 Symbole: %s\n⏱️ Timeframe: %s\n🔧 Mode: %s",
		app.mode, symbol, timeframe, app.mode)

	// Notification de démarrage ASYNCHRONE
	// Délai court maintenant que le timeout client est à 30s
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("🔔 Envoi notification démarrage...")
		if err := app.notifier.SendStatusNotification(startMsg); err != nil {
			fmt.Printf("⚠️  Échec: %v\n", err)
		} else {
			fmt.Println("✅ Notification démarrage envoyée")
		}
	}()

	// 1️⃣ Charger historique initial (300 dernières klines)
	fmt.Println("\n📂 Chargement historique initial...")
	if err := app.loadInitialKlines(); err != nil {
		app.notifier.SendErrorNotification(fmt.Sprintf("Erreur chargement initial: %v", err))
		return fmt.Errorf("erreur chargement initial: %w", err)
	}
	fmt.Printf("✅ %d klines initiales chargées\n", len(app.klines))

	// 2️⃣ Démarrer loop timer
	fmt.Println("\n🔄 Démarrage loop trading...")
	if err := app.runTimerLoop(ctx); err != nil {
		app.notifier.SendErrorNotification(fmt.Sprintf("Erreur loop: %v", err))
		return fmt.Errorf("erreur loop: %w", err)
	}

	// Notification d'arrêt (avec logs)
	fmt.Println("🔔 Envoi notification arrêt...")
	stopMsg := fmt.Sprintf("🛑 Arrêt Scalping %s\n\n📊 Signaux détectés: %d", app.mode, len(app.signals))
	if err := app.notifier.SendStatusNotification(stopMsg); err != nil {
		fmt.Printf("⚠️  Notification arrêt échouée: %v\n", err)
	} else {
		fmt.Println("✅ Notification arrêt envoyée")
	}

	return nil
}

// loadInitialKlines charge les 300 dernières klines via SDK Binance
func (app *ScalpingPaperApp) loadInitialKlines() error {
	symbol := app.config.BinanceData.Symbols[0]
	timeframe := app.config.Strategy.ScalpingConfig.Timeframe

	// Appel SDK Binance Futures
	klines, err := app.binanceClient.NewKlinesService().
		Symbol(symbol).
		Interval(timeframe).
		Limit(300).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("Binance SDK error: %w", err)
	}

	// Convertir Binance Klines → notre struct Kline
	app.klines = make([]Kline, len(klines))
	for i, k := range klines {
		app.klines[i] = Kline{
			Timestamp:        k.OpenTime,
			Open:             parseFloat(k.Open),
			High:             parseFloat(k.High),
			Low:              parseFloat(k.Low),
			Close:            parseFloat(k.Close),
			Volume:           parseFloat(k.Volume),
			QuoteAssetVolume: parseFloat(k.QuoteAssetVolume),
		}
	}

	// Mémoriser dernier timestamp connu
	if len(app.klines) > 0 {
		app.lastKnownTimestamp = app.klines[len(app.klines)-1].Timestamp
	}

	return nil
}

// runTimerLoop exécute la boucle principale toutes les 10 secondes (specs paper/live)
// Responsabilités:
// 1. Toutes les 10s : Mise à jour trailing stops (positions ouvertes)
// 2. Sur clôture bougie : Calcul indicateurs + détection signaux
// IMPORTANT: Synchronisé sur :00, :10, :20, :30, :40, :50 pour coïncider avec clôtures
func (app *ScalpingPaperApp) runTimerLoop(ctx context.Context) error {
	loopInterval := 10 * time.Second
	timeframe := app.config.Strategy.ScalpingConfig.Timeframe

	// 1️⃣ Calculer délai jusqu'au prochain multiple de 10 secondes
	now := time.Now()
	currentSecond := now.Second()
	secondsUntilNext := 10 - (currentSecond % 10)
	if secondsUntilNext == 10 {
		secondsUntilNext = 0 // Déjà sur un multiple de 10
	}
	nextSync := now.Add(time.Duration(secondsUntilNext) * time.Second).Truncate(time.Second)

	fmt.Printf("⏱️  Synchronisation sur multiples de 10s...\n")
	fmt.Printf("   Heure actuelle: %s\n", now.Format("15:04:05"))
	fmt.Printf("   Prochain tick: %s (dans %ds)\n", nextSync.Format("15:04:05"), secondsUntilNext)
	fmt.Printf("   Timeframe bougie: %s\n\n", timeframe)

	// 2️⃣ Attendre la synchronisation initiale
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(time.Duration(secondsUntilNext) * time.Second):
		// Premier tick synchronisé
		fmt.Printf("[%s] 🔔 Synchronisé!\n", time.Now().Format("15:04:05"))
		if err := app.processTimerTick(); err != nil {
			fmt.Printf("⚠️  Erreur tick initial: %v\n", err)
		}
	}

	// 3️⃣ Créer ticker pour les ticks suivants
	ticker := time.NewTicker(loopInterval)
	defer ticker.Stop()

	fmt.Printf("⏱️  Loop active (tick toutes les %v)\n\n", loopInterval)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n🛑 Arrêt demandé")
			return nil

		case <-ticker.C:
			if err := app.processTimerTick(); err != nil {
				fmt.Printf("⚠️  Erreur tick: %v\n", err)
			}
		}
	}
}

// processTimerTick traite un tick du timer (toutes les 10 secondes)
// ULTRA-RAPIDE : Lance uniquement les tâches async, ne bloque JAMAIS
// Logique complète selon docs/workflow/04_engine_temporal.md:
// - Toujours : mise à jour trailing stops si position ouverte (async)
// - Sur clôture : calcul indicateurs + détection signaux (async)
func (app *ScalpingPaperApp) processTimerTick() error {
	now := time.Now()
	fmt.Printf("[%s] 🔄 Tick...\n", now.Format("15:04:05"))

	// Lancer le traitement COMPLET en goroutine pour ne JAMAIS bloquer
	go app.processTickAsync()

	return nil
}

// processTickAsync effectue le traitement réel du tick en arrière-plan
func (app *ScalpingPaperApp) processTickAsync() {
	// 1️⃣ Fetch nouvelles klines (peut être lent, mais n'bloque pas le ticker)
	newKlines, err := app.fetchLatestKlines()
	if err != nil {
		fmt.Printf("   ⚠️  Erreur fetch klines: %v\n", err)
		return
	}

	// 2️⃣ Si position ouverte → Mise à jour trailing stop
	// TODO: Implémenter quand position management sera ajouté
	// if app.hasOpenPosition() {
	//     app.updateTrailingStop(newKlines)
	//     app.checkStopHit()
	// }

	// 3️⃣ Détecter bougies fermées (avec lock)
	completedCandles := app.detectNewCompletedCandles(newKlines)

	if len(completedCandles) > 0 {
		fmt.Printf("   📊 %d nouvelle(s) bougie(s) fermée(s)\n", len(completedCandles))

		// 4️⃣ Pour chaque bougie fermée → Calcul indicateurs + signaux
		// IMPORTANT: Traitement asynchrone pour ne pas bloquer même ce worker
		for _, candleTimestamp := range completedCandles {
			// Lancer dans une goroutine séparée
			go func(ts int64) {
				if err := app.processMarker(ts); err != nil {
					fmt.Printf("   ⚠️  Erreur traitement marqueur: %v\n", err)
				}
			}(candleTimestamp)
		}
	}
}

// fetchLatestKlines récupère les dernières klines via SDK Binance
func (app *ScalpingPaperApp) fetchLatestKlines() ([]Kline, error) {
	symbol := app.config.BinanceData.Symbols[0]
	timeframe := app.config.Strategy.ScalpingConfig.Timeframe

	// Récupérer seulement les 10 dernières via SDK
	binanceKlines, err := app.binanceClient.NewKlinesService().
		Symbol(symbol).
		Interval(timeframe).
		Limit(10).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("Binance SDK error: %w", err)
	}

	// Convertir
	klines := make([]Kline, len(binanceKlines))
	for i, k := range binanceKlines {
		klines[i] = Kline{
			Timestamp:        k.OpenTime,
			Open:             parseFloat(k.Open),
			High:             parseFloat(k.High),
			Low:              parseFloat(k.Low),
			Close:            parseFloat(k.Close),
			Volume:           parseFloat(k.Volume),
			QuoteAssetVolume: parseFloat(k.QuoteAssetVolume),
		}
	}

	return klines, nil
}

// detectNewCompletedCandles détecte les nouvelles bougies fermées
func (app *ScalpingPaperApp) detectNewCompletedCandles(newKlines []Kline) []int64 {
	app.mu.Lock()
	defer app.mu.Unlock()

	var completed []int64

	for _, kline := range newKlines {
		if kline.Timestamp > app.lastKnownTimestamp {
			// Nouvelle bougie détectée
			completed = append(completed, kline.Timestamp)

			// Ajouter à l'historique
			app.klines = append(app.klines, kline)

			// Garder seulement 300 dernières (rolling window)
			if len(app.klines) > 300 {
				app.klines = app.klines[len(app.klines)-300:]
			}

			// Mettre à jour dernier timestamp connu
			app.lastKnownTimestamp = kline.Timestamp
		}
	}

	return completed
}

// processMarker traite un marqueur de bougie fermée
// ATTENTION: Appelé en goroutine asynchrone, doit protéger les accès partagés
func (app *ScalpingPaperApp) processMarker(timestamp int64) error {
	t := time.Unix(timestamp/1000, 0)

	// Log marqueur avec bordure (comme scalping_engine)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🕐 %s | MARQUEUR 5M DÉTECTÉ\n", t.Format("15:04:05"))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// 1️⃣ Copier les données partagées avec lock (rapide)
	app.mu.Lock()
	klinesCopy := make([]Kline, len(app.klines))
	copy(klinesCopy, app.klines)
	pendingCopy := make([]PendingAnalysis, len(app.pendingAnalyses))
	copy(pendingCopy, app.pendingAnalyses)
	app.mu.Unlock()

	// 2️⃣ Créer stratégie avec window de 300 klines (sans lock)
	scalpingConfig := ScalpingConfig{
		CCISurachat:      app.config.Strategy.ScalpingConfig.CCISurachat,
		CCISurvente:      app.config.Strategy.ScalpingConfig.CCISurvente,
		MFISurachat:      app.config.Strategy.ScalpingConfig.MFISurachat,
		MFISurvente:      app.config.Strategy.ScalpingConfig.MFISurvente,
		StochSurachat:    app.config.Strategy.ScalpingConfig.StochSurachat,
		StochSurvente:    app.config.Strategy.ScalpingConfig.StochSurvente,
		ValidationWindow: app.config.Strategy.ScalpingConfig.ValidationWindow,
		VolumeEnabled:    true, // Toujours activé
		VolumeThreshold:  app.config.Strategy.ScalpingConfig.VolumeThreshold,
		VolumePeriod:     app.config.Strategy.ScalpingConfig.VolumePeriod,
		VolumeMaxExt:     app.config.Strategy.ScalpingConfig.VolumeMaxExt,
	}

	strategy := &ScalpingStrategy{
		config:           scalpingConfig,
		klines:           klinesCopy, // Copie locale
		windowStartIndex: 0,
		pendingAnalyses:  pendingCopy, // Copie locale
	}

	// 3️⃣ Calculer indicateurs (LENT, sans lock)
	if err := strategy.calculateIndicators(); err != nil {
		return fmt.Errorf("calcul indicateurs: %w", err)
	}

	// 4️⃣ Détecter signaux (sans lock)
	newSignals := strategy.DetectSignals()

	if len(newSignals) > 0 {
		fmt.Printf("   🎯 %d signal(aux) détecté(s)!\n", len(newSignals))

		symbol := app.config.BinanceData.Symbols[0]

		for _, sig := range newSignals {
			fmt.Printf("      → %s à %.2f (CCI=%.1f, MFI=%.1f, K=%.1f)\n",
				sig.Type, sig.Price, sig.CCI, sig.MFI, sig.StochK)

			// Envoyer notification ntfy (sans lock, I/O externe)
			signalInfo := notifications.SignalInfo{
				Type:   sig.Type,
				Symbol: symbol,
				Price:  sig.Price,
				Time:   timestampMsToTime(sig.Timestamp),
				CCI:    sig.CCI,
				MFI:    sig.MFI,
				StochK: sig.StochK,
				StochD: sig.StochD,
				Volume: sig.Volume,
				Mode:   app.mode,
			}

			if err := app.notifier.SendSignalNotification(signalInfo); err != nil {
				fmt.Printf("      ⚠️  Notification échouée: %v\n", err)
			} else {
				fmt.Printf("      ✅ Notification envoyée\n")
			}
		}

		// 5️⃣ Sauvegarder résultats avec lock (rapide)
		app.mu.Lock()
		app.signals = append(app.signals, newSignals...)
		app.pendingAnalyses = strategy.pendingAnalyses
		app.mu.Unlock()
	} else {
		// Pas de signaux mais on sauvegarde quand même les analyses en cours
		app.mu.Lock()
		app.pendingAnalyses = strategy.pendingAnalyses
		app.mu.Unlock()
	}

	return nil
}

// calculateIndicators calcule les indicateurs techniques
func (s *ScalpingStrategy) calculateIndicators() error {
	high := make([]float64, len(s.klines))
	low := make([]float64, len(s.klines))
	close := make([]float64, len(s.klines))
	volume := make([]float64, len(s.klines))

	for i, k := range s.klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volume[i] = k.Volume
	}

	// CCI TV Standard (20)
	cciTV := indicators.NewCCITVStandard(20)
	s.cciValues = cciTV.Calculate(high, low, close)

	// MFI TV Standard (14)
	mfiTV := indicators.NewMFITVStandard(14)
	s.mfiValues = mfiTV.Calculate(high, low, close, volume)

	// Stochastique TV Standard (14,3,3)
	stochTV := indicators.NewStochTVStandard(14, 3, 3)
	s.stochKValues, s.stochDValues = stochTV.Calculate(high, low, close)

	// Logs indicateurs calculés (comme scalping_engine)
	fmt.Printf("✅ Indicateurs calculés: CCI=%d, MFI=%d, StochK=%d, StochD=%d\n",
		len(s.cciValues), len(s.mfiValues), len(s.stochKValues), len(s.stochDValues))

	// DEBUG: Afficher dernières valeurs pour vérification
	if len(s.klines) >= 5 && len(s.cciValues) >= len(s.klines) {
		lastIdx := len(s.klines) - 1
		fmt.Printf("\n📊 INDICATEURS CALCULÉS:\n")
		fmt.Printf("   CCI(N-1): %.1f | MFI(N-1): %.1f\n",
			s.cciValues[lastIdx], s.mfiValues[lastIdx])
		fmt.Printf("   Stoch K(N-1): %.1f D(N-1): %.1f\n\n",
			s.stochKValues[lastIdx], s.stochDValues[lastIdx])

		// Afficher 5 dernières valeurs pour debug
		fmt.Printf("[DEBUG] 5 dernières valeurs indicateurs:\n")
		for i := lastIdx - 4; i <= lastIdx; i++ {
			if i >= 0 && i < len(s.cciValues) {
				fmt.Printf("[DEBUG] [%d] CCI=%.2f, MFI=%.2f, StochK=%.2f, StochD=%.2f, Vol=%.0f\n",
					i, s.cciValues[i], s.mfiValues[i], s.stochKValues[i], s.stochDValues[i], s.klines[i].Volume)
			}
		}
		fmt.Println()
	}

	return nil
}

// DetectSignals détecte les signaux de scalping avec logs debug détaillés
func (s *ScalpingStrategy) DetectSignals() []Signal {
	var signals []Signal

	// Vérifier qu'on a assez de données
	minRequiredIndex := 20 + s.config.ValidationWindow // CCI(20) + window
	lastIdx := len(s.klines) - 1                       // Dernière kline (N-1)

	fmt.Printf("\n🔍 DÉTECTION SIGNAUX:\n")
	fmt.Printf("[DEBUG] DetectSignals: lastIdx=%d, minReq=%d, CCI_len=%d, MFI_len=%d, StochK_len=%d\n",
		lastIdx, minRequiredIndex, len(s.cciValues), len(s.mfiValues), len(s.stochKValues))

	if lastIdx < minRequiredIndex {
		fmt.Printf("[DEBUG] ⚠️  Pas assez de données: lastIdx(%d) < minReq(%d)\n\n", lastIdx, minRequiredIndex)
		return signals
	}

	n2Index := lastIdx - 1 // N-2
	n1Index := lastIdx     // N-1

	// DEBUG: Afficher valeurs indicateurs N-2 et N-1
	if n2Index >= 0 && n2Index < len(s.cciValues) && n1Index < len(s.cciValues) {
		fmt.Printf("[DEBUG] Indicateurs N-2[%d]: CCI=%.2f, MFI=%.2f, StochK=%.2f, StochD=%.2f\n",
			n2Index, s.cciValues[n2Index], s.mfiValues[n2Index], s.stochKValues[n2Index], s.stochDValues[n2Index])
		fmt.Printf("[DEBUG] Indicateurs N-1[%d]: CCI=%.2f, MFI=%.2f, StochK=%.2f, StochD=%.2f\n",
			n1Index, s.cciValues[n1Index], s.mfiValues[n1Index], s.stochKValues[n1Index], s.stochDValues[n1Index])
	}

	// 1️⃣ Vérifier triple extrême sur N-2 OU N-1
	extremeOnN2 := s.isTripleExtreme(n2Index)
	extremeOnN1 := s.isTripleExtreme(n1Index)

	fmt.Printf("[DEBUG] Triple extrême: N-2=%v, N-1=%v\n", extremeOnN2, extremeOnN1)

	if extremeOnN2 || extremeOnN1 {
		fmt.Printf("[DEBUG] 🎯 Triple extrême DÉTECTÉ!\n")

		// 2️⃣ Vérifier croisement Stochastique (comparer N-2 vs N-1)
		crossingType := s.detectStochCrossing(n1Index)
		fmt.Printf("[DEBUG] Croisement stochastique: type=%s\n", crossingType)

		if crossingType != "" {
			fmt.Printf("[DEBUG] ✅ CROISEMENT DÉTECTÉ: %s\n", crossingType)

			// 3️⃣ Validation window
			signal := s.validateInWindow(n2Index, crossingType)
			if signal != nil {
				fmt.Printf("[DEBUG] ✅ SIGNAL VALIDÉ dans window!\n")
				signals = append(signals, *signal)
			} else {
				fmt.Printf("[DEBUG] ❌ Signal NON validé (pas de bougie/volume conforme)\n")
			}
		} else {
			fmt.Printf("[DEBUG] ❌ Pas de croisement stochastique détecté\n")
		}
	} else {
		fmt.Printf("[DEBUG] ❌ Triple extrême non détecté\n")
	}

	fmt.Printf("[DEBUG] Résultat: %d signal(aux) détecté(s)\n\n", len(signals))
	return signals
}

// isTripleExtreme vérifie si CCI, MFI et Stochastic sont tous en zones extrêmes
func (s *ScalpingStrategy) isTripleExtreme(index int) bool {
	if index >= len(s.cciValues) || index >= len(s.mfiValues) ||
		index >= len(s.stochKValues) || index >= len(s.stochDValues) {
		return false
	}

	cci := s.cciValues[index]
	mfi := s.mfiValues[index]
	k := s.stochKValues[index]
	d := s.stochDValues[index]

	// SURACHAT
	isOverbought := cci > s.config.CCISurachat && mfi > s.config.MFISurachat &&
		(k >= s.config.StochSurachat || d >= s.config.StochSurachat)

	// SURVENTE
	isOversold := cci < s.config.CCISurvente && mfi < s.config.MFISurvente &&
		(k <= s.config.StochSurvente || d <= s.config.StochSurvente)

	return isOverbought || isOversold
}

// detectStochCrossing détecte le croisement stochastique
func (s *ScalpingStrategy) detectStochCrossing(index int) string {
	if index < 1 || index >= len(s.stochKValues) || index >= len(s.stochDValues) {
		return ""
	}

	n2 := index - 1 // N-2
	n1 := index     // N-1

	kN2 := s.stochKValues[n2]
	dN2 := s.stochDValues[n2]
	kN1 := s.stochKValues[n1]
	dN1 := s.stochDValues[n1]

	// Croisement haussier (SHORT): K passe SOUS D
	if kN2 > dN2 && kN1 < dN1 {
		return "SHORT"
	}

	// Croisement baissier (LONG): K passe AU-DESSUS de D
	if kN2 < dN2 && kN1 > dN1 {
		return "LONG"
	}

	return ""
}

// validateInWindow valide le signal dans la fenêtre de validation
func (s *ScalpingStrategy) validateInWindow(crossingIndex int, signalType string) *Signal {
	// Rechercher dans la fenêtre de validation
	for i := crossingIndex; i < crossingIndex+s.config.ValidationWindow && i < len(s.klines); i++ {
		bougiValide := false
		if signalType == "SHORT" {
			bougiValide = s.klines[i].Close < s.klines[i].Open // Bougie rouge
		} else {
			bougiValide = s.klines[i].Close > s.klines[i].Open // Bougie verte
		}

		if !bougiValide {
			continue
		}

		// TODO: Ajouter check volume si activé

		// Signal validé
		return &Signal{
			Type:      signalType,
			Timestamp: s.klines[i].Timestamp,
			Price:     s.klines[i].Close,
			CCI:       s.cciValues[crossingIndex],
			MFI:       s.mfiValues[crossingIndex],
			StochK:    s.stochKValues[crossingIndex],
			StochD:    s.stochDValues[crossingIndex],
			Volume:    s.klines[i].Volume,
		}
	}

	return nil
}

// parseFloat convertit une interface{} en float64
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return math.NaN()
	}
}

// timestampMsToTime convertit un timestamp en millisecondes vers time.Time
// FONCTION CENTRALE : Garantit la cohérence avec scalping_engine
func timestampMsToTime(timestampMs int64) time.Time {
	return time.Unix(timestampMs/1000, 0).UTC()
}
