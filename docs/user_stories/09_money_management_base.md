# User Stories Money Management BASE - Core Invariant

## 📋 Vue d'Ensemble

User Stories pour le Money Management de base invariant : circuit breakers globaux, limites risques, position sizing et métriques communes à toutes stratégies.

---

## 🚨 USER STORIES - CIRCUIT BREAKERS

### 🛑 Story #1 : Arrêt d'Urgence Perte Journalière

```
En tant que : Gestionnaire de risque automatisé
Je veux : Arrêter complètement le trading si perte journalière dépasse -5%
Pour que : Mon capital soit protégé contre les journées catastrophiques

Critères d'Acceptation :
✅ Calcul PnL journalier en temps réel (toutes positions fermées)
✅ Surveillance continue du seuil -5% par rapport capital début jour
✅ Déclenchement immédiat si seuil atteint ou dépassé
✅ Fermeture market de TOUTES positions ouvertes instantanément
✅ Désactivation complète trading jusqu'à 00h00 UTC lendemain
✅ Notification urgence : "DAILY_LIMIT_BREACH" avec détails
✅ Log complet : capital initial, PnL final, positions fermées

Scénarios de Test :
- Capital début : 10000 USDT, PnL -500 USDT → Surveillance active
- PnL atteint -501 USDT (-5.01%) → Déclenchement immédiat
- 3 positions ouvertes → Toutes fermées en <30 secondes  
- Tentative nouveau trade → Rejet avec message explicite

Données de Test :
- Capital jour : 10000 USDT
- Positions actives : BTC LONG (-200), ETH SHORT (+50), SOL LONG (-351)
- PnL total : -501 USDT (-5.01%)

Définition de Fini (DoD) :
- Monitoring PnL temps réel sans latence
- Fermeture multi-positions simultanée fiable
- Blocage trading effectif jusqu'à reset minuit
- Logs audit trail complets pour compliance
```

### 📊 Story #2 : Limite Mensuelle avec Retry

```
En tant que : Système de contrôle des risques
Je veux : Gérer les pertes mensuelles avec arrêt et retry automatique
Pour que : Les mauvais mois n'épuisent pas le capital sur la durée

Critères d'Acceptation :
✅ Calcul PnL mensuel glissant (30 derniers jours calendaires)
✅ Surveillance seuil -15% par rapport capital début mois
✅ Fermeture toutes positions si seuil atteint
✅ Arrêt trading pour reste de la journée courante
✅ Réactivation automatique à 00h00 UTC jour suivant
✅ Notification : "MONTHLY_LIMIT_BREACH" + plan retry
✅ Historique mensuel sauvegardé pour analyse

Scénarios de Test :
- Capital début mois : 10000 USDT, PnL -30 jours : -1480 USDT  
- Déclenchement à -1501 USDT (-15.01%)
- Arrêt 15h30 → Réactivation 00h00 lendemain
- Vérification calcul glissant correct (pas calendaire fixe)

Données de Test :
- Capital mensuel : 10000 USDT  
- PnL 30 jours : -1501 USDT (-15.01%)
- Heure déclenchement : 15h30 UTC
- Retry attendu : 00h00 UTC jour+1

Définition de Fini (DoD) :
- Calcul mensuel glissant précis au jour près
- Mécanisme retry automatique fiable 
- Persistence état entre redémarrages système
- Métriques longue durée pour reporting mensuel
```

---

## 💰 USER STORIES - POSITION SIZING BASE

### 🎯 Story #3 : Position Sizing Montant Fixe

```
En tant que : Trader avec approche simplifiée
Je veux : Utiliser des montants fixes par trade sans calculs complexes
Pour que : Ma gestion soit prévisible et mes risques maîtrisés

Critères d'Acceptation :
✅ Spot : 1000 USDT par trade (configurable)
✅ Futures : 500 USDT par trade avec levier 10x (configurable)
✅ Validation solde suffisant avant ouverture position
✅ Calcul quantité automatique selon prix marché
✅ Respect minimums/maximums exchange BingX
✅ Ajustement précision selon symbole (8 décimales BTC, 2 ETH, etc.)
✅ Log montant, quantité, prix d'exécution

Scénarios de Test :
- Signal BTC spot : 1000 USDT à 45000 USD/BTC → 0.02222 BTC
- Signal ETH futures : 500 USDT×10 levier à 3000 USD/ETH → 1.667 ETH
- Solde insuffisant (800 USDT) → Rejet avec message clair
- Prix très élevé → Quantité très petite mais > minimum exchange

Données de Test :
- Montant spot configuré : 1000 USDT
- Montant futures configuré : 500 USDT, levier 10x
- Prix BTC : 45000 USD, minimum 0.00001 BTC
- Prix ETH : 3000 USD, minimum 0.001 ETH

Définition de Fini (DoD) :
- Calculs quantité précis selon règles exchange
- Gestion erreurs montant insuffisant élégante
- Configuration montants runtime sans redémarrage
- Validation limites exchange temps réel
```

---

## 📈 USER STORIES - MONITORING GLOBAL

### 📊 Story #4 : Surveillance Métriques Cross-Strategy Locale

```
En tant que : Agent économique autonome
Je veux : Collecter et sauvegarder toutes les métriques de performance
Pour que : Je puisse détecter les anomalies et optimiser globalement

Critères d'Acceptation :
✅ PnL global mis à jour toutes les secondes (toutes stratégies)
✅ Métriques globales : win rate, profit factor, PnL journalier/mensuel
✅ Métriques par stratégie : performance isolée + comparaison
✅ Sauvegarde métriques fichiers JSON locaux temps réel
✅ Logs alertes préventives : approche limites, performance dégradée
✅ Persistence métriques historiques fichiers pour analyse post-mortem

Scénarios de Test :
- 3 stratégies actives : performance globale + breakdown détaillé
- Stratégie A : +2.3%, Stratégie B : -0.8%, Stratégie C : +1.1% → Global : +2.6%
- PnL approche -4.5% → Alerte préventive "APPROACHING_DAILY_LIMIT"
- Comparaison stratégies : ranking performance + recommandations

Données de Test :
- Stratégies multiples avec PnL variés
- Historique 30 jours pour métriques consolidées
- Simulation approche limites risque

Définition de Fini (DoD) :
- Métriques temps réel <1 seconde de latence
- Calculs agrégation précis (cross-strategy)
- Système logs alertes configurable et fiable
- Sauvegarde fichiers métriques performante
```

### 📝 Story #5 : Reporting Automatique Local

```
En tant que : Agent économique autonome
Je veux : Générer des rapports automatiques consolidés locaux
Pour que : Je puisse analyser la performance globale et par stratégie

Critères d'Acceptation :
✅ Rapport journalier automatique à 23h59 UTC
✅ Synthèse globale : PnL jour, win rate, profit factor consolidés
✅ Breakdown par stratégie : performance individuelle + ranking
✅ Respect limites : distance aux seuils journalier/mensuel
✅ Rapport mensuel : tendances, évolution comparative stratégies
✅ Format structuré : JSON + résumé exécutif lisible
✅ Sauvegarde automatique fichiers logs locaux

Scénarios de Test :
- Fin journée : 45 trades total, 3 stratégies, PnL global +1.8%
- Breakdown : MACD/CCI/DMI +2.1%, RSI/BB +0.9%, EMA/Vol +2.4%
- Analyse mensuelle : évolution win rate, profit factor, drawdown max
- Recommandations : allocation optimale entre stratégies

Données de Test :
- Historique multi-stratégies complet mois
- Métriques performance calculées par stratégie
- Configuration reporting multiples destinations

Définition de Fini (DoD) :
- Génération rapports automatique fiable cross-strategy
- Calculs métriques consolidées exactes
- Templates rapports informatifs + actionables
- Sauvegarde fichiers rapports robuste et structurée
```

---

## 🔧 USER STORIES - CONFIGURATION SYSTÈME

### ⚙️ Story #6 : Configuration Runtime Globale

```
En tant que : Administrateur système trading
Je veux : Modifier les paramètres Money Management global sans redémarrage
Pour que : Je puisse ajuster rapidement selon conditions marché globales

Critères d'Acceptation :
✅ Modification limites risque (-5%/-15%) via API/config sans arrêt
✅ Ajustement montants fixes base en temps réel
✅ Configuration circuit breakers (seuils, actions) dynamique
✅ Validation paramètres avant application (cohérence, limites)
✅ Application immédiate pour nouvelles opérations
✅ Positions existantes non impactées (continuité)
✅ Log changements configuration avec audit trail complet

Scénarios de Test :
- Modification limite journalière 5% → 4% → Application immédiate
- Changement montants base : spot 1000→1200, futures 500→600
- Paramètre invalide (limite -50%) → Rejet avec erreur claire
- Positions actives pendant changement → Continuité garantie

Données de Test :
- Configuration globale actuelle complète
- Nouveaux paramètres avec validations à tester
- Positions actives multi-stratégies pour test continuité

Définition de Fini (DoD) :
- API configuration runtime complète et sécurisée
- Validation paramètres exhaustive avant application
- Application smooth sans interruption opérations
- Logs audit trail modifications configuration complets
```

---

## 📋 MATRICE PRIORITÉS USER STORIES BASE

### 🎯 Classification par Criticité BASE

#### **Priorité CRITIQUE (Must Have) :**
- Story #1 : Arrêt Perte Journalière  
- Story #2 : Limite Mensuelle
- Story #3 : Position Sizing Fixe

#### **Priorité HAUTE (Should Have) :**
- Story #4 : Monitoring Cross-Strategy
- Story #6 : Configuration Runtime

#### **Priorité MOYENNE (Could Have) :**
- Story #5 : Reporting Automatique

### 🔄 Ordre d'Implémentation Recommandé BASE

```
Sprint 1 (Protection) : Stories #1, #2
Sprint 2 (Foundation) : Story #3  
Sprint 3 (Monitoring) : Stories #4, #6
Sprint 4 (Reporting) : Story #5
```

— Fin User Stories Money Management BASE —
