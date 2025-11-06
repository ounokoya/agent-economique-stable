package money_management

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// MoneyManager is the main coordinator for all Money Management BASE components
type MoneyManager struct {
	// Core MM components
	circuitBreaker      *CircuitBreaker
	positionSizer       *PositionSizer
	metricsCollector    *GlobalMetricsCollector
	auditLogger         *AuditLogger
	
	// Configuration
	config BaseConfiguration
	
	// State tracking
	isActive            bool
	currentTotalPnL     float64
	currentCapital      float64
	
	// Integration callbacks
	emergencyStopCallback func() error // Function to close all positions
	
	// Thread safety
	mutex sync.RWMutex
}

// MoneyManagerConfig holds initialization configuration
type MoneyManagerConfig struct {
	BaseConfig         BaseConfiguration
	DataDirectory      string  // Directory for logs and persistence
	StrategyName       string  // Strategy name for metrics
}

// NewMoneyManager creates a new Money Manager with all components
func NewMoneyManager(config MoneyManagerConfig) (*MoneyManager, error) {
	if err := config.BaseConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid base configuration: %w", err)
	}
	
	// Initialize audit logger
	auditDir := filepath.Join(config.DataDirectory, "audit")
	auditLogger, err := NewAuditLogger(auditDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}
	
	// Initialize circuit breaker
	stateFile := filepath.Join(config.DataDirectory, "circuit_breaker_state.json")
	circuitBreaker, err := NewCircuitBreaker(config.BaseConfig, stateFile, auditLogger)
	if err != nil {
		auditLogger.Close()
		return nil, fmt.Errorf("failed to create circuit breaker: %w", err)
	}
	
	// Initialize position sizer
	positionSizer, err := NewPositionSizer(config.BaseConfig, auditLogger)
	if err != nil {
		auditLogger.Close()
		return nil, fmt.Errorf("failed to create position sizer: %w", err)
	}
	
	// Initialize metrics collector
	metricsFile := filepath.Join(config.DataDirectory, "global_metrics.json")
	metricsCollector, err := NewGlobalMetricsCollector(config.BaseConfig, metricsFile, auditLogger)
	if err != nil {
		auditLogger.Close()
		return nil, fmt.Errorf("failed to create metrics collector: %w", err)
	}
	
	mm := &MoneyManager{
		circuitBreaker:   circuitBreaker,
		positionSizer:    positionSizer,
		metricsCollector: metricsCollector,
		auditLogger:      auditLogger,
		config:           config.BaseConfig,
		isActive:         true,
		currentCapital:   config.BaseConfig.CurrentCapital,
	}
	
	// Set emergency stop callback for circuit breaker
	circuitBreaker.SetEmergencyStopCallback(func() error {
		if mm.emergencyStopCallback != nil {
			return mm.emergencyStopCallback()
		}
		return fmt.Errorf("no emergency stop callback configured")
	})
	
	// Start background processes
	circuitBreaker.StartPeriodicChecks()
	
	return mm, nil
}

// SetEmergencyStopCallback sets the function to call for emergency position closure
func (mm *MoneyManager) SetEmergencyStopCallback(callback func() error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.emergencyStopCallback = callback
}

// CalculatePositionSize calculates position size and validates against limits
func (mm *MoneyManager) CalculatePositionSize(request PositionSizingRequest) (PositionSizingResult, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	// Check if trading is halted by circuit breakers
	if mm.circuitBreaker.IsActive() {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: "Trading halted - circuit breaker active",
		}, ErrCircuitBreakerActive
	}
	
	// Check if MM is active
	if !mm.isActive {
		return PositionSizingResult{
			IsValid:         false,
			ValidationError: "Money management is disabled",
		}, fmt.Errorf("money management disabled")
	}
	
	// Calculate position size
	result := mm.positionSizer.CalculatePositionSize(request)
	
	return result, nil
}

// ValidateNewPosition validates if a new position can be opened
func (mm *MoneyManager) ValidateNewPosition(request PositionSizingRequest) error {
	result, err := mm.CalculatePositionSize(request)
	if err != nil {
		return err
	}
	
	if !result.IsValid {
		return fmt.Errorf("position validation failed: %s", result.ValidationError)
	}
	
	return nil
}

// OnPositionOpened should be called when a position is successfully opened
func (mm *MoneyManager) OnPositionOpened(positionValue float64, strategyName string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	// Update position counters
	mm.positionSizer.AddPosition()
	
	// Log position opening
	mm.auditLogger.LogEmergencyAction("POSITION_OPENED", map[string]interface{}{
		"strategy":       strategyName,
		"position_value": positionValue,
		"timestamp":      time.Now().UTC(),
	}, true, fmt.Sprintf("Position opened: %.2f USDT by %s", positionValue, strategyName))
}

// OnPositionClosed should be called when a position is closed
func (mm *MoneyManager) OnPositionClosed(trade TradeRecord) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	// Update position counters
	mm.positionSizer.RemovePosition()
	
	// Record trade in metrics
	mm.metricsCollector.RecordTrade(trade)
	
	// Update total PnL
	mm.currentTotalPnL += trade.PnL
	
	// Update current capital if in percentage mode
	if mm.config.PositionSizing.Mode == PositionSizingPercentage {
		mm.currentCapital += trade.PnL
		mm.positionSizer.UpdateCapital(mm.currentCapital)
		mm.metricsCollector.UpdateRealTimePnL(mm.currentTotalPnL)
	}
	
	return nil
}

// UpdateRealTimePnL updates floating PnL for risk monitoring
func (mm *MoneyManager) UpdateRealTimePnL(totalPnL float64) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.currentTotalPnL = totalPnL
	
	// Update metrics collector
	mm.metricsCollector.UpdateRealTimePnL(totalPnL)
	
	// Check circuit breaker limits
	if err := mm.circuitBreaker.CheckDailyLimit(totalPnL); err != nil {
		if err == ErrDailyLimitBreached {
			mm.auditLogger.LogCircuitBreaker("DAILY_BREACH_TRIGGERED", map[string]interface{}{
				"total_pnl": totalPnL,
				"timestamp": time.Now().UTC(),
			}, true, "Daily limit breached - emergency stop executed")
		}
		return err
	}
	
	if err := mm.circuitBreaker.CheckMonthlyLimit(totalPnL); err != nil {
		if err == ErrMonthlyLimitBreached {
			mm.auditLogger.LogCircuitBreaker("MONTHLY_BREACH_TRIGGERED", map[string]interface{}{
				"total_pnl": totalPnL,
				"timestamp": time.Now().UTC(),
			}, true, "Monthly limit breached - emergency stop executed")
		}
		return err
	}
	
	return nil
}

// GetCurrentMetrics returns current performance metrics
func (mm *MoneyManager) GetCurrentMetrics() GlobalMetrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.metricsCollector.GetMetrics()
}

// GetCircuitBreakerState returns current circuit breaker state
func (mm *MoneyManager) GetCircuitBreakerState() CircuitBreakerState {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.circuitBreaker.GetState()
}

// IsActive returns whether money management is currently active
func (mm *MoneyManager) IsActive() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.isActive && !mm.circuitBreaker.IsActive()
}

// IsTradingHalted returns whether trading is currently halted
func (mm *MoneyManager) IsTradingHalted() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.circuitBreaker.IsActive()
}

// GetRiskAnalysis returns current risk analysis
func (mm *MoneyManager) GetRiskAnalysis() RiskAnalysis {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	report := mm.metricsCollector.GenerateDailyReport()
	return report.RiskAnalysis
}

// GenerateDailyReport generates comprehensive daily report
func (mm *MoneyManager) GenerateDailyReport() DailyReport {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	report := mm.metricsCollector.GenerateDailyReport()
	report.CircuitBreakerState = mm.circuitBreaker.GetState()
	
	return report
}

// UpdateConfiguration updates MM configuration at runtime
func (mm *MoneyManager) UpdateConfiguration(newConfig BaseConfiguration) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	oldConfig := mm.config
	mm.config = newConfig
	
	// Update all components
	if err := mm.circuitBreaker.UpdateConfiguration(newConfig); err != nil {
		mm.config = oldConfig // Rollback
		return fmt.Errorf("failed to update circuit breaker config: %w", err)
	}
	
	if err := mm.positionSizer.UpdateConfiguration(newConfig); err != nil {
		mm.config = oldConfig // Rollback
		mm.circuitBreaker.UpdateConfiguration(oldConfig) // Rollback CB too
		return fmt.Errorf("failed to update position sizer config: %w", err)
	}
	
	// Update current capital if changed
	if newConfig.CurrentCapital != mm.currentCapital {
		mm.currentCapital = newConfig.CurrentCapital
		mm.positionSizer.UpdateCapital(mm.currentCapital)
	}
	
	// Log configuration change
	mm.auditLogger.LogConfigChange("MoneyManager", oldConfig, newConfig, "Money Manager configuration updated successfully")
	
	return nil
}

// ResetDailyCounters resets daily metrics and counters
func (mm *MoneyManager) ResetDailyCounters() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	// Reset position sizer daily counters
	mm.positionSizer.ResetDailyCounters()
	
	// Reset metrics collector daily metrics
	mm.metricsCollector.ResetDailyMetrics()
	
	// Log daily reset
	mm.auditLogger.LogEmergencyAction("DAILY_RESET", map[string]interface{}{
		"timestamp": time.Now().UTC(),
	}, true, "Daily counters and metrics reset at midnight UTC")
}

// GetCurrentCapital returns current capital amount
func (mm *MoneyManager) GetCurrentCapital() float64 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.currentCapital
}

// GetPositionSizingMode returns current position sizing mode
func (mm *MoneyManager) GetPositionSizingMode() PositionSizingMode {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.config.PositionSizing.Mode
}

// Enable enables money management operations
func (mm *MoneyManager) Enable() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.isActive = true
	
	mm.auditLogger.LogEmergencyAction("MM_ENABLED", map[string]interface{}{
		"timestamp": time.Now().UTC(),
	}, true, "Money Management enabled")
}

// Disable disables money management operations (emergency use)
func (mm *MoneyManager) Disable() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.isActive = false
	
	mm.auditLogger.LogEmergencyAction("MM_DISABLED", map[string]interface{}{
		"timestamp": time.Now().UTC(),
	}, true, "Money Management disabled")
}

// Stop gracefully stops all MM components and saves state
func (mm *MoneyManager) Stop() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	// Stop metrics collector
	if err := mm.metricsCollector.Stop(); err != nil {
		fmt.Printf("Warning: error stopping metrics collector: %v\n", err)
	}
	
	// Close audit logger
	if err := mm.auditLogger.Close(); err != nil {
		fmt.Printf("Warning: error closing audit logger: %v\n", err)
	}
	
	mm.isActive = false
	
	return nil
}

// MoneyManagerStatus provides comprehensive status information
type MoneyManagerStatus struct {
	IsActive              bool                    `json:"is_active"`
	IsTradingHalted       bool                    `json:"is_trading_halted"`
	CurrentCapital        float64                 `json:"current_capital"`
	CurrentTotalPnL       float64                 `json:"current_total_pnl"`
	PositionSizingMode    PositionSizingMode      `json:"position_sizing_mode"`
	CircuitBreakerState   CircuitBreakerState     `json:"circuit_breaker_state"`
	GlobalMetrics         GlobalMetrics           `json:"global_metrics"`
	RiskAnalysis          RiskAnalysis            `json:"risk_analysis"`
}

// GetStatus returns comprehensive MM status
func (mm *MoneyManager) GetStatus() MoneyManagerStatus {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	return MoneyManagerStatus{
		IsActive:            mm.isActive,
		IsTradingHalted:     mm.circuitBreaker.IsActive(),
		CurrentCapital:      mm.currentCapital,
		CurrentTotalPnL:     mm.currentTotalPnL,
		PositionSizingMode:  mm.config.PositionSizing.Mode,
		CircuitBreakerState: mm.circuitBreaker.GetState(),
		GlobalMetrics:       mm.metricsCollector.GetMetrics(),
		RiskAnalysis:        mm.GetRiskAnalysis(),
	}
}
