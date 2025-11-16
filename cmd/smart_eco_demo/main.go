// Package main provides Smart ECO demo that loads the last N klines and loops over candles
package main

import (
	"flag"
	"fmt"
	"log"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("🧪 SMART ECO DEMO - Last N Klines")
	fmt.Println("=================================")

	// CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	symbol := flag.String("symbol", "", "Symbole (ex: SOLUSDT) - override config")
	n := flag.Int("n", 1000, "Nombre de bougies à charger (dernieres)")
	flag.Parse()

	// Config
	fmt.Printf("\n📝 Chargement configuration: %s\n", *configPath)
	config, err := shared.LoadConfig(*configPath)
	if err != nil { log.Fatalf("❌ Erreur chargement config: %v", err) }
	if *symbol != "" { config.BinanceData.Symbols = []string{*symbol} }

	// App
	app := NewSmartEcoDemoApp(config, *n)

	// Run
	fmt.Println("\n🚀 Démarrage DEMO Smart ECO (boucle sur N bougies)...")
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}
	fmt.Println("\n✅ Demo terminée")
}
