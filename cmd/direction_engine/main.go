// Package main provides direction strategy with temporal engine integration
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  DIRECTION ENGINE - Moteur Temporel + Binance")
	fmt.Println("═══════════════════════════════════════════════════")

	// 1️⃣ Parser arguments CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	startDate := flag.String("start", "", "Date de début (YYYY-MM-DD) - override config")
	endDate := flag.String("end", "", "Date de fin (YYYY-MM-DD) - override config")
	symbol := flag.String("symbol", "", "Symbole (ex: SOLUSDT) - override config")
	flag.Parse()

	// 2️⃣ Charger configuration
	fmt.Printf("\n📝 Chargement configuration: %s\n", *configPath)
	config, err := shared.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Erreur chargement config: %v", err)
	}

	// Override avec arguments CLI si fournis
	if *startDate != "" {
		config.DataPeriod.StartDate = *startDate
	}
	if *endDate != "" {
		config.DataPeriod.EndDate = *endDate
	}
	if *symbol != "" {
		config.BinanceData.Symbols = []string{*symbol}
	}

	// 3️⃣ Charger config direction (YAML ou defaults)
	app := NewDirectionEngineApp(config, nil) // dates seront assignées après
	dirCfg := app.directionCfg

	// 4️⃣ Afficher paramètres backtest
	fmt.Printf("\n⚙️  Paramètres Backtest:\n")
	fmt.Printf("   • Symbole: %s\n", config.BinanceData.Symbols[0])
	fmt.Printf("   • Période: %s → %s\n", config.DataPeriod.StartDate, config.DataPeriod.EndDate)
	fmt.Printf("\n📊 Paramètres Direction:\n")
	fmt.Printf("   • Timeframe: %s\n", dirCfg.Timeframe)
	fmt.Printf("   • VWMA: %d\n", dirCfg.VWMAPeriod)
	fmt.Printf("   • Slope: %d\n", dirCfg.SlopePeriod)
	fmt.Printf("   • K-Confirmation: %d\n", dirCfg.KConfirmation)
	fmt.Printf("   • ATR: %d (coef %.2f)\n", dirCfg.ATRPeriod, dirCfg.ATRCoefficient)
	if dirCfg.UseDynamicThreshold {
		fmt.Printf("   • Seuil: DYNAMIQUE (ATR × %.2f)\n", dirCfg.ATRCoefficient)
	} else {
		fmt.Printf("   • Seuil: FIXE (%.2f)\n", dirCfg.FixedThreshold)
	}
	fmt.Printf("\n💾 Cache: %s\n", config.BinanceData.CacheRoot)

	// 5️⃣ Générer liste de dates pour la période
	dates, err := generateDateRange(config.DataPeriod.StartDate, config.DataPeriod.EndDate)
	if err != nil {
		log.Fatalf("❌ Erreur génération dates: %v", err)
	}
	fmt.Printf("   • Jours à traiter: %d\n", len(dates))

	// 6️⃣ Assigner les dates à l'application
	app.dates = dates

	fmt.Println("\n🚀 Démarrage backtest - traitement trade par trade...")
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}

	fmt.Println("\n✅ Backtest terminé!")
}

// generateDateRange génère la liste des dates entre start et end
func generateDateRange(startStr, endStr string) ([]string, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, fmt.Errorf("date début invalide: %w", err)
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return nil, fmt.Errorf("date fin invalide: %w", err)
	}

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}

	return dates, nil
}
