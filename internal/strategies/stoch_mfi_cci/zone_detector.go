package stoch_mfi_cci

import (
	"agent-economique/internal/indicators"
)

// detectZoneEvents detects zone monitoring events for STOCH/MFI/CCI strategy
func detectZoneEvents(results *indicators.IndicatorResults, positionCtx *PositionState) []indicators.ZoneEvent {
	var events []indicators.ZoneEvent
	
	if positionCtx == nil || !positionCtx.IsOpen {
		return events
	}
	
	// Convert current indicator results to snapshot
	currentSnapshot := ConvertIndicatorResults(results)
	
	// Detect STOCH inverse zone event
	if stochEvent := detectSTOCHInverseEvent(currentSnapshot, positionCtx); stochEvent != nil {
		events = append(events, *stochEvent)
	}
	
	// Detect MFI inverse zone event
	if mfiEvent := detectMFIInverseEvent(currentSnapshot, positionCtx); mfiEvent != nil {
		events = append(events, *mfiEvent)
	}
	
	// Detect CCI inverse zone event
	if cciEvent := detectCCIInverseEvent(currentSnapshot, positionCtx); cciEvent != nil {
		events = append(events, *cciEvent)
	}
	
	// Detect triple inverse alignment
	if tripleEvent := detectTripleInverseEvent(currentSnapshot, positionCtx); tripleEvent != nil {
		events = append(events, *tripleEvent)
	}
	
	return events
}

// detectSTOCHInverseEvent detects STOCH moving to inverse zone
func detectSTOCHInverseEvent(current IndicatorSnapshot, positionCtx *PositionState) *indicators.ZoneEvent {
	entryZone := positionCtx.EntrySTOCHZone
	currentZone := current.STOCHZone
	direction := positionCtx.Direction
	
	// Check if STOCH moved to inverse zone
	isInverse := DetectSTOCHInverse(entryZone, currentZone, direction)
	
	if isInverse {
		return &indicators.ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "STOCH_INVERSE",
			IsInverse:       true,
			RequiresProfit:  false, // STOCH inverse activates monitoring immediately
			ProfitThreshold: 0.0,
			CurrentProfit:   positionCtx.CalculateCurrentProfit(current.STOCHValue), // Using STOCH K as price proxy
			Timestamp:       current.Timestamp.UnixMilli(),
		}
	}
	
	return nil
}

// detectMFIInverseEvent detects MFI moving to inverse zone
func detectMFIInverseEvent(current IndicatorSnapshot, positionCtx *PositionState) *indicators.ZoneEvent {
	entryZone := positionCtx.EntryMFIZone
	currentZone := current.MFIZone
	direction := positionCtx.Direction
	
	var isInverse bool
	switch direction {
	case "LONG":
		isInverse = entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		isInverse = entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	}
	
	if isInverse {
		return &indicators.ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "MFI_INVERSE",
			IsInverse:       true,
			RequiresProfit:  false, // MFI supports STOCH, no profit requirement
			ProfitThreshold: 0.0,
			CurrentProfit:   0.0, // Will be calculated by behavioral MM
			Timestamp:       current.Timestamp.UnixMilli(),
		}
	}
	
	return nil
}

// detectCCIInverseEvent detects CCI moving to inverse zone
func detectCCIInverseEvent(current IndicatorSnapshot, positionCtx *PositionState) *indicators.ZoneEvent {
	entryZone := positionCtx.EntryCCIZone
	currentZone := current.CCIZone
	direction := positionCtx.Direction
	
	var isInverse bool
	switch direction {
	case "LONG":
		isInverse = entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		isInverse = entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	}
	
	if isInverse {
		return &indicators.ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "CCI_INVERSE",
			IsInverse:       true,
			RequiresProfit:  false, // CCI supports STOCH, no profit requirement
			ProfitThreshold: 0.0,
			CurrentProfit:   0.0, // Will be calculated by behavioral MM
			Timestamp:       current.Timestamp.UnixMilli(),
		}
	}
	
	return nil
}

// detectTripleInverseEvent detects when all three indicators are in inverse zones
func detectTripleInverseEvent(current IndicatorSnapshot, positionCtx *PositionState) *indicators.ZoneEvent {
	// Reconstruct entry snapshot from position state
	entrySnapshot := IndicatorSnapshot{
		STOCHZone: positionCtx.EntrySTOCHZone,
		MFIZone:   positionCtx.EntryMFIZone,
		CCIZone:   positionCtx.EntryCCIZone,
	}
	
	// Check if all three are inverse
	isTripleInverse := DetectTripleInverse(entrySnapshot, current, positionCtx.Direction)
	
	if isTripleInverse {
		return &indicators.ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "TRIPLE_INVERSE_ALIGNMENT",
			IsInverse:       true,
			RequiresProfit:  false, // Triple inverse is critical, no profit requirement
			ProfitThreshold: 0.0,
			CurrentProfit:   0.0, // Will be calculated by behavioral MM
			Timestamp:       current.Timestamp.UnixMilli(),
		}
	}
	
	return nil
}

// DetectZoneEvents is the main entry point for zone detection (Engine integration)
func DetectZoneEvents(results *indicators.IndicatorResults, positionCtx *PositionState) []indicators.ZoneEvent {
	return detectZoneEvents(results, positionCtx)
}

// GetMonitoringState determines current monitoring state based on indicators
func GetMonitoringState(current IndicatorSnapshot, positionCtx *PositionState) MonitoringState {
	if positionCtx == nil || !positionCtx.IsOpen {
		return StateNormal
	}
	
	// Check for triple inverse (highest priority)
	entrySnapshot := IndicatorSnapshot{
		STOCHZone: positionCtx.EntrySTOCHZone,
		MFIZone:   positionCtx.EntryMFIZone,
		CCIZone:   positionCtx.EntryCCIZone,
	}
	
	if DetectTripleInverse(entrySnapshot, current, positionCtx.Direction) {
		return StateTripleInverse
	}
	
	// Check for STOCH inverse (triggers tick-by-tick monitoring)
	if DetectSTOCHInverse(positionCtx.EntrySTOCHZone, current.STOCHZone, positionCtx.Direction) {
		return StateSTOCHInverse
	}
	
	return StateNormal
}

// IsSTOCHInverseActive checks if STOCH inverse monitoring is active
func IsSTOCHInverseActive(current IndicatorSnapshot, positionCtx *PositionState) bool {
	if positionCtx == nil || !positionCtx.IsOpen {
		return false
	}
	
	return DetectSTOCHInverse(positionCtx.EntrySTOCHZone, current.STOCHZone, positionCtx.Direction)
}

// IsMFISupporting checks if MFI is supporting inverse conditions
func IsMFISupporting(current IndicatorSnapshot, positionCtx *PositionState) bool {
	if positionCtx == nil || !positionCtx.IsOpen {
		return false
	}
	
	entryZone := positionCtx.EntryMFIZone
	currentZone := current.MFIZone
	direction := positionCtx.Direction
	
	switch direction {
	case "LONG":
		return entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		return entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	default:
		return false
	}
}

// IsCCISupporting checks if CCI is supporting inverse conditions
func IsCCISupporting(current IndicatorSnapshot, positionCtx *PositionState) bool {
	if positionCtx == nil || !positionCtx.IsOpen {
		return false
	}
	
	entryZone := positionCtx.EntryCCIZone
	currentZone := current.CCIZone
	direction := positionCtx.Direction
	
	switch direction {
	case "LONG":
		return entryZone == "OVERSOLD" && currentZone == "OVERBOUGHT"
	case "SHORT":
		return entryZone == "OVERBOUGHT" && currentZone == "OVERSOLD"
	default:
		return false
	}
}

// CalculateInverseIntensity calculates how many indicators are in inverse zones (0-3)
func CalculateInverseIntensity(current IndicatorSnapshot, positionCtx *PositionState) int {
	if positionCtx == nil || !positionCtx.IsOpen {
		return 0
	}
	
	intensity := 0
	
	if DetectSTOCHInverse(positionCtx.EntrySTOCHZone, current.STOCHZone, positionCtx.Direction) {
		intensity++
	}
	
	if IsMFISupporting(current, positionCtx) {
		intensity++
	}
	
	if IsCCISupporting(current, positionCtx) {
		intensity++
	}
	
	return intensity
}

// GetInverseAdjustmentMultiplier returns adjustment multiplier based on inverse intensity
func GetInverseAdjustmentMultiplier(intensity int, config StrategyConfig) float64 {
	switch intensity {
	case 0:
		return 0.0 // No inverse conditions
	case 1:
		if config.STOCHInverseAdjust > 0 {
			return config.STOCHInverseAdjust // Assume STOCH is the primary
		}
		return 0.2 // Default
	case 2:
		// STOCH + one supporting indicator
		return config.STOCHInverseAdjust + config.MFIInverseAdjust // Combine adjustments
	case 3:
		// Triple inverse - use dedicated multiplier
		return config.TripleInverseAdjust
	default:
		return 0.0
	}
}

// ShouldActivateTickByTickMonitoring determines if tick-by-tick monitoring should be activated
func ShouldActivateTickByTickMonitoring(current IndicatorSnapshot, positionCtx *PositionState) bool {
	if positionCtx == nil || !positionCtx.IsOpen {
		return false
	}
	
	// Activate if STOCH is in inverse zone (primary trigger)
	return DetectSTOCHInverse(positionCtx.EntrySTOCHZone, current.STOCHZone, positionCtx.Direction)
}

// ValidateZoneTransition validates that a zone transition is legitimate
func ValidateZoneTransition(previous, current IndicatorSnapshot) bool {
	// Basic sanity checks for zone transitions
	if previous.Timestamp.After(current.Timestamp) {
		return false // Time should be moving forward
	}
	
	// Allow any valid zone transition - indicators can move freely
	validZones := []string{"OVERSOLD", "NEUTRAL", "OVERBOUGHT"}
	
	isValidZone := func(zone string) bool {
		for _, valid := range validZones {
			if zone == valid {
				return true
			}
		}
		return false
	}
	
	return isValidZone(current.STOCHZone) && isValidZone(current.MFIZone) && isValidZone(current.CCIZone)
}
