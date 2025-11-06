package money_management

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewGlobalMetricsCollector(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, err := NewAuditLogger(filepath.Join(tempDir, "audit"))
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Verify initial state
	metrics := gmc.GetMetrics()
	if metrics.TotalPositions != 0 {
		t.Errorf("Expected 0 total positions, got %d", metrics.TotalPositions)
	}
	if metrics.StrategyMetrics == nil {
		t.Error("Strategy metrics map should be initialized")
	}
}

func TestRecordTrade(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Record winning trade
	winningTrade := TradeRecord{
		Timestamp:    time.Now().UTC(),
		StrategyName: "MACD_CCI_DMI",
		Symbol:       "BTC-USDT",
		Side:         "LONG",
		EntryPrice:   50000.0,
		ExitPrice:    51000.0,
		Quantity:     0.02,
		PnL:          20.0,
		PnLPercent:   2.0,
		Duration:     3600, // 1 hour
		IsWinning:    true,
	}
	
	gmc.RecordTrade(winningTrade)
	
	// Verify global metrics
	metrics := gmc.GetMetrics()
	if metrics.TotalPositions != 1 {
		t.Errorf("Expected 1 total position, got %d", metrics.TotalPositions)
	}
	if metrics.WinningPositions != 1 {
		t.Errorf("Expected 1 winning position, got %d", metrics.WinningPositions)
	}
	if metrics.WinRate != 100.0 {
		t.Errorf("Expected 100%% win rate, got %.2f%%", metrics.WinRate)
	}
	if metrics.TotalPnL != 20.0 {
		t.Errorf("Expected 20.0 total PnL, got %.2f", metrics.TotalPnL)
	}
	
	// Verify strategy metrics
	strategyMetrics, exists := gmc.GetStrategyMetrics("MACD_CCI_DMI")
	if !exists {
		t.Fatal("Strategy metrics should exist")
	}
	if strategyMetrics.Positions != 1 {
		t.Errorf("Expected 1 strategy position, got %d", strategyMetrics.Positions)
	}
	if strategyMetrics.PnL != 20.0 {
		t.Errorf("Expected 20.0 strategy PnL, got %.2f", strategyMetrics.PnL)
	}
	if strategyMetrics.WinRate != 100.0 {
		t.Errorf("Expected 100%% strategy win rate, got %.2f%%", strategyMetrics.WinRate)
	}
}

func TestMultipleTradesAndStrategies(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Record trades from different strategies
	trades := []TradeRecord{
		// MACD_CCI_DMI strategy - 2 wins, 1 loss
		{
			StrategyName: "MACD_CCI_DMI",
			PnL:          100.0,
			IsWinning:    true,
		},
		{
			StrategyName: "MACD_CCI_DMI",
			PnL:          50.0,
			IsWinning:    true,
		},
		{
			StrategyName: "MACD_CCI_DMI",
			PnL:          -30.0,
			IsWinning:    false,
		},
		// RSI_BB strategy - 1 win, 2 losses
		{
			StrategyName: "RSI_BB",
			PnL:          80.0,
			IsWinning:    true,
		},
		{
			StrategyName: "RSI_BB",
			PnL:          -40.0,
			IsWinning:    false,
		},
		{
			StrategyName: "RSI_BB",
			PnL:          -25.0,
			IsWinning:    false,
		},
	}
	
	for _, trade := range trades {
		trade.Timestamp = time.Now().UTC()
		gmc.RecordTrade(trade)
	}
	
	// Verify global metrics
	metrics := gmc.GetMetrics()
	if metrics.TotalPositions != 6 {
		t.Errorf("Expected 6 total positions, got %d", metrics.TotalPositions)
	}
	if metrics.WinningPositions != 3 {
		t.Errorf("Expected 3 winning positions, got %d", metrics.WinningPositions)
	}
	if metrics.LosingPositions != 3 {
		t.Errorf("Expected 3 losing positions, got %d", metrics.LosingPositions)
	}
	
	expectedWinRate := 50.0 // 3/6 * 100
	if abs(metrics.WinRate-expectedWinRate) > 0.01 {
		t.Errorf("Expected %.2f%% win rate, got %.2f%%", expectedWinRate, metrics.WinRate)
	}
	
	expectedTotalPnL := 135.0 // 100+50-30+80-40-25
	if abs(metrics.TotalPnL-expectedTotalPnL) > 0.01 {
		t.Errorf("Expected %.2f total PnL, got %.2f", expectedTotalPnL, metrics.TotalPnL)
	}
	
	// Verify strategy-specific metrics
	macdMetrics, exists := gmc.GetStrategyMetrics("MACD_CCI_DMI")
	if !exists {
		t.Fatal("MACD_CCI_DMI strategy metrics should exist")
	}
	if macdMetrics.Positions != 3 {
		t.Errorf("Expected 3 MACD positions, got %d", macdMetrics.Positions)
	}
	expectedMacdPnL := 120.0 // 100+50-30
	if abs(macdMetrics.PnL-expectedMacdPnL) > 0.01 {
		t.Errorf("Expected %.2f MACD PnL, got %.2f", expectedMacdPnL, macdMetrics.PnL)
	}
	expectedMacdWinRate := 66.67 // 2/3 * 100
	if abs(macdMetrics.WinRate-expectedMacdWinRate) > 0.1 {
		t.Errorf("Expected %.2f%% MACD win rate, got %.2f%%", expectedMacdWinRate, macdMetrics.WinRate)
	}
	
	rsiMetrics, exists := gmc.GetStrategyMetrics("RSI_BB")
	if !exists {
		t.Fatal("RSI_BB strategy metrics should exist")
	}
	if rsiMetrics.Positions != 3 {
		t.Errorf("Expected 3 RSI positions, got %d", rsiMetrics.Positions)
	}
	expectedRsiPnL := 15.0 // 80-40-25
	if abs(rsiMetrics.PnL-expectedRsiPnL) > 0.01 {
		t.Errorf("Expected %.2f RSI PnL, got %.2f", expectedRsiPnL, rsiMetrics.PnL)
	}
}

func TestProfitFactor(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Record trades with known profit factor
	trades := []TradeRecord{
		{PnL: 100.0, IsWinning: true},  // Total wins: 150
		{PnL: 50.0, IsWinning: true},
		{PnL: -30.0, IsWinning: false}, // Total losses: -80
		{PnL: -50.0, IsWinning: false},
	}
	
	for _, trade := range trades {
		trade.Timestamp = time.Now().UTC()
		trade.StrategyName = "TEST"
		gmc.RecordTrade(trade)
	}
	
	metrics := gmc.GetMetrics()
	expectedProfitFactor := 150.0 / 80.0 // 1.875
	if abs(metrics.ProfitFactor-expectedProfitFactor) > 0.01 {
		t.Errorf("Expected %.3f profit factor, got %.3f", expectedProfitFactor, metrics.ProfitFactor)
	}
}

func TestRealTimePnLUpdate(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.StartingCapital = 10000.0
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Update real-time PnL
	gmc.UpdateRealTimePnL(-300.0) // -3% loss
	
	metrics := gmc.GetMetrics()
	if metrics.TotalPnL != -300.0 {
		t.Errorf("Expected -300.0 total PnL, got %.2f", metrics.TotalPnL)
	}
	
	expectedDailyLossPercent := -3.0 // -300/10000 * 100
	if abs(metrics.DailyLossPercent-expectedDailyLossPercent) > 0.01 {
		t.Errorf("Expected %.2f%% daily loss, got %.2f%%", expectedDailyLossPercent, metrics.DailyLossPercent)
	}
}

func TestTopStrategiesRanking(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Record trades with different strategy performance
	strategies := []struct {
		name string
		pnl  float64
	}{
		{"Strategy_A", 200.0}, // Best
		{"Strategy_B", 100.0}, // Second
		{"Strategy_C", -50.0}, // Worst
	}
	
	for _, strategy := range strategies {
		trade := TradeRecord{
			Timestamp:    time.Now().UTC(),
			StrategyName: strategy.name,
			PnL:          strategy.pnl,
			IsWinning:    strategy.pnl > 0,
		}
		gmc.RecordTrade(trade)
	}
	
	// Get top strategies
	topStrategies := gmc.GetTopStrategies(3)
	if len(topStrategies) != 3 {
		t.Errorf("Expected 3 strategies, got %d", len(topStrategies))
	}
	
	// Verify ranking order
	if topStrategies[0].StrategyName != "Strategy_A" {
		t.Errorf("Expected Strategy_A first, got %s", topStrategies[0].StrategyName)
	}
	if topStrategies[1].StrategyName != "Strategy_B" {
		t.Errorf("Expected Strategy_B second, got %s", topStrategies[1].StrategyName)
	}
	if topStrategies[2].StrategyName != "Strategy_C" {
		t.Errorf("Expected Strategy_C third, got %s", topStrategies[2].StrategyName)
	}
}

func TestRiskAnalysis(t *testing.T) {
	config := DefaultBaseConfiguration()
	config.DailyLimitPercent = 5.0   // -5% daily limit
	config.MonthlyLimitPercent = 15.0 // -15% monthly limit
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Simulate critical risk scenario
	gmc.UpdateRealTimePnL(-400.0) // -4% loss, 1% from daily limit
	
	report := gmc.GenerateDailyReport()
	riskAnalysis := report.RiskAnalysis
	
	if riskAnalysis.RiskLevel != "CRITICAL" {
		t.Errorf("Expected CRITICAL risk level, got %s", riskAnalysis.RiskLevel)
	}
	
	expectedDistance := 1.0 // 5% limit - 4% loss = 1%
	if abs(riskAnalysis.DistanceToDailyLimit-expectedDistance) > 0.1 {
		t.Errorf("Expected %.1f%% distance to daily limit, got %.1f%%", expectedDistance, riskAnalysis.DistanceToDailyLimit)
	}
	
	if len(riskAnalysis.AlertsGenerated) == 0 {
		t.Error("Expected alerts to be generated for critical risk")
	}
}

func TestDailyReset(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	gmc, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create global metrics collector: %v", err)
	}
	defer gmc.Stop()
	
	// Record some trades
	trade := TradeRecord{
		Timestamp:    time.Now().UTC(),
		StrategyName: "TEST",
		PnL:          100.0,
		IsWinning:    true,
	}
	gmc.RecordTrade(trade)
	
	// Update real-time PnL
	gmc.UpdateRealTimePnL(150.0)
	
	// Verify initial state
	metrics := gmc.GetMetrics()
	if metrics.DailyPnL != 150.0 {
		t.Errorf("Expected 150.0 daily PnL before reset, got %.2f", metrics.DailyPnL)
	}
	
	// Reset daily metrics
	gmc.ResetDailyMetrics()
	
	// Verify reset
	metrics = gmc.GetMetrics()
	if metrics.DailyPnL != 0.0 {
		t.Errorf("Expected 0.0 daily PnL after reset, got %.2f", metrics.DailyPnL)
	}
	
	// Total PnL should remain unchanged
	if metrics.TotalPnL != 150.0 {
		t.Errorf("Expected 150.0 total PnL after reset, got %.2f", metrics.TotalPnL)
	}
}

func TestMetricsPersistence(t *testing.T) {
	config := DefaultBaseConfiguration()
	tempDir := t.TempDir()
	metricsFile := filepath.Join(tempDir, "metrics.json")
	auditLogger, _ := NewAuditLogger(filepath.Join(tempDir, "audit"))
	defer auditLogger.Close()
	
	// Create first collector and record trade
	gmc1, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create first metrics collector: %v", err)
	}
	
	trade := TradeRecord{
		Timestamp:    time.Now().UTC(),
		StrategyName: "TEST",
		PnL:          100.0,
		IsWinning:    true,
	}
	gmc1.RecordTrade(trade)
	gmc1.persistMetrics()
	gmc1.Stop()
	
	// Create second collector (simulating restart)
	gmc2, err := NewGlobalMetricsCollector(config, metricsFile, auditLogger)
	if err != nil {
		t.Fatalf("Failed to create second metrics collector: %v", err)
	}
	defer gmc2.Stop()
	
	// Verify data persisted
	metrics := gmc2.GetMetrics()
	if metrics.TotalPositions != 1 {
		t.Errorf("Expected 1 total position after reload, got %d", metrics.TotalPositions)
	}
	if metrics.TotalPnL != 100.0 {
		t.Errorf("Expected 100.0 total PnL after reload, got %.2f", metrics.TotalPnL)
	}
	
	strategyMetrics, exists := gmc2.GetStrategyMetrics("TEST")
	if !exists {
		t.Fatal("Strategy metrics should exist after reload")
	}
	if strategyMetrics.PnL != 100.0 {
		t.Errorf("Expected 100.0 strategy PnL after reload, got %.2f", strategyMetrics.PnL)
	}
}
