package macd_cci_dmi

import (
	"fmt"
	"time"
	
	"agent-economique/internal/indicators"
)

// EngineIntegration provides integration points between Behavioral MM and Engine
type EngineIntegration struct {
	behavioralMM *BehavioralMoneyManager
	
	// Engine callbacks
	onPositionClose func(reason string) error
	onStopAdjustment func(newStopPrice float64) error
	
	// State tracking
	lastProcessedTime time.Time
}

// NewEngineIntegration creates integration between Behavioral MM and Engine
func NewEngineIntegration(config BehavioralConfig) (*EngineIntegration, error) {
	bmm, err := NewBehavioralMoneyManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create behavioral MM: %w", err)
	}
	
	integration := &EngineIntegration{
		behavioralMM: bmm,
	}
	
	// Set up behavioral MM callbacks
	bmm.SetCallbacks(
		integration.onTrailingAdjustment,
		integration.onEarlyExit,
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

// OnPositionOpened notifies behavioral MM when position is opened
func (ei *EngineIntegration) OnPositionOpened(
	direction string,
	entryPrice float64,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Convert engine indicator results to behavioral MM format
	snapshot := ei.convertIndicatorResults(indicatorResults)
	
	return ei.behavioralMM.OnPositionOpened(direction, entryPrice, snapshot)
}

// OnPositionClosed notifies behavioral MM when position is closed
func (ei *EngineIntegration) OnPositionClosed() {
	ei.behavioralMM.OnPositionClosed()
}

// OnIndicatorUpdate processes new indicator values from engine
func (ei *EngineIntegration) OnIndicatorUpdate(
	currentPrice float64,
	indicatorResults *indicators.IndicatorResults,
) error {
	// Convert engine indicator results
	snapshot := ei.convertIndicatorResults(indicatorResults)
	
	// Update behavioral MM
	return ei.behavioralMM.UpdateIndicators(currentPrice, snapshot)
}

// ShouldTriggerStop checks if current price should trigger stop
func (ei *EngineIntegration) ShouldTriggerStop(currentPrice float64) bool {
	return ei.behavioralMM.ShouldTriggerStop(currentPrice)
}

// GetCurrentStopPrice returns current stop price
func (ei *EngineIntegration) GetCurrentStopPrice() float64 {
	return ei.behavioralMM.GetCurrentStopPrice()
}

// GetStatus returns behavioral MM status
func (ei *EngineIntegration) GetStatus() BehavioralMMStatus {
	return ei.behavioralMM.GetStatus()
}

// onTrailingAdjustment handles trailing stop adjustments
func (ei *EngineIntegration) onTrailingAdjustment(adjustment TrailingAdjustment) {
	if adjustment.Success && ei.onStopAdjustment != nil {
		err := ei.onStopAdjustment(adjustment.NewStopPrice)
		if err != nil {
			fmt.Printf("[Behavioral MM] ⚠️  Failed to update engine stop price: %v\n", err)
		} else {
			fmt.Printf("[Behavioral MM] ✅ Engine stop updated: %.2f → %.2f (reason: %s)\n",
				adjustment.OldStopPrice, adjustment.NewStopPrice, adjustment.Reason)
		}
	}
}

// onEarlyExit handles early exit decisions
func (ei *EngineIntegration) onEarlyExit(decision EarlyExitDecision) error {
	if decision.ShouldExit && ei.onPositionClose != nil {
		fmt.Printf("[Behavioral MM] 🚪 Early exit triggered: %s\n", decision.Reason)
		return ei.onPositionClose(fmt.Sprintf("EARLY_EXIT: %s", decision.Reason))
	}
	return nil
}

// convertIndicatorResults converts engine indicator results to behavioral MM format
func (ei *EngineIntegration) convertIndicatorResults(results *indicators.IndicatorResults) IndicatorSnapshot {
	snapshot := IndicatorSnapshot{
		Timestamp: time.Now().UTC(),
	}
	
	// Extract MACD values
	if results.MACD != nil {
		snapshot.MACDValue = results.MACD.MACD
		snapshot.MACDSignal = results.MACD.Signal
	}
	
	// Extract CCI values
	if results.CCI != nil {
		snapshot.CCIValue = results.CCI.Value
	}
	
	// Extract DMI values
	if results.DMI != nil {
		snapshot.DIPlus = results.DMI.PlusDI
		snapshot.DIMinus = results.DMI.MinusDI
		snapshot.ADXValue = results.DMI.ADX
	}
	
	return snapshot
}

// PrintStatus prints current behavioral MM status for debugging
func (ei *EngineIntegration) PrintStatus() {
	status := ei.GetStatus()
	
	if !status.IsActive {
		fmt.Printf("[Behavioral MM] 💤 Inactive\n")
		return
	}
	
	pos := status.PositionState
	fmt.Printf("[Behavioral MM] 📊 Active: %s at %.2f | Stop: %.2f (%.2f%%) | Adjustments: %d\n",
		pos.Direction, pos.EntryPrice, pos.CurrentStopPrice, 
		pos.CurrentTrailingPercent, pos.TotalAdjustments)
	
	if len(status.RecentAdjustments) > 0 {
		latest := status.RecentAdjustments[len(status.RecentAdjustments)-1]
		fmt.Printf("[Behavioral MM] 🔄 Last: %s at %s (%.2f%% → %.2f%%)\n",
			latest.Reason, latest.Timestamp.Format("15:04:05"), 
			latest.OldTrailingPercent, latest.NewTrailingPercent)
	}
}

// ProcessEngineZoneEvents processes zone events from engine
func (ei *EngineIntegration) ProcessEngineZoneEvents(zoneEvents []indicators.ZoneEvent, currentPrice float64) error {
	if !ei.behavioralMM.GetStatus().IsActive {
		return nil // No position to manage
	}
	
	// Process each zone event
	for _, event := range zoneEvents {
		err := ei.processZoneEvent(event, currentPrice)
		if err != nil {
			fmt.Printf("[Behavioral MM] ⚠️  Error processing zone event %s: %v\n", event.Type, err)
		}
	}
	
	return nil
}

// processZoneEvent handles individual zone events
func (ei *EngineIntegration) processZoneEvent(event indicators.ZoneEvent, currentPrice float64) error {
	switch event.ZoneType {
	case "CCI_INVERSE":
		// CCI moved to inverse zone - handled by UpdateIndicators
		fmt.Printf("[Behavioral MM] 🔄 CCI Zone Event: %s\n", event.ZoneType)
		
	case "MACD_INVERSE":
		// MACD inverse signal - handled by UpdateIndicators
		fmt.Printf("[Behavioral MM] 🔄 MACD Event: %s\n", event.ZoneType)
		
	case "DI_COUNTER":
		// DMI counter trend - handled by UpdateIndicators
		fmt.Printf("[Behavioral MM] 🔄 DMI Event: %s\n", event.ZoneType)
		
	default:
		// Unknown event type
		fmt.Printf("[Behavioral MM] ❓ Unknown zone event: %s\n", event.ZoneType)
	}
	
	return nil
}

// GetBehavioralConfig returns current behavioral configuration (for engine inspection)
func (ei *EngineIntegration) GetBehavioralConfig() BehavioralConfig {
	return ei.behavioralMM.config
}

// UpdateBehavioralConfig updates behavioral configuration at runtime
func (ei *EngineIntegration) UpdateBehavioralConfig(newConfig BehavioralConfig) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid behavioral config: %w", err)
	}
	
	// Create new behavioral MM with new config
	newBMM, err := NewBehavioralMoneyManager(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new behavioral MM: %w", err)
	}
	
	// Transfer state if position is open
	if ei.behavioralMM.GetStatus().IsActive {
		// Note: In production, you might want to transfer the position state
		// For now, we'll just log the config change
		fmt.Printf("[Behavioral MM] ⚙️  Configuration updated (position active)\n")
	}
	
	// Set up callbacks for new instance
	newBMM.SetCallbacks(
		ei.onTrailingAdjustment,
		ei.onEarlyExit,
	)
	
	// Replace behavioral MM
	ei.behavioralMM = newBMM
	
	fmt.Printf("[Behavioral MM] ✅ Configuration updated successfully\n")
	return nil
}
