# 🧭 Guide de Navigation - Documentation Agent Économique

**Version:** 1.2  
**Dernière mise à jour:** 2025-11-06  
**Objectif:** Orienter la lecture selon les besoins et priorités

**🔴 NOUVEAU v1.2.0:** Documentation contraintes génération signaux scalping ([`CONTRAINTES_SIGNAUX_SCALPING.md`](CONTRAINTES_SIGNAUX_SCALPING.md))

## 🚀 **Démarrage rapide - Parcours essentiel**

### **Pour comprendre le projet (20 min) :**
1. [`architecture_agent_general.md`](architecture_agent_general.md) - **Vision et architecture globale**
2. [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) - **Stratégie MACD/CCI/DMI (héritée)**
3. [`workflow/09_strategy_stoch_mfi_cci.md`](workflow/09_strategy_stoch_mfi_cci.md) - **Stratégie STOCH/MFI/CCI (v1.1.0+)** ⭐
4. [`data_specification_binance.md`](data_specification_binance.md) - **Données utilisées**

### **Pour développer (30 min) :**
4. [`workflow_methodology_development.md`](workflow_methodology_development.md) - **Comment développer**
5. [`constraints_development_go.md`](constraints_development_go.md) - **Contraintes techniques**
6. [`FILE_ORGANIZATION_RULES.md`](FILE_ORGANIZATION_RULES.md) - **Organisation documentation**

## 🏆 **Hiérarchie de priorité**

### **🔴 Niveau 1 : CRITIQUE** (Fondations du système)
**À lire en premier - Requis pour comprendre le projet**

| Document | Description | Temps |
|----------|-------------|--------|
| [`architecture_agent_general.md`](architecture_agent_general.md) | Vision, composants, flux principaux | 10 min |
| [`constraints_development_go.md`](constraints_development_go.md) | Standards Go, limites techniques | 8 min |
| [`FILE_ORGANIZATION_RULES.md`](FILE_ORGANIZATION_RULES.md) | Organisation et nommage des fichiers | 5 min |
| [`workflow_methodology_development.md`](workflow_methodology_development.md) | Méthodologie développement complète | 15 min |

### **🟡 Niveau 2 : ESSENTIEL** (Contexte métier)
**Définit QUOI on construit**

| Document | Description | Temps |
|----------|-------------|--------|
| [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) | Stratégie MACD/CCI/DMI (héritée) | 12 min |
| [`workflow/09_strategy_stoch_mfi_cci.md`](workflow/09_strategy_stoch_mfi_cci.md) | **Stratégie STOCH/MFI/CCI (v1.1.0+)** ⭐ | 15 min |
| [`user_stories/09_strategy_stoch_mfi_cci.md`](user_stories/09_strategy_stoch_mfi_cci.md) | **User stories STOCH/MFI/CCI** | 8 min |
| [`data_specification_binance.md`](data_specification_binance.md) | Sources données, formats, structure cache | 10 min |
| [`constraints_risk_management.md`](constraints_risk_management.md) | Règles de risque, money management | 8 min |
| [`config_strategy_parameters.md`](config_strategy_parameters.md) | Paramètres techniques stratégie | 5 min |

### **🟢 Niveau 3 : IMPLÉMENTATION** (Comment construire)
**Workflows séquentiels d'implémentation**

| Document | Description | Temps |
|----------|-------------|--------|
| [`workflow/01_infrastructure_base.md`](workflow/01_infrastructure_base.md) | Cache, téléchargeur, streaming | 15 min |
| [`workflow/02_pipeline_donnees.md`](workflow/02_pipeline_donnees.md) | Parsers, intégration multi-timeframes | 12 min |
| [`workflow/03_integration_agent.md`](workflow/03_integration_agent.md) | Connecteurs engines, contexte unifié | 18 min |

### **🔵 Niveau 4 : DÉTAILS** (Spécifications fines)
**Documentation détaillée pour implémentation**

| Catégorie | Documents | Description |
|-----------|-----------|-------------|
| **User Stories** | [`user_stories/`](user_stories/) | 5 stories détaillées avec critères d'acceptation |
| **Tests** | [`tests/`](tests/) | Documentation logique de test par module |
| **Historique** | [`change_log/`](change_log/) | Changelogs versionnés par composant |

## 📂 **Navigation par catégorie**

### **🏗️ Architecture & Conception**
- [`architecture_agent_general.md`](architecture_agent_general.md) - Vue d'ensemble système
- [`constraints_development_go.md`](constraints_development_go.md) - Standards développement
- [`workflow_methodology_development.md`](workflow_methodology_development.md) - Méthodologie complète

### **📈 Trading & Stratégies**
- [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) - Stratégie MACD/CCI/DMI (héritée)
- [`workflow/09_strategy_stoch_mfi_cci.md`](workflow/09_strategy_stoch_mfi_cci.md) - **Stratégie STOCH/MFI/CCI (v1.1.0+)** ⭐
- [`CONTRAINTES_SIGNAUX_SCALPING.md`](CONTRAINTES_SIGNAUX_SCALPING.md) - **6 contraintes génération signaux scalping** 🔴 **CRITIQUE v1.2.0**
- [`user_stories/09_strategy_stoch_mfi_cci.md`](user_stories/09_strategy_stoch_mfi_cci.md) - User stories STOCH/MFI/CCI
- [`tests/strategy_stoch_mfi_cci_test_plan.md`](tests/strategy_stoch_mfi_cci_test_plan.md) - Plan de test STOCH/MFI/CCI
- [`config_strategy_parameters.md`](config_strategy_parameters.md) - Configuration stratégies
- [`constraints_risk_management.md`](constraints_risk_management.md) - Gestion des risques

### **💾 Données & Pipeline**
- [`data_specification_binance.md`](data_specification_binance.md) - Données Binance Vision
- [`workflow/02_pipeline_donnees.md`](workflow/02_pipeline_donnees.md) - Traitement données
- [`user_stories/03_lecteur_streaming_performance.md`](user_stories/03_lecteur_streaming_performance.md) - Streaming ZIP

### **📋 Guides Pratiques**
- [`guides/development_guide.md`](guides/development_guide.md) - Guide de développement complet
- [`guides/performance_monitoring.md`](guides/performance_monitoring.md) - Performance et monitoring

### **📊 Indicateurs Techniques & Validation**
- [`indicateurs/`](indicateurs/) - **Spécifications et recherche indicateurs** ⭐ **NOUVEAU**
  - [`indicateurs/binance_precision_guide.md`](indicateurs/binance_precision_guide.md) - Guide précision tous indicateurs Binance ⭐ **NOUVEAU**
  - [`indicateurs/gateio_mfi_precision_guide.md`](indicateurs/gateio_mfi_precision_guide.md) - Guide précision MFI Gate.io
  - [`indicateurs/indicateur_precision_rules.md`](indicateurs/indicateur_precision_rules.md) - Règles précision 100%
  - [`indicateurs/mfi_tradingview_research.md`](indicateurs/mfi_tradingview_research.md) - Recherche MFI TradingView
  - [`indicateurs/macd_tradingview_research.md`](indicateurs/macd_tradingview_research.md) - Recherche MACD TradingView
  - [`indicateurs/cci_tradingview_research.md`](indicateurs/cci_tradingview_research.md) - Recherche CCI TradingView
  - [`indicateurs/dmi_tradingview_research.md`](indicateurs/dmi_tradingview_research.md) - Recherche DMI TradingView
  - [`indicateurs/stoch_tradingview_research.md`](indicateurs/stoch_tradingview_research.md) - Recherche Stochastic TradingView
  - [`indicateurs/ema_tradingview_research.md`](indicateurs/ema_tradingview_research.md) - Recherche EMA TradingView
  - [`indicateurs/rma_tradingview_research.md`](indicateurs/rma_tradingview_research.md) - Recherche RMA TradingView
  - [`indicateurs/sma_tradingview_research.md`](indicateurs/sma_tradingview_research.md) - Recherche SMA TradingView

### **🔄 Workflows & Processus**
- [`workflow/01_infrastructure_base.md`](workflow/01_infrastructure_base.md) - Infrastructure
- [`workflow/02_pipeline_donnees.md`](workflow/02_pipeline_donnees.md) - Pipeline données
- [`workflow/03_integration_agent.md`](workflow/03_integration_agent.md) - Intégration finale

### **⚙️ Configuration & Paramètres**
- [`config_strategy_parameters.md`](config_strategy_parameters.md) - Paramètres MACD/CCI/DMI
- Voir aussi: sections configuration dans chaque workflow

### **🧪 Tests & Validation**
- [`tests/cache_module_test_documentation.md`](tests/cache_module_test_documentation.md) - Tests cache
- [`tests/downloader_module_test_documentation.md`](tests/downloader_module_test_documentation.md) - Tests téléchargeur
- [`tests/streaming_module_test_documentation.md`](tests/streaming_module_test_documentation.md) - Tests streaming
- [`tests/parsers_module_test_documentation.md`](tests/parsers_module_test_documentation.md) - Tests parsers
- [`tests/connectors_module_test_documentation.md`](tests/connectors_module_test_documentation.md) - Tests connecteurs
- **Tests indicateurs TV Standard** - [`../cmd/indicators_validation/`](../cmd/indicators_validation/) ⭐ **NOUVEAU**
  - [`../cmd/indicators_validation/mfi_tv_standard_validation.go`](../cmd/indicators_validation/mfi_tv_standard_validation.go) - Validation MFI précision 100%
  - [`../cmd/indicators_validation/macd_gateio_application.go`](../cmd/indicators_validation/macd_gateio_application.go) - Validation MACD Gate.io
  - [`../cmd/indicators_validation/cci_gateio_application.go`](../cmd/indicators_validation/cci_gateio_application.go) - Validation CCI Gate.io
  - [`../cmd/indicators_validation/dmi_gateio_application.go`](../cmd/indicators_validation/dmi_gateio_application.go) - Validation DMI Gate.io
  - [`../cmd/indicators_validation/stoch_gateio_application.go`](../cmd/indicators_validation/stoch_gateio_application.go) - Validation Stochastic Gate.io

## 🎯 **Parcours par rôle**

### **👨‍💼 Chef de projet / Product Owner**
1. [`architecture_agent_general.md`](architecture_agent_general.md) - Vision globale
2. [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) - Stratégie métier
3. [`user_stories/`](user_stories/) - Exigences fonctionnelles
4. [`change_log/CHANGELOG.md`](change_log/CHANGELOG.md) - État d'avancement

### **👨‍💻 Développeur débutant sur le projet**
1. [`architecture_agent_general.md`](architecture_agent_general.md) - Comprendre l'architecture
2. [`guides/development_guide.md`](guides/development_guide.md) - Guide de développement complet
3. [`constraints_development_go.md`](constraints_development_go.md) - Standards à respecter
4. [`workflow/01_infrastructure_base.md`](workflow/01_infrastructure_base.md) - Commencer par l'infrastructure

### **👨‍💻 Développeur d'indicateurs techniques**
1. [`indicateurs/gateio_mfi_precision_guide.md`](indicateurs/gateio_mfi_precision_guide.md) - Maîtriser la précision des données
2. [`indicateurs/indicateur_precision_rules.md`](indicateurs/indicateur_precision_rules.md) - Règles de précision 100%
3. [`indicateurs/mfi_tradingview_research.md`](indicateurs/mfi_tradingview_research.md) - Spécifications MFI
4. [`../cmd/indicators_validation/`](../cmd/indicators_validation/) - Tests de validation fonctionnels

### **👨‍💻 Développeur expérimenté**
1. [`workflow_methodology_development.md`](workflow_methodology_development.md) - Processus de développement
2. [`constraints_development_go.md`](constraints_development_go.md) - Contraintes techniques
3. Workflows séquentiels [`workflow/`](workflow/) selon besoin
4. [`tests/`](tests/) - Documentation des tests

### **🧪 Testeur / QA**
1. [`architecture_agent_general.md`](architecture_agent_general.md) - Comprendre le système
2. [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) - Logique métier à valider
3. [`user_stories/`](user_stories/) - Critères d'acceptation
4. [`tests/`](tests/) - Logique de test détaillée

### **📊 Analyste / Trader**
1. [`CONTRAINTES_SIGNAUX_SCALPING.md`](CONTRAINTES_SIGNAUX_SCALPING.md) - **6 contraintes validation signaux** 🔴 **CRITIQUE**
2. [`workflow/09_strategy_stoch_mfi_cci.md`](workflow/09_strategy_stoch_mfi_cci.md) - **Stratégie STOCH/MFI/CCI (v1.1.0+)** ⭐
3. [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md) - Stratégie MACD/CCI/DMI (héritée)
4. [`config_strategy_parameters.md`](config_strategy_parameters.md) - Paramètres configurables
5. [`constraints_risk_management.md`](constraints_risk_management.md) - Rules de risque
6. [`data_specification_binance.md`](data_specification_binance.md) - Données utilisées

## 🔍 **Recherche par mots-clés**

### **Architecture**
→ [`architecture_agent_general.md`](architecture_agent_general.md), [`constraints_development_go.md`](constraints_development_go.md)

### **Scalping** 🔴 **CRITIQUE v1.2.0**
→ [`CONTRAINTES_SIGNAUX_SCALPING.md`](CONTRAINTES_SIGNAUX_SCALPING.md) - **6 contraintes validation signaux**

### **STOCH/MFI/CCI** ⭐ **NOUVEAU v1.1.0**
→ [`workflow/09_strategy_stoch_mfi_cci.md`](workflow/09_strategy_stoch_mfi_cci.md), [`user_stories/09_strategy_stoch_mfi_cci.md`](user_stories/09_strategy_stoch_mfi_cci.md)

### **MACD/CCI/DMI** (Héritée)
→ [`strategy_macd_cci_dmi_pure.md`](strategy_macd_cci_dmi_pure.md), [`config_strategy_parameters.md`](config_strategy_parameters.md)

### **Binance/Données**
→ [`data_specification_binance.md`](data_specification_binance.md), [`workflow/02_pipeline_donnees.md`](workflow/02_pipeline_donnees.md)

### **Cache/Téléchargement**
→ [`workflow/01_infrastructure_base.md`](workflow/01_infrastructure_base.md), [`user_stories/01_cache_intelligent_local.md`](user_stories/01_cache_intelligent_local.md)

### **Streaming/Performance**
→ [`user_stories/03_lecteur_streaming_performance.md`](user_stories/03_lecteur_streaming_performance.md), [`tests/streaming_module_test_documentation.md`](tests/streaming_module_test_documentation.md)

### **Tests**
→ [`tests/`](tests/) (tous les modules), [`workflow_methodology_development.md`](workflow_methodology_development.md)

### **Indicateurs Techniques** ⭐ **NOUVEAU**
→ [`indicateurs/`](indicateurs/) (spécifications), [`../cmd/indicators_validation/`](../cmd/indicators_validation/) (tests)

### **Précision Données Gate.io**
→ [`indicateurs/gateio_mfi_precision_guide.md`](indicateurs/gateio_mfi_precision_guide.md), [`indicateurs/indicateur_precision_rules.md`](indicateurs/indicateur_precision_rules.md)

### **Précision Données Binance** ⭐ **NOUVEAU**
→ [`indicateurs/binance_precision_guide.md`](indicateurs/binance_precision_guide.md), [`../cmd/indicators_validation/all_binance_validation.go`](../cmd/indicators_validation/all_binance_validation.go)

### **Configuration**
→ [`config_strategy_parameters.md`](config_strategy_parameters.md), sections config des workflows

## 📋 **Checklist de lecture**

### **Compréhension générale (✅ cocher au fur et à mesure) :**
- [ ] Architecture globale comprise (`architecture_agent_general.md`)
- [ ] **Stratégie STOCH/MFI/CCI comprise** (`workflow/09_strategy_stoch_mfi_cci.md`) ⭐ **v1.1.0**
- [ ] Stratégie MACD/CCI/DMI assimilée (`strategy_macd_cci_dmi_pure.md`)
- [ ] Contraintes Go connues (`constraints_development_go.md`)
- [ ] Méthodologie de développement comprise (`workflow_methodology_development.md`)

### **Prêt pour implémentation :**
- [ ] Workflow 1 étudié (Infrastructure)
- [ ] Workflow 2 étudié (Pipeline)
- [ ] Workflow 3 étudié (Intégration)
- [ ] Tests documentés consultés
- [ ] User stories comprises

## 🆘 **Aide & Support**

### **En cas de confusion :**
1. **Relire** [`architecture_agent_general.md`](architecture_agent_general.md) pour le contexte global
2. **Consulter** [`workflow_methodology_development.md`](workflow_methodology_development.md) pour la méthodologie
3. **Vérifier** [`FILE_ORGANIZATION_RULES.md`](FILE_ORGANIZATION_RULES.md) pour l'organisation

### **Pour contribuer :**
1. **Suivre** la méthodologie dans [`workflow_methodology_development.md`](workflow_methodology_development.md)
2. **Respecter** les contraintes de [`constraints_development_go.md`](constraints_development_go.md)
3. **Nommer** selon [`FILE_ORGANIZATION_RULES.md`](FILE_ORGANIZATION_RULES.md)

---

**💡 Conseil :** Commencez toujours par le **Niveau 1 (CRITIQUE)** avant de plonger dans les détails. La compréhension globale facilite l'assimilation des spécificités techniques.
