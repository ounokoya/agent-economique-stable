# Règles d'organisation et nommage des fichiers documentation

**Version:** 0.1  
**Statut:** Standards documentation  
**Objectif:** Organisation cohérente du contenu selon les préoccupations

## 🎯 **Problème résolu**

**Avant:** Fichiers avec contenu mélangé non identifiable par le nom
- `agent_economique_conception_initiale.md` → Architecture + Contraintes Go + Spéc Binance + Pipeline
- `strategie_macd_cci_dmi.md` → Stratégie + Configuration + Contraintes risque

**Après:** Fichiers focalisés avec nommage explicite

## 📋 **Convention de nommage**

### **Format standard:**
```
[DOMAINE]_[TYPE]_[SPECIFIQUE].md
```

### **Domaines principaux:**
- `architecture_` - Conception et architecture système
- `strategy_` - Stratégies de trading 
- `config_` - Configuration et paramètres
- `constraints_` - Contraintes et règles
- `data_` - Gestion des données
- `workflow_` - Processus et workflows
- `tests_` - Tests et validation

### **Types principaux:**
- `specification_` - Spécifications techniques détaillées
- `rules_` - Règles et standards
- `parameters_` - Paramètres de configuration
- `management_` - Gestion et administration

## 🗂️ **Nouvelle organisation**

### **Fichiers réorganisés:**

| Ancien fichier | Nouveau(x) fichier(s) | Contenu |
|----------------|----------------------|---------|
| `agent_economique_conception_initiale.md` | **Éclaté en 3 fichiers:** | |
| | `architecture_agent_general.md` | ✅ Architecture et vision générale |
| | `constraints_development_go.md` | ✅ Contraintes techniques Go |
| | `data_specification_binance.md` | ✅ Spéc données Binance Vision |
| `strategie_macd_cci_dmi.md` | **Éclaté en 3 fichiers:** | |
| | `strategy_macd_cci_dmi_pure.md` | ✅ Règles pures de trading |
| | `config_strategy_parameters.md` | ✅ Paramètres de configuration |
| | `constraints_risk_management.md` | ✅ Contraintes et gestion risques |

### **Structure finale cohérente:**

```
docs/
├── architecture_agent_general.md           # Vue d'ensemble architecture
├── strategy_macd_cci_dmi_pure.md          # Règles de trading pures
├── data_specification_binance.md          # Spécifications données
├── constraints_development_go.md          # Contraintes développement
├── constraints_risk_management.md         # Contraintes risques  
├── config_strategy_parameters.md          # Configuration stratégie
├── workflow/                               # Workflows d'implémentation
│   ├── 01_infrastructure_base.md
│   ├── 02_pipeline_donnees.md
│   └── 03_integration_agent.md
├── user_stories/                           # Stories utilisateur
│   ├── 01_cache_intelligent_local.md
│   ├── 02_telechargeur_robuste.md
│   ├── 03_lecteur_streaming_performance.md
│   ├── 04_integration_strategie_macd_cci_dmi.md
│   └── 05_monitoring_diagnostics.md
├── tests/                                  # Documentation tests
│   ├── cache_module_test_documentation.md
│   ├── downloader_module_test_documentation.md
│   ├── streaming_module_test_documentation.md
│   ├── parsers_module_test_documentation.md
│   └── connectors_module_test_documentation.md
└── change_log/                             # Historique changements
    ├── CHANGELOG.md
    ├── v0.1.0_2025-10-30_infrastructure_base.md
    └── v0.1.0_2025-10-30_pipeline_donnees.md
```

## ✅ **Avantages de la réorganisation**

### **Clarté**
- **Nom = Contenu** : Le nom du fichier indique exactement son contenu
- **Préoccupation unique** : Chaque fichier traite un seul sujet
- **Navigation facile** : Trouvez rapidement ce que vous cherchez

### **Maintenabilité**
- **Modifications ciblées** : Changement dans un domaine = un seul fichier
- **Évolutivité** : Ajout de nouveaux contenus sans mélange
- **Responsabilités claires** : Qui maintient quoi

### **Réutilisabilité**
- **Références précises** : Pointer vers le bon fichier
- **Import sélectif** : Utiliser seulement ce qui est nécessaire
- **Documentation modulaire** : Combiner selon les besoins

## 🎯 **Règles d'application**

### **Avant d'ajouter du contenu:**
1. **Identifier le domaine** (architecture, strategy, config, etc.)
2. **Identifier le type** (specification, rules, parameters, etc.)  
3. **Vérifier unicité** : Le contenu n'appartient qu'à cette préoccupation
4. **Nommer selon convention** : `[DOMAINE]_[TYPE]_[SPECIFIQUE].md`

### **Si contenu mixte détecté:**
1. **Identifier chaque préoccupation**
2. **Créer fichiers séparés** selon convention
3. **Migrer contenu** vers fichiers appropriés
4. **Supprimer ancien fichier** mélangé
5. **Mettre à jour références**

### **Validation:**
- ✅ Le nom du fichier décrit-il précisément son contenu ?
- ✅ Le contenu appartient-il à une seule préoccupation ?
- ✅ Peut-on modifier ce fichier sans impacter d'autres domaines ?
- ✅ La navigation est-elle intuitive ?

## 📖 **Migration complétée**

**Status:** ✅ **Terminée**  
**Date:** 2025-10-30  
**Fichiers migrés:** 2 fichiers → 6 fichiers focalisés  
**Cohérence:** 100% - chaque fichier a un objectif unique et clair

**Prochaine étape:** Maintenir cette organisation pour tous les nouveaux contenus.
