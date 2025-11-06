package stoch_mfi_cci

import (
	"fmt"
	"math"
	"sync"
	"time"
	
	"agent-economique/internal/indicators"
)

// BehavioralMoneyManager manages dynamic trailing stops for STOCH/MFI/CCI strategy
type BehavioralMoneyManager struct {
	config            StrategyConfig
	positionState     PositionState
	adjustmentHistory []TrailingAdjustment
	
	// State tracking
	lastIndicatorSnapshot IndicatorSnapshot
	
	// Callbacks for engine integration
	onTrailingAdjustment func(TrailingAdjustment)
	onEarlyExit         func(EarlyExitDecision) error
	onStateChange       func(MonitoringState)
	
	// Thread safety
	mutex sync.RWMutex
}

// NewBehavioralMoneyManager creates a new behavioral MM instance
func NewBehavioralMoneyManager(config StrategyConfig) (*BehavioralMoneyManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	bmm := &BehavioralMoneyManager{
		config:            config,
		adjustmentHistory: make([]TrailingAdjustment, 0),
		positionState:     PositionState{IsOpen: false},
	}
	
	return bmm, nil
}

// SetCallbacks sets callback functions for engine integration
func (bmm *BehavioralMoneyManager) SetCallbacks(
	onTrailingAdjustment func(TrailingAdjustment),
	onEarlyExit func(EarlyExitDecision) error,
	onStateChange func(MonitoringState),
) {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	bmm.onTrailingAdjustment = onTrailingAdjustment
	bmm.onEarlyExit = onEarlyExit
	bmm.onStateChange = onStateChange
}

// OnPositionOpened is called when a new position is opened
func (bmm *BehavioralMoneyManager) OnPositionOpened(
	direction string,
	entryPrice float64,
	signalStrength SignalStrength,
	indicatorResults *indicators.IndicatorResults,
	trendClassification string,
) error {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	// Convert indicator results to snapshot
	snapshot := ConvertIndicatorResults(indicatorResults)
	
	// Determine initial trailing stop
	initialTrailing := bmm.calculateInitialTrailingStop(trendClassification)
	stopPrice := bmm.calculateStopPrice(entryPrice, initialTrailing, direction)
	
	bmm.positionState = PositionState{
		IsOpen:         true,
		EntryTime:      time.Now().UTC(),
		EntryPrice:     entryPrice,
		Direction:      direction,
		SignalStrength: signalStrength,
		
		// Entry conditions from indicators
		EntrySTOCHK:       snapshot.STOCHValue,
		EntrySTOCHD:       snapshot.STOCHSignal,
		EntrySTOCHZone:    snapshot.STOCHZone,
		EntryMFIValue:     snapshot.MFIValue,
		EntryMFIZone:      snapshot.MFIZone,
		EntryCCIValue:     snapshot.CCIValue,
		EntryCCIZone:      snapshot.CCIZone,
		
		// Multi-timeframe context
		TrendClassification: trendClassification,
		
		// Trailing state
		InitialTrailing:     initialTrailing,
		CurrentTrailing:     initialTrailing,
		CurrentStopPrice:    stopPrice,
		
		// Monitoring state
		MonitoringState:    StateNormal,
		LastAdjustmentTime: time.Now().UTC(),
		TotalAdjustments:   0,
		CumulativeAdjust:   0.0,
	}
	
	bmm.lastIndicatorSnapshot = snapshot
	
	fmt.Printf("[STOCH MM] 📈 Position opened: %s at %.2f, initial trailing: %.2f%%, stop: %.2f\n",
		direction, entryPrice, initialTrailing, stopPrice)
	
	return nil
}

// OnPositionClosed is called when position is closed
func (bmm *BehavioralMoneyManager) OnPositionClosed() {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	fmt.Printf("[STOCH MM] 📊 Position closed after %d adjustments, cumulative: %.2f%%\n", 
		bmm.positionState.TotalAdjustments, bmm.positionState.CumulativeAdjust)
	
	// Notify state change
	if bmm.onStateChange != nil {
		bmm.onStateChange(StateNormal)
	}
	
	bmm.positionState = PositionState{IsOpen: false}
}

// OnTickUpdate processes tick-by-tick updates (called on every trade when monitoring active)
func (bmm *BehavioralMoneyManager) OnTickUpdate(currentPrice float64, indicatorResults *indicators.IndicatorResults) error {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	if !bmm.positionState.IsOpen {
		return nil // No position to manage
	}
	
	// Convert current indicators
	currentSnapshot := ConvertIndicatorResults(indicatorResults)
	
	// Determine current monitoring state
	newState := GetMonitoringState(currentSnapshot, &bmm.positionState)
	
	// Handle state transitions
	if newState != bmm.positionState.MonitoringState {
		bmm.handleStateTransition(bmm.positionState.MonitoringState, newState)
		bmm.positionState.MonitoringState = newState
	}
	
	// Process adjustments based on current state
	switch newState {
	case StateNormal:
		// Normal trailing stop updates (less frequent)
		bmm.updateNormalTrailingStop(currentPrice)
		
	case StateSTOCHInverse:
		// STOCH inverse detected - activate tick-by-tick monitoring
		bmm.processSTOCHInverseAdjustments(currentPrice, currentSnapshot)
		
	case StateTripleInverse:
		// Triple inverse - maximum protection
		bmm.processTripleInverseAdjustments(currentPrice, currentSnapshot)
	}
	
	// Check for early exit conditions
	if bmm.config.EnableEarlyExit {
		decision := bmm.evaluateEarlyExit(currentPrice, currentSnapshot)
		if decision.ShouldExit && bmm.onEarlyExit != nil {
			return bmm.onEarlyExit(decision)
		}
	}
	
	// Update last snapshot
	bmm.lastIndicatorSnapshot = currentSnapshot
	
	return nil
}

// calculateInitialTrailingStop determines initial trailing stop percentage
func (bmm *BehavioralMoneyManager) calculateInitialTrailingStop(trendClassification string) float64 {
	switch trendClassification {
	case "TREND":
		return bmm.config.TrendTrailingPercent
	case "COUNTER_TREND":
		return bmm.config.CounterTrendTrailing
	default:
		return bmm.config.BaseTrailingPercent
	}
}

// handleStateTransition handles monitoring state changes
func (bmm *BehavioralMoneyManager) handleStateTransition(oldState, newState MonitoringState) {
	fmt.Printf("[STOCH MM] 🔄 State transition: %s → %s\n", oldState, newState)
	
	if bmm.onStateChange != nil {
		bmm.onStateChange(newState)
	}
	
	switch newState {
	case StateSTOCHInverse:
		fmt.Printf("[STOCH MM] ⚡ Activating tick-by-tick monitoring (STOCH inverse detected)\n")
	case StateTripleInverse:
		fmt.Printf("[STOCH MM] 🚨 TRIPLE INVERSE detected - Maximum protection mode\n")
	case StateNormal:
		fmt.Printf("[STOCH MM] ✅ Returning to normal monitoring\n")
	}
}

// processSTOCHInverseAdjustments handles STOCH inverse adjustments
func (bmm *BehavioralMoneyManager) processSTOCHInverseAdjustments(currentPrice float64, snapshot IndicatorSnapshot) {
	intensity := CalculateInverseIntensity(snapshot, &bmm.positionState)
	adjustmentPercent := GetInverseAdjustmentMultiplier(intensity, bmm.config)
	
	if adjustmentPercent > 0 {
		reason := AdjustmentSTOCHInverse
		if intensity >= 2 {
			reason = AdjustmentTripleInverse
		}
		
		bmm.makeAdjustment(currentPrice, reason, adjustmentPercent, snapshot)
	}
}

// processTripleInverseAdjustments handles triple inverse maximum protection
func (bmm *BehavioralMoneyManager) processTripleInverseAdjustments(currentPrice float64, snapshot IndicatorSnapshot) {
	// Apply maximum protection adjustment
	bmm.makeAdjustment(currentPrice, AdjustmentTripleInverse, bmm.config.TripleInverseAdjust, snapshot)
	
	// Check for emergency early exit if enabled
	if bmm.config.TripleInverseEarlyExit {
		profit := bmm.positionState.CalculateCurrentProfit(currentPrice)
		if profit >= bmm.config.MinProfitForEarlyExit {
			decision := EarlyExitDecision{
				Timestamp:         time.Now().UTC(),
				ShouldExit:        true,
				Reason:            "Triple inverse with minimum profit achieved",
				CurrentPrice:      currentPrice,
				ProfitPercent:     profit,
				IndicatorSnapshot: snapshot,
				TriggerCondition:  "TRIPLE_INVERSE_EMERGENCY",
			}
			
			if bmm.onEarlyExit != nil {
				bmm.onEarlyExit(decision)
			}
		}
	}
}

// makeAdjustment applies a trailing stop adjustment
func (bmm *BehavioralMoneyManager) makeAdjustment(
	currentPrice float64,
	reason AdjustmentReason,
	adjustmentPercent float64,
	snapshot IndicatorSnapshot,
) {
	// Check cumulative adjustment limits
	if bmm.positionState.CumulativeAdjust+adjustmentPercent > bmm.config.MaxCumulativeAdjust {
		adjustmentPercent = bmm.config.MaxCumulativeAdjust - bmm.positionState.CumulativeAdjust
		if adjustmentPercent <= 0 {
			return // No more adjustments allowed
		}
	}
	
	oldTrailing := bmm.positionState.CurrentTrailing
	oldStop := bmm.positionState.CurrentStopPrice
	
	// Calculate new trailing percentage (tighter)
	newTrailing := oldTrailing - adjustmentPercent
	newTrailing = math.Max(bmm.config.MinTrailingPercent, newTrailing)
	
	// Calculate new stop price
	newStop := bmm.calculateStopPrice(currentPrice, newTrailing, bmm.positionState.Direction)
	
	// Verify this is a beneficial adjustment
	beneficial := bmm.isBeneficialAdjustment(oldStop, newStop, bmm.positionState.Direction)
	
	adjustment := TrailingAdjustment{
		Timestamp:            time.Now().UTC(),
		Reason:               reason,
		OldTrailingPercent:   oldTrailing,
		NewTrailingPercent:   newTrailing,
		OldStopPrice:         oldStop,
		NewStopPrice:         newStop,
		CurrentPrice:         currentPrice,
		CurrentProfitPercent: bmm.positionState.CalculateCurrentProfit(currentPrice),
		IndicatorSnapshot:    snapshot,
		CumulativeAdjustment: bmm.positionState.CumulativeAdjust + adjustmentPercent,
		Success:              beneficial,
	}
	
	if beneficial {
		bmm.positionState.CurrentTrailing = newTrailing
		bmm.positionState.CurrentStopPrice = newStop
		bmm.positionState.LastAdjustmentTime = time.Now().UTC()
		bmm.positionState.TotalAdjustments++
		bmm.positionState.CumulativeAdjust += adjustmentPercent
		
		fmt.Printf("[STOCH MM] ⚡ %s adjustment: %.2f%% → %.2f%%, Stop %.2f → %.2f\n",
			reason, oldTrailing, newTrailing, oldStop, newStop)
	} else {
		adjustment.ErrorMessage = "Adjustment would not improve stop price"
	}
	
	// Record adjustment
	bmm.adjustmentHistory = append(bmm.adjustmentHistory, adjustment)
	
	// Limit history size
	if len(bmm.adjustmentHistory) > 100 {
		bmm.adjustmentHistory = bmm.adjustmentHistory[len(bmm.adjustmentHistory)-50:]
	}
	
	// Notify callback
	if bmm.onTrailingAdjustment != nil {
		bmm.onTrailingAdjustment(adjustment)
	}
}

// updateNormalTrailingStop updates trailing stop in normal mode
func (bmm *BehavioralMoneyManager) updateNormalTrailingStop(currentPrice float64) {
	currentPercent := bmm.positionState.CurrentTrailing
	newStop := bmm.calculateStopPrice(currentPrice, currentPercent, bmm.positionState.Direction)
	
	// Only update if new stop is better (favorable price movement)
	if bmm.isBeneficialAdjustment(bmm.positionState.CurrentStopPrice, newStop, bmm.positionState.Direction) {
		bmm.positionState.CurrentStopPrice = newStop
	}
}

// evaluateEarlyExit evaluates early exit conditions
func (bmm *BehavioralMoneyManager) evaluateEarlyExit(currentPrice float64, snapshot IndicatorSnapshot) EarlyExitDecision {
	decision := EarlyExitDecision{
		Timestamp:         time.Now().UTC(),
		ShouldExit:        false,
		CurrentPrice:      currentPrice,
		ProfitPercent:     bmm.positionState.CalculateCurrentProfit(currentPrice),
		IndicatorSnapshot: snapshot,
	}
	
	// Check minimum profit requirement
	if decision.ProfitPercent < bmm.config.MinProfitForEarlyExit {
		decision.Reason = fmt.Sprintf("Profit %.2f%% below minimum %.2f%%", 
			decision.ProfitPercent, bmm.config.MinProfitForEarlyExit)
		return decision
	}
	
	// Triple inverse early exit
	if bmm.config.TripleInverseEarlyExit {
		entrySnapshot := IndicatorSnapshot{
			STOCHZone: bmm.positionState.EntrySTOCHZone,
			MFIZone:   bmm.positionState.EntryMFIZone,
			CCIZone:   bmm.positionState.EntryCCIZone,
		}
		
		if DetectTripleInverse(entrySnapshot, snapshot, bmm.positionState.Direction) {
			decision.ShouldExit = true
			decision.Reason = "Triple inverse detected with sufficient profit"
			decision.TriggerCondition = "TRIPLE_INVERSE"
		}
	}
	
	return decision
}

// calculateStopPrice calculates stop price based on current price and trailing percentage
func (bmm *BehavioralMoneyManager) calculateStopPrice(currentPrice, trailingPercent float64, direction string) float64 {
	switch direction {
	case "LONG":
		return currentPrice * (1 - trailingPercent/100)
	case "SHORT":
		return currentPrice * (1 + trailingPercent/100)
	default:
		return currentPrice
	}
}

// isBeneficialAdjustment checks if new stop is better than old stop
func (bmm *BehavioralMoneyManager) isBeneficialAdjustment(oldStop, newStop float64, direction string) bool {
	switch direction {
	case "LONG":
		return newStop > oldStop // Higher stop is better for LONG
	case "SHORT":
		return newStop < oldStop // Lower stop is better for SHORT
	default:
		return false
	}
}

// ShouldTriggerStop checks if current price should trigger stop
func (bmm *BehavioralMoneyManager) ShouldTriggerStop(currentPrice float64) bool {
	bmm.mutex.RLock()
	defer bmm.mutex.RUnlock()
	
	return bmm.positionState.ShouldTriggerStop(currentPrice)
}

// GetCurrentStopPrice returns current stop price
func (bmm *BehavioralMoneyManager) GetCurrentStopPrice() float64 {
	bmm.mutex.RLock()
	defer bmm.mutex.RUnlock()
	
	return bmm.positionState.CurrentStopPrice
}

// GetStatus returns current behavioral MM status
func (bmm *BehavioralMoneyManager) GetStatus() StrategyStatus {
	bmm.mutex.RLock()
	defer bmm.mutex.RUnlock()
	
	// Get recent adjustments (last 10)
	recentAdjustments := bmm.adjustmentHistory
	if len(recentAdjustments) > 10 {
		recentAdjustments = recentAdjustments[len(recentAdjustments)-10:]
	}
	
	return StrategyStatus{
		IsActive:              bmm.positionState.IsOpen,
		PositionState:         bmm.positionState,
		MonitoringState:       bmm.positionState.MonitoringState,
		RecentAdjustments:     recentAdjustments,
		TotalSignalsGenerated: 0, // Will be tracked by strategy manager
		PremiumSignalsCount:   0, // Will be tracked by strategy manager
		AverageConfidence:     0, // Will be tracked by strategy manager
		TickProcessingCount:   0, // Will be tracked by strategy manager
	}
}

// IsMonitoringActive returns true if tick-by-tick monitoring is active
func (bmm *BehavioralMoneyManager) IsMonitoringActive() bool {
	bmm.mutex.RLock()
	defer bmm.mutex.RUnlock()
	
	return bmm.positionState.IsOpen && 
		(bmm.positionState.MonitoringState == StateSTOCHInverse || 
		 bmm.positionState.MonitoringState == StateTripleInverse)
}
