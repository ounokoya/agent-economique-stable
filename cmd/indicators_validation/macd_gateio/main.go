package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/indicators"
)

// Test du MACD TV Standard sur Gate.io
func main() {
	fmt.Println("🎯 MACD TV STANDARD - TEST GATE.IO")
	fmt.Println("=" + strings.Repeat("=", 45))

	// Créer le client Gate.io
	client := gateio.NewClient()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Récupérer 300 klines depuis Gate.io
	fmt.Println("📡 Récupération des 300 dernières klines depuis Gate.io...")
	klines, err := client.GetKlines(ctx, "SOL_USDT", "5m", 300)
	if err != nil {
		fmt.Printf("❌ Erreur klines Gate.io: %v\n", err)
		return
	}

	// Trier chronologiquement
	for i := 0; i < len(klines); i++ {
		for j := i + 1; j < len(klines); j++ {
			if klines[j].OpenTime.Before(klines[i].OpenTime) {
				klines[i], klines[j] = klines[j], klines[i]
			}
		}
	}

	fmt.Printf("✅ %d klines récupérées depuis Gate.io\n", len(klines))

	// Préparer les données (utiliser les prix de clôture)
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Close
	}

	// Initialiser et calculer le MACD TV Standard
	fmt.Println("\n🔧 Calcul MACD TV Standard sur données Gate.io...")
	macdIndicator := indicators.NewMACDTVStandard(12, 26, 9)
	macdLine, signalLine, histogram := macdIndicator.Calculate(prices)

	if macdLine == nil || signalLine == nil || histogram == nil {
		fmt.Println("❌ Erreur calcul MACD")
		return
	}

	// Afficher les 15 dernières valeurs
	fmt.Println("\n📊 MACD TV STANDARD - GATE.IO (15 dernières valeurs):")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Printf("%-12s %-10s %-12s %-12s %-12s %-12s %-10s\n", 
		"TIME", "CLOSE", "MACD_LINE", "SIGNAL", "HISTOGRAM", "MOMENTUM", "SIGNAL")
	fmt.Println(strings.Repeat("-", 95))

	startIdx := len(klines) - 15
	for i := startIdx; i < len(klines); i++ {
		k := klines[i]
		macdVal := macdLine[i]
		signalVal := signalLine[i]
		histVal := histogram[i]
		
		momentum := macdIndicator.GetMomentumStrength(histVal)
		signal := macdIndicator.GetMACDSignal(macdVal, signalVal)
		
		fmt.Printf("%-12s %-10.2f %-12.4f %-12.4f %-12.4f %-12s %-10s\n",
			k.OpenTime.Format("15:04"), k.Close, 
			macdVal, signalVal, histVal, momentum, signal)
	}

	fmt.Println(strings.Repeat("=", 95))

	// Analyse des dernières valeurs
	fmt.Println("\n📈 ANALYSE MACD GATE.IO:")
	fmt.Println(strings.Repeat("=", 30))

	lastMACD, lastSignal, lastHist := macdIndicator.GetLastValues(macdLine, signalLine, histogram)
	
	fmt.Printf("Dernière bougie (%s):\n", klines[len(klines)-1].OpenTime.Format("15:04"))
	fmt.Printf("  Prix: %.2f\n", klines[len(klines)-1].Close)
	fmt.Printf("  MACD Line: %.4f\n", lastMACD)
	fmt.Printf("  Signal Line: %.4f\n", lastSignal)
	fmt.Printf("  Histogram: %.4f\n", lastHist)
	fmt.Printf("  Momentum: %s\n", macdIndicator.GetMomentumStrength(lastHist))
	fmt.Printf("  Signal MACD: %s\n", macdIndicator.GetMACDSignal(lastMACD, lastSignal))
	
	// Analyse des croisements récents
	fmt.Println("\n🔍 ANALYSE DES CROISEMENTS RÉCENTS:")
	fmt.Println(strings.Repeat("=", 35))
	
	recentBullishCrosses := 0
	recentBearishCrosses := 0
	recentZeroUp := 0
	recentZeroDown := 0
	
	// Analyser les 15 dernières valeurs
	for i := startIdx; i < len(klines); i++ {
		if macdIndicator.IsBullishCrossover(macdLine, signalLine, i) {
			recentBullishCrosses++
			fmt.Printf("  🟢 Croisement haussier MACD/Signal à %s\n", klines[i].OpenTime.Format("15:04"))
		}
		if macdIndicator.IsBearishCrossover(macdLine, signalLine, i) {
			recentBearishCrosses++
			fmt.Printf("  🔴 Croisement baissier MACD/Signal à %s\n", klines[i].OpenTime.Format("15:04"))
		}
		if macdIndicator.IsZeroLineCrossoverUp(macdLine, i) {
			recentZeroUp++
			fmt.Printf("  ⬆️ Croisement ligne zéro vers le haut à %s\n", klines[i].OpenTime.Format("15:04"))
		}
		if macdIndicator.IsZeroLineCrossoverDown(macdLine, i) {
			recentZeroDown++
			fmt.Printf("  ⬇️ Croisement ligne zéro vers le bas à %s\n", klines[i].OpenTime.Format("15:04"))
		}
	}
	
	if recentBullishCrosses == 0 && recentBearishCrosses == 0 {
		fmt.Println("  ➡️ Aucun croisement MACD/Signal récent")
	}
	if recentZeroUp == 0 && recentZeroDown == 0 {
		fmt.Println("  ➡️ Aucun croisement ligne zéro récent")
	}

	// Analyse des divergences
	fmt.Println("\n📊 ANALYSE DES DIVERGENCES:")
	fmt.Println(strings.Repeat("=", 30))
	
	divergence5 := macdIndicator.GetDivergenceType(prices, macdLine, 5)
	divergence10 := macdIndicator.GetDivergenceType(prices, macdLine, 10)
	
	fmt.Printf("  Divergence 5 périodes: %s\n", divergence5)
	fmt.Printf("  Divergence 10 périodes: %s\n", divergence10)

	// Statistiques sur les 15 dernières valeurs
	fmt.Println("\n📊 STATISTIQUES (15 dernières valeurs):")
	last15MACD := macdLine[startIdx:]
	last15Signal := signalLine[startIdx:]
	last15Hist := histogram[startIdx:]
	
	validCount := 0
	bullishCount := 0
	bearishCount := 0
	positiveHist := 0
	negativeHist := 0
	
	for i := range last15MACD {
		if !math.IsNaN(last15MACD[i]) && !math.IsNaN(last15Signal[i]) {
			validCount++
			if macdIndicator.GetMACDSignal(last15MACD[i], last15Signal[i]) == "Haussier" {
				bullishCount++
			} else if macdIndicator.GetMACDSignal(last15MACD[i], last15Signal[i]) == "Baissier" {
				bearishCount++
			}
			
			if !math.IsNaN(last15Hist[i]) {
				if last15Hist[i] > 0 {
					positiveHist++
				} else if last15Hist[i] < 0 {
					negativeHist++
				}
			}
		}
	}
	
	fmt.Printf("Valeurs valides: %d/15\n", validCount)
	fmt.Printf("Signal haussier: %d fois (%.1f%%)\n", bullishCount, float64(bullishCount)/float64(validCount)*100)
	fmt.Printf("Signal baissier: %d fois (%.1f%%)\n", bearishCount, float64(bearishCount)/float64(validCount)*100)
	fmt.Printf("Histogram positif: %d fois (%.1f%%)\n", positiveHist, float64(positiveHist)/float64(validCount)*100)
	fmt.Printf("Histogram négatif: %d fois (%.1f%%)\n", negativeHist, float64(negativeHist)/float64(validCount)*100)

	// Analyse des extrêmes
	maxMACD := getMaxValue(last15MACD)
	minMACD := getMinValue(last15MACD)
	maxSignal := getMaxValue(last15Signal)
	minSignal := getMinValue(last15Signal)
	maxHist := getMaxValue(last15Hist)
	minHist := getMinValue(last15Hist)
	
	fmt.Printf("\n📈 VALEURS EXTRÊMES (15 dernières):\n")
	fmt.Printf("  MACD Line: Min %.4f | Max %.4f\n", minMACD, maxMACD)
	fmt.Printf("  Signal Line: Min %.4f | Max %.4f\n", minSignal, maxSignal)
	fmt.Printf("  Histogram: Min %.4f | Max %.4f\n", minHist, maxHist)

	// Configuration actuelle
	fmt.Println("\n🎯 CONFIGURATION ACTUELLE:")
	fmt.Println(strings.Repeat("=", 30))
	
	if !math.IsNaN(lastMACD) && !math.IsNaN(lastSignal) {
		if lastMACD > lastSignal {
			if lastHist > 0 {
				fmt.Println("  📈 Configuration: MACD > Signal + Histogram positif")
				fmt.Println("  💡 Interprétation: Momentum haussier confirmé")
			} else {
				fmt.Println("  ⚠️ Configuration: MACD > Signal mais Histogram négatif")
				fmt.Println("  💡 Interprétation: Signal haussier affaibli")
			}
		} else {
			if lastHist < 0 {
				fmt.Println("  📉 Configuration: MACD < Signal + Histogram négatif")
				fmt.Println("  💡 Interprétation: Momentum baissier confirmé")
			} else {
				fmt.Println("  ⚠️ Configuration: MACD < Signal mais Histogram positif")
				fmt.Println("  💡 Interprétation: Signal baissier affaibli")
			}
		}
	}

	// Recommandations de trading
	fmt.Println("\n💡 RECOMMANDATIONS DE TRADING:")
	fmt.Println(strings.Repeat("=", 35))
	
	if bullishCount > bearishCount {
		fmt.Println("📊 Tendance MACD globalement haussière")
		fmt.Println("   → Rechercher des opportunités d'achat")
		fmt.Println("   → Attendre confirmation croisement MACD/Signal")
	} else if bearishCount > bullishCount {
		fmt.Println("📉 Tendance MACD globalement baissière")
		fmt.Println("   → Rechercher des opportunités de vente")
		fmt.Println("   → Attendre confirmation croisement MACD/Signal")
	} else {
		fmt.Println("➡️ Tendance MACD neutre / changeante")
		fmt.Println("   → Attendre une direction claire")
		fmt.Println("   → Utiliser filtres additionnels")
	}
	
	if positiveHist > negativeHist {
		fmt.Println("🚀 Momentum positif dominant")
		fmt.Println("   → Pression d'achat supérieure")
		fmt.Println("   → Favoriser les positions longues")
	} else {
		fmt.Println("📉 Momentum négatif dominant")
		fmt.Println("   → Pression de vente supérieure")
		fmt.Println("   → Favoriser les positions courtes")
	}

	fmt.Println("\n✅ MACD TV STANDARD testé avec succès sur Gate.io!")
}

// Fonctions utilitaires
func getMaxValue(values []float64) float64 {
	max := math.Inf(-1)
	for _, v := range values {
		if !math.IsNaN(v) && v > max {
			max = v
		}
	}
	return max
}

func getMinValue(values []float64) float64 {
	min := math.Inf(1)
	for _, v := range values {
		if !math.IsNaN(v) && v < min {
			min = v
		}
	}
	return min
}
