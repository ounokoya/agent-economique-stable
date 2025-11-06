# Workflow 3: Intégration avec l'agent

**Version:** 0.1  
**Statut:** Spécification technique  
**Module:** Intégration données ↔ Agent économique

## Vue d'ensemble
Ce workflow finalise l'intégration du module de données avec les composants de l'agent économique, permettant l'exécution de la stratégie MACD/CCI/DMI.

## Composants principaux

### 7. Connecteur Kline Engine
**Fichier:** `internal/data/binance/kline_connector.go`

**Responsabilités:**
- Interface vers le Kline Engine de l'agent
- Alimentation des calculs d'indicateurs MACD/CCI/DMI
- Gestion des timeframes multiples
- Optimisation des requêtes de données

**Fonctions clés:**
```go
func NewKlineConnector(engine KlineEngine) *KlineConnector
func (c *KlineConnector) FeedTimeframe(tf string, data []Kline) error
func (c *KlineConnector) GetIndicatorSnapshot(symbol, tf string) (*IndicatorSnapshot, error)
func (c *KlineConnector) StartRealTimeFeeding() error
func (c *KlineConnector) GetAvailableIndicators() []string
```

### 8. Connecteur Tick Engine
**Fichier:** `internal/data/binance/tick_connector.go`

**Responsabilités:**
- Interface vers le Tick Engine de l'agent
- Alimentation de l'analyse microstructure
- Calculs order flow et volatilité réalisée
- Détection d'événements en temps réel

**Fonctions clés:**
```go
func NewTickConnector(engine TickEngine) *TickConnector
func (c *TickConnector) FeedTrades(trades []Trade) error
func (c *TickConnector) GetTickAnalytics(symbol string) (*TickAnalytics, error)
func (c *TickConnector) EnableEventDetection(events []string) error
func (c *TickConnector) GetVolatilityMetrics() *VolatilityMetrics
```

### 9. Gestionnaire de contexte
**Fichier:** `internal/data/binance/context_manager.go`

**Responsabilités:**
- Fusion des données pour contexte unifié
- Coordination avec l'Agrégateur de Contexte
- Versioning et horodatage des contextes
- Optimisation des accès de données

**Fonctions clés:**
```go
func NewContextManager(aggregator ContextAggregator) *ContextManager
func (m *ContextManager) GenerateContext(timestamp int64) (*Context, error)
func (m *ContextManager) GetContextHistory(hours int) ([]Context, error)
func (m *ContextManager) ValidateContextIntegrity(ctx *Context) error
func (m *ContextManager) OptimizeDataAccess(strategy string) error
```

## Flux d'exécution

### Phase 1: Initialisation connecteurs
1. **Configuration interfaces:**
   - Établissement connexions avec Kline Engine
   - Établissement connexions avec Tick Engine
   - Configuration de l'Agrégateur de Contexte
   - Validation des contrats d'interface

2. **Préparation données:**
   - Chargement index des données disponibles
   - Préparation des buffers de streaming
   - Configuration des timeframes actifs

### Phase 2: Alimentation engines
1. **Flux Kline Engine:**
   ```
   Données ZIP → Parser Klines → Buffer Multi-TF → Kline Engine
                                                  ↓
   MACD(12,26,9) ← CCI(14) ← DMI(14) ← Indicateurs calculés
   ```

2. **Flux Tick Engine:**
   ```
   Données ZIP → Parser Trades → Agrégation → Tick Engine
                                            ↓
   Microstructure ← Order Flow ← Volatilité ← Analytics calculées
   ```

### Phase 3: Génération contexte unifié
1. **Fusion données:**
   - Synchronisation IndicatorSnapshot + TickAnalytics
   - Ajout des signaux de la stratégie
   - Intégration métadonnées d'environnement
   - Versioning et horodatage

2. **Validation et distribution:**
   - Validation cohérence du contexte
   - Distribution vers Money Management
   - Archivage pour diagnostics
   - Métriques de performance

## Intégration stratégie MACD/CCI/DMI

### Configuration indicateurs
```yaml
indicators:
  macd:
    fast_period: 12
    slow_period: 26
    signal_period: 9
    timeframes: ["5m", "15m", "1h", "4h"]
  
  cci:
    period: 14
    timeframes: ["5m", "15m", "1h", "4h"]
    thresholds:
      long_trend_oversold: -100
      long_trend_overbought: 100
      short_trend_oversold: -120
      short_trend_overbought: 120
  
  dmi:
    period: 14
    adx_period: 14
    timeframes: ["5m", "15m", "1h", "4h"]
```

### Signaux générés
```go
type StrategySignal struct {
    Timestamp   int64
    Symbol      string
    Timeframe   string
    SignalType  string  // "LONG_ENTRY", "SHORT_ENTRY", "EXIT"
    Confidence  float64
    Indicators  map[string]float64
    Context     *MarketContext
}
```

## Interface avec l'agent

### Context structure étendue
```go
type Context struct {
    // Données de base
    Timestamp    int64
    Symbol       string
    Environment  string
    
    // Indicateurs techniques
    Indicators   map[string]*IndicatorSnapshot
    
    // Analytics temps réel
    TickAnalytics *TickAnalytics
    
    // Signaux stratégie
    Signals      []StrategySignal
    
    // Métadonnées
    DataQuality  *QualityMetrics
    Latency      *LatencyMetrics
    Version      string
}
```

## Tests d'intégration requis

- `TestKlineEngineIntegration`
- `TestTickEngineIntegration`
- `TestContextGenerationComplete`
- `TestStrategySignalGeneration`
- `TestMultiTimeframeCoherence`
- `TestEndToEndBacktest`
- `TestPerformanceConstraints`

## Configuration complète

```yaml
integration:
  # Connecteurs
  kline_engine:
    buffer_size: 1000
    calculation_lag_ms: 100
    indicators: ["macd", "cci", "dmi"]
  
  tick_engine:
    aggregation_window_ms: 1000
    event_detection: true
    volatility_window_minutes: 15
  
  context_manager:
    versioning: true
    quality_validation: true
    archival_hours: 168  # 1 semaine
  
  # Performance
  performance:
    max_latency_ms: 500
    min_data_quality: 0.95
    context_generation_timeout_ms: 1000
```

## Critères d'acceptation

✅ Intégration complète avec Kline Engine  
✅ Intégration complète avec Tick Engine  
✅ Génération contexte unifié versionné  
✅ Calculs indicateurs MACD/CCI/DMI fonctionnels  
✅ Signaux stratégie générés correctement  
✅ Performance respectée (< 500ms latence)  
✅ Tests d'intégration end-to-end réussis  
✅ Qualité données validée (> 95%)
