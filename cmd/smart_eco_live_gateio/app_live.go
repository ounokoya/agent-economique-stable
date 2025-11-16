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

// Defaults (alignés à smart_eco app)
var (
    DEFAULT_TIMEFRAME          = "1m"
    DEFAULT_ATR_PERIOD         = 3
    DEFAULT_BODY_PCT_MIN       = 0.60
    DEFAULT_BODY_ATR_MIN       = 0.60
    DEFAULT_STOCH_K_PERIOD     = 14
    DEFAULT_STOCH_K_SMOOTH     = 3
    DEFAULT_STOCH_D_PERIOD     = 3
    DEFAULT_STOCH_K_LONG_MAX   = 40.0
    DEFAULT_STOCH_K_SHORT_MIN  = 60.0
    DEFAULT_ENABLE_STOCH_CROSS = false
    // VWMA
    DEFAULT_ENABLE_VWMA_CROSS = false
    DEFAULT_VWMA_FAST         = 6
    DEFAULT_VWMA_SLOW         = 36
    // Stoch extremes
    DEFAULT_ENABLE_STOCH_EXTREMES = false
    // DMI/MFI/CCI
    DEFAULT_ENABLE_DMI_CROSS  = false
    DEFAULT_DMI_PERIOD        = 14
    DEFAULT_ENABLE_MFI_FILTER = false
    DEFAULT_MFI_PERIOD        = 14
    DEFAULT_MFI_OVERSOLD      = 20.0
    DEFAULT_MFI_OVERBOUGHT    = 80.0
    DEFAULT_ENABLE_CCI_FILTER = false
    DEFAULT_CCI_PERIOD        = 20
    DEFAULT_CCI_OVERSOLD      = -100.0
    DEFAULT_CCI_OVERBOUGHT    = 100.0
    // MACD
    DEFAULT_ENABLE_MACD_HIST_FILTER  = false
    DEFAULT_ENABLE_MACD_SIGNE_FILTER = false
    DEFAULT_MACD_FAST                = 12
    DEFAULT_MACD_SLOW                = 26
    DEFAULT_MACD_SIGNAL              = 9
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

// default helpers
func ifZeroInt(v, def int) int {
    if v == 0 { return def }
    return v
}

func ifZeroFloat(v, def float64) float64 {
    if v == 0 { return def }
    return v
}

// SmartEcoLiveGateIOApp: live loop Gate.io utilisant le générateur smart_eco
type SmartEcoLiveGateIOApp struct {
    config *shared.Config
    // Données
    klines []Kline
    // Dernière bougie connue
    lastKnownTimestamp int64
    // Client Gate.io
    gateioClient *gateio.Client
    // Paramètres N
    initN   int
    updateN int
}

func NewSmartEcoLiveGateIOApp(config *shared.Config, initN, updateN int) *SmartEcoLiveGateIOApp {
    return &SmartEcoLiveGateIOApp{
        config:       config,
        klines:       make([]Kline, 0, 300),
        gateioClient: gateio.NewClient(),
        initN:        initN,
        updateN:      updateN,
    }
}

func (app *SmartEcoLiveGateIOApp) Run(ctx context.Context) error {
    // 1) Historique initial N (300) dernières klines
    n := app.initN
    if n <= 0 { n = 300 }
    if err := app.loadInitialKlines(n); err != nil {
        return err
    }
    fmt.Printf("✅ %d klines initiales chargées\n", len(app.klines))

    // 2) Boucle timer synchronisée (décalée +30s comme gateio live)
    if err := app.runTimerLoop(ctx); err != nil { return err }
    return nil
}

func (app *SmartEcoLiveGateIOApp) loadInitialKlines(limit int) error {
    symbol := app.config.BinanceData.Symbols[0]
    cm := app.config.Strategy.ScalpingMomentium
    timeframe := cm.Timeframe
    if timeframe == "" { timeframe = DEFAULT_TIMEFRAME }
    gateioSymbol := convertSymbolToGateIO(symbol)

    gk, err := app.gateioClient.GetKlines(context.Background(), gateioSymbol, timeframe, limit)
    if err != nil { return fmt.Errorf("Gate.io SDK error: %w", err) }

    app.klines = make([]Kline, len(gk))
    for i, k := range gk {
        app.klines[i] = Kline{
            Timestamp:        k.OpenTime.UnixMilli(),
            Open:             k.Open,
            High:             k.High,
            Low:              k.Low,
            Close:            k.Close,
            Volume:           k.Volume,
            QuoteAssetVolume: k.Volume * k.Close,
        }
    }
    if len(app.klines) > 0 { app.lastKnownTimestamp = app.klines[len(app.klines)-1].Timestamp }
    return nil
}

func (app *SmartEcoLiveGateIOApp) runTimerLoop(ctx context.Context) error {
    loopInterval := 10 * time.Second
    timeframe := app.config.Strategy.ScalpingMomentium.Timeframe
    if timeframe == "" { timeframe = DEFAULT_TIMEFRAME }

    now := time.Now()
    currentSecond := now.Second()
    var secondsUntilNext int
    switch {
    case currentSecond < 30:
        secondsUntilNext = 30 - currentSecond
    case currentSecond < 40:
        secondsUntilNext = 40 - currentSecond
    case currentSecond < 50:
        secondsUntilNext = 50 - currentSecond
    default:
        secondsUntilNext = 60 - currentSecond
    }
    nextSync := now.Add(time.Duration(secondsUntilNext) * time.Second).Truncate(time.Second)

    fmt.Printf("⏱️  Sync décalée +30s | now=%s next=%s in %ds | tf=%s\n", now.Format("15:04:05"), nextSync.Format("15:04:05"), secondsUntilNext, timeframe)

    select {
    case <-ctx.Done():
        return nil
    case <-time.After(time.Duration(secondsUntilNext) * time.Second):
        fmt.Printf("[%s] 🔔 Synchronisé!\n", time.Now().Format("15:04:05"))
        if err := app.processTimerTick(); err != nil { fmt.Printf("tick init err: %v\n", err) }
    }

    ticker := time.NewTicker(loopInterval)
    defer ticker.Stop()

    fmt.Printf("⏱️  Loop active (tick %v)\n", loopInterval)
    for {
        select {
        case <-ctx.Done():
            fmt.Println("🛑 Arrêt demandé")
            return nil
        case <-ticker.C:
            if err := app.processTimerTick(); err != nil { fmt.Printf("tick err: %v\n", err) }
        }
    }
}

func (app *SmartEcoLiveGateIOApp) processTimerTick() error {
    // Récupérer seulement les 10 dernières klines pour mise à jour
    n := app.updateN
    if n <= 0 { n = 10 }
    newK, err := app.fetchLatestKlines(n)
    if err != nil { return err }
    completed := app.detectNewCompletedCandles(newK)
    if len(completed) == 0 { return nil }
    for _, ts := range completed {
        // Traiter chaque bougie fermée
        if err := app.processMarker(ts); err != nil { fmt.Printf("process marker err: %v\n", err) }
    }
    return nil
}

func (app *SmartEcoLiveGateIOApp) fetchLatestKlines(limit int) ([]Kline, error) {
    symbol := app.config.BinanceData.Symbols[0]
    timeframe := app.config.Strategy.ScalpingMomentium.Timeframe
    if timeframe == "" { timeframe = DEFAULT_TIMEFRAME }
    gateioSymbol := convertSymbolToGateIO(symbol)

    gk, err := app.gateioClient.GetKlines(context.Background(), gateioSymbol, timeframe, limit)
    if err != nil { return nil, fmt.Errorf("SDK error: %w", err) }

    out := make([]Kline, len(gk))
    for i, k := range gk {
        out[i] = Kline{
            Timestamp:        k.OpenTime.UnixMilli(),
            Open:             k.Open,
            High:             k.High,
            Low:              k.Low,
            Close:            k.Close,
            Volume:           k.Volume,
            QuoteAssetVolume: k.Volume * k.Close,
        }
    }
    return out, nil
}

func (app *SmartEcoLiveGateIOApp) detectNewCompletedCandles(newKlines []Kline) []int64 {
    var completed []int64
    // Mettre à jour les existantes et détecter nouvelles
    for _, nk := range newKlines {
        updated := false
        for i := range app.klines {
            if app.klines[i].Timestamp == nk.Timestamp {
                app.klines[i] = nk // refresh
                updated = true
                break
            }
        }
        if !updated && nk.Timestamp > app.lastKnownTimestamp {
            completed = append(completed, nk.Timestamp)
            app.klines = append(app.klines, nk)
            if len(app.klines) > 300 { app.klines = app.klines[len(app.klines)-300:] }
            app.lastKnownTimestamp = nk.Timestamp
        }
    }
    return completed
}

func (app *SmartEcoLiveGateIOApp) processMarker(timestamp int64) error {
    // Construire fenêtre complète (jusqu'à 300 dernières)
    win := make([]signals.Kline, 0, len(app.klines))
    for _, k := range app.klines {
        win = append(win, signals.Kline{
            OpenTime: time.Unix(0, k.Timestamp*1e6),
            Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume,
        })
    }
    // Ajouter une bougie synthétique en formation (OHLC = dernier close) pour que lastClosedIdx = len(win)-2
    if len(app.klines) > 0 {
        last := app.klines[len(app.klines)-1]
        tf := app.config.Strategy.ScalpingMomentium.Timeframe
        if tf == "" { tf = DEFAULT_TIMEFRAME }
        nextOpenMs := last.Timestamp + tfMillis(tf)
        lastClose := last.Close
        win = append(win, signals.Kline{
            OpenTime: time.Unix(0, nextOpenMs*1e6),
            Open: lastClose, High: lastClose, Low: lastClose, Close: lastClose, Volume: 0,
        })
    }

    // Créer un générateur frais par marqueur (évite mismatch d'index)
    cm := app.config.Strategy.ScalpingMomentium
    g := smarteco.NewGenerator(smarteco.Config{
        ATRPeriod:                 ifZeroInt(cm.ATRPeriod, DEFAULT_ATR_PERIOD),
        BodyPctMin:                ifZeroFloat(cm.BodyPctMin, DEFAULT_BODY_PCT_MIN),
        BodyATRMin:                ifZeroFloat(cm.BodyATRMin, DEFAULT_BODY_ATR_MIN),
        StochKPeriod:              ifZeroInt(cm.StochKPeriod, DEFAULT_STOCH_K_PERIOD),
        StochKSmooth:              ifZeroInt(cm.StochKSmooth, DEFAULT_STOCH_K_SMOOTH),
        StochDPeriod:              ifZeroInt(cm.StochDPeriod, DEFAULT_STOCH_D_PERIOD),
        StochKLongMax:             ifZeroFloat(cm.StochKLongMax, DEFAULT_STOCH_K_LONG_MAX),
        StochKShortMin:            ifZeroFloat(cm.StochKShortMin, DEFAULT_STOCH_K_SHORT_MIN),
        EnableStochCross:          cm.EnableStochCross,
        EnableVwmaCross:           DEFAULT_ENABLE_VWMA_CROSS,
        VwmaFast:                  DEFAULT_VWMA_FAST,
        VwmaSlow:                  DEFAULT_VWMA_SLOW,
        EnableStochExtremes:       DEFAULT_ENABLE_STOCH_EXTREMES,
        EnableDMICross:            cm.EnableDMICross,
        DMIPeriod:                 ifZeroInt(cm.DMIPeriod, DEFAULT_DMI_PERIOD),
        EnableMFIFilter:           cm.EnableMFIFilter,
        MFIPeriod:                 ifZeroInt(cm.MFIPeriod, DEFAULT_MFI_PERIOD),
        MFIOversold:               ifZeroFloat(cm.MFIOversold, DEFAULT_MFI_OVERSOLD),
        MFIOverbought:             ifZeroFloat(cm.MFIOverbought, DEFAULT_MFI_OVERBOUGHT),
        EnableCCIFilter:           cm.EnableCCIFilter,
        CCIPeriod:                 ifZeroInt(cm.CCIPeriod, DEFAULT_CCI_PERIOD),
        CCIOversold:               cm.CCIOversold,
        CCIOverbought:             ifZeroFloat(cm.CCIOverbought, DEFAULT_CCI_OVERBOUGHT),
        EnableMacdHistogramFilter: DEFAULT_ENABLE_MACD_HIST_FILTER,
        EnableMacdSigneFilter:     DEFAULT_ENABLE_MACD_SIGNE_FILTER,
        MacdFast:                  DEFAULT_MACD_FAST,
        MacdSlow:                  DEFAULT_MACD_SLOW,
        MacdSignalPeriod:          DEFAULT_MACD_SIGNAL,
    })

    cm = app.config.Strategy.ScalpingMomentium
    tf := cm.Timeframe
    if tf == "" { tf = DEFAULT_TIMEFRAME }
    _ = g.Initialize(signals.GeneratorConfig{Symbol: app.config.BinanceData.Symbols[0], Timeframe: tf, HistorySize: 1000})

    if err := g.CalculateIndicators(win); err != nil { return err }
    sigs, err := g.DetectSignals(win)
    if err != nil { return err }

    // Afficher les signaux sur la bougie qui vient de se fermer
    tsPrev := time.Unix(0, timestamp*1e6)
    for _, s := range sigs {
        if s.Timestamp.Equal(tsPrev) {
            fmt.Printf("[SIG] %s | %s | price=%.6f | conf=%.2f\n", s.Timestamp.Format(time.RFC3339), s.Type, s.Price, s.Confidence)
        }
    }
    return nil
}

// Helpers
func convertSymbolToGateIO(symbol string) string {
    if strings.HasSuffix(symbol, "USDT") {
        base := strings.TrimSuffix(symbol, "USDT")
        return base + "_USDT"
    }
    return symbol
}

// tfMillis retourne la durée en millisecondes pour un timeframe string
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
        return 300_000 // par défaut 5m
    }
}
