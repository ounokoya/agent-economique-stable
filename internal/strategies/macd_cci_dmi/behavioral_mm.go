package macd_cci_dmi

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// BehavioralMoneyManager manages behavioral MM for MACD/CCI/DMI strategy
type BehavioralMoneyManager struct {
	config         BehavioralConfig
	positionState  PositionState
	adjustmentHistory []TrailingAdjustment
	
	// State tracking
	lastIndicatorValues IndicatorSnapshot
	
	// Callbacks for actions
	onTrailingAdjustment func(adjustment TrailingAdjustment)
	onEarlyExit         func(decision EarlyExitDecision) error
	
	// Thread safety
	mutex sync.RWMutex
}

// NewBehavioralMoneyManager creates a new behavioral MM instance
func NewBehavioralMoneyManager(config BehavioralConfig) (*BehavioralMoneyManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid behavioral config: %w", err)
	}
	
	bmm := &BehavioralMoneyManager{
		config:            config,
		adjustmentHistory: make([]TrailingAdjustment, 0),
	}
	
	return bmm, nil
}

// SetCallbacks sets callback functions for actions
func (bmm *BehavioralMoneyManager) SetCallbacks(
	onTrailingAdjustment func(TrailingAdjustment),
	onEarlyExit func(EarlyExitDecision) error,
) {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	bmm.onTrailingAdjustment = onTrailingAdjustment
	bmm.onEarlyExit = onEarlyExit
}

// OnPositionOpened is called when a new position is opened
func (bmm *BehavioralMoneyManager) OnPositionOpened(
	direction string,
	entryPrice float64,
	indicators IndicatorSnapshot,
) error {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	// Determine initial trailing stop based on DMI trend
	initialTrailing := bmm.calculateInitialTrailingStop(indicators)
	
	// Determine entry conditions
	cciZone := DetermineCCIZone(indicators.CCIValue, bmm.config.CCIMonitoringThreshold)
	dmiTrend := DetermineTrendType(indicators.DIPlus, indicators.DIMinus)
	
	// Calculate initial stop price
	stopPrice := bmm.calculateStopPrice(entryPrice, initialTrailing, direction)
	
	bmm.positionState = PositionState{
		IsOpen:                 true,
		EntryTime:              time.Now().UTC(),
		EntryPrice:             entryPrice,
		Direction:              direction,
		InitialTrailingPercent: initialTrailing,
		CurrentTrailingPercent: initialTrailing,
		CurrentStopPrice:       stopPrice,
		
		// Entry conditions
		EntryMACDValue:         indicators.MACDValue,
		EntryMACDSignal:        indicators.MACDSignal,
		EntryCCIValue:          indicators.CCIValue,
		EntryCCIZone:           cciZone,
		EntryDMITrend:          dmiTrend,
		
		// State tracking
		LastAdjustmentTime:     time.Now().UTC(),
		TotalAdjustments:       0,
		LastIndicatorValues:    indicators,
	}
	
	fmt.Printf("[Behavioral MM] 📈 Position opened: %s at %.2f, initial trailing: %.2f%%, stop: %.2f\n",
		direction, entryPrice, initialTrailing, stopPrice)
	
	return nil
}

// OnPositionClosed is called when position is closed
func (bmm *BehavioralMoneyManager) OnPositionClosed() {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	fmt.Printf("[Behavioral MM] 📊 Position closed after %d adjustments\n", 
		bmm.positionState.TotalAdjustments)
	
	bmm.positionState = PositionState{IsOpen: false}
}

// UpdateIndicators updates current indicator values and processes events
func (bmm *BehavioralMoneyManager) UpdateIndicators(currentPrice float64, indicators IndicatorSnapshot) error {
	bmm.mutex.Lock()
	defer bmm.mutex.Unlock()
	
	if !bmm.positionState.IsOpen {
		return nil // No position to manage
	}
	
	// Store previous values for comparison
	previousIndicators := bmm.positionState.LastIndicatorValues
	bmm.positionState.LastIndicatorValues = indicators
	
	// Check for early exit conditions
	if bmm.config.EnableEarlyExit {
		decision := bmm.evaluateEarlyExit(currentPrice, indicators, previousIndicators)
		if decision.ShouldExit {
			if bmm.onEarlyExit != nil {
				return bmm.onEarlyExit(decision)
			}
		}
	}
	
	// Process dynamic adjustments
	if bmm.config.EnableDynamicAdjustments {
		bmm.processDynamicAdjustments(currentPrice, indicators, previousIndicators)
	}
	
	// Process profit grid adjustments
	if bmm.config.EnableProfitGrid {
		bmm.processProfitGridAdjustments(currentPrice)
	}
	
	// Update trailing stop based on current price movement
	bmm.updateTrailingStop(currentPrice)
	
	return nil
}

// calculateInitialTrailingStop determines initial trailing stop percentage
func (bmm *BehavioralMoneyManager) calculateInitialTrailingStop(indicators IndicatorSnapshot) float64 {
	switch bmm.config.TrailingMode {
	case TrailingModeFixed:
		return bmm.config.TrendTrailingPercent
		
	case TrailingModeAdaptive:
		// Use DMI to determine trend vs counter-trend
		if indicators.DIPlus > indicators.DIMinus {
			return bmm.config.TrendTrailingPercent
		}
		return bmm.config.CounterTrendTrailingPercent
		
	case TrailingModeDynamic:
		// Start with adaptive and allow dynamic adjustments
		base := bmm.config.TrendTrailingPercent
		if indicators.DIPlus <= indicators.DIMinus {
			base = bmm.config.CounterTrendTrailingPercent
		}
		
		// Adjust based on volatility (simplified using ADX)
		if indicators.ADXValue > 25 {
			base *= 1.2 // Wider stops in high volatility
		} else if indicators.ADXValue < 15 {
			base *= 0.8 // Tighter stops in low volatility
		}
		
		return math.Max(bmm.config.MinTrailingPercent, 
			math.Min(bmm.config.MaxTrailingPercent, base))
		
	default:
		return bmm.config.TrendTrailingPercent
	}
}

// processDynamicAdjustments handles indicator-based adjustments
func (bmm *BehavioralMoneyManager) processDynamicAdjustments(
	currentPrice float64,
	current, previous IndicatorSnapshot,
) {
	// CCI Zone Inverse Adjustment
	if bmm.detectCCIZoneInverse(current, previous) {
		bmm.makeAdjustment(currentPrice, AdjustmentCCIInverse, bmm.config.CCIZoneAdjustmentPercent)
	}
	
	// MACD Inverse Signal Adjustment
	if bmm.detectMACDInverse(current, previous) {
		// Only adjust if in profit
		profit := bmm.positionState.CalculateCurrentProfit(currentPrice)
		if profit > 0 {
			bmm.makeAdjustment(currentPrice, AdjustmentMACDInverse, bmm.config.MACDInverseAdjustmentPercent)
		}
	}
	
	// DMI Counter-Trend Adjustment
	if bmm.detectDMICounter(current, previous) {
		// Only adjust if in profit
		profit := bmm.positionState.CalculateCurrentProfit(currentPrice)
		if profit > 0 {
			bmm.makeAdjustment(currentPrice, AdjustmentDMICounter, bmm.config.DMICounterAdjustmentPercent)
		}
	}
}

// processProfitGridAdjustments handles profit-based adjustments
func (bmm *BehavioralMoneyManager) processProfitGridAdjustments(currentPrice float64) {
	currentProfit := bmm.positionState.CalculateCurrentProfit(currentPrice)
	
	for _, level := range bmm.config.ProfitAdjustmentLevels {
		if currentProfit >= level.ProfitThresholdPercent {
			// Check if we haven't already adjusted for this level
			if !bmm.hasAdjustedForProfitLevel(level.ProfitThresholdPercent) {
				bmm.makeAdjustment(currentPrice, AdjustmentProfitGrid, level.AdjustmentPercent)
			}
		}
	}
}

// evaluateEarlyExit checks if position should be exited early
func (bmm *BehavioralMoneyManager) evaluateEarlyExit(
	currentPrice float64,
	current, previous IndicatorSnapshot,
) EarlyExitDecision {
	decision := EarlyExitDecision{
		Timestamp:         time.Now().UTC(),
		ShouldExit:        false,
		CurrentPrice:      currentPrice,
		ProfitPercent:     bmm.positionState.CalculateCurrentProfit(currentPrice),
		IndicatorSnapshot: current,
	}
	
	// Check minimum profit requirement
	if decision.ProfitPercent < bmm.config.MinProfitForEarlyExit {
		decision.Reason = fmt.Sprintf("Profit %.2f%% below minimum %.2f%%", 
			decision.ProfitPercent, bmm.config.MinProfitForEarlyExit)
		return decision
	}
	
	// MACD Inverse Early Exit
	if bmm.config.EarlyExitOnMACDInverse && bmm.detectMACDInverse(current, previous) {
		// Additional condition: trailing stop must not be positive yet
		if !bmm.isTrailingStopInProfit(currentPrice) {
			decision.ShouldExit = true
			decision.Reason = "MACD inverse signal detected before trailing stop in profit"
		}
	}
	
	return decision
}

// makeAdjustment applies a trailing stop adjustment
func (bmm *BehavioralMoneyManager) makeAdjustment(
	currentPrice float64,
	reason AdjustmentReason,
	adjustmentPercent float64,
) {
	// Limit adjustment to maximum allowed
	adjustmentPercent = math.Min(adjustmentPercent, bmm.config.MaxAdjustmentPercent)
	
	oldTrailing := bmm.positionState.CurrentTrailingPercent
	oldStop := bmm.positionState.CurrentStopPrice
	
	// Calculate new trailing percentage (tighter)
	newTrailing := oldTrailing - adjustmentPercent
	newTrailing = math.Max(bmm.config.MinTrailingPercent, newTrailing)
	
	// Calculate new stop price
	newStop := bmm.calculateStopPrice(currentPrice, newTrailing, bmm.positionState.Direction)
	
	// Only apply if it's a beneficial adjustment (tighter stop)
	beneficial := false
	switch bmm.positionState.Direction {
	case "LONG":
		beneficial = newStop > oldStop
	case "SHORT":
		beneficial = newStop < oldStop
	}
	
	adjustment := TrailingAdjustment{
		Timestamp:            time.Now().UTC(),
		Reason:               reason,
		OldTrailingPercent:   oldTrailing,
		NewTrailingPercent:   newTrailing,
		OldStopPrice:         oldStop,
		NewStopPrice:         newStop,
		CurrentPrice:         currentPrice,
		CurrentProfitPercent: bmm.positionState.CalculateCurrentProfit(currentPrice),
		IndicatorSnapshot:    bmm.positionState.LastIndicatorValues,
		Success:              beneficial,
	}
	
	if beneficial {
		bmm.positionState.CurrentTrailingPercent = newTrailing
		bmm.positionState.CurrentStopPrice = newStop
		bmm.positionState.LastAdjustmentTime = time.Now().UTC()
		bmm.positionState.TotalAdjustments++
		
		fmt.Printf("[Behavioral MM] ⚡ Adjustment: %s - Trailing %.2f%% → %.2f%%, Stop %.2f → %.2f\n",
			reason, oldTrailing, newTrailing, oldStop, newStop)
	} else {
		adjustment.ErrorMessage = "Adjustment would worsen stop price"
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

// updateTrailingStop updates trailing stop based on favorable price movement
func (bmm *BehavioralMoneyManager) updateTrailingStop(currentPrice float64) {
	currentPercent := bmm.positionState.CurrentTrailingPercent
	newStop := bmm.calculateStopPrice(currentPrice, currentPercent, bmm.positionState.Direction)
	
	// Only update if new stop is better (tighter in favorable direction)
	shouldUpdate := false
	switch bmm.positionState.Direction {
	case "LONG":
		shouldUpdate = newStop > bmm.positionState.CurrentStopPrice
	case "SHORT":
		shouldUpdate = newStop < bmm.positionState.CurrentStopPrice
	}
	
	if shouldUpdate {
		bmm.positionState.CurrentStopPrice = newStop
	}
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

// Helper functions for detection
func (bmm *BehavioralMoneyManager) detectCCIZoneInverse(current, previous IndicatorSnapshot) bool {
	currentZone := DetermineCCIZone(current.CCIValue, bmm.config.CCIMonitoringThreshold)
	entryZone := bmm.positionState.EntryCCIZone
	
	// Detect if CCI moved to opposite zone
	switch bmm.positionState.Direction {
	case "LONG":
		return entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		return entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	default:
		return false
	}
}

func (bmm *BehavioralMoneyManager) detectMACDInverse(current, previous IndicatorSnapshot) bool {
	// Detect MACD signal inverse to entry direction
	macdCrossUp := previous.MACDValue <= previous.MACDSignal && current.MACDValue > current.MACDSignal
	macdCrossDown := previous.MACDValue >= previous.MACDSignal && current.MACDValue < current.MACDSignal
	
	switch bmm.positionState.Direction {
	case "LONG":
		return macdCrossDown // MACD crosses down for LONG position
	case "SHORT":
		return macdCrossUp // MACD crosses up for SHORT position
	default:
		return false
	}
}

func (bmm *BehavioralMoneyManager) detectDMICounter(current, previous IndicatorSnapshot) bool {
	currentTrend := DetermineTrendType(current.DIPlus, current.DIMinus)
	entryTrend := bmm.positionState.EntryDMITrend
	
	// Detect trend change
	return currentTrend != entryTrend
}

func (bmm *BehavioralMoneyManager) hasAdjustedForProfitLevel(profitThreshold float64) bool {
	for _, adj := range bmm.adjustmentHistory {
		if adj.Reason == AdjustmentProfitGrid && 
		   adj.CurrentProfitPercent >= profitThreshold-0.1 && // Small tolerance
		   adj.Success {
			return true
		}
	}
	return false
}

func (bmm *BehavioralMoneyManager) isTrailingStopInProfit(currentPrice float64) bool {
	stopPrice := bmm.positionState.CurrentStopPrice
	entryPrice := bmm.positionState.EntryPrice
	
	switch bmm.positionState.Direction {
	case "LONG":
		return stopPrice > entryPrice
	case "SHORT":
		return stopPrice < entryPrice
	default:
		return false
	}
}

// GetStatus returns current behavioral MM status
func (bmm *BehavioralMoneyManager) GetStatus() BehavioralMMStatus {
	bmm.mutex.RLock()
	defer bmm.mutex.RUnlock()
	
	// Get recent adjustments (last 10)
	recentAdjustments := bmm.adjustmentHistory
	if len(recentAdjustments) > 10 {
		recentAdjustments = recentAdjustments[len(recentAdjustments)-10:]
	}
	
	// Calculate average trailing percent
	avgTrailing := bmm.positionState.CurrentTrailingPercent
	if len(bmm.adjustmentHistory) > 0 {
		total := 0.0
		count := 0
		for _, adj := range bmm.adjustmentHistory {
			if adj.Success {
				total += adj.NewTrailingPercent
				count++
			}
		}
		if count > 0 {
			avgTrailing = total / float64(count)
		}
	}
	
	return BehavioralMMStatus{
		IsActive:               bmm.positionState.IsOpen,
		PositionState:          bmm.positionState,
		RecentAdjustments:      recentAdjustments,
		TotalAdjustmentsMade:   bmm.positionState.TotalAdjustments,
		AverageTrailingPercent: avgTrailing,
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
