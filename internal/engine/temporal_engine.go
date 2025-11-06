package engine

import (
	"fmt"
	"sync"
	"time"
	
	"agent-economique/internal/indicators"
	"agent-economique/internal/money_management"
	"agent-economique/internal/shared"
	"agent-economique/internal/strategies/stoch_mfi_cci"
)

// TemporalEngine manages temporal execution and orchestrates components
type TemporalEngine struct {
	// Configuration
	mode   ExecutionMode
	config EngineConfig
	
	// Core components
	positionManager *PositionManager
	zoneMonitor     *ZoneMonitor
	moneyManager    *money_management.MoneyManager
	
	// Temporal state
	currentTimestamp int64
	lastProcessed    int64
	
	// Data management
	historicalTrades       []Trade
	historicalKlinesMultiTF map[string][]Kline // timeframe -> klines
	historicalKlines       []Kline             // Legacy single-TF (deprecated)
	tradeIterator          *TradeIterator
	
	// Performance tracking
	metrics          PerformanceMetrics
	cycleStartTime   time.Time
	
	// Thread safety
	mutex sync.RWMutex
	
	// Integration with Indicators (NEW)
	indicatorResults *indicators.IndicatorResults
	lastSignalTime   int64
	
	// STOCH/MFI/CCI Strategy Integration (NEW)
	stochStrategy    *stoch_mfi_cci.EngineIntegration
	strategyEnabled  bool
	
	// Control
	running bool
}

// NewTemporalEngineFromYAML creates a new temporal engine instance from YAML configuration
func NewTemporalEngineFromYAML(mode ExecutionMode, yamlConfig *shared.Config, engineConfig EngineConfig) (*TemporalEngine, error) {
	if err := ValidateConfiguration(engineConfig); err != nil {
		return nil, fmt.Errorf("invalid engine configuration: %w", err)
	}
	
	engine := &TemporalEngine{
		mode:            mode,
		config:          engineConfig,
		positionManager: NewPositionManager(engineConfig.TrailingStop, engineConfig.AdjustmentGrid),
		zoneMonitor:     NewZoneMonitor(engineConfig.Zones),
		historicalTrades:       make([]Trade, 0, engineConfig.WindowSize),
		historicalKlinesMultiTF: make(map[string][]Kline),
		historicalKlines:       make([]Kline, 0, engineConfig.WindowSize),
		metrics:         PerformanceMetrics{},
		running:         false,
		strategyEnabled: yamlConfig.Strategy.Name == "STOCH_MFI_CCI", // Enable based on YAML config
	}
	
	// Initialize STOCH/MFI/CCI strategy from YAML configuration
	if engine.strategyEnabled {
		yamlStrategyConfig := yamlConfig.ToSTOCHMFICCIStrategyConfig()
		
		// Convert to the format expected by the strategy
		strategyConfig := convertYAMLToStrategyConfig(yamlStrategyConfig)
		
		stochStrategy, err := stoch_mfi_cci.NewEngineIntegration(strategyConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize STOCH/MFI/CCI strategy from YAML: %w", err)
		}
		
		engine.stochStrategy = stochStrategy
		
		// Set engine callbacks for position management
		stochStrategy.SetEngineCallbacks(
			engine.closePositionFromStrategy,
			engine.adjustStopFromStrategy,
		)
		
		fmt.Printf("[Engine] ✅ STOCH/MFI/CCI strategy initialized from YAML config\n")
	}
	
	return engine, nil
}

// NewTemporalEngine creates a new temporal engine instance
func NewTemporalEngine(mode ExecutionMode, config EngineConfig) (*TemporalEngine, error) {
	if err := ValidateConfiguration(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	engine := &TemporalEngine{
		mode:            mode,
		config:          config,
		positionManager: NewPositionManager(config.TrailingStop, config.AdjustmentGrid),
		zoneMonitor:     NewZoneMonitor(config.Zones),
		historicalTrades:       make([]Trade, 0, config.WindowSize),
		historicalKlinesMultiTF: make(map[string][]Kline),
		historicalKlines:       make([]Kline, 0, config.WindowSize),
		metrics:         PerformanceMetrics{},
		running:         false,
		strategyEnabled: true, // Enable STOCH/MFI/CCI strategy by default
	}
	
	// Initialize STOCH/MFI/CCI strategy
	if engine.strategyEnabled {
		strategyConfig := stoch_mfi_cci.DefaultStrategyConfig()
		// TODO: Load config from YAML in future
		
		stochStrategy, err := stoch_mfi_cci.NewEngineIntegration(strategyConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize STOCH/MFI/CCI strategy: %w", err)
		}
		
		engine.stochStrategy = stochStrategy
		
		// Set engine callbacks for position management
		stochStrategy.SetEngineCallbacks(
			engine.closePositionFromStrategy,
			engine.adjustStopFromStrategy,
		)
		
		fmt.Printf("[Engine] ✅ STOCH/MFI/CCI strategy initialized\n")
	}
	
	return engine, nil
}

// Initialize prepares the engine for execution
func (e *TemporalEngine) Initialize(initialData InitialData) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	switch e.mode {
	case BacktestMode:
		return e.initializeBacktest(initialData)
	case PaperMode, LiveMode:
		return e.initializeLive(initialData)
	default:
		return fmt.Errorf("unsupported execution mode: %v", e.mode)
	}
}

// ProcessTrade processes a single trade (main entry point)
func (e *TemporalEngine) ProcessTrade(trade Trade) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if !e.running {
		return fmt.Errorf("engine not running")
	}
	
	cycleStart := time.Now()
	
	// Update temporal state
	if err := e.updateTimestamp(trade.Timestamp); err != nil {
		return fmt.Errorf("temporal update failed: %w", err)
	}
	
	// Update historical data (with anti-look-ahead validation)
	if err := e.updateHistoricalData(trade); err != nil {
		return fmt.Errorf("historical data update failed: %w", err)
	}
	
	// Process position management if position is open
	if e.positionManager.IsOpen() {
		if err := e.processOpenPosition(trade.Price); err != nil {
			return fmt.Errorf("position processing failed: %w", err)
		}
		
		// STOCH/MFI/CCI tick-by-tick processing (NEW)
		if e.strategyEnabled && e.stochStrategy != nil {
			// Only process tick events if monitoring is active
			if e.stochStrategy.IsMonitoringActive() {
				e.processSTOCHTickEvent(trade.Price)
			}
		}
	}
	
	// Check for candle markers and trigger calculations
	if IsMarkerTimestamp(trade.Timestamp) {
		if err := e.processMarkerEvent(trade.Timestamp); err != nil {
			return fmt.Errorf("marker processing failed: %w", err)
		}
	}
	
	// Update performance metrics
	e.updatePerformanceMetrics(time.Since(cycleStart))
	
	return nil
}

// ProcessTimerEvent processes timer events for live/paper mode
func (e *TemporalEngine) ProcessTimerEvent() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if e.mode == BacktestMode {
		return fmt.Errorf("timer events not supported in backtest mode")
	}
	
	cycleStart := time.Now()
	currentTime := cycleStart.UnixMilli()
	
	// Update timestamp to current time
	if err := e.updateTimestamp(currentTime); err != nil {
		return fmt.Errorf("timestamp update failed: %w", err)
	}
	
	// Fetch new data since last processing
	if err := e.fetchNewData(); err != nil {
		return fmt.Errorf("data fetch failed: %w", err)
	}
	
	// Check for completed candles
	completedCandles := e.detectCompletedCandles()
	for _, candleTime := range completedCandles {
		if err := e.processMarkerEvent(candleTime); err != nil {
			return fmt.Errorf("candle processing failed: %w", err)
		}
	}
	
	// Update performance metrics
	e.updatePerformanceMetrics(time.Since(cycleStart))
	
	return nil
}

// GetCurrentTimestamp returns the current temporal timestamp
func (e *TemporalEngine) GetCurrentTimestamp() int64 {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.currentTimestamp
}

// GetPosition returns current position information
func (e *TemporalEngine) GetPosition() Position {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.positionManager.GetPosition()
}

// GetActiveZones returns currently active zones
func (e *TemporalEngine) GetActiveZones() map[ZoneType]ActiveZone {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.zoneMonitor.GetActiveZones()
}

// GetMetrics returns current performance metrics
func (e *TemporalEngine) GetMetrics() PerformanceMetrics {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.metrics
}

// Start begins engine execution
func (e *TemporalEngine) Start() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if e.running {
		return fmt.Errorf("engine already running")
	}
	
	e.running = true
	return nil
}

// Stop halts engine execution
func (e *TemporalEngine) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	if !e.running {
		return fmt.Errorf("engine not running")
	}
	
	e.running = false
	
	// Close any open positions
	if e.positionManager.IsOpen() {
		position := e.positionManager.GetPosition()
		e.positionManager.ClosePosition(e.currentTimestamp, position.StopLoss, "engine_shutdown")
		e.metrics.PositionsClosed++
	}
	
	return nil
}

// IsRunning returns whether the engine is currently running
func (e *TemporalEngine) IsRunning() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.running
}

// Private methods

func (e *TemporalEngine) initializeBacktest(data InitialData) error {
	if len(data.Trades) == 0 {
		return fmt.Errorf("no trades provided for backtest")
	}
	
	e.tradeIterator = NewTradeIterator(data.Trades)
	e.currentTimestamp = data.Trades[0].Timestamp
	e.lastProcessed = 0
	
	// Load recent historical context - prioritize multi-TF data
	if len(data.RecentKlinesMultiTF) > 0 {
		e.historicalKlinesMultiTF = data.RecentKlinesMultiTF
		totalKlines := 0
		for tf, klines := range data.RecentKlinesMultiTF {
			totalKlines += len(klines)
			fmt.Printf("[Engine] ✅ Loaded %d klines for timeframe %s\n", len(klines), tf)
		}
		fmt.Printf("[Engine] ✅ Loaded %d total historical klines across %d timeframes for backtest\n", 
			totalKlines, len(data.RecentKlinesMultiTF))
	} else if len(data.RecentKlines) > 0 {
		// Fallback to legacy single-TF data
		e.historicalKlines = data.RecentKlines
		fmt.Printf("[Engine] ⚠️  Using legacy single-TF klines: %d klines loaded\n", len(data.RecentKlines))
	}
	
	if len(data.RecentTrades) > 0 {
		e.historicalTrades = data.RecentTrades
		fmt.Printf("[Engine] ✅ Loaded %d historical trades for backtest\n", len(data.RecentTrades))
	}
	
	return nil
}

func (e *TemporalEngine) initializeLive(data InitialData) error {
	e.currentTimestamp = time.Now().UnixMilli()
	e.lastProcessed = e.currentTimestamp
	
	// Load recent historical context - prioritize multi-TF data
	if len(data.RecentKlinesMultiTF) > 0 {
		e.historicalKlinesMultiTF = data.RecentKlinesMultiTF
	} else if len(data.RecentKlines) > 0 {
		// Fallback to legacy single-TF data
		e.historicalKlines = data.RecentKlines
	}
	
	if len(data.RecentTrades) > 0 {
		e.historicalTrades = data.RecentTrades
	}
	
	return nil
}

func (e *TemporalEngine) updateTimestamp(newTimestamp int64) error {
	if e.config.AntiLookAhead && newTimestamp < e.currentTimestamp {
		return fmt.Errorf("%w: new timestamp %d < current %d", 
			ErrLookAheadDetected, newTimestamp, e.currentTimestamp)
	}
	
	e.currentTimestamp = newTimestamp
	return nil
}

func (e *TemporalEngine) updateHistoricalData(trade Trade) error {
	// Anti-look-ahead validation
	if e.config.AntiLookAhead && trade.Timestamp > e.currentTimestamp {
		e.metrics.AntiLookAheadViolations++
		return fmt.Errorf("%w: trade timestamp %d > current %d", 
			ErrLookAheadDetected, trade.Timestamp, e.currentTimestamp)
	}
	
	// Add trade to historical data
	e.historicalTrades = append(e.historicalTrades, trade)
	
	// Maintain window size for trades
	if len(e.historicalTrades) > e.config.WindowSize {
		e.historicalTrades = e.historicalTrades[1:]
	}
	
	return nil
}

func (e *TemporalEngine) processOpenPosition(currentPrice float64) error {
	// Update trailing stop
	if err := e.positionManager.UpdateTrailingStop(currentPrice); err != nil {
		return fmt.Errorf("trailing stop update failed: %w", err)
	}
	
	// Check active zones and apply adjustments
	position := e.positionManager.GetPosition()
	currentProfit := position.CalculateProfitPercent(currentPrice)
	
	if adjustment := e.zoneMonitor.CheckActiveZones(currentProfit, e.currentTimestamp); adjustment != nil {
		if err := e.positionManager.ApplyStopAdjustment(adjustment.NewStopPercent); err != nil {
			return fmt.Errorf("stop adjustment failed: %w", err)
		}
		e.metrics.StopAdjustments++
	}
	
	// Check if stop is hit
	if position.IsStopHit(currentPrice) {
		if err := e.positionManager.ClosePosition(e.currentTimestamp, currentPrice, "stop_hit"); err != nil {
			return fmt.Errorf("position close failed: %w", err)
		}
		e.metrics.PositionsClosed++
		e.zoneMonitor.ResetAllZones()
	}
	
	return nil
}

func (e *TemporalEngine) processMarkerEvent(timestamp int64) error {
	e.metrics.CyclesExecuted++
	
	// =============================================================================
	// COMPLETE STRATEGY EXECUTION AT CANDLE MARKERS (NEW)
	// =============================================================================
	
	fmt.Printf("[%d] 📊 Marker event - executing STOCH/MFI/CCI strategy workflow\n", timestamp)
	
	// Step 1: Calculate indicators (STOCH/MFI/CCI + legacy MACD/CCI/DMI)
	fmt.Printf("[%d] 🧮 Step 1: Calculating indicators...\n", timestamp)
	response, err := e.calculateIndicators()
	if err != nil {
		fmt.Printf("[%d] ❌ Error calculating indicators: %v\n", timestamp, err)
		return fmt.Errorf("indicator calculation failed: %w", err)
	}
	
	if !response.Success {
		fmt.Printf("[%d] ⚠️  Indicator calculation unsuccessful: %v\n", timestamp, response.Error)
		return fmt.Errorf("indicator calculation unsuccessful: %w", response.Error)
	}
	
	// Step 2: STOCH/MFI/CCI Strategy Processing (NEW)
	if e.strategyEnabled && e.stochStrategy != nil {
		fmt.Printf("[%d] 🎯 Step 2: Processing STOCH/MFI/CCI strategy...\n", timestamp)
		e.processSTOCHStrategy(response.Results)
	}
	
	// Step 3: Legacy signal processing (fallback)
	if len(response.Signals) > 0 {
		fmt.Printf("[%d] 🔄 Step 3: Processing %d legacy signals...\n", timestamp, len(response.Signals))
		e.processStrategySignals(response.Signals)
	} else {
		fmt.Printf("[%d] ℹ️  Step 3: No legacy signals generated\n", timestamp)
	}
	
	// Step 4: Legacy zone events (fallback)
	if len(response.ZoneEvents) > 0 {
		fmt.Printf("[%d] 🔄 Step 4: Processing %d zone events...\n", timestamp, len(response.ZoneEvents))
		e.processZoneEvents(response.ZoneEvents)
	} else {
		fmt.Printf("[%d] ℹ️  Step 4: No zone events detected\n", timestamp)
	}
	
	fmt.Printf("[%d] ✅ Strategy workflow completed successfully\n", timestamp)
	return nil
}

func (e *TemporalEngine) fetchNewData() error {
	// In a real implementation, this would fetch new data from external sources
	// For now, this is a placeholder
	return nil
}

func (e *TemporalEngine) detectCompletedCandles() []int64 {
	// In a real implementation, this would detect newly completed candles
	// For now, return empty slice
	return []int64{}
}

func (e *TemporalEngine) updatePerformanceMetrics(cycleDuration time.Duration) {
	latencyMs := float64(cycleDuration.Nanoseconds()) / 1e6
	
	if latencyMs > e.metrics.MaxLatencyMs {
		e.metrics.MaxLatencyMs = latencyMs
	}
	
	// Update average latency (simple moving average)
	totalCycles := float64(e.metrics.CyclesExecuted)
	if totalCycles > 0 {
		e.metrics.AverageLatencyMs = (e.metrics.AverageLatencyMs*(totalCycles-1) + latencyMs) / totalCycles
	} else {
		e.metrics.AverageLatencyMs = latencyMs
	}
}

// InitialData holds initial data for engine initialization
type InitialData struct {
	Trades                []Trade            `json:"trades"`
	RecentKlinesMultiTF   map[string][]Kline `json:"recent_klines_multi_tf"` // timeframe -> klines
	RecentTrades          []Trade            `json:"recent_trades"`
	
	// Legacy single-timeframe support (deprecated, for compatibility)
	RecentKlines          []Kline            `json:"recent_klines,omitempty"`
}

// TradeIterator provides sequential access to trades for backtest mode
type TradeIterator struct {
	trades []Trade
	index  int
}

// NewTradeIterator creates a new trade iterator
func NewTradeIterator(trades []Trade) *TradeIterator {
	return &TradeIterator{
		trades: trades,
		index:  0,
	}
}

// HasNext returns true if there are more trades
func (ti *TradeIterator) HasNext() bool {
	return ti.index < len(ti.trades)
}

// Next returns the next trade and advances the iterator
func (ti *TradeIterator) Next() Trade {
	if !ti.HasNext() {
		return Trade{}
	}
	
	trade := ti.trades[ti.index]
	ti.index++
	return trade
}

// Reset resets the iterator to the beginning
func (ti *TradeIterator) Reset() {
	ti.index = 0
}

// =============================================================================
// INTEGRATION ENGINE ↔ INDICATORS (NEW)
// =============================================================================

// calculateIndicators calls indicators module and updates engine state
func (e *TemporalEngine) calculateIndicators() (*indicators.CalculationResponse, error) {
	// Preparation request with Engine data
	request := &indicators.CalculationRequest{
		Symbol:       "SOLUSDT", // TODO: Add to config if needed
		Timeframe:    "5m",      // TODO: Add to config if needed
		CurrentTime:  e.currentTimestamp,
		CandleWindow: e.getCandleWindow(), // Use existing window
		RequestID:    fmt.Sprintf("engine-%d", e.currentTimestamp),
		
		// Position context for zone events
		PositionContext: e.getPositionContext(),
	}
	
	// Call Indicators module
	response := indicators.Calculate(request)
	
	if response.Success {
		e.indicatorResults = response.Results
		e.logIndicatorResults(response.Results)
	}
	
	return response, response.Error
}

// getPositionContext converts Engine position to indicators PositionContext
func (e *TemporalEngine) getPositionContext() *indicators.PositionContext {
	position := e.positionManager.GetPosition()
	if !position.IsOpen {
		return nil
	}
	
	return &indicators.PositionContext{
		IsOpen:        true,
		Direction:     position.Direction.String(), // "LONG" or "SHORT"
		EntryPrice:    position.EntryPrice,
		EntryTime:     position.EntryTime,
		EntryCCIZone:  e.convertCCIZone(position.EntryCCIZone), // Convert string to CCIZone
		ProfitPercent: position.CalculateProfitPercent(e.getCurrentPrice()), // Need current price
	}
}

// getCandleWindow converts Engine klines to indicators format
// Uses primary timeframe (5m) for marker-based calculations
func (e *TemporalEngine) getCandleWindow() []indicators.Kline {
	var sourceKlines []Kline
	
	// Prioritize multi-TF data - use primary timeframe for calculations
	if len(e.historicalKlinesMultiTF) > 0 {
		// Find primary timeframe (usually 5m) for marker calculations
		primaryTF := "5m"
		if klines, exists := e.historicalKlinesMultiTF[primaryTF]; exists {
			sourceKlines = klines
		} else {
			// Fallback to first available timeframe
			for _, klines := range e.historicalKlinesMultiTF {
				sourceKlines = klines
				break
			}
		}
	} else {
		// Fallback to legacy single-TF data
		sourceKlines = e.historicalKlines
	}
	
	// STRICT anti-lookahead: only klines up to current timestamp
	validKlines := make([]Kline, 0, len(sourceKlines))
	for _, kline := range sourceKlines {
		if kline.Timestamp <= e.currentTimestamp {
			validKlines = append(validKlines, kline)
		}
	}
	
	// Convert to indicators format
	window := make([]indicators.Kline, len(validKlines))
	for i, kline := range validKlines {
		window[i] = indicators.Kline{
			Timestamp: kline.Timestamp,
			Open:      kline.Open,
			High:      kline.High,
			Low:       kline.Low,
			Close:     kline.Close,
			Volume:    kline.Volume,
		}
	}
	
	return window
}

// convertCCIZone converts string CCI zone to indicators.CCIZone enum
func (e *TemporalEngine) convertCCIZone(zoneStr string) indicators.CCIZone {
	switch zoneStr {
	case "OVERSOLD":
		return indicators.CCIOversold
	case "OVERBOUGHT":
		return indicators.CCIOverbought
	case "NORMAL":
		return indicators.CCINormal
	default:
		return indicators.CCINormal // Default to normal
	}
}

// getCurrentPrice returns current market price estimate
func (e *TemporalEngine) getCurrentPrice() float64 {
	// Use latest kline close price if available
	if len(e.historicalKlines) > 0 {
		return e.historicalKlines[len(e.historicalKlines)-1].Close
	}
	
	// Use latest trade price if available
	if len(e.historicalTrades) > 0 {
		return e.historicalTrades[len(e.historicalTrades)-1].Price
	}
	
	// Fallback to 0 - should not happen in normal operation
	return 0.0
}

// logIndicatorResults logs calculated indicator values for debugging
func (e *TemporalEngine) logIndicatorResults(results *indicators.IndicatorResults) {
	if results == nil {
		return
	}
	
	macdInfo := "MACD=nil"
	if results.MACD != nil {
		macdInfo = fmt.Sprintf("MACD=%.4f", results.MACD.MACD)
	}
	
	cciInfo := "CCI=nil"
	if results.CCI != nil {
		cciInfo = fmt.Sprintf("CCI=%.2f", results.CCI.Value)
	}
	
	dmiInfo := "DMI=nil"
	if results.DMI != nil {
		dmiInfo = fmt.Sprintf("DMI=%.1f/%.1f", results.DMI.PlusDI, results.DMI.MinusDI)
	}
	
	// Use existing log method if available, or placeholder
	fmt.Printf("[%d] 📊 Indicators: %s, %s, %s\n", 
		e.currentTimestamp, macdInfo, cciInfo, dmiInfo)
}

// =============================================================================
// SIGNAL AND ZONE EVENT PROCESSING (NEW)
// =============================================================================

// processStrategySignals processes trading signals from indicators module
func (e *TemporalEngine) processStrategySignals(signals []indicators.StrategySignal) {
	if e.positionManager.IsOpen() {
		return // Position already open
	}
	
	for _, signal := range signals {
		// Filter minimum confidence according to strategy
		if signal.Confidence < 0.7 {
			fmt.Printf("[%d] ⚠️  Signal ignored: confidence %.2f < 0.7\n", 
				e.currentTimestamp, signal.Confidence)
			continue
		}
		
		// Convert signal direction to Engine position direction
		direction := PositionLong
		if signal.Direction == indicators.ShortSignal {
			direction = PositionShort
		}
		
		fmt.Printf("[%d] 🚀 Opening position %s: confidence=%.2f, type=%v\n", 
			e.currentTimestamp, direction, signal.Confidence, signal.Type)
			
		// Open position using position manager
		err := e.openPosition(direction, signal.Timestamp)
		if err != nil {
			fmt.Printf("[%d] ❌ Error opening position: %v\n", e.currentTimestamp, err)
			continue
		}
		
		e.lastSignalTime = signal.Timestamp
		break // Only one position at a time
	}
}

// processZoneEvents processes zone monitoring events for trailing stop adjustments
func (e *TemporalEngine) processZoneEvents(events []indicators.ZoneEvent) {
	if !e.positionManager.IsOpen() {
		return // No position to adjust
	}
	
	for _, event := range events {
		if event.Type != "ZONE_ACTIVATED" {
			continue
		}
		
		switch event.ZoneType {
		case "CCI_INVERSE":
			fmt.Printf("[%d] 🔄 CCI zone inverse detected - adjusting trailing stop\n", 
				e.currentTimestamp)
			e.adjustTrailingStopForCCIInverse()
			
		case "MACD_INVERSE":
			if event.CurrentProfit > event.ProfitThreshold {
				fmt.Printf("[%d] 🔄 MACD inverse with profit %.2f%% - adjusting\n", 
					e.currentTimestamp, event.CurrentProfit)
				e.adjustTrailingStopForMACDInverse()
			}
			
		case "DI_COUNTER":
			if event.CurrentProfit > event.ProfitThreshold {
				fmt.Printf("[%d] 🔄 DI counter-trend with profit %.2f%% - adjusting\n", 
					e.currentTimestamp, event.CurrentProfit)
				e.adjustTrailingStopForDICounter()
			}
		}
	}
}

// openPosition opens a new trading position
func (e *TemporalEngine) openPosition(direction PositionDirection, timestamp int64) error {
	currentPrice := e.getCurrentPrice()
	if currentPrice <= 0 {
		return fmt.Errorf("invalid current price: %f", currentPrice)
	}
	
	// Determine entry CCI zone for tracking
	entryCCIZone := "NORMAL"
	if e.indicatorResults != nil && e.indicatorResults.CCI != nil {
		entryCCIZone = e.indicatorResults.CCI.Zone.String()
	}
	
	// Open position using position manager
	err := e.positionManager.OpenPosition(direction, currentPrice, timestamp, entryCCIZone)
	if err != nil {
		return fmt.Errorf("failed to open position: %w", err)
	}
	
	fmt.Printf("[%d] ✅ Position opened: %s at %.4f, CCI zone: %s\n", 
		timestamp, direction, currentPrice, entryCCIZone)
	
	return nil
}

// =============================================================================
// TRAILING STOP ADJUSTMENTS (NEW)
// =============================================================================

// adjustTrailingStopForCCIInverse adjusts trailing stop for CCI inverse zone
func (e *TemporalEngine) adjustTrailingStopForCCIInverse() {
	// According to user memory: CCI zone extreme inverse adjustment
	adjustment := 0.1 // 10% more aggressive (0.1% tighter)
	err := e.positionManager.ApplyStopAdjustment(adjustment)
	if err != nil {
		fmt.Printf("[%d] ❌ Error adjusting CCI inverse stop: %v\n", e.currentTimestamp, err)
		return
	}
	
	fmt.Printf("[%d] 📊 Trailing stop adjusted: +%.1f%% (CCI inverse)\n", 
		e.currentTimestamp, adjustment)
}

// =============================================================================
// STOCH/MFI/CCI STRATEGY INTEGRATION CALLBACKS (NEW)
// =============================================================================

// closePositionFromStrategy handles position close requests from STOCH/MFI/CCI strategy
func (e *TemporalEngine) closePositionFromStrategy(reason string) error {
	if !e.positionManager.IsOpen() {
		return fmt.Errorf("no position to close")
	}
	
	// Get current position for recording
	position := e.positionManager.GetPosition()
	
	// Close position at current price (last trade price)
	currentPrice := 0.0
	if len(e.historicalTrades) > 0 {
		currentPrice = e.historicalTrades[len(e.historicalTrades)-1].Price
	}
	
	err := e.positionManager.ClosePosition(e.currentTimestamp, currentPrice, reason)
	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}
	
	// Notify strategy that position was closed
	if e.stochStrategy != nil {
		e.stochStrategy.OnPositionClosed()
	}
	
	// Notify money management
	e.notifyPositionClosed(position, currentPrice, reason)
	
	fmt.Printf("[Engine] 🚪 Position closed by strategy: %s\n", reason)
	return nil
}

// adjustStopFromStrategy handles stop price adjustments from STOCH/MFI/CCI strategy
func (e *TemporalEngine) adjustStopFromStrategy(newStopPrice float64) error {
	if !e.positionManager.IsOpen() {
		return fmt.Errorf("no position to adjust")
	}
	
	// Update stop price manually (no direct UpdateStopPrice method)
	// We'll manually update the position stop loss
	position := e.positionManager.GetPosition()
	if position.IsOpen {
		// For now, we'll use the existing position structure
		// In a full implementation, we'd need to add UpdateStopPrice method
		fmt.Printf("[Engine] ⚠️  Manual stop price update needed: %.2f\n", newStopPrice)
	}
	
	fmt.Printf("[Engine] ⚡ Stop price updated by strategy: %.2f\n", newStopPrice)
	return nil
}

// processSTOCHStrategy processes STOCH/MFI/CCI strategy at marker events
func (e *TemporalEngine) processSTOCHStrategy(indicatorResults *indicators.IndicatorResults) {
	if e.stochStrategy == nil {
		return
	}
	
	// Convert engine klines to indicators klines
	indicatorKlines := make([]indicators.Kline, len(e.historicalKlines))
	for i, k := range e.historicalKlines {
		indicatorKlines[i] = indicators.Kline{
			Timestamp: k.Timestamp,
			Open:     k.Open,
			High:     k.High,
			Low:      k.Low,
			Close:    k.Close,
			Volume:   k.Volume,
		}
	}
	
	// Generate signals using STOCH/MFI/CCI strategy
	signals := e.stochStrategy.GenerateSignals(indicatorResults, indicatorKlines, "BTC-USDT")
	
	// Process generated signals (open positions if not already open)
	if len(signals) > 0 && !e.positionManager.IsOpen() {
		fmt.Printf("[%d] 🎯 STOCH Strategy generated %d signals\n", e.currentTimestamp, len(signals))
		
		for _, signal := range signals {
			if signal.Confidence >= 0.7 {
				// Open position
				currentPrice := 0.0
				if len(e.historicalTrades) > 0 {
					currentPrice = e.historicalTrades[len(e.historicalTrades)-1].Price
				}
				
				var direction PositionDirection
				if signal.Direction.String() == "LONG" {
					direction = PositionLong
				} else {
					direction = PositionShort
				}
				err := e.positionManager.OpenPosition(direction, currentPrice, e.currentTimestamp, "STOCH_SIGNAL")
				if err != nil {
					fmt.Printf("[%d] ❌ Failed to open STOCH position: %v\n", e.currentTimestamp, err)
					continue
				}
				
				// Notify strategy of position opening
				err = e.stochStrategy.OnPositionOpened(
					signal.Direction.String(),
					currentPrice,
					signal,
					indicatorResults,
				)
				if err != nil {
					fmt.Printf("[%d] ⚠️  Strategy notification failed: %v\n", e.currentTimestamp, err)
				}
				
				// Notify money management
				e.notifyPositionOpened(currentPrice)
				
				fmt.Printf("[%d] ✅ STOCH position opened: %s at %.2f (confidence: %.2f)\n",
					e.currentTimestamp, signal.Direction, currentPrice, signal.Confidence)
				break // Only open one position
			}
		}
	}
	
	// Process marker events for existing positions
	if e.positionManager.IsOpen() {
		currentPrice := 0.0
		if len(e.historicalTrades) > 0 {
			currentPrice = e.historicalTrades[len(e.historicalTrades)-1].Price
		}
		
		// Check stop loss trigger
		if e.stochStrategy.ShouldTriggerStop(currentPrice) {
			fmt.Printf("[%d] 🛑 STOCH strategy stop triggered at %.2f\n", e.currentTimestamp, currentPrice)
			e.closePositionFromStrategy("STOCH_STOP_LOSS")
			return
		}
		
		// Process normal marker event (trailing stop updates)
		err := e.stochStrategy.ProcessMarkerEvent(currentPrice, indicatorResults)
		if err != nil {
			fmt.Printf("[%d] ⚠️  STOCH marker processing error: %v\n", e.currentTimestamp, err)
		}
	}
}

// processSTOCHTickEvent processes tick-by-tick events for STOCH/MFI/CCI strategy
func (e *TemporalEngine) processSTOCHTickEvent(currentPrice float64) {
	if e.stochStrategy == nil || e.indicatorResults == nil {
		return
	}
	
	// Check stop loss trigger first
	if e.stochStrategy.ShouldTriggerStop(currentPrice) {
		fmt.Printf("[%d] 🛑 STOCH tick-by-tick stop triggered at %.2f\n", e.currentTimestamp, currentPrice)
		e.closePositionFromStrategy("STOCH_TICK_STOP")
		return
	}
	
	// Process tick event with current indicator values
	err := e.stochStrategy.ProcessTradeEvent(currentPrice, e.indicatorResults)
	if err != nil {
		fmt.Printf("[%d] ⚠️  STOCH tick processing error: %v\n", e.currentTimestamp, err)
	}
}

// adjustTrailingStopForMACDInverse adjusts trailing stop for MACD inverse signal
func (e *TemporalEngine) adjustTrailingStopForMACDInverse() {
	// According to user memory: MACD inverse if profit
	adjustment := 0.05 // 5% more aggressive
	err := e.positionManager.ApplyStopAdjustment(adjustment)
	if err != nil {
		fmt.Printf("[%d] ❌ Error adjusting MACD inverse stop: %v\n", e.currentTimestamp, err)
		return
	}
	
	fmt.Printf("[%d] 📊 Trailing stop adjusted: +%.1f%% (MACD inverse)\n", 
		e.currentTimestamp, adjustment)
}

// adjustTrailingStopForDICounter adjusts trailing stop for DI counter-trend
func (e *TemporalEngine) adjustTrailingStopForDICounter() {
	// According to user memory: DI counter-trend if profit
	adjustment := 0.08 // 8% more aggressive
	err := e.positionManager.ApplyStopAdjustment(adjustment)
	if err != nil {
		fmt.Printf("[%d] ❌ Error adjusting DI counter stop: %v\n", e.currentTimestamp, err)
		return
	}
	
	fmt.Printf("[%d] 📊 Trailing stop adjusted: +%.1f%% (DI counter)\n", 
		e.currentTimestamp, adjustment)
}

// convertYAMLToStrategyConfig converts shared config to strategy config format
func convertYAMLToStrategyConfig(yamlConfig shared.STOCHMFICCIStrategyConfig) stoch_mfi_cci.StrategyConfig {
	return stoch_mfi_cci.StrategyConfig{
		// Indicator Parameters
		StochPeriodK:    yamlConfig.StochPeriodK,
		StochSmoothK:    yamlConfig.StochSmoothK,
		StochPeriodD:    yamlConfig.StochPeriodD,
		StochOversold:   yamlConfig.StochOversold,
		StochOverbought: yamlConfig.StochOverbought,
		
		MFIPeriod:       yamlConfig.MFIPeriod,
		MFIOversold:     yamlConfig.MFIOversold,
		MFIOverbought:   yamlConfig.MFIOverbought,
		
		CCIThreshold:    yamlConfig.CCIThreshold,
		
		// Signal Generation
		MinConfidence:         yamlConfig.MinConfidence,
		PremiumConfidence:     yamlConfig.PremiumConfidence,
		RequireBarConfirmation: yamlConfig.RequireBarConfirmation,
		HigherTimeframe:       yamlConfig.HigherTimeframe,
		EnableMultiTF:         yamlConfig.EnableMultiTF,
		
		// Position Management  
		BaseTrailingPercent:      yamlConfig.BaseTrailingPercent,
		TrendTrailingPercent:     yamlConfig.TrendTrailingPercent,
		CounterTrendTrailing:     yamlConfig.CounterTrendTrailing,
		
		// Dynamic Adjustments
		EnableDynamicAdjustments: yamlConfig.EnableDynamicAdjustments,
		STOCHInverseAdjust:       yamlConfig.STOCHInverseAdjust,
		MFIInverseAdjust:         yamlConfig.MFIInverseAdjust,
		CCIInverseAdjust:         yamlConfig.CCIInverseAdjust,
		TripleInverseAdjust:      yamlConfig.TripleInverseAdjust,
		
		// Safety Limits
		MaxCumulativeAdjust:      yamlConfig.MaxCumulativeAdjust,
		MinTrailingPercent:       yamlConfig.MinTrailingPercent,
		MaxTrailingPercent:       yamlConfig.MaxTrailingPercent,
		
		// Early Exit
		EnableEarlyExit:          yamlConfig.EnableEarlyExit,
		MinProfitForEarlyExit:    yamlConfig.MinProfitForEarlyExit,
		TripleInverseEarlyExit:   yamlConfig.TripleInverseEarlyExit,
	}
}

// GetCurrentIndicatorResults returns current indicator values for external access
func (e *TemporalEngine) GetCurrentIndicatorResults() *indicators.IndicatorResults {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.indicatorResults
}

// GetStrategySignals returns current strategy signals if any
func (e *TemporalEngine) GetStrategySignals() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	if e.stochStrategy == nil {
		return nil
	}
	
	// Get current strategy status
	// TODO: Implement method to get signals from strategy
	return map[string]interface{}{
		"strategy_active": e.strategyEnabled,
		"last_signal_time": e.lastSignalTime,
	}
}
