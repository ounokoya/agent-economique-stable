// Package main provides ban_eco: simple scanner for body% and ATR(3) filters
package main

import (
	"flag"
	"fmt"
	"log"

	"agent-economique/internal/shared"
)

func main() {
	fmt.Println(" BAN ECO - Body% & ATR(3) filter")
	fmt.Println("=================================")

	configPath := flag.String("config", "config/config.yaml", "Chemin vers la config (optionnel, pour symbole par défaut)")
	symbol := flag.String("symbol", "", "Symbole ex: SOLUSDT (si vide, lit le premier du YAML)")
	n := flag.Int("n", 1000, "Nombre de bougies (dernières)")
	flag.Parse()

	var sym string
	if *symbol != "" {
		sym = *symbol
	} else {
		cfg, err := shared.LoadConfig(*configPath)
		if err != nil { log.Fatalf(" Erreur chargement config: %v", err) }
		if len(cfg.BinanceData.Symbols) == 0 {
			log.Fatalf(" Aucun symbole fourni (-symbol) et aucun dans le YAML")
		}
		sym = cfg.BinanceData.Symbols[0]
	}

	app := NewBanEcoApp(sym, *n)
	fmt.Printf("\n Scan sur %s, dernières %d bougies, timeframe=1m\n", sym, *n)
	if err := app.Run(); err != nil {
		log.Fatalf(" Erreur: %v", err)
	}
	fmt.Println("\n Terminé")
}
