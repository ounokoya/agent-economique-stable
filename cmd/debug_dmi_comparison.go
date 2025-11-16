package main

import (
	"context"
	"fmt"
	"log"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/direction_dmi"
)

func main() {
	fmt.Println("🔍 DEBUG: COMPARAISON DMI TV Standard vs Direction DMI")
	fmt.Println("=" + fmt.Sprintf("%60s", "="))

	// Récupérer les mêmes données
	client := gateio.NewClient()
	ctx := context.Background()
	
	fmt.Println("📡 Récupération klines...")
	gateioKlines, err := client.GetKlines(ctx, "SOL_USDT", "5m", 100)
	if err != nil {
		log.Fatalf("❌ Erreur klines: %v", err)
	}

	// Trier chronologiquement (même logique que validateur)
	for i := 0; i < len(gateioKlines); i++ {
		for j := i + 1; j < len(gateioKlines); j++ {
			if gateioKlines[j].OpenTime.Before(gateioKlines[i].OpenTime) {
				gateioKlines[i], gateioKlines[j] = gateioKlines[j], gateioKlines[i]
			}
		}
	}

	// Convertir en signals.Kline pour Direction DMI
	klines := make([]signals.Kline, len(gateioKlines))
	for i, gk := range gateioKlines {
		klines[i] = signals.Kline{
			OpenTime: gk.OpenTime,
			Open:     gk.Open,
			High:     gk.High,
			Low:      gk.Low,
			Close:    gk.Close,
			Volume:   gk.Volume,
		}
	}

	fmt.Printf("✅ %d klines récupérées\n", len(klines))

	// === TEST 1: DMI TV Standard direct (comme validateur) ===
	fmt.Println("\n🔧 TEST 1: DMI TV Standard direct")
	
	highs1 := make([]float64, len(gateioKlines))
	lows1 := make([]float64, len(gateioKlines))
	closes1 := make([]float64, len(gateioKlines))

	for i, k := range gateioKlines {
		highs1[i] = k.High
		lows1[i] = k.Low
		closes1[i] = k.Close
	}

	dmiDirect := indicators.NewDMITVStandard(14)
	diPlusDirect, diMinusDirect, adxDirect := dmiDirect.Calculate(highs1, lows1, closes1)

	// === TEST 2: Direction DMI Generator ===
	fmt.Println("\n🔧 TEST 2: Direction DMI Generator")
	
	config := direction_dmi.Config{
		VWMAPeriod:          20,
		SlopePeriod:         6,
		KConfirmation:       2,
		UseDynamicThreshold: true,
		ATRPeriod:           8,
		ATRCoefficient:      0.25,
		DMIPeriod:           14,  // Même période que test direct
		DMISmooth:           14,
		GammaGapDI:          2.0,
		GammaGapDX:          2.0,
		WindowGammaValidate: 5,
		WindowMatching:      5,
		EnableEntryTrend:        true,
		EnableEntryCounterTrend: true,
		EnableExitTrend:         true,
		EnableExitCounterTrend:  true,
	}

	generatorConfig := signals.GeneratorConfig{}

	generator := direction_dmi.NewDirectionDMIGenerator(generatorConfig, config)
	
	// Calculer indicateurs dans le generator
	err = generator.CalculateIndicators(klines)
	if err != nil {
		log.Fatalf("❌ Erreur calcul generator: %v", err)
	}

	// Récupérer les valeurs DMI du generator
	diPlusGen, diMinusGen, adxGen := generator.GetDMIValues()

	// === COMPARAISON DES 10 DERNIÈRES VALEURS ===
	fmt.Println("\n📊 COMPARAISON DES 10 DERNIÈRES VALEURS:")
	fmt.Printf("%-8s %-10s %-12s %-12s %-12s %-12s %-12s %-12s\n", 
		"INDEX", "CLOSE", "DI+ Direct", "DI+ Gen", "DI- Direct", "DI- Gen", "ADX Direct", "ADX Gen")
	fmt.Println("-" + fmt.Sprintf("%96s", "-"))

	start := len(klines) - 10
	if start < 14 { start = 14 } // DMI needs at least 14 periods

	for i := start; i < len(klines); i++ {
		if i < len(diPlusDirect) && i < len(diPlusGen) {
			fmt.Printf("%-8d %-10.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f\n",
				i,
				klines[i].Close,
				diPlusDirect[i],
				diPlusGen[i],
				diMinusDirect[i], 
				diMinusGen[i],
				adxDirect[i],
				adxGen[i])
		}
	}

	// === ANALYSE DES DIFFÉRENCES ===
	fmt.Println("\n🚨 ANALYSE DES DIFFÉRENCES:")
	
	lastIdx := len(klines) - 1
	if lastIdx < len(diPlusDirect) && lastIdx < len(diPlusGen) {
		diffDIPlus := diPlusDirect[lastIdx] - diPlusGen[lastIdx]
		diffDIMinus := diMinusDirect[lastIdx] - diMinusGen[lastIdx]
		diffADX := adxDirect[lastIdx] - adxGen[lastIdx]

		fmt.Printf("Dernière valeur (index %d):\n", lastIdx)
		fmt.Printf("  DI+ : Direct=%.2f | Generator=%.2f | Diff=%.2f\n", 
			diPlusDirect[lastIdx], diPlusGen[lastIdx], diffDIPlus)
		fmt.Printf("  DI- : Direct=%.2f | Generator=%.2f | Diff=%.2f\n", 
			diMinusDirect[lastIdx], diMinusGen[lastIdx], diffDIMinus)
		fmt.Printf("  ADX : Direct=%.2f | Generator=%.2f | Diff=%.2f\n", 
			adxDirect[lastIdx], adxGen[lastIdx], diffADX)

		if abs(diffDIPlus) > 0.01 || abs(diffDIMinus) > 0.01 || abs(diffADX) > 0.01 {
			fmt.Println("\n❌ DIFFÉRENCES SIGNIFICATIVES DÉTECTÉES!")
		} else {
			fmt.Println("\n✅ Valeurs identiques ou très proches")
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
