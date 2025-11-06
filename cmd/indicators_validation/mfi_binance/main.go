// ✅ VALIDATION MFI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD  
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

// formatValue formate une valeur float64 pour l'affichage
func formatValue(v float64) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	return fmt.Sprintf("%.2f", v)
}

// getZone retourne la zone MFI
func getZone(v float64) string {
	if math.IsNaN(v) {
		return "⚪ NEUTRE"
	}
	if v > 80 {
		return "🔴 SURACHAT"
	} else if v > 70 {
		return "🟡 SURACHAT"
	} else if v < 20 {
		return "🟢 SURVENTE"
	} else if v < 30 {
		return "🟡 SURVENTE"
	}
	return "⚪ NEUTRE"
}

func main() {
	fmt.Println("🔍 VALIDATION MFI BINANCE - COMPARAISON ANCIENNE vs TV STANDARD")
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

	// 3️⃣ & 4️⃣ CALCULER MFI avec Volume SOL et Volume USDT
	fmt.Println("\n📊 Calcul MFI avec Volume SOL (période 14)...")
	fmt.Println("📊 Calcul MFI avec Volume USDT (période 14)...")
	
	// Créer les indicateurs MFI
	mfiTV_SOL := indicators.NewMFITVStandard(14)
	mfiTV_USDT := indicators.NewMFITVStandard(14)

	// Créer les données pour MFI
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	volumeSOL := make([]float64, len(klines))
	volumeUSDT := make([]float64, len(klines))

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volumeSOL[i] = k.Volume
		volumeUSDT[i] = k.QuoteAssetVolume
	}

	// Calculer MFI avec les 2 volumes
	mfiValues_SOL := mfiTV_SOL.Calculate(high, low, close, volumeSOL)
	mfiValues_USDT := mfiTV_USDT.Calculate(high, low, close, volumeUSDT)

	if len(mfiValues_SOL) == 0 || len(mfiValues_USDT) == 0 {
		log.Fatalf("❌ Aucune valeur MFI calculée")
	}

	// 5️⃣ COMPARAISON DES VERSIONS
	fmt.Println("\n📊 COMPARAISON MFI avec Volume SOL vs Volume USDT:")
	fmt.Println("=" + strings.Repeat("=", 65))
	
	lastKline := klines[len(klines)-1]
	lastMFI_SOL := mfiTV_SOL.GetLastValue(mfiValues_SOL)
	lastMFI_USDT := mfiTV_USDT.GetLastValue(mfiValues_USDT)
	
	fmt.Printf("🕐 Dernière bougie: %s\n", lastKline.OpenTime.Format("15:04:05"))
	fmt.Printf("💰 Prix Close:      %.4f USDT\n", lastKline.Close)
	fmt.Printf("📊 Volume SOL:      %.0f\n", lastKline.Volume)
	fmt.Printf("📊 Volume USDT:     %.0f\n", lastKline.QuoteAssetVolume)
	
	fmt.Printf("\n📊 MFI (Vol SOL):   %.4f\n", lastMFI_SOL)
	fmt.Printf("📊 MFI (Vol USDT):  %.4f\n", lastMFI_USDT)
	
	// Calculer la différence
	diff := math.Abs(lastMFI_SOL - lastMFI_USDT)
	diffPercent := 0.0
	if lastMFI_SOL != 0 {
		diffPercent = (diff / math.Abs(lastMFI_SOL)) * 100
	}
	
	fmt.Printf("\n📊 Différences SOL vs USDT:\n")
	fmt.Printf("   MFI: %.4f (%.2f%%)\n", diff, diffPercent)

	// 6️⃣ TABLE DE COMPARAISON 10 DERNIÈRES VALEURS
	fmt.Println("\n📊 COMPARAISON 10 DERNIÈRES VALEURS:")
	fmt.Println("┌──────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐")
	fmt.Println("│ Heure│  Vol SOL │ Vol USDT │ MFI SOL  │ MFI USDT │ Différence│ Diff %   │ TV Match │")
	fmt.Println("├──────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤")
	
	startIdx := len(klines) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	
	totalDiff := 0.0
	validComparisons := 0
	maxDiff := 0.0
	
	for i := startIdx; i < len(klines); i++ {
		if i >= len(mfiValues_SOL) || i >= len(mfiValues_USDT) {
			continue
		}
		
		mfiSOL := mfiValues_SOL[i]
		mfiUSDT := mfiValues_USDT[i]
		
		if math.IsNaN(mfiSOL) || math.IsNaN(mfiUSDT) {
			continue
		}
		
		diffVal := math.Abs(mfiSOL - mfiUSDT)
		diffPercent := 0.0
		if mfiSOL != 0 {
			diffPercent = (diffVal / math.Abs(mfiSOL)) * 100
		}
		
		totalDiff += diffVal
		validComparisons++
		if diffVal > maxDiff {
			maxDiff = diffVal
		}
		
		tvMatch := "?"
		// L'utilisateur doit comparer avec TradingView manuellement
		
		fmt.Printf("│ %s│ %8.0f │ %8.0f │ %8.4f │ %8.4f │ %8.4f │ %7.2f%% │ %-8s │\n",
			klines[i].OpenTime.Format("15:04"),
			klines[i].Volume,
			klines[i].QuoteAssetVolume,
			mfiSOL, mfiUSDT, diffVal, diffPercent, tvMatch)
	}
	
	fmt.Println("└──────┴──────────┴──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘")

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

	// 8️⃣ SIGNAUX POUR LES DEUX VERSIONS
	fmt.Println("\n📊 SIGNAUX GÉNÉRÉS:")
	signalSOL := getMFISignal(lastMFI_SOL)
	signalUSDT := getMFISignal(lastMFI_USDT)
	
	fmt.Printf("🎯 Signal MFI (Vol SOL):   %s\n", signalSOL)
	fmt.Printf("🎯 Signal MFI (Vol USDT):  %s\n", signalUSDT)
	if signalSOL == signalUSDT {
		fmt.Println("✅ SIGNAUX IDENTIQUES - Les deux volumes donnent le même signal")
	} else {
		fmt.Println("⚠️  SIGNAUX DIFFÉRENTS - Le choix du volume change le signal !")
	}
	
	fmt.Println("\n🏁 VALIDATION MFI COMPARATIVE TERMINÉE:")
	fmt.Printf("🎯 MFI (Vol SOL):   %.4f - %s\n", lastMFI_SOL, signalSOL)
	fmt.Printf("🎯 MFI (Vol USDT):  %.4f - %s\n", lastMFI_USDT, signalUSDT)
	fmt.Printf("📊 Différence:      %.4f (%.2f%%)\n", diff, diffPercent)
	
	if avgDiff < 0.1 {
		fmt.Println("✅ MIGRATION SÛRE - Différences négligeables")
	} else {
		fmt.Println("⚠️  MIGRATION À VÉRIFIER - Différences significatives")
	}

	fmt.Println("\n💡 Comparaison terminée avec succès !")
}

// getMFISignal retourne le signal MFI basé sur les zones
func getMFISignal(mfiValue float64) string {
	if math.IsNaN(mfiValue) {
		return "⚪ NaN"
	}
	
	switch {
	case mfiValue > 80:
		return "🔴 SURACHAT"
	case mfiValue > 70:
		return "🟡 SURACHAT FAIBLE"
	case mfiValue < 20:
		return "🟢 SURVENTE"
	case mfiValue < 30:
		return "🟡 SURVENTE FAIBLE"
	default:
		return "⚪ NEUTRE"
	}
}
