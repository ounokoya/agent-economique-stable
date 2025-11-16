package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/shared"
	"agent-economique/internal/signals"
	smarteco "agent-economique/internal/signals/smart_eco"
)

// Kline local
type Kline struct {
	Timestamp        int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	QuoteAssetVolume float64
}

// Defaults alignés à smart_eco
var (
	DEFAULT_TIMEFRAME          = "1m"
	DEFAULT_ATR_PERIOD         = 3
	DEFAULT_BODY_PCT_MIN       = 0.60
	DEFAULT_BODY_ATR_MIN       = 0.90
	DEFAULT_STOCH_K_PERIOD     = 14
	DEFAULT_STOCH_K_SMOOTH     = 3
	DEFAULT_STOCH_D_PERIOD     = 3
	DEFAULT_STOCH_K_LONG_MAX   = 40.0
	DEFAULT_STOCH_K_SHORT_MIN  = 60.0
	DEFAULT_ENABLE_STOCH_CROSS = false
	// VWMA/MACD: pas exposés dans YAML actuel
	DEFAULT_ENABLE_VWMA_CROSS        = false
	DEFAULT_VWMA_FAST                = 6
	DEFAULT_VWMA_SLOW                = 36
	DEFAULT_ENABLE_STOCH_EXTREMES    = false
	DEFAULT_ENABLE_DMI_CROSS         = false
	DEFAULT_DMI_PERIOD               = 14
	DEFAULT_ENABLE_MFI_FILTER        = false
	DEFAULT_MFI_PERIOD               = 14
	DEFAULT_MFI_OVERSOLD             = 20.0
	DEFAULT_MFI_OVERBOUGHT           = 80.0
	DEFAULT_ENABLE_CCI_FILTER        = false
	DEFAULT_CCI_PERIOD               = 20
	DEFAULT_CCI_OVERSOLD             = -100.0
	DEFAULT_CCI_OVERBOUGHT           = 100.0
	DEFAULT_ENABLE_MACD_HIST_FILTER  = false
	DEFAULT_ENABLE_MACD_SIGNE_FILTER = false
	DEFAULT_MACD_FAST                = 12
	DEFAULT_MACD_SLOW                = 26
	DEFAULT_MACD_SIGNAL              = 9
)

// Helpers
func ifZeroInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
func ifZeroFloat(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

// App
type SmartEcoDemoApp struct {
	cfg    *shared.Config
	n      int
	klines []Kline
	gate   *gateio.Client
}

func NewSmartEcoDemoApp(cfg *shared.Config, n int) *SmartEcoDemoApp {
	if n <= 0 {
		n = 1000
	}
	return &SmartEcoDemoApp{cfg: cfg, n: n, klines: make([]Kline, 0, n), gate: gateio.NewClient()}
}

func (a *SmartEcoDemoApp) Run() error {
	if err := a.loadKlines(); err != nil {
		return err
	}
	fmt.Printf("✅ %d klines chargées (dernieres)\n", len(a.klines))

	// Préparer la config SmartEco UNIQUEMENT depuis les constantes locales (ignore YAML)
	cfg := smarteco.Config{
		ATRPeriod:                 DEFAULT_ATR_PERIOD,
		BodyPctMin:                DEFAULT_BODY_PCT_MIN,
		BodyATRMin:                DEFAULT_BODY_ATR_MIN,
		StochKPeriod:              DEFAULT_STOCH_K_PERIOD,
		StochKSmooth:              DEFAULT_STOCH_K_SMOOTH,
		StochDPeriod:              DEFAULT_STOCH_D_PERIOD,
		StochKLongMax:             DEFAULT_STOCH_K_LONG_MAX,
		StochKShortMin:            DEFAULT_STOCH_K_SHORT_MIN,
		EnableStochCross:          DEFAULT_ENABLE_STOCH_CROSS,
		EnableVwmaCross:           DEFAULT_ENABLE_VWMA_CROSS,
		VwmaFast:                  DEFAULT_VWMA_FAST,
		VwmaSlow:                  DEFAULT_VWMA_SLOW,
		EnableStochExtremes:       DEFAULT_ENABLE_STOCH_EXTREMES,
		EnableDMICross:            DEFAULT_ENABLE_DMI_CROSS,
		DMIPeriod:                 DEFAULT_DMI_PERIOD,
		EnableMFIFilter:           DEFAULT_ENABLE_MFI_FILTER,
		MFIPeriod:                 DEFAULT_MFI_PERIOD,
		MFIOversold:               DEFAULT_MFI_OVERSOLD,
		MFIOverbought:             DEFAULT_MFI_OVERBOUGHT,
		EnableCCIFilter:           DEFAULT_ENABLE_CCI_FILTER,
		CCIPeriod:                 DEFAULT_CCI_PERIOD,
		CCIOversold:               DEFAULT_CCI_OVERSOLD,
		CCIOverbought:             DEFAULT_CCI_OVERBOUGHT,
		EnableMacdHistogramFilter: DEFAULT_ENABLE_MACD_HIST_FILTER,
		EnableMacdSigneFilter:     DEFAULT_ENABLE_MACD_SIGNE_FILTER,
		MacdFast:                  DEFAULT_MACD_FAST,
		MacdSlow:                  DEFAULT_MACD_SLOW,
		MacdSignalPeriod:          DEFAULT_MACD_SIGNAL,
	}

	tf := DEFAULT_TIMEFRAME
	// Boucle sur les bougies: rolling window et détection à chaque i
	// Fenêtre max 300 (comme backtest)
	windowSize := 300
	totalSignals := 0
	for i := range a.klines {
		start := i - windowSize + 1
		if start < 0 {
			start = 0
		}
		win := make([]signals.Kline, 0, i-start+2)
		for j := start; j <= i; j++ {
			k := a.klines[j]
			win = append(win, signals.Kline{
				OpenTime: time.Unix(0, k.Timestamp*1e6),
				Open:     k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume,
			})
		}
		// bougie synthétique (forming) pour aligner lastClosedIdx = len-2
		if len(win) > 0 {
			last := a.klines[i]
			nextOpenMs := last.Timestamp + tfMillis(tf)
			lastClose := last.Close
			win = append(win, signals.Kline{OpenTime: time.Unix(0, nextOpenMs*1e6), Open: lastClose, High: lastClose, Low: lastClose, Close: lastClose, Volume: 0})
		}

		// IMPORTANT: créer un générateur frais par itération pour éviter les mismatches d'index
		g := smarteco.NewGenerator(cfg)
		_ = g.Initialize(signals.GeneratorConfig{Symbol: a.cfg.BinanceData.Symbols[0], Timeframe: tf, HistorySize: 1000})
		if err := g.CalculateIndicators(win); err != nil {
			continue
		}
		sigs, err := g.DetectSignals(win)
		if err != nil {
			continue
		}

		// Afficher uniquement les signaux émis sur la bougie précédente
		if i > 0 {
			prevTs := time.Unix(0, a.klines[i-1].Timestamp*1e6)
			for _, s := range sigs {
				if s.Timestamp.Equal(prevTs) {
					totalSignals++
					fmt.Printf("[SIG] %s | %s | price=%.6f | conf=%.2f\n", s.Timestamp.Format(time.RFC3339), s.Type, s.Price, s.Confidence)
				}
			}
		}
	}

	fmt.Printf("\nRésumé: %d signaux sur %d bougies\n", totalSignals, len(a.klines))
	return nil
}

func (a *SmartEcoDemoApp) loadKlines() error {
	symbol := a.cfg.BinanceData.Symbols[0]
	tf := DEFAULT_TIMEFRAME
	gsym := convertSymbolToGateIO(symbol)

	gk, err := a.gate.GetKlines(context.Background(), gsym, tf, a.n)
	if err != nil {
		return fmt.Errorf("Gate.io SDK error: %w", err)
	}

	a.klines = make([]Kline, len(gk))
	for i, k := range gk {
		a.klines[i] = Kline{
			Timestamp:        k.OpenTime.UnixMilli(),
			Open:             k.Open,
			High:             k.High,
			Low:              k.Low,
			Close:            k.Close,
			Volume:           k.Volume,
			QuoteAssetVolume: k.Volume * k.Close,
		}
	}
	return nil
}

func convertSymbolToGateIO(symbol string) string {
	if strings.HasSuffix(symbol, "USDT") {
		base := strings.TrimSuffix(symbol, "USDT")
		return base + "_USDT"
	}
	return symbol
}

func tfMillis(tf string) int64 {
	switch tf {
	case "1m":
		return 60_000
	case "3m":
		return 180_000
	case "5m":
		return 300_000
	case "15m":
		return 900_000
	case "30m":
		return 1_800_000
	case "1h":
		return 3_600_000
	case "2h":
		return 7_200_000
	case "4h":
		return 14_400_000
	case "1d":
		return 86_400_000
	default:
		return 300_000
	}
}
