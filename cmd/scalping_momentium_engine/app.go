package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/execution"
	"agent-economique/internal/shared"
	"agent-economique/internal/signals"
	momentium "agent-economique/internal/signals/scalping_momentium"
)

type ScalpingConfig struct {
	Timeframe         string
	ATRPeriod         int
	BodyPctMin        float64
	BodyATRMin        float64
	StochKPeriod      int
	StochKSmooth      int
	StochDPeriod      int
	StochKLongMax     float64
	StochKShortMin    float64
	TrailingCapPct    float64
	EnableStochCross  bool
	TrailingATRCoeff  float64
	// Optional filters
	EnableDMICross    bool
	DMIPeriod         int
	EnableMFIFilter   bool
	MFIPeriod         int
	MFIOversold       float64
	MFIOverbought     float64
	EnableCCIFilter   bool
	CCIPeriod         int
	CCIOversold       float64
	CCIOverbought     float64
}

func (app *ScalpingApp) writeLog(line string) {
    if app.logFile != nil {
        fmt.Fprintln(app.logFile, line)
    }
}

func (app *ScalpingApp) logSignal(s signals.Signal) {
    ts := s.Timestamp.Format(time.RFC3339)
    bodyPct := asFloat(s.Metadata["body_pct"]) * 100
    b2atr := asFloat(s.Metadata["body_to_atr"])
    kval := asFloat(s.Metadata["stoch_k"])
    line := fmt.Sprintf("[SIG] %s | %s | %s | price=%.6f | conf=%.2f | body=%.1f%% | b/atr=%.2f | K=%.2f",
        ts, s.Action, s.Type, s.Price, s.Confidence, bodyPct, b2atr, kval)
    app.writeLog(line)
}

func (app *ScalpingApp) hasTradesAvailable() bool {
    // Check existence of at least one trades file in cache
    cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
    if err != nil { return false }
    symbol := app.config.BinanceData.Symbols[0]
    for _, date := range app.dates {
        p := cache.GetFilePath(symbol, "trades", date)
        if _, err := os.Stat(p); err == nil { return true }
    }
    return false
}

// processTemporalTrades runs trade-by-trade; indicators use Vision klines window; trades drive trailing and minute markers
func (app *ScalpingApp) processTemporalTrades() error {
    cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
    if err != nil { return err }
    streamConfig := shared.StreamingConfig{}
    reader, err := binance.NewStreamingReader(cache, streamConfig)
    if err != nil { return err }

    symbol := app.config.BinanceData.Symbols[0]
    var currentMinute int64 = -1 // ms epoch truncated to minute

    // helper to process marker (minute boundary): use closed Vision kline at prevMinute
    processMarker := func(prevMinute int64) {
        idxPrev, ok := app.kIndex[prevMinute]
        if !ok || idxPrev < 299 || idxPrev >= len(app.klines)-1 {
            return
        }
        // Build window of 300 closed Vision klines up to idxPrev
        // plus one synthetic forming candle (to align generator lastClosedIdx on idxPrev)
        win := make([]signals.Kline, 0, 301)
        for j := idxPrev - 299; j <= idxPrev; j++ {
            k := app.klines[j]
            win = append(win, signals.Kline{
                OpenTime: time.Unix(0, k.Timestamp*1e6),
                Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume,
            })
        }
        // Append synthetic current forming candle (no look-ahead): openTime = prevMinute+60s, OHLC = last close
        nextOpenTime := app.klines[idxPrev+1].Timestamp
        lastClose := app.klines[idxPrev].Close
        win = append(win, signals.Kline{
            OpenTime: time.Unix(0, nextOpenTime*1e6),
            Open: lastClose, High: lastClose, Low: lastClose, Close: lastClose, Volume: 0,
        })
        // Use a fresh generator per marker to avoid lastProcessedIdx mismatch
        mg := momentium.NewGenerator(momentium.Config{
            ATRPeriod:    app.scalpCfg.ATRPeriod,
            BodyPctMin:   app.scalpCfg.BodyPctMin,
            BodyATRMin:   app.scalpCfg.BodyATRMin,
            StochKPeriod: app.scalpCfg.StochKPeriod,
            StochKSmooth: app.scalpCfg.StochKSmooth,
            StochDPeriod: app.scalpCfg.StochDPeriod,
            StochKLongMax:  app.scalpCfg.StochKLongMax,
            StochKShortMin: app.scalpCfg.StochKShortMin,
            EnableStochCross: app.scalpCfg.EnableStochCross,
            EnableDMICross:   app.scalpCfg.EnableDMICross,
            DMIPeriod:        app.scalpCfg.DMIPeriod,
            EnableMFIFilter:  app.scalpCfg.EnableMFIFilter,
            MFIPeriod:        app.scalpCfg.MFIPeriod,
            MFIOversold:      app.scalpCfg.MFIOversold,
            MFIOverbought:    app.scalpCfg.MFIOverbought,
            EnableCCIFilter:  app.scalpCfg.EnableCCIFilter,
            CCIPeriod:        app.scalpCfg.CCIPeriod,
            CCIOversold:      app.scalpCfg.CCIOversold,
            CCIOverbought:    app.scalpCfg.CCIOverbought,
        })
        _ = mg.Initialize(signals.GeneratorConfig{Symbol: app.config.BinanceData.Symbols[0], Timeframe: app.scalpCfg.Timeframe, HistorySize: 1000})
        if err := mg.CalculateIndicators(win); err != nil { return }
        sigs, err := mg.DetectSignals(win)
        if err != nil { return }
        // Log all signals like demo and collect for export
        for _, s := range sigs { app.logSignal(s) }
        app.signals = append(app.signals, sigs...)
        tsPrev := time.Unix(0, prevMinute*1e6)
        var exits, entries []signals.Signal
        for _, s := range sigs {
            if s.Timestamp.Equal(tsPrev) {
                if s.Action == signals.SignalActionExit { exits = append(exits, s) }
                if s.Action == signals.SignalActionEntry { entries = append(entries, s) }
            }
        }
        // EXIT at Vision close of idxPrev
        if app.currentPos != nil {
            for _, s := range exits {
                if s.Type != app.currentPos.Type { continue }
                exitPrice := app.klines[idxPrev].Close
                app.closePositionAt(tsPrev, exitPrice)
                break
            }
        }
        // ENTRY at Vision open of idxPrev+1 (after processing EXITs)
        if len(entries) > 0 {
            entryK := app.klines[idxPrev+1]
            entryTime := time.Unix(0, entryK.Timestamp*1e6)
            entryOpen := entryK.Open
            for _, s := range entries {
                if app.currentPos == nil {
                    // No open position -> open directly at next open
                    app.openPositionAt(s, entryTime, entryOpen)
                    break
                } else if s.Type != app.currentPos.Type {
                    // Opposite signal -> close current at prev close, then open new at next open
                    exitPx := app.klines[idxPrev].Close
                    app.closePositionAt(tsPrev, exitPx)
                    app.openPositionAt(s, entryTime, entryOpen)
                    break
                }
            }
        }
        // Signals already collected above
    }

    for _, date := range app.dates {
        tradesFile := cache.GetFilePath(symbol, "trades", date)
        // Stream trades if file exists
        if _, err := os.Stat(tradesFile); err != nil { continue }
        err := reader.StreamTrades(tradesFile, func(td shared.TradeData) error {
            // trailing intrabar
            if app.currentPos != nil {
                app.currentPos.Trail.Update(td.Price)
                if hit, stopPx := app.currentPos.Trail.Hit(td.Price); hit {
                    ts := time.Unix(0, td.Time*1e6)
                    app.forceClosePosition(ts, stopPx)
                }
            }
            // marker detection
            minute := td.Time - (td.Time % 60000)
            if currentMinute == -1 {
                currentMinute = minute
            }
            if minute != currentMinute {
                prev := currentMinute
                currentMinute = minute
                processMarker(prev)
            }
            return nil
        })
        if err != nil { return err }
        // Process last minute of the day if any trade was seen
        if currentMinute != -1 {
            processMarker(currentMinute)
        }
    }
    return nil
}

// (helper asFloat defined later)

func DefaultScalpingConfig() ScalpingConfig {
	return ScalpingConfig{
		Timeframe:      "1m",
		ATRPeriod:      3,
		BodyPctMin:     0.60,
		BodyATRMin:     0.60,
		StochKPeriod:   14,
		StochKSmooth:   3,
		StochDPeriod:   3,
		StochKLongMax:  40.0,
		StochKShortMin: 60.0,
		TrailingCapPct: 0.005,
		EnableStochCross: false,
		TrailingATRCoeff: 1.0,
		// Optional filters defaults
		EnableDMICross:  false,
		DMIPeriod:       14,
		EnableMFIFilter: false,
		MFIPeriod:       14,
		MFIOversold:     20.0,
		MFIOverbought:   80.0,
		EnableCCIFilter: false,
		CCIPeriod:       20,
		CCIOversold:     -100.0,
		CCIOverbought:   100.0,
	}
}

type Kline struct {
	Timestamp        int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	QuoteAssetVolume float64
}

type Position struct {
	Type       signals.SignalType
	EntryTime  time.Time
	EntryPrice float64
	Trail      *execution.Trailing
	ExitTime   *time.Time
	ExitPrice  *float64
	PnLPercent float64
	Duration   time.Duration
}

type ScalpingApp struct {
    config       *shared.Config
    dates        []string
    scalpCfg     ScalpingConfig
    generator    *momentium.Generator
    klines       []Kline
    signals      []signals.Signal
    currentPos   *Position
    closedPos    []Position
    kIndex       map[int64]int // map OpenTime(ms) -> index in klines
    outDir       string
    logFile      *os.File
    sumLong      float64
    sumShort     float64
}

func NewScalpingApp(config *shared.Config, dates []string) *ScalpingApp {
	cfg := DefaultScalpingConfig()
	// Override with YAML config if provided
	cm := config.Strategy.ScalpingMomentium
	if cm.Timeframe != "" { cfg.Timeframe = cm.Timeframe }
	if cm.ATRPeriod > 0 { cfg.ATRPeriod = cm.ATRPeriod }
	if cm.BodyPctMin > 0 { cfg.BodyPctMin = cm.BodyPctMin }
	if cm.BodyATRMin > 0 { cfg.BodyATRMin = cm.BodyATRMin }
	if cm.StochKPeriod > 0 { cfg.StochKPeriod = cm.StochKPeriod }
	if cm.StochKSmooth > 0 { cfg.StochKSmooth = cm.StochKSmooth }
	if cm.StochDPeriod > 0 { cfg.StochDPeriod = cm.StochDPeriod }
	if cm.StochKLongMax > 0 { cfg.StochKLongMax = cm.StochKLongMax }
	if cm.StochKShortMin > 0 { cfg.StochKShortMin = cm.StochKShortMin }
	if cm.TrailingCapPct > 0 { cfg.TrailingCapPct = cm.TrailingCapPct }
    cfg.EnableStochCross = cm.EnableStochCross
    if cm.TrailingATRCoeff > 0 { cfg.TrailingATRCoeff = cm.TrailingATRCoeff }
    // Optional filters from YAML
    cfg.EnableDMICross = cm.EnableDMICross
    if cm.DMIPeriod > 0 { cfg.DMIPeriod = cm.DMIPeriod }
    cfg.EnableMFIFilter = cm.EnableMFIFilter
    if cm.MFIPeriod > 0 { cfg.MFIPeriod = cm.MFIPeriod }
    if cm.MFIOversold > 0 { cfg.MFIOversold = cm.MFIOversold }
    if cm.MFIOverbought > 0 { cfg.MFIOverbought = cm.MFIOverbought }
    cfg.EnableCCIFilter = cm.EnableCCIFilter
    if cm.CCIPeriod > 0 { cfg.CCIPeriod = cm.CCIPeriod }
    // Note: CCIOversold can be negative; don't gate on >0
    cfg.CCIOversold = cm.CCIOversold
    if cm.CCIOverbought != 0 { cfg.CCIOverbought = cm.CCIOverbought }
    return &ScalpingApp{
        config:    config,
        dates:     dates,
        scalpCfg:  cfg,
        signals:   make([]signals.Signal, 0),
        closedPos: make([]Position, 0),
    }
}

func (app *ScalpingApp) Run() error {
	fmt.Println("\n📂 Chargement klines Binance Vision...")
	if err := app.loadKlines(); err != nil {
		return fmt.Errorf("chargement klines: %w", err)
	}
	fmt.Printf("✅ %d klines chargées\n", len(app.klines))

	if err := app.initializeGenerator(); err != nil {
		return fmt.Errorf("init générateur: %w", err)
	}

	// Build kline index

	// Prepare output folder and log file
	exportRoot := app.config.Backtest.ExportPath
	if exportRoot == "" { exportRoot = "backtest_results" }
	app.outDir = filepath.Join(exportRoot, "scalping_momentium_engine_"+time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(app.outDir, 0755); err != nil { return fmt.Errorf("mkdir outDir: %w", err) }
	lf, err := os.Create(filepath.Join(app.outDir, "engine.log"))
	if err != nil { return fmt.Errorf("create log: %w", err) }
	app.logFile = lf
	defer app.logFile.Close()
	// Log bundle/output directory path for user visibility
	fmt.Printf("\n📁 Dossier bundle: %s\n", app.outDir)

	// Backtest requires trade-by-trade cycle; no kline fallback when trades are missing
	if app.hasTradesAvailable() {
		fmt.Println("\n🔄 Exécution temporelle trade-par-trade avec marqueurs minute...")
		if err := app.processTemporalTrades(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("backtest requires trades (cycle_type=trades). No trade files found in cache for configured dates/symbol")
	}

	app.displayResults()
	if app.config.Backtest.ExportJSON {
		_ = app.exportResults()
		_ = app.exportBundle()
		// Remind where the bundle files (klines/positions/signals) were written
		fmt.Printf("📁 Dossier bundle: %s\n", app.outDir)
	}
	return nil
}

func (app *ScalpingApp) loadKlines() error {
    cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
    if err != nil { return err }
    streamConfig := shared.StreamingConfig{}
    reader, err := binance.NewStreamingReader(cache, streamConfig)
    if err != nil { return err }

    aggConfig := shared.AggregationConfig{}
    processor, err := binance.NewParsedDataProcessor(cache, reader, aggConfig)
    if err != nil { return err }

    symbol := app.config.BinanceData.Symbols[0]
    timeframe := app.scalpCfg.Timeframe
    app.klines = make([]Kline, 0, len(app.dates)*1500)

    for _, date := range app.dates {
        klinesFile := cache.GetFilePath(symbol, "klines", date, timeframe)
        batch, err := processor.ParseKlinesBatch(klinesFile, symbol, timeframe, date)
        if err != nil {
            fmt.Printf("  ⚠️  Skip date %s: %v\n", date, err)
            continue
        }
        for _, kd := range batch.KlinesData {
            app.klines = append(app.klines, Kline{
                Timestamp:        kd.OpenTime,
                Open:             kd.Open,
                High:             kd.High,
                Low:              kd.Low,
                Close:            kd.Close,
                Volume:           kd.Volume,
                QuoteAssetVolume: kd.QuoteAssetVolume,
            })
        }
    }
    // Build index of klines by open time (ms)
    app.kIndex = make(map[int64]int, len(app.klines))
    for i, k := range app.klines { app.kIndex[k.Timestamp] = i }
    return nil
}

func (app *ScalpingApp) initializeGenerator() error {
    cfg := momentium.Config{
        ATRPeriod:      app.scalpCfg.ATRPeriod,
        BodyPctMin:     app.scalpCfg.BodyPctMin,
        BodyATRMin:     app.scalpCfg.BodyATRMin,
        StochKPeriod:   app.scalpCfg.StochKPeriod,
        StochKSmooth:   app.scalpCfg.StochKSmooth,
        StochDPeriod:   app.scalpCfg.StochDPeriod,
        StochKLongMax:  app.scalpCfg.StochKLongMax,
        StochKShortMin: app.scalpCfg.StochKShortMin,
        EnableStochCross: app.scalpCfg.EnableStochCross,
        EnableDMICross:   app.scalpCfg.EnableDMICross,
        DMIPeriod:        app.scalpCfg.DMIPeriod,
        EnableMFIFilter:  app.scalpCfg.EnableMFIFilter,
        MFIPeriod:        app.scalpCfg.MFIPeriod,
        MFIOversold:      app.scalpCfg.MFIOversold,
        MFIOverbought:    app.scalpCfg.MFIOverbought,
        EnableCCIFilter:  app.scalpCfg.EnableCCIFilter,
        CCIPeriod:        app.scalpCfg.CCIPeriod,
        CCIOversold:      app.scalpCfg.CCIOversold,
        CCIOverbought:    app.scalpCfg.CCIOverbought,
    }
    app.generator = momentium.NewGenerator(cfg)
	return app.generator.Initialize(signals.GeneratorConfig{
		Symbol:    app.config.BinanceData.Symbols[0],
		Timeframe: app.scalpCfg.Timeframe,
		HistorySize: 1000,
	})
}

func (app *ScalpingApp) processLoop() error {
	// Indexer mapping timestamp->index pour accès signal
	// Fenêtre de 300 bougies maximum
	windowSize := 300
	for i := range app.klines {
		// Construire fenêtre [max(0,i-windowSize+1) .. i]
		start := i - windowSize + 1
		if start < 0 { start = 0 }
		win := make([]signals.Kline, 0, i-start+1)
		for j := start; j <= i; j++ {
			k := app.klines[j]
			win = append(win, signals.Kline{
				OpenTime: time.Unix(0, k.Timestamp*1e6),
				Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume,
			})
		}

		if err := app.generator.CalculateIndicators(win); err != nil { continue }
		sigs, err := app.generator.DetectSignals(win)
		if err != nil { continue }

		k := app.klines[i]
        // Logs: kline
        if app.config.Backtest.Logging.EnableKlineLogs {
            fmt.Printf("[KLINE] %s O:%.4f H:%.4f L:%.4f C:%.4f\n", time.Unix(0, k.Timestamp*1e6).Format(time.RFC3339), k.Open, k.High, k.Low, k.Close)
        }

        // 1) Traiter les signaux émis sur la bougie précédente (EXIT d'abord)
        var exits, entries []signals.Signal
        if i > 0 {
            prev := app.klines[i-1]
            prevTs := time.Unix(0, prev.Timestamp*1e6)
            for _, s := range sigs {
                if s.Timestamp.Equal(prevTs) {
                    if s.Action == signals.SignalActionExit { exits = append(exits, s) }
                    if s.Action == signals.SignalActionEntry { entries = append(entries, s) }
                }
            }
        }
        // Logs: signals
        if app.config.Backtest.Logging.EnableSignalLogs {
            if len(entries) > 0 || len(exits) > 0 {
                fmt.Printf("[SIG] %s entries:%d exits:%d\n", time.Unix(0, k.Timestamp*1e6).Format(time.RFC3339), len(entries), len(exits))
                for _, s := range append(entries, exits...) {
                    stochK := s.Metadata["stoch_k"]
                    atr := s.Metadata["atr"]
                    fmt.Printf("  - %s | %s | price=%.6f | K=%.2f | ATR=%.6f | conf=%.2f\n", s.Action, s.Type, s.Price, asFloat(stochK), asFloat(atr), s.Confidence)
                }
            }
        }
        // Collect for export
        if len(entries) > 0 { app.signals = append(app.signals, entries...) }
        if len(exits) > 0 { app.signals = append(app.signals, exits...) }

        // Appliquer d'abord les EXITs sur la position courante (au close de la bougie précédente)
        if app.currentPos != nil {
            for _, s := range exits {
                if s.Type != app.currentPos.Type { continue }
                app.closePositionAt(s.Timestamp, s.Price)
                break
            }
        }
        // Puis les ENTRY: ouvrir à l'open de la bougie courante
        for _, s := range entries {
            entryTime := time.Unix(0, k.Timestamp*1e6)
            entryOpen := k.Open
            if app.currentPos == nil {
                app.openPositionAt(s, entryTime, entryOpen)
            } else if s.Type != app.currentPos.Type {
                // Clôture au close précédent puis ouverture à l'open courant
                app.closePositionAt(s.Timestamp, s.Price)
                app.openPositionAt(s, entryTime, entryOpen)
            }
        }

        // 2) Mettre à jour le trailing avec le Close courant et vérifier stop
        if app.currentPos != nil {
            if app.config.Backtest.Logging.EnableTradeLogs {
                fmt.Printf("[TRAIL] before update trail=%.6f close=%.6f\n", app.currentPos.Trail.Trail, k.Close)
            }
            app.currentPos.Trail.Update(k.Close)
            if hit, stopPx := app.currentPos.Trail.Hit(k.Close); hit {
                if app.config.Backtest.Logging.EnableTradeLogs {
                    fmt.Printf("[TRAIL-HIT] stop=%.6f at %s\n", stopPx, time.Unix(0, k.Timestamp*1e6).Format(time.RFC3339))
                }
                app.forceClosePosition(time.Unix(0, k.Timestamp*1e6), stopPx)
            }
        }
    }
    return nil
}

func (app *ScalpingApp) openPositionAt(sig signals.Signal, entryTime time.Time, entryOpen float64) {
    atr := 0.0
    if v, ok := sig.Metadata["atr"].(float64); ok { atr = v }
    side := execution.SideShort
    if sig.Type == signals.SignalTypeLong { side = execution.SideLong }
    // Apply ATR coefficient for trailing sizing
    effATR := atr * app.scalpCfg.TrailingATRCoeff
    tr := execution.NewTrailing(side, entryOpen, effATR, app.scalpCfg.TrailingCapPct)
	app.currentPos = &Position{
		Type: sig.Type,
		EntryTime: entryTime,
		EntryPrice: entryOpen,
		Trail: tr,
	}
	    if app.config.Backtest.Logging.EnableTradeLogs {
        fmt.Printf("[OPEN] %s %s @ %.6f (atr=%.6f cap=%.4f%%)\n", entryTime.Format(time.RFC3339), sig.Type, entryOpen, atr, app.scalpCfg.TrailingCapPct*100)
    }
    app.writeLog(fmt.Sprintf("[OPEN] %s | %s @ %.6f | atr=%.6f cap=%.4f%%",
        entryTime.Format(time.RFC3339), sig.Type, entryOpen, atr, app.scalpCfg.TrailingCapPct*100))
}

func (app *ScalpingApp) closePositionAt(exitTime time.Time, exitPrice float64) {
    if app.currentPos == nil { return }
    app.currentPos.ExitTime = &exitTime
    app.currentPos.ExitPrice = &exitPrice
    app.currentPos.Duration = exitTime.Sub(app.currentPos.EntryTime)
    if app.currentPos.Type == signals.SignalTypeLong {
        app.currentPos.PnLPercent = (exitPrice - app.currentPos.EntryPrice) / app.currentPos.EntryPrice * 100
    } else {
        app.currentPos.PnLPercent = (app.currentPos.EntryPrice - exitPrice) / app.currentPos.EntryPrice * 100
    }
    // Capture raw/dir (SPEC): raw = Exit - Entry; dir = raw for LONG, -raw for SHORT
    captureRaw := exitPrice - app.currentPos.EntryPrice
    captureDir := captureRaw
    if app.currentPos.Type == signals.SignalTypeShort { captureDir = -captureRaw }
    if app.currentPos.Type == signals.SignalTypeLong { app.sumLong += captureRaw } else { app.sumShort += captureRaw }
    app.closedPos = append(app.closedPos, *app.currentPos)
    if app.config.Backtest.Logging.EnableTradeLogs {
        fmt.Printf("[CLOSE] %s %s @ %.6f | PnL=%.4f%% | dur=%s\n", exitTime.Format(time.RFC3339), app.closedPos[len(app.closedPos)-1].Type, exitPrice, app.closedPos[len(app.closedPos)-1].PnLPercent, app.closedPos[len(app.closedPos)-1].Duration)
    }
    app.writeLog(fmt.Sprintf("[CLOSE] %s | %s @ %.6f | raw=%.6f dir=%.6f | sumLong=%.6f sumShort=%.6f | sum=%.6f dirSum=%.6f",
        exitTime.Format(time.RFC3339), app.closedPos[len(app.closedPos)-1].Type, exitPrice,
        captureRaw, captureDir,
        app.sumLong, app.sumShort,
        app.sumLong+app.sumShort, app.sumLong+(-1*app.sumShort)))
    app.currentPos = nil
}

func (app *ScalpingApp) forceClosePosition(exitTime time.Time, exitPrice float64) {
    if app.currentPos == nil { return }
    app.currentPos.ExitTime = &exitTime
    app.currentPos.ExitPrice = &exitPrice
    app.currentPos.Duration = exitTime.Sub(app.currentPos.EntryTime)
    if app.currentPos.Type == signals.SignalTypeLong {
        app.currentPos.PnLPercent = (exitPrice - app.currentPos.EntryPrice) / app.currentPos.EntryPrice * 100
    } else {
        app.currentPos.PnLPercent = (app.currentPos.EntryPrice - exitPrice) / app.currentPos.EntryPrice * 100
    }
    // Capture raw/dir (SPEC) on force close
    captureRaw := exitPrice - app.currentPos.EntryPrice
    captureDir := captureRaw
    if app.currentPos.Type == signals.SignalTypeShort { captureDir = -captureRaw }
    if app.currentPos.Type == signals.SignalTypeLong { app.sumLong += captureRaw } else { app.sumShort += captureRaw }
    app.closedPos = append(app.closedPos, *app.currentPos)
    if app.config.Backtest.Logging.EnableTradeLogs {
        fmt.Printf("[FORCE-CLOSE] %s %s @ %.6f | PnL=%.4f%% | dur=%s\n", exitTime.Format(time.RFC3339), app.closedPos[len(app.closedPos)-1].Type, exitPrice, app.closedPos[len(app.closedPos)-1].PnLPercent, app.closedPos[len(app.closedPos)-1].Duration)
    }
    app.writeLog(fmt.Sprintf("[FORCE-CLOSE] %s | %s @ %.6f | raw=%.6f dir=%.6f | sumLong=%.6f sumShort=%.6f | sum=%.6f dirSum=%.6f",
        exitTime.Format(time.RFC3339), app.closedPos[len(app.closedPos)-1].Type, exitPrice,
        captureRaw, captureDir,
        app.sumLong, app.sumShort,
        app.sumLong+app.sumShort, app.sumLong+(-1*app.sumShort)))
    app.currentPos = nil
}

func (app *ScalpingApp) exportBundle() error {
    // klines.json
    type outK struct { Timestamp int64 `json:"t"`; Open, High, Low, Close, Volume float64 }
    ko := make([]outK, len(app.klines))
    for i, k := range app.klines { ko[i] = outK{Timestamp: k.Timestamp, Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume} }
    if err := writeJSON(filepath.Join(app.outDir, "klines.json"), ko); err != nil { return err }

    // positions.json with capture fields + cumulative sums per side (pct)
    type outP struct {
        Type       signals.SignalType `json:"type"`
        EntryTime  time.Time          `json:"entry_time"`
        EntryPrice float64            `json:"entry_price"`
        ExitTime   *time.Time         `json:"exit_time"`
        ExitPrice  *float64           `json:"exit_price"`
        PnLPercent float64            `json:"pnl_percent"`
        Duration   time.Duration      `json:"duration"`
        CaptureRaw float64            `json:"capture_raw"`
        CaptureDir float64            `json:"capture_dir"`
        CaptureRawPct float64         `json:"capture_raw_pct"`
        CaptureDirPct float64         `json:"capture_dir_pct"`
        SumLongCapturePct float64     `json:"sum_long_capture_pct"`
        SumShortCapturePct float64    `json:"sum_short_capture_pct"`
        SumLongDirCapturePct float64  `json:"sum_long_dir_capture_pct"`
        SumShortDirCapturePct float64 `json:"sum_short_dir_capture_pct"`
    }
    po := make([]outP, 0, len(app.closedPos))
    // running sums of raw and directional capture percent per side
    cumLongPct := 0.0
    cumShortPct := 0.0
    cumLongDirPct := 0.0
    cumShortDirPct := 0.0
    for _, p := range app.closedPos {
        var raw, dir float64
        var rawPct, dirPct float64
        if p.ExitPrice != nil {
            raw = *p.ExitPrice - p.EntryPrice
            dir = raw
            if p.Type == signals.SignalTypeShort { dir = -raw }
            if p.EntryPrice != 0 {
                rawPct = (raw / p.EntryPrice) * 100
                dirPct = rawPct
                if p.Type == signals.SignalTypeShort { dirPct = -rawPct }
            }
        }
        if p.Type == signals.SignalTypeLong {
            cumLongPct += rawPct
            cumLongDirPct += dirPct
        } else {
            cumShortPct += rawPct
            cumShortDirPct += dirPct
        }
        po = append(po, outP{
            Type: p.Type, EntryTime: p.EntryTime, EntryPrice: p.EntryPrice,
            ExitTime: p.ExitTime, ExitPrice: p.ExitPrice,
            PnLPercent: p.PnLPercent, Duration: p.Duration,
            CaptureRaw: raw, CaptureDir: dir,
            CaptureRawPct: rawPct, CaptureDirPct: dirPct,
            SumLongCapturePct: cumLongPct, SumShortCapturePct: cumShortPct,
            SumLongDirCapturePct: cumLongDirPct, SumShortDirCapturePct: cumShortDirPct,
        })
    }
    if err := writeJSON(filepath.Join(app.outDir, "positions.json"), po); err != nil { return err }

    // signals.json
    if err := writeJSON(filepath.Join(app.outDir, "signals.json"), app.signals); err != nil { return err }
    return nil
}

func writeJSON(path string, v interface{}) error {
    f, err := os.Create(path)
    if err != nil { return err }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(v)
}

func (app *ScalpingApp) displayResults() {
    fmt.Println("\n══════════════════════════════════════════════════════════════════════════")
    fmt.Println("  RÉSULTATS BACKTEST SCALPING MOMENTIUM")
    fmt.Println("══════════════════════════════════════════════════════════════════════════")
    fmt.Printf("Positions fermées: %d\n", len(app.closedPos))
    if len(app.closedPos) == 0 { return }
    wins := 0
    sumPct := 0.0
    nbLong, nbShort := 0, 0
    sumRawPct := 0.0
    sumDirPct := 0.0
    for _, p := range app.closedPos {
        if p.PnLPercent > 0 { wins++ }
        sumPct += p.PnLPercent
        if p.Type == signals.SignalTypeLong { nbLong++ } else { nbShort++ }
        if p.ExitPrice != nil && p.EntryPrice != 0 {
            rawPct := (*p.ExitPrice - p.EntryPrice) / p.EntryPrice * 100
            dirPct := rawPct
            if p.Type == signals.SignalTypeShort { dirPct = -rawPct }
            sumRawPct += rawPct
            sumDirPct += dirPct
        }
    }
    fmt.Printf("Nb positions : %d (LONG=%d, SHORT=%d)\n", len(app.closedPos), nbLong, nbShort)
    fmt.Printf("Somme Capture %% : %.2f%% | Somme Dir %%: %.2f%%\n", sumRawPct, sumDirPct)
    fmt.Printf("Win rate: %.1f%% | PnL moyen: %.2f%% | Total: %.2f%%\n", float64(wins)/float64(len(app.closedPos))*100, sumPct/float64(len(app.closedPos)), sumPct)
}

func (app *ScalpingApp) exportResults() error {
    exportPath := app.config.Backtest.ExportPath
    if exportPath == "" { exportPath = "backtest_results" }
    if err := os.MkdirAll(exportPath, 0755); err != nil { return err }
    filename := fmt.Sprintf("scalping_momentium_%s.json", time.Now().Format("20060102_150405"))
    fp := filepath.Join(exportPath, filename)
    file, err := os.Create(fp)
    if err != nil { return err }
    defer file.Close()
    total := len(app.closedPos)
    wins := 0
    sumPct := 0.0
    nbLong, nbShort := 0, 0
    // percentage-based sums
    sumRawPct := 0.0
    sumDirPct := 0.0
    for _, p := range app.closedPos {
        if p.PnLPercent > 0 { wins++ }
        sumPct += p.PnLPercent
        if p.Type == signals.SignalTypeLong { nbLong++ } else { nbShort++ }
        if p.ExitPrice != nil && p.EntryPrice != 0 {
            rawPct := (*p.ExitPrice - p.EntryPrice) / p.EntryPrice * 100
            dirPct := rawPct
            if p.Type == signals.SignalTypeShort { dirPct = -rawPct }
            sumRawPct += rawPct
            sumDirPct += dirPct
        }
    }
    winRate := 0.0
    avgPct := 0.0
    if total > 0 {
        winRate = float64(wins) / float64(total) * 100
        avgPct = sumPct / float64(total)
    }
    payload := map[string]interface{}{
        "timestamp": time.Now().Format(time.RFC3339),
        "positions": app.closedPos,
        "signals":   app.signals,
        "stats": map[string]interface{}{
            "total_positions": total,
            "long": nbLong,
            "short": nbShort,
            "win_rate": winRate,
            "pnl_avg_pct": avgPct,
            "pnl_total_pct": sumPct,
            // percentage-based captures (replacing unit-based sums)
            "capture_sum_pct": sumRawPct,
            "capture_dir_sum_pct": sumDirPct,
        },
    }
    enc := json.NewEncoder(file)
    enc.SetIndent("", "  ")
    if err := enc.Encode(payload); err != nil { return err }
    fmt.Printf("\n💾 Export: %s\n", fp)
    return nil
}

func generateDateRange(startStr, endStr string) ([]string, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, fmt.Errorf("date début invalide: %w", err)
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return nil, fmt.Errorf("date fin invalide: %w", err)
	}
	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates, nil
}

// asFloat converts interface{} to float64 when possible, otherwise returns 0
func asFloat(v interface{}) float64 {
    if v == nil {
        return 0
    }
    if f, ok := v.(float64); ok {
        return f
    }
    return 0
}
