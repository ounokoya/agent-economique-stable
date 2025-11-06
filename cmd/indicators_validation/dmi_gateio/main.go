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

// Test du DMI TV Standard sur Gate.io
func main() {
	fmt.Println("🎯 DMI TV STANDARD - TEST GATE.IO")
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

	// Initialiser et calculer le DMI TV Standard
	fmt.Println("\n🔧 Calcul DMI TV Standard sur données Gate.io...")
	dmiIndicator := indicators.NewDMITVStandard(14)
	plusDI, minusDI, adx := dmiIndicator.Calculate(high, low, close)

	if plusDI == nil || minusDI == nil || adx == nil {
		fmt.Println("❌ Erreur calcul DMI")
		return
	}

	// Afficher les 15 dernières valeurs
	fmt.Println("\n📊 DMI TV STANDARD - GATE.IO (15 dernières valeurs):")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Printf("%-12s %-10s %-10s %-10s %-10s %-10s %-12s %-10s\n", 
		"TIME", "CLOSE", "+DI", "-DI", "DX", "ADX", "TREND", "SIGNAL")
	fmt.Println(strings.Repeat("-", 95))

	startIdx := len(klines) - 15
	for i := startIdx; i < len(klines); i++ {
		k := klines[i]
		plusDIVal := plusDI[i]
		minusDIVal := minusDI[i]
		adxVal := adx[i]
		
		// Calculer DX pour l'affichage
		dxVal := calculateDXForDisplay(plusDIVal, minusDIVal)
		
		trend := getTrendInfo(plusDIVal, minusDIVal, adxVal)
		signal := getSignalInfo(plusDIVal, minusDIVal, adxVal)
		
		fmt.Printf("%-12s %-10.2f %-10.2f %-10.2f %-10.2f %-10.2f %-12s %-10s\n",
			k.OpenTime.Format("15:04"), k.Close, 
			plusDIVal, minusDIVal, dxVal, adxVal, trend, signal)
	}

	fmt.Println(strings.Repeat("=", 85))

	// Analyse des dernières valeurs
	fmt.Println("\n📈 ANALYSE DMI GATE.IO:")
	fmt.Println(strings.Repeat("=", 30))

	lastPlusDI, lastMinusDI, lastADX := dmiIndicator.GetLastValues(plusDI, minusDI, adx)
	lastDX := calculateDXForDisplay(lastPlusDI, lastMinusDI)
	
	fmt.Printf("Dernière bougie (%s):\n", klines[len(klines)-1].OpenTime.Format("15:04"))
	fmt.Printf("  Prix: %.2f\n", klines[len(klines)-1].Close)
	fmt.Printf("  +DI: %.2f\n", lastPlusDI)
	fmt.Printf("  -DI: %.2f\n", lastMinusDI)
	fmt.Printf("  DX: %.2f\n", lastDX)
	fmt.Printf("  ADX: %.2f\n", lastADX)
	fmt.Printf("  Force tendance: %s\n", dmiIndicator.GetTrendStrength(lastADX))
	fmt.Printf("  Direction tendance: %s\n", dmiIndicator.GetTrendDirection(lastPlusDI, lastMinusDI))
	
	// Signal actuel
	signal := dmiIndicator.GetSignal(lastPlusDI, lastMinusDI, lastADX)
	fmt.Printf("  🎯 SIGNAL ACTUEL: %s\n", signal)
	
	// Analyse détaillée
	if dmiIndicator.IsStrongTrend(lastADX) {
		if dmiIndicator.IsBullish(lastPlusDI, lastMinusDI) {
			fmt.Printf("  📈 CONFIGURATION: Trend fort haussier (+DI > -DI et ADX > 25)\n")
		} else if dmiIndicator.IsBearish(lastPlusDI, lastMinusDI) {
			fmt.Printf("  📉 CONFIGURATION: Trend fort baissier (-DI > +DI et ADX > 25)\n")
		}
	} else {
		fmt.Printf("  ⚠️ CONFIGURATION: Trend faible ou inexistant (ADX < 25)\n")
		fmt.Printf("  💡 CONSEIL: Attendre un ADX > 25 pour des signaux fiables\n")
	}

	// Statistiques sur les 15 dernières valeurs
	fmt.Println("\n📊 STATISTIQUES (15 dernières valeurs):")
	last15PlusDI := plusDI[startIdx:]
	last15MinusDI := minusDI[startIdx:]
	last15ADX := adx[startIdx:]
	
	validCount := 0
	strongTrendCount := 0
	bullishCount := 0
	bearishCount := 0
	buySignals := 0
	sellSignals := 0
	
	for i := range last15ADX {
		if !math.IsNaN(last15ADX[i]) {
			validCount++
			if dmiIndicator.IsStrongTrend(last15ADX[i]) {
				strongTrendCount++
			}
			if dmiIndicator.IsBullish(last15PlusDI[i], last15MinusDI[i]) {
				bullishCount++
			}
			if dmiIndicator.IsBearish(last15PlusDI[i], last15MinusDI[i]) {
				bearishCount++
			}
			
			signal := dmiIndicator.GetSignal(last15PlusDI[i], last15MinusDI[i], last15ADX[i])
			if signal == "ACHAT" {
				buySignals++
			} else if signal == "VENTE" {
				sellSignals++
			}
		}
	}
	
	fmt.Printf("Valeurs valides: %d/15\n", validCount)
	fmt.Printf("Trends forts (ADX > 25): %d fois (%.1f%%)\n", strongTrendCount, float64(strongTrendCount)/float64(validCount)*100)
	fmt.Printf("Direction haussière: %d fois (%.1f%%)\n", bullishCount, float64(bullishCount)/float64(validCount)*100)
	fmt.Printf("Direction baissière: %d fois (%.1f%%)\n", bearishCount, float64(bearishCount)/float64(validCount)*100)
	fmt.Printf("Signaux d'achat: %d\n", buySignals)
	fmt.Printf("Signaux de vente: %d\n", sellSignals)

	// Analyse des extrêmes
	maxADX := getMaxValue(last15ADX)
	minADX := getMinValue(last15ADX)
	maxPlusDI := getMaxValue(last15PlusDI)
	maxMinusDI := getMaxValue(last15MinusDI)
	
	// Calculer les valeurs DX pour les statistiques
	last15DX := make([]float64, len(last15PlusDI))
	for i := range last15PlusDI {
		last15DX[i] = calculateDXForDisplay(last15PlusDI[i], last15MinusDI[i])
	}
	maxDX := getMaxValue(last15DX)
	minDX := getMinValue(last15DX)
	
	fmt.Printf("\n📈 VALEURS EXTRÊMES (15 dernières):\n")
	fmt.Printf("  DX: Min %.2f | Max %.2f\n", minDX, maxDX)
	fmt.Printf("  ADX: Min %.2f | Max %.2f\n", minADX, maxADX)
	fmt.Printf("  +DI: Max %.2f\n", maxPlusDI)
	fmt.Printf("  -DI: Max %.2f\n", maxMinusDI)

	// Recommandations de trading
	fmt.Println("\n💡 RECOMMANDATIONS DE TRADING:")
	fmt.Println(strings.Repeat("=", 35))
	
	if strongTrendCount > validCount/2 {
		fmt.Println("🚀 Marché en tendance forte fréquente")
		fmt.Println("   → Utiliser les croisements DI pour entrées/sorties")
		fmt.Println("   → ADX > 25 confirme la validité des signaux")
	} else {
		fmt.Println("⚠️ Marché souvent sans tendance claire")
		fmt.Println("   → Éviter les signaux DMI en range")
		fmt.Println("   → Attendre ADX > 25 avant de trader")
	}
	
	if bullishCount > bearishCount {
		fmt.Println("📊 Tendance générale haussière dominante")
		fmt.Println("   → Favoriser les positions longues")
		fmt.Println("   → Chercher les points d'entrée sur +DI > -DI")
	} else if bearishCount > bullishCount {
		fmt.Println("📉 Tendance générale baissière dominante")
		fmt.Println("   → Favoriser les positions courtes")
		fmt.Println("   → Chercher les points d'entrée sur -DI > +DI")
	} else {
		fmt.Println("➡️ Tendance neutre / changeante")
		fmt.Println("   → Attendre une direction claire")
		fmt.Println("   → Utiliser filtres additionnels")
	}

	fmt.Println("\n✅ DMI TV STANDARD testé avec succès sur Gate.io!")
}

// Fonctions utilitaires
func getTrendInfo(plusDI, minusDI, adx float64) string {
	if math.IsNaN(adx) {
		return "Inconnu"
	}
	
	strength := getADXStrength(adx)
	direction := getDIDirection(plusDI, minusDI)
	
	return fmt.Sprintf("%s %s", strength, direction)
}

func getSignalInfo(plusDI, minusDI, adx float64) string {
	if math.IsNaN(adx) || adx <= 25 {
		return "RANGE"
	}
	
	if plusDI > minusDI {
		return "ACHAT"
	} else if minusDI > plusDI {
		return "VENTE"
	} else {
		return "NEUTRE"
	}
}

func getADXStrength(adx float64) string {
	if adx > 25 {
		return "Fort"
	} else if adx > 20 {
		return "Modéré"
	} else if adx > 15 {
		return "Faible"
	} else {
		return "Nul"
	}
}

func getDIDirection(plusDI, minusDI float64) string {
	if math.IsNaN(plusDI) || math.IsNaN(minusDI) {
		return "?"
	}
	
	if plusDI > minusDI {
		return "↑"
	} else if minusDI > plusDI {
		return "↓"
	} else {
		return "→"
	}
}

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

// calculateDXForDisplay calcule le DX pour l'affichage
func calculateDXForDisplay(plusDI, minusDI float64) float64 {
	if math.IsNaN(plusDI) || math.IsNaN(minusDI) {
		return math.NaN()
	}
	
	sum := plusDI + minusDI
	if sum != 0 {
		return 100 * math.Abs(plusDI-minusDI) / sum
	} else {
		return 0
	}
}
