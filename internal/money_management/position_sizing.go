package money_management

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// PositionSizer handles position sizing calculations with flexible configuration
type PositionSizer struct {
	config BaseConfiguration
	mutex  sync.RWMutex
	
	// Symbol precision cache
	symbolPrecision map[string]int
	
	// Current tracking
	currentCapital  float64
	dailyPositions  int
	currentTrades   int
	
	// Logging
	auditLogger *AuditLogger
}

// NewPositionSizer creates a new position sizer
func NewPositionSizer(config BaseConfiguration, auditLogger *AuditLogger) (*PositionSizer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	ps := &PositionSizer{
		config:          config,
		currentCapital:  config.CurrentCapital,
		symbolPrecision: make(map[string]int),
		auditLogger:     auditLogger,
	}
	
	// Initialize default precisions
	ps.initializeDefaultPrecisions()
	
	return ps, nil
}

// CalculatePositionSize calculates position size based on configuration mode
func (ps *PositionSizer) CalculatePositionSize(request PositionSizingRequest) PositionSizingResult {
	ps.mutex.RLock()
	config := ps.config
	ps.mutex.RUnlock()
	
	// Validate request
	if err := ps.validateRequest(request); err != nil {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: err.Error(),
		}
	}
	
	// Check daily position limits
	if ps.dailyPositions >= config.MaxDailyPositions {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("daily position limit exceeded (%d/%d)", ps.dailyPositions, config.MaxDailyPositions),
		}
	}
	
	// Check concurrent trade limits
	if ps.currentTrades >= config.MaxConcurrentTrades {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("concurrent trade limit exceeded (%d/%d)", ps.currentTrades, config.MaxConcurrentTrades),
		}
	}
	
	var result PositionSizingResult
	
	// Calculate based on mode
	switch config.PositionSizing.Mode {
	case PositionSizingFixed:
		result = ps.calculateFixed(request)
	case PositionSizingPercentage:
		result = ps.calculatePercentage(request)
	default:
		result = PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("unknown position sizing mode: %s", config.PositionSizing.Mode),
		}
	}
	
	// Apply limits and precision
	if result.IsValid {
		result = ps.applyLimitsAndPrecision(result, request)
	}
	
	// Log the calculation
	ps.auditLogger.LogPositionSizing(request, result, result.IsValid, 
		fmt.Sprintf("Position sizing calculation: mode=%s, valid=%t", config.PositionSizing.Mode, result.IsValid))
	
	return result
}

// calculateFixed calculates position size using fixed USDT amounts
func (ps *PositionSizer) calculateFixed(request PositionSizingRequest) PositionSizingResult {
	config := ps.config.PositionSizing
	
	var amountUSDT float64
	switch strings.ToLower(request.PositionType) {
	case "spot":
		amountUSDT = config.SpotAmountUSDT
	case "futures":
		amountUSDT = config.FuturesAmountUSDT
	default:
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("unknown position type: %s", request.PositionType),
		}
	}
	
	// Check if sufficient balance
	requiredBalance := amountUSDT
	if strings.ToLower(request.PositionType) == "futures" {
		requiredBalance = amountUSDT / float64(request.Leverage) // Margin required
	}
	
	if request.AvailableBalance < requiredBalance {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("insufficient balance: need %.2f USDT, have %.2f USDT", requiredBalance, request.AvailableBalance),
		}
	}
	
	// Calculate quantity
	var quantity float64
	if strings.ToLower(request.PositionType) == "futures" {
		// For futures: Quantity = (Amount * Leverage) / Price
		quantity = (amountUSDT * float64(request.Leverage)) / request.Price
	} else {
		// For spot: Quantity = Amount / Price
		quantity = amountUSDT / request.Price
	}
	
	return PositionSizingResult{
		Quantity:        quantity,
		NotionalValue:   amountUSDT * float64(request.Leverage), // Total position value
		RequiredBalance: requiredBalance,
		IsValid:         true,
	}
}

// calculatePercentage calculates position size using percentage of capital
func (ps *PositionSizer) calculatePercentage(request PositionSizingRequest) PositionSizingResult {
	config := ps.config.PositionSizing
	
	var percentage float64
	switch strings.ToLower(request.PositionType) {
	case "spot":
		percentage = config.SpotPercentage
	case "futures":
		percentage = config.FuturesPercentage
	default:
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("unknown position type: %s", request.PositionType),
		}
	}
	
	// Calculate amount based on current capital
	amountUSDT := ps.currentCapital * (percentage / 100.0)
	
	// Check if sufficient balance
	requiredBalance := amountUSDT
	if strings.ToLower(request.PositionType) == "futures" {
		requiredBalance = amountUSDT / float64(request.Leverage) // Margin required
	}
	
	if request.AvailableBalance < requiredBalance {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("insufficient balance: need %.2f USDT (%.2f%% of %.2f capital), have %.2f USDT", 
				requiredBalance, percentage, ps.currentCapital, request.AvailableBalance),
		}
	}
	
	// Calculate quantity
	var quantity float64
	if strings.ToLower(request.PositionType) == "futures" {
		// For futures: Quantity = (Amount * Leverage) / Price
		quantity = (amountUSDT * float64(request.Leverage)) / request.Price
	} else {
		// For spot: Quantity = Amount / Price
		quantity = amountUSDT / request.Price
	}
	
	return PositionSizingResult{
		Quantity:        quantity,
		NotionalValue:   amountUSDT * float64(request.Leverage), // Total position value
		RequiredBalance: requiredBalance,
		IsValid:         true,
	}
}

// applyLimitsAndPrecision applies position limits and symbol precision
func (ps *PositionSizer) applyLimitsAndPrecision(result PositionSizingResult, request PositionSizingRequest) PositionSizingResult {
	config := ps.config.PositionSizing
	
	// Apply position size limits
	if result.NotionalValue > config.MaxPositionSize {
		// Scale down to max position size
		scaleFactor := config.MaxPositionSize / result.NotionalValue
		result.Quantity *= scaleFactor
		result.NotionalValue = config.MaxPositionSize
		result.RequiredBalance *= scaleFactor
	}
	
	if result.NotionalValue < config.MinPositionSize {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("position size %.2f USDT below minimum %.2f USDT", result.NotionalValue, config.MinPositionSize),
		}
	}
	
	// Apply symbol precision
	precision := ps.getSymbolPrecision(request.Symbol)
	result.Quantity = ps.roundToPrecision(result.Quantity, precision)
	result.Precision = precision
	
	// Get minimum quantity for symbol
	minQty := ps.getMinimumQuantity(request.Symbol)
	result.MinimumQuantity = minQty
	
	if result.Quantity < minQty {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: fmt.Sprintf("calculated quantity %.8f below minimum %.8f for symbol %s", result.Quantity, minQty, request.Symbol),
		}
	}
	
	// Recalculate values with final quantity
	result.NotionalValue = result.Quantity * request.Price
	if strings.ToLower(request.PositionType) == "futures" {
		result.RequiredBalance = result.NotionalValue / float64(request.Leverage)
	} else {
		result.RequiredBalance = result.NotionalValue
	}
	
	return result
}

// UpdateCapital updates current capital for percentage calculations
func (ps *PositionSizer) UpdateCapital(newCapital float64) error {
	if newCapital <= 0 {
		return fmt.Errorf("capital must be positive")
	}
	
	ps.mutex.Lock()
	oldCapital := ps.currentCapital
	ps.currentCapital = newCapital
	ps.config.CurrentCapital = newCapital
	ps.mutex.Unlock()
	
	// Log capital update
	ps.auditLogger.LogConfigChange("PositionSizer", 
		map[string]interface{}{"old_capital": oldCapital}, 
		map[string]interface{}{"new_capital": newCapital},
		fmt.Sprintf("Capital updated from %.2f to %.2f USDT", oldCapital, newCapital))
	
	return nil
}

// UpdateConfiguration updates position sizing configuration
func (ps *PositionSizer) UpdateConfiguration(newConfig BaseConfiguration) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid new configuration: %w", err)
	}
	
	ps.mutex.Lock()
	oldConfig := ps.config
	ps.config = newConfig
	ps.currentCapital = newConfig.CurrentCapital
	ps.mutex.Unlock()
	
	// Log configuration change
	ps.auditLogger.LogConfigChange("PositionSizer", oldConfig, newConfig, "Position sizing configuration updated")
	
	return nil
}

// GetConfiguration returns current configuration (thread-safe copy)
func (ps *PositionSizer) GetConfiguration() BaseConfiguration {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.config
}

// AddPosition increments position counters
func (ps *PositionSizer) AddPosition() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.dailyPositions++
	ps.currentTrades++
}

// RemovePosition decrements current trade counter
func (ps *PositionSizer) RemovePosition() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if ps.currentTrades > 0 {
		ps.currentTrades--
	}
}

// ResetDailyCounters resets daily position counter
func (ps *PositionSizer) ResetDailyCounters() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.dailyPositions = 0
}

// validateRequest validates position sizing request
func (ps *PositionSizer) validateRequest(request PositionSizingRequest) error {
	if request.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if request.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if request.AvailableBalance < 0 {
		return fmt.Errorf("available balance cannot be negative")
	}
	if request.Leverage < 1 {
		request.Leverage = ps.config.PositionSizing.DefaultLeverage
	}
	if request.Leverage > 125 {
		return fmt.Errorf("leverage cannot exceed 125x")
	}
	return nil
}

// initializeDefaultPrecisions sets up default symbol precisions
func (ps *PositionSizer) initializeDefaultPrecisions() {
	// Common crypto pairs
	ps.symbolPrecision["BTC-USDT"] = 8
	ps.symbolPrecision["ETH-USDT"] = 6
	ps.symbolPrecision["SOL-USDT"] = 4
	ps.symbolPrecision["SUI-USDT"] = 3
	ps.symbolPrecision["ADA-USDT"] = 1
	ps.symbolPrecision["DOGE-USDT"] = 0
	
	// BingX format
	ps.symbolPrecision["BTCUSDT"] = 8
	ps.symbolPrecision["ETHUSDT"] = 6
	ps.symbolPrecision["SOLUSDT"] = 4
	ps.symbolPrecision["SUIUSDT"] = 3
}

// getSymbolPrecision returns decimal precision for symbol
func (ps *PositionSizer) getSymbolPrecision(symbol string) int {
	if precision, exists := ps.symbolPrecision[symbol]; exists {
		return precision
	}
	
	// Default precision based on symbol patterns
	symbol = strings.ToUpper(symbol)
	if strings.Contains(symbol, "BTC") {
		return 8
	}
	if strings.Contains(symbol, "ETH") {
		return 6
	}
	if strings.Contains(symbol, "SOL") || strings.Contains(symbol, "SUI") {
		return 4
	}
	
	// Default precision
	return 3
}

// getMinimumQuantity returns minimum quantity for symbol
func (ps *PositionSizer) getMinimumQuantity(symbol string) float64 {
	// Default minimums based on precision
	precision := ps.getSymbolPrecision(symbol)
	return math.Pow(10, -float64(precision))
}

// roundToPrecision rounds value to specified decimal places
func (ps *PositionSizer) roundToPrecision(value float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}
