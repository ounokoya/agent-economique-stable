# Guide d'Utilisation - Application CLI Agent Économique

**Version :** 1.0.0  
**Date :** 2025-10-31  
**Public :** Développeurs, Traders, Analystes

## 🚀 Introduction

L'application CLI Agent Économique permet de télécharger et analyser automatiquement les données de marché Binance Vision multi-timeframes. Elle utilise une approche modulaire sans agrégation, téléchargeant directement chaque timeframe depuis les serveurs officiels Binance.

## 📋 Prérequis

- **Go 1.19+** installé
- **Connexion internet** stable
- **Espace disque** suffisant (recommandé : 10GB+)
- **Fichier de configuration** YAML valide

## 🔧 Installation et Compilation

### Compilation
```bash
cd /path/to/agent-economique
go build -o agent-economique ./cmd/agent/
```

### Permissions (Linux/Mac)
```bash
chmod +x agent-economique
```

## ⚙️ Configuration

### Fichier Principal : `config/config.yaml`

```yaml
# Configuration des données Binance
binance_data:
  cache_root: "data/binance"
  symbols: 
    - "SOLUSDT"
    - "ETHUSDT"
    - "BTCUSDT"
  timeframes:
    - "5m"
    - "15m"
    - "1h"
    - "4h"
    - "1d"
  
  downloader:
    base_url: "https://data.binance.vision"
    max_retries: 3
    timeout: "10m"
    max_concurrent: 5
    
  streaming:
    buffer_size: 65536
    max_memory_mb: 100
    enable_metrics: true

# Configuration CLI
cli:
  execution_mode: "default"    # Mode d'exécution par défaut
  memory_limit_mb: 512         # Limite mémoire streaming
  force_redownload: false      # Re-téléchargement forcé
  verbose: false               # Logs détaillés
  enable_metrics: true         # Métriques de performance

# Période de données
data_period:
  start_date: "2023-06-01"
  end_date: "2023-06-30"
```

## 🎮 Utilisation de Base

### Commande Minimale
```bash
./agent-economique --config config/config.yaml
```

### Syntaxe Complète
```bash
./agent-economique [OPTIONS]
```

## 📝 Options Ligne de Commande

| Option | Description | Exemple |
|--------|-------------|---------|
| `--config <file>` | **Obligatoire** - Fichier de configuration YAML | `--config config.yaml` |
| `--symbols <list>` | Liste de symboles (remplace config) | `--symbols SOLUSDT,ETHUSDT` |
| `--timeframes <list>` | Liste de timeframes (remplace config) | `--timeframes 5m,1h,1d` |
| `--mode <mode>` | Mode d'exécution | `--mode download-only` |
| `--memory-limit <MB>` | Limite mémoire (mode streaming) | `--memory-limit 256` |
| `--force-redownload` | Forcer le re-téléchargement | `--force-redownload` |
| `--verbose` | Logs détaillés | `--verbose` |
| `--enable-metrics` | Activer métriques | `--enable-metrics` |

## 🎯 Modes d'Exécution

### 1. Mode Default (Complet)
**Description :** Téléchargement + Traitement + Statistiques
```bash
./agent-economique --config config/config.yaml
# ou
./agent-economique --config config/config.yaml --mode default
```

**Utilisation :** Production, analyse complète des données

### 2. Mode Download-Only
**Description :** Téléchargement uniquement, pas de traitement
```bash
./agent-economique --config config/config.yaml --mode download-only
```

**Utilisation :** Mise en cache rapide, synchronisation de données

### 3. Mode Streaming
**Description :** Traitement optimisé mémoire avec contraintes
```bash
./agent-economique --config config/config.yaml --mode streaming --memory-limit 128
```

**Utilisation :** Serveurs avec mémoire limitée, traitement temps réel

### 4. Mode Batch
**Description :** Traitement par lots pour de gros volumes
```bash
./agent-economique --config config/config.yaml --mode batch
```

**Utilisation :** Traitement historique massif

## 📊 Exemples d'Utilisation

### Exemple 1 : Téléchargement Rapide SOL
```bash
./agent-economique \
  --config config/config.yaml \
  --symbols SOLUSDT \
  --timeframes 5m,1h \
  --mode download-only
```

### Exemple 2 : Analyse Complète Multi-Symboles
```bash
./agent-economique \
  --config config/config.yaml \
  --symbols SOLUSDT,ETHUSDT,BTCUSDT \
  --timeframes 15m,1h,4h \
  --verbose \
  --enable-metrics
```

### Exemple 3 : Mode Économie Mémoire
```bash
./agent-economique \
  --config config/config.yaml \
  --mode streaming \
  --memory-limit 64 \
  --symbols SOLUSDT
```

### Exemple 4 : Re-téléchargement Forcé
```bash
# Force le re-téléchargement (ignore le cache)
./agent-economique \
  --config config/config.yaml \
  --force-redownload \
  --symbols ETHUSDT \
  --timeframes 1h

# ⚡ Performance comparison:
# Sans --force-redownload: ~400µs (cache hit)
# Avec --force-redownload: ~24s (re-téléchargement complet)
```

## 📈 Interprétation des Résultats

### Rapport d'Exécution Type
```
📊 RAPPORT D'EXÉCUTION
===================================================
Résumé: Successfully processed 12 files for 3 symbols in 21.95s
Temps d'exécution: 21.951314034s
Taux de succès: 100.0%
Volume de données: 12.00 MB

Symboles traités:
  - SOLUSDT
  - ETHUSDT
  - BTCUSDT

Timeframes générés:
  - 5m
  - 15m
  - 1h
  - 4h

Recommandations:
  💡 Tous les téléchargements ont réussi
  💡 Mémoire utilisée dans les limites normales
```

### Codes de Sortie
- **0** : Succès complet
- **1** : Erreur d'exécution (voir logs)

## 🗂️ Structure des Données Téléchargées

```
data/binance/
├── binance/
│   └── futures_um/
│       ├── klines/
│       │   ├── SOLUSDT/
│       │   │   ├── 5m/
│       │   │   │   └── SOLUSDT-5m-2023-06-01.zip
│       │   │   ├── 15m/
│       │   │   ├── 1h/
│       │   │   └── 4h/
│       │   └── ETHUSDT/
│       └── trades/
└── index.json  # Cache metadata
```

## 🔍 Gestion d'Erreurs

### Erreurs Communes

#### 1. Fichier de Configuration Introuvable
```
❌ Erreur d'arguments: configuration file not found: config.yaml
```
**Solution :** Vérifiez le chemin du fichier de configuration

#### 2. Symbole Invalide
```
❌ Erreur d'arguments: invalid symbol format: invalid-symbol
```
**Solution :** Utilisez des symboles valides (ex: SOLUSDT, ETHUSDT)

#### 3. Timeframe Non Supporté
```
❌ Erreur d'arguments: unsupported timeframe: 2h
```
**Solution :** Utilisez : 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d

#### 4. Limite Mémoire Dépassée
```
⚠️ Memory constraints validation failed, proceeding with caution
```
**Solution :** Augmentez `--memory-limit` ou utilisez moins de symboles

### Débogage

#### Mode Verbose
```bash
./agent-economique --config config/config.yaml --verbose
```

#### Vérification Configuration
```bash
# Test avec configuration minimale
./agent-economique --config config/config.yaml --symbols SOLUSDT --timeframes 1h
```

## 🚀 Optimisations Performance

### Recommandations Générales
1. **Réseau :** Connexion stable > 10 Mbps
2. **Mémoire :** 8GB+ recommandé
3. **Stockage :** SSD préférable
4. **Parallélisme :** max_concurrent = nombre de cœurs CPU

### Paramètres de Performance
```yaml
binance_data:
  downloader:
    max_concurrent: 8      # Augmenter si bonne connexion
    timeout: "5m"          # Réduire pour connexions rapides
  streaming:
    buffer_size: 131072    # 128KB pour gros fichiers
    max_memory_mb: 1024    # Ajuster selon RAM disponible
```

### Mode Streaming pour Gros Volumes
```bash
# Pour traiter plusieurs mois de données
./agent-economique \
  --config config/config.yaml \
  --mode streaming \
  --memory-limit 512 \
  --symbols SOLUSDT,ETHUSDT,BTCUSDT
```

## 🔐 Sécurité et Bonnes Pratiques

### Permissions Fichiers
```bash
chmod 600 config/config.yaml  # Configuration sécurisée
chmod 755 data/              # Dossier de données
```

### Validation des Données
- **Checksums automatiques** : Validation SHA256
- **Continuité temporelle** : Vérification des gaps
- **Format des données** : Validation CSV structure

### Sauvegarde
```bash
# Sauvegarde cache avant gros téléchargement
tar -czf backup_cache_$(date +%Y%m%d).tar.gz data/binance/
```

## 📞 Support et Dépannage

### Logs de Débogage
Les logs détaillés sont disponibles avec `--verbose`

### Issues Communes
1. **Timeouts réseau** → Augmenter `timeout` dans config
2. **Mémoire insuffisante** → Mode streaming + limite mémoire
3. **Espace disque plein** → Nettoyer cache ou augmenter limite

### Performance Monitoring
Avec `--enable-metrics`, l'application affiche :
- Vitesse de téléchargement (MB/s)
- Utilisation mémoire (MB)
- Taux d'erreur réseau
- Statistiques de cache

---

**🎯 L'Application CLI Agent Économique est maintenant prête pour un usage professionnel avec téléchargement direct multi-timeframes !**
