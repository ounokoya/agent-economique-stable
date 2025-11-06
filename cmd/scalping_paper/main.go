// Package main provides scalping strategy for paper and live trading
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
	fmt.Println("🎯 SCALPING PAPER/LIVE - Trading Temps Réel")
	fmt.Println("============================================")

	// 1️⃣ Parser arguments CLI
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	mode := flag.String("mode", "paper", "Mode d'exécution: paper ou live")
	symbol := flag.String("symbol", "", "Symbole (ex: SOLUSDT) - override config")
	flag.Parse()

	// Valider mode
	if *mode != "paper" && *mode != "live" {
		log.Fatalf("❌ Mode invalide: %s (doit être 'paper' ou 'live')", *mode)
	}

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

	// Valider configuration scalping
	if config.Strategy.Name != "SCALPING" {
		log.Fatalf("❌ Stratégie doit être 'SCALPING', trouvé: %s", config.Strategy.Name)
	}

	// Confirmation supplémentaire pour LIVE
	if *mode == "live" {
		fmt.Println("\n⚠️  MODE LIVE - TRADING RÉEL AVEC ARGENT RÉEL")
		fmt.Print("Tapez 'CONFIRM' pour continuer: ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if confirmation != "CONFIRM" {
			log.Fatal("❌ Annulé par l'utilisateur")
		}
	}

	// 3️⃣ Afficher paramètres
	fmt.Printf("\n📊 Paramètres Trading:\n")
	fmt.Printf("   - Mode: %s\n", *mode)
	fmt.Printf("   - Stratégie: %s\n", config.Strategy.Name)
	fmt.Printf("   - Symbole: %s\n", config.BinanceData.Symbols[0])
	fmt.Printf("   - Timeframe: %s\n", config.Strategy.ScalpingConfig.Timeframe)
	
	if *mode == "paper" {
		fmt.Printf("   - Endpoint: https://testnet.binance.vision\n")
	} else {
		fmt.Printf("   - Endpoint: https://api.binance.com\n")
	}

	// 4️⃣ Créer application
	app := NewScalpingPaperApp(config, *mode)

	// 5️⃣ Setup signal handler pour arrêt propre
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Signal d'arrêt reçu...")
		cancel()
	}()

	// 6️⃣ Démarrer trading
	fmt.Printf("\n🚀 Démarrage %s trading...\n", *mode)
	if err := app.Run(ctx); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}

	fmt.Println("\n✅ Trading arrêté proprement")
}
