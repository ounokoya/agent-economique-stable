package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/bybit"
	"agent-economique/internal/execution"
	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
	bfm "agent-economique/internal/signals/ban_fin_momentium"
)

// ===============================
// Paramètres par défaut (modifiables)
// ===============================
const (
	SYMBOL          = "SOLUSDT" // Bybit symbol format
	TIMEFRAME       = "5m"
	TREND_TIMEFRAME = "1h"
	NB_CANDLES      = 2000

	ATR_PERIOD     = 10
	BODY_ATR_MULT  = 0.8 // Body > MULT * ATR(10)
	VOL_SMA_PERIOD = 10
	VOL_COEFF      = 0.8   // Volume > VOL_COEFF * SMA(volume)
	AGGREGATE_3    = false // Analyse par bougie unique par défaut

	// VWMA trend filter (fast/slow) – activé par défaut
	VWMA_FAST_PERIOD       = 4
	VWMA_SLOW_PERIOD       = 12
	ENABLE_VWMA_TREND_GATE = false
	ENABLE_VWMA_CROSS_GATE = false
	VWMA_TRAIL_OFFSET      = 0.005
	TRAIL_PCT              = 0.005 // 0.3% trailing standard sur le prix
	TRAIL_PCT1             = 0.003
	TRAIL_PCT2             = 0.01
	TRAIL_PCT3             = 0.01
	TRAIL_PROFIT_THR2      = 0.005
	TRAIL_PROFIT_THR3      = 0.03

	STOCH_K_PERIOD = 14
	STOCH_K_SMOOTH = 2
	STOCH_D_PERIOD = 3
	// Bornes directionnelles Stoch (optionnelles)
	STOCH_K_OVERSOLD_LONG    = 40.0
	STOCH_K_OVERBOUGHT_LONG  = 70.0
	STOCH_K_OVERSOLD_SHORT   = 30.0
	STOCH_K_OVERBOUGHT_SHORT = 60.0
	// Toggles Stoch
	ENABLE_STOCH_CROSS                   = true  // Step 3 optionnel
	ENABLE_STOCH_K_OVERSOLD_LONG_GATE    = false // Step 3 optionnel
	ENABLE_STOCH_K_OVERBOUGHT_LONG_GATE  = false // Step 3 optionnel
	ENABLE_STOCH_K_OVERSOLD_SHORT_GATE   = false // Step 3 optionnel
	ENABLE_STOCH_K_OVERBOUGHT_SHORT_GATE = false // Step 3 optionnel

	MFI_PERIOD           = 30
	MFI_OVERBOUGHT_LONG  = 80.0
	MFI_OVERSOLD_LONG    = 20.0
	MFI_OVERBOUGHT_SHORT = 80.0
	MFI_OVERSOLD_SHORT   = 20.0
	// Toggles MFI (par extrême)
	ENABLE_MFI_OVERSOLD_LONG_GATE    = false
	ENABLE_MFI_OVERBOUGHT_LONG_GATE  = false
	ENABLE_MFI_OVERSOLD_SHORT_GATE   = false
	ENABLE_MFI_OVERBOUGHT_SHORT_GATE = false

	CCI_PERIOD           = 20
	CCI_OVERBOUGHT_LONG  = 100.0
	CCI_OVERSOLD_LONG    = -100.0
	CCI_OVERBOUGHT_SHORT = 100.0
	CCI_OVERSOLD_SHORT   = -100.0
	// Toggles CCI (par extrême)
	ENABLE_CCI_OVERSOLD_LONG_GATE    = false
	ENABLE_CCI_OVERBOUGHT_LONG_GATE  = false
	ENABLE_CCI_OVERSOLD_SHORT_GATE   = false
	ENABLE_CCI_OVERBOUGHT_SHORT_GATE = false
	// CCI trend (long) for trend gating
	CCI_TREND_PERIOD      = 240
	ENABLE_CCI_TREND_GATE = true
	ENABLE_BAR_GATE       = false

	// MACD (TV standard)
	MACD_FAST_PERIOD   = 12
	MACD_SLOW_PERIOD   = 26
	MACD_SIGNAL_PERIOD = 9
	// Toggles MACD (gates optionnels)
	ENABLE_MACD_CROSS_GATE = false
	ENABLE_MACD_SIGN_GATE  = false
	ENABLE_MACD_HIST_GATE  = false

	ENABLE_WINDOW_MODE = false
	// WINDOW_OPEN_MODE: "auto" (défaut), "cross", "bar", "stoch_cross", "vwma_cross", "macd_cross"
	WINDOW_BARS      = 10
	WINDOW_OPEN_MODE = "stoch_cross"
)

func main() {
	fmt.Println("=== DEMO BAN_FIN_MOMENTIUM ===")
	fmt.Printf("Symbole   : %s\n", SYMBOL)
	fmt.Printf("Timeframe : %s\n", TIMEFRAME)
	fmt.Printf("Règles Open/Close avec ATR(%d), agg3=%v, body>%.2f*ATR, vol>SMA*%.2f, VWMA(%d/%d) trend=%v, Stoch/MFI/CCI gates, cross=%v\n",
		ATR_PERIOD, AGGREGATE_3, BODY_ATR_MULT, VOL_COEFF, VWMA_FAST_PERIOD, VWMA_SLOW_PERIOD, ENABLE_VWMA_TREND_GATE, ENABLE_STOCH_CROSS)

	// 1) Données (Bybit)
	ctx := context.Background()
	client := bybit.NewClient()
	klines, err := client.GetKlines(ctx, SYMBOL, TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("erreur klines: %v", err)
	}
	sort.Slice(klines, func(i, j int) bool { return klines[i].OpenTime.Before(klines[j].OpenTime) })
	fmt.Printf("Klines: %d\n", len(klines))
	if len(klines) < 50 {
		fmt.Println("Pas assez de données")
		return
	}
	trendKlines, err := client.GetKlines(ctx, SYMBOL, TREND_TIMEFRAME, NB_CANDLES)
	if err != nil {
		log.Fatalf("erreur klines tendance: %v", err)
	}
	sort.Slice(trendKlines, func(i, j int) bool { return trendKlines[i].OpenTime.Before(trendKlines[j].OpenTime) })

	// 2) Convertir en format generator
	sigK := make([]signals.Kline, len(klines))
	for i, k := range klines {
		sigK[i] = signals.Kline{OpenTime: k.OpenTime, Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume}
	}
	trendCCI := computeTrendCCI(trendKlines)
	trendSign := buildTrendSign(sigK, trendKlines, trendCCI)

	// 3) Instancier générateur
	gen := bfm.NewGenerator(bfm.Config{
		ATRPeriod:           ATR_PERIOD,
		BodyATRMultiplier:   BODY_ATR_MULT,
		VolumeSMAPeriod:     VOL_SMA_PERIOD,
		VolumeCoeff:         VOL_COEFF,
		Aggregate3:          AGGREGATE_3,
		EnableBarGate:       ENABLE_BAR_GATE,
		VWMAFastPeriod:      VWMA_FAST_PERIOD,
		VWMASlowPeriod:      VWMA_SLOW_PERIOD,
		EnableVWMATrendGate: ENABLE_VWMA_TREND_GATE,
		EnableVWMACross:     ENABLE_VWMA_CROSS_GATE,
		// Setups actifs par défaut (Step 1) ; filtres optionnels désactivés
		EnableOpenLong:   true,
		EnableOpenShort:  true,
		EnableCloseLong:  true,
		EnableCloseShort: true,
		StochKPeriod:     STOCH_K_PERIOD,
		StochKSmooth:     STOCH_K_SMOOTH,
		StochDPeriod:     STOCH_D_PERIOD,
		EnableStochCross: ENABLE_STOCH_CROSS,
		// Stoch bornes par direction + toggles indépendants
		StochKOversoldLong:              STOCH_K_OVERSOLD_LONG,
		StochKOverboughtLong:            STOCH_K_OVERBOUGHT_LONG,
		StochKOversoldShort:             STOCH_K_OVERSOLD_SHORT,
		StochKOverboughtShort:           STOCH_K_OVERBOUGHT_SHORT,
		EnableStochKOversoldLongGate:    ENABLE_STOCH_K_OVERSOLD_LONG_GATE,
		EnableStochKOverboughtLongGate:  ENABLE_STOCH_K_OVERBOUGHT_LONG_GATE,
		EnableStochKOversoldShortGate:   ENABLE_STOCH_K_OVERSOLD_SHORT_GATE,
		EnableStochKOverboughtShortGate: ENABLE_STOCH_K_OVERBOUGHT_SHORT_GATE,
		MFIPeriod:                       MFI_PERIOD,
		MFIOverboughtLong:               MFI_OVERBOUGHT_LONG,
		MFIOverboughtShort:              MFI_OVERBOUGHT_SHORT,
		MFIOversoldLong:                 MFI_OVERSOLD_LONG,
		MFIOversoldShort:                MFI_OVERSOLD_SHORT,
		CCIPeriod:                       CCI_PERIOD,
		CCIOverboughtLong:               CCI_OVERBOUGHT_LONG,
		CCIOverboughtShort:              CCI_OVERBOUGHT_SHORT,
		CCIOversoldLong:                 CCI_OVERSOLD_LONG,
		CCIOversoldShort:                CCI_OVERSOLD_SHORT,
		EnableMFIGate:                   (ENABLE_MFI_OVERSOLD_LONG_GATE || ENABLE_MFI_OVERBOUGHT_LONG_GATE || ENABLE_MFI_OVERSOLD_SHORT_GATE || ENABLE_MFI_OVERBOUGHT_SHORT_GATE),
		EnableCCIGate:                   (ENABLE_CCI_OVERSOLD_LONG_GATE || ENABLE_CCI_OVERBOUGHT_LONG_GATE || ENABLE_CCI_OVERSOLD_SHORT_GATE || ENABLE_CCI_OVERBOUGHT_SHORT_GATE),
		// Toggles indépendants MFI/CCI
		EnableMFIOversoldLongGate:    ENABLE_MFI_OVERSOLD_LONG_GATE,
		EnableMFIOverboughtLongGate:  ENABLE_MFI_OVERBOUGHT_LONG_GATE,
		EnableMFIOversoldShortGate:   ENABLE_MFI_OVERSOLD_SHORT_GATE,
		EnableMFIOverboughtShortGate: ENABLE_MFI_OVERBOUGHT_SHORT_GATE,
		EnableCCIOversoldLongGate:    ENABLE_CCI_OVERSOLD_LONG_GATE,
		EnableCCIOverboughtLongGate:  ENABLE_CCI_OVERBOUGHT_LONG_GATE,
		EnableCCIOversoldShortGate:   ENABLE_CCI_OVERSOLD_SHORT_GATE,
		EnableCCIOverboughtShortGate: ENABLE_CCI_OVERBOUGHT_SHORT_GATE,
		MACDFastPeriod:               MACD_FAST_PERIOD,
		MACDSlowPeriod:               MACD_SLOW_PERIOD,
		MACDSignalPeriod:             MACD_SIGNAL_PERIOD,
		EnableMACDCrossGate:          ENABLE_MACD_CROSS_GATE,
		EnableMACDSignGate:           ENABLE_MACD_SIGN_GATE,
		EnableMACDHistGate:           ENABLE_MACD_HIST_GATE,
	})
	if err := gen.Initialize(signals.GeneratorConfig{Symbol: SYMBOL, Timeframe: TIMEFRAME, HistorySize: NB_CANDLES}); err != nil {
		log.Fatalf("init: %v", err)
	}
	if err := gen.CalculateIndicators(sigK); err != nil {
		log.Fatalf("calc: %v", err)
	}

	// 4) Détecter signaux (structure ban_fin)
	// Warmup minimal selon filtres activés
	warmup := ATR_PERIOD
	if VOL_SMA_PERIOD > warmup {
		warmup = VOL_SMA_PERIOD
	}
	if VWMA_FAST_PERIOD > warmup {
		warmup = VWMA_FAST_PERIOD
	}
	if ENABLE_VWMA_TREND_GATE {
		if VWMA_SLOW_PERIOD > warmup {
			warmup = VWMA_SLOW_PERIOD
		}
	}
	if ENABLE_CCI_TREND_GATE {
		if CCI_TREND_PERIOD > warmup {
			warmup = CCI_TREND_PERIOD
		}
	}
	// Warmup Stoch (K/D) uniquement si on utilise le cross comme gate de timing.
	if ENABLE_STOCH_CROSS {
		if t := STOCH_K_PERIOD + STOCH_K_SMOOTH + STOCH_D_PERIOD; t > warmup {
			warmup = t
		}
	}
	needMFI := ENABLE_MFI_OVERSOLD_LONG_GATE || ENABLE_MFI_OVERBOUGHT_LONG_GATE || ENABLE_MFI_OVERSOLD_SHORT_GATE || ENABLE_MFI_OVERBOUGHT_SHORT_GATE
	if needMFI {
		if MFI_PERIOD > warmup {
			warmup = MFI_PERIOD
		}
	}
	needCCI := ENABLE_CCI_OVERSOLD_LONG_GATE || ENABLE_CCI_OVERBOUGHT_LONG_GATE || ENABLE_CCI_OVERSOLD_SHORT_GATE || ENABLE_CCI_OVERBOUGHT_SHORT_GATE
	if needCCI {
		if CCI_PERIOD > warmup {
			warmup = CCI_PERIOD
		}
	}
	needMACD := ENABLE_MACD_CROSS_GATE || ENABLE_MACD_SIGN_GATE || ENABLE_MACD_HIST_GATE
	if needMACD {
		if t := MACD_SLOW_PERIOD + MACD_SIGNAL_PERIOD; t > warmup {
			warmup = t
		}
	}
	if AGGREGATE_3 && warmup < 2 {
		warmup = 2
	}

	var all []signals.Signal
	if ENABLE_WINDOW_MODE {
		// Mode fenêtre : utilisation du WindowFinder
		finder := bfm.NewWindowFinder(gen, bfm.WindowConfig{
			WindowBars: WINDOW_BARS,
			OpenMode:   bfm.WindowOpenMode(WINDOW_OPEN_MODE),
		})
		for i := warmup; i < len(sigK); i++ {
			// i est l'index de la dernière bougie fermée évaluée par le Finder
			sig, err := finder.Step(sigK, i)
			if err != nil {
				log.Fatalf("window step: %v", err)
			}
			if sig != nil {
				if ENABLE_CCI_TREND_GATE {
					if i < len(trendSign) {
						s := trendSign[i]
						if s == 0 {
							continue
						}
						if sig.Type == signals.SignalTypeLong && s <= 0 {
							continue
						}
						if sig.Type == signals.SignalTypeShort && s >= 0 {
							continue
						}
					}
				}
				all = append(all, *sig)
			}
		}
	} else {
		// Mode classique : EvaluateLast sur fenêtre roulante
		for i := warmup; i < len(sigK); i++ {
			prefix := sigK[:i+1]
			sig, err := gen.EvaluateLast(prefix)
			if err != nil {
				log.Fatalf("evaluate last: %v", err)
			}
			if sig != nil {
				if ENABLE_CCI_TREND_GATE {
					if i < len(trendSign) {
						s := trendSign[i]
						if s == 0 {
							continue
						}
						if sig.Type == signals.SignalTypeLong && s <= 0 {
							continue
						}
						if sig.Type == signals.SignalTypeShort && s >= 0 {
							continue
						}
					}
				}
				all = append(all, *sig)
			}
		}
	}
	fmt.Printf("Signaux détectés: %d\n", len(all))

	// 5) Construire intervalles avec trailing hybride (selon la règle 3%)
	inters := buildIntervalles(all, sigK)

	// 6) Affichage
	displaySignals(all)
	displayIntervalles(inters)
	displayStats(inters)

	// 7) Export JSON
	if err := exportResults(sigK, inters, all); err != nil {
		log.Printf("export: %v", err)
	}

	fmt.Println("=== FIN DEMO BAN_FIN_MOMENTIUM ===")
}

type Intervalle struct {
	Numero     int
	Type       signals.SignalType
	DateDebut  time.Time
	DateFin    time.Time
	PrixDebut  float64
	PrixFin    float64
	NbBougies  int
	CaptureRaw float64
	CaptureDir float64
}

type trailingStop interface {
	Update(price float64)
	Hit(price float64) (bool, float64)
}

func buildIntervalles(sigs []signals.Signal, klines []signals.Kline) []Intervalle {
	var inters []Intervalle
	// index des klines
	idxByTime := make(map[time.Time]int, len(klines))
	for i, k := range klines {
		idxByTime[k.OpenTime] = i
	}
	// group by index
	sigsByIdx := make(map[int][]signals.Signal)
	for _, s := range sigs {
		if idx, ok := idxByTime[s.Timestamp]; ok {
			sigsByIdx[idx] = append(sigsByIdx[idx], s)
		}
	}

	type position struct {
		typ        signals.SignalType
		entryIdx   int
		entryT     time.Time
		entryPrice float64
		tr         trailingStop
	}
	var pos *position
	num := 0

	for i := 0; i < len(klines); i++ {
		k := klines[i]

		// 1) MAJ trailing et test stop (sur la CLOSE)
		if pos != nil {
			pos.tr.Update(k.Close)
			if hit, stopPrice := pos.tr.Hit(k.Close); hit {
				num++
				it := Intervalle{Numero: num, Type: pos.typ, DateDebut: pos.entryT, DateFin: k.OpenTime, PrixDebut: pos.entryPrice, PrixFin: stopPrice, NbBougies: i - pos.entryIdx + 1}
				delta := stopPrice - pos.entryPrice
				pct := 0.0
				if pos.entryPrice != 0 {
					pct = (delta / pos.entryPrice) * 100.0
				}
				it.CaptureRaw = pct
				if pos.typ == signals.SignalTypeLong {
					it.CaptureDir = pct
				} else {
					it.CaptureDir = -pct
				}
				inters = append(inters, it)
				pos = nil
			}
		}

		// 2) Traiter signaux de cette bougie: EXIT avant ENTRY
		if list, ok := sigsByIdx[i]; ok {
			exits := make([]signals.Signal, 0)
			entries := make([]signals.Signal, 0)
			for _, s := range list {
				if s.Action == signals.SignalActionExit {
					exits = append(exits, s)
				} else if s.Action == signals.SignalActionEntry {
					entries = append(entries, s)
				}
			}

			// a) EXIT
			if pos != nil {
				for _, s := range exits {
					if s.Type != pos.typ {
						continue
					}
					num++
					it := Intervalle{Numero: num, Type: pos.typ, DateDebut: pos.entryT, DateFin: k.OpenTime, PrixDebut: pos.entryPrice, PrixFin: k.Close, NbBougies: i - pos.entryIdx + 1}
					delta := k.Close - pos.entryPrice
					pct := 0.0
					if pos.entryPrice != 0 {
						pct = (delta / pos.entryPrice) * 100.0
					}
					it.CaptureRaw = pct
					if pos.typ == signals.SignalTypeLong {
						it.CaptureDir = pct
					} else {
						it.CaptureDir = -pct
					}
					inters = append(inters, it)
					pos = nil
				}
			}

			// b) ENTRY
			for _, s := range entries {
				if pos == nil {
					// ouvrir au CLOSE
					entry := k.Close
					var side execution.Side
					if s.Type == signals.SignalTypeLong {
						side = execution.SideLong
					} else {
						side = execution.SideShort
					}
					tr := execution.NewMultiStagePercentTrailing(side, entry, TRAIL_PCT1, TRAIL_PCT2, TRAIL_PCT3, TRAIL_PROFIT_THR2, TRAIL_PROFIT_THR3)
					pos = &position{typ: s.Type, entryIdx: i, entryT: k.OpenTime, entryPrice: entry, tr: tr}
				} else if s.Type != pos.typ {
					// clôturer au CLOSE puis ouvrir au CLOSE
					num++
					it := Intervalle{Numero: num, Type: pos.typ, DateDebut: pos.entryT, DateFin: k.OpenTime, PrixDebut: pos.entryPrice, PrixFin: k.Close, NbBougies: i - pos.entryIdx + 1}
					delta := k.Close - pos.entryPrice
					pct := 0.0
					if pos.entryPrice != 0 {
						pct = (delta / pos.entryPrice) * 100.0
					}
					it.CaptureRaw = pct
					if pos.typ == signals.SignalTypeLong {
						it.CaptureDir = pct
					} else {
						it.CaptureDir = -pct
					}
					inters = append(inters, it)

					entry := k.Close
					var side execution.Side
					if s.Type == signals.SignalTypeLong {
						side = execution.SideLong
					} else {
						side = execution.SideShort
					}
					tr := execution.NewMultiStagePercentTrailing(side, entry, TRAIL_PCT1, TRAIL_PCT2, TRAIL_PCT3, TRAIL_PROFIT_THR2, TRAIL_PROFIT_THR3)
					pos = &position{typ: s.Type, entryIdx: i, entryT: k.OpenTime, entryPrice: entry, tr: tr}
				}
			}
		}
	}

	return inters
}

func displaySignals(sigs []signals.Signal) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("SIGNAUX BAN_FIN_MOMENTIUM")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-19s | %-6s | %-6s | %8s | %10s | %8s | %8s | %7s | %7s | %8s | %8s\n",
		"#", "Date/Heure", "Action", "Type", "Prix", "Confiance", "Body/ATR", "Vol/SMA", "K", "MFI", "VWMAf", "VWMAs")
	fmt.Println(strings.Repeat("-", 140))
	// tri chrono
	sort.Slice(sigs, func(i, j int) bool { return sigs[i].Timestamp.Before(sigs[j].Timestamp) })
	for i, s := range sigs {
		b2atr := getFloat(s.Metadata, "body_to_atr")
		v2sma := getFloat(s.Metadata, "vol_to_sma")
		kval := getFloat(s.Metadata, "stoch_k")
		mfiv := getFloat(s.Metadata, "mfi")
		vwf := getFloat(s.Metadata, "vwma_fast")
		vws := getFloat(s.Metadata, "vwma_slow")
		fmt.Printf("%-4d | %s | %-6s | %-6s | %8.2f | %10.2f | %8.2f | %8.2f | %7.2f | %7.2f | %8.2f | %8.2f\n",
			i+1, s.Timestamp.Format("2006-01-02 15:04:05"), s.Action, s.Type, s.Price, s.Confidence, b2atr, v2sma, kval, mfiv, vwf, vws)
	}
	fmt.Println(strings.Repeat("=", 140))
}

func displayIntervalles(inters []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("INTERVALLES ENTRY→EXIT")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s | %-6s | %-19s | %-19s | %-8s | %-8s | %-7s | %-9s | %-9s\n",
		"#", "Type", "Début", "Fin", "P.Début", "P.Fin", "Bougies", "Capture%", "CaptureDir%")
	fmt.Println(strings.Repeat("-", 120))
	// chrono
	sort.Slice(inters, func(i, j int) bool { return inters[i].DateDebut.Before(inters[j].DateDebut) })
	for _, it := range inters {
		fmt.Printf("%-4d | %-6s | %s | %s | %8.2f | %8.2f | %-7d | %9.2f | %9.2f\n",
			it.Numero, it.Type, it.DateDebut.Format("2006-01-02 15:04:05"), it.DateFin.Format("2006-01-02 15:04:05"), it.PrixDebut, it.PrixFin, it.NbBougies, it.CaptureRaw, it.CaptureDir)
	}
	fmt.Println(strings.Repeat("=", 120))
}

func displayStats(inters []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES")
	fmt.Println(strings.Repeat("=", 100))
	if len(inters) == 0 {
		fmt.Println("Aucun intervalle")
		return
	}
	var sumRaw, sumDir float64
	var nbLong, nbShort int
	for _, it := range inters {
		sumRaw += it.CaptureRaw
		sumDir += it.CaptureDir
		if it.Type == signals.SignalTypeLong {
			nbLong++
		} else {
			nbShort++
		}
	}
	fmt.Printf("Nb intervalles : %d (LONG=%d, SHORT=%d)\n", len(inters), nbLong, nbShort)
	fmt.Printf("Somme Capture  : %.2f | Somme Dir: %.2f\n", sumRaw, sumDir)
	fmt.Println(strings.Repeat("=", 100))
}

func exportResults(klines []signals.Kline, inters []Intervalle, sigs []signals.Signal) error {
	outDir := "out/ban_fin_momentium_" + time.Now().Format("20060102_150405")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err := saveJSON(filepath.Join(outDir, "klines.json"), klines); err != nil {
		return err
	}
	if err := saveJSON(filepath.Join(outDir, "intervalles.json"), inters); err != nil {
		return err
	}
	if err := saveJSON(filepath.Join(outDir, "signals.json"), sigs); err != nil {
		return err
	}
	fmt.Printf("\n✅ Export: %s\n", outDir)
	return nil
}

func saveJSON(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func getFloat(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
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
	return math.NaN()
}

// computeTrendCCI calcule CCI_trend sur les klines du timeframe de tendance.
func computeTrendCCI(trendKlines []bybit.Kline) []float64 {
	if len(trendKlines) == 0 {
		return nil
	}
	highs := make([]float64, len(trendKlines))
	lows := make([]float64, len(trendKlines))
	closes := make([]float64, len(trendKlines))
	for i, k := range trendKlines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
	}
	cciInd := indicators.NewCCITVStandard(CCI_TREND_PERIOD)
	return cciInd.Calculate(highs, lows, closes)
}

// buildTrendSign projette le signe de CCI_trend (TF tendance) sur chaque bougie du TF signaux.
// Retourne un slice de même taille que sigK, contenant -1 (tendance short), +1 (tendance long) ou 0 (indéfini).
func buildTrendSign(sigK []signals.Kline, trendKlines []bybit.Kline, trendCCI []float64) []int {
	n := len(sigK)
	res := make([]int, n)
	if len(trendKlines) == 0 || len(trendCCI) == 0 {
		return res
	}
	j := 0
	for i := 0; i < n; i++ {
		// avancer j tant que la prochaine bougie de tendance commence avant ou à l'heure de la bougie signal
		for j+1 < len(trendKlines) && trendKlines[j+1].OpenTime.Before(sigK[i].OpenTime.Add(1*time.Nanosecond)) {
			j++
		}
		if j >= len(trendCCI) {
			continue
		}
		v := trendCCI[j]
		if math.IsNaN(v) {
			continue
		}
		if v > 0 {
			res[i] = 1
		} else if v < 0 {
			res[i] = -1
		} else {
			res[i] = 0
		}
	}
	return res
}
