# User Story 5: Monitoring et diagnostics

**Epic:** Gestionnaire de données historiques Binance  
**Priorité:** Moyenne  
**Estimation:** 5 points  
**Sprint:** 3

## Description

> **En tant qu'** développeur/opérateur  
> **Je veux** surveiller l'état du téléchargement et la qualité des données  
> **Afin de** diagnostiquer rapidement les problèmes de backtest

## Contexte métier

L'agent économique traite des volumes importants de données historiques dans un contexte de backtest critique. Les problèmes de qualité des données, les échecs de téléchargement, ou les dégradations de performance peuvent impacter directement les résultats de trading. Un système de monitoring robuste est essentiel.

## Critères d'acceptation

### ✅ Logs structurés JSON avec niveaux
- **ÉTANT DONNÉ** que l'agent traite des données
- **QUAND** des événements significatifs surviennent
- **ALORS** ils doivent être loggés avec niveau approprié (debug/info/warn/error)

### ✅ Métriques de performance et santé
- **ÉTANT DONNÉ** que le système est en fonctionnement
- **QUAND** les métriques sont collectées
- **ALORS** elles doivent inclure: taux cache hit, vitesse téléchargement, erreurs réseau

### ✅ Diagnostics de qualité des données
- **ÉTANT DONNÉ** des données ingérées depuis Binance
- **QUAND** la validation qualité s'exécute
- **ALORS** les gaps, fichiers corrompus, et timeframes manquants doivent être détectés

### ✅ Dashboard optionnel pour monitoring visuel
- **ÉTANT DONNÉ** que l'opérateur souhaite superviser le système
- **QUAND** il accède au dashboard
- **ALORS** les métriques clés doivent être affichées en temps réel

## Définition of Done

- [ ] Logger structuré JSON avec niveaux configurables
- [ ] Collecteur de métriques avec export Prometheus
- [ ] Détecteur de qualité données avec alertes
- [ ] Dashboard web simple (optionnel)
- [ ] Tests de charge pour validation métriques
- [ ] Documentation opérationnelle complète

## Cas de test principaux

### Test 1: Logging structuré
```go
func TestStructuredLogging() {
    logger := NewStructuredLogger(LogConfig{Level: "info"})
    
    logger.Info("download_started", map[string]interface{}{
        "symbol": "SOLUSDT",
        "date": "2023-06-01",
        "size_mb": 125.4,
    })
    
    // Vérifier format JSON en sortie
    logOutput := captureLogOutput()
    var logEntry map[string]interface{}
    err := json.Unmarshal(logOutput, &logEntry)
    
    assert.NoError(t, err)
    assert.Equal(t, "info", logEntry["level"])
    assert.Equal(t, "download_started", logEntry["message"])
}
```

### Test 2: Métriques de performance
```go
func TestPerformanceMetrics() {
    collector := NewMetricsCollector()
    
    // Simuler activité
    collector.RecordDownload("SOLUSDT", 100*1024*1024, time.Second*30)
    collector.RecordCacheHit("ETHUSDT")
    collector.RecordError("network_timeout")
    
    metrics := collector.GetMetrics()
    assert.Greater(t, metrics.DownloadSpeed, 3.0) // > 3MB/s
    assert.Equal(t, 1, metrics.CacheHits)
    assert.Equal(t, 1, metrics.ErrorCount)
}
```

### Test 3: Détection gaps de données
```go
func TestDataGapDetection() {
    validator := NewDataQualityValidator()
    
    // Données avec gap de 2 heures
    klines := generateKlinesWithGap("2023-06-01T10:00Z", "2023-06-01T12:00Z")
    
    issues := validator.ValidateKlines(klines)
    
    assert.Len(t, issues, 1)
    assert.Equal(t, "data_gap", issues[0].Type)
    assert.Equal(t, 2*time.Hour, issues[0].Duration)
}
```

## Spécifications techniques

### Logger structuré
```go
type StructuredLogger struct {
    level      LogLevel
    writer     io.Writer
    formatter  Formatter
    context    map[string]interface{}
}

type LogEntry struct {
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Component string                 `json:"component"`
    Context   map[string]interface{} `json:"context"`
    RunID     string                 `json:"run_id"`
}

func (l *StructuredLogger) Info(msg string, ctx map[string]interface{})
func (l *StructuredLogger) Warn(msg string, ctx map[string]interface{})
func (l *StructuredLogger) Error(msg string, err error, ctx map[string]interface{})
```

### Collecteur de métriques
```go
type MetricsCollector struct {
    downloadSpeed    prometheus.Histogram
    cacheHitRate     prometheus.Counter
    errorCount       prometheus.CounterVec
    dataQuality      prometheus.Gauge
    activeDownloads  prometheus.Gauge
}

type SystemMetrics struct {
    // Performance
    DownloadSpeed      float64   `json:"download_speed_mbps"`
    CacheHitRate       float64   `json:"cache_hit_rate"`
    AvgLatency         time.Duration `json:"avg_latency"`
    
    // Santé système
    ErrorRate          float64   `json:"error_rate"`
    ActiveConnections  int       `json:"active_connections"`
    MemoryUsage        uint64    `json:"memory_usage_mb"`
    
    // Qualité données
    DataQualityScore   float64   `json:"data_quality_score"`
    MissingDataPoints  int       `json:"missing_data_points"`
    CorruptedFiles     int       `json:"corrupted_files"`
}
```

### Validateur qualité données
```go
type DataQualityValidator struct {
    rules []ValidationRule
}

type DataQualityIssue struct {
    Type        string        `json:"type"`
    Severity    string        `json:"severity"`
    Symbol      string        `json:"symbol"`
    Timeframe   string        `json:"timeframe"`
    Timestamp   int64         `json:"timestamp"`
    Description string        `json:"description"`
    Duration    time.Duration `json:"duration,omitempty"`
}

func (v *DataQualityValidator) ValidateKlines(klines []Kline) []DataQualityIssue
func (v *DataQualityValidator) ValidateTrades(trades []Trade) []DataQualityIssue
func (v *DataQualityValidator) ValidateTimeframes(data map[string][]Kline) []DataQualityIssue
```

## Configuration monitoring

```yaml
monitoring:
  logging:
    level: "info"                    # debug, info, warn, error
    format: "json"                   # json, text
    output: "stdout"                 # stdout, file, both
    file_path: "logs/agent.log"
    rotation_size: "100MB"
    max_files: 10
  
  metrics:
    enabled: true
    port: 9090                       # Port Prometheus
    endpoint: "/metrics"
    collection_interval: "30s"
    
  quality:
    validation_enabled: true
    gap_tolerance: "5m"              # Tolérance gaps données
    corruption_threshold: 0.01       # 1% max fichiers corrompus
    alert_on_issues: true
    
  dashboard:
    enabled: false                   # Optionnel
    port: 8080
    refresh_interval: "10s"
```

## Dashboard web simple (optionnel)

### Pages principales
- **Overview:** Métriques système temps réel
- **Downloads:** Status téléchargements en cours
- **Data Quality:** Scores qualité par paire/timeframe
- **Logs:** Stream logs avec filtrage

### Endpoints API
```go
GET /api/metrics          // Métriques système JSON
GET /api/downloads        // Status téléchargements
GET /api/quality          // Issues qualité données
GET /api/logs?level=info  // Logs filtrés
```

## Alertes configurables

### Types d'alertes
- **Performance:** Vitesse téléchargement < seuil
- **Cache:** Taux hit rate < seuil
- **Erreurs:** Taux erreur > seuil
- **Qualité:** Score qualité < seuil
- **Système:** Utilisation mémoire > seuil

### Canaux notifications
- Logs (toujours actif)
- Console (développement)
- Webhook (production optionnel)

## Dépendances

- Dépend de: Toutes les user stories précédentes
- Packages Go: `prometheus`, `logrus`, `net/http`
- Optionnel: Dashboard web framework léger

## Métriques cibles

### Performance
- Download speed: > 10 MB/s
- Cache hit rate: > 80%
- Error rate: < 1%
- Latency: < 500ms

### Qualité
- Data quality score: > 95%
- Missing data: < 0.1%
- Corrupted files: < 0.01%

## Critères de validation

- Logs structurés analysables par outils standards
- Métriques exportables vers systèmes de monitoring
- Détection proactive des problèmes qualité
- Interface de supervision accessible et informative
- Performance du monitoring < 5% overhead système
