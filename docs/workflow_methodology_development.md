# Méthodologie de développement - Idée vers Implémentation

**Version:** 0.1  
**Statut:** Méthodologie standard  
**Scope:** Processus complet transformation idée → implémentation cohérente

## 🎯 **Vue d'ensemble**

Cette méthodologie définit le processus systématique pour transformer une idée en implémentation fonctionnelle, en garantissant la cohérence avec l'écosystème documentaire et architectural existant de l'agent économique.

**Objectifs :**
- Maintenir la cohérence architecturale
- Respecter les contraintes techniques Go
- Assurer la qualité via tests et validation
- Préserver la traçabilité des décisions

## 🔄 **Workflow en 6 phases**

### **Phase 1 : Clarification et Cadrage**

#### **1.1 Définir l'idée précisément**
- **Problématique** : Quel problème l'idée résout-elle ?
- **Domaine** : Architecture, stratégie, données, configuration, contraintes ?
- **Impact** : Comment s'intègre-t-elle dans l'agent économique ?
- **Priorité** : Critique, importante, amélioration ?

#### **1.2 Valider la cohérence globale**
**Documents de référence obligatoires :**
- `architecture_agent_general.md` → L'idée s'intègre-t-elle dans la vision ?
- `constraints_development_go.md` → Respecte-t-elle les contraintes (500 lignes, tests, etc.) ?
- `strategy_macd_cci_dmi_pure.md` → Compatible avec la stratégie MACD/CCI/DMI ?
- `data_specification_binance.md` → Cohérente avec les données Binance Vision ?
- `constraints_risk_management.md` → Respecte-t-elle les règles de risque ?

**Sortie Phase 1 :** ✅ Idée cadrée et validée conceptuellement

---

### **Phase 2 : Spécification Technique**

#### **2.1 Créer la spécification détaillée**
**Nommage selon `FILE_ORGANIZATION_RULES.md` :**
```
[DOMAINE]_specification_[SPECIFIQUE].md
```

**Contenu obligatoire :**
- Vue d'ensemble et objectifs
- Interfaces et signatures de fonctions
- Structures de données
- Contraintes techniques spécifiques
- Dépendances avec modules existants
- Critères d'acceptation

#### **2.2 Décomposer en User Stories**
**Modèle des user stories existantes :**
- Format : "En tant que... je veux... afin de..."
- Critères d'acceptation SMART
- Estimation en points (Fibonacci : 1, 2, 3, 5, 8, 13)
- Dépendances identifiées
- Sprint assigné

**Sortie Phase 2 :** ✅ Spécification technique + User Stories

---

### **Phase 3 : Conception Détaillée**

#### **3.1 Créer le workflow d'implémentation**
**Modèle des 3 workflows existants :**
- **Workflow Infrastructure** : Fondations et composants de base
- **Workflow Pipeline** : Traitement et transformation des données  
- **Workflow Intégration** : Connexion avec l'agent et stratégie

**Structure standard :**
```markdown
# Workflow N: [NOM]
## Composants principaux
### N. [Composant]
**Fichier:** `internal/[chemin]/[nom].go`
**Responsabilités:** [liste]
**Fonctions clés:** [signatures Go]
```

#### **3.2 Planifier les tests**
**Documentation obligatoire selon modèle :**
- Fichier : `tests/[module]_test_documentation.md`
- Pour chaque fonction : logique à tester, conditions d'entrée, résultats attendus
- Tests de performance si applicable
- Tests d'intégration end-to-end

**Sortie Phase 3 :** ✅ Workflow détaillé + Documentation tests

---

### **Phase 4 : Validation Préalable**

#### **4.1 Review de cohérence documentaire**
**Checklist obligatoire :**
- [ ] Pas de conflit avec `architecture_agent_general.md`
- [ ] Respect `constraints_development_go.md` (500 lignes max, etc.)
- [ ] Compatibilité interfaces existantes
- [ ] Pas de régression sur `strategy_macd_cci_dmi_pure.md`
- [ ] Cohérence avec `data_specification_binance.md`

#### **4.2 Validation utilisateur**
**Présentation pour accord :**
- Spécification technique complète
- User stories avec estimations
- Workflow d'implémentation détaillé
- Impact sur l'existant

**Sortie Phase 4 :** ✅ Validation utilisateur obtenue

---

### **Phase 5 : Implémentation Progressive**

#### **5.1 Suivre l'ordre des workflows**
**Progression obligatoire :**
1. **Infrastructure** (Workflow 1) : Composants de base
2. **Pipeline** (Workflow 2) : Traitement des données
3. **Intégration** (Workflow 3) : Connexion avec l'agent

#### **5.2 Respecter les contraintes Go**
**Pour chaque fichier :**
- Max 500 lignes de code
- Une fonction = un objectif unique
- Tests unitaires obligatoires
- Éviter les pointeurs sauf nécessité performance
- Documentation Godoc pour fonctions publiques

#### **5.3 Validation continue**
**Après chaque fonction :**
- Tests unitaires passent (couverture > 90%)
- Review de code selon contraintes Go
- Intégration sans régression

**Sortie Phase 5 :** ✅ Code implémenté et testé

---

### **Phase 6 : Documentation et Intégration**

#### **6.1 Mettre à jour la documentation**
**Obligations :**
- Changelog : `v[VERSION]_[DATE]_[SCOPE].md`
- README si impact sur architecture globale
- Mise à jour user stories (status → completed)

#### **6.2 Tests d'intégration finale**
**Validation end-to-end :**
- Tests d'intégration avec données Binance réelles
- Validation performance selon critères
- Tests robustesse (pannes, corruptions)
- Compatibilité stratégie MACD/CCI/DMI

**Sortie Phase 6 :** ✅ Fonctionnalité intégrée et documentée

---

## ✅ **Points de contrôle obligatoires**

### **Avant implémentation :**
- [ ] **Cohérence architecturale** : Consulté `architecture_agent_general.md`
- [ ] **Contraintes Go respectées** : Max 500 lignes, tests obligatoires
- [ ] **Stratégie compatible** : Pas d'impact négatif sur MACD/CCI/DMI
- [ ] **Données cohérentes** : Compatible avec Binance Vision streaming
- [ ] **Validation utilisateur** : Accord explicite obtenu
- [ ] **Tests planifiés** : Documentation complète créée

### **Pendant implémentation :**
- [ ] **Workflow suivi** : Ordre infrastructure → pipeline → intégration
- [ ] **Fonctions testées** : Tests unitaires après chaque fonction
- [ ] **Review de code** : Contraintes Go validées
- [ ] **Pas de régression** : Tests existants passent toujours

### **Après implémentation :**
- [ ] **Documentation à jour** : Changelogs et README mis à jour
- [ ] **Tests end-to-end** : Validation complète fonctionnelle
- [ ] **Performance validée** : Critères respectés
- [ ] **User stories complètes** : Status updated

---

## 📋 **Templates et exemples**

### **Template spécification :**
```markdown
# [Domaine] - Spécification [Fonctionnalité]

**Version:** 0.1
**Statut:** Spécification technique
**Scope:** [Description courte]

## Vue d'ensemble
[Problématique résolue]

## Interfaces
[Signatures Go des fonctions]

## Structures de données
[Types Go définis]

## Contraintes spécifiques
[Limitations et exigences]

## Intégration
[Comment s'intègre avec l'existant]
```

### **Template user story :**
```markdown
# User Story X: [Titre]

**Epic:** [Groupe fonctionnel]
**Priorité:** Haute/Moyenne/Basse
**Estimation:** X points
**Sprint:** N

## Description
> **En tant qu'** [acteur]
> **Je veux** [action]
> **Afin de** [bénéfice]

## Critères d'acceptation
### ✅ [Critère 1]
- **ÉTANT DONNÉ** [contexte]
- **QUAND** [action]
- **ALORS** [résultat attendu]
```

---

## 🚀 **Exemple d'application**

### **Idée :** "Ajouter indicateur RSI à la stratégie"

**Phase 1 - Cadrage :**
- Domaine : Strategy (filtrage supplémentaire)
- Impact : Extension de `strategy_macd_cci_dmi_pure.md`
- Cohérence : Compatible avec logique existante

**Phase 2 - Spécification :**
- Fichier : `strategy_specification_rsi_integration.md`
- User Story : "En tant qu'agent, je veux utiliser RSI pour filtrer les signaux MACD"

**Phase 3 - Conception :**
- Workflow : Extension Workflow 3 (intégration agent)
- Tests : `tests/rsi_integration_test_documentation.md`

**Phase 4 - Validation :**
- Review : Pas de conflit avec MACD/CCI/DMI
- Accord utilisateur : Obtenu

**Phase 5 - Implémentation :**
- Code : Fonction `calculateRSI()` avec tests
- Intégration : Module indicateurs existant

**Phase 6 - Documentation :**
- Changelog : `v0.2.0_2025-10-30_rsi_integration.md`
- Tests : End-to-end avec données réelles

**Résultat :** RSI intégré sans régression, documentation cohérente, tests passants.

---

## 🎯 **Checklist de validation finale**

### **Documentation cohérente :**
- [ ] Nommage selon `FILE_ORGANIZATION_RULES.md`
- [ ] Contenu mono-préoccupation
- [ ] Références croisées correctes
- [ ] Changelog à jour

### **Architecture respectée :**
- [ ] Compatible `architecture_agent_general.md`
- [ ] Contraintes Go `constraints_development_go.md` respectées
- [ ] Pas d'impact négatif sur composants existants

### **Qualité assurée :**
- [ ] Tests unitaires > 90% couverture
- [ ] Tests d'intégration passants
- [ ] Performance selon critères
- [ ] Code review validée

### **Processus suivi :**
- [ ] Workflow en 6 phases respecté
- [ ] Points de contrôle validés
- [ ] Validation utilisateur obtenue
- [ ] Traçabilité complète

---

**Cette méthodologie garantit la cohérence, la qualité et la maintenabilité de l'agent économique tout en préservant la flexibilité d'évolution.**
