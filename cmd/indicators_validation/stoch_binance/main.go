// ✅ VALIDATION STOCHASTIC BINANCE - COMPARAISON ANCIENNE vs TV STANDARD
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
	fmt.Println("🔍 VALIDATION STOCHASTIC BINANCE - COMPARAISON ANCIENNE vs TV STANDARD")
	fmt.Println("=" + strings.Repeat("=", 65))

	// 1️⃣ CRÉER CLIENT BINANCE FUTURES
	client := binance.NewFuturesClient()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2️⃣ RÉCUPÉRER 300 KLINES DEPUIS BINANCE FUTURES
	fmt.Println("📡 Récupération des 300 dernières klines depuis Binance Futures...")
	futuresKlines, err := client.GetKlines(ctx, "SOLUSDT", "5m", 300)
	if err != nil {
		log.Fatalf("❌ Erreur récupération klines: %v", err)
	}

	// Convertir en format standard
	klines := client.ConvertToStandardKline(futuresKlines)

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

	// 3️⃣ CALCULER STOCHASTIC ANCIENNE VERSION
	fmt.Println("\n📊 Calcul Stochastic Ancienne Version (%K=14, %D=3)...")
	
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
	
	stochOldK, stochOldD := indicators.StochasticFromKlines(indicatorsKlines, 14, 3, 3)

	// 4️⃣ CALCULER STOCHASTIC TV STANDARD
	fmt.Println("📊 Calcul Stochastic TV Standard (%K=14, %D=3)...")
	
	// Préparer les données pour Stochastic TV Standard
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	
	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
	}
	
	stochTV := indicators.NewStochTVStandard(14, 3, 3)
	stochTVK, stochTVD := stochTV.Calculate(high, low, close)

	if len(stochOldK) == 0 || len(stochTVK) == 0 {
		log.Fatalf("❌ Aucune valeur Stochastic calculée")
	}

	// 5️⃣ COMPARAISON DES VERSIONS
	fmt.Println("\n📊 COMPARAISON ANCIENNE vs TV STANDARD:")
	fmt.Println("=" + strings.Repeat("=", 65))
	
	lastKline := klines[len(klines)-1]
	lastKOld := stochOldK[len(stochOldK)-1]
	lastDOld := stochOldD[len(stochOldD)-1]
	lastKTV := stochTVK[len(stochTVK)-1]
	lastDTV := stochTVD[len(stochTVD)-1]
	
	fmt.Printf("🕐 Dernière bougie: %s\n", lastKline.OpenTime.Format("15:04:05"))
	fmt.Printf("💰 Prix Close:      %.4f USDT\n", lastKline.Close)
	
	// Affichage avec gestion des NaN
	fmt.Printf("\n📊 %K Ancienne:     ")
	if math.IsNaN(lastKOld) {
		fmt.Printf("NaN\n")
	} else {
		fmt.Printf("%.4f\n", lastKOld)
	}
	
	fmt.Printf("📊 %K TV Standard:  ")
	if math.IsNaN(lastKTV) {
		fmt.Printf("NaN\n")
	} else {
		fmt.Printf("%.4f\n", lastKTV)
	}
	
	fmt.Printf("\n📊 %D Ancienne:     ")
	if math.IsNaN(lastDOld) {
		fmt.Printf("NaN\n")
	} else {
		fmt.Printf("%.4f\n", lastDOld)
	}
	
	fmt.Printf("📊 %D TV Standard:  ")
	if math.IsNaN(lastDTV) {
		fmt.Printf("NaN\n")
	} else {
		fmt.Printf("%.4f\n", lastDTV)
	}
	
	// Calculer les différences
	diffK := 0.0
	diffD := 0.0
	if !math.IsNaN(lastKOld) && !math.IsNaN(lastKTV) {
		diffK = math.Abs(lastKOld - lastKTV)
	}
	if !math.IsNaN(lastDOld) && !math.IsNaN(lastDTV) {
		diffD = math.Abs(lastDOld - lastDTV)
	}
	
	fmt.Printf("\n📊 Différences:\n")
	fmt.Printf("   %K: %.4f\n", diffK)
	fmt.Printf("   %D: %.4f\n", diffD)

	// 6️⃣ TABLE DE COMPARAISON 10 DERNIÈRES VALEURS
	fmt.Println("\n📊 COMPARAISON 10 DERNIÈRES VALEURS:")
	fmt.Println("┌──────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐")
	fmt.Println("│ Heure│  %K Old  │  %K TV   │  %D Old  │  %D TV   │ DiffK    │ DiffD    │")
	fmt.Println("├──────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤")
	
	startIdx := len(klines) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	
	totalDiffK := 0.0
	totalDiffD := 0.0
	validComparisons := 0
	maxDiffK := 0.0
	maxDiffD := 0.0
	
	for i := startIdx; i < len(klines); i++ {
		if i >= len(stochOldK) || i >= len(stochTVK) {
			continue
		}
		
		kOldVal := stochOldK[i]
		kTVVal := stochTVK[i]
		dOldVal := stochOldD[i]
		dTVVal := stochTVD[i]
		
		if math.IsNaN(kOldVal) || math.IsNaN(kTVVal) ||
		   math.IsNaN(dOldVal) || math.IsNaN(dTVVal) {
			continue
		}
		
		diffKVal := math.Abs(kOldVal - kTVVal)
		diffDVal := math.Abs(dOldVal - dTVVal)
		
		totalDiffK += diffKVal
		totalDiffD += diffDVal
		validComparisons++
		
		if diffKVal > maxDiffK {
			maxDiffK = diffKVal
		}
		if diffDVal > maxDiffD {
			maxDiffD = diffDVal
		}
		
		fmt.Printf("│ %s│ %8.4f │ %8.4f │ %8.4f │ %8.4f │ %8.4f │ %8.4f │\n",
			klines[i].OpenTime.Format("15:04"),
			kOldVal, kTVVal, dOldVal, dTVVal, diffKVal, diffDVal)
	}
	
	fmt.Println("└──────┴──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘")
	
	// 7️⃣ STATISTIQUES DE COMPARAISON
	avgDiffK := 0.0
	avgDiffD := 0.0
	
	if validComparisons > 0 {
		avgDiffK = totalDiffK / float64(validComparisons)
		avgDiffD = totalDiffD / float64(validComparisons)
	}
	
	fmt.Printf("\n📊 STATISTIQUES COMPARAISON:\n")
	fmt.Printf("✅ Comparaisons valides: %d/10\n", validComparisons)
	fmt.Printf("📊 Différence moyenne %K: %.4f\n", avgDiffK)
	fmt.Printf("📊 Différence moyenne %D: %.4f\n", avgDiffD)
	fmt.Printf("📊 Différence maximale %K: %.4f\n", maxDiffK)
	fmt.Printf("📊 Différence maximale %D: %.4f\n", maxDiffD)
	
	// Évaluation globale
	avgGlobalDiff := (avgDiffK + avgDiffD) / 2.0
	
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
	
	signalOld := getStochSignal(lastKOld, lastDOld)
	signalTV := getStochSignal(lastKTV, lastDTV)
	
	fmt.Printf("🎯 Signal Ancienne:     %s\n", signalOld)
	fmt.Printf("🎯 Signal TV Standard:  %s\n", signalTV)
	
	if signalOld == signalTV {
		fmt.Printf("✅ SIGNAUX IDENTIQUES - Cohérence parfaite\n")
	} else {
		fmt.Printf("⚠️  SIGNAUX DIFFÉRENTS - Vérification requise\n")
	}

	// 9️⃣ CONCLUSION COMPARATIVE
	fmt.Println("\n🏁 VALIDATION STOCHASTIC COMPARATIVE TERMINÉE:")
	fmt.Printf("🎯 Stoch Ancienne:    %%K:%.4f %%D:%.4f - %s\n", 
		lastKOld, lastDOld, signalOld)
	fmt.Printf("🎯 Stoch TV Standard: %%K:%.4f %%D:%.4f - %s\n", 
		lastKTV, lastDTV, signalTV)
	fmt.Printf("📊 Différences:      %%K:%.4f %%D:%.4f\n", 
		diffK, diffD)
	
	if avgGlobalDiff < 0.5 {
		fmt.Println("✅ MIGRATION SÛRE - Différences négligeables")
	} else {
		fmt.Println("⚠️  MIGRATION À VÉRIFIER - Différences significatives")
	}

	fmt.Println("\n💡 Comparaison terminée avec succès !")
}

// getStochSignal retourne le signal Stochastic basé sur les valeurs K et D
func getStochSignal(k, d float64) string {
	if math.IsNaN(k) || math.IsNaN(d) {
		return "⚪ NaN"
	}
	
	switch {
	case k > 80 && d > 80:
		if k < d {
			return "🔴 SURACHAT + croisement %K sous %D"
		} else {
			return "🔴 SURACHAT"
		}
	case k < 20 && d < 20:
		if k > d {
			return "🟢 SURVENTE + croisement %K sur %D"
		} else {
			return "🟢 SURVENTE"
		}
	case k > d:
		return "🟡 MOMENTUM HAUSSIER"
	case k < d:
		return "🟡 MOMENTUM BAISSIER"
	default:
		return "⚪ NEUTRE"
	}
}
