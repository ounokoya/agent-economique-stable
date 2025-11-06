// Package main provides the Engine Demo CLI application
// Specialized in demonstrating the temporal engine with live BingX data
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("🔥 Engine Demo CLI - Version 1.1.0")
	fmt.Println("🚀 Démo du Temporal Engine avec données BingX RÉELLES")
	fmt.Println()

	// Check for minimum arguments
	if len(os.Args) < 5 {
		printUsage()
		os.Exit(1)
	}

	// Parse arguments
	configPath := ""
	symbol := ""
	timeframe := ""
	candleCount := 100

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--config":
			if i+1 >= len(os.Args) {
				log.Fatal("❌ --config requires a value")
			}
			configPath = os.Args[i+1]
			i++
		case "--symbol":
			if i+1 >= len(os.Args) {
				log.Fatal("❌ --symbol requires a value")
			}
			symbol = os.Args[i+1]
			i++
		case "--timeframe":
			if i+1 >= len(os.Args) {
				log.Fatal("❌ --timeframe requires a value")
			}
			timeframe = os.Args[i+1]
			i++
		case "--candles":
			if i+1 >= len(os.Args) {
				log.Fatal("❌ --candles requires a value")
			}
			var err error
			candleCount, err = strconv.Atoi(os.Args[i+1])
			if err != nil {
				log.Fatalf("❌ Invalid candle count: %s", os.Args[i+1])
			}
			i++
		default:
			log.Fatalf("❌ Unknown argument: %s", os.Args[i])
		}
	}

	if configPath == "" || symbol == "" || timeframe == "" {
		log.Fatal("❌ Missing required arguments")
	}

	// Load configuration
	config, err := shared.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Create engine demo application
	app := NewEngineDemoApp(config, symbol, timeframe, candleCount)

	// Execute demo workflow
	err = app.RunDemo()
	if err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}

	fmt.Println("✅ Test terminé avec succès!")
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  demo --config <config.yaml> --symbol <SYMBOL> --timeframe <TF> [options]")
	fmt.Println()
	fmt.Println("🎯 Objectif: Démonstration Temporal Engine avec données BingX RÉELLES")
	fmt.Println()
	fmt.Println("Arguments obligatoires:")
	fmt.Println("  --config <file>        Configuration YAML (required)")
	fmt.Println("  --symbol <symbol>      Symbole à tester (ex: SOLUSDT)")
	fmt.Println("  --timeframe <tf>       Timeframe (ex: 5m, 15m, 1h, 4h)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --candles <count>      Nombre de bougies à récupérer (défaut: 100)")
	fmt.Println()
	fmt.Println("Exemples:")
	fmt.Println("  demo --config config.yaml --symbol SOLUSDT --timeframe 5m")
	fmt.Println("  demo --config config.yaml --symbol ETHUSDT --timeframe 1h --candles 200")
	fmt.Println()
	fmt.Println("📊 Fonctionnalités:")
	fmt.Println("  • Récupération données BingX via SDK")
	fmt.Println("  • Calcul indicateurs (MACD/CCI/DMI ou STOCH/MFI/CCI)")
	fmt.Println("  • Simulation temporal engine")
	fmt.Println("  • Logs détaillés à chaque bougie")
}
