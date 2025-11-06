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

// Application du CCI TV Standard à Gate.io
func main() {
	fmt.Println("🎯 CCI TV STANDARD - APPLICATION GATE.IO")
	fmt.Println("=" + strings.Repeat("=", 45))

	// Créer le client Gate.io (pas besoin de config complexe)
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

	// Initialiser et calculer le CCI TV Standard
	fmt.Println("\n🔧 Calcul CCI TV Standard sur données Gate.io...")
	cciIndicator := indicators.NewCCITVStandard(20)
	cciValues := cciIndicator.Calculate(high, low, close)

	if cciValues == nil {
		fmt.Println("❌ Erreur calcul CCI")
		return
	}

	// Afficher les 15 dernières valeurs
	fmt.Println("\n📊 CCI TV STANDARD - GATE.IO (15 dernières valeurs):")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("%-12s %-10s %-12s %-12s %-10s\n", 
		"TIME", "CLOSE", "CCI_VALUE", "ZONE", "SIGNAL")
	fmt.Println(strings.Repeat("-", 70))

	startIdx := len(klines) - 15
	for i := startIdx; i < len(klines); i++ {
		k := klines[i]
		cciVal := cciValues[i]
		zone := cciIndicator.GetZone(cciVal)
		signal := getSignal(cciVal)
		
		cciStr := formatCCI(cciVal)
		
		fmt.Printf("%-12s %-10.2f %-12s %-12s %-10s\n",
			k.OpenTime.Format("15:04"), k.Close, cciStr, zone, signal)
	}

	fmt.Println(strings.Repeat("=", 70))

	// Analyse des dernières valeurs
	fmt.Println("\n📈 ANALYSE CCI GATE.IO:")
	fmt.Println(strings.Repeat("=", 30))

	lastCCI := cciIndicator.GetLastValue(cciValues)
	lastZone := cciIndicator.GetZone(lastCCI)
	
	fmt.Printf("Dernière bougie (%s):\n", klines[len(klines)-1].OpenTime.Format("15:04"))
	fmt.Printf("  Prix: %.2f\n", klines[len(klines)-1].Close)
	fmt.Printf("  CCI: %.2f\n", lastCCI)
	fmt.Printf("  Zone: %s\n", lastZone)
	
	// Signaux actuels
	if cciIndicator.IsOverbought(lastCCI) {
		fmt.Printf("  ⚠️ SIGNAL: SURACHAT - Potentielle correction baissière\n")
	} else if cciIndicator.IsOversold(lastCCI) {
		fmt.Printf("  📈 SIGNAL: SURVENTE - Potentiel rebond haussier\n")
	} else if cciIndicator.IsBullish(lastCCI) {
		fmt.Printf("  📊 SIGNAL: HAUSSIER - Tendance haussière en cours\n")
	} else if cciIndicator.IsBearish(lastCCI) {
		fmt.Printf("  📉 SIGNAL: BAISSIER - Tendance baissière en cours\n")
	} else {
		fmt.Printf("  ➡️ SIGNAL: NEUTRE - Pas de tendance claire\n")
	}

	// Statistiques sur les 15 dernières valeurs
	fmt.Println("\n📊 STATISTIQUES (15 dernières valeurs):")
	last15Values := cciValues[startIdx:]
	
	validCount := 0
	overboughtCount := 0
	oversoldCount := 0
	bullishCount := 0
	bearishCount := 0
	
	for _, v := range last15Values {
		if !math.IsNaN(v) {
			validCount++
			if cciIndicator.IsOverbought(v) {
				overboughtCount++
			}
			if cciIndicator.IsOversold(v) {
				oversoldCount++
			}
			if cciIndicator.IsBullish(v) {
				bullishCount++
			}
			if cciIndicator.IsBearish(v) {
				bearishCount++
			}
		}
	}
	
	fmt.Printf("Valeurs valides: %d/15\n", validCount)
	fmt.Printf("Zone Surachat: %d fois (%.1f%%)\n", overboughtCount, float64(overboughtCount)/float64(validCount)*100)
	fmt.Printf("Zone Survente: %d fois (%.1f%%)\n", oversoldCount, float64(oversoldCount)/float64(validCount)*100)
	fmt.Printf("Tendance Haussière: %d fois (%.1f%%)\n", bullishCount, float64(bullishCount)/float64(validCount)*100)
	fmt.Printf("Tendance Baissière: %d fois (%.1f%%)\n", bearishCount, float64(bearishCount)/float64(validCount)*100)

	// Recommandations de trading
	fmt.Println("\n💡 RECOMMANDATIONS DE TRADING:")
	fmt.Println(strings.Repeat("=", 35))
	
	if overboughtCount > validCount/2 {
		fmt.Println("⚠️ Marché en surachat fréquent")
		fmt.Println("   → Attendre une correction avant d'acheter")
		fmt.Println("   → Considérer prendre profits sur positions longues")
	} else if oversoldCount > validCount/4 {
		fmt.Println("📈 Opportunités d'achat en survente")
		fmt.Println("   → Chercher points d'entrée longs")
		fmt.Println("   → Surveiller les retournements haussiers")
	}
	
	if bullishCount > bearishCount {
		fmt.Println("📊 Tendance générale haussière")
		fmt.Println("   → Favoriser les positions longues")
		fmt.Println("   → Utiliser les replis pour acheter")
	} else {
		fmt.Println("📉 Tendance générale baissière")
		fmt.Println("   → Favoriser les positions courtes")
		fmt.Println("   → Utiliser les rebonds pour vendre")
	}

	fmt.Println("\n✅ CCI TV STANDARD appliqué avec succès à Gate.io!")
}

// Fonctions utilitaires
func formatCCI(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	return fmt.Sprintf("%.2f", value)
}

func getSignal(value float64) string {
	if math.IsNaN(value) {
		return "INCONNU"
	}
	
	if value > 100 {
		return "SURACHAT"
	} else if value < -100 {
		return "SURVENTE"
	} else if value > 50 {
		return "HAUSSIER+"
	} else if value > 0 {
		return "HAUSSIER"
	} else if value > -50 {
		return "BAISSIER"
	} else {
		return "BAISSIER-"
	}
}
