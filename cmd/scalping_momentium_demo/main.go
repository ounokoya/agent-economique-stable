package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/signals"
	"agent-economique/internal/execution"
	momentium "agent-economique/internal/signals/scalping_momentium"
)

// Configuration par défaut (alignée sur la démo Momentum)
const (
	SYMBOL     = "SOL_USDT"
	TIMEFRAME  = "1m"
	NB_CANDLES = 1000

	ATR_PERIOD        = 3
	BODY_PCT_MIN      = 0.60
	BODY_ATR_MIN      = 0.60
	STOCH_K_PERIOD    = 14
	STOCH_K_SMOOTH    = 3
	STOCH_D_PERIOD    = 3
	STOCH_K_LONG_MAX  = 50.0
	STOCH_K_SHORT_MIN = 50.0
)

type Intervalle struct {
	Numero     int
	Type       signals.SignalType
	DateDebut  time.Time
	DateFin    time.Time
	PrixDebut  float64
	PrixFin    float64
	NbBougies  int
	CaptureRaw float64   // Close(sortie) - Open(entree)
	CaptureDir float64   // Orienté: LONG= +diff, SHORT= -diff
}

func main() {
	fmt.Println("=== DEMO GENERATEUR SCALPING MOMENTIUM ===")
	fmt.Printf("Exchange  : Gate.io\n")
	fmt.Printf("Symbole   : %s\n", SYMBOL)
	fmt.Printf("Timeframe : %s\n", TIMEFRAME)
	fmt.Printf("Klines    : %d\n", NB_CANDLES)
	fmt.Printf("Params    : ATR=%d, body%%>=%.0f%%, body>=%.0f%%*ATR, StochK=%d,%d,%d, K<long=%.0f, K>short=%.0f\n",
		ATR_PERIOD, BODY_PCT_MIN*100, BODY_ATR_MIN*100, STOCH_K_PERIOD, STOCH_K_SMOOTH, STOCH_D_PERIOD, STOCH_K_LONG_MAX, STOCH_K_SHORT_MIN)

	// 1) Récupération des klines Gate.io
	ctx := context.Background()
	client := gateio.NewClient()
	klines, err := client.GetKlines(ctx, SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}
	// Tri chronologique défensif
	sort.Slice(klines, func(i, j int) bool { return klines[i].OpenTime.Before(klines[j].OpenTime) })
	fmt.Printf("Klines recuperees: %d\n", len(klines))

	// 2) Convertir en format signals.Kline
	sigKlines := make([]signals.Kline, len(klines))
	for i, k := range klines {
		sigKlines[i] = signals.Kline{OpenTime: k.OpenTime, Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume}
	}

	// 3) Instancier le générateur
	gen := momentium.NewGenerator(momentium.Config{
		ATRPeriod:      ATR_PERIOD,
		BodyPctMin:     BODY_PCT_MIN,
		BodyATRMin:     BODY_ATR_MIN,
		StochKPeriod:   STOCH_K_PERIOD,
		StochKSmooth:   STOCH_K_SMOOTH,
		StochDPeriod:   STOCH_D_PERIOD,
		StochKLongMax:  STOCH_K_LONG_MAX,
		StochKShortMin: STOCH_K_SHORT_MIN,
	})
	if err := gen.Initialize(signals.GeneratorConfig{Symbol: SYMBOL, Timeframe: TIMEFRAME, HistorySize: NB_CANDLES}); err != nil {
		log.Fatalf("Erreur initialisation: %v", err)
	}

	// 4) Calculer indicateurs
	fmt.Println("\nCalcul des indicateurs...")
	if err := gen.CalculateIndicators(sigKlines); err != nil {
		log.Fatalf("Erreur calcul indicateurs: %v", err)
	}

	// 5) Détecter signaux
	fmt.Println("Detection des signaux...")
	allSignals, err := gen.DetectSignals(sigKlines)
	if err != nil {
		log.Fatalf("Erreur detection: %v", err)
	}
	fmt.Printf("Signaux detectes: %d\n", len(allSignals))

	// 6) Intervalles ENTRY→EXIT
	intervalles := buildIntervalles(allSignals, sigKlines)

	// 7) Affichage
	displaySignals(allSignals)
	displayIntervalles(intervalles)
	displayStatistics(intervalles)

	// 8) Métriques
	m := gen.GetMetrics()
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("METRIQUES GENERATEUR MOMENTIUM")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Total signaux : %d | Entry=%d | Exit=%d | Long=%d | Short=%d\n", m.TotalSignals, m.EntrySignals, m.ExitSignals, m.LongSignals, m.ShortSignals)
	fmt.Printf("Confiance moy.: %.2f\n", m.AvgConfidence)
	fmt.Println(strings.Repeat("=", 100))

	// 9) Export JSON
	if err := exportResults(sigKlines, intervalles, allSignals); err != nil {
		log.Printf("⚠️  Export echoue: %v", err)
	}

	fmt.Println("\n=== FIN DEMO SCALPING MOMENTIUM ===")
}

func buildIntervalles(sigs []signals.Signal, klines []signals.Kline) []Intervalle {
	var intervalles []Intervalle

	// Index des klines par timestamp pour accès O(1)
	idxByTime := make(map[time.Time]int, len(klines))
	for i, k := range klines { idxByTime[k.OpenTime] = i }

	// Regrouper signaux par index de bougie
	sigsByIdx := make(map[int][]signals.Signal)
	for _, s := range sigs {
		if idx, ok := idxByTime[s.Timestamp]; ok {
			sigsByIdx[idx] = append(sigsByIdx[idx], s)
		}
	}

	type position struct {
		typ      signals.SignalType
		entryIdx int
		entryT   time.Time
		entryO   float64
		tr       *execution.Trailing
	}
	var pos *position
	num := 0

	// Itérer bougie par bougie
	for i := 0; i < len(klines); i++ {
		k := klines[i]

		// Mettre à jour trailing et tester stop sur la CLOSE de cette bougie
		if pos != nil {
			pos.tr.Update(k.Close)
			if hit, stopPrice := pos.tr.Hit(k.Close); hit {
				num++
				inter := Intervalle{
					Numero:    num,
					Type:      pos.typ,
					DateDebut: pos.entryT,
					DateFin:   k.OpenTime,
					PrixDebut: pos.entryO,
					PrixFin:   stopPrice,
					NbBougies: i - pos.entryIdx + 1,
				}
				inter.CaptureRaw = stopPrice - pos.entryO
				if pos.typ == signals.SignalTypeLong { inter.CaptureDir = inter.CaptureRaw } else { inter.CaptureDir = -inter.CaptureRaw }
				intervalles = append(intervalles, inter)
				pos = nil
			}
		}

		// Récupérer signaux de cette bougie, ordonner EXIT avant ENTRY
		if sigList, ok := sigsByIdx[i]; ok {
			exits := make([]signals.Signal, 0)
			entries := make([]signals.Signal, 0)
			for _, s := range sigList {
				if s.Action == signals.SignalActionExit { exits = append(exits, s) } else if s.Action == signals.SignalActionEntry { entries = append(entries, s) }
			}

			// 1) Traiter EXIT (fermeture au CLOSE)
			if pos != nil {
				for _, s := range exits {
					if s.Type != pos.typ { continue }
					num++
					inter := Intervalle{
						Numero:    num,
						Type:      pos.typ,
						DateDebut: pos.entryT,
						DateFin:   k.OpenTime,
						PrixDebut: pos.entryO,
						PrixFin:   k.Close,
						NbBougies: i - pos.entryIdx + 1,
					}
					inter.CaptureRaw = k.Close - pos.entryO
					if pos.typ == signals.SignalTypeLong { inter.CaptureDir = inter.CaptureRaw } else { inter.CaptureDir = -inter.CaptureRaw }
					intervalles = append(intervalles, inter)
					pos = nil
					break // un seul EXIT suffit
				}
			}

			// 2) Traiter ENTRY
			for _, s := range entries {
				if pos == nil {
					// Ouvrir position au OPEN de la bougie
					atr := getFloat(s.Metadata, "atr")
					side := execution.SideShort
					if s.Type == signals.SignalTypeLong { side = execution.SideLong }
					tr := execution.NewTrailing(side, k.Open, atr, 0.005)
					p := &position{typ: s.Type, entryIdx: i, entryT: k.OpenTime, entryO: k.Open, tr: tr}
					pos = p
				} else if s.Type != pos.typ {
					// ENTRY opposé: fermer au CLOSE puis ouvrir au OPEN
					num++
					inter := Intervalle{
						Numero:    num,
						Type:      pos.typ,
						DateDebut: pos.entryT,
						DateFin:   k.OpenTime,
						PrixDebut: pos.entryO,
						PrixFin:   k.Close,
						NbBougies: i - pos.entryIdx + 1,
					}
					inter.CaptureRaw = k.Close - pos.entryO
					if pos.typ == signals.SignalTypeLong { inter.CaptureDir = inter.CaptureRaw } else { inter.CaptureDir = -inter.CaptureRaw }
					intervalles = append(intervalles, inter)

					// Ouvrir nouvelle position
					atr := getFloat(s.Metadata, "atr")
					side := execution.SideShort
					if s.Type == signals.SignalTypeLong { side = execution.SideLong }
					tr := execution.NewTrailing(side, k.Open, atr, 0.005)
					p := &position{typ: s.Type, entryIdx: i, entryT: k.OpenTime, entryO: k.Open, tr: tr}
					pos = p
				}
			}
		}
	}

	return intervalles
}

func displaySignals(sigs []signals.Signal) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("SIGNAUX SCALPING MOMENTIUM")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-19s | %-6s | %-6s | %8s | %10s | %8s | %8s | %7s\n",
		"#", "Date/Heure", "Action", "Type", "Prix", "Confiance", "Body%", "B/ATR", "K")
	fmt.Println(strings.Repeat("-", 140))

	// Chrono
	sort.Slice(sigs, func(i, j int) bool { return sigs[i].Timestamp.Before(sigs[j].Timestamp) })
	for i, s := range sigs {
		bodyPct := getFloat(s.Metadata, "body_pct") * 100
		b2atr := getFloat(s.Metadata, "body_to_atr")
		kval := getFloat(s.Metadata, "stoch_k")
		fmt.Printf("%-4d | %s | %-6s | %-6s | %8.2f | %10.2f | %7.1f%% | %8.2f | %7.2f\n",
			i+1,
			s.Timestamp.Format("2006-01-02 15:04:05"),
			s.Action,
			s.Type,
			s.Price,
			s.Confidence,
			bodyPct,
			b2atr,
			kval,
		)
	}
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("Total signaux: %d\n", len(sigs))
}

func displayIntervalles(inters []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("INTERVALLES ENTRY→EXIT")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s | %-6s | %-19s | %-19s | %-8s | %-8s | %-7s | %-9s | %-9s\n",
		"#", "Type", "Debut", "Fin", "P.Debut", "P.Fin", "Bougies", "Capture", "CaptureDir")
	fmt.Println(strings.Repeat("-", 120))

	// Chrono
	sort.Slice(inters, func(i, j int) bool { return inters[i].DateDebut.Before(inters[j].DateDebut) })
	for _, inter := range inters {
		fmt.Printf("%-4d | %-6s | %s | %s | %8.2f | %8.2f | %-7d | %9.2f | %9.2f\n",
			inter.Numero,
			inter.Type,
			inter.DateDebut.Format("2006-01-02 15:04:05"),
			inter.DateFin.Format("2006-01-02 15:04:05"),
			inter.PrixDebut,
			inter.PrixFin,
			inter.NbBougies,
			inter.CaptureRaw,
			inter.CaptureDir,
		)
	}
	fmt.Println(strings.Repeat("=", 120))
}

func displayStatistics(inters []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES INTERVALLES")
	fmt.Println(strings.Repeat("=", 100))
	if len(inters) == 0 {
		fmt.Println("Aucun intervalle complet.")
		return
	}
	var sumRaw, sumDir float64
	var nbLong, nbShort int
	for _, it := range inters {
		sumRaw += it.CaptureRaw
		sumDir += it.CaptureDir
		if it.Type == signals.SignalTypeLong { nbLong++ } else { nbShort++ }
	}
	fmt.Printf("Nb intervalles : %d (LONG=%d, SHORT=%d)\n", len(inters), nbLong, nbShort)
	fmt.Printf("Somme Capture  : %.2f | Somme Dir: %.2f\n", sumRaw, sumDir)
	fmt.Println(strings.Repeat("=", 100))
}

func exportResults(klines []signals.Kline, inters []Intervalle, sigs []signals.Signal) error {
	outDir := "out/scalping_momentium_generator_demo_" + time.Now().Format("20060102_150405")
	if err := os.MkdirAll(outDir, 0755); err != nil { return fmt.Errorf("mkdir: %w", err) }
	if err := saveJSON(filepath.Join(outDir, "klines.json"), klines); err != nil { return err }
	if err := saveJSON(filepath.Join(outDir, "intervalles.json"), inters); err != nil { return err }
	if err := saveJSON(filepath.Join(outDir, "signals.json"), sigs); err != nil { return err }
	fmt.Printf("\n✅ Resultats exportes dans: %s\n", outDir)
	return nil
}

func saveJSON(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func getFloat(m map[string]interface{}, key string) float64 {
	if m == nil { return 0 }
	if v, ok := m[key]; ok {
		switch x := v.(type) {
		case float64:
			return x
		case float32:
			return float64(x)
		case int:
			return float64(x)
		case int64:
			return float64(x)
		}
	}
	return 0
}
