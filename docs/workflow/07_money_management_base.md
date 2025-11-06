# Workflow Money Management BASE - Core Invariant

## 📋 Vue d'Ensemble

Workflow Money Management de base invariant pour toutes stratégies : circuit breakers globaux, limites risques, position sizing de base et métriques communes.

---

## 🚨 WORKFLOW CIRCUIT BREAKERS GLOBAUX

### 💥 Workflow Arrêt Journalier (-5%)

#### Étapes Surveillance Continue :

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

### 📊 Workflow Arrêt Mensuel (-15%)

#### Étapes Surveillance Mensuelle :

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

## 💰 WORKFLOW POSITION SIZING BASE

### 🎯 Workflow Montants Fixes

#### Étapes Calcul Position :

1. **Validation Paramètres**
   ```yaml
   default_position_sizing:
     spot_amount: 1000.0      # USDT par trade spot
     futures_amount: 500.0    # USDT par trade futures
     futures_leverage: 10     # Levier fixe futures
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

4. **Ajustement Précision**
   - Selon symbole (8 dec BTC, 3 dec ETH)
   - Respect minimums exchange
   - Validation finale quantité

---

## 📊 WORKFLOW MÉTRIQUES GLOBALES

### 📈 Workflow Collecte Performance Locale

#### Étapes Temps Réel :

1. **Collecte Métriques Position (en mémoire)**
   - PnL flottant en continu
   - % profit par rapport entrée  
   - Drawdown maximum depuis ouverture
   - Durée position ouverte

2. **Collecte Métriques Globales (en mémoire)**
   - PnL journalier cumulé
   - Nombre positions ouvertes/fermées
   - Win rate journalier
   - Profit factor

3. **Agrégation Cross-Strategy (locale)**
   - Performance globale toutes stratégies
   - Métriques par stratégie isolées
   - Comparaison relative stratégies

4. **Sauvegarde Fichiers Locaux**
   - Snapshots métriques réguliers (JSON)
   - Historique performance journalière
   - Logs métriques pour analyse post-mortem

---

## 🔍 WORKFLOW AUDIT TRAIL

### 📝 Workflow Logs Compliance

#### Étapes Logging Obligatoire :

1. **Logs Circuit Breaker**
   - Activation : Timestamp + cause + positions fermées
   - Limite approchée : Alertes préventives -4% journalier
   - Recovery : Réactivation après période

2. **Logs Configuration**
   - Changement paramètres : Qui/Quand/Quoi modifié
   - Validation paramètres : Rejets + raisons
   - État système : Snapshots réguliers

3. **Logs Métriques Critiques**
   - Performance journalière/mensuelle
   - Dépassements seuils
   - Actions correctives appliquées

4. **Sécurité Audit**
   - Hash chain logs : Protection tamper-proof
   - Backup séparé : Stockage audit externe
   - Contrôle accès : Authentification logs

---

## 🔄 WORKFLOW INTÉGRATION STRATÉGIES

### 🎯 Workflow Coordination BASE ↔ Comportemental

#### Étapes Intégration :

1. **Validation Globale (Priorité Absolue)**
   ```
   Strategy MM prend décision → Core MM valide limites globales
   Si limite OK → Exécution + log audit
   Si limite KO → Circuit breaker + arrêt global
   ```

2. **Collecte Métriques Remontées**
   - Réception métriques strategy-specific
   - Agrégation vers métriques globales
   - Calcul impact sur limites globales

3. **Application Circuit Breakers**
   - Monitoring continu toutes stratégies
   - Déclenchement basé sur PnL global
   - Actions impactent toutes stratégies

4. **Reporting Consolidé**
   - Métriques globales + breakdown par stratégie
   - Performance comparison ranking
   - Recommandations basées sur global

---

## ⚙️ WORKFLOW CONFIGURATION RUNTIME

### 🔧 Workflow Modification Paramètres

#### Étapes Configuration Dynamique :

1. **Réception Demande Changement**
   - API/interface configuration
   - Validation utilisateur autorisé
   - Parsing nouveaux paramètres

2. **Validation Paramètres**
   ```yaml
   circuit_breakers:
     daily_limit_percent: 5.0         # Validé 0.1-20.0%
     monthly_limit_percent: 15.0      # Validé 1.0-50.0%
   
   position_sizing:
     default_spot_amount: 1000.0      # Validé >0
     default_futures_amount: 500.0    # Validé >0
   ```

3. **Application Progressive**
   - Nouveaux paramètres pour futures opérations
   - Positions existantes gardent anciens paramètres
   - Monitoring transition smooth

4. **Logging Changement**
   - Audit trail modification complète
   - Avant/après paramètres
   - Impact estimé sur opérations

---

## 📊 WORKFLOW REPORTING AUTOMATIQUE

### 📈 Workflow Rapports BASE

#### Étapes Reporting Quotidien :

1. **Génération Rapport Fin Jour (23h59 UTC)**
   - Synthèse performance journalière globale
   - État limites journalière/mensuelle
   - Positions fermées avec PnL détaillé

2. **Calcul Métriques Consolidées**
   - Win rate global toutes stratégies
   - Profit factor consolidé
   - Drawdown maximum période

3. **Analyse Respect Limites**
   - Distance actuelle aux seuils critiques
   - Historique déclenchements circuit breakers
   - Recommandations préventives

4. **Sauvegarde Rapport Local**
   - Format JSON structuré + résumé texte
   - Écriture fichiers logs locaux
   - Archivage local pour historique compliance

---

## 🎯 RESPONSABILITÉS CORE MM

### ✅ Workflow Responsabilités Invariantes

#### **Ce que gère Core MM :**
- **Protection globale** : Circuit breakers toutes stratégies
- **Position sizing base** : Calculs montants fixes standard
- **Monitoring global** : Métriques cross-strategy
- **Compliance** : Audit trail complet
- **Configuration** : Paramètres runtime globaux

#### **Interface avec Strategy MM :**
- Validation limites avant exécution strategy decisions
- Collecte métriques strategy pour agrégation globale
- Application circuit breakers prioritaire sur strategy logic
- Fourniture services base (position sizing, logging, config)

#### **Workflow Délégation :**
- Strategy MM gère trailing stops spécifiques
- Strategy MM gère réactions événements indicateurs
- Strategy MM gère conditions sortie anticipée
- Core MM valide + protège + monitore globalement

— Fin Workflow Money Management BASE —
