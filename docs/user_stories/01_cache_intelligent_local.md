# User Story 1: Cache intelligent local

**Epic:** Gestionnaire de données historiques Binance  
**Priorité:** Haute  
**Estimation:** 5 points  
**Sprint:** 1

## Description

> **En tant qu'** agent de trading en backtest  
> **Je veux** un système de cache local organisé par date/paire/timeframe  
> **Afin de** éviter les re-téléchargements et optimiser les temps d'exécution

## Contexte métier


## Critères d'acceptation

### ✅ Structure hiérarchique
- **ÉTANT DONNÉ** que l'agent a besoin de données organisées
- **QUAND** le cache est initialisé
- **ALORS** la structure `data/binance/futures_um/{klines|trades}/SYMBOL/[TIMEFRAME/]` doit être créée

### ✅ Index local JSON
- **ÉTANT DONNÉ** que le cache contient des fichiers
- **QUAND** un fichier est ajouté ou modifié
- **ALORS** l'index JSON doit être mis à jour avec les métadonnées (taille, checksum, date)
### ✅ Vérification d'existence
- **ÉTANT DONNÉ** qu'un fichier est requis pour le backtest
- **QUAND** le système vérifie sa disponibilité
- **ALORS** il doit consulter l'index local avant tout téléchargement

### ✅ Gestion fichiers corrompus
- **ÉTANT DONNÉ** qu'un fichier peut être corrompu
- **QUAND** une vérification d'intégrité échoue
- **ALORS** le fichier doit être marqué pour re-téléchargement automatique

## Définition of Done

- [ ] Code implémenté selon contraintes (Go, max 500 lignes)
- [ ] Tests unitaires couvrant 90%+ du code
- [ ] Tests d'intégration pour structure complète
- [ ] Documentation technique complète
- [ ] Validation par review code
- [ ] Performance testée (< 100ms pour vérifications)

## Cas de test principaux

### Test 1: Initialisation cache vide
```go
func TestCacheInitialization() {
    cache := InitializeCache("./test_data")
    assert.True(t, cache.DirectoryExists("data/binance/futures_um"))
    assert.FileExists(t, "data/binance/futures_um/index.json")
}
```

### Test 2: Vérification existence fichier
```go
func TestFileExistenceCheck() {
    cache := setupTestCache()
    exists := cache.FileExists("SOLUSDT", "klines", "2023-06-01")
    assert.False(t, exists) // fichier pas encore téléchargé
}
```

### Test 3: Détection corruption
```go
func TestCorruptionDetection() {
    cache := setupTestCache()
    // Créer fichier corrompu
    corrupted := cache.IsFileCorrupted("./test_corrupted.zip")
    assert.True(t, corrupted)
}
```

## Dépendances

- Aucune dépendance externe
- Prérequis: structure projet Go initialisée
- Bloque: User Story 2 (Téléchargeur robuste)

## Notes techniques

- Utilisation de `encoding/json` pour l'index
- Checksums SHA256 pour validation d'intégrité
- Gestion concurrent-safe avec mutex
- Optimisation: cache en mémoire de l'index pour performances

## Risques identifiés

- **Espace disque:** Gros volumes de données (plusieurs GB)
- **Concurrence:** Accès simultanés pendant téléchargements
- **Corruption:** Détection fiable des fichiers partiels

## Critères de validation

- Temps de vérification d'existence < 100ms
- Consommation mémoire < 50MB pour l'index
- Gestion propre des erreurs I/O
- Nettoyage automatique des fichiers orphelins
