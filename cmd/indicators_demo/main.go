// Récupère 500 klines BingX, calcule les indicateurs, affiche les 10 dernières valeurs
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/datasource/bingx"
	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/datasource/kucoin"
	"agent-economique/internal/indicators"
	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("🔥 Strategy Demo CLI - Version 5.0.0")
	fmt.Println("🎯 Stratégie MACD + Stochastic Multi-Exchange (BingX + Gate.io + Binance + KuCoin)")
	fmt.Println("📊 Versions Classiques + Dataframes Enhanced")
	fmt.Println()

	// Check for minimum arguments
	if len(os.Args) < 6 {
		fmt.Println("Usage: indicators-demo --config <path> --symbol <symbol> --timeframe <tf>")
		fmt.Println("Example: indicators-demo --config config/config.yaml --symbol SOL-USDT --timeframe 5m")
		os.Exit(1)
	}

	// Parse command line arguments
	var configPath, symbol, timeframe string
	for i, arg := range os.Args {
		switch arg {
		case "--config":
			if i+1 < len(os.Args) {
				configPath = os.Args[i+1]
			}
		case "--symbol":
			if i+1 < len(os.Args) {
				symbol = os.Args[i+1]
			}
		case "--timeframe":
			if i+1 < len(os.Args) {
				timeframe = os.Args[i+1]
			}
		}
	}

	// Validate arguments
	if configPath == "" || symbol == "" || timeframe == "" {
		fmt.Println("❌ Tous les arguments sont requis: --config, --symbol, --timeframe")
		os.Exit(1)
	}

	// Load configuration
	config, err := shared.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Erreur lors du chargement de la configuration: %v", err)
	}

	fmt.Printf("🚀 Démarrage demo DUAL-EXCHANGE pour %s (%s) - 500 klines\n", symbol, timeframe)
	fmt.Println(strings.Repeat("=", 80))

	// Initialize clients
	fmt.Println("🔌 Initialisation BingX + Gate.io clients...")

	// Fetch klines from both exchanges
	// BingX
	_, bingxMarketService, err := initializeBingXClient(config)
	if err != nil {
		log.Printf("⚠️ Erreur initialisation BingX: %v", err)
	}

	var bingxKlines []bingx.Kline
	if bingxMarketService != nil {
		bingxKlines, err = fetchKlines(bingxMarketService, symbol, timeframe, 500)
	}
	if err != nil {
		log.Printf("⚠️ Erreur BingX: %v", err)
	}

	gateKlines, err := fetchGateIOKlines(symbol, timeframe, 500)
	if err != nil {
		log.Printf("⚠️ Erreur Gate.io: %v", err)
	}

	binanceKlines, err := fetchBinanceKlines(symbol, timeframe, 500)
	if err != nil {
		log.Printf("⚠️ Erreur Binance: %v", err)
	}

	kucoinKlines, err := fetchKuCoinKlines(symbol, timeframe, 500)
	if err != nil {
		log.Printf("⚠️ Erreur KuCoin: %v", err)
	}

	// Process results
	if bingxKlines != nil && len(bingxKlines) > 0 {
		fmt.Println("\n📊 CALCULS BINGX:")

		bingxIndicatorKlines := convertBingXToIndicatorKlines(bingxKlines)
		bingxResults, err := calculateStrategyIndicators(bingxIndicatorKlines, bingxKlines, config)
		if err == nil {
			fmt.Println("\n🏢 BINGX - Stratégie MACD + Stochastic:")
			displayStrategyValues(bingxResults, symbol+" (BingX)", timeframe)

			// Test Strategy Dataframes
			fmt.Println("\n🧮 BINGX - STRATEGY DATAFRAMES ANALYSIS:")
			testStrategyDataframes(bingxIndicatorKlines, "BingX", symbol, config)
		} else {
			fmt.Printf("⚠️ Erreur calcul BingX: %v\n", err)
		}
	}

	// Process Gate.io results
	if gateKlines != nil && len(gateKlines) > 0 {
		fmt.Println("\n📊 CALCULS GATE.IO:")
		fmt.Println(strings.Repeat("-", 50))

		gateIndicatorKlines := convertGateIOToIndicatorKlines(gateKlines)
		gateResults, err := calculateStrategyIndicatorsGateIO(gateIndicatorKlines, gateKlines, config)
		if err == nil {
			fmt.Println("\n🏢 GATE.IO - Dernières valeurs:")
			displayLast10Values(gateResults, symbol+" (Gate.io)", timeframe)
		} else {
			fmt.Printf("⚠️ Erreur calcul Gate.io: %v\n", err)
		}
	}

	// Process Binance results
	if binanceKlines != nil && len(binanceKlines) > 0 {
		fmt.Println("\n📊 CALCULS BINANCE:")
		fmt.Println(strings.Repeat("-", 50))

		binanceIndicatorKlines := convertBinanceToIndicatorKlines(binanceKlines)
		binanceResults, err := calculateAllIndicatorsBinance(binanceIndicatorKlines, binanceKlines, config)
		if err == nil {
			fmt.Println("\n🏢 BINANCE - Dernières valeurs:")
			displayLast10Values(binanceResults, symbol+" (Binance)", timeframe)
		} else {
			fmt.Printf("⚠️ Erreur calcul Binance: %v\n", err)
		}
	}

	// Process KuCoin results
	if kucoinKlines != nil && len(kucoinKlines) > 0 {
		fmt.Println("\n📊 CALCULS KUCOIN:")
		fmt.Println(strings.Repeat("-", 50))

		kucoinIndicatorKlines := convertKuCoinToIndicatorKlines(kucoinKlines)
		kucoinResults, err := calculateAllIndicatorsKuCoin(kucoinIndicatorKlines, kucoinKlines, config)
		if err == nil {
			fmt.Println("\n🏢 KUCOIN - Dernières valeurs:")
			displayLast10Values(kucoinResults, symbol+" (KuCoin)", timeframe)
		} else {
			fmt.Printf("⚠️ Erreur calcul KuCoin: %v\n", err)
		}
	}

	fmt.Println("\n🎯 Comparaison 4-exchanges terminée (BingX + Gate.io + Binance + KuCoin) - Analysez les écarts MFI !")
}

// calculateStrategyIndicators calculates MACD + Stochastic for strategy
func calculateStrategyIndicators(klines []indicators.Kline, bingxKlines []bingx.Kline, config *shared.Config) (*StrategyTimeSeries, error) {
	fmt.Printf("🎯 Calcul stratégie sur %d klines...\n", len(klines))

	if len(klines) < 50 {
		return nil, fmt.Errorf("insufficient data: need at least 50 klines, got %d", len(klines))
	}

	// Get parameters from config
	macdFast := config.Strategy.Indicators.MACD.FastPeriod
	macdSlow := config.Strategy.Indicators.MACD.SlowPeriod
	macdSignal := config.Strategy.Indicators.MACD.SignalPeriod
	stochK := config.Strategy.Indicators.Stochastic.PeriodK
	stochSmoothK := config.Strategy.Indicators.Stochastic.SmoothK
	stochD := config.Strategy.Indicators.Stochastic.PeriodD

	fmt.Printf("📋 Paramètres: MACD(%d,%d,%d) STOCH(%d,%d,%d)\n",
		macdFast, macdSlow, macdSignal, stochK, stochSmoothK, stochD)

	// Calculate MACD (classic version)
	macdLine, signalLine, histogram := indicators.MACDFromKlines(klines, macdFast, macdSlow, macdSignal, func(k indicators.Kline) float64 { return k.Close })

	// Calculate Stochastic (classic version)
	stochKValues, stochDValues := indicators.StochasticFromKlines(klines, stochK, stochSmoothK, stochD)

	return &StrategyTimeSeries{
		Timestamps: extractBingXTimestamps(bingxKlines),
		Prices:     extractPrices(klines),
		MACD:       macdLine,
		Signal:     signalLine,
		Histogram:  histogram,
		StochK:     stochKValues,
		StochD:     stochDValues,
	}, nil
}

// testStrategyDataframes tests the new MACD + Stochastic dataframes implementation
func testStrategyDataframes(klines []indicators.Kline, exchange, symbol string, config *shared.Config) {
	// Calculate enhanced strategy analysis using dataframes
	strategyAnalysis, err := indicators.CalculateStrategySignals(klines,
		config.Strategy.Indicators.MACD.FastPeriod,
		config.Strategy.Indicators.MACD.SlowPeriod,
		config.Strategy.Indicators.MACD.SignalPeriod,
		config.Strategy.Indicators.Stochastic.PeriodK,
		config.Strategy.Indicators.Stochastic.SmoothK,
		config.Strategy.Indicators.Stochastic.PeriodD)
	if err != nil {
		fmt.Printf("⚠️ Erreur calcul Strategy Dataframes: %v\n", err)
		return
	}

	// Print detailed analysis
	strategyAnalysis.PrintStrategyAnalysis(exchange, symbol)
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  indicators-demo --config <config.yaml> --symbol <SYMBOL> --timeframe <TF>")
	fmt.Println()
	fmt.Println("🎯 Objectif: Calculer indicateurs sur 500 klines BingX et afficher 10 dernières valeurs")
	fmt.Println()
	fmt.Println("Arguments obligatoires:")
	fmt.Println("  --config <file>        Configuration YAML (required)")
	fmt.Println("  --symbol <symbol>      Symbole (ex: SOL-USDT)")
	fmt.Println("  --timeframe <tf>       Timeframe (ex: 5m, 15m, 1h, 4h)")
	fmt.Println()
	fmt.Println("Exemples:")
	fmt.Println("  indicators-demo --config config.yaml --symbol SOL-USDT --timeframe 5m")
	fmt.Println("  indicators-demo --config config.yaml --symbol ETH-USDT --timeframe 1h")
}

// runIndicatorsDemo executes the indicators demo
func runIndicatorsDemo(config *shared.Config, symbol, timeframe string) error {
	fmt.Printf("🚀 Démarrage demo pour %s (%s) - 500 klines\n", symbol, timeframe)

	// 1. Initialize BingX client
	_, marketService, err := initializeBingXClient(config)
	if err != nil {
		return fmt.Errorf("failed to initialize BingX: %w", err)
	}

	// 2. Fetch 500 klines
	klines, err := fetchKlines(marketService, symbol, timeframe, 500)
	if err != nil {
		return fmt.Errorf("failed to fetch klines: %w", err)
	}

	// 3. Convert to indicators format
	indicatorKlines := convertBingXToIndicatorKlines(klines)

	// 4. Calculate all indicators with real BingX timestamps using config parameters
	results, err := calculateAllIndicators(indicatorKlines, klines, config)
	if err != nil {
		return fmt.Errorf("failed to calculate indicators: %w", err)
	}

	// 5. Display last 10 values
	displayLast10Values(results, symbol, timeframe)

	return nil
}

// initializeBingXClient sets up BingX client
func initializeBingXClient(config *shared.Config) (*bingx.Client, *bingx.MarketDataService, error) {
	// Parse timeout
	timeout, err := time.ParseDuration(config.BingXData.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	// Create client config
	clientConfig := bingx.ClientConfig{
		Environment: bingx.DemoEnvironment,
		Credentials: bingx.APICredentials{
			APIKey:    config.BingXData.Credentials.APIKey,
			SecretKey: config.BingXData.Credentials.SecretKey,
		},
		Timeout: timeout,
	}

	// Create client
	client, err := bingx.NewClient(clientConfig)
	if err != nil {
		return nil, nil, err
	}

	marketService := bingx.NewMarketDataService(client)

	fmt.Printf("🔌 BingX client initialisé (mode %s)\n", config.BingXData.Environment)
	return client, marketService, nil
}

// fetchKlines retrieves klines from BingX
func fetchKlines(marketService *bingx.MarketDataService, symbol, timeframe string, count int) ([]bingx.Kline, error) {
	fmt.Printf("📡 Récupération %d klines %s pour %s...\n", count, timeframe, symbol)

	ctx := context.Background()
	klines, err := marketService.GetKlines(ctx, symbol, timeframe, count, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines returned")
	}

	// Sort by time (oldest first)
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime.Before(klines[j].OpenTime)
	})

	fmt.Printf("✅ %d klines récupérées (de %s à %s)\n",
		len(klines),
		klines[0].OpenTime.Format("15:04:05"),
		klines[len(klines)-1].CloseTime.Format("15:04:05"))

	return klines, nil
}

// fetchGateIOKlines retrieves klines from Gate.io
func fetchGateIOKlines(symbol, timeframe string, limit int) ([]gateio.Kline, error) {
	client := gateio.NewClient()
	ctx := context.Background()

	// Convert symbol format: SOL-USDT -> SOL_USDT for Gate.io
	gateSymbol := strings.Replace(symbol, "-", "_", -1)

	fmt.Printf("📡 Récupération %d klines %s pour %s (Gate.io)...\n", limit, timeframe, gateSymbol)

	klines, err := client.GetKlines(ctx, gateSymbol, timeframe, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Gate.io klines: %v", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines returned from Gate.io")
	}

	fmt.Printf("✅ %d klines récupérées (de %s à %s) - Gate.io\n",
		len(klines),
		klines[0].OpenTime.Format("15:04:05"),
		klines[len(klines)-1].CloseTime.Format("15:04:05"))

	return klines, nil
}

// convertBingXToIndicatorKlines converts BingX klines to indicator format
func convertBingXToIndicatorKlines(bingxKlines []bingx.Kline) []indicators.Kline {
	indicatorKlines := make([]indicators.Kline, len(bingxKlines))

	for i, k := range bingxKlines {
		indicatorKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix() * 1000, // Début de barre
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	return indicatorKlines
}

// convertGateIOToIndicatorKlines converts Gate.io klines to indicator format
func convertGateIOToIndicatorKlines(gateKlines []gateio.Kline) []indicators.Kline {
	indicatorKlines := make([]indicators.Kline, len(gateKlines))

	for i, k := range gateKlines {
		indicatorKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix() * 1000, // Début de barre
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	return indicatorKlines
}

// fetchBinanceKlines retrieves klines from Binance API
func fetchBinanceKlines(symbol, interval string, limit int) ([]binance.Kline, error) {
	fmt.Printf("📡 Récupération %d klines %s pour %s (Binance)...\n", limit, interval, strings.ReplaceAll(symbol, "-", ""))

	client := binance.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert symbol format: "SOL-USDT" → "SOLUSDT"
	binanceSymbol := strings.ReplaceAll(symbol, "-", "")

	klines, err := client.GetKlines(ctx, binanceSymbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("Binance API error: %v", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no data received from Binance")
	}

	fmt.Printf("✅ %d klines récupérées (de %s à %s) - Binance\n",
		len(klines),
		klines[0].OpenTime.Format("02/01 15:04:05"),
		klines[len(klines)-1].CloseTime.Format("02/01 15:04:05"))

	return klines, nil
}

// fetchKuCoinKlines retrieves klines from KuCoin API
func fetchKuCoinKlines(symbol, interval string, limit int) ([]kucoin.Kline, error) {
	fmt.Printf("📡 Récupération %d klines %s pour %s (KuCoin)...\n", limit, interval, symbol)

	client := kucoin.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	klines, err := client.GetKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("KuCoin API error: %v", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no data received from KuCoin")
	}

	fmt.Printf("✅ %d klines récupérées (de %s à %s) - KuCoin\n",
		len(klines),
		klines[0].OpenTime.Format("02/01 15:04:05"),
		klines[len(klines)-1].CloseTime.Format("02/01 15:04:05"))

	return klines, nil
}

// convertBinanceToIndicatorKlines converts Binance klines to indicator format
func convertBinanceToIndicatorKlines(binanceKlines []binance.Kline) []indicators.Kline {
	indicatorKlines := make([]indicators.Kline, len(binanceKlines))

	for i, k := range binanceKlines {
		indicatorKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix() * 1000, // Début de barre
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	return indicatorKlines
}

// convertKuCoinToIndicatorKlines converts KuCoin klines to indicator format
func convertKuCoinToIndicatorKlines(kucoinKlines []kucoin.Kline) []indicators.Kline {
	indicatorKlines := make([]indicators.Kline, len(kucoinKlines))

	for i, k := range kucoinKlines {
		indicatorKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix() * 1000, // Début de barre
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	return indicatorKlines
}

// calculateAllIndicators calculates all indicators using config parameters
func calculateAllIndicators(klines []indicators.Kline, bingxKlines []bingx.Kline, config *shared.Config) (*StrategyTimeSeries, error) {
	fmt.Printf("🧮 Calcul des indicateurs sur %d klines avec paramètres de configuration...\n", len(klines))

	if len(klines) < 50 {
		return nil, fmt.Errorf("insufficient data: need at least 50 klines, got %d", len(klines))
	}

	// 📊 TRI CHRONOLOGIQUE OBLIGATOIRE avant calculs indicateurs
	sort.SliceStable(klines, func(i, j int) bool {
		return klines[i].Timestamp < klines[j].Timestamp
	})
	sort.SliceStable(bingxKlines, func(i, j int) bool {
		return bingxKlines[i].OpenTime.Unix() < bingxKlines[j].OpenTime.Unix()
	})

	// Vérification de l'ordre chronologique
	firstTime := time.Unix(klines[0].Timestamp/1000, 0)
	lastTime := time.Unix(klines[len(klines)-1].Timestamp/1000, 0)
	fmt.Printf("📅 Données triées chronologiquement: %s → %s\n",
		firstTime.Format("02/01 15:04:05"), lastTime.Format("02/01 15:04:05"))

	// Get parameters from config
	macdFast := config.Strategy.Indicators.MACD.FastPeriod
	macdSlow := config.Strategy.Indicators.MACD.SlowPeriod
	macdSignal := config.Strategy.Indicators.MACD.SignalPeriod
	cciPeriod := config.Strategy.Indicators.CCI.Period
	dmiPeriod := config.Strategy.Indicators.DMI.Period
	stochK := config.Strategy.Indicators.Stochastic.PeriodK
	stochSmoothK := config.Strategy.Indicators.Stochastic.SmoothK
	stochD := config.Strategy.Indicators.Stochastic.PeriodD
	mfiPeriod := config.Strategy.Indicators.MFI.Period

	fmt.Printf("📋 Paramètres: MACD(%d,%d,%d) CCI(%d) DMI(%d) STOCH(%d,%d,%d) MFI(%d)\n",
		macdFast, macdSlow, macdSignal, cciPeriod, dmiPeriod, stochK, stochSmoothK, stochD, mfiPeriod)

	// Calculate MACD using config parameters
	macdLine, signalLine, histogram := indicators.MACDFromKlines(klines, macdFast, macdSlow, macdSignal, func(k indicators.Kline) float64 {
		return k.Close
	})

	// Calculate Stochastic using config parameters
	stochKValues, stochDValues := indicators.StochasticFromKlines(klines, stochK, stochSmoothK, stochD)

	return &StrategyTimeSeries{
		Timestamps: extractBingXTimestamps(bingxKlines), // Use real BingX timestamps
		Prices:     extractPrices(klines),
		MACD:       macdLine,
		Signal:     signalLine,
		Histogram:  histogram,
		StochK:     stochKValues,
		StochD:     stochDValues,
	}, nil
}

// calculateStrategyIndicatorsGateIO calculates MACD + Stochastic for Gate.io
func calculateStrategyIndicatorsGateIO(klines []indicators.Kline, gateKlines []gateio.Kline, config *shared.Config) (*StrategyTimeSeries, error) {
	fmt.Printf("🎯 Calcul stratégie sur %d klines...\n", len(klines))

	if len(klines) < 50 {
		return nil, fmt.Errorf("insufficient data: need at least 50 klines, got %d", len(klines))
	}

	// Get parameters from config
	macdFast := config.Strategy.Indicators.MACD.FastPeriod
	macdSlow := config.Strategy.Indicators.MACD.SlowPeriod
	macdSignal := config.Strategy.Indicators.MACD.SignalPeriod
	stochK := config.Strategy.Indicators.Stochastic.PeriodK
	stochSmoothK := config.Strategy.Indicators.Stochastic.SmoothK
	stochD := config.Strategy.Indicators.Stochastic.PeriodD

	fmt.Printf("📋 Paramètres: MACD(%d,%d,%d) STOCH(%d,%d,%d)\n",
		macdFast, macdSlow, macdSignal, stochK, stochSmoothK, stochD)

	// Calculate MACD using config parameters
	macdLine, signalLine, histogram := indicators.MACDFromKlines(klines, macdFast, macdSlow, macdSignal, func(k indicators.Kline) float64 {
		return k.Close
	})

	// Calculate Stochastic using config parameters
	stochKValues, stochDValues := indicators.StochasticFromKlines(klines, stochK, stochSmoothK, stochD)

	return &StrategyTimeSeries{
		Timestamps: extractGateIOTimestamps(gateKlines),
		Prices:     extractPrices(klines),
		MACD:       macdLine,
		Signal:     signalLine,
		Histogram:  histogram,
		StochK:     stochKValues,
		StochD:     stochDValues,
	}, nil
}

// calculateAllIndicatorsBinance calculates MACD + Stochastic for Binance
func calculateAllIndicatorsBinance(klines []indicators.Kline, binanceKlines []binance.Kline, config *shared.Config) (*StrategyTimeSeries, error) {
	fmt.Printf("🎯 Calcul stratégie sur %d klines...\n", len(klines))

	if len(klines) < 50 {
		return nil, fmt.Errorf("insufficient data: need at least 50 klines, got %d", len(klines))
	}

	// Get parameters from config
	macdFast := config.Strategy.Indicators.MACD.FastPeriod
	macdSlow := config.Strategy.Indicators.MACD.SlowPeriod
	macdSignal := config.Strategy.Indicators.MACD.SignalPeriod
	stochK := config.Strategy.Indicators.Stochastic.PeriodK
	stochSmoothK := config.Strategy.Indicators.Stochastic.SmoothK
	stochD := config.Strategy.Indicators.Stochastic.PeriodD

	fmt.Printf("📋 Paramètres: MACD(%d,%d,%d) STOCH(%d,%d,%d)\n",
		macdFast, macdSlow, macdSignal, stochK, stochSmoothK, stochD)

	// Calculate MACD using config parameters
	macdLine, signalLine, histogram := indicators.MACDFromKlines(klines, macdFast, macdSlow, macdSignal, func(k indicators.Kline) float64 {
		return k.Close
	})

	// Calculate Stochastic using config parameters
	stochKValues, stochDValues := indicators.StochasticFromKlines(klines, stochK, stochSmoothK, stochD)

	return &StrategyTimeSeries{
		Timestamps: extractBinanceTimestamps(binanceKlines),
		Prices:     extractPrices(klines),
		MACD:       macdLine,
		Signal:     signalLine,
		Histogram:  histogram,
		StochK:     stochKValues,
		StochD:     stochDValues,
	}, nil
}

// calculateAllIndicatorsKuCoin calculates MACD + Stochastic for KuCoin
func calculateAllIndicatorsKuCoin(klines []indicators.Kline, kucoinKlines []kucoin.Kline, config *shared.Config) (*StrategyTimeSeries, error) {
	fmt.Printf("🎯 Calcul stratégie sur %d klines...\n", len(klines))

	if len(klines) < 50 {
		return nil, fmt.Errorf("insufficient data: need at least 50 klines, got %d", len(klines))
	}

	// Get parameters from config
	macdFast := config.Strategy.Indicators.MACD.FastPeriod
	macdSlow := config.Strategy.Indicators.MACD.SlowPeriod
	macdSignal := config.Strategy.Indicators.MACD.SignalPeriod
	stochK := config.Strategy.Indicators.Stochastic.PeriodK
	stochSmoothK := config.Strategy.Indicators.Stochastic.SmoothK
	stochD := config.Strategy.Indicators.Stochastic.PeriodD

	fmt.Printf("📋 Paramètres: MACD(%d,%d,%d) STOCH(%d,%d,%d)\n",
		macdFast, macdSlow, macdSignal, stochK, stochSmoothK, stochD)

	// Calculate MACD using config parameters
	macdLine, signalLine, histogram := indicators.MACDFromKlines(klines, macdFast, macdSlow, macdSignal, func(k indicators.Kline) float64 {
		return k.Close
	})

	// Calculate Stochastic using config parameters
	stochKValues, stochDValues := indicators.StochasticFromKlines(klines, stochK, stochSmoothK, stochD)

	return &StrategyTimeSeries{
		Timestamps: extractKuCoinTimestamps(kucoinKlines),
		Prices:     extractPrices(klines),
		MACD:       macdLine,
		Signal:     signalLine,
		Histogram:  histogram,
		StochK:     stochKValues,
		StochD:     stochDValues,
	}, nil
}

// StrategyTimeSeries holds MACD + Stochastic strategy data
type StrategyTimeSeries struct {
	Timestamps []int64
	Prices     []float64
	MACD       []float64
	Signal     []float64
	Histogram  []float64
	StochK     []float64
	StochD     []float64
}

// extractTimestamps extracts timestamps from klines
func extractTimestamps(klines []indicators.Kline) []int64 {
	timestamps := make([]int64, len(klines))
	for i, k := range klines {
		timestamps[i] = k.Timestamp
	}
	return timestamps
}

// extractBingXTimestamps extracts timestamps from BingX klines
func extractBingXTimestamps(bingxKlines []bingx.Kline) []int64 {
	timestamps := make([]int64, len(bingxKlines))
	for i, k := range bingxKlines {
		timestamps[i] = k.OpenTime.Unix() * 1000 // Real BingX open time (début de barre)
	}
	return timestamps
}

// extractGateIOTimestamps extracts timestamps from Gate.io klines
func extractGateIOTimestamps(gateKlines []gateio.Kline) []int64 {
	timestamps := make([]int64, len(gateKlines))
	for i, k := range gateKlines {
		timestamps[i] = k.OpenTime.Unix() * 1000 // Real Gate.io open time (début de barre)
	}
	return timestamps
}

// extractBinanceTimestamps extracts timestamps from Binance klines
func extractBinanceTimestamps(binanceKlines []binance.Kline) []int64 {
	timestamps := make([]int64, len(binanceKlines))
	for i, k := range binanceKlines {
		timestamps[i] = k.OpenTime.Unix() * 1000 // Real Binance open time (début de barre)
	}
	return timestamps
}

// extractKuCoinTimestamps extracts timestamps from KuCoin klines
func extractKuCoinTimestamps(kucoinKlines []kucoin.Kline) []int64 {
	timestamps := make([]int64, len(kucoinKlines))
	for i, k := range kucoinKlines {
		timestamps[i] = k.OpenTime.Unix() * 1000 // Real KuCoin open time (début de barre)
	}
	return timestamps
}

// extractPrices extracts close prices from klines
func extractPrices(klines []indicators.Kline) []float64 {
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Close
	}
	return prices
}

// displayStrategyValues displays the last 10 values of MACD + Stochastic strategy
func displayStrategyValues(results *StrategyTimeSeries, symbol, timeframe string) {
	fmt.Printf("\n📊 DERNIÈRES 10 VALEURS STRATÉGIE - %s (%s)\n", symbol, timeframe)
	fmt.Println(strings.Repeat("=", 80))

	if results == nil {
		fmt.Printf("⚠️ Aucun résultat de stratégie disponible\n")
		return
	}

	totalLen := len(results.Timestamps)
	if totalLen == 0 {
		fmt.Printf("⚠️ Aucune donnée temporelle disponible\n")
		return
	}

	start := totalLen - 10
	if start < 0 {
		start = 0
	}

	fmt.Printf("%-16s %-10s %-8s %-8s %-8s %-8s %-8s\n", 
		"Date/Heure", "Prix", "MACD", "Signal", "Histog", "StochK", "StochD")
	fmt.Println(strings.Repeat("-", 80))

	for i := start; i < totalLen; i++ {
		timestamp := time.Unix(results.Timestamps[i]/1000, 0)
		dateStr := timestamp.Format("02/01 15:04:05")

		var macdVal, signalVal, histVal, stochKVal, stochDVal string
		
		if !math.IsNaN(results.MACD[i]) && results.MACD[i] != 0 {
			macdVal = fmt.Sprintf("%.4f", results.MACD[i])
		} else {
			macdVal = "NaN"
		}
		
		if !math.IsNaN(results.Signal[i]) && results.Signal[i] != 0 {
			signalVal = fmt.Sprintf("%.4f", results.Signal[i])
		} else {
			signalVal = "NaN"
		}

		if !math.IsNaN(results.Histogram[i]) && results.Histogram[i] != 0 {
			histVal = fmt.Sprintf("%.4f", results.Histogram[i])
		} else {
			histVal = "NaN"
		}

		if !math.IsNaN(results.StochK[i]) && results.StochK[i] != 0 {
			stochKVal = fmt.Sprintf("%.2f", results.StochK[i])
		} else {
			stochKVal = "NaN"
		}

		if !math.IsNaN(results.StochD[i]) && results.StochD[i] != 0 {
			stochDVal = fmt.Sprintf("%.2f", results.StochD[i])
		} else {
			stochDVal = "NaN"
		}

		fmt.Printf("%-16s %-10.2f %-8s %-8s %-8s %-8s %-8s\n",
			dateStr, results.Prices[i], macdVal, signalVal, histVal, stochKVal, stochDVal)
	}

	fmt.Println(strings.Repeat("=", 80))
	lastTime := time.Unix(results.Timestamps[totalLen-1]/1000, 0)
	fmt.Printf("📈 Dernière valeur (la plus récente): %s - Prix: %.2f\n", 
		lastTime.Format("02/01/2006 15:04:05"), results.Prices[totalLen-1])
}

// displayLast10Values is an alias for displayStrategyValues for backward compatibility
func displayLast10Values(results *StrategyTimeSeries, symbol, timeframe string) {
	displayStrategyValues(results, symbol, timeframe)
}

