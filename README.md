# Agent Économique de Trading - Module Binance Vision

[![Version](https://img.shields.io/badge/version-1.3.0-blue.svg)](docs/change_log/CHANGELOG.md)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![Licence](https://img.shields.io/badge/licence-MIT-green.svg)](LICENSE)

> Module multi-exchanges de trading avec précision 100% des indicateurs techniques (Gate.io + Binance + BingX) et stratégie MACD/CCI/DMI.


L'Agent Économique est un système modulaire d'exécution et de décision de trading, piloté par un contexte unifié. Ce repository contient spécifiquement le **module de téléchargement des données Binance Vision** qui alimente la stratégie de trading basée sur les indicateurs MACD, CCI et DMI.

**📋 Architecture complète**: [`architecture_agent_general.md`](docs/architecture_agent_general.md)  
**⚙️ Configuration**: [`config_strategy_parameters.md`](docs/config_strategy_parameters.md)  
**🚫 Contraintes Go**: [`constraints_development_go.md`](docs/constraints_development_go.md)  
**🖥️ Guide CLI**: [`guide_utilisation_cli.md`](docs/guide_utilisation_cli.md)  
**🎯 Contraintes Signaux Scalping**: [`CONTRAINTES_SIGNAUX_SCALPING.md`](docs/CONTRAINTES_SIGNAUX_SCALPING.md) - **6 contraintes validation** 🔴  
**📋 Changelog Projet**: [`CHANGELOG.md`](CHANGELOG.md) - **v1.2.0 : Correction critique synchronisation** 🔴  
**📋 Changelog DevOps**: [`devops/CHANGELOG.md`](devops/CHANGELOG.md) - Infrastructure  
**📋 Changelog v1.3.0**: [`CHANGELOG_v1.3.0.md`](docs/change_log/CHANGELOG_v1.3.0.md) - **Précision Binance 100%** ⭐  
**📋 Changelog v1.2.0**: [`CHANGELOG_v1.2.0.md`](docs/change_log/CHANGELOG_v1.2.0.md) - **Précision Gate.io 100%**  
**📋 Changelog v1.1.0**: [`CHANGELOG_v1.1.0.md`](docs/change_log/CHANGELOG_v1.1.0.md)  
**📋 Changelog v1.0.1**: [`CHANGELOG_v1.0.1.md`](docs/change_log/CHANGELOG_v1.0.1.md)  
**📋 Changelog v1.0.0**: [`CHANGELOG_v1.0.0.md`](docs/change_log/CHANGELOG_v1.0.0.md)

### **Données traitées**
- **Paires**: SOLUSDT, SUIUSDT, ETHUSDT
- **Période**: 2023-06-01 au 29/06/2025-06-30
- **Timeframes**: 5m, 15m, 1h, 4h (klines) + trades tick-by-tick
- **Source**: [Binance Data Vision](https://data.binance.vision) (futures USDT-M)

### **Exchanges supportés** ⭐ **v1.3.0**
- **Gate.io Futures** : Précision 100% indicateurs (v1.2.0+)  
- **Binance Futures** : Précision 100% indicateurs (v1.3.0+)  
- **BingX** : Intégration existante  

### **Indicateurs techniques validés** ⭐ **Précision 100%**
- **MFI (Money Flow Index)** : Période 14, zones surachat/survente  
- **MACD (12,26,9)** : Croisements EMA, histogramme momentum  
- **CCI (Commodity Channel Index)** : Période 20, zones extrêmes  
- **DMI (Directional Movement Index)** : DI+/DI- + ADX, tendance/force  
- **Stochastic (%K=14, %D=3)** : Oscillateur momentum, lissage SMA  

**📋 Documentation précision** : [`binance_precision_guide.md`](docs/indicateurs/binance_precision_guide.md) | [`gateio_mfi_precision_guide.md`](docs/indicateurs/gateio_mfi_precision_guide.md)

### **Stratégies implémentées**

#### **⚡ Applications Scalping Live** (v1.2.0) 🔴 **CORRIGÉ**
- **scalping_live_bybit**: Trading live 5m Bybit (déployé production)
- **scalping_live_gateio**: Trading live 5m Gate.io
- **scalping_engine**: Moteur de backtesting scalping

**Triple Extrême Flexible** :
- CCI, MFI, Stochastic en zones extrêmes (N-1 OU N-2)
- **Synchronisation obligatoire** : Les 3 indicateurs bougent ensemble
- Croisement stochastique + validation bougie + volume conditionné
- **6 contraintes** documentées : [`CONTRAINTES_SIGNAUX_SCALPING.md`](docs/CONTRAINTES_SIGNAUX_SCALPING.md)

**Correction critique v1.2.0** 🔴 :
- ✅ Contrainte de synchronisation ajoutée (absente auparavant)
- ✅ Prévient signaux avec divergences indicateurs
- ✅ Triple extrême flexible (chaque indicateur N-1 ou N-2)
- ✅ Cohérence directionnelle garantie (SURACHAT→SHORT, SURVENTE→LONG)

#### **🎯 Stratégie STOCH/MFI/CCI** (v1.1.0+) ⭐ **NOUVEAU**
- **Stochastic(14,3,3)**: Oscillateur principal avec zones extrêmes
- **MFI(14)**: Money Flow Index pour validation volume  
- **CCI(14)**: Commodity Channel Index pour confirmation momentum
- **Signaux**: Triple validation (Premium) ou double validation (Minimal)
- **Gestion**: Money management tick-by-tick adaptatif avec états monitoring
- **Protection**: Early exit sur triple inverse + trailing dynamique

#### **📈 Stratégie MACD/CCI/DMI** (Héritée)
- **MACD(12,26,9)**: Signal principal de croisement
- **CCI(14)**: Zones extrêmes survente/surachat  
- **DMI(14)**: Analyse tendance/contre-tendance
- **Gestion**: Trailing stop dynamique + sortie anticipée

**📋 Documentation stratégies**: [`strategy_macd_cci_dmi_pure.md`](docs/strategy_macd_cci_dmi_pure.md) | [`workflow/09_strategy_stoch_mfi_cci.md`](docs/workflow/09_strategy_stoch_mfi_cci.md)

## 🧪 **Tests de Validation**

### **Validation précision indicateurs** ⭐ **v1.3.0**
```bash
# Validation Binance Futures (tous indicateurs)
go run cmd/indicators_validation/all_binance_validation.go

# Validation individuelle Binance
go run cmd/indicators_validation/mfi_binance_validation.go
go run cmd/indicators_validation/macd_binance_validation.go
go run cmd/indicators_validation/cci_binance_validation.go
go run cmd/indicators_validation/dmi_binance_validation.go
go run cmd/indicators_validation/stoch_binance_validation.go

# Validation Gate.io (référence v1.2.0)
go run cmd/indicators_validation/mfi_tv_standard_validation.go
```

### **Résultats attendus**
- **300 klines** par exchange (futures perpétuels)
- **5 dernières valeurs** affichées pour validation
- **Précision 100%** conforme TradingView
- **Volume SOL** correct (base currency)

**📋 Documentation validation** : [`binance_precision_guide.md`](docs/indicateurs/binance_precision_guide.md)

## 📚 **Documentation**

### **🧭 Navigation de la documentation**
**[📖 Guide de Navigation Complet](docs/NAVIGATION.md)** - **Commencer ici pour s'orienter**

### **Implémentation**
- **Workflows**: 3 phases progressives ([Infrastructure](docs/workflow/01_infrastructure_base.md) → [Pipeline](docs/workflow/02_pipeline_donnees.md) → [Intégration](docs/workflow/03_integration_agent.md))
- **User Stories**: 5 stories détaillées ([`user_stories/`](docs/user_stories/))
- **Tests**: Documentation complète ([`tests/`](docs/tests/))
- **Historique**: [Changelogs versionnés](docs/change_log/)

## 🚀 **Démarrage rapide**

### **Installation**
```bash
git clone <repository-url>
cd agent_economique_stable
go mod tidy
go build -o agent-economique ./cmd/agent/
```

### **Utilisation CLI** ⭐
```bash
# Utilisation basique
./agent-economique --config config/config.yaml

# Symboles et timeframes spécifiques
./agent-economique --config config/config.yaml --symbols SOLUSDT --timeframes 1h

# Mode téléchargement seulement
./agent-economique --config config/config.yaml --mode download-only

# Mode streaming économie mémoire
./agent-economique --config config/config.yaml --mode streaming --memory-limit 128
```

### **Configuration**
```yaml
# config/config.yaml
binance_data:
  cache_root: "data/binance"
  symbols: ["SOLUSDT", "SUIUSDT", "ETHUSDT"]
  timeframes: ["5m", "15m", "1h", "4h"]

# Configuration stratégie (v1.1.0+)
strategy:
  name: "STOCH_MFI_CCI"  # ou "MACD_CCI_DMI"
  position_management:
    enable_dynamic_adjustments: true
    triple_inverse_early_exit: true

# Section CLI
cli:
  execution_mode: "default"
  memory_limit_mb: 512
  verbose: false
  enable_metrics: true
```

### **Tests** (organisation mixte, 95%+ couverture)
```bash
# Tests unitaires par module (white box)
go test -cover ./internal/engine                # Tests Engine (39.1% couverture)
go test -cover ./internal/indicators           # Tests Indicators (47.0% couverture)
go test -cover ./internal/strategies/...       # Tests Stratégies (100% couverture)
go test -cover ./internal/...                  # Tous modules internes

# Tests fonctionnels centralisés (black box)
go test ./tests -v                              # Tests Binance/CLI complets
go test ./tests -cover                          # Avec couverture globale
go test ./tests -run="TestCLI" -v              # Tests CLI spécifiques

# Tests spécifiques stratégies (v1.1.0+)
go test ./internal/strategies/stoch_mfi_cci/... -v  # 18 tests STOCH/MFI/CCI
```

**📋 Guide complet**: [`guides/development_guide.md`](docs/guides/development_guide.md)

## 📊 **Performance & Monitoring**

**Contraintes**: Streaming sans accumulation, <500ms latence, >80% cache hit rate  
**Monitoring**: Métriques temps réel, logs structurés JSON  
**📈 Guide complet**: [`guides/performance_monitoring.md`](docs/guides/performance_monitoring.md)

## 🤝 **Contribution**

**Processus**: Méthodologie 6 phases → Contraintes Go → Tests >90% → Validation utilisateur  
**📋 Guide complet**: [`guides/development_guide.md`](docs/guides/development_guide.md)

## 📄 **Licence**

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 🔗 **Liens utiles**

- [Binance Data Vision](https://data.binance.vision) - Source des données historiques
- [Documentation Go](https://golang.org/doc/) - Référence langage
- [Keep a Changelog](https://keepachangelog.com/fr/) - Format changelog
- [Semantic Versioning](https://semver.org/lang/fr/) - Versioning sémantique

---

## 🎯 **État du Projet**

**Version actuelle**: 1.1.0 (Stratégie STOCH/MFI/CCI) - ⭐ **NOUVELLE STRATÉGIE MAJEURE**  
**Statut**: ✅ **Production Ready** - Architecture multi-stratégies opérationnelle  
**Prochaine étape**: v1.2.0 - Export multi-format, interface web, optimisations batch  
**Maintenance**: Active - Évolutions stratégies et performance selon besoins

### **🚀 Nouvelles fonctionnalités v1.1.0**
- ✅ **Stratégie STOCH/MFI/CCI** : Volume + momentum + oscillateur
- ✅ **Monitoring tick-by-tick** : États adaptatifs (NORMAL → STOCH_INVERSE → TRIPLE_INVERSE)
- ✅ **Money management dynamique** : Ajustements temps réel selon inverse zones
- ✅ **Architecture multi-stratégies** : Support simultané MACD/CCI/DMI + STOCH/MFI/CCI
- ✅ **18 nouveaux tests** : Validation complète behavioral MM et signal generation
- ✅ **Protection avancée** : Early exit sur triple inverse + profit minimum

### **🏆 Accomplissements cumulés**
- ✅ **57+ tests unitaires** (95%+ couverture) + **18 tests stratégie STOCH/MFI/CCI**
- ✅ **Application CLI** multi-modes avec cache intelligent sub-milliseconde
- ✅ **2 stratégies complètes** : MACD/CCI/DMI + STOCH/MFI/CCI (configuration YAML)
- ✅ **Engine optimisé** : Monitoring sélectif, tick-by-tick adaptatif, trailing dynamique
- ✅ **Architecture évolutive** : Base extensible pour futures stratégies
- ✅ **Rétrocompatibilité** : Migration transparente, configuration préservée

**📋 Nouvelle stratégie v1.1.0**: [`CHANGELOG_v1.1.0.md`](docs/change_log/CHANGELOG_v1.1.0.md)  
**📋 Corrections v1.0.1**: [`CHANGELOG_v1.0.1.md`](docs/change_log/CHANGELOG_v1.0.1.md)  
**📋 Release initiale**: [`CHANGELOG_v1.0.0.md`](docs/change_log/CHANGELOG_v1.0.0.md)
