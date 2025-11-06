package money_management

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// CircuitBreaker manages global risk limits and emergency stops
type CircuitBreaker struct {
	config BaseConfiguration
	state  CircuitBreakerState
	mutex  sync.RWMutex
	
	// State persistence
	stateFilePath string
	
	// Metrics tracking
	startingCapital float64
	dailyStartPnL   float64
	monthlyPnLs     []float64 // Last 30 days of PnL
	
	// Callbacks for emergency actions
	emergencyStopCallback func() error // Function to close all positions
	
	// Logging
	auditLogger *AuditLogger
}

// NewCircuitBreaker creates a new circuit breaker with configuration
func NewCircuitBreaker(config BaseConfiguration, stateFilePath string, auditLogger *AuditLogger) (*CircuitBreaker, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	cb := &CircuitBreaker{
		config:          config,
		stateFilePath:   stateFilePath,
		startingCapital: config.StartingCapital,
		auditLogger:     auditLogger,
		monthlyPnLs:     make([]float64, 0),
	}
	
	// Load persisted state or initialize
	if err := cb.loadState(); err != nil {
		log.Printf("Warning: could not load circuit breaker state: %v", err)
		cb.initializeState()
	}
	
	// Reset daily if needed
	if cb.state.ShouldResetDaily() {
		cb.resetDaily()
	}
	
	// Retry monthly if needed
	if cb.state.ShouldRetryMonthly() {
		cb.retryMonthly()
	}
	
	return cb, nil
}

// SetEmergencyStopCallback sets the callback function for emergency position closure
func (cb *CircuitBreaker) SetEmergencyStopCallback(callback func() error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.emergencyStopCallback = callback
}

// CheckDailyLimit checks if daily PnL exceeds limit and triggers circuit breaker
func (cb *CircuitBreaker) CheckDailyLimit(currentPnL float64) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	// Skip if circuit breaker already active
	if cb.state.IsActive() {
		return ErrCircuitBreakerActive
	}
	
	// Calculate daily loss percentage
	dailyPnL := currentPnL - cb.dailyStartPnL
	dailyLossPercent := (dailyPnL / cb.startingCapital) * 100
	
	// Check if daily limit breached
	if dailyLossPercent <= -cb.config.DailyLimitPercent {
		return cb.triggerDailyBreaker(dailyPnL, dailyLossPercent)
	}
	
	return nil
}

// CheckMonthlyLimit checks if monthly PnL exceeds limit and triggers circuit breaker
func (cb *CircuitBreaker) CheckMonthlyLimit(currentPnL float64) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	// Skip if circuit breaker already active
	if cb.state.IsActive() {
		return ErrCircuitBreakerActive
	}
	
	// Calculate monthly PnL (simplified: current vs 30 days ago)
	monthlyPnL := cb.calculateMonthlyPnL(currentPnL)
	monthlyLossPercent := (monthlyPnL / cb.startingCapital) * 100
	
	// Check if monthly limit breached
	if monthlyLossPercent <= -cb.config.MonthlyLimitPercent {
		return cb.triggerMonthlyBreaker(monthlyPnL, monthlyLossPercent)
	}
	
	return nil
}

// IsActive returns true if any circuit breaker is currently active
func (cb *CircuitBreaker) IsActive() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state.IsActive()
}

// GetState returns current circuit breaker state (thread-safe copy)
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// UpdateConfiguration updates MM configuration runtime
func (cb *CircuitBreaker) UpdateConfiguration(newConfig BaseConfiguration) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid new configuration: %w", err)
	}
	
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	oldConfig := cb.config
	cb.config = newConfig
	
	// Log configuration change
	cb.auditLogger.LogConfigChange("CircuitBreaker", oldConfig, newConfig, "Configuration updated successfully")
	
	return cb.persistState()
}

// triggerDailyBreaker activates daily circuit breaker and executes emergency stop
func (cb *CircuitBreaker) triggerDailyBreaker(dailyPnL, lossPercent float64) error {
	cb.state.DailyBreakerActive = true
	cb.state.LastDailyBreach = time.Now().UTC()
	cb.state.DailyResetTime = getNextMidnightUTC()
	cb.state.TotalBreaches++
	
	// Execute emergency stop
	emergencyError := cb.executeEmergencyStop()
	
	// Log circuit breaker activation
	cb.auditLogger.LogCircuitBreaker("DAILY_LIMIT_BREACH", map[string]interface{}{
		"daily_pnl":        dailyPnL,
		"loss_percent":     lossPercent,
		"limit_percent":    cb.config.DailyLimitPercent,
		"emergency_stop":   emergencyError == nil,
		"next_reset_time":  cb.state.DailyResetTime,
	}, emergencyError == nil, fmt.Sprintf("Emergency stop: %v", emergencyError))
	
	// Persist state
	cb.persistState()
	
	if emergencyError != nil {
		return fmt.Errorf("daily circuit breaker triggered but emergency stop failed: %w", emergencyError)
	}
	
	return ErrDailyLimitBreached
}

// triggerMonthlyBreaker activates monthly circuit breaker
func (cb *CircuitBreaker) triggerMonthlyBreaker(monthlyPnL, lossPercent float64) error {
	cb.state.MonthlyBreakerActive = true
	cb.state.LastMonthlyBreach = time.Now().UTC()
	cb.state.MonthlyRetryTime = getNextMidnightUTC()
	cb.state.TotalBreaches++
	
	// Execute emergency stop
	emergencyError := cb.executeEmergencyStop()
	
	// Log circuit breaker activation
	cb.auditLogger.LogCircuitBreaker("MONTHLY_LIMIT_BREACH", map[string]interface{}{
		"monthly_pnl":      monthlyPnL,
		"loss_percent":     lossPercent,
		"limit_percent":    cb.config.MonthlyLimitPercent,
		"emergency_stop":   emergencyError == nil,
		"retry_time":       cb.state.MonthlyRetryTime,
	}, emergencyError == nil, fmt.Sprintf("Emergency stop: %v", emergencyError))
	
	// Persist state
	cb.persistState()
	
	if emergencyError != nil {
		return fmt.Errorf("monthly circuit breaker triggered but emergency stop failed: %w", emergencyError)
	}
	
	return ErrMonthlyLimitBreached
}

// executeEmergencyStop calls the emergency stop callback to close all positions
func (cb *CircuitBreaker) executeEmergencyStop() error {
	if cb.emergencyStopCallback == nil {
		return fmt.Errorf("no emergency stop callback configured")
	}
	
	log.Printf("CRITICAL: Executing emergency stop - closing all positions")
	return cb.emergencyStopCallback()
}

// resetDaily resets daily circuit breaker at midnight UTC
func (cb *CircuitBreaker) resetDaily() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if cb.state.DailyBreakerActive {
		cb.state.DailyBreakerActive = false
		
		// Log reset
		cb.auditLogger.LogCircuitBreaker("DAILY_RESET", map[string]interface{}{
			"reset_time": time.Now().UTC(),
		}, true, "Daily circuit breaker reset at midnight UTC")
		
		log.Printf("Daily circuit breaker reset at midnight UTC")
	}
	
	// Update daily start tracking
	cb.dailyStartPnL = 0 // This should be set to current total PnL
	cb.persistState()
}

// retryMonthly retries monthly circuit breaker the next day
func (cb *CircuitBreaker) retryMonthly() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if cb.state.MonthlyBreakerActive {
		cb.state.MonthlyBreakerActive = false
		
		// Log retry
		cb.auditLogger.LogCircuitBreaker("MONTHLY_RETRY", map[string]interface{}{
			"retry_time": time.Now().UTC(),
		}, true, "Monthly circuit breaker retry activated")
		
		log.Printf("Monthly circuit breaker retry activated")
	}
	
	cb.persistState()
}

// calculateMonthlyPnL calculates PnL over last 30 days (simplified)
func (cb *CircuitBreaker) calculateMonthlyPnL(currentPnL float64) float64 {
	// Simplified: assume current PnL represents monthly performance
	// In production, this should track daily PnLs over 30-day window
	if len(cb.monthlyPnLs) == 0 {
		return currentPnL
	}
	
	// Return difference from 30 days ago (or earliest available)
	oldestIndex := 0
	if len(cb.monthlyPnLs) > 30 {
		oldestIndex = len(cb.monthlyPnLs) - 30
	}
	
	return currentPnL - cb.monthlyPnLs[oldestIndex]
}

// initializeState creates initial circuit breaker state
func (cb *CircuitBreaker) initializeState() {
	cb.state = CircuitBreakerState{
		DailyBreakerActive:   false,
		MonthlyBreakerActive: false,
		DailyResetTime:       getNextMidnightUTC(),
		TotalBreaches:        0,
	}
}

// loadState loads circuit breaker state from file
func (cb *CircuitBreaker) loadState() error {
	data, err := os.ReadFile(cb.stateFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("state file does not exist")
	}
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}
	
	return json.Unmarshal(data, &cb.state)
}

// persistState saves circuit breaker state to file
func (cb *CircuitBreaker) persistState() error {
	data, err := json.MarshalIndent(cb.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	
	return os.WriteFile(cb.stateFilePath, data, 0644)
}

// getNextMidnightUTC returns next midnight UTC time
func getNextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

// StartPeriodicChecks starts periodic circuit breaker validation (call in main loop)
func (cb *CircuitBreaker) StartPeriodicChecks() {
	go func() {
		ticker := time.NewTicker(time.Minute) // Check every minute
		defer ticker.Stop()
		
		for range ticker.C {
			// Check if daily reset is needed
			if cb.state.ShouldResetDaily() {
				cb.resetDaily()
			}
			
			// Check if monthly retry is needed
			if cb.state.ShouldRetryMonthly() {
				cb.retryMonthly()
			}
		}
	}()
}
