package engine

import (
	"fmt"
	"time"
	
	"agent-economique/internal/money_management"
)

// executeEmergencyStopAll closes all positions immediately (circuit breaker callback)
func (e *TemporalEngine) executeEmergencyStopAll() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	fmt.Printf("[EMERGENCY] 🚨 Circuit breaker triggered - closing all positions\n")
	
	// Get current position if open
	if e.positionManager.IsOpen() {
		position := e.positionManager.GetPosition()
		
		// Force close position
		err := e.positionManager.ClosePosition(e.currentTimestamp, 0, "EMERGENCY_STOP")
		if err != nil {
			fmt.Printf("[EMERGENCY] ❌ Failed to close position: %v\n", err)
			return fmt.Errorf("emergency position closure failed: %w", err)
		}
		
		// Record the emergency closure as a trade
		trade := money_management.TradeRecord{
			Timestamp:    time.Unix(e.currentTimestamp, 0).UTC(),
			StrategyName: "STOCH_MFI_CCI", // Updated for new strategy
			Symbol:       "BTC-USDT", // TODO: Get from config
			Side:         position.Direction.String(),
			EntryPrice:   position.EntryPrice,
			ExitPrice:    position.EntryPrice, // Use entry price for emergency
			Quantity:     1.0, // TODO: Get from position
			PnL:          0.0, // Emergency close at entry price
			PnLPercent:   0.0,
			Duration:     e.currentTimestamp - position.EntryTime,
			IsWinning:    false, // Emergency stops are losses
		}
		
		// Notify money manager of position closure
		e.moneyManager.OnPositionClosed(trade)
		
		fmt.Printf("[EMERGENCY] ✅ Position closed: %.2f USDT PnL\n", 0.0)
	} else {
		fmt.Printf("[EMERGENCY] ℹ️  No positions to close\n")
	}
	
	fmt.Printf("[EMERGENCY] 🛑 Trading halted by circuit breaker\n")
	return nil
}

// validatePositionWithMM validates new position with Money Manager before opening
func (e *TemporalEngine) validatePositionWithMM(direction PositionDirection, entryPrice float64) error {
	if e.moneyManager == nil {
		return fmt.Errorf("money manager not initialized")
	}
	
	// Check if trading is halted
	if !e.moneyManager.IsActive() {
		return fmt.Errorf("trading halted by money management")
	}
	
	// Create position sizing request
	request := money_management.PositionSizingRequest{
		Symbol:           "BTC-USDT", // TODO: Get from config
		Price:            entryPrice,
		PositionType:     "futures", // TODO: Get from config
		AvailableBalance: 1000.0,    // TODO: Get from account balance
		Leverage:         10,        // TODO: Get from config
	}
	
	// Log validation with direction info
	fmt.Printf("[MM] 🔍 Validating %s position at %.2f\n", direction.String(), entryPrice)
	
	// Validate position size
	result, err := e.moneyManager.CalculatePositionSize(request)
	if err != nil {
		return fmt.Errorf("position sizing failed: %w", err)
	}
	
	if !result.IsValid {
		return fmt.Errorf("position validation failed: %s", result.ValidationError)
	}
	
	fmt.Printf("[MM] ✅ Position validated: %.6f quantity, %.2f USDT notional\n", 
		result.Quantity, result.NotionalValue)
	
	return nil
}

// updateMMWithRealTimePnL updates Money Manager with current floating PnL
func (e *TemporalEngine) updateMMWithRealTimePnL() error {
	if e.moneyManager == nil {
		return nil // MM not initialized
	}
	
	// Calculate total PnL from position manager
	var totalPnL float64
	if e.positionManager.IsOpen() {
		position := e.positionManager.GetPosition()
		// Calculate current PnL (need to pass current price)
		// For now, use a simple calculation
		if position.Direction == PositionLong {
			totalPnL = (position.EntryPrice * 0.02) // Mock PnL calculation
		} else {
			totalPnL = -(position.EntryPrice * 0.02) // Mock PnL calculation
		}
	}
	
	// Update Money Manager with real-time PnL
	err := e.moneyManager.UpdateRealTimePnL(totalPnL)
	if err != nil {
		// Circuit breaker triggered - trading will be halted
		fmt.Printf("[MM] 🚨 Circuit breaker triggered: %v\n", err)
		return err
	}
	
	return nil
}

// notifyPositionOpened notifies MM when position is successfully opened
func (e *TemporalEngine) notifyPositionOpened(positionValue float64) {
	if e.moneyManager == nil {
		return
	}
	
	e.moneyManager.OnPositionOpened(positionValue, "STOCH_MFI_CCI")
}

// notifyPositionClosed notifies MM when position is closed
func (e *TemporalEngine) notifyPositionClosed(position Position, exitPrice float64, exitReason string) {
	if e.moneyManager == nil {
		return
	}
	
	// Create trade record
	trade := money_management.TradeRecord{
		Timestamp:    time.Unix(e.currentTimestamp, 0).UTC(),
		StrategyName: "STOCH_MFI_CCI", // Updated for new strategy
		Symbol:       "BTC-USDT", // TODO: Get from config
		Side:         position.Direction.String(),
		EntryPrice:   position.EntryPrice,
		ExitPrice:    exitPrice,
		Quantity:     1.0, // TODO: Get actual quantity
		PnL:          (exitPrice - position.EntryPrice) * 1.0, // Simple PnL calculation
		PnLPercent:   ((exitPrice - position.EntryPrice) / position.EntryPrice) * 100,
		Duration:     e.currentTimestamp - position.EntryTime,
		IsWinning:    exitPrice > position.EntryPrice,
	}
	
	// Notify money manager
	err := e.moneyManager.OnPositionClosed(trade)
	if err != nil {
		fmt.Printf("[MM] ⚠️  Error recording trade: %v\n", err)
	}
	
	fmt.Printf("[MM] 📊 Trade recorded: %.2f USDT PnL, Strategy: %s\n", 
		trade.PnL, trade.StrategyName)
}

// getMMStatus returns current Money Manager status for debugging
func (e *TemporalEngine) getMMStatus() money_management.MoneyManagerStatus {
	if e.moneyManager == nil {
		return money_management.MoneyManagerStatus{}
	}
	
	return e.moneyManager.GetStatus()
}

// printMMStatusSummary prints MM status summary for monitoring
func (e *TemporalEngine) printMMStatusSummary() {
	if e.moneyManager == nil {
		return
	}
	
	status := e.moneyManager.GetStatus()
	metrics := status.GlobalMetrics
	
	fmt.Printf("[MM Status] Active: %t | Halted: %t | Total PnL: %.2f | Win Rate: %.1f%% | Profit Factor: %.2f\n",
		status.IsActive, status.IsTradingHalted, metrics.TotalPnL, metrics.WinRate, metrics.ProfitFactor)
	
	if status.RiskAnalysis.RiskLevel != "LOW" {
		fmt.Printf("[MM Risk] ⚠️  Risk Level: %s | Distance to Daily Limit: %.1f%%\n",
			status.RiskAnalysis.RiskLevel, status.RiskAnalysis.DistanceToDailyLimit)
	}
}

// shutdownMoneyManager gracefully shuts down MM components
func (e *TemporalEngine) shutdownMoneyManager() error {
	if e.moneyManager == nil {
		return nil
	}
	
	fmt.Printf("[MM] 🛑 Shutting down Money Manager...\n")
	
	// Generate final daily report
	report := e.moneyManager.GenerateDailyReport()
	fmt.Printf("[MM] 📊 Final Report - Total PnL: %.2f USDT | Trades: %d | Win Rate: %.1f%%\n",
		report.GlobalMetrics.TotalPnL, report.GlobalMetrics.TotalPositions, report.GlobalMetrics.WinRate)
	
	// Stop MM
	return e.moneyManager.Stop()
}
