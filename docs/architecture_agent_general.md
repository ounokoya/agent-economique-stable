# Architecture Agent Économique - Vue d'ensemble Modulaire

## 📋 Vue d'Ensemble

L'agent économique est conçu comme un **système modulaire multi-stratégies** permettant de composer différentes stratégies de trading en combinant des **modules indicateurs réutilisables** avec des **comportements Money Management adaptatifs**.

## 🏗️ Architecture Modulaire

### 🎯 Principe Architectural

**"Séparer l'Infrastructure Invariante des Compositions Stratégiques Variables"**

- **Core System** : Infrastructure stable, réutilisable par toutes stratégies
- **Indicateurs Modules** : Calculs techniques réutilisables, événements standardisés  
- **Stratégies Compositions** : Assemblage indicateurs + MM comportemental spécifique

## 🏛️ MODULES CORE (Invariants)

### 1. **Engine Temporal**
- **Responsabilité** : Orchestrateur principal + cycle temporel
- **Fonctions** : Ticks, barres, coordination modules, event bus
- **Réutilisation** : Identique pour toutes stratégies

### 2. **Infrastructure**
- **Responsabilité** : SDK exchanges, monitoring, configuration, logs
- **Fonctions** : APIs BingX/Binance, health checks, audit trail
- **Réutilisation** : Base commune toutes stratégies

### 3. **Pipeline Données**
- **Responsabilité** : Ingestion, cache, streaming données market
- **Fonctions** : Cache TTL, streaming performant, parsers
- **Réutilisation** : Infrastructure commune toutes stratégies

### 4. **Multi-Comptes**
- **Responsabilité** : Gestion sous-comptes, isolation risques
- **Fonctions** : Création comptes, transferts, permissions
- **Réutilisation** : Service global multi-stratégies

### 5. **Money Management BASE**
- **Responsabilité** : Circuit breakers globaux, limites invariantes
- **Fonctions** : Arrêts -5%/-15%, position sizing base, métriques globales
- **Réutilisation** : Protection commune toutes stratégies

## 📊 MODULES INDICATEURS (Réutilisables)

### 🎯 Principe Indicateurs
- **Un module = Un indicateur** technique standard
- **Événements standardisés** : Interface commune pour stratégies
- **Réutilisation maximale** : MACD utilisé par N stratégies différentes

### 📈 Indicateurs Disponibles
- **MACD Module** : Croisements + événements MACD_CROSS_UP/DOWN
- **CCI Module** : Zones extrêmes + événements CCI_ZONE_INVERSE, CCI_OVERSOLD  
- **DMI Module** : Tendances + événements DMI_TREND_BULLISH, DMI_COUNTER_CROSS
- **RSI Module** : Surachat/survente + événements RSI_OVERSOLD, RSI_DIVERGENCE
- **Bollinger Module** : Squeeze/breakouts + événements BB_SQUEEZE, BB_BREAKOUT
- **EMA/SMA Module** : Croisements moyennes + événements MA_CROSS_UP/DOWN
- **Volume Module** : Anomalies volume + événements VOLUME_SPIKE, VOLUME_CONFIRM

## 🎨 COMPOSITIONS STRATÉGIQUES (Variables)

### 🧩 Stratégie = Assemblage Modulaire
```
Stratégie X = {
    Indicateurs Choisis: [MACD, CCI, DMI]
    + Signal Generator: Logique combinaison → Signal final  
    + MM Comportemental: Réactions événements indicateurs
    + Position Manager: Gestion selon logique stratégie
}
```

### 📊 Exemple : Stratégie MACD/CCI/DMI
- **Indicateurs** : MACD + CCI + DMI modules
- **Signaux** : MACD_CROSS_UP + CCI_OVERSOLD + DMI_TREND → LONG_ENTRY
- **MM Comportemental** : 
  - Trailing stop selon DMI (2% tendance, 1.5% contre-tendance)
  - Ajustements CCI_ZONE_INVERSE, MACD_CROSS_DOWN, DMI_COUNTER_CROSS
  - Sortie anticipée MACD inverse si trailing stop pas positif

## 🔄 FLUX ARCHITECTURAL

### 📡 Event-Driven Architecture
```
1. MarketData → Modules Indicateurs → Événements standardisés
2. Événements Indicateurs → Signal Generators → Signaux stratégie
3. Signaux → MM Comportemental → Décisions trailing stops
4. Décisions MM → Core MM validation → Exécution si limites OK
```

## 🎯 AVANTAGES ARCHITECTURE MODULAIRE

### ✅ **Réutilisabilité Maximale**
- MACD calculé 1 fois → Utilisé par N stratégies
- Core MM → Circuit breakers pour toutes stratégies
- Infrastructure → Base commune (SDK, Engine, Data)

### 🎨 **Flexibilité Comportementale**
- MM adaptatif : Chaque stratégie réagit selon SES indicateurs
- Événements sur mesure : Réactions spécifiques aux signaux choisis
- Compositions infinies : N stratégies avec mêmes modules de base

### 🔧 **Maintenance Simplifiée**
- Bug indicateur : Fix unique pour toutes stratégies l'utilisant
- Amélioration Core : Bénéfice automatique toutes stratégies
- Nouvelle stratégie : Composition modules existants + MM comportemental

## 🚀 WORKFLOW CRÉATION NOUVELLE STRATÉGIE

### 📋 Processus Simplifié
1. **Conception** : Identifier indicateurs + définir MM comportemental
2. **Composition** : Réutiliser modules + créer MM spécifique
3. **Validation** : Tests unitaires + intégration + backtests
4. **Déploiement** : Configuration runtime + monitoring dédié

— Fin Architecture Agent Économique Modulaire —
