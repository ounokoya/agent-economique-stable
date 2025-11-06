package stoch_mfi_cci

import (
	"fmt"
	"time"
	
	"agent-economique/internal/indicators"
)

// EngineIntegration provides integration points between STOCH/MFI/CCI strategy and Engine
type EngineIntegration struct {
	config       StrategyConfig
	behavioralMM *BehavioralMoneyManager
	
	// Engine callbacks
	onPositionClose  func(reason string) error
	onStopAdjustment func(newStopPrice float64) error
	
	// Strategy state
	totalSignalsGenerated int
	premiumSignalsCount   int
	totalConfidence       float64
	tickProcessingCount   int64
	
	// Multi-timeframe cache
	higherTFCache map[string]*indicators.IndicatorResults
	lastCacheTime time.Time
}

// NewEngineIntegration creates integration between STOCH/MFI/CCI strategy and Engine
func NewEngineIntegration(config StrategyConfig) (*EngineIntegration, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	bmm, err := NewBehavioralMoneyManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create behavioral MM: %w", err)
	}
	
	integration := &EngineIntegration{
		config:            config,
		behavioralMM:      bmm,
		higherTFCache:     make(map[string]*indicators.IndicatorResults),
		totalSignalsGenerated: 0,
		premiumSignalsCount:   0,
		totalConfidence:       0.0,
		tickProcessingCount:   0,
	}
	
	// Set up behavioral MM callbacks
	bmm.SetCallbacks(
		integration.onTrailingAdjustment,
		integration.onEarlyExit,
		integration.onStateChange,
	)
	
	return integration, nil
}

// SetEngineCallbacks sets callbacks to communicate with engine
func (ei *EngineIntegration) SetEngineCallbacks(
	onPositionClose func(string) error,
	onStopAdjustment func(float64) error,
) {
	ei.onPositionClose = onPositionClose
	ei.onStopAdjustment = onStopAdjustment
}

// GenerateSignals processes new market data and generates signals
func (ei *EngineIntegration) GenerateSignals(
	results *indicators.IndicatorResults,
	klines []indicators.Kline,
	symbol string,
) []indicators.StrategySignal {
	// Get higher timeframe data for multi-TF analysis
	var higherTFResults *indicators.IndicatorResults
	if ei.config.EnableMultiTF {
		higherTFResults = ei.getHigherTimeframeData(symbol)
	}
	
	// Generate signals using strategy logic
	signals := GenerateStrategySignals(results, klines, ei.config, higherTFResults)
	
	// Update strategy statistics
	ei.totalSignalsGenerated += len(signals)
	for _, signal := range signals {
		ei.totalConfidence += signal.Confidence
		
		// Count premium signals (high confidence)
		if signal.Confidence >= ei.config.PremiumConfidence {
			ei.premiumSignalsCount++
		}
	}
	
	return signals
}

// OnPositionOpened notifies strategy when position is opened
func (ei *EngineIntegration) OnPositionOpened(
	direction string,
	entryPrice float64,
	signal indicators.StrategySignal,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Determine signal strength
	var signalStrength SignalStrength
	if signal.Confidence >= ei.config.PremiumConfidence {
		signalStrength = SignalPremium
	} else {
		signalStrength = SignalMinimal
	}
	
	// Determine trend classification
	trendClassification := "TREND"
	if signal.Type == indicators.CounterTrendSignal {
		trendClassification = "COUNTER_TREND"
	}
	
	return ei.behavioralMM.OnPositionOpened(
		direction,
		entryPrice,
		signalStrength,
		indicatorResults,
		trendClassification,
	)
}

// OnPositionClosed notifies strategy when position is closed
func (ei *EngineIntegration) OnPositionClosed() {
	ei.behavioralMM.OnPositionClosed()
}

// ProcessMarkerEvent processes new candle/marker events (standard frequency)
func (ei *EngineIntegration) ProcessMarkerEvent(
	currentPrice float64,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Always process marker events for normal trailing stop updates
	return ei.behavioralMM.OnTickUpdate(currentPrice, indicatorResults)
}

// ProcessTradeEvent processes individual trade events (tick-by-tick when monitoring active)
func (ei *EngineIntegration) ProcessTradeEvent(
	currentPrice float64,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Only process if tick-by-tick monitoring is active
	if !ei.behavioralMM.IsMonitoringActive() {
		return nil
	}
	
	ei.tickProcessingCount++
	return ei.behavioralMM.OnTickUpdate(currentPrice, indicatorResults)
}

// ProcessZoneEvents processes zone events from engine
func (ei *EngineIntegration) ProcessZoneEvents(
	zoneEvents []indicators.ZoneEvent,
	currentPrice float64,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Zone events are handled internally by the behavioral MM
	// This method provides compatibility with engine architecture
	
	for _, event := range zoneEvents {
		fmt.Printf("[STOCH Strategy] 🌊 Zone Event: %s - %s\n", event.ZoneType, event.Type)
	}
	
	// Process through normal tick update which handles zone detection internally
	return ei.ProcessTradeEvent(currentPrice, indicatorResults)
}

// ShouldTriggerStop checks if current price should trigger stop
func (ei *EngineIntegration) ShouldTriggerStop(currentPrice float64) bool {
	return ei.behavioralMM.ShouldTriggerStop(currentPrice)
}

// GetCurrentStopPrice returns current stop price
func (ei *EngineIntegration) GetCurrentStopPrice() float64 {
	return ei.behavioralMM.GetCurrentStopPrice()
}

// GetStrategyStatus returns comprehensive strategy status
func (ei *EngineIntegration) GetStrategyStatus() StrategyStatus {
	status := ei.behavioralMM.GetStatus()
	
	// Add engine-level statistics
	status.TotalSignalsGenerated = ei.totalSignalsGenerated
	status.PremiumSignalsCount = ei.premiumSignalsCount
	status.TickProcessingCount = ei.tickProcessingCount
	
	// Calculate average confidence
	if ei.totalSignalsGenerated > 0 {
		status.AverageConfidence = ei.totalConfidence / float64(ei.totalSignalsGenerated)
	}
	
	return status
}

// UpdateConfig updates strategy configuration at runtime
func (ei *EngineIntegration) UpdateConfig(newConfig StrategyConfig) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid new config: %w", err)
	}
	
	// Create new behavioral MM with new config
	newBMM, err := NewBehavioralMoneyManager(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new behavioral MM: %w", err)
	}
	
	// Transfer position state if active
	oldStatus := ei.behavioralMM.GetStatus()
	if oldStatus.IsActive {
		// Note: In production, you might want to transfer the complete state
		fmt.Printf("[STOCH Strategy] ⚙️  Configuration updated with active position\n")
	}
	
	// Set up callbacks for new instance
	newBMM.SetCallbacks(
		ei.onTrailingAdjustment,
		ei.onEarlyExit,
		ei.onStateChange,
	)
	
	// Replace components
	ei.config = newConfig
	ei.behavioralMM = newBMM
	
	fmt.Printf("[STOCH Strategy] ✅ Configuration updated successfully\n")
	return nil
}

// GetConfiguration returns current strategy configuration
func (ei *EngineIntegration) GetConfiguration() StrategyConfig {
	return ei.config
}

// onTrailingAdjustment handles trailing stop adjustments from behavioral MM
func (ei *EngineIntegration) onTrailingAdjustment(adjustment TrailingAdjustment) {
	if adjustment.Success && ei.onStopAdjustment != nil {
		err := ei.onStopAdjustment(adjustment.NewStopPrice)
		if err != nil {
			fmt.Printf("[STOCH Strategy] ⚠️  Failed to update engine stop price: %v\n", err)
		} else {
			fmt.Printf("[STOCH Strategy] ✅ Engine stop updated: %.2f → %.2f (%s)\n",
				adjustment.OldStopPrice, adjustment.NewStopPrice, adjustment.Reason)
		}
	}
}

// onEarlyExit handles early exit decisions from behavioral MM
func (ei *EngineIntegration) onEarlyExit(decision EarlyExitDecision) error {
	if decision.ShouldExit && ei.onPositionClose != nil {
		fmt.Printf("[STOCH Strategy] 🚪 Early exit triggered: %s\n", decision.Reason)
		return ei.onPositionClose(fmt.Sprintf("EARLY_EXIT: %s", decision.Reason))
	}
	return nil
}

// onStateChange handles monitoring state changes from behavioral MM
func (ei *EngineIntegration) onStateChange(newState MonitoringState) {
	switch newState {
	case StateSTOCHInverse:
		fmt.Printf("[STOCH Strategy] ⚡ Activated tick-by-tick monitoring\n")
	case StateTripleInverse:
		fmt.Printf("[STOCH Strategy] 🚨 Triple inverse protection activated\n")
	case StateNormal:
		fmt.Printf("[STOCH Strategy] ✅ Returned to normal monitoring\n")
	}
}

// getHigherTimeframeData retrieves higher timeframe data for multi-TF analysis
func (ei *EngineIntegration) getHigherTimeframeData(symbol string) *indicators.IndicatorResults {
	// Check cache validity (cache for 1 minute)
	if time.Since(ei.lastCacheTime) > time.Minute {
		ei.higherTFCache = make(map[string]*indicators.IndicatorResults)
		ei.lastCacheTime = time.Now()
	}
	
	// Return cached data if available
	if cached, exists := ei.higherTFCache[symbol]; exists {
		return cached
	}
	
	// In a full implementation, this would fetch higher TF data from engine
	// For now, return nil (strategy will work without multi-TF)
	return nil
}

// PrintStatus prints current strategy status for debugging
func (ei *EngineIntegration) PrintStatus() {
	status := ei.GetStrategyStatus()
	
	fmt.Printf("[STOCH Strategy] 📊 Status Report:\n")
	fmt.Printf("  Active: %t | Monitoring: %s\n", status.IsActive, status.MonitoringState)
	
	if status.IsActive {
		pos := status.PositionState
		fmt.Printf("  Position: %s at %.2f | Stop: %.2f (%.2f%%)\n",
			pos.Direction, pos.EntryPrice, pos.CurrentStopPrice, pos.CurrentTrailing)
		fmt.Printf("  Signal: %s | Trend: %s\n", pos.SignalStrength, pos.TrendClassification)
		fmt.Printf("  Adjustments: %d | Cumulative: %.2f%%\n", 
			pos.TotalAdjustments, pos.CumulativeAdjust)
	}
	
	fmt.Printf("  Signals Generated: %d | Premium: %d | Avg Confidence: %.2f\n",
		status.TotalSignalsGenerated, status.PremiumSignalsCount, status.AverageConfidence)
	fmt.Printf("  Tick Processing Count: %d\n", status.TickProcessingCount)
}

// ValidateIntegration validates that the integration is properly configured
func (ei *EngineIntegration) ValidateIntegration() error {
	if ei.behavioralMM == nil {
		return fmt.Errorf("behavioral MM not initialized")
	}
	
	if err := ei.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	if ei.onPositionClose == nil || ei.onStopAdjustment == nil {
		return fmt.Errorf("engine callbacks not set")
	}
	
	return nil
}

// GetPerformanceMetrics returns performance metrics for monitoring
func (ei *EngineIntegration) GetPerformanceMetrics() map[string]interface{} {
	status := ei.GetStrategyStatus()
	
	return map[string]interface{}{
		"signals_generated":     status.TotalSignalsGenerated,
		"premium_signals_count": status.PremiumSignalsCount,
		"average_confidence":    status.AverageConfidence,
		"tick_processing_count": status.TickProcessingCount,
		"is_monitoring_active":  ei.behavioralMM.IsMonitoringActive(),
		"monitoring_state":      status.MonitoringState,
		"total_adjustments":     status.PositionState.TotalAdjustments,
		"cumulative_adjust":     status.PositionState.CumulativeAdjust,
	}
}

// ResetStatistics resets strategy statistics (useful for backtesting)
func (ei *EngineIntegration) ResetStatistics() {
	ei.totalSignalsGenerated = 0
	ei.premiumSignalsCount = 0
	ei.totalConfidence = 0.0
	ei.tickProcessingCount = 0
	
	fmt.Printf("[STOCH Strategy] 📊 Statistics reset\n")
}

// IsMonitoringActive returns true if tick-by-tick monitoring is active
func (ei *EngineIntegration) IsMonitoringActive() bool {
	if ei.behavioralMM == nil {
		return false
	}
	return ei.behavioralMM.IsMonitoringActive()
}
