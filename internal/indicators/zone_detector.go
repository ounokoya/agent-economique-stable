package indicators

// detectZoneEvents detects zone monitoring events using Engine-compatible types
func detectZoneEvents(results *IndicatorResults, positionCtx *PositionContext) []ZoneEvent {
	var events []ZoneEvent
	
	if positionCtx == nil || !positionCtx.IsOpen {
		return events
	}
	
	// CCI Zone Inverse Detection (from user memory strategy)
	// Monitoring CCI continu en zone opposée
	if cciEvent := detectCCIInverseEvent(results, positionCtx); cciEvent != nil {
		events = append(events, *cciEvent)
	}
	
	// MACD Inverse Detection (if profit conditions met)
	// Ajustements trailing stop: MACD inverse si profit
	if macdEvent := detectMACDInverseEvent(results, positionCtx); macdEvent != nil {
		events = append(events, *macdEvent)
	}
	
	// DI Counter Detection (if profit conditions met) 
	// DI contre-tendance si profit
	if diEvent := detectDICounterEvent(results, positionCtx); diEvent != nil {
		events = append(events, *diEvent)
	}
	
	return events
}

// detectCCIInverseEvent detects CCI moving to inverse zone
func detectCCIInverseEvent(results *IndicatorResults, positionCtx *PositionContext) *ZoneEvent {
	cciValue := results.CCI.Value
	entryCCIZone := positionCtx.EntryCCIZone
	
	// Check for CCI zone inverse based on entry zone
	var isInverse bool
	
	if positionCtx.Direction == "LONG" {
		// LONG entered in oversold, check if now overbought (inverse)
		if entryCCIZone == CCIOversold && cciValue > 100 {
			isInverse = true
		}
	} else if positionCtx.Direction == "SHORT" {
		// SHORT entered in overbought, check if now oversold (inverse)
		if entryCCIZone == CCIOverbought && cciValue < -100 {
			isInverse = true
		}
	}
	
	if isInverse {
		return &ZoneEvent{
			Type:        "ZONE_ACTIVATED",
			ZoneType:    "CCI_INVERSE",
			IsInverse:   true,
			Timestamp:   results.Timestamp,
		}
	}
	
	return nil
}

// detectMACDInverseEvent detects MACD crossover inverse to position direction
func detectMACDInverseEvent(results *IndicatorResults, positionCtx *PositionContext) *ZoneEvent {
	// Only trigger if position is profitable (from user memory strategy)
	if positionCtx.ProfitPercent <= 0.5 {
		return nil
	}
	
	macdCross := results.MACD.CrossoverType
	var isInverse bool
	
	if positionCtx.Direction == "LONG" && macdCross == CrossDown {
		// LONG position but MACD crosses down (inverse)
		isInverse = true
	} else if positionCtx.Direction == "SHORT" && macdCross == CrossUp {
		// SHORT position but MACD crosses up (inverse) 
		isInverse = true
	}
	
	if isInverse {
		return &ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "MACD_INVERSE", 
			RequiresProfit:  true,
			ProfitThreshold: 0.5,
			CurrentProfit:   positionCtx.ProfitPercent,
			Timestamp:       results.Timestamp,
		}
	}
	
	return nil
}

// detectDICounterEvent detects DI movement counter to position direction
func detectDICounterEvent(results *IndicatorResults, positionCtx *PositionContext) *ZoneEvent {
	// Only trigger if position is profitable (from user memory strategy)
	if positionCtx.ProfitPercent <= 1.0 {
		return nil
	}
	
	diPlus := results.DMI.PlusDI
	diMinus := results.DMI.MinusDI
	var isCounter bool
	
	if positionCtx.Direction == "LONG" && diMinus > diPlus {
		// LONG position but DI- > DI+ (counter trend)
		isCounter = true
	} else if positionCtx.Direction == "SHORT" && diPlus > diMinus {
		// SHORT position but DI+ > DI- (counter trend)
		isCounter = true
	}
	
	if isCounter {
		return &ZoneEvent{
			Type:            "ZONE_ACTIVATED",
			ZoneType:        "DI_COUNTER",
			RequiresProfit:  true, 
			ProfitThreshold: 1.0,
			CurrentProfit:   positionCtx.ProfitPercent,
			Timestamp:       results.Timestamp,
		}
	}
	
	return nil
}
