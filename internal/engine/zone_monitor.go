package engine

import (
	"fmt"
	"time"
)

// ZoneMonitor manages active zone monitoring and stop adjustments
type ZoneMonitor struct {
	config      ZoneConfig
	activeZones map[ZoneType]ActiveZone
	
	// Statistics
	totalActivations   int
	totalAdjustments   int
	zoneActivationLog  []ZoneActivation
}

// NewZoneMonitor creates a new zone monitor
func NewZoneMonitor(config ZoneConfig) *ZoneMonitor {
	return &ZoneMonitor{
		config:            config,
		activeZones:       make(map[ZoneType]ActiveZone),
		totalActivations:  0,
		totalAdjustments:  0,
		zoneActivationLog: make([]ZoneActivation, 0),
	}
}

// ActivateZone activates monitoring for a specific zone type
func (zm *ZoneMonitor) ActivateZone(zoneType ZoneType, timestamp int64, metadata ZoneMetadata) error {
	if !zm.isZoneEnabled(zoneType) {
		return fmt.Errorf("zone type %v is not enabled", zoneType)
	}
	
	// Check if zone is already active
	if zone, exists := zm.activeZones[zoneType]; exists && zone.Active {
		return fmt.Errorf("zone %v is already active", zoneType)
	}
	
	// Activate the zone
	zm.activeZones[zoneType] = ActiveZone{
		Type:        zoneType,
		Active:      true,
		EntryTime:   timestamp,
		TriggerTime: timestamp,
	}
	
	zm.totalActivations++
	
	// Log activation
	zm.logZoneActivation(ZoneActivation{
		ZoneType:  zoneType,
		Action:    "ACTIVATED",
		Timestamp: timestamp,
		Metadata:  metadata,
	})
	
	return nil
}

// DeactivateZone deactivates monitoring for a specific zone type
func (zm *ZoneMonitor) DeactivateZone(zoneType ZoneType, timestamp int64, reason string) error {
	zone, exists := zm.activeZones[zoneType]
	if !exists || !zone.Active {
		return fmt.Errorf("zone %v is not active", zoneType)
	}
	
	// Deactivate the zone
	zone.Active = false
	zm.activeZones[zoneType] = zone
	
	// Log deactivation
	zm.logZoneActivation(ZoneActivation{
		ZoneType:  zoneType,
		Action:    "DEACTIVATED",
		Timestamp: timestamp,
		Metadata: ZoneMetadata{
			Reason: reason,
		},
	})
	
	return nil
}

// CheckActiveZones checks all active zones and returns adjustment if needed
func (zm *ZoneMonitor) CheckActiveZones(currentProfit float64, timestamp int64) *StopAdjustment {
	for zoneType, zone := range zm.activeZones {
		if !zone.Active {
			continue
		}
		
		adjustment := zm.checkZoneCondition(zoneType, zone, currentProfit, timestamp)
		if adjustment != nil {
			zm.totalAdjustments++
			return adjustment
		}
	}
	
	return nil
}

// ProcessZoneEvent processes an external zone event (from indicators)
func (zm *ZoneMonitor) ProcessZoneEvent(event ZoneEvent) error {
	switch event.Type {
	case "CCI_ZONE_ENTERED":
		if event.IsInverse {
			return zm.ActivateZone(ZoneCCIInverse, event.Timestamp, ZoneMetadata{
				Value:  event.CurrentProfit,
				Reason: "CCI entered inverse zone",
			})
		}
		
	case "CCI_ZONE_EXITED":
		return zm.DeactivateZone(ZoneCCIInverse, event.Timestamp, "CCI exited zone")
		
	case "MACD_INVERSE_CROSS":
		if event.RequiresProfit && event.CurrentProfit >= event.ProfitThreshold {
			return zm.ActivateZone(ZoneMACDInverse, event.Timestamp, ZoneMetadata{
				Value:           event.CurrentProfit,
				ProfitThreshold: event.ProfitThreshold,
				Reason:          "MACD inverse crossover with profit",
			})
		}
		
	case "DI_COUNTER_CROSS":
		if event.RequiresProfit && event.CurrentProfit >= event.ProfitThreshold {
			return zm.ActivateZone(ZoneDICounter, event.Timestamp, ZoneMetadata{
				Value:           event.CurrentProfit,
				ProfitThreshold: event.ProfitThreshold,
				Reason:          "DI counter-trend crossover with profit",
			})
		}
	}
	
	return nil
}

// ResetAllZones resets all active zones (typically when position closes)
func (zm *ZoneMonitor) ResetAllZones() {
	timestamp := time.Now().UnixMilli()
	
	for zoneType := range zm.activeZones {
		if zm.activeZones[zoneType].Active {
			zm.DeactivateZone(zoneType, timestamp, "position_closed")
		}
	}
}

// GetActiveZones returns copy of all active zones
func (zm *ZoneMonitor) GetActiveZones() map[ZoneType]ActiveZone {
	result := make(map[ZoneType]ActiveZone)
	for k, v := range zm.activeZones {
		if v.Active {
			result[k] = v
		}
	}
	return result
}

// GetZoneStatistics returns zone monitoring statistics
func (zm *ZoneMonitor) GetZoneStatistics() ZoneStatistics {
	return ZoneStatistics{
		TotalActivations: zm.totalActivations,
		TotalAdjustments: zm.totalAdjustments,
		ActiveZoneCount:  zm.countActiveZones(),
		ActivationLog:    zm.getRecentActivations(10),
	}
}

// Private methods

func (zm *ZoneMonitor) isZoneEnabled(zoneType ZoneType) bool {
	switch zoneType {
	case ZoneCCIInverse:
		return zm.config.CCIInverse.Enabled
	case ZoneMACDInverse:
		return zm.config.MACDInverse.Enabled
	case ZoneDICounter:
		return zm.config.DICounter.Enabled
	default:
		return false
	}
}

func (zm *ZoneMonitor) checkZoneCondition(zoneType ZoneType, zone ActiveZone, currentProfit float64, timestamp int64) *StopAdjustment {
	switch zoneType {
	case ZoneCCIInverse:
		return zm.checkCCIInverseZone(zone, currentProfit, timestamp)
	case ZoneMACDInverse:
		return zm.checkMACDInverseZone(zone, currentProfit, timestamp)
	case ZoneDICounter:
		return zm.checkDICounterZone(zone, currentProfit, timestamp)
	default:
		return nil
	}
}

func (zm *ZoneMonitor) checkCCIInverseZone(zone ActiveZone, currentProfit float64, timestamp int64) *StopAdjustment {
	// CCI inverse zone has continuous monitoring
	config := zm.config.CCIInverse
	
	if config.Monitoring != "continuous" {
		return nil
	}
	
	// Apply adjustment based on profit level
	adjustmentPercent := zm.calculateAdjustmentPercent(currentProfit)
	
	// Check if enough time has passed since last adjustment (throttling)
	if zm.shouldThrottleAdjustment(zone, timestamp) {
		return nil
	}
	
	return &StopAdjustment{
		NewStopPercent: adjustmentPercent,
		Reason:         "CCI inverse zone continuous monitoring",
		ZoneType:       ZoneCCIInverse,
		TriggerProfit:  currentProfit,
		Timestamp:      timestamp,
	}
}

func (zm *ZoneMonitor) checkMACDInverseZone(zone ActiveZone, currentProfit float64, timestamp int64) *StopAdjustment {
	// MACD inverse is one-time event
	config := zm.config.MACDInverse
	
	if config.Monitoring != "event" {
		return nil
	}
	
	// Check if profit meets threshold
	if currentProfit < config.ProfitThreshold {
		return nil
	}
	
	adjustmentPercent := zm.calculateAdjustmentPercent(currentProfit)
	
	// Deactivate zone after use (one-time event)
	zm.DeactivateZone(ZoneMACDInverse, timestamp, "adjustment_applied")
	
	return &StopAdjustment{
		NewStopPercent: adjustmentPercent,
		Reason:         "MACD inverse event triggered",
		ZoneType:       ZoneMACDInverse,
		TriggerProfit:  currentProfit,
		Timestamp:      timestamp,
	}
}

func (zm *ZoneMonitor) checkDICounterZone(zone ActiveZone, currentProfit float64, timestamp int64) *StopAdjustment {
	// DI counter is one-time event
	config := zm.config.DICounter
	
	if config.Monitoring != "event" {
		return nil
	}
	
	// Check if profit meets threshold
	if currentProfit < config.ProfitThreshold {
		return nil
	}
	
	adjustmentPercent := zm.calculateAdjustmentPercent(currentProfit)
	
	// Deactivate zone after use (one-time event)
	zm.DeactivateZone(ZoneDICounter, timestamp, "adjustment_applied")
	
	return &StopAdjustment{
		NewStopPercent: adjustmentPercent,
		Reason:         "DI counter-trend event triggered",
		ZoneType:       ZoneDICounter,
		TriggerProfit:  currentProfit,
		Timestamp:      timestamp,
	}
}

func (zm *ZoneMonitor) calculateAdjustmentPercent(currentProfit float64) float64 {
	// This should use the adjustment grid from PositionManager
	// For now, implement simple logic
	
	if currentProfit >= 20.0 {
		return 0.5 // Very tight stop for high profits
	} else if currentProfit >= 10.0 {
		return 1.0
	} else if currentProfit >= 5.0 {
		return 1.5
	} else {
		return 2.0 // Default trailing stop
	}
}

func (zm *ZoneMonitor) shouldThrottleAdjustment(zone ActiveZone, timestamp int64) bool {
	// Prevent too frequent adjustments (minimum 1 minute between adjustments)
	minInterval := int64(60 * 1000) // 1 minute in milliseconds
	
	return zone.LastAdjustment > 0 && (timestamp-zone.LastAdjustment) < minInterval
}

func (zm *ZoneMonitor) countActiveZones() int {
	count := 0
	for _, zone := range zm.activeZones {
		if zone.Active {
			count++
		}
	}
	return count
}

func (zm *ZoneMonitor) logZoneActivation(activation ZoneActivation) {
	zm.zoneActivationLog = append(zm.zoneActivationLog, activation)
	
	// Keep only last 100 activations to prevent memory growth
	if len(zm.zoneActivationLog) > 100 {
		zm.zoneActivationLog = zm.zoneActivationLog[1:]
	}
}

func (zm *ZoneMonitor) getRecentActivations(count int) []ZoneActivation {
	if len(zm.zoneActivationLog) == 0 {
		return []ZoneActivation{}
	}
	
	start := len(zm.zoneActivationLog) - count
	if start < 0 {
		start = 0
	}
	
	result := make([]ZoneActivation, len(zm.zoneActivationLog[start:]))
	copy(result, zm.zoneActivationLog[start:])
	return result
}

// Supporting types

// ZoneMetadata holds metadata about zone events
type ZoneMetadata struct {
	Value           float64 `json:"value,omitempty"`
	ProfitThreshold float64 `json:"profit_threshold,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// ZoneActivation logs zone activation/deactivation events
type ZoneActivation struct {
	ZoneType  ZoneType     `json:"zone_type"`
	Action    string       `json:"action"` // "ACTIVATED" or "DEACTIVATED"
	Timestamp int64        `json:"timestamp"`
	Metadata  ZoneMetadata `json:"metadata"`
}

// ZoneStatistics holds statistics about zone monitoring
type ZoneStatistics struct {
	TotalActivations int              `json:"total_activations"`
	TotalAdjustments int              `json:"total_adjustments"`
	ActiveZoneCount  int              `json:"active_zone_count"`
	ActivationLog    []ZoneActivation `json:"activation_log"`
}

// ValidateZoneConfig validates zone configuration
func ValidateZoneConfig(config ZoneConfig) error {
	if config.MACDInverse.Enabled && config.MACDInverse.ProfitThreshold < 0 {
		return fmt.Errorf("MACD inverse profit threshold must be non-negative")
	}
	
	if config.DICounter.Enabled && config.DICounter.ProfitThreshold < 0 {
		return fmt.Errorf("DI counter profit threshold must be non-negative")
	}
	
	// Validate monitoring types
	validMonitoring := map[string]bool{
		"continuous": true,
		"event":      true,
	}
	
	if config.CCIInverse.Enabled && !validMonitoring[config.CCIInverse.Monitoring] {
		return fmt.Errorf("invalid CCI inverse monitoring type: %s", config.CCIInverse.Monitoring)
	}
	
	if config.MACDInverse.Enabled && !validMonitoring[config.MACDInverse.Monitoring] {
		return fmt.Errorf("invalid MACD inverse monitoring type: %s", config.MACDInverse.Monitoring)
	}
	
	if config.DICounter.Enabled && !validMonitoring[config.DICounter.Monitoring] {
		return fmt.Errorf("invalid DI counter monitoring type: %s", config.DICounter.Monitoring)
	}
	
	return nil
}
