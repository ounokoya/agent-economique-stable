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

// Test du Stochastic TV Standard sur Gate.io
func main() {
	fmt.Println("🎯 STOCHASTIC TV STANDARD - TEST GATE.IO")
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

	// Préparer les données
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
	}

	// Initialiser et calculer le Stochastic TV Standard
	fmt.Println("\n🔧 Calcul Stochastic TV Standard sur données Gate.io...")
	stochIndicator := indicators.NewStochTVStandard(14, 3, 3)
	k, d := stochIndicator.Calculate(high, low, close)

	if k == nil || d == nil {
		fmt.Println("❌ Erreur calcul Stochastic")
		return
	}

	// Afficher les 15 dernières valeurs
	fmt.Println("\n📊 STOCHASTIC TV STANDARD - GATE.IO (15 dernières valeurs):")
	fmt.Println(strings.Repeat("=", 105))
	fmt.Printf("%-12s %-10s %-10s %-10s %-12s %-15s %-10s\n", 
		"TIME", "CLOSE", "%K", "%D", "ZONE", "POSITION", "SIGNAL")
	fmt.Println(strings.Repeat("-", 105))

	startIdx := len(klines) - 15
	for i := startIdx; i < len(klines); i++ {
		kVal := k[i]
		dVal := d[i]
		
		zone := stochIndicator.GetZone(kVal, dVal)
		position := stochIndicator.GetPositionInRange(kVal)
		signal := stochIndicator.GetSignal(kVal, dVal)
		
		fmt.Printf("%-12s %-10.2f %-10.2f %-10.2f %-12s %-15s %-10s\n",
			klines[i].OpenTime.Format("15:04"), klines[i].Close, 
			kVal, dVal, zone, position, signal)
	}

	fmt.Println(strings.Repeat("=", 105))

	// Analyse des dernières valeurs
	fmt.Println("\n📈 ANALYSE STOCHASTIC GATE.IO:")
	fmt.Println(strings.Repeat("=", 35))

	lastK, lastD := stochIndicator.GetLastValues(k, d)
	
	fmt.Printf("Dernière bougie (%s):\n", klines[len(klines)-1].OpenTime.Format("15:04"))
	fmt.Printf("  Prix: %.2f\n", klines[len(klines)-1].Close)
	fmt.Printf("  %%K: %.2f\n", lastK)
	fmt.Printf("  %%D: %.2f\n", lastD)
	fmt.Printf("  Zone: %s\n", stochIndicator.GetZone(lastK, lastD))
	fmt.Printf("  Position: %s\n", stochIndicator.GetPositionInRange(lastK))
	fmt.Printf("  Momentum: %s\n", stochIndicator.GetMomentumStrength(lastK))
	fmt.Printf("  Signal: %s\n", stochIndicator.GetSignal(lastK, lastD))
	
	// Analyse des niveaux extrêmes
	fmt.Println("\n🔍 ANALYSE DES NIVEAUX EXTRÊMES:")
	fmt.Println(strings.Repeat("=", 35))
	
	overbought80 := 0
	oversold20 := 0
	extremeOB90 := 0
	extremeOS10 := 0
	
	// Analyser les 15 dernières valeurs
	for i := startIdx; i < len(klines); i++ {
		kVal := k[i]
		dVal := d[i]
		
		if stochIndicator.IsOverbought(kVal, dVal, 80) {
			overbought80++
			if kVal > 90 && dVal > 90 {
				extremeOB90++
			}
		}
		if stochIndicator.IsOversold(kVal, dVal, 20) {
			oversold20++
			if kVal < 10 && dVal < 10 {
				extremeOS10++
			}
		}
	}
	
	fmt.Printf("  Surachat (>80): %d fois\n", overbought80)
	fmt.Printf("  Survente (<20): %d fois\n", oversold20)
	fmt.Printf("  Surachat extrême (>90): %d fois\n", extremeOB90)
	fmt.Printf("  Survente extrême (<10): %d fois\n", extremeOS10)

	// Analyse des croisements récents
	fmt.Println("\n🔄 ANALYSE DES CROISEMENTS RÉCENTS:")
	fmt.Println(strings.Repeat("=", 35))
	
	recentBullishCrosses := 0
	recentBearishCrosses := 0
	
	for i := startIdx; i < len(klines); i++ {
		if stochIndicator.IsBullishCrossover(k, d, i) {
			recentBullishCrosses++
			fmt.Printf("  🟢 Croisement haussier %%K/%%D à %s\n", klines[i].OpenTime.Format("15:04"))
		}
		if stochIndicator.IsBearishCrossover(k, d, i) {
			recentBearishCrosses++
			fmt.Printf("  🔴 Croisement baissier %%K/%%D à %s\n", klines[i].OpenTime.Format("15:04"))
		}
	}
	
	if recentBullishCrosses == 0 && recentBearishCrosses == 0 {
		fmt.Println("  ➡️ Aucun croisement %%K/%%D récent")
	}

	// Analyse des divergences
	fmt.Println("\n📊 ANALYSE DES DIVERGENCES:")
	fmt.Println(strings.Repeat("=", 30))
	
	divergence5 := stochIndicator.GetDivergenceType(close, k, 5)
	divergence10 := stochIndicator.GetDivergenceType(close, k, 10)
	
	fmt.Printf("  Divergence 5 périodes: %s\n", divergence5)
	fmt.Printf("  Divergence 10 périodes: %s\n", divergence10)

	// Statistiques sur les 15 dernières valeurs
	fmt.Println("\n📊 STATISTIQUES (15 dernières valeurs):")
	last15K := k[startIdx:]
	last15D := d[startIdx:]
	
	validCount := 0
	overboughtCount := 0
	oversoldCount := 0
	bullishCount := 0
	bearishCount := 0
	
	for i := range last15K {
		if !math.IsNaN(last15K[i]) && !math.IsNaN(last15D[i]) {
			validCount++
			if stochIndicator.IsOverbought(last15K[i], last15D[i], 80) {
				overboughtCount++
			}
			if stochIndicator.IsOversold(last15K[i], last15D[i], 20) {
				oversoldCount++
			}
			
			signal := stochIndicator.GetSignal(last15K[i], last15D[i])
			if signal == "ACHAT" || signal == "HAUSSIER" {
				bullishCount++
			} else if signal == "VENTE" || signal == "BAISSIER" {
				bearishCount++
			}
		}
	}
	
	fmt.Printf("Valeurs valides: %d/15\n", validCount)
	fmt.Printf("Zone surachat: %d fois (%.1f%%)\n", overboughtCount, float64(overboughtCount)/float64(validCount)*100)
	fmt.Printf("Zone survente: %d fois (%.1f%%)\n", oversoldCount, float64(oversoldCount)/float64(validCount)*100)
	fmt.Printf("Signaux haussiers: %d fois (%.1f%%)\n", bullishCount, float64(bullishCount)/float64(validCount)*100)
	fmt.Printf("Signaux baissiers: %d fois (%.1f%%)\n", bearishCount, float64(bearishCount)/float64(validCount)*100)

	// Analyse des extrêmes
	maxK := getMaxValue(last15K)
	minK := getMinValue(last15K)
	maxD := getMaxValue(last15D)
	minD := getMinValue(last15D)
	
	fmt.Printf("\n📈 VALEURS EXTRÊMES (15 dernières):\n")
	fmt.Printf("  %%K: Min %.2f | Max %.2f\n", minK, maxK)
	fmt.Printf("  %%D: Min %.2f | Max %.2f\n", minD, maxD)

	// Configuration actuelle
	fmt.Println("\n🎯 CONFIGURATION ACTUELLE:")
	fmt.Println(strings.Repeat("=", 30))
	
	if !math.IsNaN(lastK) && !math.IsNaN(lastD) {
		if stochIndicator.IsOverbought(lastK, lastD, 80) {
			fmt.Println("  🔴 Configuration: Surachat confirmé")
			fmt.Println("  💡 Interprétation: Risque de retournement baissier")
		} else if stochIndicator.IsOversold(lastK, lastD, 20) {
			fmt.Println("  🟢 Configuration: Survente confirmée")
			fmt.Println("  💡 Interprétation: Opportunité d'achat potentielle")
		} else if lastK > lastD {
			fmt.Println("  📈 Configuration: Momentum haussier")
			fmt.Println("  💡 Interprétation: Pression d'achat supérieure")
		} else {
			fmt.Println("  📉 Configuration: Momentum baissier")
			fmt.Println("  💡 Interprétation: Pression de vente supérieure")
		}
	}

	// Recommandations de trading
	fmt.Println("\n💡 RECOMMANDATIONS DE TRADING:")
	fmt.Println(strings.Repeat("=", 35))
	
	if overboughtCount > oversoldCount {
		fmt.Println("⚠️ Marché souvent en surachat")
		fmt.Println("   → Attendre correction avant d'acheter")
		fmt.Println("   → Considérer prendre profits sur positions longues")
	} else if oversoldCount > overboughtCount {
		fmt.Println("🚀 Marché souvent en survente")
		fmt.Println("   → Rechercher opportunités d'achat")
		fmt.Println("   → Surveiller retournements haussiers")
	} else {
		fmt.Println("➡️ Marché équilibré")
		fmt.Println("   → Utiliser croisements %%K/%%D pour entrées")
		fmt.Println("   → Confirmer avec autres indicateurs")
	}
	
	if bullishCount > bearishCount {
		fmt.Println("📊 Momentum haussier dominant")
		fmt.Println("   → Favoriser les positions longues")
		fmt.Println("   → Utiliser replis pour acheter")
	} else {
		fmt.Println("📉 Momentum baissier dominant")
		fmt.Println("   → Favoriser les positions courtes")
		fmt.Println("   → Utiliser rebonds pour vendre")
	}

	fmt.Println("\n✅ STOCHASTIC TV STANDARD testé avec succès sur Gate.io!")
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
