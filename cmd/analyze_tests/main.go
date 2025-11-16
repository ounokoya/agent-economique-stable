package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Intervalle struct {
	Numero          int       `json:"Numero"`
	Type            string    `json:"Type"`
	DateDebut       time.Time `json:"DateDebut"`
	DateFin         time.Time `json:"DateFin"`
	PrixDebut       float64   `json:"PrixDebut"`
	PrixFin         float64   `json:"PrixFin"`
	NbBougies       int       `json:"NbBougies"`
	VariationCaptee float64   `json:"VariationCaptee"`
}

type TestResult struct {
	Path            string
	Timeframe       string
	VWMAPeriod      int
	SlopePeriod     int
	ATRPeriod       int
	ATRCoeff        float64
	TotalIntervals  int
	LongIntervals   int
	ShortIntervals  int
	VariationLong   float64
	VariationShort  float64
	TotalCapte      float64
	AvgBougiesLong  float64
	AvgBougiesShort float64
}

func main() {
	fmt.Println("═══════════════════════════════════════════════════════════════════════════")
	fmt.Println("  ANALYSE COMPARATIVE DES TESTS DIRECTION")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════")

	outDir := "out"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	// Parcourir tous les sous-dossiers
	entries, err := os.ReadDir(outDir)
	if err != nil {
		fmt.Printf("❌ Erreur lecture %s: %v\n", outDir, err)
		return
	}

	var results []TestResult

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(outDir, entry.Name())
		intervallesPath := filepath.Join(dirPath, "intervalles.json")

		// Vérifier si intervalles.json existe
		if _, err := os.Stat(intervallesPath); os.IsNotExist(err) {
			continue
		}

		// Lire les intervalles
		data, err := os.ReadFile(intervallesPath)
		if err != nil {
			fmt.Printf("⚠️  Erreur lecture %s: %v\n", intervallesPath, err)
			continue
		}

		var intervalles []Intervalle
		if err := json.Unmarshal(data, &intervalles); err != nil {
			fmt.Printf("⚠️  Erreur parsing %s: %v\n", intervallesPath, err)
			continue
		}

		// Extraire les paramètres du nom du dossier
		params := parseDirectoryName(entry.Name())
		if params == nil {
			continue
		}

		// Calculer les statistiques
		result := TestResult{
			Path:          entry.Name(),
			Timeframe:     params["timeframe"].(string),
			VWMAPeriod:    params["vwma"].(int),
			SlopePeriod:   params["slope"].(int),
			ATRPeriod:     params["atr"].(int),
			ATRCoeff:      params["coeff"].(float64),
			TotalIntervals: len(intervalles),
		}

		totalBougiesLong := 0
		totalBougiesShort := 0

		for _, inter := range intervalles {
			if inter.Type == "LONG" {
				result.LongIntervals++
				result.VariationLong += inter.VariationCaptee
				totalBougiesLong += inter.NbBougies
			} else {
				result.ShortIntervals++
				result.VariationShort += inter.VariationCaptee
				totalBougiesShort += inter.NbBougies
			}
		}

		result.TotalCapte = result.VariationLong - result.VariationShort

		if result.LongIntervals > 0 {
			result.AvgBougiesLong = float64(totalBougiesLong) / float64(result.LongIntervals)
		}
		if result.ShortIntervals > 0 {
			result.AvgBougiesShort = float64(totalBougiesShort) / float64(result.ShortIntervals)
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		fmt.Println("\n❌ Aucun test trouvé dans", outDir)
		return
	}

	// Trier par TOTAL CAPTÉ décroissant
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCapte > results[j].TotalCapte
	})

	// Afficher le tableau
	fmt.Printf("\n📊 RÉSULTATS (%d tests analysés)\n\n", len(results))
	fmt.Println(strings.Repeat("─", 160))
	fmt.Printf("%-8s | %-6s | %-5s | %-3s | %-4s | %-10s | %-4s | %-8s | %-8s | %-10s | %-12s | %-12s\n",
		"Rank", "TF", "VWMA", "Slp", "ATR", "Coef", "#Int", "Long%", "Short%", "CAPTÉ%", "AvgBougie_L", "AvgBougie_S")
	fmt.Println(strings.Repeat("─", 160))

	for i, r := range results {
		rank := fmt.Sprintf("#%d", i+1)
		if i < 3 {
			rank = fmt.Sprintf("🥇#%d", i+1)
		} else if i < 10 {
			rank = fmt.Sprintf("⭐#%d", i+1)
		}

		fmt.Printf("%-8s | %-6s | %5d | %3d | %4d | %10.2f | %4d | %+8.2f | %+8.2f | %+10.2f | %12.1f | %12.1f\n",
			rank,
			r.Timeframe,
			r.VWMAPeriod,
			r.SlopePeriod,
			r.ATRPeriod,
			r.ATRCoeff,
			r.TotalIntervals,
			r.VariationLong,
			r.VariationShort,
			r.TotalCapte,
			r.AvgBougiesLong,
			r.AvgBougiesShort)
	}
	fmt.Println(strings.Repeat("─", 160))

	// Analyse par catégories
	fmt.Println("\n" + strings.Repeat("═", 100))
	fmt.Println("📈 ANALYSE PAR CATÉGORIES")
	fmt.Println(strings.Repeat("═", 100))

	analyzeByVWMA(results)
	analyzeByATRCoeff(results)
	analyzeByCandleDuration(results)

	// Recommandations
	fmt.Println("\n" + strings.Repeat("═", 100))
	fmt.Println("💡 RECOMMANDATIONS STRATÉGIQUES")
	fmt.Println(strings.Repeat("═", 100))

	recommandations(results)
}

func parseDirectoryName(name string) map[string]interface{} {
	// Format: direction_demo_5m_vwma6_slope3_k2_atr3_coef0.50
	re := regexp.MustCompile(`direction_demo_(\w+)_vwma(\d+)_slope(\d+)_k(\d+)_atr(\d+)_coef([\d.]+)`)
	matches := re.FindStringSubmatch(name)

	if len(matches) != 7 {
		return nil
	}

	vwma, _ := strconv.Atoi(matches[2])
	slope, _ := strconv.Atoi(matches[3])
	atr, _ := strconv.Atoi(matches[5])
	coeff, _ := strconv.ParseFloat(matches[6], 64)

	return map[string]interface{}{
		"timeframe": matches[1],
		"vwma":      vwma,
		"slope":     slope,
		"atr":       atr,
		"coeff":     coeff,
	}
}

func analyzeByVWMA(results []TestResult) {
	fmt.Println("\n🎯 MEILLEURS PAR VWMA PERIOD:")

	vwmaGroups := make(map[int][]TestResult)
	for _, r := range results {
		vwmaGroups[r.VWMAPeriod] = append(vwmaGroups[r.VWMAPeriod], r)
	}

	type VWMAStats struct {
		Period     int
		Count      int
		AvgCapte   float64
		BestCapte  float64
		WorstCapte float64
	}

	var stats []VWMAStats
	for period, group := range vwmaGroups {
		total := 0.0
		best := group[0].TotalCapte
		worst := group[0].TotalCapte

		for _, r := range group {
			total += r.TotalCapte
			if r.TotalCapte > best {
				best = r.TotalCapte
			}
			if r.TotalCapte < worst {
				worst = r.TotalCapte
			}
		}

		stats = append(stats, VWMAStats{
			Period:     period,
			Count:      len(group),
			AvgCapte:   total / float64(len(group)),
			BestCapte:  best,
			WorstCapte: worst,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].AvgCapte > stats[j].AvgCapte
	})

	fmt.Printf("   %-10s | %-6s | %-12s | %-12s | %-12s\n", "VWMA", "Tests", "Avg Capté", "Best", "Worst")
	fmt.Println("   " + strings.Repeat("─", 70))
	for _, s := range stats {
		fmt.Printf("   %-10d | %6d | %+12.2f%% | %+12.2f%% | %+12.2f%%\n",
			s.Period, s.Count, s.AvgCapte, s.BestCapte, s.WorstCapte)
	}
}

func analyzeByATRCoeff(results []TestResult) {
	fmt.Println("\n⚡ MEILLEURS PAR ATR COEFFICIENT:")

	coeffGroups := make(map[float64][]TestResult)
	for _, r := range results {
		coeffGroups[r.ATRCoeff] = append(coeffGroups[r.ATRCoeff], r)
	}

	type CoeffStats struct {
		Coeff      float64
		Count      int
		AvgCapte   float64
		BestCapte  float64
		WorstCapte float64
	}

	var stats []CoeffStats
	for coeff, group := range coeffGroups {
		total := 0.0
		best := group[0].TotalCapte
		worst := group[0].TotalCapte

		for _, r := range group {
			total += r.TotalCapte
			if r.TotalCapte > best {
				best = r.TotalCapte
			}
			if r.TotalCapte < worst {
				worst = r.TotalCapte
			}
		}

		stats = append(stats, CoeffStats{
			Coeff:      coeff,
			Count:      len(group),
			AvgCapte:   total / float64(len(group)),
			BestCapte:  best,
			WorstCapte: worst,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].AvgCapte > stats[j].AvgCapte
	})

	fmt.Printf("   %-10s | %-6s | %-12s | %-12s | %-12s\n", "Coeff", "Tests", "Avg Capté", "Best", "Worst")
	fmt.Println("   " + strings.Repeat("─", 70))
	for _, s := range stats {
		fmt.Printf("   %-10.2f | %6d | %+12.2f%% | %+12.2f%% | %+12.2f%%\n",
			s.Coeff, s.Count, s.AvgCapte, s.BestCapte, s.WorstCapte)
	}
}

func analyzeByCandleDuration(results []TestResult) {
	fmt.Println("\n⏱️  ANALYSE PAR DURÉE MOYENNE D'INTERVALLE:")

	// Catégoriser par durée moyenne
	var shortTerm, mediumTerm, longTerm []TestResult

	for _, r := range results {
		avgDuration := (r.AvgBougiesLong + r.AvgBougiesShort) / 2.0

		if avgDuration < 20 {
			shortTerm = append(shortTerm, r)
		} else if avgDuration < 50 {
			mediumTerm = append(mediumTerm, r)
		} else {
			longTerm = append(longTerm, r)
		}
	}

	printCategoryStats := func(name string, results []TestResult) {
		if len(results) == 0 {
			return
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].TotalCapte > results[j].TotalCapte
		})

		avgCapte := 0.0
		for _, r := range results {
			avgCapte += r.TotalCapte
		}
		avgCapte /= float64(len(results))

		best := results[0]
		worst := results[len(results)-1]

		fmt.Printf("\n   %s (%d tests, avg bougie < XX):\n", name, len(results))
		fmt.Printf("      • Moyenne capté: %+.2f%%\n", avgCapte)
		fmt.Printf("      • Meilleur: %+.2f%% (VWMA=%d, ATR_coef=%.2f, avg_bougie=%.1f)\n",
			best.TotalCapte, best.VWMAPeriod, best.ATRCoeff, (best.AvgBougiesLong+best.AvgBougiesShort)/2)
		fmt.Printf("      • Pire: %+.2f%% (VWMA=%d, ATR_coef=%.2f, avg_bougie=%.1f)\n",
			worst.TotalCapte, worst.VWMAPeriod, worst.ATRCoeff, (worst.AvgBougiesLong+worst.AvgBougiesShort)/2)
	}

	printCategoryStats("📍 COURT TERME (<20 bougies)", shortTerm)
	printCategoryStats("📊 MOYEN TERME (20-50 bougies)", mediumTerm)
	printCategoryStats("📈 LONG TERME (>50 bougies)", longTerm)
}

func recommandations(results []TestResult) {
	// Court terme: <20 bougies moyennes
	// Moyen terme: 20-50 bougies
	// Long terme: >50 bougies

	var shortTerm, mediumTerm, longTerm []TestResult

	for _, r := range results {
		avgDuration := (r.AvgBougiesLong + r.AvgBougiesShort) / 2.0

		if avgDuration < 20 {
			shortTerm = append(shortTerm, r)
		} else if avgDuration < 50 {
			mediumTerm = append(mediumTerm, r)
		} else {
			longTerm = append(longTerm, r)
		}
	}

	// Trier chaque catégorie par TOTAL CAPTÉ
	sort.Slice(shortTerm, func(i, j int) bool {
		return shortTerm[i].TotalCapte > shortTerm[j].TotalCapte
	})
	sort.Slice(mediumTerm, func(i, j int) bool {
		return mediumTerm[i].TotalCapte > mediumTerm[j].TotalCapte
	})
	sort.Slice(longTerm, func(i, j int) bool {
		return longTerm[i].TotalCapte > longTerm[j].TotalCapte
	})

	fmt.Println("\n🎯 COURT TERME (Scalping, <20 bougies = <2h en 5m):")
	if len(shortTerm) > 0 {
		best := shortTerm[0]
		fmt.Printf("   • Meilleure config: VWMA=%d, Slope=%d, ATR=%d, Coef=%.2f\n",
			best.VWMAPeriod, best.SlopePeriod, best.ATRPeriod, best.ATRCoeff)
		fmt.Printf("   • Performance: %+.2f%% capté\n", best.TotalCapte)
		fmt.Printf("   • Intervalles: %d (avg %.1f bougies)\n", best.TotalIntervals,
			(best.AvgBougiesLong+best.AvgBougiesShort)/2)
		fmt.Printf("   • Interprétation: VWMA court = réactivité élevée, ATR_coef faible = moins de bruit\n")
	}

	fmt.Println("\n📊 MOYEN TERME (Intraday, 20-50 bougies = 2-8h en 5m):")
	if len(mediumTerm) > 0 {
		best := mediumTerm[0]
		fmt.Printf("   • Meilleure config: VWMA=%d, Slope=%d, ATR=%d, Coef=%.2f\n",
			best.VWMAPeriod, best.SlopePeriod, best.ATRPeriod, best.ATRCoeff)
		fmt.Printf("   • Performance: %+.2f%% capté\n", best.TotalCapte)
		fmt.Printf("   • Intervalles: %d (avg %.1f bougies)\n", best.TotalIntervals,
			(best.AvgBougiesLong+best.AvgBougiesShort)/2)
		fmt.Printf("   • Interprétation: Équilibre entre réactivité et stabilité\n")
	}

	fmt.Println("\n📈 LONG TERME (Swing, >50 bougies = >8h en 5m):")
	if len(longTerm) > 0 {
		best := longTerm[0]
		fmt.Printf("   • Meilleure config: VWMA=%d, Slope=%d, ATR=%d, Coef=%.2f\n",
			best.VWMAPeriod, best.SlopePeriod, best.ATRPeriod, best.ATRCoeff)
		fmt.Printf("   • Performance: %+.2f%% capté\n", best.TotalCapte)
		fmt.Printf("   • Intervalles: %d (avg %.1f bougies)\n", best.TotalIntervals,
			(best.AvgBougiesLong+best.AvgBougiesShort)/2)
		fmt.Printf("   • Interprétation: VWMA long = filtre bruit, suit tendances majeures\n")
	}

	fmt.Println("\n💡 PRINCIPES GÉNÉRAUX:")
	fmt.Println("   • VWMA court (3-6) → Réactif, nombreux signaux, court terme")
	fmt.Println("   • VWMA moyen (12-20) → Équilibré, signaux qualité, moyen terme")
	fmt.Println("   • VWMA long (48+) → Filtré, peu de signaux, long terme")
	fmt.Println("   • ATR_coef bas (0.25-0.50) → Sensible, capte petits mouvements")
	fmt.Println("   • ATR_coef élevé (0.80-1.50) → Conservateur, tendances fortes uniquement")
}
