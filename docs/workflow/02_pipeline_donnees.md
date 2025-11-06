# Workflow 2: Pipeline de données

**Version:** 0.1  
**Statut:** Spécification technique  
**Module:** Traitement données Binance Vision

## Vue d'ensemble
Ce workflow définit le pipeline de traitement des données téléchargées pour alimenter les engines Kline et Tick de l'agent économique.

## Composants principaux

### 4. Parser Klines
**Fichier:** `internal/data/binance/klines.go`

**Responsabilités:**
- Extraction OHLCV depuis archives ZIP Binance
- Support multi-timeframes (5m, 15m, 1h, 4h)
- Parsing streaming sans full-load mémoire
- Validation format et cohérence temporelle

**Fonctions clés:**
```go
func NewKlineParser(timeframe string) *KlineParser
func (p *KlineParser) Parse(zipReader *ZipStreamReader) (*KlineIterator, error)
func (k *KlineIterator) Next() (*Kline, error)
func (k *KlineIterator) HasNext() bool
func ValidateKlineSequence(klines []Kline) error
```

**Structure Kline:**
```go
type Kline struct {
    Symbol    string
    Timeframe string
    OpenTime  int64
    CloseTime int64
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    float64
    Trades    int64
}
```

### 5. Parser Trades
**Fichier:** `internal/data/binance/trades.go`

**Responsabilités:**
- Extraction trades depuis archives ZIP
- Parsing pour microstructure et order flow
- Agrégation temps réel configurable
- Détection d'anomalies dans les trades

**Fonctions clés:**
```go
func NewTradeParser() *TradeParser
func (p *TradeParser) Parse(zipReader *ZipStreamReader) (*TradeIterator, error)
func (t *TradeIterator) Next() (*Trade, error)
func (t *TradeIterator) HasNext() bool
func AggregateTradesByTimeWindow(trades []Trade, windowMs int64) []TradeWindow
```

**Structure Trade:**
```go
type Trade struct {
    Symbol    string
    TradeID   int64
    Price     float64
    Quantity  float64
    Timestamp int64
    IsBuyerMaker bool
}
```

### 6. Intégrateur multi-timeframes
**Fichier:** `internal/data/binance/integrator.go`

**Responsabilités:**
- Synchronisation données multi-timeframes
- Agrégation cohérente sans décalage temporel
- Buffer glissant pour fenêtres d'analyse
- Génération de contexte unifié

**Fonctions clés:**
```go
func NewMultiTimeframeIntegrator(timeframes []string) *Integrator
func (i *Integrator) AddKlineData(timeframe string, klines []Kline) error
func (i *Integrator) AddTradeData(trades []Trade) error
func (i *Integrator) GetSynchronizedContext(timestamp int64) (*MarketContext, error)
func (i *Integrator) GetAvailableTimeRange() (start, end int64)
```

## Flux d'exécution

### Phase 1: Configuration parsers
1. Initialisation parsers selon timeframes requis
2. Configuration des fenêtres d'agrégation
3. Préparation des buffers de données
4. Validation des paramètres d'entrée

### Phase 2: Parsing streaming
1. **Pour chaque fichier ZIP:**
   - Ouverture stream ZIP
   - Parsing ligne par ligne (klines ou trades)
   - Validation format et cohérence
   - Accumulation dans buffers temporaires

2. **Validation temporelle:**
   - Vérification séquences temporelles
   - Détection de gaps de données
   - Signalement des anomalies

### Phase 3: Intégration multi-timeframes
1. **Synchronisation:**
   - Alignement des timeframes sur timestamps communs
   - Gestion des décalages entre sources
   - Interpolation si nécessaire pour gaps mineurs

2. **Agrégation:**
   - Construction du contexte unifié par timestamp
   - Validation cohérence cross-timeframe
   - Préparation pour engines de calcul

## Configuration requise

```yaml
parsing:
  supported_timeframes: ["5m", "15m", "1h", "4h"]
  validation_strict: true
  buffer_size_mb: 64
  trade_aggregation_window_ms: 1000
  
integration:
  sync_tolerance_ms: 1000
  gap_interpolation: false
  context_buffer_hours: 24
```

## Structures de données

### MarketContext
```go
type MarketContext struct {
    Timestamp    int64
    Symbol       string
    Klines       map[string]*Kline  // timeframe -> kline
    TradeMetrics *TradeMetrics
    Metadata     map[string]interface{}
}

type TradeMetrics struct {
    Volume       float64
    TradeCount   int64
    BuyVolume    float64
    SellVolume   float64
    VWAP         float64
    Spread       float64
}
```

## Tests unitaires requis

- `TestKlineParserValidFormats`
- `TestTradeParserValidFormats`
- `TestMultiTimeframeSynchronization`
- `TestGapDetectionAndHandling`
- `TestDataIntegrityValidation`
- `TestMemoryUsageConstraints`
- `TestStreamingPerformance`

## Critères d'acceptation

✅ Parsing streaming sans full-load mémoire  
✅ Support multi-timeframes synchronisé  
✅ Validation d'intégrité des données  
✅ Gestion des gaps et anomalies  
✅ Performance optimisée (< 100MB RAM)  
✅ Interface cohérente pour engines  
✅ Tests de robustesse extensifs
