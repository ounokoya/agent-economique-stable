# User Story 2: Téléchargeur robuste

**Epic:** Gestionnaire de données historiques Binance  
**Priorité:** Haute  
**Estimation:** 8 points  
**Sprint:** 1

## Description

> **En tant qu'** agent de backtest  
> **Je veux** un téléchargeur qui gère les interruptions et reprises  
> **Afin de** récupérer de façon fiable les données SOLUSDT/SUIUSDT/ETHUSDT (2023-06-01 à 2025-06-29)

## Contexte métier

Les données Binance Vision représentent plusieurs GB de fichiers ZIP. Les téléchargements peuvent être interrompus par des problèmes réseau, des redémarrages système, ou des timeouts. L'agent doit pouvoir reprendre les téléchargements là où ils se sont arrêtés.

## Critères d'acceptation

### ✅ Téléchargement par chunks avec reprise
- **ÉTANT DONNÉ** qu'un téléchargement est interrompu à 60%
- **QUAND** le système redémarre le téléchargement
- **ALORS** il doit reprendre à partir de 60% et non depuis le début

### ✅ Gestion erreurs réseau avec retry exponentiel
- **ÉTANT DONNÉ** qu'une erreur réseau temporaire survient
- **QUAND** le téléchargement échoue
- **ALORS** le système doit retry avec délais croissants (1s, 2s, 4s, 8s)

### ✅ Validation checksums
- **ÉTANT DONNÉ** qu'un fichier est complètement téléchargé
- **QUAND** la validation d'intégrité est effectuée
- **ALORS** le checksum doit correspondre à celui fourni par Binance (si disponible)

### ✅ Progress tracking
- **ÉTANT DONNÉ** qu'un téléchargement de gros volume est en cours
- **QUAND** l'utilisateur consulte le statut
- **ALORS** le pourcentage de progression doit être affiché

## Définition of Done

- [ ] Implémentation Go avec gestion HTTP ranges
- [ ] Tests unitaires incluant simulation d'interruptions
- [ ] Tests d'intégration avec vrais URLs Binance
- [ ] Logging structuré des événements de téléchargement
- [ ] Métriques de performance (vitesse, taux d'erreur)
- [ ] Documentation API complète

## Cas de test principaux

### Test 1: Téléchargement complet sans interruption
```go
func TestFullDownload() {
    downloader := NewDownloader(testConfig)
    err := downloader.DownloadFile(testURL, testPath)
    assert.NoError(t, err)
    assert.FileExists(t, testPath)
}
```

### Test 2: Reprise après interruption
```go
func TestResumeDownload() {
    // Simuler interruption à 50%
    partialFile := createPartialFile(testPath, 50)
    
    downloader := NewDownloader(testConfig)  
    err := downloader.ResumeDownload(testURL, testPath, partialFile.Size())
    
    assert.NoError(t, err)
    assert.True(t, isFileComplete(testPath))
}
```

### Test 3: Retry avec backoff exponentiel
```go
func TestRetryWithBackoff() {
    // Mock serveur qui échoue 3 fois puis réussit
    mockServer := setupFailingServer(3)
    
    downloader := NewDownloader(testConfig)
    start := time.Now()
    err := downloader.DownloadFile(mockServer.URL, testPath)
    elapsed := time.Since(start)
    
    assert.NoError(t, err)
    assert.True(t, elapsed > 7*time.Second) // 1+2+4 = 7s minimum
}
```

## Configuration requise

```yaml
downloader:
  base_url: "https://data.binance.vision"
  timeout: 30s
  max_retries: 5
  initial_retry_delay: 1s
  max_retry_delay: 30s
  chunk_size: 8192
  max_concurrent_downloads: 3
  user_agent: "agent-economique/0.1"
```

## Spécifications techniques

### Structure principale
```go
type Downloader struct {
    client      *http.Client
    config      DownloaderConfig
    activeDownloads map[string]*DownloadState
    mutex       sync.RWMutex
}

type DownloadState struct {
    URL         string
    LocalPath   string
    TotalSize   int64
    Downloaded  int64
    StartTime   time.Time
    LastUpdate  time.Time
    Status      DownloadStatus
}
```

### Interface publique
```go
func NewDownloader(config DownloaderConfig) *Downloader
func (d *Downloader) DownloadFile(url, localPath string) error
func (d *Downloader) DownloadFiles(requests []DownloadRequest) error
func (d *Downloader) GetProgress(localPath string) (*DownloadProgress, error)
func (d *Downloader) CancelDownload(localPath string) error
```

## Dépendances

- Dépend de: User Story 1 (Cache intelligent)
- Bloque: User Story 3 (Lecteur streaming)
- Packages Go: `net/http`, `io`, `os`, `crypto/sha256`

## URLs cibles Binance Vision

```
Klines:
https://data.binance.vision/data/futures/um/daily/klines/SOLUSDT/5m/SOLUSDT-5m-2023-06-01.zip

Trades:
https://data.binance.vision/data/futures/um/daily/trades/SOLUSDT/SOLUSDT-trades-2023-06-01.zip

Checksums:
https://data.binance.vision/data/futures/um/daily/klines/SOLUSDT/5m/SOLUSDT-5m-2023-06-01.zip.CHECKSUM
https://data.binance.vision/data/futures/um/daily/trades/SOLUSDT/SOLUSDT-trades-2023-06-01.zip.CHECKSUM
```

## Métriques de performance

- Débit adapté à la connexion disponible
- Taux d'erreur acceptable: < 1% avec retry
- Temps de reprise: < 5 secondes après interruption
- Overhead mémoire: Streaming sans accumulation

## Risques identifiés

- **Rate limiting:** Binance peut limiter les requêtes
- **Espace disque:** Vérifier avant téléchargement
- **Interruptions:** Système ou réseau instable
- **Checksum:** Pas toujours disponible côté Binance

## Critères de validation

- Téléchargements de plusieurs GB sans échec
- Reprises fonctionnelles après coupures réseau
- Logs détaillés pour debugging
- Performance acceptable en conditions dégradées
