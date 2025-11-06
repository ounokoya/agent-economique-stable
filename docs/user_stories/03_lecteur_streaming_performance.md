# User Story 3: Lecteur streaming haute performance

**Epic:** Gestionnaire de données historiques Binance  
**Priorité:** Haute  
**Estimation:** 8 points  
**Sprint:** 2

## Description

> **En tant que** Kline Engine et Tick Engine  
> **Je veux** accéder aux données sans charger tout en mémoire  
> **Afin de** traiter efficacement les timeframes 5m/15m/1h/4h et les trades

## Contexte métier

Les fichiers ZIP de Binance Vision peuvent contenir des millions de lignes de données. Charger tout en mémoire dépasserait les contraintes (plusieurs GB par fichier). Le streaming permet de traiter les données au fur et à mesure de leur lecture.

## Critères d'acceptation

### ✅ Décompression ZIP en streaming
- **ÉTANT DONNÉ** un fichier ZIP de données Binance
- **QUAND** la lecture commence
- **ALORS** la décompression doit se faire ligne par ligne sans extraction complète

### ✅ Interface iterator pour parcours séquentiel
- **ÉTANT DONNÉ** des données en cours de streaming
- **QUAND** le code client demande la ligne suivante
- **ALORS** l'iterator doit fournir la prochaine ligne disponible

### ✅ Buffer glissant minimal
- **ÉTANT DONNÉ** des données streamées
- **QUAND** le traitement nécessite une fenêtre d'analyse
- **ALORS** seul le minimum requis doit être maintenu en mémoire

### ✅ Parsing direct sans stockage intermédiaire
- **ÉTANT DONNÉ** une ligne de données lue depuis le ZIP
- **QUAND** elle est parsée en structure Go
- **ALORS** aucun stockage temporaire ne doit être utilisé

## Définition of Done

- [ ] Implémentation streaming avec `archive/zip` et `bufio`
- [ ] Interface iterator propre et réutilisable
- [ ] Tests de performance avec gros fichiers (>100MB)
- [ ] Tests de consommation mémoire (contrainte <100MB)
- [ ] Gestion d'erreurs robuste pendant streaming
- [ ] Benchmarks de performance documentés

## Cas de test principaux

### Test 1: Streaming ZIP complet
```go
func TestZipStreamReading() {
    reader, err := NewZipStreamReader("test_data.zip")
    assert.NoError(t, err)
    defer reader.Close()
    
    lineCount := 0
    for reader.HasNext() {
        line, err := reader.ReadLine()
        assert.NoError(t, err)
        assert.NotEmpty(t, line)
        lineCount++
    }
    
    assert.Greater(t, lineCount, 1000) // Vérifier volume données
}
```

### Test 2: Contrainte mémoire respectée
```go
func TestMemoryConstraint() {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Traiter gros fichier
    processLargeFile("large_data.zip")
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    memoryUsed := m2.Alloc - m1.Alloc
    assert.Less(t, memoryUsed, 100*1024*1024) // < 100MB
}
```

### Test 3: Performance lecture séquentielle
```go
func BenchmarkStreamReading(b *testing.B) {
    reader, _ := NewZipStreamReader("benchmark_data.zip")
    defer reader.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if reader.HasNext() {
            reader.ReadLine()
        }
    }
}
```

## Spécifications techniques

### Architecture streaming
```go
type ZipStreamReader struct {
    zipFile    *zip.ReadCloser
    currentEntry *zip.File
    scanner    *bufio.Scanner
    buffer     []byte
    index      int
    isEOF      bool
}

type StreamEntry struct {
    Filename string
    Size     int64
    Data     io.Reader
}
```

### Interface publique
```go
func NewZipStreamReader(filePath string) (*ZipStreamReader, error)
func (r *ZipStreamReader) NextEntry() (*StreamEntry, error)
func (r *ZipStreamReader) ReadLine() ([]byte, error)
func (r *ZipStreamReader) HasNext() bool
func (r *ZipStreamReader) Close() error
func (r *ZipStreamReader) GetProgress() float64
```

### Parsers spécialisés
```go
type KlineStreamParser struct {
    reader *ZipStreamReader
    buffer []*Kline
    window int
}

type TradeStreamParser struct {
    reader *ZipStreamReader
    aggregator *TradeAggregator
    windowMs int64
}
```

## Configuration optimisation

```yaml
streaming:
  buffer_size: 65536      # 64KB buffer de lecture
  scan_buffer_size: 4096  # 4KB buffer scanner
  max_line_size: 1024     # Limite taille ligne
  compression_level: 6    # Balance CPU/mémoire
  
performance:
  max_memory_mb: 100      # Contrainte mémoire stricte
  min_throughput_mbps: 50 # Débit minimum attendu
  gc_frequency: 1000      # GC après N lignes
```

## Formats de données traités

### Klines Binance (CSV)
```
1687654800000,30.45,30.48,30.44,30.47,1234.56,1687654859999,37658.789,123,567.89,17234.567
[timestamp],[open],[high],[low],[close],[volume],[close_time],[quote_asset_volume],[count],[taker_buy_base_asset_volume],[taker_buy_quote_asset_volume]
```

### Trades Binance (CSV)
```
123456789,30.45,12.34,1687654800123,true,100
[trade_id],[price],[quantity],[timestamp],[is_buyer_maker],[best_match]
```

## Dépendances

- Dépend de: User Story 2 (Téléchargeur robuste)
- Bloque: User Story 4 (Intégration stratégie)
- Packages Go: `archive/zip`, `bufio`, `io`, `runtime`

## Optimisations implémentées

### Buffer management
- Buffer circulaire pour fenêtres glissantes
- Pooling des buffers pour éviter allocations
- GC hints pour libération mémoire proactive

### CPU optimisations
- Parsing in-place sans string allocations
- Réutilisation des structures de données
- Inlining des fonctions critiques

### I/O optimisations
- Lecture buffurisée avec taille optimale
- Préfetch pour masquer latences disque
- Streaming parallèle pour fichiers multiples

## Métriques de performance cibles

- **Débit:** Optimisé selon capacités système
- **Mémoire:** Streaming sans accumulation - pas de chargement complet
- **Latence:** < 1ms par ligne parsée
- **CPU:** < 50% utilisation mono-core

### Test contrainte mémoire streaming
**Objectif:** Validation streaming sans accumulation
**Logique testée:**
- Traitement fichier ZIP > 500MB en streaming pur
- Mesure consommation mémoire constante
- Pas de croissance mémoire avec taille fichier
- GC efficace des buffers temporaires
- Compatibilité avec tous formats Binance Vision
