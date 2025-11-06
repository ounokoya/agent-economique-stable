# Plan de Tests - Engine Temporel (Organisation White Box)

**Version:** 1.1 - **TESTS UNITAIRES DANS MODULE**  
**Organisation:** Tests colocalisés avec code source (`internal/engine/*_test.go`)  
**Coverage:** `go test -cover ./internal/engine` (39.1% atteint)  
**Contraintes:** Go standards, fichiers < 500 lignes

## Tests Unitaires

### TemporalEngine Tests (implémenté)
```go
// internal/engine/temporal_engine_test.go
func TestTemporalEngine_CalculateIndicators(t *testing.T)       ✅ Implémenté
func TestTemporalEngine_GetPositionContext(t *testing.T)        ✅ Implémenté
func TestTemporalEngine_ProcessStrategySignals(t *testing.T)    ✅ Implémenté
func TestTemporalEngine_GetCandleWindow(t *testing.T)           ✅ Implémenté
func TestTemporalEngine_ConvertCCIZone(t *testing.T)            ✅ Implémenté
func TestTemporalEngine_GetCurrentPrice(t *testing.T)           ✅ Implémenté
```

### PositionManager Tests (existant)
```go
// internal/engine/position_manager_test.go
func TestNewPositionManager(t *testing.T)                      ✅ Existant
func TestPositionManager_OpenPosition(t *testing.T)            ✅ Existant
func TestPositionManager_ClosePosition(t *testing.T)           ✅ Existant
func TestPositionManager_UpdateTrailingStop(t *testing.T)      ✅ Existant
func TestValidateStopLevel(t *testing.T)                       ✅ Existant
func TestCalculateRisk(t *testing.T)                           ✅ Existant
```

### ZoneMonitor Tests  
```go
// internal/engine/zone_monitor_test.go
func TestZoneMonitor_ActivateZone(t *testing.T)
func TestZoneMonitor_DeactivateZone(t *testing.T)
func TestZoneMonitor_CheckZoneConditions(t *testing.T)
func TestZoneMonitor_ApplyAdjustmentGrid(t *testing.T)
```

## Tests d'Intégration

### Cycles Backtest avec données fixes
```go
func TestBacktestCycle_Complete(t *testing.T) {
    // Chargement dataset fixe
    dataset, err := testdata.LoadDataset("engine_temporal/basic_cycle")
    require.NoError(t, err)
    
    expected, err := testdata.LoadExpected("engine_temporal/basic_cycle")
    require.NoError(t, err)
    
    engine := NewTemporalEngine(BacktestMode)
    
    // Simulation avec 150 trades chronologiques
    for _, trade := range dataset.Data.Trades {
        engine.ProcessTrade(trade)
    }
    
    // Validation vs résultats pré-calculés
    assert.Equal(t, expected.EngineImporal.CyclesExecuted, 150)
    assert.Equal(t, expected.EngineImporal.MarkersDetected, 6)
    assert.Equal(t, expected.EngineImporal.AntiLookaheadViolations, 0)
}
```

### Cycles Live/Paper simulation
```go
func TestLiveCycle_10SecondLoop(t *testing.T) {
    // Dataset simulation temps réel
    dataset, err := testdata.LoadDataset("engine_temporal/live_simulation")
    require.NoError(t, err)
    
    engine := NewTemporalEngine(LiveMode)
    
    // Simulation timer 10s avec mock time.Now()
    mockTimer := &MockTimer{}
    engine.SetTimer(mockTimer)
    
    // Test détection nouvelles bougies
    for _, event := range dataset.Data.TimerEvents {
        mockTimer.Advance(10 * time.Second)
        engine.ProcessTimerEvent()
    }
    
    // Validation cohérence vs mode backtest
    assert.Equal(t, expected.DecisionCount, engine.GetDecisionCount())
}
```

## Tests Performance

### Métriques cibles
- **Latence cycle:** < 50ms (backtest), < 200ms (live)
- **Mémoire:** < 500MB sustained
- **Throughput:** > 1000 trades/sec

### Benchmarks avec données fixes
```go
func BenchmarkTemporalEngine_ProcessTrade(b *testing.B) {
    dataset, _ := testdata.LoadDataset("engine_temporal/performance_stress")
    engine := NewTemporalEngine()
    
    trades := dataset.Data.Trades
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        for _, trade := range trades {
            engine.ProcessTrade(trade)
        }
        engine.Reset() // Reset pour iteration suivante
    }
}

func BenchmarkPositionManager_UpdateStop(b *testing.B) {
    dataset, _ := testdata.LoadDataset("engine_temporal/position_long_complete")
    positionManager := NewPositionManager()
    
    // Pre-load position ouverte
    positionManager.OpenPosition(LONG, 25.50, time.Now().UnixMilli())
    
    trades := dataset.Data.Trades
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        for _, trade := range trades {
            positionManager.UpdateTrailingStop(trade.Price)
        }
    }
}

func BenchmarkZoneMonitor_CheckAll(b *testing.B) {
    dataset, _ := testdata.LoadDataset("engine_temporal/cci_zone_activation")
    zoneMonitor := NewZoneMonitor()
    
    // Activer toutes les zones pour test stress
    zoneMonitor.ActivateZone(CCI_INVERSE, time.Now().UnixMilli())
    zoneMonitor.ActivateZone(MACD_INVERSE, time.Now().UnixMilli())
    
    trades := dataset.Data.Trades
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        for _, trade := range trades {
            zoneMonitor.CheckActiveZones(trade.Price, 5.0) // 5% profit
        }
    }
}

## Datasets TestData

### Référence : `testdata_specification.md`

### Datasets Engine Temporal disponibles
```yaml
basic_cycle:
  path: "engine_temporal/basic_cycle"
  scenario: "Signal LONG → Position → Fermeture standard"
  size: ~15KB, duration: 30min
  
position_long_complete:
  path: "engine_temporal/position_long_complete"  
  scenario: "LONG avec CCI inverse + ajustements stops"
  size: ~25KB, duration: 45min
  
cci_zone_activation:
  path: "engine_temporal/cci_zone_activation"
  scenario: "Zone CCI monitoring continu"
  size: ~10KB, duration: 20min
  
anti_lookahead_test:
  path: "engine_temporal/anti_lookahead_test"
  scenario: "Validation temporelle avec pièges futur"
  size: ~5KB, expected_violations: 0
  
performance_stress:
  path: "engine_temporal/performance_stress"
  scenario: "Volume réaliste pour benchmarks"
  size: ~200KB, duration: 4h
```

### Tests par dataset
```go
// Test anti-look-ahead avec piège
func TestAntiLookAhead_WithTraps(t *testing.T) {
    dataset, _ := testdata.LoadDataset("engine_temporal/anti_lookahead_test")
    expected, _ := testdata.LoadExpected("engine_temporal/anti_lookahead_test")
    
    engine := NewTemporalEngine()
    violations := 0
    
    // Dataset contient trade "piège" dans futur
    for _, trade := range dataset.Data.Trades {
        if err := engine.ProcessTrade(trade); err != nil {
            if errors.Is(err, ErrLookAheadDetected) {
                violations++
            }
        }
    }
    
    // Doit détecter violations = résistance anti-look-ahead
    assert.Equal(t, expected.EngineImporal.AntiLookaheadViolations, violations)
}

// Test zones CCI actives
func TestZoneMonitoring_CCIInverse(t *testing.T) {
    dataset, _ := testdata.LoadDataset("engine_temporal/cci_zone_activation") 
    expected, _ := testdata.LoadExpected("engine_temporal/cci_zone_activation")
    
    engine := NewTemporalEngine()
    
    // Simulation complète avec zone CCI inverse
    for _, trade := range dataset.Data.Trades {
        engine.ProcessTrade(trade)
    }
    
    // Validation zones activées et ajustements
    assert.Contains(t, expected.EngineImporal.ZonesActivated, "CCI_INVERSE")
    assert.Greater(t, engine.GetStopAdjustmentsCount(), 0)
}
```

---

*Tests Engine Temporel : Validation complète gestion temporelle et positions*
