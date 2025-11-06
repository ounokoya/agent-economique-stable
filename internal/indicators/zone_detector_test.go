package indicators

import (
	"testing"
	"time"
)

func TestDetectZoneEvents_CCIInverse_Long(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: 120}, // Now overbought (inverse for LONG)
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:       true,
		Direction:    "LONG",
		EntryCCIZone: CCIOversold, // Entered in oversold
	}
	
	events := detectZoneEvents(results, positionCtx)
	
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.Type != "ZONE_ACTIVATED" {
		t.Errorf("Expected ZONE_ACTIVATED, got %s", event.Type)
	}
	
	if event.ZoneType != "CCI_INVERSE" {
		t.Errorf("Expected CCI_INVERSE, got %s", event.ZoneType)
	}
	
	if !event.IsInverse {
		t.Error("Expected IsInverse=true")
	}
}

func TestDetectZoneEvents_CCIInverse_Short(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: -120}, // Now oversold (inverse for SHORT)
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:       true,
		Direction:    "SHORT",
		EntryCCIZone: CCIOverbought, // Entered in overbought
	}
	
	events := detectZoneEvents(results, positionCtx)
	
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.ZoneType != "CCI_INVERSE" {
		t.Errorf("Expected CCI_INVERSE, got %s", event.ZoneType)
	}
}

func TestDetectZoneEvents_MACDInverse_WithProfit(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{CrossoverType: CrossDown}, // Inverse for LONG
		CCI:  &CCIValues{Value: 0}, // Add to prevent nil pointer
		DMI:  &DMIValues{PlusDI: 20, MinusDI: 20}, // Add to prevent nil pointer
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:        true,
		Direction:     "LONG",
		ProfitPercent: 1.0, // Above threshold (0.5)
	}
	
	events := detectZoneEvents(results, positionCtx)
	
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.ZoneType != "MACD_INVERSE" {
		t.Errorf("Expected MACD_INVERSE, got %s", event.ZoneType)
	}
	
	if !event.RequiresProfit {
		t.Error("Expected RequiresProfit=true")
	}
	
	if event.ProfitThreshold != 0.5 {
		t.Errorf("Expected ProfitThreshold=0.5, got %.1f", event.ProfitThreshold)
	}
}

func TestDetectZoneEvents_MACDInverse_InsufficientProfit(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{CrossoverType: CrossDown},
		CCI:  &CCIValues{Value: 0}, // Add to prevent nil pointer
		DMI:  &DMIValues{PlusDI: 20, MinusDI: 20}, // Add to prevent nil pointer
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:        true,
		Direction:     "LONG",
		ProfitPercent: 0.3, // Below threshold (0.5)
	}
	
	events := detectZoneEvents(results, positionCtx)
	
	if len(events) != 0 {
		t.Errorf("Expected 0 events for insufficient profit, got %d", len(events))
	}
}

func TestDetectZoneEvents_DICounter_WithProfit(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{CrossoverType: NoCrossover}, // Add to prevent nil pointer
		CCI:  &CCIValues{Value: 0}, // Add to prevent nil pointer
		DMI: &DMIValues{
			PlusDI:  15,
			MinusDI: 25, // DI- > DI+ (counter for LONG)
		},
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:        true,
		Direction:     "LONG",
		ProfitPercent: 1.5, // Above threshold (1.0)
	}
	
	events := detectZoneEvents(results, positionCtx)
	
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.ZoneType != "DI_COUNTER" {
		t.Errorf("Expected DI_COUNTER, got %s", event.ZoneType)
	}
	
	if event.ProfitThreshold != 1.0 {
		t.Errorf("Expected ProfitThreshold=1.0, got %.1f", event.ProfitThreshold)
	}
}

func TestDetectZoneEvents_NoPosition(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: 120},
		Timestamp: time.Now().UnixMilli(),
	}
	
	// No position context
	events := detectZoneEvents(results, nil)
	
	if len(events) != 0 {
		t.Errorf("Expected 0 events for no position, got %d", len(events))
	}
	
	// Closed position
	positionCtx := &PositionContext{
		IsOpen: false,
	}
	
	events = detectZoneEvents(results, positionCtx)
	
	if len(events) != 0 {
		t.Errorf("Expected 0 events for closed position, got %d", len(events))
	}
}

func TestDetectCCIInverseEvent_NoInverse(t *testing.T) {
	results := &IndicatorResults{
		CCI: &CCIValues{Value: -50}, // Still normal range
		Timestamp: time.Now().UnixMilli(),
	}
	
	positionCtx := &PositionContext{
		IsOpen:       true,
		Direction:    "LONG",
		EntryCCIZone: CCIOversold,
	}
	
	event := detectCCIInverseEvent(results, positionCtx)
	
	if event != nil {
		t.Error("Expected no event for non-inverse CCI")
	}
}

func TestDetectMACDInverseEvent_NoInverse(t *testing.T) {
	results := &IndicatorResults{
		MACD: &MACDValues{CrossoverType: CrossUp}, // Same direction as LONG
	}
	
	positionCtx := &PositionContext{
		IsOpen:        true,
		Direction:     "LONG",
		ProfitPercent: 1.0,
	}
	
	event := detectMACDInverseEvent(results, positionCtx)
	
	if event != nil {
		t.Error("Expected no event for non-inverse MACD")
	}
}

func TestDetectDICounterEvent_NoCounter(t *testing.T) {
	results := &IndicatorResults{
		DMI: &DMIValues{
			PlusDI:  25,
			MinusDI: 15, // DI+ > DI- (same as LONG direction)
		},
	}
	
	positionCtx := &PositionContext{
		IsOpen:        true,
		Direction:     "LONG",
		ProfitPercent: 1.5,
	}
	
	event := detectDICounterEvent(results, positionCtx)
	
	if event != nil {
		t.Error("Expected no event for non-counter DI")
	}
}
