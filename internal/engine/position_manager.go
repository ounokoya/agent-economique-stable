package engine

import (
	"fmt"
	"math"
)

// PositionManager manages trading positions and trailing stops
type PositionManager struct {
	position       Position
	trailingConfig TrailingStopConfig
	adjustmentGrid []AdjustmentLevel
	
	// Statistics
	totalPositions int
	totalProfit    float64
}

// NewPositionManager creates a new position manager
func NewPositionManager(trailingConfig TrailingStopConfig, adjustmentGrid []AdjustmentLevel) *PositionManager {
	return &PositionManager{
		position: Position{
			IsOpen:    false,
			Direction: PositionNone,
		},
		trailingConfig: trailingConfig,
		adjustmentGrid: adjustmentGrid,
		totalPositions: 0,
		totalProfit:    0.0,
	}
}

// OpenPosition opens a new trading position
func (pm *PositionManager) OpenPosition(direction PositionDirection, entryPrice float64, timestamp int64, entryCCIZone string) error {
	if pm.position.IsOpen {
		return ErrPositionAlreadyOpen
	}
	
	if entryPrice <= 0 {
		return fmt.Errorf("invalid entry price: %f", entryPrice)
	}
	
	initialStop := pm.calculateInitialStop(direction, entryPrice)
	
	pm.position = Position{
		IsOpen:       true,
		Direction:    direction,
		EntryPrice:   entryPrice,
		EntryTime:    timestamp,
		StopLoss:     initialStop,
		EntryCCIZone: entryCCIZone,
	}
	
	pm.totalPositions++
	
	return nil
}

// ClosePosition closes the current position
func (pm *PositionManager) ClosePosition(timestamp int64, exitPrice float64, reason string) error {
	if !pm.position.IsOpen {
		return ErrNoOpenPosition
	}
	
	if exitPrice <= 0 {
		return fmt.Errorf("invalid exit price: %f", exitPrice)
	}
	
	// Calculate final profit
	profit := pm.position.CalculateProfitPercent(exitPrice)
	pm.totalProfit += profit
	
	// Reset position
	pm.position = Position{
		IsOpen:    false,
		Direction: PositionNone,
	}
	
	return nil
}

// UpdateTrailingStop updates the trailing stop based on current price
func (pm *PositionManager) UpdateTrailingStop(currentPrice float64) error {
	if !pm.position.IsOpen {
		return ErrNoOpenPosition
	}
	
	if currentPrice <= 0 {
		return fmt.Errorf("invalid current price: %f", currentPrice)
	}
	
	// Determine trailing stop percentage based on direction
	var stopPercent float64
	
	// For now, use trend percent as default
	// In real implementation, this would be determined by signal type
	stopPercent = pm.trailingConfig.TrendPercent
	
	// Calculate new stop level
	newStop := pm.calculateTrailingStop(currentPrice, stopPercent)
	
	// Only update if new stop is tighter (better)
	if pm.isStopTighter(newStop, pm.position.StopLoss) {
		pm.position.StopLoss = newStop
	}
	
	return nil
}

// ApplyStopAdjustment applies a stop adjustment based on profit grid
func (pm *PositionManager) ApplyStopAdjustment(adjustmentPercent float64) error {
	if !pm.position.IsOpen {
		return ErrNoOpenPosition
	}
	
	if adjustmentPercent <= 0 || adjustmentPercent > 100 {
		return fmt.Errorf("invalid adjustment percent: %f", adjustmentPercent)
	}
	
	// Get current price (for calculation, we need this to be passed or stored)
	// For now, calculate based on current stop and direction
	currentPrice := pm.estimateCurrentPrice()
	
	// Calculate new stop with adjustment
	newStop := pm.calculateTrailingStop(currentPrice, adjustmentPercent)
	
	// Only apply if it tightens the stop
	if pm.isStopTighter(newStop, pm.position.StopLoss) {
		pm.position.StopLoss = newStop
	}
	
	return nil
}

// IsOpen returns true if there is an open position
func (pm *PositionManager) IsOpen() bool {
	return pm.position.IsOpen
}

// GetPosition returns the current position
func (pm *PositionManager) GetPosition() Position {
	return pm.position
}

// GetStatistics returns position manager statistics
func (pm *PositionManager) GetStatistics() PositionStatistics {
	return PositionStatistics{
		TotalPositions: pm.totalPositions,
		TotalProfit:    pm.totalProfit,
		AverageProfit:  pm.calculateAverageProfit(),
	}
}

// GetAdjustmentForProfit returns the appropriate trailing percentage for given profit
func (pm *PositionManager) GetAdjustmentForProfit(profitPercent float64) float64 {
	for _, level := range pm.adjustmentGrid {
		if profitPercent >= level.ProfitMin && profitPercent < level.ProfitMax {
			return level.TrailingPercent
		}
	}
	
	// Default to last level if profit exceeds all ranges
	if len(pm.adjustmentGrid) > 0 {
		return pm.adjustmentGrid[len(pm.adjustmentGrid)-1].TrailingPercent
	}
	
	return pm.trailingConfig.TrendPercent
}

// Private methods

func (pm *PositionManager) calculateInitialStop(direction PositionDirection, entryPrice float64) float64 {
	stopPercent := pm.trailingConfig.TrendPercent / 100.0
	
	switch direction {
	case PositionLong:
		return entryPrice * (1.0 - stopPercent)
	case PositionShort:
		return entryPrice * (1.0 + stopPercent)
	default:
		return entryPrice
	}
}

func (pm *PositionManager) calculateTrailingStop(currentPrice float64, stopPercent float64) float64 {
	stopRatio := stopPercent / 100.0
	
	switch pm.position.Direction {
	case PositionLong:
		return currentPrice * (1.0 - stopRatio)
	case PositionShort:
		return currentPrice * (1.0 + stopRatio)
	default:
		return currentPrice
	}
}

func (pm *PositionManager) isStopTighter(newStop, currentStop float64) bool {
	switch pm.position.Direction {
	case PositionLong:
		// For long positions, higher stop is tighter
		return newStop > currentStop
	case PositionShort:
		// For short positions, lower stop is tighter
		return newStop < currentStop
	default:
		return false
	}
}

func (pm *PositionManager) estimateCurrentPrice() float64 {
	// This is a placeholder - in real implementation, current price should be passed
	// For now, estimate based on stop and direction
	stopPercent := pm.trailingConfig.TrendPercent / 100.0
	
	switch pm.position.Direction {
	case PositionLong:
		return pm.position.StopLoss / (1.0 - stopPercent)
	case PositionShort:
		return pm.position.StopLoss / (1.0 + stopPercent)
	default:
		return pm.position.EntryPrice
	}
}

func (pm *PositionManager) calculateAverageProfit() float64 {
	if pm.totalPositions == 0 {
		return 0.0
	}
	return pm.totalProfit / float64(pm.totalPositions)
}

// PositionStatistics holds statistics about position management
type PositionStatistics struct {
	TotalPositions int     `json:"total_positions"`
	TotalProfit    float64 `json:"total_profit"`
	AverageProfit  float64 `json:"average_profit"`
}

// StopAdjustment represents a stop adjustment recommendation
type StopAdjustment struct {
	NewStopPercent   float64   `json:"new_stop_percent"`
	Reason          string    `json:"reason"`
	ZoneType        ZoneType  `json:"zone_type"`
	TriggerProfit   float64   `json:"trigger_profit"`
	Timestamp       int64     `json:"timestamp"`
}

// ValidateStopLevel validates if a stop level is reasonable
func ValidateStopLevel(direction PositionDirection, entryPrice, stopPrice float64) error {
	if entryPrice <= 0 || stopPrice <= 0 {
		return fmt.Errorf("prices must be positive")
	}
	
	stopPercent := math.Abs((stopPrice - entryPrice) / entryPrice * 100.0)
	
	// Reasonable stop loss should be between 0.1% and 50%
	if stopPercent < 0.1 {
		return fmt.Errorf("stop too tight: %f%%", stopPercent)
	}
	
	if stopPercent > 50.0 {
		return fmt.Errorf("stop too wide: %f%%", stopPercent)
	}
	
	// Check direction makes sense
	switch direction {
	case PositionLong:
		if stopPrice >= entryPrice {
			return fmt.Errorf("long stop must be below entry price")
		}
	case PositionShort:
		if stopPrice <= entryPrice {
			return fmt.Errorf("short stop must be above entry price")
		}
	}
	
	return nil
}

// CalculateRisk calculates position risk in percentage
func CalculateRisk(direction PositionDirection, entryPrice, stopPrice float64) float64 {
	if entryPrice <= 0 || stopPrice <= 0 {
		return 0.0
	}
	
	return math.Abs((stopPrice - entryPrice) / entryPrice * 100.0)
}

// CalculateRewardRatio calculates reward-to-risk ratio for a target
func CalculateRewardRatio(direction PositionDirection, entryPrice, stopPrice, targetPrice float64) float64 {
	risk := CalculateRisk(direction, entryPrice, stopPrice)
	if risk == 0 {
		return 0.0
	}
	
	var reward float64
	switch direction {
	case PositionLong:
		reward = (targetPrice - entryPrice) / entryPrice * 100.0
	case PositionShort:
		reward = (entryPrice - targetPrice) / entryPrice * 100.0
	default:
		return 0.0
	}
	
	if reward <= 0 {
		return 0.0
	}
	
	return reward / risk
}
