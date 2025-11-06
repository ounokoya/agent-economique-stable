// Package main provides scalping strategy for LIVE trading
// This is a dedicated live trading entry point (forces live mode)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println("🎯 SCALPING LIVE - Trading Production")
	fmt.Println("========================================")

	// 1️⃣ Parser arguments CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	symbol := flag.String("symbol", "", "Symbole (ex: SOLUSDT) - override config")
	flag.Parse()

	// 2️⃣ Charger configuration
	fmt.Printf("\n📋 Chargement configuration: %s\n", *configPath)
	config, err := shared.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Erreur chargement config: %v", err)
	}

	// Override avec arguments CLI si fournis
	if *symbol != "" {
		config.BinanceData.Symbols = []string{*symbol}
	}

	// 3️⃣ Afficher paramètres
	fmt.Println("\n📊 Paramètres Trading:")
	fmt.Printf("   - Mode: live\n")
	fmt.Printf("   - Stratégie: %s\n", config.Strategy.Name)
	fmt.Printf("   - Symbole: %s\n", config.BinanceData.Symbols[0])
	fmt.Printf("   - Timeframe: %s\n", config.Strategy.ScalpingConfig.Timeframe)
	fmt.Println("   - Endpoint: PRODUCTION BINANCE")

	// 4️⃣ Créer application (MODE LIVE FORCÉ)
	app := NewScalpingLiveApp(config, "live")

	// 5️⃣ Gérer arrêt gracieux
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Signal d'arrêt reçu...")
		cancel()
	}()

	// 6️⃣ Démarrer trading LIVE
	fmt.Println("\n🚀 Démarrage LIVE trading...")
	if err := app.Run(ctx); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}

	fmt.Println("\n✅ Trading arrêté proprement")
}
