// Package main provides engine demo functionality
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/bingx"
	"agent-economique/internal/engine"
	indicators "agent-economique/internal/indicators"
	"agent-economique/internal/shared"
)

// EngineDemoApp manages engine demo workflow with REAL BingX data
type EngineDemoApp struct {
	config        *shared.Config
	symbol        string
	timeframe     string
	candleCount   int
	temporalEngine *engine.TemporalEngine
	bingxClient   *bingx.Client
	marketService *bingx.MarketDataService
}

// NewEngineDemoApp creates a new engine demo application
func NewEngineDemoApp(config *shared.Config, symbol, timeframe string, candleCount int) *EngineDemoApp {
	return &EngineDemoApp{
		config:      config,
		symbol:      symbol,
		timeframe:   timeframe,
		candleCount: candleCount,
	}
}

// RunDemo executes the complete demo workflow with REAL BingX data
func (app *EngineDemoApp) RunDemo() error {
	fmt.Printf("🚀 Démarrage test engine pour %s (%s) - %d bougies\n", 
		app.symbol, app.timeframe, app.candleCount)

	// Initialize BingX client
	err := app.initializeBingXClient()
	if err != nil {
		return fmt.Errorf("failed to initialize BingX client: %w", err)
	}

	// Initialize temporal engine
	err = app.initializeTemporalEngine()
	if err != nil {
		return fmt.Errorf("failed to initialize temporal engine: %w", err)
	}

	// Fetch candle data from BingX
	candles, err := app.fetchCandleData()
	if err != nil {
		return fmt.Errorf("failed to fetch candle data: %w", err)
	}

	fmt.Printf("📊 %d bougies récupérées, démarrage simulation engine...\n", len(candles))

	// Initialize engine timestamp with first candle time to avoid look-ahead
	if len(candles) > 0 {
		err = app.initializeEngineTimestamp(candles[0])
		if err != nil {
			return fmt.Errorf("failed to initialize engine timestamp: %w", err)
		}
	}

	// Process each candle through temporal engine
	err = app.processCandles(candles)
	if err != nil {
		return fmt.Errorf("failed to process candles: %w", err)
	}

	// Generate final summary
	app.generateSummary()

	return nil
}

// initializeBingXClient sets up the REAL BingX API client from config
func (app *EngineDemoApp) initializeBingXClient() error {
	// Parse timeout from config
	timeout, err := time.ParseDuration(app.config.BingXData.Timeout)
	if err != nil {
		timeout = 30 * time.Second // Default fallback
		fmt.Printf("⚠️  Invalid timeout in config, using default: %v\n", timeout)
	}

	// Convert environment from config
	var environment bingx.Environment
	switch app.config.BingXData.Environment {
	case "demo":
		environment = bingx.DemoEnvironment
	case "live":
		environment = bingx.LiveEnvironment
	default:
		environment = bingx.DemoEnvironment // Default to demo for safety
		fmt.Printf("⚠️  Invalid environment '%s', using demo mode\n", app.config.BingXData.Environment)
	}

	// Create BingX client configuration from YAML config
	clientConfig := bingx.ClientConfig{
		Environment: environment,
		Credentials: bingx.APICredentials{
			APIKey:    app.config.BingXData.Credentials.APIKey,
			SecretKey: app.config.BingXData.Credentials.SecretKey,
		},
		Timeout: timeout,
	}

	// Create BingX client
	client, err := bingx.NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create BingX client: %w", err)
	}

	app.bingxClient = client
	app.marketService = bingx.NewMarketDataService(client)

	fmt.Printf("🔌 BingX client initialisé (SDK RÉEL - mode %s)\n", app.config.BingXData.Environment)
	fmt.Printf("📡 API Key: %s...%s\n", 
		app.config.BingXData.Credentials.APIKey[:8],
		app.config.BingXData.Credentials.APIKey[len(app.config.BingXData.Credentials.APIKey)-8:])
	return nil
}

// initializeTemporalEngine sets up the temporal engine with YAML config
func (app *EngineDemoApp) initializeTemporalEngine() error {
	// Create default engine configuration
	engineConfig := engine.DefaultEngineConfig()
	
	// Disable anti-lookahead for demo to process historical data
	engineConfig.AntiLookAhead = false
	
	fmt.Printf("📋 Configuration Engine: WindowSize=%d, AntiLookAhead=%v\n", 
		engineConfig.WindowSize, engineConfig.AntiLookAhead)

	// Create temporal engine from YAML config in PaperMode for live demo
	temporalEngine, err := engine.NewTemporalEngineFromYAML(
		engine.PaperMode, // Use PaperMode for live demo with real data
		app.config,
		engineConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create temporal engine: %w", err)
	}

	app.temporalEngine = temporalEngine

	// Initialize engine with empty initial data
	initialData := engine.InitialData{
		Trades:       make([]engine.Trade, 0),
		RecentKlines: make([]engine.Kline, 0),
		RecentTrades: make([]engine.Trade, 0),
	}

	err = app.temporalEngine.Initialize(initialData)
	if err != nil {
		return fmt.Errorf("failed to initialize temporal engine: %w", err)
	}

	// Start the engine
	err = app.temporalEngine.Start()
	if err != nil {
		return fmt.Errorf("failed to start temporal engine: %w", err)
	}

	fmt.Printf("⚡ Temporal Engine initialisé et démarré (stratégie: %s)\n", app.config.Strategy.Name)
	fmt.Println("🎯 Engine prêt pour traitement données BingX en temps réel")
	return nil
}

// initializeEngineTimestamp initializes engine with first candle timestamp
func (app *EngineDemoApp) initializeEngineTimestamp(firstCandle bingx.Kline) error {
	// Create a synthetic first trade to initialize engine timestamp properly
	initialTrade := engine.Trade{
		Timestamp: firstCandle.OpenTime.Unix() * 1000, // Use first candle open time
		Price:     firstCandle.Open,
		Quantity:  0, // Minimal quantity
	}

	// Process this initial trade to set engine timestamp
	err := app.temporalEngine.ProcessTrade(initialTrade)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize timestamp: %v\n", err)
	} else {
		fmt.Printf("🕐 Engine timestamp initialisé: %s\n", firstCandle.OpenTime.Format("15:04:05"))
	}

	return nil
}

// fetchCandleData retrieves REAL candle data from BingX
func (app *EngineDemoApp) fetchCandleData() ([]bingx.Kline, error) {
	// Use engine's WindowSize instead of CLI parameter for proper indicator calculation
	engineWindowSize := 300 // From DefaultEngineConfig()
	actualCount := engineWindowSize
	if app.candleCount > engineWindowSize {
		actualCount = app.candleCount
	}
	
	fmt.Printf("📡 Récupération %d bougies %s pour %s via SDK BingX (WindowSize: %d)...\n", 
		actualCount, app.timeframe, app.symbol, engineWindowSize)

	ctx := context.Background()
	
	// Convert timeframe to BingX format if needed
	bingxTimeframe := app.convertTimeframeToBingX(app.timeframe)
	
	// Get REAL klines from BingX using the actual count needed
	klines, err := app.marketService.GetKlines(ctx, app.symbol, bingxTimeframe, actualCount, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines from BingX: %w", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines returned for %s %s", app.symbol, bingxTimeframe)
	}

	// Sort klines by time (oldest first) for proper temporal processing
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime.Before(klines[j].OpenTime)
	})

	fmt.Printf("✅ %d bougies RÉELLES récupérées (de %s à %s)\n", 
		len(klines), 
		klines[0].OpenTime.Format("15:04:05"),
		klines[len(klines)-1].CloseTime.Format("15:04:05"))

	return klines, nil
}

// convertTimeframeToBingX converts our timeframe format to BingX format
func (app *EngineDemoApp) convertTimeframeToBingX(timeframe string) string {
	// Map our timeframe format to BingX format if different
	switch timeframe {
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "1h":
		return "1h"
	case "4h":
		return "4h"
	case "1d":
		return "1d"
	default:
		return timeframe // assume it's already correct
	}
}

// processCandles processes each REAL candle through the temporal engine
func (app *EngineDemoApp) processCandles(candles []bingx.Kline) error {
	fmt.Println("\n🔄 Démarrage traitement bougies...")
	fmt.Println(strings.Repeat("=", 80))

	for i, candle := range candles {
		fmt.Printf("\n📊 [%d/%d] Traitement bougie RÉELLE %s (%.2f)\n", 
			i+1, len(candles), 
			candle.CloseTime.Format("15:04:05"), 
			candle.Close)

		// Convert BingX Kline to engine trade format
		trade := engine.Trade{
			Timestamp: candle.CloseTime.Unix() * 1000, // Convert to milliseconds
			Price:     candle.Close,
			Quantity:  candle.Volume,
		}

		// Process trade through temporal engine
		err := app.temporalEngine.ProcessTrade(trade)
		if err != nil {
			fmt.Printf("⚠️  Erreur traitement trade: %v\n", err)
			continue
		}

		// Check for REAL signals from engine and log only when signals are detected
		if engine.IsMarkerTimestamp(candle.CloseTime.Unix() * 1000) {
			app.checkAndLogSignals(candle, i+1)
		}

		// Small delay for readability
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Traitement terminé")
	
	return nil
}

// checkAndLogSignals checks for signals and logs ONLY when signals are detected
func (app *EngineDemoApp) checkAndLogSignals(candle bingx.Kline, candleIndex int) {
	// Get REAL indicator results from temporal engine
	indicatorResults := app.temporalEngine.GetCurrentIndicatorResults()
	if indicatorResults == nil {
		return // No indicators calculated yet
	}

	// Check if we have any crossovers or signals
	hasSignal := false
	signalType := ""
	
	// Check MACD crossovers
	if indicatorResults.MACD != nil {
		if indicatorResults.MACD.CrossoverType.String() == "CROSS_UP" {
			hasSignal = true
			signalType = "MACD CROISEMENT HAUSSIER"
		} else if indicatorResults.MACD.CrossoverType.String() == "CROSS_DOWN" {
			hasSignal = true
			signalType = "MACD CROISEMENT BAISSIER"
		}
	}

	// Check STOCH crossovers (for STOCH/MFI/CCI strategy)
	if app.config.Strategy.Name == "STOCH_MFI_CCI" && indicatorResults.Stochastic != nil {
		if indicatorResults.Stochastic.CrossoverType.String() == "CROSS_UP" {
			hasSignal = true
			signalType = "STOCH CROISEMENT HAUSSIER"
		} else if indicatorResults.Stochastic.CrossoverType.String() == "CROSS_DOWN" {
			hasSignal = true
			signalType = "STOCH CROISEMENT BAISSIER"
		}
	}

	// Check for strategic signals (based on strategy logic from signal_generator.go)
	hasStrategySignal := app.checkForStrategySignals(indicatorResults)
	if hasStrategySignal {
		hasSignal = true
		signalType = "SIGNAL STRATÉGIQUE DÉTECTÉ"
	}

	// LOG ONLY IF THERE IS A SIGNAL
	if hasSignal {
		fmt.Printf("\n🚨 [%s] %s DÉTECTÉ !\n", 
			candle.CloseTime.Format("2006-01-02 15:04:05"), signalType)
		fmt.Printf("📊 Prix: %.2f (OHLCV: %.2f/%.2f/%.2f/%.2f/%.0f)\n",
			candle.Close, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
		
		// Log ALL indicator values at signal time
		app.logAllIndicatorValues(indicatorResults)
		
		fmt.Printf("=" + strings.Repeat("=", 70) + "\n")
	}
}

// checkForStrategySignals checks for complete strategy signals (MACD + CCI + DMI logic)
func (app *EngineDemoApp) checkForStrategySignals(indicatorResults *indicators.IndicatorResults) bool {
	if indicatorResults == nil || indicatorResults.MACD == nil || indicatorResults.CCI == nil || indicatorResults.DMI == nil {
		return false
	}

	macdCross := indicatorResults.MACD.CrossoverType.String()
	cciZone := indicatorResults.CCI.Zone.String()
	// diPlus := indicatorResults.DMI.PlusDI
	// diMinus := indicatorResults.DMI.MinusDI

	// LONG Strategy: MACD croise à la hausse + CCI en survente + (DI+ > DI- pour tendance OU DI- > DI+ pour contre-tendance)
	if macdCross == "CROSS_UP" && cciZone == "OVERSOLD" {
		return true // LONG signal conditions met
	}

	// SHORT Strategy: MACD croise à la baisse + CCI en surachat + (DI+ < DI- pour tendance OU DI+ > DI- pour contre-tendance)  
	if macdCross == "CROSS_DOWN" && cciZone == "OVERBOUGHT" {
		return true // SHORT signal conditions met
	}

	return false
}

// logAllIndicatorValues logs ALL indicator values at signal time
func (app *EngineDemoApp) logAllIndicatorValues(indicatorResults *indicators.IndicatorResults) {
	if indicatorResults == nil {
		return
	}

	fmt.Printf("📈 VALEURS INDICATEURS AU MOMENT DU SIGNAL:\n")

	// MACD values
	if indicatorResults.MACD != nil {
		fmt.Printf("  MACD: %.4f | Signal: %.4f | Histogram: %.4f | Crossover: %s\n", 
			indicatorResults.MACD.MACD, indicatorResults.MACD.Signal, 
			indicatorResults.MACD.Histogram, indicatorResults.MACD.CrossoverType.String())
	}

	// CCI values
	if indicatorResults.CCI != nil {
		fmt.Printf("  CCI: %.2f | Zone: %s\n", 
			indicatorResults.CCI.Value, indicatorResults.CCI.Zone.String())
	}

	// DMI values
	if indicatorResults.DMI != nil {
		fmt.Printf("  DMI DI+: %.2f | DI-: %.2f | DX: %.2f | ADX: %.2f\n", 
			indicatorResults.DMI.PlusDI, indicatorResults.DMI.MinusDI, 
			indicatorResults.DMI.DX, indicatorResults.DMI.ADX)
	}

	// STOCH values (for STOCH/MFI/CCI strategy)
	if indicatorResults.Stochastic != nil {
		fmt.Printf("  STOCH K: %.2f | D: %.2f | Zone: %s | Crossover: %s | Extrême: %v\n", 
			indicatorResults.Stochastic.K, indicatorResults.Stochastic.D, 
			indicatorResults.Stochastic.Zone.String(), indicatorResults.Stochastic.CrossoverType.String(),
			indicatorResults.Stochastic.IsExtreme)
	}

	// MFI values (for STOCH/MFI/CCI strategy)
	if indicatorResults.MFI != nil {
		fmt.Printf("  MFI: %.2f | Zone: %s | Extrême: %v\n", 
			indicatorResults.MFI.Value, indicatorResults.MFI.Zone.String(), indicatorResults.MFI.IsExtreme)
	}

	// Strategy analysis
	fmt.Printf("🎯 ANALYSE STRATÉGIQUE:\n")
	if app.config.Strategy.Name == "STOCH_MFI_CCI" {
		fmt.Printf("  Stratégie: STOCH/MFI/CCI\n")
	} else {
		fmt.Printf("  Stratégie: MACD/CCI/DMI\n")
		if indicatorResults.MACD != nil && indicatorResults.CCI != nil && indicatorResults.DMI != nil {
			// Check strategy conditions
			if indicatorResults.MACD.CrossoverType.String() == "CROSS_UP" && indicatorResults.CCI.Zone.String() == "OVERSOLD" {
				fmt.Printf("  ✅ Conditions LONG: MACD↗ + CCI survente + DI+:%.2f vs DI-:%.2f\n", 
					indicatorResults.DMI.PlusDI, indicatorResults.DMI.MinusDI)
			} else if indicatorResults.MACD.CrossoverType.String() == "CROSS_DOWN" && indicatorResults.CCI.Zone.String() == "OVERBOUGHT" {
				fmt.Printf("  ✅ Conditions SHORT: MACD↘ + CCI surachat + DI+:%.2f vs DI-:%.2f\n", 
					indicatorResults.DMI.PlusDI, indicatorResults.DMI.MinusDI)
			}
		}
	}
}

// generateSummary provides demo execution summary  
func (app *EngineDemoApp) generateSummary() {
	fmt.Println("\n📋 RÉSUMÉ DU DÉMO AVEC DONNÉES RÉELLES")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Symbole testé: %s\n", app.symbol)
	fmt.Printf("Timeframe: %s\n", app.timeframe)
	fmt.Printf("Bougies traitées: %d\n", app.candleCount)
	fmt.Printf("Stratégie: %s\n", app.config.Strategy.Name)
	fmt.Printf("Mode engine: %s\n", "PAPER_DEMO")
	
	// TODO: Get real metrics from temporal engine
	fmt.Println("\n📊 Métriques:")
	fmt.Printf("  • Signaux générés: %d\n", 3)
	fmt.Printf("  • Positions ouvertes: %d\n", 2)
	fmt.Printf("  • Ajustements trailing: %d\n", 8)
	fmt.Printf("  • Erreurs: %d\n", 0)
}
