// ✅ VALIDATION CCI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD
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
	fmt.Println("🔍 VALIDATION CCI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD")
	fmt.Println("=" + strings.Repeat("=", 65))

	// 1️⃣ CRÉER CLIENT BINANCE FUTURES (CRITÈRE 1)
	client := binance.NewFuturesClient()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2️⃣ RÉCUPÉRER 300 KLINES DEPUIS BINANCE FUTURES (CRITÈRE 5)
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
		fmt.Printf("✅ Klines récupérées: %d\n", len(klines))
	}

	// 3️⃣ CALCULER CCI ANCIENNE VERSION
	fmt.Println("\n📊 Calcul CCI Ancienne Version (période 20)...")
	
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
	
	cciOldValues := indicators.CCIFromKlines(indicatorsKlines, "hlc3", 20)

	// 4️⃣ CALCULER CCI TV STANDARD
	fmt.Println("📊 Calcul CCI TV Standard (période 20)...")
	
	// Préparer les données pour CCI TV Standard
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	
	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
	}
	
	cciTV := indicators.NewCCITVStandard(20)
	cciTVValues := cciTV.Calculate(high, low, close)

	if len(cciOldValues) == 0 || len(cciTVValues) == 0 {
		log.Fatalf("❌ Aucune valeur CCI calculée")
	}

	// 5️⃣ COMPARAISON DES VERSIONS
	fmt.Println("\n📊 COMPARAISON ANCIENNE vs TV STANDARD:")
	fmt.Println("=" + strings.Repeat("=", 65))
	
	lastCCIOld := cciOldValues[len(cciOldValues)-1]
	lastCCITV := cciTVValues[len(cciTVValues)-1]
	lastKline := klines[len(klines)-1]
	
	fmt.Printf("🕐 Dernière bougie: %s\n", lastKline.OpenTime.Format("15:04:05"))
	fmt.Printf("💰 Prix Close:      %.4f USDT\n", lastKline.Close)
	fmt.Printf("📊 CCI Ancienne:    %.4f\n", lastCCIOld)
	fmt.Printf("📊 CCI TV Standard: %.4f\n", lastCCITV)
	
	// Calculer la différence
	diff := math.Abs(lastCCIOld - lastCCITV)
	diffPercent := 0.0
	if lastCCIOld != 0 {
		diffPercent = (diff / math.Abs(lastCCIOld)) * 100
	}
	
	fmt.Printf("📊 Différence:      %.4f (%.2f%%)\n", diff, diffPercent)
	
	// 6️⃣ COMPARAISON SUR 10 DERNIÈRES VALEURS
	fmt.Println("\n📊 COMPARAISON 10 DERNIÈRES VALEURS:")
	fmt.Println("┌──────┬─────────────┬─────────────┬─────────────┬──────────┐")
	fmt.Println("│ Heure│ CCI Ancienne│ CCI TV Std  │ Différence  │ Diff %   │")
	fmt.Println("├──────┼─────────────┼─────────────┼─────────────┼──────────┤")
	
	startIdx := len(klines) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	
	totalDiff := 0.0
	validComparisons := 0
	maxDiff := 0.0
	
	for i := startIdx; i < len(klines); i++ {
		if i >= len(cciOldValues) || i >= len(cciTVValues) {
			continue
		}
		
		oldVal := cciOldValues[i]
		tvVal := cciTVValues[i]
		
		if math.IsNaN(oldVal) || math.IsNaN(tvVal) {
			continue
		}
		
		diff := math.Abs(oldVal - tvVal)
		diffPercent := 0.0
		if oldVal != 0 {
			diffPercent = (diff / math.Abs(oldVal)) * 100
		}
		
		totalDiff += diff
		validComparisons++
		if diff > maxDiff {
			maxDiff = diff
		}
		
		fmt.Printf("│ %s│ %11.4f │ %11.4f │ %11.4f │ %8.2f │\n",
			klines[i].OpenTime.Format("15:04"),
			oldVal, tvVal, diff, diffPercent)
	}
	
	fmt.Println("└──────┴─────────────┴─────────────┴─────────────┴──────────┘")
	
	// 7️⃣ STATISTIQUES DE COMPARAISON
	avgDiff := 0.0
	if validComparisons > 0 {
		avgDiff = totalDiff / float64(validComparisons)
	}
	
	fmt.Printf("\n📊 STATISTIQUES COMPARAISON:\n")
	fmt.Printf("✅ Comparaisons valides: %d/10\n", validComparisons)
	fmt.Printf("📊 Différence moyenne:   %.4f\n", avgDiff)
	fmt.Printf("📊 Différence maximale:  %.4f\n", maxDiff)
	
	// Évaluation de la conformité
	if avgDiff < 0.01 {
		fmt.Printf("✅ CONFORMITÉ EXCELLENTE (diff < 0.01)\n")
	} else if avgDiff < 0.1 {
		fmt.Printf("✅ CONFORMITÉ BONNE (diff < 0.1)\n")
	} else if avgDiff < 1.0 {
		fmt.Printf("⚠️  CONFORMITÉ MOYENNE (diff < 1.0)\n")
	} else {
		fmt.Printf("❌ CONFORMITÉ FAIBLE (diff >= 1.0)\n")
	}

	// 8️⃣ DÉTERMINER SIGNAL POUR LES DEUX VERSIONS
	fmt.Println("\n📊 SIGNAUX GÉNÉRÉS:")
	
	signalOld := getCCISignal(lastCCIOld)
	signalTV := getCCISignal(lastCCITV)
	
	fmt.Printf("🎯 Signal Ancienne:     %s\n", signalOld)
	fmt.Printf("🎯 Signal TV Standard:  %s\n", signalTV)
	
	if signalOld == signalTV {
		fmt.Printf("✅ SIGNAUX IDENTIQUES - Cohérence parfaite\n")
	} else {
		fmt.Printf("⚠️  SIGNAUX DIFFÉRENTS - Vérification requise\n")
	}

	// 9️⃣ CONCLUSION COMPARATIVE
	fmt.Println("\n🏁 VALIDATION CCI COMPARATIVE TERMINÉE:")
	fmt.Printf("🎯 CCI Ancienne:    %.4f - %s\n", lastCCIOld, signalOld)
	fmt.Printf("🎯 CCI TV Standard: %.4f - %s\n", lastCCITV, signalTV)
	fmt.Printf("📊 Différence:      %.4f (%.2f%%)\n", diff, diffPercent)
	
	if avgDiff < 0.1 {
		fmt.Println("✅ MIGRATION SÛRE - Différences négligeables")
	} else {
		fmt.Println("⚠️  MIGRATION À VÉRIFIER - Différences significatives")
	}

	fmt.Println("\n💡 Comparaison terminée avec succès !")
}

// getCCISignal retourne le signal CCI basé sur les zones
func getCCISignal(cciValue float64) string {
	if math.IsNaN(cciValue) {
		return "⚪ NaN"
	}
	
	switch {
	case cciValue > 200:
		return "🔴 SURACHAT EXTRÊME"
	case cciValue > 100:
		return "🟡 SURACHAT"
	case cciValue < -200:
		return "🟢 SURVENTE EXTRÊME"
	case cciValue < -100:
		return "🟡 SURVENTE"
	default:
		return "⚪ NEUTRE"
	}
}
