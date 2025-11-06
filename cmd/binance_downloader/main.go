// Package main provides the Binance Data Downloader CLI application
// Specialized in downloading historical data from Binance Vision for backtesting
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"agent-economique/internal/cli"
)

func main() {
	fmt.Println("🏦 Binance Data Downloader - Version 1.1.0")
	fmt.Println("📊 Spécialisé dans le téléchargement de données historiques Binance Vision")
	fmt.Println()

	// Check for minimum arguments
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	// Create CLI application
	app := cli.NewCLIApp()

	// Parse command line arguments
	if err := app.ParseArguments(os.Args); err != nil {
		log.Fatalf("❌ Erreur d'arguments: %v", err)
	}

	// Execute workflow
	result, err := app.ExecuteWorkflow()
	if err != nil {
		log.Fatalf("❌ Erreur d'exécution: %v", err)
	}

	// Generate and display final report
	report, err := app.GenerateFinalReport(result)
	if err != nil {
		log.Printf("⚠️ Erreur génération rapport: %v", err)
	} else {
		printReport(report)
	}

	if result.Success {
		fmt.Println("✅ Téléchargement terminé avec succès!")
		os.Exit(0)
	} else {
		fmt.Printf("❌ Téléchargement échoué avec %d erreurs\n", len(result.Errors))
		os.Exit(1)
	}
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  binance-downloader --config <config.yaml> [options]")
	fmt.Println()
	fmt.Println("🎯 Objectif: Télécharger données historiques Binance Vision pour backtesting")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config <file>           Configuration YAML file (required)")
	fmt.Println("  --symbols <list>          Comma-separated list of symbols (overrides config)")
	fmt.Println("  --timeframes <list>       Comma-separated list of timeframes (overrides config)")
	fmt.Println("  --mode <mode>             Execution mode: default|download-only|streaming")
	fmt.Println("  --memory-limit <MB>       Memory limit in MB for streaming mode")
	fmt.Println("  --force-redownload        Force re-download of existing files")
	fmt.Println("  --verbose                 Enable verbose logging")
	fmt.Println("  --enable-metrics          Enable performance metrics collection")
	fmt.Println()
	fmt.Println("Exemples:")
	fmt.Println("  binance-downloader --config config.yaml")
	fmt.Println("  binance-downloader --config config.yaml --mode download-only")
	fmt.Println("  binance-downloader --config config.yaml --symbols SOLUSDT,ETHUSDT --timeframes 5m,1h")
	fmt.Println()
	fmt.Println("📁 Données sauvegardées dans: data/binance/")
}

// printReport displays the final execution report
func printReport(report *cli.FinalReport) {
	fmt.Println("\n📊 RAPPORT DE TÉLÉCHARGEMENT")
	fmt.Println("=" + strings.Repeat("=", 60))
	
	fmt.Printf("Résumé: %s\n", report.Summary)
	fmt.Printf("Temps d'exécution: %v\n", report.ExecutionTime)
	fmt.Printf("Taux de succès: %.1f%%\n", report.SuccessRate)
	fmt.Printf("Volume de données: %.2f MB\n", float64(report.TotalDataVolume)/(1024*1024))
	
	fmt.Println("\n📈 Symboles téléchargés:")
	for _, symbol := range report.SymbolsProcessed {
		fmt.Printf("  - %s\n", symbol)
	}
	
	fmt.Println("\n⏱️  Timeframes générés:")
	for _, tf := range report.TimeframesGenerated {
		fmt.Printf("  - %s\n", tf)
	}
	
	if len(report.Recommendations) > 0 {
		fmt.Println("\n💡 Recommandations:")
		for _, rec := range report.Recommendations {
			fmt.Printf("  💡 %s\n", rec)
		}
	}
	
	if len(report.Warnings) > 0 {
		fmt.Println("\n⚠️  Avertissements:")
		for _, warning := range report.Warnings {
			fmt.Printf("  ⚠️  %s\n", warning)
		}
	}
	
	if len(report.Errors) > 0 {
		fmt.Println("\n❌ Erreurs:")
		for _, err := range report.Errors {
			fmt.Printf("  ❌ %s\n", err.Error())
		}
	}
}
