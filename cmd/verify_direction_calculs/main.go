package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/signals"
)

// Intervalle pour chargement depuis JSON
type IntervalleJSON struct {
	Numero          int
	Type            signals.SignalType
	DateDebut       time.Time
	DateFin         time.Time
	PrixDebut       float64
	PrixFin         float64
	NbBougies       int
	VariationCaptee float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_calculs.go <dossier_out>")
		fmt.Println("Exemple: go run verify_calculs.go out/direction_demo_5m_vwma6_slope4_k3_atr4_coef0.50")
		os.Exit(1)
	}

	outDir := os.Args[1]
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("  VÉRIFICATION DES CALCULS - Direction Generator Demo")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Printf("\n📂 Dossier: %s\n", outDir)

	// Charger klines
	klinesFile := filepath.Join(outDir, "klines.json")
	klines, err := loadKlines(klinesFile)
	if err != nil {
		log.Fatalf("❌ Erreur chargement klines: %v", err)
	}
	fmt.Printf("✅ Klines chargées: %d\n", len(klines))

	// Charger intervalles
	intervallesFile := filepath.Join(outDir, "intervalles.json")
	intervalles, err := loadIntervalles(intervallesFile)
	if err != nil {
		log.Fatalf("❌ Erreur chargement intervalles: %v", err)
	}
	fmt.Printf("✅ Intervalles chargés: %d\n", len(intervalles))

	// Vérifier calculs
	fmt.Println("\n" + repeatStr("─", 70))
	fmt.Println("VÉRIFICATION INTERVALLE PAR INTERVALLE")
	fmt.Println(repeatStr("─", 70))

	erreurs := 0
	tolerance := 0.01 // tolérance de 0.01%

	for _, inter := range intervalles {
		// Recalculer variation
		variationRecalculee := 0.0
		if inter.PrixDebut != 0 {
			variationRecalculee = (inter.PrixFin - inter.PrixDebut) / inter.PrixDebut * 100
		}

		// Comparer
		diff := math.Abs(variationRecalculee - inter.VariationCaptee)
		status := "✅"
		if diff > tolerance {
			status = "❌"
			erreurs++
		}

		typeStr := "LONG"
		if inter.Type == signals.SignalTypeShort {
			typeStr = "SHORT"
		}

		fmt.Printf("%s #%-2d | %s | %.2f → %.2f | Variation: %+.2f%% (recalc: %+.2f%%, diff: %.4f%%)\n",
			status, inter.Numero, typeStr,
			inter.PrixDebut, inter.PrixFin,
			inter.VariationCaptee, variationRecalculee, diff)
	}

	// Recalculer totaux
	fmt.Println("\n" + repeatStr("─", 70))
	fmt.Println("VÉRIFICATION DES TOTAUX")
	fmt.Println(repeatStr("─", 70))

	countLong := 0
	countShort := 0
	variationLongDemo := 0.0
	variationShortDemo := 0.0
	variationLongRecalc := 0.0
	variationShortRecalc := 0.0

	for _, inter := range intervalles {
		variationRecalculee := (inter.PrixFin - inter.PrixDebut) / inter.PrixDebut * 100

		if inter.Type == signals.SignalTypeLong {
			countLong++
			variationLongDemo += inter.VariationCaptee
			variationLongRecalc += variationRecalculee
		} else {
			countShort++
			variationShortDemo += inter.VariationCaptee
			variationShortRecalc += variationRecalculee
		}
	}

	// Affichage LONG
	fmt.Println("\n📈 LONG:")
	fmt.Printf("   • Intervalles: %d\n", countLong)
	fmt.Printf("   • Variation démo:      %+.2f%%\n", variationLongDemo)
	fmt.Printf("   • Variation recalc:    %+.2f%%\n", variationLongRecalc)
	fmt.Printf("   • Différence:          %.4f%%\n", math.Abs(variationLongDemo-variationLongRecalc))

	// Affichage SHORT
	fmt.Println("\n📉 SHORT:")
	fmt.Printf("   • Intervalles: %d\n", countShort)
	fmt.Printf("   • Variation démo:      %+.2f%%\n", variationShortDemo)
	fmt.Printf("   • Variation recalc:    %+.2f%%\n", variationShortRecalc)
	fmt.Printf("   • Différence:          %.4f%%\n", math.Abs(variationShortDemo-variationShortRecalc))

	// Total capté
	totalCapteDemo := variationLongDemo - variationShortDemo
	totalCapteRecalc := variationLongRecalc - variationShortRecalc

	fmt.Println("\n💰 TOTAL CAPTÉ (bidirectionnel):")
	fmt.Printf("   • Démo:       %.2f%%\n", totalCapteDemo)
	fmt.Printf("   • Recalculé:  %.2f%%\n", totalCapteRecalc)
	fmt.Printf("   • Différence: %.4f%%\n", math.Abs(totalCapteDemo-totalCapteRecalc))

	// Verdict
	fmt.Println("\n" + repeatStr("═", 70))
	fmt.Println("VERDICT")
	fmt.Println(repeatStr("═", 70))

	if erreurs == 0 && math.Abs(totalCapteDemo-totalCapteRecalc) < tolerance {
		fmt.Println("✅ TOUS LES CALCULS SONT CORRECTS")
		fmt.Printf("   • Intervalles vérifiés: %d/%d\n", len(intervalles), len(intervalles))
		fmt.Printf("   • Erreurs détectées: 0\n")
		fmt.Printf("   • Précision: < %.2f%%\n", tolerance)
	} else {
		fmt.Println("❌ ERREURS DÉTECTÉES")
		fmt.Printf("   • Intervalles avec erreur: %d/%d\n", erreurs, len(intervalles))
		fmt.Printf("   • Écart total capté: %.4f%%\n", math.Abs(totalCapteDemo-totalCapteRecalc))
	}

	fmt.Println(repeatStr("═", 70))
}

func loadKlines(filename string) ([]signals.Kline, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var klines []signals.Kline
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&klines); err != nil {
		return nil, err
	}

	return klines, nil
}

func loadIntervalles(filename string) ([]IntervalleJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var intervalles []IntervalleJSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&intervalles); err != nil {
		return nil, err
	}

	return intervalles, nil
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
