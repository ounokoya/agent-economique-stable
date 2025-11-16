// Package main provides Smart ECO strategy for LIVE trading on Gate.io
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
	fmt.Println(" SMART ECO LIVE GATE.IO - Production")
	fmt.Println("========================================")

	// 1) CLI flags
	configPath := flag.String("config", "config/config.yaml", "Chemin vers le fichier de configuration")
	symbol := flag.String("symbol", "", "Symbole (ex: SOL_USDT ou SOLUSDT)")
	nInit := flag.Int("ninit", 300, "Nombre de klines initiales à charger")
	nUpdate := flag.Int("nupdate", 10, "Nombre de klines à rafraîchir à chaque tick")
	flag.Parse()

	// 2) Load config
	fmt.Printf("\n Chargement configuration: %s\n", *configPath)
	config, err := shared.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf(" Erreur chargement config: %v", err)
	}

	// Optional override symbol
	if *symbol != "" {
		// Accept SOLUSDT or SOL_USDT. Normaliser vers Binance format pour la config globale
		cfgSym := *symbol
		if len(cfgSym) >= 2 && cfgSym[len(cfgSym)-5:] == "_USDT" {
			cfgSym = cfgSym[:len(cfgSym)-5] + "USDT"
		}
		config.BinanceData.Symbols = []string{cfgSym}
	}

	fmt.Println("\n Paramètres:")
	fmt.Printf("   - Mode: live\n")
	fmt.Printf("   - Stratégie: smart_eco\n")
	fmt.Printf("   - Symbole: %s\n", config.BinanceData.Symbols[0])
	fmt.Printf("   - Timeframe: %s\n", config.Strategy.ScalpingConfig.Timeframe)
	fmt.Println("   - Exchange: Gate.io")

	// 3) Create app
	app := NewSmartEcoLiveGateIOApp(config, *nInit, *nUpdate)

	// 4) Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Signal d'arrêt reçu...")
		cancel()
	}()

	// 5) Run
	fmt.Println("\n🚀 Démarrage LIVE Smart ECO Gate.io...")
	if err := app.Run(ctx); err != nil {
		log.Fatalf("❌ Erreur exécution: %v", err)
	}
	fmt.Println("\n✅ Arrêt propre")
}
