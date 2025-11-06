package bingx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// TrailingStopCondition represents conditions for adjusting trailing stops
type TrailingStopCondition struct {
	Indicator     string  // "CCI", "MACD", "DMI"
	Trigger       string  // "inverse_zone", "signal_cross", "counter_trend"
	Action        string  // "tighten", "loosen", "close"
	AdjustmentPct float64 // Percentage adjustment
	MinProfit     float64 // Minimum profit required to trigger
}

// TrailingStopManager manages intelligent trailing stops with conditions
type TrailingStopManager struct {
	client     *Client
	trading    *TradingService
	positions  map[string]*ManagedPosition
	conditions []TrailingStopCondition
	mutex      sync.RWMutex
	monitoring bool
	stopChan   chan bool
}

// ManagedPosition represents a position with trailing stop management
type ManagedPosition struct {
	Position          *Position
	TrailingStopRate  float64
	ActivationPrice   float64
	CurrentStopPrice  float64
	LastUpdateTime    time.Time
	Conditions        []TrailingStopCondition
	IsActive          bool
	ProfitThreshold   float64
}

// NewTrailingStopManager creates a new trailing stop manager
func NewTrailingStopManager(client *Client, trading *TradingService) *TrailingStopManager {
	return &TrailingStopManager{
		client:     client,
		trading:    trading,
		positions:  make(map[string]*ManagedPosition),
		conditions: make([]TrailingStopCondition, 0),
		stopChan:   make(chan bool, 1),
	}
}

// AddTrailingStop adds a trailing stop to a position
func (tsm *TrailingStopManager) AddTrailingStop(ctx context.Context, symbol string, positionSide PositionSide, trailingRate, activationPrice float64) error {
	if trailingRate <= 0 || trailingRate > 1 {
		return fmt.Errorf("invalid trailing rate: %f (must be between 0 and 1)", trailingRate)
	}

	// Get current position
	positions, err := tsm.trading.GetPositions(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get position for %s: %w", symbol, err)
	}

	var targetPosition *Position
	for _, pos := range positions {
		if pos.PositionSide == positionSide && pos.Size != 0 {
			targetPosition = &pos
			break
		}
	}

	if targetPosition == nil {
		return fmt.Errorf("no open position found for %s %s", symbol, positionSide)
	}

	// Place trailing stop order
	params := map[string]string{
		"symbol":          symbol,
		"side":            tsm.getClosingSide(positionSide),
		"positionSide":    string(positionSide),
		"type":            string(OrderTypeTrailingStopMarket),
		"quantity":        strconv.FormatFloat(targetPosition.Size, 'f', -1, 64),
		"callbackRate":    strconv.FormatFloat(trailingRate, 'f', 4, 64),
	}

	if activationPrice > 0 {
		params["activationPrice"] = strconv.FormatFloat(activationPrice, 'f', -1, 64)
	}

	_, err = tsm.client.DoRequest(ctx, http.MethodPost, "/openApi/swap/v2/trade/order", params, EndpointTypeTrading)
	if err != nil {
		return fmt.Errorf("failed to place trailing stop for %s: %w", symbol, err)
	}

	// Add to managed positions
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	positionKey := fmt.Sprintf("%s_%s", symbol, positionSide)
	tsm.positions[positionKey] = &ManagedPosition{
		Position:         targetPosition,
		TrailingStopRate: trailingRate,
		ActivationPrice:  activationPrice,
		LastUpdateTime:   time.Now(),
		IsActive:         true,
		Conditions:       tsm.conditions,
	}

	return nil
}

// AddCondition adds an adjustment condition to the trailing stop manager
func (tsm *TrailingStopManager) AddCondition(condition TrailingStopCondition) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()
	
	tsm.conditions = append(tsm.conditions, condition)
	
	// Update existing positions with new condition
	for _, managedPos := range tsm.positions {
		managedPos.Conditions = append(managedPos.Conditions, condition)
	}
}

// UpdateTrailingStop updates trailing stop based on market conditions and indicators
func (tsm *TrailingStopManager) UpdateTrailingStop(ctx context.Context, symbol string, positionSide PositionSide, newRate float64, reason string) error {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	positionKey := fmt.Sprintf("%s_%s", symbol, positionSide)
	managedPos, exists := tsm.positions[positionKey]
	if !exists {
		return fmt.Errorf("no managed position found for %s %s", symbol, positionSide)
	}

	if !managedPos.IsActive {
		return fmt.Errorf("trailing stop is not active for %s %s", symbol, positionSide)
	}

	// Validate new rate
	if newRate <= 0 || newRate > 1 {
		return fmt.Errorf("invalid new trailing rate: %f", newRate)
	}

	// Cancel existing trailing stop order
	orders, err := tsm.trading.GetOpenOrders(ctx, symbol, false)
	if err != nil {
		return fmt.Errorf("failed to get open orders: %w", err)
	}

	for _, order := range orders {
		if order.Type == OrderTypeTrailingStopMarket {
			if err := tsm.trading.CancelOrder(ctx, symbol, order.OrderID, false); err != nil {
				return fmt.Errorf("failed to cancel existing trailing stop: %w", err)
			}
			break
		}
	}

	// Place new trailing stop with updated rate
	params := map[string]string{
		"symbol":       symbol,
		"side":         tsm.getClosingSide(positionSide),
		"positionSide": string(positionSide),
		"type":         string(OrderTypeTrailingStopMarket),
		"quantity":     strconv.FormatFloat(managedPos.Position.Size, 'f', -1, 64),
		"callbackRate": strconv.FormatFloat(newRate, 'f', 4, 64),
	}

	_, err = tsm.client.DoRequest(ctx, http.MethodPost, "/openApi/swap/v2/trade/order", params, EndpointTypeTrading)
	if err != nil {
		return fmt.Errorf("failed to update trailing stop: %w", err)
	}

	// Update managed position
	managedPos.TrailingStopRate = newRate
	managedPos.LastUpdateTime = time.Now()

	return nil
}

// ProcessIndicatorSignal processes indicator signals for trailing stop adjustments
func (tsm *TrailingStopManager) ProcessIndicatorSignal(ctx context.Context, symbol string, indicator, signal string, value float64) error {
	tsm.mutex.RLock()
	positions := make(map[string]*ManagedPosition)
	for key, pos := range tsm.positions {
		if pos.Position.Symbol == symbol && pos.IsActive {
			positions[key] = pos
		}
	}
	tsm.mutex.RUnlock()

	for _, managedPos := range positions {
		for _, condition := range managedPos.Conditions {
			if condition.Indicator != indicator {
				continue
			}

			shouldTrigger := tsm.evaluateCondition(condition, signal, value, managedPos)
			if !shouldTrigger {
				continue
			}

			// Check minimum profit requirement
			currentProfit := tsm.calculatePositionProfit(managedPos.Position)
			if currentProfit < condition.MinProfit {
				continue
			}

			switch condition.Action {
			case "tighten":
				newRate := managedPos.TrailingStopRate * (1 - condition.AdjustmentPct)
				if newRate < 0.001 {
					newRate = 0.001 // Minimum 0.1%
				}
				err := tsm.UpdateTrailingStop(ctx, symbol, managedPos.Position.PositionSide, newRate, fmt.Sprintf("%s_%s", indicator, signal))
				if err != nil {
					return fmt.Errorf("failed to tighten trailing stop: %w", err)
				}

			case "loosen":
				newRate := managedPos.TrailingStopRate * (1 + condition.AdjustmentPct)
				if newRate > 0.1 {
					newRate = 0.1 // Maximum 10%
				}
				err := tsm.UpdateTrailingStop(ctx, symbol, managedPos.Position.PositionSide, newRate, fmt.Sprintf("%s_%s", indicator, signal))
				if err != nil {
					return fmt.Errorf("failed to loosen trailing stop: %w", err)
				}

			case "close":
				err := tsm.ClosePositionEarly(ctx, symbol, managedPos.Position.PositionSide, fmt.Sprintf("%s_%s_early_exit", indicator, signal))
				if err != nil {
					return fmt.Errorf("failed to close position early: %w", err)
				}
			}
		}
	}

	return nil
}

// ClosePositionEarly closes a position before trailing stop is hit
func (tsm *TrailingStopManager) ClosePositionEarly(ctx context.Context, symbol string, positionSide PositionSide, reason string) error {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	positionKey := fmt.Sprintf("%s_%s", symbol, positionSide)
	managedPos, exists := tsm.positions[positionKey]
	if !exists {
		return fmt.Errorf("no managed position found for %s %s", symbol, positionSide)
	}

	// Cancel trailing stop order
	orders, err := tsm.trading.GetOpenOrders(ctx, symbol, false)
	if err != nil {
		return fmt.Errorf("failed to get open orders: %w", err)
	}

	for _, order := range orders {
		if order.Type == OrderTypeTrailingStopMarket {
			if err := tsm.trading.CancelOrder(ctx, symbol, order.OrderID, false); err != nil {
				return fmt.Errorf("failed to cancel trailing stop: %w", err)
			}
			break
		}
	}

	// Close position with market order
	if positionSide == PositionSideLong {
		_, err = tsm.trading.CloseFuturesLong(ctx, symbol, managedPos.Position.Size, OrderTypeMarket, 0)
	} else {
		_, err = tsm.trading.CloseFuturesShort(ctx, symbol, managedPos.Position.Size, OrderTypeMarket, 0)
	}

	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	// Mark position as inactive
	managedPos.IsActive = false
	managedPos.LastUpdateTime = time.Now()

	return nil
}

// StartMonitoring starts monitoring trailing stops and positions
func (tsm *TrailingStopManager) StartMonitoring(ctx context.Context, checkInterval time.Duration) {
	if tsm.monitoring {
		return
	}

	tsm.monitoring = true
	go tsm.monitoringLoop(ctx, checkInterval)
}

// StopMonitoring stops the monitoring loop
func (tsm *TrailingStopManager) StopMonitoring() {
	if !tsm.monitoring {
		return
	}

	tsm.monitoring = false
	select {
	case tsm.stopChan <- true:
	default:
	}
}

// monitoringLoop continuously monitors positions and trailing stops
func (tsm *TrailingStopManager) monitoringLoop(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tsm.checkPositions(ctx)
		case <-tsm.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkPositions checks all managed positions for status updates
func (tsm *TrailingStopManager) checkPositions(ctx context.Context) {
	tsm.mutex.RLock()
	positionsToCheck := make([]*ManagedPosition, 0, len(tsm.positions))
	for _, pos := range tsm.positions {
		if pos.IsActive {
			positionsToCheck = append(positionsToCheck, pos)
		}
	}
	tsm.mutex.RUnlock()

	for _, managedPos := range positionsToCheck {
		// Check if position still exists
		positions, err := tsm.trading.GetPositions(ctx, managedPos.Position.Symbol)
		if err != nil {
			continue
		}

		stillOpen := false
		for _, pos := range positions {
			if pos.PositionSide == managedPos.Position.PositionSide && pos.Size != 0 {
				// Update position data
				managedPos.Position = &pos
				stillOpen = true
				break
			}
		}

		if !stillOpen {
			// Position was closed (trailing stop triggered or manual close)
			tsm.mutex.Lock()
			positionKey := fmt.Sprintf("%s_%s", managedPos.Position.Symbol, managedPos.Position.PositionSide)
			if pos, exists := tsm.positions[positionKey]; exists {
				pos.IsActive = false
				pos.LastUpdateTime = time.Now()
			}
			tsm.mutex.Unlock()
		}
	}
}

// Helper functions

func (tsm *TrailingStopManager) getClosingSide(positionSide PositionSide) string {
	if positionSide == PositionSideLong {
		return string(OrderSideSell)
	}
	return string(OrderSideBuy)
}

func (tsm *TrailingStopManager) evaluateCondition(condition TrailingStopCondition, signal string, value float64, managedPos *ManagedPosition) bool {
	switch condition.Trigger {
	case "inverse_zone":
		// CCI inverse zone logic
		if condition.Indicator == "CCI" {
			if managedPos.Position.PositionSide == PositionSideLong && value < -100 {
				return true // CCI back to oversold while in long
			}
			if managedPos.Position.PositionSide == PositionSideShort && value > 100 {
				return true // CCI back to overbought while in short
			}
		}
	case "signal_cross":
		// MACD signal cross logic
		if condition.Indicator == "MACD" && signal == "bearish_cross" && managedPos.Position.PositionSide == PositionSideLong {
			return true
		}
		if condition.Indicator == "MACD" && signal == "bullish_cross" && managedPos.Position.PositionSide == PositionSideShort {
			return true
		}
	case "counter_trend":
		// DMI counter trend logic
		if condition.Indicator == "DMI" && signal == "counter_trend" {
			return true
		}
	}
	return false
}

func (tsm *TrailingStopManager) calculatePositionProfit(position *Position) float64 {
	// Calculate unrealized PnL percentage
	if position.EntryPrice == 0 {
		return 0
	}

	priceDiff := position.MarkPrice - position.EntryPrice
	if position.PositionSide == PositionSideShort {
		priceDiff = -priceDiff
	}

	return priceDiff / position.EntryPrice
}

// GetManagedPositions returns all currently managed positions
func (tsm *TrailingStopManager) GetManagedPositions() map[string]*ManagedPosition {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	result := make(map[string]*ManagedPosition)
	for key, pos := range tsm.positions {
		if pos.IsActive {
			result[key] = pos
		}
	}

	return result
}

// RemovePosition removes a position from management
func (tsm *TrailingStopManager) RemovePosition(symbol string, positionSide PositionSide) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	positionKey := fmt.Sprintf("%s_%s", symbol, positionSide)
	delete(tsm.positions, positionKey)
}
