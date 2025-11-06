# Workflow 1: Infrastructure de base

**Version:** 0.1  
**Statut:** Spécification technique  
**Module:** Téléchargement données Binance Vision

## Vue d'ensemble
Ce workflow établit les fondations du module de téléchargement des données historiques Binance Vision pour l'agent économique de trading.

## Composants principaux

### 1. Gestionnaire de cache local
**Fichier:** `internal/data/binance/cache.go`

**Responsabilités:**
- Structure hiérarchique `data/binance/futures_um/klines/SYMBOL/TIMEFRAME/` et `data/binance/futures_um/trades/SYMBOL/`
- Index local JSON des fichiers disponibles
- Vérification d'existence avant téléchargement
- Gestion des métadonnées (taille, checksum, date modification)

**Fonctions clés:**
```go
func InitializeCache(rootPath string) (*CacheManager, error)
func (c *CacheManager) FileExists(symbol, dataType, date string, timeframe ...string) bool
func (c *CacheManager) GetFilePath(symbol, dataType, date string, timeframe ...string) string
func (c *CacheManager) UpdateIndex(fileInfo FileMetadata) error
func (c *CacheManager) IsFileCorrupted(filePath string) (bool, error)
func (c *CacheManager) GetCacheStats() *CacheStatistics
```

### 2. Téléchargeur intelligent
**Fichier:** `internal/data/binance/downloader.go`

**Responsabilités:**
- Client HTTP configuré pour Binance Data Vision
- Vérification d'existence avant téléchargement
- Gestion des reprises sur interruption
- Validation d'intégrité post-téléchargement

**Fonctions clés:**
```go
func NewDownloader(config DownloaderConfig) *Downloader
func (d *Downloader) DownloadFile(url, localPath string) error
func (d *Downloader) ResumeDownload(url, localPath string, offset int64) error
func (d *Downloader) ValidateChecksum(filePath, expectedChecksum string) bool
func (d *Downloader) GetFileSize(url string) (int64, error)
func (d *Downloader) DownloadFiles(requests []DownloadRequest) error
func (d *Downloader) GetProgress(localPath string) (*DownloadProgress, error)
func (d *Downloader) CancelDownload(localPath string) error
```

### 3. Lecteur streaming
**Fichier:** `internal/data/binance/streaming.go`

**Responsabilités:**
- Décompression ZIP à la volée
- Interface iterator pour parcours séquentiel
- Buffer glissant minimal pour optimisation mémoire
- Gestion des erreurs de lecture

**Fonctions clés:**
```go
func NewZipStreamReader(filePath string) (*ZipStreamReader, error)
func (r *ZipStreamReader) NextEntry() (*StreamEntry, error)
func (r *ZipStreamReader) ReadLine() ([]byte, error)
func (r *ZipStreamReader) Close() error
func (r *ZipStreamReader) HasNext() bool
```

## Flux d'exécution

### Phase 1: Initialisation
1. Création du répertoire de cache si inexistant
2. Chargement de l'index des fichiers existants
3. Configuration du client HTTP avec timeouts appropriés
4. Validation des paramètres de configuration

### Phase 2: Vérification de cache
1. Pour chaque fichier requis (paire/date/timeframe):
   - Vérification existence locale
   - Validation intégrité si présent
   - Marquage pour téléchargement si absent/corrompu

### Phase 3: Téléchargements manquants
1. Téléchargement parallélisé avec limite de connexions
2. Validation des checksums post-téléchargement
3. Mise à jour de l'index local
4. Notification des erreurs critiques

## Configuration requise

```yaml
binance_data:
  cache_root: "data/binance"
  base_url: "https://data.binance.vision"
  http_timeout: 30s
  max_parallel_downloads: 5
  retry_attempts: 3
  retry_delay: 1s
  checksum_validation: true
```

## Tests unitaires requis

- `TestCacheManagerInitialization`
- `TestFileExistenceCheck`
- `TestDownloadWithResume`
- `TestChecksumValidation`
- `TestZipStreamReading`
- `TestConcurrentDownloads`

## Critères d'acceptation

✅ Structure de cache hiérarchique créée  
✅ Index local maintenu à jour  
✅ Téléchargements optimisés (éviter les doublons)  
✅ Reprises fonctionnelles sur interruption  
✅ Validation d'intégrité systématique  
✅ Gestion d'erreurs robuste  
✅ Tests unitaires couvrant 90%+ du code
