# Plan de Tests - Module Indicateurs (Organisation White Box)

**Version:** 1.2 - **TESTS UNITAIRES DANS MODULE**  
**Organisation:** Tests colocalisés avec code source (`internal/indicators/*_test.go`)  
**Coverage:** `go test -cover ./internal/indicators` (47.0% atteint)  
**Précision:** < 0.001% vs TradingView (indicateurs existants déjà testés)

## Tests Unitaires (focus nouveaux composants)

### Interface Communication (implémenté)
```go
// internal/indicators/calculator_test.go
func TestCalculate_BasicFunctionality(t *testing.T)        ✅ Implémenté
func TestCalculate_EmptyKlines(t *testing.T)              ✅ Implémenté
func TestDetectMACDCrossover_CrossUp(t *testing.T)        ✅ Implémenté

// Test interface communication sans redondance Engine
func TestCalculate_NoValidationDuplication(t *testing.T) {
    // Engine fournit déjà données propres
    klines := testdata.LoadValidKlines(300) // Déjà >= 35, déjà validées
    
    request := &CalculationRequest{
        Symbol: "SOLUSDT",
        CandleWindow: klines,  // Engine a déjà validé
        PositionContext: &PositionContext{IsOpen: true, Direction: PositionLong},
    }
    
    // Module ne fait PAS de validation (Engine l'a déjà fait)
    response := Calculate(request)
    
    assert.True(t, response.Success)
    assert.NotNil(t, response.Results)
    assert.NotEmpty(t, response.Signals) // Si conditions remplies
    // Pas de test validation car Engine s'en charge déjà
}
```

### Signal Generator Stratégie (implémenté)
```go
// internal/indicators/signal_generator_test.go
func TestGenerateStrategySignals_LongTrendSignal(t *testing.T)     ✅ Implémenté
func TestGenerateStrategySignals_ShortTrendSignal(t *testing.T)    ✅ Implémenté
func TestCalculateConfidence_HighConfidence(t *testing.T)          ✅ Implémenté

// internal/indicators/zone_detector_test.go  
func TestDetectZoneEvents_CCIInverse_Long(t *testing.T)            ✅ Implémenté
func TestDetectZoneEvents_MACDInverse_WithProfit(t *testing.T)     ✅ Implémenté

// Test signal generator conforme mémoire utilisateur
func TestGenerateStrategySignals_MemoryCompliance(t *testing.T) {
    results := &IndicatorResults{
        MACD: &MACDValues{CrossoverType: CrossUp},
        CCI: &CCIValues{Zone: CCIOversold},  
        DMI: &DMIValues{PlusDI: 25, MinusDI: 15}, // DI+ > DI-
    }
    
    signals := generateStrategySignals(results)
    
    // Signal LONG tendance selon mémoire utilisateur
    require.Len(t, signals, 1)
    assert.Equal(t, LongSignal, signals[0].Direction)
    assert.Equal(t, TrendSignal, signals[0].Type)
}

// Test détection événements zones (utilise types Engine)
func TestDetectZoneEvents_UsesEngineTypes(t *testing.T) {
    positionCtx := &PositionContext{
        IsOpen: true,
        Direction: PositionLong,
        EntryCCIZone: CCIOversold,  // Engine type existant
    }
    
    results := &IndicatorResults{
        CCI: &CCIValues{Value: 120}, // Maintenant surachat (inverse)
    }
    
    events := detectZoneEvents(results, positionCtx)
    
    // Utilise ZoneEvent et ZoneType Engine existants
    require.Len(t, events, 1)
    assert.Equal(t, ZoneCCIInverse, events[0].ZoneType) // Engine type
}
```

### DMI Calculator
```go
// internal/indicators/dmi_test.go
func TestDMICalculator_Calculate(t *testing.T)
func TestDMICalculator_TrendAnalysis(t *testing.T)
func TestDMICalculator_WildersSmoothing(t *testing.T)

// Test croisements DI
func TestDMI_DetectCrossings(t *testing.T) {
    // DI+ passe au-dessus DI-
}
```

## Tests Précision avec références externes

### Datasets de validation disponibles
```yaml
macd_precision:
  path: "indicators/macd_precision"
  scenario: "300 candles SOLUSDT vs TradingView"
  size: ~50KB
  tolerance: 0.00001 (< 0.001%)
  
cci_zones:
  path: "indicators/cci_zones"
  scenarios: ["trend_long", "counter_short", "multi_thresholds"]
  size: ~30KB
  
dmi_trend_analysis:
  path: "indicators/dmi_trend_analysis"
  scenario: "DMI/ADX vs MetaTrader référence"
  size: ~40KB
```

### Tests validation croisée
```go
func TestIndicators_PrecisionTradingView(t *testing.T) {
    dataset, _ := testdata.LoadDataset("indicators/macd_precision")
    expected, _ := testdata.LoadExpected("indicators/macd_precision")
    
    // Tous indicateurs sur mêmes données
    macd := CalculateMACD(dataset.Data.Klines)
    cci := CalculateCCI(dataset.Data.Klines)
    dmi := CalculateDMI(dataset.Data.Klines)
    
    // Validation vs TradingView pré-calculés
    assert.InDelta(t, expected.TradingViewReference.MACD, macd.FinalValue, 0.00001)
    assert.InDelta(t, expected.TradingViewReference.CCI, cci.FinalValue, 0.00001)
    assert.InDelta(t, expected.TradingViewReference.ADX, dmi.ADX, 0.00001)
}

func TestIndicators_PrecisionMetaTrader(t *testing.T) {
    dataset, _ := testdata.LoadDataset("indicators/dmi_trend_analysis")
    expected, _ := testdata.LoadExpected("indicators/dmi_trend_analysis")
    
    dmi := CalculateDMI(dataset.Data.Klines)
    
    // Validation croisée MT4/MT5 DMI
    assert.InDelta(t, expected.MetaTraderReference.DIPlus, dmi.DIPlus, 0.00001)
    assert.InDelta(t, expected.MetaTraderReference.DIMinus, dmi.DIMinus, 0.00001)
}
```

## Tests Signal Generation avec scénarios

### Datasets signaux disponibles
```yaml
signal_generation:
  path: "indicators/signal_generation"
  scenarios:
    long_perfect: "MACD↗ + CCI oversold + DI+ > DI-"
    short_rejected: "Conditions partielles + filtres échouent"
    confidence_scores: "Calcul pénalités scores"
  size: ~40KB
```

### Tests par scénario
```go
func TestSignalGenerator_LongPerfectSignal(t *testing.T) {
    dataset, _ := testdata.LoadDataset("indicators/signal_generation")
    expected, _ := testdata.LoadExpected("indicators/signal_generation")
    
    // Charger scénario LONG parfait
    scenario := dataset.Scenarios["long_perfect"]
    
    macd := CalculateMACD(scenario.Klines)
    cci := CalculateCCI(scenario.Klines)
    dmi := CalculateDMI(scenario.Klines)
    
    // Génération signal
    signals := GenerateSignals(macd, cci, dmi, DefaultSignalConfig)
    
    // Validation signal LONG généré
    require.Len(t, signals, 1)
    signal := signals[0]
    assert.Equal(t, LONG_ENTRY, signal.Type)
    assert.GreaterOrEqual(t, signal.Confidence, 80.0)
    assert.Contains(t, signal.TriggerReason, "MACD cross up")
}

func TestSignalGenerator_FilterRejection(t *testing.T) {
    dataset, _ := testdata.LoadDataset("indicators/signal_generation")
    expected, _ := testdata.LoadExpected("indicators/signal_generation")
    
    // Scénario avec filtres qui échouent
    scenario := dataset.Scenarios["short_rejected"]
    
    macd := CalculateMACD(scenario.Klines)
    cci := CalculateCCI(scenario.Klines)
    dmi := CalculateDMI(scenario.Klines)
    
    // Config avec filtres stricts
    config := SignalConfig{
        MACDSameSignFilter: true,
        DXADXFilterEnabled: true,
    }
    
    signals := GenerateSignals(macd, cci, dmi, config)
    
    // Aucun signal généré (rejeté par filtres)
    assert.Len(t, signals, 0)
}
```

### Score de confiance
```go
func TestSignalGenerator_ConfidenceScore(t *testing.T) {
    // Base 100% - pénalités
    // MACD faible: -10%
    // CCI proche seuil: -15%
    // ADX faible: -20%
}
```

## Tests Performance

### Benchmarks avec données fixes
```go
func BenchmarkMACD_300Candles(b *testing.B) {
    dataset, _ := testdata.LoadDataset("indicators/macd_precision")
    klines := dataset.Data.Klines
    config := MACDConfig{12, 26, 9}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        CalculateMACD(klines, config)
    }
    // Target: < 15ms per iteration
}

func BenchmarkCCI_300Candles(b *testing.B) {
    dataset, _ := testdata.LoadDataset("indicators/cci_zones")
    klines := dataset.Data.Klines
    config := CCIConfig{14}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        CalculateCCI(klines, config)
    }
    // Target: < 10ms per iteration
}

func BenchmarkDMI_300Candles(b *testing.B) {
    dataset, _ := testdata.LoadDataset("indicators/dmi_trend_analysis")
    klines := dataset.Data.Klines
    config := DMIConfig{14, 14}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        CalculateDMI(klines, config)
    }
    // Target: < 15ms per iteration
}

func BenchmarkSignalGeneration_Complete(b *testing.B) {
    dataset, _ := testdata.LoadDataset("indicators/signal_generation")
    klines := dataset.Data.Klines
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        macd := CalculateMACD(klines, MACDConfig{12, 26, 9})
        cci := CalculateCCI(klines, CCIConfig{14})
        dmi := CalculateDMI(klines, DMIConfig{14, 14})
        GenerateSignals(macd, cci, dmi, DefaultSignalConfig)
    }
    // Target: < 50ms pipeline complète
}

## Tests Événements Zones

### CCI Zone Events
```go
func TestZoneEvents_CCIInverse(t *testing.T) {
    // Position LONG + CCI overbought → ZoneEvent
}

func TestZoneEvents_MACDInverse(t *testing.T) {
    // MACD inverse + profit > seuil → Event
}
```

## Référence TestData

### Voir `testdata_specification.md` pour structure complète

### Chargement standardisé
```go
// Utilitaires tests indicateurs
package indicators_test

import "internal/testdata"

func setupIndicatorTest(datasetPath string) (*testdata.Dataset, *testdata.ExpectedResults, error) {
    dataset, err := testdata.LoadDataset(datasetPath)
    if err != nil {
        return nil, nil, err
    }
    
    expected, err := testdata.LoadExpected(datasetPath)
    if err != nil {
        return nil, nil, err
    }
    
    return dataset, expected, nil
}

// Helper extraction klines
func extractKlines(dataset *testdata.Dataset) []Kline {
    return dataset.Data.Klines
}

// Helper validation précision
func assertPrecision(t *testing.T, expected, actual float64, tolerance float64) {
    assert.InDelta(t, expected, actual, tolerance, 
        "Precision error: expected %f, got %f, tolerance %f", 
        expected, actual, tolerance)
}
```

### Edge cases intégrés dans datasets
```yaml
Datasets incluent naturellement :
- Gaps temporels (weekends crypto)
- Volatilité extrême (pumps/dumps)
- Volumes variables (pics et creux)
- Périodes de consolidation
- Tendances fortes et corrections
```

---

*Tests Indicateurs : Validation précision calculs et génération signaux*
