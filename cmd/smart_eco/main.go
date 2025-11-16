// Package main provides smart_eco strategy with temporal engine integration (Binance Vision)
package main

import (
	"flag"
	"fmt"
	"log"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  SMART ECO ENGINE - Moteur + Binance")
	fmt.Println("═══════════════════════════════════════════════════")

	// 1) Args CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	startDate := flag.String("start", "", "Date de début (YYYY-MM-DD) - override config")
	endDate := flag.String("end", "", "Date de fin (YYYY-MM-DD) - override config")
	symbol := flag.String("symbol", "", "Symbole (ex: SOLUSDT) - override config")
	flag.Parse()

	// 2) Charger configuration
	fmt.Printf("\n📝 Chargement configuration: %s\n", *configPath)
	config, err := shared.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Erreur chargement config: %v", err)
	}

	if *startDate != "" { config.DataPeriod.StartDate = *startDate }
	if *endDate != "" { config.DataPeriod.EndDate = *endDate }
	if *symbol != "" { config.BinanceData.Symbols = []string{*symbol} }

	// 3) Créer app
	app := NewScalpingApp(config, nil)

	// 4) Générer liste de dates
	dates, err := generateDateRange(config.DataPeriod.StartDate, config.DataPeriod.EndDate)
	if err != nil { log.Fatalf("❌ Erreur génération dates: %v", err) }
	fmt.Printf("   • Jours à traiter: %d\n", len(dates))
	app.dates = dates

	// 5) Run
	fmt.Println("\n🚀 Démarrage backtest - traitement trade par trade...")
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}
	fmt.Println("\n✅ Backtest terminé!")
}
