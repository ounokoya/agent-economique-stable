# Guide Performance et Monitoring

**Version:** 0.1  
**Statut:** Guide opérationnel  
**Scope:** Performance, métriques et monitoring du système

## 📊 **Objectifs de performance**

### **Métriques cibles**
| Métrique | Cible | Contrainte |
|----------|-------|------------|
| Débit streaming | Optimisé système | Lecture ZIP streaming |
| Mémoire | Streaming sans accumulation | Pas de chargement complet |
| Latence signaux | <500ms | End-to-end |
| Cache hit rate | >80% | Optimisation téléchargements |
| Qualité données | >95% | Score validation |

### **Contraintes strictes**
- **Streaming pur** : Mémoire constante indépendamment de la taille fichier
- **Pas d'accumulation** : Traitement ligne par ligne sans stockage intermédiaire
- **Validation** : Checksums systématiques pour intégrité
- **Robustesse** : Reprises automatiques sur interruption

## 🔍 **Monitoring et Métriques**

### **Métriques exposées**
- **Performance**: Vitesse téléchargement, latence, throughput
- **Cache**: Hit rate, taille, évictions
- **Qualité**: Score données, gaps détectés, fichiers corrompus
- **Système**: Mémoire, CPU, connexions réseau

### **Logs structurés**
Format JSON standardisé avec corrélation :

```json
{
  "timestamp": "2025-10-30T16:27:14Z",
  "level": "info",
  "message": "download_completed",
  "component": "downloader",
  "context": {
    "symbol": "SOLUSDT",
    "date": "2023-06-01",
    "size_mb": 125.4,
    "duration_ms": 12500
  },
  "run_id": "backtest-20251030-162714"
}
```

## 🧪 **Tests de performance**

### **Benchmarks disponibles**
```bash
# Exécuter les benchmarks
go test -bench=. ./internal/data/binance/...

# Tests de performance mémoire
go test -memprofile=mem.prof ./internal/data/binance/
go tool pprof mem.prof

# Tests de charge
go test -tags=loadtest ./tests/performance/...
```

### **Validation performance**
- **Traitement fichiers > 1GB** sans overflow mémoire
- **Performance stable** sur durées longues
- **Gestion propre** des erreurs I/O
- **Compatibilité** avec tous formats Binance Vision

## 📈 **Optimisations implémentées**

### **Streaming ZIP**
- Décompression à la volée sans extraction complète
- Buffer circulaire pour fenêtres glissantes
- Pooling des buffers pour éviter allocations
- GC hints pour libération mémoire proactive

### **Cache intelligent**
- Vérification existence < 1ms via index mémoire
- Thread-safe avec RWMutex
- Checksums SHA256 pour validation
- Nettoyage automatique fichiers corrompus

### **Réseau optimisé**
- HTTP ranges pour reprises téléchargement
- Retry exponentiel configurable
- Connexions parallèles limitées
- Validation checksums systématique

## 🚨 **Alertes et Seuils**

### **Seuils critiques**
- **Vitesse téléchargement** < 1 MB/s → Alerte réseau
- **Cache hit rate** < 50% → Problème configuration
- **Erreur rate** > 5% → Instabilité système
- **Mémoire** > contrainte streaming → Fuite mémoire

### **Actions automatiques**
- **Retry automatique** sur erreurs temporaires
- **Fallback cache** si serveur indisponible
- **Limitation débit** si surcharge détectée
- **Circuit breaker** si taux d'erreur critique
