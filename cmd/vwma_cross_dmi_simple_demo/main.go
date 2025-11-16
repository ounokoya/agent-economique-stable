package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/signals"
	"agent-economique/internal/signals/vwma_cross_dmi_simple"
)

// ============================================================================
// CONSTANTES
// ============================================================================

// Symbole et donnees
const (
	SYMBOL     = "SOL_USDT" // Format Gate.io
	TIMEFRAME  = "5m"
	NB_CANDLES = 1000
)

// Paramètres VWMA Cross + DMI Simple
const (
	VWMA_SHORT_PERIOD = 4  // VWMA court
	VWMA_LONG_PERIOD  = 12 // VWMA long
	DMI_PERIOD        = 14 // Période DMI standard
	DMI_SMOOTH        = 6  // Lissage DMI
	WINDOW_MATCHING   = 5  // Fenêtre matching 3 conditions
)

// ============================================================================
// STRUCTURES
// ============================================================================

type Config struct {
	Symbol          string
	Timeframe       string
	NbCandles       int
	VWMAShortPeriod int
	VWMALongPeriod  int
	DMIPeriod       int
	DMISmooth       int
	WindowMatching  int
}

// SignalInfo représente un signal détecté avec contexte
type SignalInfo struct {
	Index         int
	Timestamp     time.Time
	Action        string
	Type          string
	Mode          string
	Price         float64
	Confidence    float64
	PositionAvant string
	PositionApres string
	VWMADirection string
	DIDominant    string
	DXCrossDir    string
}

// TradeInfo représente un trade complet (entry + exit)
type TradeInfo struct {
	Numero          int
	Type            string
	EntryTime       time.Time
	ExitTime        time.Time
	EntryPrice      float64
	ExitPrice       float64
	Duration        int
	Variation       float64
	EntryMode       string
	ExitMode        string
	EntryConfidence float64
	ExitConfidence  float64
}

// ============================================================================
// DÉTECTION DE SIGNAUX
// ============================================================================

func detectSignals(config Config, klines []gateio.Kline) ([]SignalInfo, error) {
	// Convertir klines en format signals.Kline
	signalsKlines := make([]signals.Kline, len(klines))
	for i, k := range klines {
		signalsKlines[i] = signals.Kline{
			OpenTime: k.OpenTime,
			Close:    k.Close,
			High:     k.High,
			Low:      k.Low,
			Volume:   k.Volume,
		}
	}

	// Créer le générateur
	generatorConfig := vwma_cross_dmi_simple.Config{
		VWMAShortPeriod: config.VWMAShortPeriod,
		VWMALongPeriod:  config.VWMALongPeriod,
		DMIPeriod:       config.DMIPeriod,
		DMISmooth:       config.DMISmooth,
		WindowMatching:  config.WindowMatching,
	}

	generator := vwma_cross_dmi_simple.NewVWMACrossDMISimpleGenerator(generatorConfig)

	// Initialiser
	err := generator.Initialize(signals.GeneratorConfig{
		Symbol:    config.Symbol,
		Timeframe: config.Timeframe,
	})
	if err != nil {
		return nil, fmt.Errorf("erreur initialisation générateur: %v", err)
	}

	// Calculer les indicateurs
	err = generator.CalculateIndicators(signalsKlines)
	if err != nil {
		return nil, fmt.Errorf("erreur calcul indicateurs: %v", err)
	}

	// Détecter les signaux
	signalsDetected, err := generator.DetectSignals(signalsKlines)
	if err != nil {
		return nil, fmt.Errorf("erreur détection signaux: %v", err)
	}

	// Convertir en SignalInfo
	var signalInfos []SignalInfo
	currentPosition := signals.SignalTypeLong // Forcer première entrée comme LONG

	for _, sig := range signalsDetected {
		positionAvant := string(currentPosition)

		// Mettre à jour position selon signal
		if sig.Action == signals.SignalActionEntry {
			currentPosition = sig.Type
		} else if sig.Action == signals.SignalActionExit {
			// Après exit, la prochaine position sera l'inverse du signal suivant
			// Pour le suivi, on considère qu'on est "neutral"
			positionAvant = "NEUTRAL"
		}

		positionApres := string(currentPosition)

		signalInfo := SignalInfo{
			Index:         0, // Sera rempli plus tard
			Timestamp:     sig.Timestamp,
			Action:        string(sig.Action),
			Type:          string(sig.Type),
			Mode:          sig.Metadata["mode"].(string),
			Price:         sig.Price,
			Confidence:    sig.Confidence,
			PositionAvant: positionAvant,
			PositionApres: positionApres,
			VWMADirection: sig.Metadata["vwma_direction"].(string),
			DIDominant:    sig.Metadata["di_dominant"].(string),
			DXCrossDir:    sig.Metadata["dx_cross_direction"].(string),
		}

		signalInfos = append(signalInfos, signalInfo)
	}

	return signalInfos, nil
}

// ============================================================================
// ANALYSE DES TRADES
// ============================================================================

func analyzeTrades(signals []SignalInfo) []TradeInfo {
	var trades []TradeInfo
	var currentTrade *TradeInfo
	tradeNumber := 0

	for i, sig := range signals {
		sig.Index = i // Remplir l'index

		if sig.Action == "ENTRY" {
			// Démarrer un nouveau trade
			if currentTrade == nil {
				tradeNumber++

				// DÉTERMINER LE TYPE selon le signal d'entrée
				var tradeType string
				if string(sig.Type) == "LONG" {
					tradeType = "LONG"
				} else {
					tradeType = "SHORT"
				}

				currentTrade = &TradeInfo{
					Numero:          tradeNumber,
					Type:            tradeType, // Type fixé par le signal
					EntryTime:       sig.Timestamp,
					EntryPrice:      sig.Price,
					EntryMode:       sig.Mode,
					EntryConfidence: sig.Confidence,
				}
			}
		} else if sig.Action == "EXIT" && currentTrade != nil {
			// Fermer le trade courant
			if currentTrade.Type != sig.Type {
				// Exit correct (type inverse)
				currentTrade.ExitTime = sig.Timestamp
				currentTrade.ExitPrice = sig.Price
				currentTrade.ExitMode = sig.Mode
				currentTrade.ExitConfidence = sig.Confidence
				currentTrade.Duration = int(sig.Timestamp.Sub(currentTrade.EntryTime).Minutes())

				// Calculer variation (sans modifier le type)
				currentTrade.Variation = (currentTrade.ExitPrice - currentTrade.EntryPrice) / currentTrade.EntryPrice * 100

				trades = append(trades, *currentTrade)
				currentTrade = nil
			}
		}
	}

	return trades
}

// ============================================================================
// AFFICHAGE
// ============================================================================

func displaySignals(signals []SignalInfo, limit int) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Println("SIGNAUX DÉTECTÉS (Derniers signaux)")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s | %-19s | %-6s | %-5s | %-14s | %-8s | %-8s | %-12s | %-12s | %-12s\n",
		"#", "Date/Heure", "Action", "Type", "Mode", "Price", "Conf%", "Pos Avant", "Pos Après", "VWMA/DI/DX")
	fmt.Println(strings.Repeat("-", 120))

	start := len(signals) - limit
	if start < 0 {
		start = 0
	}

	for i := start; i < len(signals); i++ {
		sig := signals[i]

		actionSymbol := "→"
		if sig.Action == "ENTRY" {
			actionSymbol = "↗"
		} else if sig.Action == "EXIT" {
			actionSymbol = "↘"
		}

		typeSymbol := sig.Type
		if sig.Type == "LONG" {
			typeSymbol = "🟢 LONG"
		} else if sig.Type == "SHORT" {
			typeSymbol = "🔴 SHORT"
		}

		vwmaSymbol := "↗"
		if sig.VWMADirection == "DOWN" {
			vwmaSymbol = "↘"
		}

		diSymbol := "+"
		if sig.DIDominant == "DI_MINUS" {
			diSymbol = "-"
		}

		dxSymbol := "↗"
		if sig.DXCrossDir == "DOWN" {
			dxSymbol = "↘"
		}

		fmt.Printf("%-4d | %s | %-6s | %-5s | %-14s | %8.2f | %7.1f%% | %-12s | %-12s | %s%s%s\n",
			sig.Index,
			sig.Timestamp.Format("2006-01-02 15:04:05"),
			actionSymbol+" "+sig.Action,
			typeSymbol,
			sig.Mode,
			sig.Price,
			sig.Confidence*100,
			sig.PositionAvant,
			sig.PositionApres,
			vwmaSymbol, diSymbol, dxSymbol)
	}

	fmt.Println(strings.Repeat("=", 120))
}

func displayTrades(trades []TradeInfo) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println("TRADES COMPLETS (Entry + Exit)")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-6s | %-19s | %-19s | %-8s | %-8s | %-8s | %-10s | %-14s | %-14s\n",
		"#", "Type", "Entry Time", "Exit Time", "Entry $", "Exit $", "Dur(min)", "Var %", "Entry Mode", "Exit Mode")
	fmt.Println(strings.Repeat("-", 140))

	totalVariation := 0.0
	gagnants := 0
	perdants := 0

	// Calculer totaux séparés pour LONG et SHORT
	totalLong := 0.0
	totalShort := 0.0
	longTrades := 0
	shortTrades := 0

	for _, trade := range trades {
		varSymbol := "📈"
		if trade.Variation < 0 {
			varSymbol = "📉"
			perdants++
		} else {
			gagnants++
		}

		fmt.Printf("%-4d | %-6s | %s | %s | %8.2f | %8.2f | %8d | %s%+8.2f%% | %-14s | %-14s\n",
			trade.Numero,
			trade.Type,
			trade.EntryTime.Format("2006-01-02 15:04:05"),
			trade.ExitTime.Format("2006-01-02 15:04:05"),
			trade.EntryPrice,
			trade.ExitPrice,
			trade.Duration,
			varSymbol,
			trade.Variation,
			trade.EntryMode,
			trade.ExitMode)

		// Totaliser séparément LONG et SHORT
		if trade.Type == "LONG" {
			totalLong += trade.Variation
			longTrades++
			fmt.Printf("DEBUG: LONG trade #%d: variation=%.2f, totalLong=%.2f\n", trade.Numero, trade.Variation, totalLong)
		} else if trade.Type == "SHORT" {
			totalShort += trade.Variation
			shortTrades++
			fmt.Printf("DEBUG: SHORT trade #%d: variation=%.2f, totalShort=%.2f\n", trade.Numero, trade.Variation, totalShort)
		} else {
			fmt.Printf("DEBUG: UNKNOWN trade type: '%s' for trade #%d\n", trade.Type, trade.Numero)
		}
	}

	// Totalisation correcte : LONG + (-SHORT)
	totalVariation = totalLong + (-totalShort)

	fmt.Println(strings.Repeat("=", 140))

	// Statistiques
	fmt.Printf("\nSTATISTIQUES TRADES:\n")
	fmt.Printf("  Total trades        : %d\n", len(trades))
	fmt.Printf("  - Trades LONG       : %d\n", longTrades)
	fmt.Printf("  - Trades SHORT      : %d\n", shortTrades)
	fmt.Printf("  Gagnants (📈)       : %d (%.1f%%)\n", gagnants, float64(gagnants)/float64(len(trades))*100)
	fmt.Printf("  Perdants (📉)       : %d (%.1f%%)\n", perdants, float64(perdants)/float64(len(trades))*100)
	fmt.Printf("  Variation LONG      : %+.2f%%\n", totalLong)
	fmt.Printf("  Variation SHORT     : %+.2f%% (gain réel = %+.2f%%)\n", totalShort, -totalShort)
	fmt.Printf("  Variation totale    : %+.2f%%\n", totalVariation)
	if len(trades) > 0 {
		fmt.Printf("  Variation moyenne   : %+.2f%% par trade\n", totalVariation/float64(len(trades)))
	}

	fmt.Println(strings.Repeat("=", 140))
}

func displayStatistics(signals []SignalInfo, trades []TradeInfo) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES GLOBALES")
	fmt.Println(strings.Repeat("=", 100))

	// Stats signaux
	totalSignals := len(signals)
	entrySignals := 0
	exitSignals := 0
	longSignals := 0
	shortSignals := 0
	trendSignals := 0
	counterTrendSignals := 0

	for _, sig := range signals {
		if sig.Action == "ENTRY" {
			entrySignals++
		} else {
			exitSignals++
		}

		if sig.Type == "LONG" {
			longSignals++
		} else {
			shortSignals++
		}

		if sig.Mode == "TREND" {
			trendSignals++
		} else {
			counterTrendSignals++
		}
	}

	fmt.Printf("\nSIGNAUX:\n")
	fmt.Printf("  Total signaux       : %d\n", totalSignals)
	fmt.Printf("  - Entrées (↗)       : %d\n", entrySignals)
	fmt.Printf("  - Sorties (↘)       : %d\n", exitSignals)
	fmt.Printf("  - LONG (🟢)         : %d\n", longSignals)
	fmt.Printf("  - SHORT (🔴)        : %d\n", shortSignals)
	fmt.Printf("  - Tendance          : %d\n", trendSignals)
	fmt.Printf("  - Contre-tendance   : %d\n", counterTrendSignals)

	// Stats trades
	if len(trades) > 0 {
		var totalDuration int
		maxDuration := 0
		minDuration := trades[0].Duration

		for _, trade := range trades {
			totalDuration += trade.Duration
			if trade.Duration > maxDuration {
				maxDuration = trade.Duration
			}
			if trade.Duration < minDuration {
				minDuration = trade.Duration
			}
		}

		fmt.Printf("\nTRADES:\n")
		fmt.Printf("  Durée moyenne       : %d minutes\n", totalDuration/len(trades))
		fmt.Printf("  Durée min/max       : %d / %d minutes\n", minDuration, maxDuration)
	}

	fmt.Println(strings.Repeat("=", 100))
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	config := Config{
		Symbol:          SYMBOL,
		Timeframe:       TIMEFRAME,
		NbCandles:       NB_CANDLES,
		VWMAShortPeriod: VWMA_SHORT_PERIOD,
		VWMALongPeriod:  VWMA_LONG_PERIOD,
		DMIPeriod:       DMI_PERIOD,
		DMISmooth:       DMI_SMOOTH,
		WindowMatching:  WINDOW_MATCHING,
	}

	fmt.Println("=== DEMO VWMA CROSS + DMI SIMPLE ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange              : Gate.io\n")
	fmt.Printf("Symbole               : %s\n", config.Symbol)
	fmt.Printf("Timeframe             : %s\n", config.Timeframe)
	fmt.Printf("VWMA court/long       : %d / %d\n", config.VWMAShortPeriod, config.VWMALongPeriod)
	fmt.Printf("DMI période/smooth    : %d / %d\n", config.DMIPeriod, config.DMISmooth)
	fmt.Printf("Fenêtre matching      : %d bougies\n", config.WindowMatching)

	// Récupération des données
	fmt.Printf("\nRécupération des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, config.Symbol, config.Timeframe, config.NbCandles)
	if err != nil {
		log.Fatalf("Erreur récupération klines: %v", err)
	}

	// Tri chronologique
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime.Before(klines[j].OpenTime)
	})

	fmt.Printf("Klines récupérées: %d\n", len(klines))

	// Détection des signaux
	fmt.Println("\nDétection des signaux...")
	signals, err := detectSignals(config, klines)
	if err != nil {
		log.Fatalf("Erreur détection signaux: %v", err)
	}

	fmt.Printf("Signaux détectés: %d\n", len(signals))

	// Analyse des trades
	fmt.Println("Analyse des trades...")
	trades := analyzeTrades(signals)
	fmt.Printf("Trades complets: %d\n", len(trades))

	// Affichage
	displaySignals(signals, 20) // Afficher les 20 derniers signaux
	displayTrades(trades)
	displayStatistics(signals, trades)

	fmt.Println("\n=== FIN DEMO ===")
}
