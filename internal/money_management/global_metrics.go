package money_management

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// GlobalMetricsCollector handles cross-strategy performance metrics collection
type GlobalMetricsCollector struct {
	config BaseConfiguration
	mutex  sync.RWMutex
	
	// Current metrics state
	metrics GlobalMetrics
	
	// Trade tracking
	tradeHistory []TradeRecord
	maxHistorySize int
	
	// Daily tracking
	dailyStartPnL    float64
	dailyStartTime   time.Time
	dailyTradeCount  int
	
	// File persistence
	metricsFilePath string
	snapshotInterval time.Duration
	
	// Logging
	auditLogger *AuditLogger
	
	// Background tasks
	stopChan chan struct{}
	running  bool
}

// TradeRecord represents a completed trade for metrics
type TradeRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	StrategyName  string    `json:"strategy_name"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`          // "LONG" or "SHORT"
	EntryPrice    float64   `json:"entry_price"`
	ExitPrice     float64   `json:"exit_price"`
	Quantity      float64   `json:"quantity"`
	PnL           float64   `json:"pnl"`           // Profit/Loss in USDT
	PnLPercent    float64   `json:"pnl_percent"`   // % profit/loss
	Duration      int64     `json:"duration"`      // Trade duration in seconds
	IsWinning     bool      `json:"is_winning"`
}

// NewGlobalMetricsCollector creates a new metrics collector
func NewGlobalMetricsCollector(config BaseConfiguration, metricsFilePath string, auditLogger *AuditLogger) (*GlobalMetricsCollector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	gmc := &GlobalMetricsCollector{
		config:           config,
		metricsFilePath:  metricsFilePath,
		maxHistorySize:   10000, // Keep last 10k trades
		snapshotInterval: time.Duration(config.MetricsUpdateIntervalSeconds) * time.Second,
		auditLogger:      auditLogger,
		stopChan:         make(chan struct{}),
	}
	
	// Initialize metrics
	gmc.initializeMetrics()
	
	// Load persisted data
	if err := gmc.loadMetrics(); err != nil {
		fmt.Printf("Warning: could not load metrics data: %v\n", err)
	}
	
	// Start background tasks
	gmc.startBackgroundTasks()
	
	return gmc, nil
}

// RecordTrade records a completed trade
func (gmc *GlobalMetricsCollector) RecordTrade(trade TradeRecord) {
	gmc.mutex.Lock()
	defer gmc.mutex.Unlock()
	
	// Add to history
	gmc.tradeHistory = append(gmc.tradeHistory, trade)
	
	// Trim history if too large
	if len(gmc.tradeHistory) > gmc.maxHistorySize {
		gmc.tradeHistory = gmc.tradeHistory[len(gmc.tradeHistory)-gmc.maxHistorySize:]
	}
	
	// Update global metrics
	gmc.updateGlobalMetrics(trade)
	
	// Update strategy-specific metrics
	gmc.updateStrategyMetrics(trade)
	
	// Log trade record
	gmc.auditLogger.LogEmergencyAction("TRADE_RECORDED", map[string]interface{}{
		"trade": trade,
	}, true, fmt.Sprintf("Trade recorded: %s %s %.4f @ %.2f, PnL: %.2f USDT", 
		trade.StrategyName, trade.Symbol, trade.Quantity, trade.ExitPrice, trade.PnL))
}

// UpdateRealTimePnL updates floating PnL for open positions
func (gmc *GlobalMetricsCollector) UpdateRealTimePnL(totalPnL float64) {
	gmc.mutex.Lock()
	defer gmc.mutex.Unlock()
	
	gmc.metrics.TotalPnL = totalPnL
	gmc.metrics.DailyPnL = totalPnL - gmc.dailyStartPnL
	gmc.metrics.LastUpdateTime = time.Now().UTC()
	
	// Calculate loss percentages
	if gmc.config.StartingCapital > 0 {
		gmc.metrics.DailyLossPercent = (gmc.metrics.DailyPnL / gmc.config.StartingCapital) * 100
		gmc.metrics.MonthlyLossPercent = (gmc.calculateMonthlyPnL() / gmc.config.StartingCapital) * 100
	}
	
	// Update max drawdown
	currentDrawdown := gmc.calculateCurrentDrawdown()
	if currentDrawdown > gmc.metrics.MaxDrawdown {
		gmc.metrics.MaxDrawdown = currentDrawdown
	}
}

// GetMetrics returns current metrics (thread-safe copy)
func (gmc *GlobalMetricsCollector) GetMetrics() GlobalMetrics {
	gmc.mutex.RLock()
	defer gmc.mutex.RUnlock()
	return gmc.metrics
}

// GetStrategyMetrics returns metrics for specific strategy
func (gmc *GlobalMetricsCollector) GetStrategyMetrics(strategyName string) (StrategyMetrics, bool) {
	gmc.mutex.RLock()
	defer gmc.mutex.RUnlock()
	
	metrics, exists := gmc.metrics.StrategyMetrics[strategyName]
	return metrics, exists
}

// GetTopStrategies returns strategies ranked by performance
func (gmc *GlobalMetricsCollector) GetTopStrategies(limit int) []StrategyMetrics {
	gmc.mutex.RLock()
	defer gmc.mutex.RUnlock()
	
	var strategies []StrategyMetrics
	for _, metrics := range gmc.metrics.StrategyMetrics {
		strategies = append(strategies, metrics)
	}
	
	// Sort by PnL descending
	for i := 0; i < len(strategies)-1; i++ {
		for j := i + 1; j < len(strategies); j++ {
			if strategies[i].PnL < strategies[j].PnL {
				strategies[i], strategies[j] = strategies[j], strategies[i]
			}
		}
	}
	
	if limit > 0 && len(strategies) > limit {
		strategies = strategies[:limit]
	}
	
	return strategies
}

// ResetDailyMetrics resets daily counters at midnight
func (gmc *GlobalMetricsCollector) ResetDailyMetrics() {
	gmc.mutex.Lock()
	defer gmc.mutex.Unlock()
	
	gmc.dailyStartPnL = gmc.metrics.TotalPnL
	gmc.dailyStartTime = time.Now().UTC()
	gmc.dailyTradeCount = 0
	gmc.metrics.DayStartTime = gmc.dailyStartTime
	gmc.metrics.DailyPnL = 0
	
	// Log daily reset
	gmc.auditLogger.LogMetricsSnapshot(gmc.metrics, "Daily metrics reset at midnight UTC")
}

// GenerateDailyReport generates end-of-day performance report
func (gmc *GlobalMetricsCollector) GenerateDailyReport() DailyReport {
	gmc.mutex.RLock()
	defer gmc.mutex.RUnlock()
	
	// Generate risk analysis
	riskAnalysis := gmc.generateRiskAnalysis()
	
	// Generate recommendations
	recommendations := gmc.generateRecommendations()
	
	report := DailyReport{
		Date:                time.Now().UTC(),
		GlobalMetrics:       gmc.metrics,
		StrategyPerformance: gmc.metrics.StrategyMetrics,
		RiskAnalysis:        riskAnalysis,
		Recommendations:     recommendations,
	}
	
	// Log report generation
	gmc.auditLogger.LogEmergencyAction("DAILY_REPORT_GENERATED", map[string]interface{}{
		"report_date": report.Date,
		"total_pnl":   report.GlobalMetrics.TotalPnL,
		"strategies":  len(report.StrategyPerformance),
	}, true, fmt.Sprintf("Daily report generated: %.2f USDT total PnL, %d strategies", 
		report.GlobalMetrics.TotalPnL, len(report.StrategyPerformance)))
	
	return report
}

// updateGlobalMetrics updates global performance metrics
func (gmc *GlobalMetricsCollector) updateGlobalMetrics(trade TradeRecord) {
	gmc.metrics.TotalPnL += trade.PnL
	gmc.metrics.DailyPnL += trade.PnL
	gmc.metrics.TotalPositions++
	gmc.dailyTradeCount++
	
	if trade.IsWinning {
		gmc.metrics.WinningPositions++
	} else {
		gmc.metrics.LosingPositions++
	}
	
	// Recalculate ratios
	if gmc.metrics.TotalPositions > 0 {
		gmc.metrics.WinRate = float64(gmc.metrics.WinningPositions) / float64(gmc.metrics.TotalPositions) * 100
	}
	
	// Calculate profit factor
	totalWins, totalLosses := gmc.calculateWinsLosses()
	if totalLosses < 0 {
		gmc.metrics.ProfitFactor = totalWins / math.Abs(totalLosses)
	} else if totalWins > 0 {
		gmc.metrics.ProfitFactor = totalWins // No losses yet
	} else {
		gmc.metrics.ProfitFactor = 0 // No trades or all break-even
	}
}

// updateStrategyMetrics updates strategy-specific metrics
func (gmc *GlobalMetricsCollector) updateStrategyMetrics(trade TradeRecord) {
	if gmc.metrics.StrategyMetrics == nil {
		gmc.metrics.StrategyMetrics = make(map[string]StrategyMetrics)
	}
	
	metrics := gmc.metrics.StrategyMetrics[trade.StrategyName]
	metrics.StrategyName = trade.StrategyName
	metrics.PnL += trade.PnL
	metrics.Positions++
	metrics.LastTradeTime = trade.Timestamp
	metrics.IsActive = true
	
	if trade.IsWinning {
		metrics.WinningPositions++
	}
	
	// Calculate strategy ratios
	if metrics.Positions > 0 {
		metrics.WinRate = float64(metrics.WinningPositions) / float64(metrics.Positions) * 100
	}
	
	// Calculate strategy profit factor
	strategyWins, strategyLosses := gmc.calculateStrategyWinsLosses(trade.StrategyName)
	if strategyLosses < 0 {
		metrics.ProfitFactor = strategyWins / math.Abs(strategyLosses)
	} else if strategyWins > 0 {
		metrics.ProfitFactor = strategyWins
	} else {
		metrics.ProfitFactor = 0
	}
	
	gmc.metrics.StrategyMetrics[trade.StrategyName] = metrics
}

// calculateWinsLosses calculates total wins and losses
func (gmc *GlobalMetricsCollector) calculateWinsLosses() (float64, float64) {
	var totalWins, totalLosses float64
	
	for _, trade := range gmc.tradeHistory {
		if trade.PnL > 0 {
			totalWins += trade.PnL
		} else {
			totalLosses += trade.PnL
		}
	}
	
	return totalWins, totalLosses
}

// calculateStrategyWinsLosses calculates wins/losses for specific strategy
func (gmc *GlobalMetricsCollector) calculateStrategyWinsLosses(strategyName string) (float64, float64) {
	var wins, losses float64
	
	for _, trade := range gmc.tradeHistory {
		if trade.StrategyName == strategyName {
			if trade.PnL > 0 {
				wins += trade.PnL
			} else {
				losses += trade.PnL
			}
		}
	}
	
	return wins, losses
}

// calculateMonthlyPnL calculates PnL over last 30 days
func (gmc *GlobalMetricsCollector) calculateMonthlyPnL() float64 {
	thirtyDaysAgo := time.Now().UTC().AddDate(0, 0, -30)
	var monthlyPnL float64
	
	for _, trade := range gmc.tradeHistory {
		if trade.Timestamp.After(thirtyDaysAgo) {
			monthlyPnL += trade.PnL
		}
	}
	
	return monthlyPnL
}

// calculateCurrentDrawdown calculates current drawdown from peak
func (gmc *GlobalMetricsCollector) calculateCurrentDrawdown() float64 {
	if len(gmc.tradeHistory) < 2 {
		return 0
	}
	
	var peak, current float64
	runningPnL := gmc.config.StartingCapital
	
	for _, trade := range gmc.tradeHistory {
		runningPnL += trade.PnL
		if runningPnL > peak {
			peak = runningPnL
		}
		current = runningPnL
	}
	
	if peak == 0 {
		return 0
	}
	
	return ((peak - current) / peak) * 100
}

// generateRiskAnalysis analyzes current risk levels
func (gmc *GlobalMetricsCollector) generateRiskAnalysis() RiskAnalysis {
	analysis := RiskAnalysis{
		DistanceToDailyLimit:   gmc.config.DailyLimitPercent + gmc.metrics.DailyLossPercent,
		DistanceToMonthlyLimit: gmc.config.MonthlyLimitPercent + gmc.metrics.MonthlyLossPercent,
		AlertsGenerated:        []string{},
	}
	
	// Determine risk level
	if analysis.DistanceToDailyLimit <= 1.0 || analysis.DistanceToMonthlyLimit <= 3.0 {
		analysis.RiskLevel = "CRITICAL"
		analysis.AlertsGenerated = append(analysis.AlertsGenerated, "APPROACHING_LIMITS")
	} else if analysis.DistanceToDailyLimit <= 2.0 || analysis.DistanceToMonthlyLimit <= 5.0 {
		analysis.RiskLevel = "HIGH"
		analysis.AlertsGenerated = append(analysis.AlertsGenerated, "HIGH_RISK_WARNING")
	} else if analysis.DistanceToDailyLimit <= 3.0 || gmc.metrics.MaxDrawdown > 8.0 {
		analysis.RiskLevel = "MEDIUM"
	} else {
		analysis.RiskLevel = "LOW"
	}
	
	return analysis
}

// generateRecommendations generates trading recommendations
func (gmc *GlobalMetricsCollector) generateRecommendations() []string {
	var recommendations []string
	
	// Risk-based recommendations
	if gmc.metrics.DailyLossPercent < -3.0 {
		recommendations = append(recommendations, "Consider reducing position sizes - significant daily losses")
	}
	
	if gmc.metrics.WinRate < 40 {
		recommendations = append(recommendations, "Review strategy parameters - low win rate")
	}
	
	if gmc.metrics.ProfitFactor < 1.2 {
		recommendations = append(recommendations, "Improve risk/reward ratio - low profit factor")
	}
	
	// Strategy-specific recommendations
	bestStrategy := ""
	bestPnL := math.Inf(-1)
	for name, metrics := range gmc.metrics.StrategyMetrics {
		if metrics.PnL > bestPnL {
			bestPnL = metrics.PnL
			bestStrategy = name
		}
	}
	
	if bestStrategy != "" {
		recommendations = append(recommendations, fmt.Sprintf("Consider increasing allocation to %s strategy", bestStrategy))
	}
	
	return recommendations
}

// initializeMetrics initializes metrics structure
func (gmc *GlobalMetricsCollector) initializeMetrics() {
	now := time.Now().UTC()
	gmc.metrics = GlobalMetrics{
		LastUpdateTime:       now,
		DayStartTime:         now,
		StrategyMetrics:      make(map[string]StrategyMetrics),
	}
	gmc.dailyStartTime = now
}

// loadMetrics loads persisted metrics data
func (gmc *GlobalMetricsCollector) loadMetrics() error {
	data, err := os.ReadFile(gmc.metricsFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("metrics file does not exist")
	}
	if err != nil {
		return fmt.Errorf("failed to read metrics file: %w", err)
	}
	
	var persistedData struct {
		Metrics      GlobalMetrics `json:"metrics"`
		TradeHistory []TradeRecord `json:"trade_history"`
	}
	
	if err := json.Unmarshal(data, &persistedData); err != nil {
		return fmt.Errorf("failed to unmarshal metrics data: %w", err)
	}
	
	gmc.metrics = persistedData.Metrics
	gmc.tradeHistory = persistedData.TradeHistory
	
	return nil
}

// persistMetrics saves metrics data to file
func (gmc *GlobalMetricsCollector) persistMetrics() error {
	// Ensure directory exists
	dir := filepath.Dir(gmc.metricsFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}
	
	persistData := struct {
		Metrics      GlobalMetrics `json:"metrics"`
		TradeHistory []TradeRecord `json:"trade_history"`
	}{
		Metrics:      gmc.metrics,
		TradeHistory: gmc.tradeHistory,
	}
	
	data, err := json.MarshalIndent(persistData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}
	
	return os.WriteFile(gmc.metricsFilePath, data, 0644)
}

// startBackgroundTasks starts background metric collection tasks
func (gmc *GlobalMetricsCollector) startBackgroundTasks() {
	gmc.running = true
	
	go func() {
		snapshotTicker := time.NewTicker(gmc.snapshotInterval)
		dailyTicker := time.NewTicker(24 * time.Hour)
		
		defer snapshotTicker.Stop()
		defer dailyTicker.Stop()
		
		for {
			select {
			case <-gmc.stopChan:
				return
			case <-snapshotTicker.C:
				// Periodic metrics snapshot
				gmc.persistMetrics()
				gmc.auditLogger.LogMetricsSnapshot(gmc.GetMetrics(), "Periodic metrics snapshot")
			case <-dailyTicker.C:
				// Daily reset and report
				report := gmc.GenerateDailyReport()
				gmc.ResetDailyMetrics()
				
				// Save daily report
				gmc.saveDailyReport(report)
			}
		}
	}()
}

// saveDailyReport saves daily report to file
func (gmc *GlobalMetricsCollector) saveDailyReport(report DailyReport) {
	reportsDir := filepath.Join(filepath.Dir(gmc.metricsFilePath), "daily_reports")
	os.MkdirAll(reportsDir, 0755)
	
	filename := fmt.Sprintf("daily_report_%s.json", report.Date.Format("2006-01-02"))
	reportPath := filepath.Join(reportsDir, filename)
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling daily report: %v\n", err)
		return
	}
	
	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		fmt.Printf("Error saving daily report: %v\n", err)
	}
}

// Stop stops the metrics collector
func (gmc *GlobalMetricsCollector) Stop() error {
	if gmc.running {
		close(gmc.stopChan)
		gmc.running = false
		
		// Final metrics save
		return gmc.persistMetrics()
	}
	return nil
}
