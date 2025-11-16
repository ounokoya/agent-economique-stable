package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/shared"
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/direction"
)

// DirectionConfig holds hardcoded direction strategy parameters
type DirectionConfig struct {
	Timeframe           string
	VWMAPeriod          int
	SlopePeriod         int
	KConfirmation       int
	UseDynamicThreshold bool
	ATRPeriod           int
	ATRCoefficient      float64
	FixedThreshold      float64
}

// DefaultDirectionConfig returns default direction parameters
// Config optimale identifiée par analyse comparative (33 tests)
// Performance: +6.03% capté sur 2.5 jours, 12 intervalles, ~3h/trade
func DefaultDirectionConfig() DirectionConfig {
	return DirectionConfig{
		Timeframe:           "5m",
		VWMAPeriod:          20,  // Optimal pour filtrage bruit + capture tendances
		SlopePeriod:         6,   // Calcul pente stable
		KConfirmation:       2,   // Confirmation standard
		UseDynamicThreshold: true, // ATR dynamique
		ATRPeriod:           8,   // Période ATR adaptée
		ATRCoefficient:      0.25, // Sensibilité optimale
		FixedThreshold:      0.1, // Utilisé si UseDynamicThreshold = false
	}
}

// PerformanceMetrics holds detailed performance timers
type PerformanceMetrics struct {
	TimeLoadData    time.Duration
	TimeIndicators  time.Duration
	TimeSignals     time.Duration
	TimeTotal       time.Duration
	
	CountMarkers    int
	CountSignals    int
}

// DirectionEngineApp manages the direction strategy with temporal engine
type DirectionEngineApp struct {
	config         *shared.Config
	dates          []string
	directionCfg   DirectionConfig

	generator *direction.DirectionGenerator

	klines []Kline
	
	tradesHistory []shared.TradeData
	
	currentKline CurrentKlineByTrade

	signals []DirectionSignal
	
	currentPosition *Position
	closedPositions []Position
	
	startTime     time.Time
	executionTime time.Duration
	metrics       PerformanceMetrics
}

// Kline represents a candle
type Kline struct {
	Timestamp        int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	QuoteAssetVolume float64
}

// CurrentKlineByTrade represents a kline constructed in real-time from trades
type CurrentKlineByTrade struct {
	StartTimestamp   int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	QuoteAssetVolume float64
	TradesCount      int
	Initialized      bool
}

// DirectionSignal represents a direction signal with context
type DirectionSignal struct {
	Timestamp  time.Time
	Type       signals.SignalType
	Action     signals.SignalAction
	Price      float64
	Confidence float64
	
	VWMA6           float64
	ATR             float64
	SlopeVariation  float64
	Threshold       float64
	
	PositionID      int
}

// Position represents an open or closed position
type Position struct {
	ID             int
	Type           signals.SignalType
	EntryTime      time.Time
	EntryPrice     float64
	ExitTime       *time.Time
	ExitPrice      *float64
	Duration       time.Duration
	PnLPercent     float64
	
	// Stats (calculés uniquement à la fermeture)
	MaxDrawdown    float64
	MaxRunup       float64
}

// LoadDirectionConfigFromYAML loads direction config from YAML or returns defaults
func LoadDirectionConfigFromYAML(config *shared.Config) DirectionConfig {
	dirCfg := config.Strategy.DirectionConfig
	
	// Si config vide ou incomplète, utiliser valeurs par défaut
	defaultCfg := DefaultDirectionConfig()
	
	// Vérifier si la config est définie (au moins VWMA period doit être > 0)
	if dirCfg.VWMAPeriod == 0 {
		fmt.Println("⚠️  Config YAML direction vide, utilisation valeurs optimales par défaut")
		return defaultCfg
	}
	
	// Construire config depuis YAML
	return DirectionConfig{
		Timeframe:           dirCfg.Timeframe,
		VWMAPeriod:          dirCfg.VWMAPeriod,
		SlopePeriod:         dirCfg.SlopePeriod,
		KConfirmation:       dirCfg.KConfirmation,
		UseDynamicThreshold: dirCfg.UseDynamicThreshold,
		ATRPeriod:           dirCfg.ATRPeriod,
		ATRCoefficient:      dirCfg.ATRCoefficient,
		FixedThreshold:      dirCfg.FixedThreshold,
	}
}

// NewDirectionEngineApp creates a new direction engine application
func NewDirectionEngineApp(config *shared.Config, dates []string) *DirectionEngineApp {
	tradesHistorySize := 300
	if config.Backtest.TradesHistorySize > 0 {
		tradesHistorySize = config.Backtest.TradesHistorySize
	}
	
	// Charger config direction depuis YAML (ou valeurs par défaut)
	directionCfg := LoadDirectionConfigFromYAML(config)
	
	return &DirectionEngineApp{
		config:          config,
		dates:           dates,
		directionCfg:    directionCfg,
		signals:         make([]DirectionSignal, 0),
		tradesHistory:   make([]shared.TradeData, 0, tradesHistorySize),
		closedPositions: make([]Position, 0),
	}
}

// Run executes the direction engine backtest
func (app *DirectionEngineApp) Run() error {
	app.startTime = time.Now()
	
	fmt.Println("\n📂 Chargement klines...")
	startLoadData := time.Now()
	if err := app.loadKlines(); err != nil {
		return fmt.Errorf("erreur chargement klines: %w", err)
	}
	app.metrics.TimeLoadData = time.Since(startLoadData)
	fmt.Printf("✅ %d klines chargées\n", len(app.klines))

	fmt.Println("\n⚙️  Initialisation générateur direction...")
	if err := app.initializeGenerator(); err != nil {
		return fmt.Errorf("erreur initialisation générateur: %w", err)
	}

	fmt.Println("\n🔄 Traitement trades en streaming...")
	if err := app.processTradesStreaming(); err != nil {
		return fmt.Errorf("erreur traitement trades: %w", err)
	}

	app.executionTime = time.Since(app.startTime)
	app.metrics.TimeTotal = app.executionTime
	
	app.displayResults()
	
	if app.config.Backtest.ExportJSON {
		fmt.Println("\n💾 Export JSON...")
		if err := app.exportSignalsToJSON(); err != nil {
			fmt.Printf("⚠️  Erreur export JSON: %v\n", err)
		}
	}

	return nil
}

// loadKlines charge les klines depuis les fichiers ZIP Binance Vision
func (app *DirectionEngineApp) loadKlines() error {
	cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
	if err != nil {
		return err
	}

	streamConfig := shared.StreamingConfig{
		BufferSize:    app.config.BinanceData.Streaming.BufferSize,
		MaxMemoryMB:   app.config.BinanceData.Streaming.MaxMemoryMB,
		EnableMetrics: app.config.BinanceData.Streaming.EnableMetrics,
	}
	reader, err := binance.NewStreamingReader(cache, streamConfig)
	if err != nil {
		return err
	}

	aggConfig := shared.AggregationConfig{}
	processor, err := binance.NewParsedDataProcessor(cache, reader, aggConfig)
	if err != nil {
		return err
	}

	symbol := app.config.BinanceData.Symbols[0]
	timeframe := app.directionCfg.Timeframe
	app.klines = make([]Kline, 0, len(app.dates)*288)

	for _, date := range app.dates {
		klinesFile := cache.GetFilePath(symbol, "klines", date, timeframe)
		batch, err := processor.ParseKlinesBatch(klinesFile, symbol, timeframe, date)
		if err != nil {
			fmt.Printf("  ⚠️  Skip date %s: %v\n", date, err)
			continue
		}

		for _, klineData := range batch.KlinesData {
			kline := Kline{
				Timestamp:        klineData.OpenTime,
				Open:             klineData.Open,
				High:             klineData.High,
				Low:              klineData.Low,
				Close:            klineData.Close,
				Volume:           klineData.Volume,
				QuoteAssetVolume: klineData.QuoteAssetVolume,
			}
			app.klines = append(app.klines, kline)
		}
	}

	return nil
}

// initializeGenerator initialise le générateur de direction
func (app *DirectionEngineApp) initializeGenerator() error {
	genConfig := direction.Config{
		VWMAPeriod:          app.directionCfg.VWMAPeriod,
		SlopePeriod:         app.directionCfg.SlopePeriod,
		KConfirmation:       app.directionCfg.KConfirmation,
		UseDynamicThreshold: app.directionCfg.UseDynamicThreshold,
		ATRPeriod:           app.directionCfg.ATRPeriod,
		ATRCoefficient:      app.directionCfg.ATRCoefficient,
		FixedThreshold:      app.directionCfg.FixedThreshold,
	}

	app.generator = direction.NewDirectionGenerator(genConfig)

	symbol := app.config.BinanceData.Symbols[0]
	return app.generator.Initialize(signals.GeneratorConfig{
		Symbol:    symbol,
		Timeframe: app.directionCfg.Timeframe,
	})
}

// Helper function for string repetition
func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// processTradesStreaming traite les trades en streaming jour par jour
func (app *DirectionEngineApp) processTradesStreaming() error {
	cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
	if err != nil {
		return fmt.Errorf("erreur init cache: %w", err)
	}

	streamConfig := shared.StreamingConfig{
		BufferSize:    app.config.BinanceData.Streaming.BufferSize,
		MaxMemoryMB:   app.config.BinanceData.Streaming.MaxMemoryMB,
		EnableMetrics: app.config.BinanceData.Streaming.EnableMetrics,
	}
	reader, err := binance.NewStreamingReader(cache, streamConfig)
	if err != nil {
		return fmt.Errorf("erreur création reader: %w", err)
	}

	symbol := app.config.BinanceData.Symbols[0]
	var nextMarker int64
	tradesProcessed := 0
	markersDetected := 0
	markerJustCrossed := false

	// Traiter chaque date
	for dateIdx, date := range app.dates {
		if app.config.Backtest.Logging.EnableProgressLogs {
			fmt.Printf("\n📅 Date %d/%d: %s\n", dateIdx+1, len(app.dates), date)
		}

		tradesFile := cache.GetFilePath(symbol, "trades", date)
		dayTrades := 0
		
		// Streamer les trades de cette date
		err = reader.StreamTrades(tradesFile, func(trade shared.TradeData) error {
			tradesProcessed++
			dayTrades++

			// Maintenir buffer circulaire
			tradesHistorySize := 300
			if app.config.Backtest.TradesHistorySize > 0 {
				tradesHistorySize = app.config.Backtest.TradesHistorySize
			}
			if len(app.tradesHistory) >= tradesHistorySize {
				app.tradesHistory = app.tradesHistory[1:]
			}
			app.tradesHistory = append(app.tradesHistory, trade)
			
			// Mettre à jour kline courante
			app.updateCurrentKline(trade)

			// Initialiser nextMarker au premier trade
			if nextMarker == 0 {
				nextMarker = app.calculateNextMarker(trade.Time)
			}

			// Détecter franchissement de marqueur
			if trade.Time >= nextMarker && !markerJustCrossed {
				markersDetected++
				markerJustCrossed = true
				app.metrics.CountMarkers++

				// Logger marqueur si activé
				if app.config.Backtest.Logging.EnableMarkerLogs {
					markerTime := timestampMsToTime(nextMarker)
					fmt.Printf("\n🕐 %s | MARQUEUR DÉTECTÉ\n", markerTime.Format("15:04:05"))
				}

				// Traiter marqueur (calculs indicateurs + détection signaux)
				app.processMarker(nextMarker)

				// Passer au marqueur suivant
				nextMarker = app.calculateNextMarker(trade.Time)
				markerJustCrossed = false
			}

			// NOTE: Direction n'a PAS de trailing stop - gestion uniquement aux marqueurs

			return nil
		})

		if err != nil {
			fmt.Printf("⚠️  Erreur streaming trades date %s: %v\n", date, err)
			continue
		}

		if app.config.Backtest.Logging.EnableProgressLogs {
			fmt.Printf("  ✅ %d trades traités\n", dayTrades)
		}
	}

	fmt.Printf("\n✅ Traitement terminé:\n")
	fmt.Printf("   • Trades: %d\n", tradesProcessed)
	fmt.Printf("   • Marqueurs: %d\n", markersDetected)
	fmt.Printf("   • Signaux: %d\n", len(app.signals))
	fmt.Printf("   • Positions fermées: %d\n", len(app.closedPositions))

	return nil
}

func (app *DirectionEngineApp) displayResults() {
	fmt.Println("\n" + repeatStr("═", 100))
	fmt.Println("  RÉSULTATS BACKTEST DIRECTION")
	fmt.Println(repeatStr("═", 100))
	
	// Statistiques signaux
	totalEntry := 0
	totalExit := 0
	totalLong := 0
	totalShort := 0
	
	for _, sig := range app.signals {
		if sig.Action == signals.SignalActionEntry {
			totalEntry++
		} else {
			totalExit++
		}
		if sig.Type == signals.SignalTypeLong {
			totalLong++
		} else {
			totalShort++
		}
	}
	
	fmt.Println("\n📊 SIGNAUX:")
	fmt.Printf("   • Total: %d\n", len(app.signals))
	fmt.Printf("   • ENTRY: %d\n", totalEntry)
	fmt.Printf("   • EXIT: %d\n", totalExit)
	fmt.Printf("   • LONG: %d\n", totalLong)
	fmt.Printf("   • SHORT: %d\n", totalShort)
	
	// Statistiques positions
	fmt.Println("\n💼 POSITIONS:")
	fmt.Printf("   • Fermées: %d\n", len(app.closedPositions))
	
	if len(app.closedPositions) > 0 {
		// Compter par type et calculer variations
		countLong := 0
		countShort := 0
		variationLong := 0.0
		variationShort := 0.0
		gagnantes := 0
		
		for _, pos := range app.closedPositions {
			if pos.Type == signals.SignalTypeLong {
				countLong++
				variationLong += pos.PnLPercent
			} else {
				countShort++
				variationShort += pos.PnLPercent
			}
			
			if pos.PnLPercent > 0 {
				gagnantes++
			}
		}
		
		winRate := float64(gagnantes) / float64(len(app.closedPositions)) * 100
		fmt.Printf("   • Gagnantes: %d (%.1f%%)\n", gagnantes, winRate)
		fmt.Printf("   • Perdantes: %d\n", len(app.closedPositions)-gagnantes)
		
		// Variations captées
		fmt.Println("\n💰 VARIATIONS CAPTÉES:")
		if countLong > 0 {
			fmt.Printf("   • LONG (↗)  : %+.2f%% total, %+.2f%% moyen\n",
				variationLong, variationLong/float64(countLong))
		}
		if countShort > 0 {
			fmt.Printf("   • SHORT (↘) : %+.2f%% total, %+.2f%% moyen\n",
				variationShort, variationShort/float64(countShort))
		}
		
		// Total capté = LONG + (SHORT × -1)
		// Les variations SHORT profitables sont négatives, donc on inverse
		totalCapte := variationLong - variationShort
		fmt.Printf("   • TOTAL CAPTÉ: %.2f%% (bidirectionnel)\n", totalCapte)
		
		// Meilleure et pire
		maxWin := 0.0
		maxLoss := 0.0
		for _, pos := range app.closedPositions {
			if pos.PnLPercent > maxWin {
				maxWin = pos.PnLPercent
			}
			if pos.PnLPercent < maxLoss {
				maxLoss = pos.PnLPercent
			}
		}
		
		fmt.Println("\n📈 PERFORMANCE:")
		fmt.Printf("   • Max Win: %+.2f%%\n", maxWin)
		fmt.Printf("   • Max Loss: %+.2f%%\n", maxLoss)
	}
	
	fmt.Println("\n" + repeatStr("═", 100))
}

func (app *DirectionEngineApp) exportSignalsToJSON() error {
	exportPath := app.config.Backtest.ExportPath
	if exportPath == "" {
		exportPath = "backtest_results"
	}

	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("direction_signals_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(exportPath, filename)

	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"signals":   app.signals,
		"positions": app.closedPositions,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// processMarker traite un marqueur (calcul indicateurs + détection signaux)
func (app *DirectionEngineApp) processMarker(markerTimestamp int64) {
	// Récupérer window de klines jusqu'au marqueur (300 dernières)
	klineIdx := app.findKlineIndexAtTimestamp(markerTimestamp)
	if klineIdx < 0 {
		return
	}

	// Construire window de signals.Kline pour le générateur
	windowSize := 300
	startIdx := klineIdx - windowSize + 1
	if startIdx < 0 {
		startIdx = 0
	}

	signalKlines := make([]signals.Kline, 0, klineIdx-startIdx+1)
	for i := startIdx; i <= klineIdx; i++ {
		k := app.klines[i]
		signalKlines = append(signalKlines, signals.Kline{
			OpenTime: timestampMsToTime(k.Timestamp),
			Open:     k.Open,
			High:     k.High,
			Low:      k.Low,
			Close:    k.Close,
			Volume:   k.Volume,
		})
	}

	// Calculer indicateurs
	startIndicators := time.Now()
	if err := app.generator.CalculateIndicators(signalKlines); err != nil {
		if app.config.Backtest.Logging.EnableMarkerLogs {
			fmt.Printf("⚠️  Erreur calcul indicateurs: %v\n", err)
		}
		return
	}
	app.metrics.TimeIndicators += time.Since(startIndicators)

	// Détecter signaux
	startSignals := time.Now()
	newSignals, err := app.generator.DetectSignals(signalKlines)
	if err != nil {
		if app.config.Backtest.Logging.EnableMarkerLogs {
			fmt.Printf("⚠️  Erreur détection signaux: %v\n", err)
		}
		return
	}
	app.metrics.TimeSignals += time.Since(startSignals)
	app.metrics.CountSignals += len(newSignals)

	// Traiter les signaux détectés
	for _, sig := range newSignals {
		app.handleSignal(sig)
	}

	// Logger indicateurs si activé
	if app.config.Backtest.Logging.EnableIndicatorLogs && len(newSignals) > 0 {
		fmt.Printf("   📊 %d nouveaux signaux détectés\n", len(newSignals))
	}
}

// handleSignal traite un signal ENTRY ou EXIT
func (app *DirectionEngineApp) handleSignal(sig signals.Signal) {
	// Créer DirectionSignal avec contexte
	dirSig := DirectionSignal{
		Timestamp:  sig.Timestamp,
		Type:       sig.Type,
		Action:     sig.Action,
		Price:      sig.Price,
		Confidence: sig.Confidence,
	}

	// Extraire metadata
	if meta, ok := sig.Metadata["vwma6"].(float64); ok {
		dirSig.VWMA6 = meta
	}

	app.signals = append(app.signals, dirSig)

	// Gérer position
	if sig.Action == signals.SignalActionEntry {
		app.openPosition(sig)
	} else if sig.Action == signals.SignalActionExit {
		app.closePosition(sig)
	}

	// Logger signal si activé
	if app.config.Backtest.Logging.EnableSignalLogs {
		typeStr := "LONG"
		if sig.Type == signals.SignalTypeShort {
			typeStr = "SHORT"
		}
		fmt.Printf("   🎯 %s %s @ %.2f (conf: %.2f)\n",
			sig.Action, typeStr, sig.Price, sig.Confidence)
	}
}

// openPosition ouvre une nouvelle position
func (app *DirectionEngineApp) openPosition(sig signals.Signal) {
	// Fermer position actuelle si existe
	if app.currentPosition != nil {
		app.forceClosePosition(sig.Timestamp, sig.Price)
	}

	// Créer nouvelle position
	position := &Position{
		ID:         len(app.closedPositions) + 1,
		Type:       sig.Type,
		EntryTime:  sig.Timestamp,
		EntryPrice: sig.Price,
	}

	app.currentPosition = position
}

// closePosition ferme la position actuelle avec signal EXIT
func (app *DirectionEngineApp) closePosition(sig signals.Signal) {
	if app.currentPosition == nil {
		return
	}

	exitTime := sig.Timestamp
	exitPrice := sig.Price
	
	app.currentPosition.ExitTime = &exitTime
	app.currentPosition.ExitPrice = &exitPrice
	app.currentPosition.Duration = exitTime.Sub(app.currentPosition.EntryTime)

	// Calculer P&L
	if app.currentPosition.Type == signals.SignalTypeLong {
		app.currentPosition.PnLPercent = (exitPrice - app.currentPosition.EntryPrice) / app.currentPosition.EntryPrice * 100
	} else {
		app.currentPosition.PnLPercent = (app.currentPosition.EntryPrice - exitPrice) / app.currentPosition.EntryPrice * 100
	}

	app.closedPositions = append(app.closedPositions, *app.currentPosition)
	app.currentPosition = nil
}

// forceClosePosition ferme position actuelle (sans signal EXIT)
func (app *DirectionEngineApp) forceClosePosition(exitTime time.Time, exitPrice float64) {
	if app.currentPosition == nil {
		return
	}

	app.currentPosition.ExitTime = &exitTime
	app.currentPosition.ExitPrice = &exitPrice
	app.currentPosition.Duration = exitTime.Sub(app.currentPosition.EntryTime)

	if app.currentPosition.Type == signals.SignalTypeLong {
		app.currentPosition.PnLPercent = (exitPrice - app.currentPosition.EntryPrice) / app.currentPosition.EntryPrice * 100
	} else {
		app.currentPosition.PnLPercent = (app.currentPosition.EntryPrice - exitPrice) / app.currentPosition.EntryPrice * 100
	}

	app.closedPositions = append(app.closedPositions, *app.currentPosition)
	app.currentPosition = nil
}

// updateCurrentKline met à jour la kline courante avec un trade
func (app *DirectionEngineApp) updateCurrentKline(trade shared.TradeData) {
	if !app.currentKline.Initialized {
		app.currentKline.StartTimestamp = trade.Time
		app.currentKline.Open = trade.Price
		app.currentKline.High = trade.Price
		app.currentKline.Low = trade.Price
		app.currentKline.Close = trade.Price
		app.currentKline.Volume = trade.Quantity
		app.currentKline.QuoteAssetVolume = trade.Price * trade.Quantity
		app.currentKline.TradesCount = 1
		app.currentKline.Initialized = true
		return
	}

	// MAJ High/Low
	if trade.Price > app.currentKline.High {
		app.currentKline.High = trade.Price
	}
	if trade.Price < app.currentKline.Low {
		app.currentKline.Low = trade.Price
	}

	app.currentKline.Close = trade.Price
	app.currentKline.Volume += trade.Quantity
	app.currentKline.QuoteAssetVolume += trade.Price * trade.Quantity
	app.currentKline.TradesCount++
}

// calculateNextMarker calcule le prochain marqueur 5m
func (app *DirectionEngineApp) calculateNextMarker(currentTimestamp int64) int64 {
	t := timestampMsToTime(currentTimestamp)
	
	// Aligner sur prochaine 5 minutes
	minute := t.Minute()
	nextMinute := ((minute / 5) + 1) * 5
	
	if nextMinute >= 60 {
		t = t.Add(time.Hour)
		nextMinute = 0
	}
	
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), nextMinute, 0, 0, t.Location())
	
	return t.UnixNano() / 1e6
}

// findKlineIndexAtTimestamp trouve l'index de la kline à un timestamp
func (app *DirectionEngineApp) findKlineIndexAtTimestamp(timestamp int64) int {
	for i, k := range app.klines {
		if k.Timestamp == timestamp {
			return i
		}
	}
	return -1
}

// timestampMsToTime convertit un timestamp ms en time.Time
func timestampMsToTime(ts int64) time.Time {
	return time.Unix(0, ts*1e6)
}
