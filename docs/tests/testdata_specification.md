# Spécification TestData - Données Fixes de Test

**Version:** 1.0  
**Objectif:** Données fixes reproductibles pour tests rapides et fiables  
**Source:** Extraction cache Binance → JSON structurés  

## Vue d'ensemble

Cette spécification définit l'organisation et la gestion des données de test fixes stockées dans `testdata/`. L'approche remplace l'utilisation directe du cache Binance par des datasets JSON pré-extraits, garantissant reproductibilité, performance et indépendance.

## Principes fondamentaux

### Avantages données fixes
- **Performance** : 10-100x plus rapide que lecture cache
- **Reproductibilité** : Exactement mêmes données à chaque test
- **Indépendance** : Tests fonctionnent sans infrastructure externe
- **Contrôle** : Scénarios précis et edge cases construits

### Contraintes respectées
- **Go standards** : Fichiers JSON < 500KB, structure modulaire
- **Versionning** : Données versionnées avec code Git
- **Maintenance** : Régénération simple depuis cache source

## Structure testdata/

### Organisation par module
```
testdata/
├── engine_temporal/              # Tests Engine Temporel
│   ├── basic_cycle/
│   │   ├── dataset.json          # Données + métadonnées
│   │   └── expected.json         # Résultats attendus
│   ├── position_long_complete/   # Cycle position LONG complet
│   ├── cci_zone_activation/      # Tests zones actives
│   ├── anti_lookahead_test/      # Validation temporelle
│   └── performance_stress/       # Benchmarks
├── indicators/                   # Tests Calculs Indicateurs
│   ├── macd_precision/          # Validation vs TradingView
│   ├── cci_zones/               # Tests seuils par type signal
│   ├── dmi_trend_analysis/      # DMI/ADX calculs
│   └── signal_generation/       # Tests signaux LONG/SHORT
├── integration/                  # Tests End-to-End
│   ├── strategy_complete/       # Stratégie complète 24h
│   ├── multi_timeframe_sync/    # Synchronisation TF
│   └── error_scenarios/         # Gestion erreurs
└── tools/                       # Utilitaires génération
    ├── extract_testdata.go      # Extraction depuis cache
    ├── validate_datasets.go     # Validation intégrité
    └── generate_expected.go     # Calcul résultats attendus
```

## Format standardisé datasets

### Structure JSON principale
```json
{
  "metadata": {
    "name": "basic_cycle_solusdt",
    "description": "Cycle complet : signal LONG → position → fermeture",
    "version": "1.0",
    "created_at": "2024-01-01T00:00:00Z",
    "source": {
      "symbol": "SOLUSDT",
      "timeframe": "5m", 
      "period": "2023-06-01T10:00:00Z to 2023-06-01T10:30:00Z",
      "extracted_from": "data/binance/SOLUSDT/klines/5m/SOLUSDT-5m-2023-06-01.zip"
    },
    "scenario": {
      "type": "engine_temporal_basic",
      "expected_signals": 1,
      "expected_positions": 1,
      "market_condition": "trending_up_with_correction"
    }
  },
  "data": {
    "klines": [
      {
        "timestamp": 1685620800000,
        "open": 25.10,
        "high": 25.50,
        "low": 25.00,
        "close": 25.30,
        "volume": 145678.50
      }
    ],
    "trades": [
      {
        "timestamp": 1685620825000,
        "price": 25.15,
        "quantity": 120.50,
        "is_buyer_maker": false
      }
    ]
  }
}
```

### Fichier résultats attendus
```json
{
  "metadata": {
    "dataset": "basic_cycle_solusdt",
    "version": "1.0",
    "calculated_at": "2024-01-01T00:00:00Z"
  },
  "engine_temporal": {
    "cycles_executed": 150,
    "markers_detected": 6,
    "positions_opened": 1,
    "positions_closed": 1,
    "zones_activated": ["CCI_INVERSE"],
    "anti_lookahead_violations": 0
  },
  "indicators": {
    "macd": {
      "final_value": 0.1234,
      "final_signal": 0.0987,
      "crossovers_detected": 1,
      "precision_vs_tradingview": 0.00001
    },
    "cci": {
      "final_value": -85.67,
      "zone_transitions": ["NORMAL", "OVERSOLD", "OVERBOUGHT"],
      "zone_events_generated": 2
    }
  },
  "performance": {
    "total_pnl_percent": 2.34,
    "max_drawdown_percent": 0.12,
    "trades_count": 1,
    "stop_adjustments": 2
  }
}
```

## Datasets par catégorie

### Engine Temporal
```yaml
basic_cycle:
  size: ~15KB
  duration: 30min
  candles: 6
  trades: ~150
  scenario: "Signal → Position → Fermeture standard"
  
position_long_complete:
  size: ~25KB  
  duration: 45min
  scenario: "LONG avec CCI inverse + ajustements stops"
  
cci_zone_activation:
  size: ~10KB
  duration: 20min
  scenario: "Zone CCI monitoring continu"
  
anti_lookahead_test:
  size: ~5KB
  scenario: "Données avec pièges futur → 0 violation"
  
performance_stress:
  size: ~200KB
  duration: 4h
  scenario: "Volume réaliste pour benchmarks"
```

### Indicators
```yaml
macd_precision:
  size: ~50KB
  candles: 300
  scenario: "Validation vs TradingView (< 0.001% erreur)"
  
cci_zones:
  size: ~30KB
  scenarios: ["trend_long", "counter_short", "multi_thresholds"]
  
signal_generation:
  size: ~40KB
  scenarios: ["long_perfect", "short_rejected", "confidence_scores"]
```

### Integration  
```yaml
strategy_complete:
  size: ~150KB
  duration: 24h
  scenario: "Multiple signaux + performance complète"
  
multi_timeframe_sync:
  size: ~80KB
  timeframes: ["5m", "15m", "1h"]
  scenario: "Synchronisation marqueurs 10:00:00"
```

## Processus extraction

### Commande génération
```bash
# Extraction dataset spécifique
go run tools/extract_testdata.go \
  --source "data/binance/SOLUSDT" \
  --scenario "basic_cycle" \
  --start "2023-06-01T10:00:00Z" \
  --duration "30m" \
  --output "testdata/engine_temporal/basic_cycle"

# Génération résultats attendus
go run tools/generate_expected.go \
  --dataset "testdata/engine_temporal/basic_cycle/dataset.json" \
  --output "testdata/engine_temporal/basic_cycle/expected.json"
```

### Validation intégrité
```bash
# Validation tous datasets
go run tools/validate_datasets.go --all

# Validation dataset spécifique  
go run tools/validate_datasets.go --dataset "engine_temporal/basic_cycle"
```

## API d'accès dans tests

### Chargement dataset
```go
// internal/testdata/loader.go
package testdata

type Dataset struct {
    Metadata DatasetMetadata `json:"metadata"`
    Data     DatasetData     `json:"data"`
}

func LoadDataset(path string) (*Dataset, error) {
    fullPath := filepath.Join("testdata", path, "dataset.json")
    return loadDatasetFromFile(fullPath)
}

func LoadExpected(path string) (*ExpectedResults, error) {
    fullPath := filepath.Join("testdata", path, "expected.json")
    return loadExpectedFromFile(fullPath)
}
```

### Utilisation dans tests
```go
func TestTemporalEngine_BasicCycle(t *testing.T) {
    // Chargement instantané
    dataset, err := testdata.LoadDataset("engine_temporal/basic_cycle")
    require.NoError(t, err)
    
    expected, err := testdata.LoadExpected("engine_temporal/basic_cycle")
    require.NoError(t, err)
    
    // Test avec données fixes
    engine := NewTemporalEngine()
    for _, trade := range dataset.Data.Trades {
        engine.ProcessTrade(trade)
    }
    
    // Validation vs résultats attendus
    assert.Equal(t, expected.Enginetemporal.PositionsOpened, engine.GetPositionsCount())
}
```

## Maintenance et évolution

### Régénération datasets
```bash
# Si stratégie évolue, régénérer expected results
make regenerate-testdata

# Si nouveaux scénarios requis
go run tools/extract_testdata.go --scenario new_indicator_rsi
```

### Versionning
- **Datasets** versionnés avec code (Git LFS si >100KB)
- **Expected results** recalculés si logique métier change
- **Backward compatibility** maintenue avec metadata.version

### CI/CD Integration
```yaml
# .github/workflows/tests.yml
- name: Validate testdata integrity
  run: go run tools/validate_datasets.go --all --strict
  
- name: Run tests with fixed data
  run: go test ./... -v -short # Utilise testdata/ automatiquement
```

## Bonnes pratiques

### Taille datasets
- **Tests unitaires** : < 20KB (rapide)
- **Tests intégration** : < 200KB (réaliste)  
- **Benchmarks** : < 1MB (performance)

### Nomenclature
```
Format : {module}_{scenario}_{symbol}
Exemples :
- engine_basic_cycle_solusdt
- indicators_macd_precision_ethusdt  
- integration_strategy_complete_multiasset
```

### Documentation inline
```json
{
  "metadata": {
    "description": "Description claire du scénario testé",
    "scenario": {
      "setup": "Conditions initiales",
      "actions": "Actions déclenchées", 
      "expectations": "Résultats attendus"
    }
  }
}
```

---

*Version 1.0 - TestData : Données fixes performantes et reproductibles pour tous tests*
