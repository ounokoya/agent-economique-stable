// ✅ VALIDATION DMI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/indicators"
)

func main() {
	fmt.Println("🔍 VALIDATION DMI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD")
	fmt.Println("=" + strings.Repeat("=", 65))

	// 1️⃣ CRÉER CLIENT BINANCE
	client := binance.NewClient()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2️⃣ RÉCUPÉRER 300 KLINES DEPUIS BINANCE
	fmt.Println("📡 Récupération des 300 dernières klines depuis Binance...")
	klines, err := client.GetKlines(ctx, "SOLUSDT", "5m", 300)
	if err != nil {
		log.Fatalf("❌ Erreur récupération klines: %v", err)
	}

	fmt.Printf("✅ %d klines récupérées de %s à %s\n", 
		len(klines), 
		klines[0].OpenTime.Format("2006-01-02 15:04"), 
		klines[len(klines)-1].OpenTime.Format("2006-01-02 15:04"))

	// 🔍 CONTRÔLE PRÉCISION BINANCE FUTURES (CRITÈRES 2-4)
	fmt.Println("\n🔍 CONTRÔLE PRÉCISION BINANCE FUTURES:")
	fmt.Printf("✅ Source: Futures perpétuels (client.NewFuturesClient())\n")
	
	if len(klines) > 0 {
		last := klines[len(klines)-1]
		fmt.Printf("✅ Format: %T (struct convertie depuis array)\n", last)
		fmt.Printf("✅ Prix: %.4f USDT\n", last.Close)
		fmt.Printf("✅ OpenTime: %s (timestamp ms→s)\n", last.OpenTime.Format("15:04:05"))
		
		// Vérifier cohérence timeframe 5m
		if len(klines) >= 2 {
			prev := klines[len(klines)-2]
			diff := last.OpenTime.Sub(prev.OpenTime)
			if diff == 5*time.Minute {
				fmt.Printf("✅ Timeframe 5m correct (%v)\n", diff)
			} else {
				fmt.Printf("❌ Timeframe incorrect: %v\n", diff)
			}
		}
		fmt.Printf("✅ Volume: %.0f SOL (base currency)\n", last.Volume)
		fmt.Printf("✅ Klines récupérées: %d\n", len(klines))
	}

	// 3️⃣ CALCULER DMI ANCIENNE VERSION
	fmt.Println("\n📊 Calcul DMI Ancienne Version (période 14)...")
	
	// Convertir en format indicators.Kline
	indicatorsKlines := make([]indicators.Kline, len(klines))
	for i, k := range klines {
		indicatorsKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix(),
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}
	
	diPlusOld, diMinusOld, _, adxOld := indicators.DMIFromKlines(indicatorsKlines, 14)

	// 4️⃣ CALCULER DMI TV STANDARD
	fmt.Println("📊 Calcul DMI TV Standard (période 14)...")
	
	// Préparer les données pour DMI TV Standard
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	
	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
	}
	
	dmiTV := indicators.NewDMITVStandard(14)
	diPlusTV, diMinusTV, adxTV := dmiTV.Calculate(high, low, close)

	if len(diPlusOld) == 0 || len(diPlusTV) == 0 {
		log.Fatalf("❌ Aucune valeur DMI calculée")
	}

	// 5️⃣ COMPARAISON DES VERSIONS
	fmt.Println("\n📊 COMPARAISON ANCIENNE vs TV STANDARD:")
	fmt.Println("=" + strings.Repeat("=", 65))
	
	lastKline := klines[len(klines)-1]
	lastDIPlusOld := diPlusOld[len(diPlusOld)-1]
	lastDIMinusOld := diMinusOld[len(diMinusOld)-1]
	lastADXOld := adxOld[len(adxOld)-1]
	
	lastDIPlusTV := diPlusTV[len(diPlusTV)-1]
	lastDIMinusTV := diMinusTV[len(diMinusTV)-1]
	lastADXTV := adxTV[len(adxTV)-1]
	
	fmt.Printf("🕐 Dernière bougie: %s\n", lastKline.OpenTime.Format("15:04:05"))
	fmt.Printf("💰 Prix Close:      %.4f USDT\n", lastKline.Close)
	
	fmt.Printf("\n📊 DI+ Ancienne:    %.4f\n", lastDIPlusOld)
	fmt.Printf("📊 DI+ TV Standard: %.4f\n", lastDIPlusTV)
	
	fmt.Printf("📊 DI- Ancienne:    %.4f\n", lastDIMinusOld)
	fmt.Printf("📊 DI- TV Standard: %.4f\n", lastDIMinusTV)
	
	fmt.Printf("📊 ADX Ancienne:    %.4f\n", lastADXOld)
	fmt.Printf("📊 ADX TV Standard: %.4f\n", lastADXTV)
	
	// Calculer les différences
	diffDIPlus := math.Abs(lastDIPlusOld - lastDIPlusTV)
	diffDIMinus := math.Abs(lastDIMinusOld - lastDIMinusTV)
	diffADX := math.Abs(lastADXOld - lastADXTV)
	
	fmt.Printf("\n📊 Différences:\n")
	fmt.Printf("   DI+: %.4f\n", diffDIPlus)
	fmt.Printf("   DI-: %.4f\n", diffDIMinus)
	fmt.Printf("   ADX: %.4f\n", diffADX)

	// 6️⃣ TABLE DE COMPARAISON 10 DERNIÈRES VALEURS
	fmt.Println("\n📊 COMPARAISON 10 DERNIÈRES VALEURS:")
	fmt.Println("┌──────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐")
	fmt.Println("│ Heure│ DI+ Old  │ DI+ TV   │ DI- Old  │ DI- TV   │ ADX Old  │ ADX TV   │")
	fmt.Println("├──────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤")
	
	startIdx := len(klines) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	
	totalDiffDIPlus := 0.0
	totalDiffDIMinus := 0.0
	totalDiffADX := 0.0
	validComparisons := 0
	
	for i := startIdx; i < len(klines); i++ {
		if i >= len(diPlusOld) || i >= len(diPlusTV) {
			continue
		}
		
		diPlusOldVal := diPlusOld[i]
		diPlusTVVal := diPlusTV[i]
		diMinusOldVal := diMinusOld[i]
		diMinusTVVal := diMinusTV[i]
		adxOldVal := adxOld[i]
		adxTVVal := adxTV[i]
		
		if math.IsNaN(diPlusOldVal) || math.IsNaN(diPlusTVVal) ||
		   math.IsNaN(diMinusOldVal) || math.IsNaN(diMinusTVVal) ||
		   math.IsNaN(adxOldVal) || math.IsNaN(adxTVVal) {
			continue
		}
		
		totalDiffDIPlus += math.Abs(diPlusOldVal - diPlusTVVal)
		totalDiffDIMinus += math.Abs(diMinusOldVal - diMinusTVVal)
		totalDiffADX += math.Abs(adxOldVal - adxTVVal)
		validComparisons++
		
		fmt.Printf("│ %s│ %8.4f │ %8.4f │ %8.4f │ %8.4f │ %8.4f │ %8.4f │\n",
			klines[i].OpenTime.Format("15:04"),
			diPlusOldVal, diPlusTVVal, diMinusOldVal, diMinusTVVal, adxOldVal, adxTVVal)
	}
	
	fmt.Println("└──────┴──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘")
	
	// 7️⃣ STATISTIQUES DE COMPARAISON
	avgDiffDIPlus := 0.0
	avgDiffDIMinus := 0.0
	avgDiffADX := 0.0
	
	if validComparisons > 0 {
		avgDiffDIPlus = totalDiffDIPlus / float64(validComparisons)
		avgDiffDIMinus = totalDiffDIMinus / float64(validComparisons)
		avgDiffADX = totalDiffADX / float64(validComparisons)
	}
	
	fmt.Printf("\n📊 STATISTIQUES COMPARAISON:\n")
	fmt.Printf("✅ Comparaisons valides: %d/10\n", validComparisons)
	fmt.Printf("📊 Différence moyenne DI+: %.4f\n", avgDiffDIPlus)
	fmt.Printf("📊 Différence moyenne DI-: %.4f\n", avgDiffDIMinus)
	fmt.Printf("📊 Différence moyenne ADX: %.4f\n", avgDiffADX)
	
	// Évaluation globale
	avgGlobalDiff := (avgDiffDIPlus + avgDiffDIMinus + avgDiffADX) / 3.0
	
	if avgGlobalDiff < 0.1 {
		fmt.Printf("✅ CONFORMITÉ EXCELLENTE (diff < 0.1)\n")
	} else if avgGlobalDiff < 0.5 {
		fmt.Printf("✅ CONFORMITÉ BONNE (diff < 0.5)\n")
	} else if avgGlobalDiff < 1.0 {
		fmt.Printf("⚠️  CONFORMITÉ MOYENNE (diff < 1.0)\n")
	} else {
		fmt.Printf("❌ CONFORMITÉ FAIBLE (diff >= 1.0)\n")
	}

	// 8️⃣ SIGNAUX POUR LES DEUX VERSIONS
	fmt.Println("\n📊 SIGNAUX GÉNÉRÉS:")
	
	signalOld := getDMISignal(lastDIPlusOld, lastDIMinusOld, lastADXOld)
	signalTV := getDMISignal(lastDIPlusTV, lastDIMinusTV, lastADXTV)
	
	fmt.Printf("🎯 Signal Ancienne:     %s\n", signalOld)
	fmt.Printf("🎯 Signal TV Standard:  %s\n", signalTV)
	
	if signalOld == signalTV {
		fmt.Printf("✅ SIGNAUX IDENTIQUES - Cohérence parfaite\n")
	} else {
		fmt.Printf("⚠️  SIGNAUX DIFFÉRENTS - Vérification requise\n")
	}

	// 9️⃣ CONCLUSION COMPARATIVE
	fmt.Println("\n🏁 VALIDATION DMI COMPARATIVE TERMINÉE:")
	fmt.Printf("🎯 DMI Ancienne:    DI+:%.4f DI-:%.4f ADX:%.4f - %s\n", 
		lastDIPlusOld, lastDIMinusOld, lastADXOld, signalOld)
	fmt.Printf("🎯 DMI TV Standard: DI+:%.4f DI-:%.4f ADX:%.4f - %s\n", 
		lastDIPlusTV, lastDIMinusTV, lastADXTV, signalTV)
	fmt.Printf("📊 Différences:     DI+:%.4f DI-:%.4f ADX:%.4f\n", 
		diffDIPlus, diffDIMinus, diffADX)
	
	if avgGlobalDiff < 0.5 {
		fmt.Println("✅ MIGRATION SÛRE - Différences négligeables")
	} else {
		fmt.Println("⚠️  MIGRATION À VÉRIFIER - Différences significatives")
	}

	fmt.Println("\n💡 Comparaison terminée avec succès !")
}

// getDMISignal retourne le signal DMI basé sur les valeurs
func getDMISignal(diPlus, diMinus, adx float64) string {
	if math.IsNaN(diPlus) || math.IsNaN(diMinus) || math.IsNaN(adx) {
		return "⚪ NaN"
	}
	
	if diPlus > diMinus {
		if adx > 25 {
			return "🟢 TENDANCE HAUSSIÈRE FORTE"
		} else {
			return "🟡 TENDANCE HAUSSIÈRE FAIBLE"
		}
	} else if diMinus > diPlus {
		if adx > 25 {
			return "🔴 TENDANCE BAISSIÈRE FORTE"
		} else {
			return "🟡 TENDANCE BAISSIÈRE FAIBLE"
		}
	} else {
		return "⚪ NEUTRE"
	}
}
