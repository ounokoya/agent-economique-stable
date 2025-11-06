package main

import (
	"context"
	"fmt"
	"time"

	"agent-economique/internal/engine"
)

// SimulationEngine orchestrates the backtest simulation using temporal engine
type SimulationEngine struct {
	temporalEngine *engine.TemporalEngine
	dataset        *HistoricalDataSet
	config         *AppConfig
	
	// Simulation state
	currentTradeIndex int
	simulationSpeed   time.Duration // Time between trades
	isRunning         bool
	isPaused          bool
	
	// Statistics
	startTime          time.Time
	totalTrades        int
	processedTrades    int
	tradesPerSecond    float64
	estimatedTimeLeft  time.Duration
	
	// Daily progress tracking
	currentDay         time.Time
	dailyStats         map[string]*DayStats
}

// DayStats tracks statistics for each trading day
type DayStats struct {
	Date            time.Time
	TradesProcessed int
	PriceStart      float64
	PriceEnd        float64
	PriceChange     float64
	PositionsOpened int
	PositionsClosed int
	PnL             float64
}

// SimulationConfig configures simulation parameters
type SimulationConfig struct {
	Speed           SimulationSpeed
	LogInterval     time.Duration
	ShowProgress    bool
	DailyReports    bool
	PauseOnSignals  bool
}

// SimulationSpeed defines how fast the simulation runs
type SimulationSpeed int

const (
	SpeedRealtime SimulationSpeed = iota // 1:1 real time speed
	SpeedFast                           // 10x faster
	SpeedMaximum                        // As fast as possible
	SpeedCustom                         // Custom interval
)

// NewSimulationEngine creates a new simulation engine
func NewSimulationEngine(
	temporalEngine *engine.TemporalEngine,
	dataset *HistoricalDataSet,
	config *AppConfig,
) *SimulationEngine {
	return &SimulationEngine{
		temporalEngine:  temporalEngine,
		dataset:         dataset,
		config:          config,
		simulationSpeed: 10 * time.Millisecond, // Default: fast simulation
		dailyStats:      make(map[string]*DayStats),
	}
}

// Initialize prepares the simulation engine
func (se *SimulationEngine) Initialize() error {
	fmt.Printf("🚀 Initializing simulation engine...\n")
	
	// Convert dataset to engine format - pass actual timeframes
	tempFetcher := NewDataFetcher(nil, "SOLUSDT", "5m", "15m", true) // TODO: get from config
	initialData, err := tempFetcher.ConvertToEngineFormat(se.dataset)
	if err != nil {
		return fmt.Errorf("failed to convert data: %w", err)
	}
	
	// Initialize temporal engine with historical data
	if err := se.temporalEngine.Initialize(*initialData); err != nil {
		return fmt.Errorf("failed to initialize temporal engine: %w", err)
	}
	
	se.totalTrades = len(se.dataset.Trades)
	se.startTime = time.Now()
	
	// Initialize first day stats
	if len(se.dataset.Trades) > 0 {
		firstTradeTime := time.UnixMilli(se.dataset.Trades[0].Timestamp)
		se.currentDay = time.Date(firstTradeTime.Year(), firstTradeTime.Month(), firstTradeTime.Day(), 0, 0, 0, 0, firstTradeTime.Location())
		se.initializeDayStats(se.currentDay, se.dataset.Trades[0].Price)
	}
	
	fmt.Printf("✅ Simulation engine initialized\n")
	fmt.Printf("📊 Ready to process %d trades over %.1f days\n", 
		se.totalTrades, 
		se.dataset.EndTime.Sub(se.dataset.StartTime).Hours()/24)
	
	return nil
}

// Run executes the main simulation loop
func (se *SimulationEngine) Run(ctx context.Context) error {
	fmt.Printf("🔄 Starting backtest simulation loop...\n")
	fmt.Printf("⚡ Simulation speed: %v between trades\n", se.simulationSpeed)
	fmt.Printf("================================================================================\n\n")
	
	se.isRunning = true
	defer func() {
		se.isRunning = false
	}()
	
	ticker := time.NewTicker(se.simulationSpeed)
	defer ticker.Stop()
	
	progressTicker := time.NewTicker(5 * time.Second) // Progress updates every 5 seconds
	defer progressTicker.Stop()
	
	for se.isRunning && se.currentTradeIndex < se.totalTrades {
		select {
		case <-ctx.Done():
			fmt.Printf("\n🛑 Simulation cancelled by context\n")
			return ctx.Err()
			
		case <-ticker.C:
			if !se.isPaused {
				if err := se.processNextTrade(); err != nil {
					return fmt.Errorf("failed to process trade %d: %w", se.currentTradeIndex, err)
				}
			}
			
		case <-progressTicker.C:
			se.displayProgress()
		}
	}
	
	fmt.Printf("\n✅ Simulation completed successfully!\n")
	se.displayFinalResults()
	
	return nil
}

// processNextTrade processes the next trade in the sequence
func (se *SimulationEngine) processNextTrade() error {
	if se.currentTradeIndex >= len(se.dataset.Trades) {
		se.isRunning = false
		return nil
	}
	
	trade := se.dataset.Trades[se.currentTradeIndex]
	
	// Check if we've moved to a new day
	tradeTime := time.UnixMilli(trade.Timestamp)
	tradeDay := time.Date(tradeTime.Year(), tradeTime.Month(), tradeTime.Day(), 0, 0, 0, 0, tradeTime.Location())
	
	if !tradeDay.Equal(se.currentDay) {
		se.finalizeDayStats(se.currentDay, se.dataset.Trades[se.currentTradeIndex-1].Price)
		se.currentDay = tradeDay
		se.initializeDayStats(tradeDay, trade.Price)
		
		fmt.Printf("\n📅 NEW DAY: %s - Starting price: %.2f\n", 
			tradeDay.Format("2006-01-02"), trade.Price)
	}
	
	// Process trade through temporal engine
	if err := se.temporalEngine.ProcessTrade(trade); err != nil {
		return fmt.Errorf("temporal engine error: %w", err)
	}
	
	// Update statistics
	se.processedTrades++
	se.currentTradeIndex++
	se.updateDailyStats(trade)
	
	// Log trade details if verbose
	if se.config.Verbose && se.processedTrades%100 == 0 {
		position := se.temporalEngine.GetPosition()
		positionStr := "CLOSED"
		if position.IsOpen {
			positionStr = fmt.Sprintf("%s @%.2f", position.Direction, position.EntryPrice)
		}
		
		fmt.Printf("🔄 Trade %d: %s @%.2f | Position: %s\n", 
			se.processedTrades, 
			tradeTime.Format("02/01 15:04:05"), 
			trade.Price,
			positionStr)
	}
	
	return nil
}

// displayProgress shows current simulation progress
func (se *SimulationEngine) displayProgress() {
	if se.totalTrades == 0 {
		return
	}
	
	progress := float64(se.processedTrades) / float64(se.totalTrades) * 100
	elapsed := time.Since(se.startTime)
	
	if se.processedTrades > 0 {
		se.tradesPerSecond = float64(se.processedTrades) / elapsed.Seconds()
		remainingTrades := se.totalTrades - se.processedTrades
		se.estimatedTimeLeft = time.Duration(float64(remainingTrades)/se.tradesPerSecond) * time.Second
	}
	
	// Get current position info
	position := se.temporalEngine.GetPosition()
	metrics := se.temporalEngine.GetMetrics()
	
	fmt.Printf("\r🔄 Progress: %.1f%% (%d/%d trades) | Speed: %.0f trades/s | ETA: %v | Pos: %v | Cycles: %d         ", 
		progress,
		se.processedTrades,
		se.totalTrades,
		se.tradesPerSecond,
		se.estimatedTimeLeft.Truncate(time.Second),
		func() string {
			if position.IsOpen {
				return fmt.Sprintf("%s@%.2f", position.Direction, position.EntryPrice)
			}
			return "CLOSED"
		}(),
		metrics.CyclesExecuted)
}

// initializeDayStats creates stats tracking for a new day
func (se *SimulationEngine) initializeDayStats(day time.Time, startPrice float64) {
	dayKey := day.Format("2006-01-02")
	se.dailyStats[dayKey] = &DayStats{
		Date:            day,
		TradesProcessed: 0,
		PriceStart:      startPrice,
		PriceEnd:        startPrice,
		PriceChange:     0,
		PositionsOpened: 0,
		PositionsClosed: 0,
		PnL:             0,
	}
}

// updateDailyStats updates statistics for the current day
func (se *SimulationEngine) updateDailyStats(trade engine.Trade) {
	dayKey := se.currentDay.Format("2006-01-02")
	if stats, exists := se.dailyStats[dayKey]; exists {
		stats.TradesProcessed++
		stats.PriceEnd = trade.Price
		stats.PriceChange = ((stats.PriceEnd - stats.PriceStart) / stats.PriceStart) * 100
	}
}

// finalizeDayStats completes the stats for a finished day
func (se *SimulationEngine) finalizeDayStats(day time.Time, endPrice float64) {
	dayKey := day.Format("2006-01-02")
	if stats, exists := se.dailyStats[dayKey]; exists {
		stats.PriceEnd = endPrice
		stats.PriceChange = ((stats.PriceEnd - stats.PriceStart) / stats.PriceStart) * 100
		
		fmt.Printf("📊 DAY COMPLETE: %s | Trades: %d | Price: %.2f→%.2f (%.2f%%) | Positions: %d opened, %d closed\n",
			day.Format("2006-01-02"),
			stats.TradesProcessed,
			stats.PriceStart,
			stats.PriceEnd,
			stats.PriceChange,
			stats.PositionsOpened,
			stats.PositionsClosed)
	}
}

// displayFinalResults shows comprehensive simulation results
func (se *SimulationEngine) displayFinalResults() {
	elapsed := time.Since(se.startTime)
	metrics := se.temporalEngine.GetMetrics()
	
	fmt.Printf("\n🎯 SIMULATION RESULTS - %s\n", se.config.Symbol)
	fmt.Printf("================================================================================\n")
	fmt.Printf("⏱️  Total Time:           %v\n", elapsed.Truncate(time.Second))
	fmt.Printf("📊 Trades Processed:      %d/%d (%.1f%%)\n", 
		se.processedTrades, se.totalTrades, 
		float64(se.processedTrades)/float64(se.totalTrades)*100)
	fmt.Printf("⚡ Average Speed:         %.0f trades/second\n", se.tradesPerSecond)
	fmt.Printf("🔄 Engine Cycles:         %d\n", metrics.CyclesExecuted)
	fmt.Printf("📍 Positions Opened:      %d\n", metrics.PositionsOpened)
	fmt.Printf("🚪 Positions Closed:      %d\n", metrics.PositionsClosed)
	fmt.Printf("⚙️  Stop Adjustments:     %d\n", metrics.StopAdjustments)
	fmt.Printf("⚠️  Anti-Lookahead:       %d violations\n", metrics.AntiLookAheadViolations)
	fmt.Printf("🚀 Performance:           Avg %.2fms/cycle, Max %.2fms\n", 
		metrics.AverageLatencyMs, metrics.MaxLatencyMs)
	
	// Display daily summary
	fmt.Printf("\n📅 DAILY SUMMARY:\n")
	fmt.Printf("--------------------------------------------------------------------------------\n")
	fmt.Printf("%-12s %-8s %-12s %-12s %-10s %-8s\n", 
		"Date", "Trades", "Start", "End", "Change%", "Pos")
	fmt.Printf("--------------------------------------------------------------------------------\n")
	
	totalPriceChange := 0.0
	totalDays := 0
	
	for dayKey, stats := range se.dailyStats {
		fmt.Printf("%-12s %-8d %-12.2f %-12.2f %-10.2f %-8d\n",
			dayKey,
			stats.TradesProcessed,
			stats.PriceStart,
			stats.PriceEnd,
			stats.PriceChange,
			stats.PositionsOpened)
		
		totalPriceChange += stats.PriceChange
		totalDays++
	}
	
	if totalDays > 0 {
		fmt.Printf("--------------------------------------------------------------------------------\n")
		fmt.Printf("%-12s %-8s %-12s %-12s %-10.2f %-8s\n",
			"AVERAGE", "-", "-", "-", totalPriceChange/float64(totalDays), "-")
	}
	
	fmt.Printf("================================================================================\n")
}

// Pause pauses the simulation
func (se *SimulationEngine) Pause() {
	se.isPaused = true
	fmt.Printf("\n⏸️  Simulation paused\n")
}

// Resume resumes the simulation
func (se *SimulationEngine) Resume() {
	se.isPaused = false
	fmt.Printf("\n▶️  Simulation resumed\n")
}

// Stop stops the simulation
func (se *SimulationEngine) Stop() {
	se.isRunning = false
	fmt.Printf("\n🛑 Simulation stopped\n")
}

// GetProgress returns current simulation progress
func (se *SimulationEngine) GetProgress() (processed, total int, percentage float64) {
	percentage = 0
	if se.totalTrades > 0 {
		percentage = float64(se.processedTrades) / float64(se.totalTrades) * 100
	}
	return se.processedTrades, se.totalTrades, percentage
}

// IsRunning returns whether simulation is currently running
func (se *SimulationEngine) IsRunning() bool {
	return se.isRunning
}

// IsPaused returns whether simulation is currently paused
func (se *SimulationEngine) IsPaused() bool {
	return se.isPaused
}

// SetSpeed adjusts the simulation speed
func (se *SimulationEngine) SetSpeed(speed time.Duration) {
	se.simulationSpeed = speed
	fmt.Printf("⚡ Simulation speed set to: %v\n", speed)
}
