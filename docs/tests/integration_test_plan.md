# Plan de Tests - Intégration End-to-End

**Version:** 1.0  
**Scope:** Engine ↔ Indicateurs intégration complète  
**Performance:** < 200ms end-to-end

## Tests d'Intégration

### Communication Engine-Indicateurs
```go
// internal/integration/communication_test.go
func TestEngineIndicators_RequestResponse(t *testing.T) {
    // CalculationRequest → CalculationResponse
    // Validation structures complètes
}

func TestEngineIndicators_ErrorHandling(t *testing.T) {
    // Gestion timeouts, données insuffisantes
    // Recovery strategies
}
```

### Workflow Complet avec datasets fixes
```go
func TestStrategy_CompleteLongExecution(t *testing.T) {
    // Chargement dataset complet 24h
    dataset, err := testdata.LoadDataset("integration/strategy_complete")
    require.NoError(t, err)
    
    expected, err := testdata.LoadExpected("integration/strategy_complete")
    require.NoError(t, err)
    
    // Initialisation système intégré
    engine := NewTemporalEngine(BacktestMode)
    indicators := NewIndicatorCalculator()
    integration := NewEngineIndicatorsIntegration(engine, indicators)
    
    // Exécution stratégie complète
    for _, trade := range dataset.Data.Trades {
        integration.ProcessTrade(trade)
    }
    
    // Validation résultats end-to-end
    performance := integration.GetPerformanceMetrics()
    assert.Equal(t, expected.Performance.TotalPnlPercent, performance.TotalPnL)
    assert.Equal(t, expected.Performance.TradesCount, performance.TradesExecuted)
    assert.Len(t, performance.ZoneAdjustments, expected.Performance.StopAdjustments)
}

func TestStrategy_MultiSymbolParallel(t *testing.T) {
    // Dataset multi-symboles simultanés
    dataset, _ := testdata.LoadDataset("integration/multi_symbol_parallel")
    expected, _ := testdata.LoadExpected("integration/multi_symbol_parallel")
    
    integration := NewMultiSymbolIntegration(["SOLUSDT", "SUIUSDT", "ETHUSDT"])
    
    // Exécution parallèle
    results := integration.ProcessParallel(dataset.Data.MultiSymbolTrades)
    
    // Validation pas d'interférence entre symboles
    for symbol, result := range results {
        expectedResult := expected.MultiSymbol[symbol]
        assert.Equal(t, expectedResult.SignalsGenerated, result.SignalsCount)
        assert.Equal(t, expectedResult.PositionsOpened, result.PositionsCount)
    }
}
```

### Performance End-to-End avec métriques
```go
func TestPerformance_BacktestComplete(t *testing.T) {
    // Dataset performance réaliste 30 jours
    dataset, _ := testdata.LoadDataset("integration/performance_30days")
    expected, _ := testdata.LoadExpected("integration/performance_30days")
    
    integration := NewFullIntegration()
    
    // Métriques de base
    startTime := time.Now()
    startMemory := GetMemoryUsage()
    
    // Exécution 30 jours de trading
    for _, trade := range dataset.Data.Trades { // ~1.3M trades
        cycleStart := time.Now()
        
        integration.ProcessTrade(trade)
        
        cycleLatency := time.Since(cycleStart)
        assert.Less(t, cycleLatency.Milliseconds(), int64(200), 
            "Cycle latency exceeded 200ms: %v", cycleLatency)
    }
    
    // Validation performance globale
    totalTime := time.Since(startTime)
    finalMemory := GetMemoryUsage()
    
    assert.Less(t, finalMemory-startMemory, int64(500*1024*1024), 
        "Memory usage exceeded 500MB: %d bytes", finalMemory-startMemory)
    
    // Throughput: > 1000 trades/sec
    throughput := float64(len(dataset.Data.Trades)) / totalTime.Seconds()
    assert.Greater(t, throughput, 1000.0, 
        "Throughput too low: %f trades/sec", throughput)
}

func TestPerformance_LiveSimulation(t *testing.T) {
    // Dataset simulation 1 heure temps réel
    dataset, _ := testdata.LoadDataset("integration/live_1hour_simulation")
    
    integration := NewFullIntegration(LiveMode)
    mockTimer := &MockTimer{interval: 10 * time.Second}
    integration.SetTimer(mockTimer)
    
    // Simulation 1 heure = 360 cycles de 10s
    for i := 0; i < 360; i++ {
        cycleStart := time.Now()
        
        mockTimer.Tick()
        integration.ProcessTimerEvent()
        
        cycleTime := time.Since(cycleStart)
        assert.Less(t, cycleTime.Milliseconds(), int64(500), 
            "Live cycle exceeded 500ms: %v", cycleTime)
    }
    
    // Validation stabilité performance (pas de dérive)
    metrics := integration.GetPerformanceMetrics()
    assert.Less(t, metrics.AverageLatency.Milliseconds(), int64(200))
    assert.Less(t, metrics.MaxLatency.Milliseconds(), int64(800))
}
```

## Tests Scénarios Métier

### Cas de trading réels
```go
func TestScenario_TrendFollowing(t *testing.T) {
    // Tendance forte avec signaux multiples
    // Validation cohérence signaux
}

func TestScenario_SidewaysMarket(t *testing.T) {
    // Marché sans tendance
    // Peu de signaux générés
}

func TestScenario_VolatileMarket(t *testing.T) {
    // Forte volatilité
    // Stops ajustés fréquemment
}
```

### Gestion position complexe
```go
func TestComplexPosition_MultipleZones(t *testing.T) {
    // CCI inverse + MACD inverse simultanés
    // Ajustements stops multiples
}

func TestComplexPosition_EarlyExit(t *testing.T) {
    // Sortie anticipée MACD inverse
    // Avant trailing stop positif
}
```

## Tests Robustesse

### Conditions dégradées avec scénarios
```go
func TestRobustness_DataGaps(t *testing.T) {
    // Dataset avec gaps intentionnels
    dataset, _ := testdata.LoadDataset("integration/error_scenarios/data_gaps")
    expected, _ := testdata.LoadExpected("integration/error_scenarios/data_gaps")
    
    integration := NewRobustIntegration()
    
    // Dataset contient gap de 5 minutes entre trades
    errors := []error{}
    for _, trade := range dataset.Data.Trades {
        if err := integration.ProcessTrade(trade); err != nil {
            errors = append(errors, err)
        }
    }
    
    // Validation recovery gracieux
    assert.Len(t, errors, expected.Recovery.ExpectedErrors)
    assert.True(t, integration.IsOperational(), 
        "System should remain operational after data gaps")
}

func TestRobustness_CalculationTimeouts(t *testing.T) {
    dataset, _ := testdata.LoadDataset("integration/error_scenarios/timeouts")
    
    integration := NewFullIntegration()
    
    // Simuler timeouts calculs (mock)
    mockIndicators := &MockIndicatorCalculator{
        timeoutProbability: 0.1, // 10% des calculs timeout
        timeoutDuration: 6 * time.Second,
    }
    integration.SetIndicators(mockIndicators)
    
    successfulCycles := 0
    for _, trade := range dataset.Data.Trades {
        if err := integration.ProcessTrade(trade); err == nil {
            successfulCycles++
        }
    }
    
    // Même avec timeouts, majorité cycles réussissent (fallback)
    successRate := float64(successfulCycles) / float64(len(dataset.Data.Trades))
    assert.Greater(t, successRate, 0.85, 
        "Success rate too low with timeouts: %f", successRate)
}

func TestRobustness_MemoryPressure(t *testing.T) {
    dataset, _ := testdata.LoadDataset("integration/performance_stress")
    
    integration := NewFullIntegration()
    
    // Limiter mémoire artificiellement
    memoryLimit := int64(200 * 1024 * 1024) // 200MB seulement
    integration.SetMemoryLimit(memoryLimit)
    
    for _, trade := range dataset.Data.Trades {
        integration.ProcessTrade(trade)
        
        currentMemory := GetMemoryUsage()
        if currentMemory > memoryLimit {
            // Déclencher GC forcé
            runtime.GC()
            runtime.GC() // Double GC pour être sûr
            
            postGCMemory := GetMemoryUsage()
            assert.Less(t, postGCMemory, memoryLimit*120/100, // +20% tolérance
                "Memory not freed after GC: %d bytes", postGCMemory)
        }
    }
    
    // Validation système reste opérationnel
    assert.True(t, integration.IsOperational())
}
```

### Edge cases système
```go
func TestEdgeCases_ConfigurationChanges(t *testing.T) {
    // Changement config runtime
    // Invalidation cache appropriée
}
```

## Tests Validation Métier

### Conformité stratégie
```go
func TestStrategy_RulesCompliance(t *testing.T) {
    // 100% conformité strategy_macd_cci_dmi_pure.md
    // Validation chaque règle individuellement
}

func TestStrategy_SignalAccuracy(t *testing.T) {
    // Comparaison vs analyse manuelle
    // Pas de faux positifs/négatifs
}
```

### Anti-look-ahead validation
```go
func TestAntiLookAhead_ComprehensiveValidation(t *testing.T) {
    // Validation sur 1M+ trades
    // 0 violation détectée
}

// Benchmarks Intégrés avec données fixes

func BenchmarkStrategy_FullExecution(b *testing.B) {
    dataset, _ := testdata.LoadDataset("integration/strategy_complete")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        integration := NewFullIntegration()
        
        // Cycle complet Engine+Indicateurs sur 24h
        for _, trade := range dataset.Data.Trades {
            integration.ProcessTrade(trade)
        }
        
        integration.Shutdown()
    }
    // Target: < 200ms moyenne end-to-end
}

func BenchmarkStrategy_ParallelSymbols(b *testing.B) {
    dataset, _ := testdata.LoadDataset("integration/multi_symbol_parallel")

### Datasets intégration disponibles
```yaml
strategy_complete:
  path: "integration/strategy_complete"
  scenario: "Stratégie 24h avec signaux multiples"
  size: ~150KB
  expected_signals: 8
  expected_positions: 5
  expected_pnl: "+12.3%"
  
multi_symbol_parallel:
  path: "integration/multi_symbol_parallel"
  symbols: ["SOLUSDT", "SUIUSDT", "ETHUSDT"]
  scenario: "Exécution parallèle sans interférence"
  size: ~200KB per symbol
  
multi_timeframe_sync:
  path: "integration/multi_timeframe_sync"
  timeframes: ["5m", "15m", "1h"]
  scenario: "Synchronisation marqueurs 10:00:00"
  size: ~80KB
  
performance_30days:
  path: "integration/performance_30days"
  scenario: "Volume réaliste production"
  trades_count: ~1.3M
  size: ~2MB compressed
  
error_scenarios/*:
  paths: ["data_gaps", "timeouts", "memory_pressure"]
  scenario: "Gestion erreurs et recovery"
  sizes: ~50KB each
```

### API chargement intégration
```go
// Utilitaires tests intégration
package integration_test

import "internal/testdata"

func setupIntegrationTest(datasetPath string) (*FullIntegration, *testdata.Dataset, *testdata.ExpectedResults) {
    dataset, _ := testdata.LoadDataset(datasetPath)
    expected, _ := testdata.LoadExpected(datasetPath)
    
    integration := NewFullIntegration()
    integration.Configure(dataset.Metadata.Configuration)
    
    return integration, dataset, expected
}

// Helper validation performance
func assertPerformanceMetrics(t *testing.T, expected *testdata.PerformanceExpected, actual *PerformanceMetrics) {
    assert.InDelta(t, expected.TotalPnlPercent, actual.TotalPnL, 0.1)
    assert.Equal(t, expected.TradesCount, actual.TradesExecuted)
    assert.LessOrEqual(t, actual.AverageLatency.Milliseconds(), int64(200))
}
```

---

*Tests Intégration : Validation complète stratégie end-to-end*
