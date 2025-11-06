# Workflow Money Management Comportemental - Stratégie MACD/CCI/DMI

## 📋 Vue d'Ensemble

Workflow Money Management comportemental spécifique à la stratégie MACD/CCI/DMI : trailing stops adaptatifs selon événements indicateurs, ajustements dynamiques et sortie anticipée.

---

## 🎯 WORKFLOW TRAILING STOP INITIAL

### 📈 Placement Trailing Stop à l'Ouverture Position

#### Étapes Détaillées :

1. **Réception Signal Trading**
   - Signal MACD/CCI/DMI validé reçu
   - Type de signal identifié (tendance/contre-tendance DMI)
   - Confidence > 0.7 confirmée

2. **Détermination Type Trailing Stop**
   - **Signal tendance DMI** → `trend_trailing_stop_percent: 2.0%`
   - **Signal contre-tendance DMI** → `counter_trend_trailing_stop_percent: 1.5%`
   - Validation paramètres configuration

3. **Calcul Prix Trailing Stop Initial**
   ```
   Position LONG:
   - Prix entrée : 45000 USDT
   - Trailing stop tendance (2.0%) : 44100 USDT
   - Trailing stop contre-tendance (1.5%) : 44325 USDT
   
   Position SHORT:
   - Prix entrée : 45000 USDT  
   - Trailing stop tendance (2.0%) : 45900 USDT
   - Trailing stop contre-tendance (1.5%) : 45675 USDT
   ```

4. **Placement Ordre Trailing Stop**
   - Utilisation API BingX trailing stop natif
   - Paramètres : symbol, side, quantity, callbackRate
   - Validation placement ordre réussi

5. **Enregistrement État Initial**
   - Prix entrée sauvegardé
   - Type trailing stop initial enregistré
   - Timestamp ouverture position
   - Activation monitoring continu

---

## 🔄 WORKFLOW AJUSTEMENT DYNAMIQUE TRAILING STOP

### 📊 Surveillance Continue et Ajustements

#### Étapes de Monitoring :

1. **Surveillance Temps Réel**
   - Monitor prix position toutes les secondes
   - Calcul % profit flottant en continu
   - Surveillance événements déclencheurs

2. **Détection Événements Déclencheurs**
   
   #### **A. CCI Zone Extrême Inverse**
   ```
   Position LONG ouverte sur CCI survente:
   → CCI passe en zone surachat (+100/+150)
   → Déclenchement ajustement immédiat
   
   Position SHORT ouverte sur CCI surachat:
   → CCI passe en zone survente (-100/-150) 
   → Déclenchement ajustement immédiat
   ```
   
   #### **B. MACD Croisement Inverse (avec profit)**
   ```
   Position LONG ouverte sur MACD hausse:
   → MACD croise à la baisse + profit > 0%
   → Déclenchement ajustement conditionnel
   
   Position SHORT ouverte sur MACD baisse:
   → MACD croise à la hausse + profit > 0%
   → Déclenchement ajustement conditionnel
   ```
   
   #### **C. DI Counter-Trend (avec profit)**
   ```
   Position ouverte DI+ > DI-:
   → DI- dépasse DI+ + profit > 0%
   → Déclenchement ajustement conditionnel
   ```

3. **Application Grille d'Ajustement**
   ```yaml
   Grille Profit → Nouveau Trailing Stop:
   - [0-5% profit]   → 2.0% (maintenu)
   - [5-10% profit]  → 1.5% (25% plus serré)
   - [10-20% profit] → 1.0% (50% plus serré)
   - [20%+ profit]   → 0.5% (75% plus serré)
   ```

4. **Mise à jour Trailing Stop**
   - Annulation ancien ordre trailing stop
   - Calcul nouveau prix selon grille
   - Placement nouveau trailing stop si plus serré
   - Validation exécution réussie

---

## ⚡ WORKFLOW SORTIE ANTICIPÉE MACD

### 🚨 Sortie Avant Trailing Stop Positif

#### Conditions et Actions :

1. **Surveillance Croisement MACD Inverse**
   - Position LONG + MACD croise à la baisse
   - Position SHORT + MACD croise à la hausse
   - Vérification statut trailing stop

2. **Évaluation "Positif" Trailing Stop**
   ```
   Position LONG:
   Prix trailing stop > Prix entrée → Trailing stop "positif"
   
   Position SHORT:  
   Prix trailing stop < Prix entrée → Trailing stop "positif"
   ```

3. **Décision Sortie Anticipée**
   - **SI** trailing stop PAS encore positif → **Sortie immédiate**
   - **SI** trailing stop déjà positif → **Laisser courir**

4. **Exécution Sortie Anticipée**
   - Annulation ordre trailing stop
   - Placement ordre market de fermeture
   - Log raison : "MACD_EARLY_EXIT"
   - Mise à jour statut position

---

## 🚨 WORKFLOW CIRCUIT BREAKERS (Stop d'Urgence)

### 💥 Arrêts Automatiques Globaux

#### Surveillance Limites Journalières :

1. **Calcul Performance Journalière**
   - Récupération toutes positions fermées du jour
   - Calcul PnL total journalier
   - Calcul % perte par rapport capital initial

2. **Vérification Seuil Journalier (-5%)**
   ```
   Capital début jour : 10000 USDT
   PnL jour actuel : -520 USDT  
   % perte : -5.2% → DÉPASSEMENT !
   ```

3. **Actions Arrêt Journalier**
   - Fermeture immédiate TOUTES positions ouvertes
   - Désactivation trading pour reste de la journée
   - Notification d'urgence : "DAILY_LIMIT_BREACH"
   - Log détaillé des positions fermées

#### Surveillance Limites Mensuelles :

1. **Calcul Performance Mensuelle**
   - Récupération PnL des 30 derniers jours
   - Calcul % perte mensuelle cumulée

2. **Vérification Seuil Mensuel (-15%)**
   - Si dépassement → Actions arrêt mensuel

3. **Actions Arrêt Mensuel**
   - Fermeture immédiate toutes positions
   - Désactivation trading jusqu'à lendemain
   - Retry automatique jour suivant
   - Alerte critique : "MONTHLY_LIMIT_BREACH"

---

## 💰 WORKFLOW MONTANTS FIXES

### 🎯 Gestion Simplifiée Position Sizing

#### Paramètres Configurés :

1. **Montants Prédéfinis**
   ```yaml
   fixed_amounts:
     spot_trade_amount: 1000.0      # 1000 USDT par trade spot
     futures_trade_amount: 500.0    # 500 USDT par trade futures  
     leverage: 10                   # Levier fixe 10x
   ```

2. **Validation Solde Disponible**
   - Vérification solde >= montant fixe
   - Calcul marge requise (futures)
   - Validation limites exchange

3. **Calcul Quantité**
   ```
   Spot: Quantité = Montant_fixe / Prix_marché
   Futures: Quantité = (Montant_fixe * Levier) / Prix_marché
   ```

---

## 📊 WORKFLOW MONITORING PERFORMANCE

### 📈 Surveillance Continue Métriques

#### Collecte Métriques Temps Réel :

1. **Métriques Position**
   - PnL flottant en continu
   - % profit par rapport entrée  
   - Drawdown maximum depuis ouverture
   - Durée position ouverte

2. **Métriques Globales**
   - PnL journalier cumulé
   - Nombre positions ouvertes/fermées
   - Win rate journalier
   - Profit factor

3. **Alertes Préventives**
   - Approche limite journalière (-4%)
   - Drawdown position excessive (>8%)
   - Durée position anormale (>24h)

#### Reporting Automatique :

1. **Rapport Fin de Journée**
   - Synthèse performance journalière
   - Positions fermées avec détails
   - Respect limites risques

2. **Rapport Hebdomadaire**
   - Performance cumulée semaine
   - Analyse win rate et profit factor
   - Recommandations ajustements

---

## 🔄 INTÉGRATION AVEC ENGINE TEMPORAL

### 🕒 Synchronisation Avec Cycle Principal

#### Points d'Intégration :

1. **À Chaque Tick (1 seconde)**
   - Mise à jour trailing stops
   - Vérification déclencheurs ajustement
   - Surveillance circuit breakers

2. **À Chaque Barre (5min)**
   - Recalcul indicateurs MACD/CCI/DMI
   - Détection événements zone inverse
   - Évaluation sortie anticipée

3. **À Chaque Signal**
   - Placement trailing stop initial
   - Configuration selon type DMI
   - Activation monitoring position

4. **À Chaque Fermeture Position**
   - Mise à jour métriques globales
   - Vérification limites journalières/mensuelles
   - Log détaillé performance

---

## ⚙️ PARAMÈTRES DE CONFIGURATION

### 📋 Configuration Money Management

```yaml
money_management:
  # Trailing Stops Initiaux
  trend_trailing_stop_percent: 2.0        # Signal tendance DMI
  counter_trend_trailing_stop_percent: 1.5 # Signal contre-tendance DMI
  
  # Grille Ajustement Trailing Stop  
  trailing_stop_adjustment_grid:
    - profit_range: [0, 5]
      trailing_stop_percent: 2.0
    - profit_range: [5, 10]  
      trailing_stop_percent: 1.5
    - profit_range: [10, 20]
      trailing_stop_percent: 1.0
    - profit_range: [20, 100]
      trailing_stop_percent: 0.5
  
  # Montants Fixes
  fixed_amounts:
    spot_trade_amount: 1000.0
    futures_trade_amount: 500.0
    leverage: 10
  
  # Circuit Breakers
  risk_limits:
    max_daily_loss_percent: 5.0
    max_monthly_loss_percent: 15.0
    daily_limit_action: "halt_for_day"
    monthly_limit_action: "halt_daily_retry"
  
  # Monitoring
  monitoring:
    update_frequency_seconds: 1
    metrics_collection: true
    alerting_enabled: true
    preventive_alerts:
      daily_warning_at_percent: 4.0
      position_drawdown_limit: 8.0
```

---

## 🎯 FLUX DÉCISIONNEL GLOBAL

### 🧠 Arbre de Décision Money Management

```
1. Signal Trading Reçu
   ├─ Valider solde disponible
   ├─ Déterminer type trailing stop (DMI)
   ├─ Placer position + trailing stop initial
   └─ Activer monitoring continu

2. Position Active
   ├─ Monitor profit % temps réel
   ├─ Surveiller déclencheurs:
   │  ├─ CCI zone inverse → Ajuster immédiat
   │  ├─ MACD inverse + profit → Ajuster conditionnel
   │  └─ DI counter + profit → Ajuster conditionnel
   └─ Vérifier sortie anticipée MACD

3. Fermeture Position  
   ├─ Trailing stop touché (priorité 1)
   ├─ Sortie anticipée MACD (priorité 2)
   └─ Circuit breaker (priorité 3)

4. Mise à Jour Globale
   ├─ Calculer PnL journalier/mensuel
   ├─ Vérifier limites risques
   └─ Actions correctives si dépassement
```

— Fin workflow Money Management —
