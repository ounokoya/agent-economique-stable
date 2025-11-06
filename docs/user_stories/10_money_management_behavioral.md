# User Stories Money Management Comportemental - Stratégie MACD/CCI/DMI

## 📋 Vue d'Ensemble

User Stories pour le Money Management comportemental spécifique à la stratégie MACD/CCI/DMI : trailing stops adaptatifs, réactions aux événements indicateurs et sortie anticipée.

---

## 🎯 USER STORIES - TRAILING STOP INITIAL

### 🟢 Story #1 : Placement Trailing Stop Selon Signal DMI

```
En tant que : Agent économique automatisé
Je veux : Placer automatiquement un trailing stop adapté au type de signal DMI
Pour que : Chaque position soit protégée dès l'ouverture avec le bon niveau de risque

Critères d'Acceptation :
✅ Signal tendance DMI → Trailing stop 2.0% placé automatiquement
✅ Signal contre-tendance DMI → Trailing stop 1.5% placé automatiquement  
✅ Trailing stop placé dans les 5 secondes après ouverture position
✅ Prix trailing stop calculé correctement selon direction (LONG/SHORT)
✅ Ordre trailing stop confirmé par API BingX avec order ID
✅ État initial sauvegardé : prix entrée, type stop, timestamp
✅ Logging détaillé : "TRAILING_STOP_PLACED" avec paramètres

Scénarios de Test :
- Signal LONG tendance DMI : BTC à 45000 → Trailing stop à 44100 (2%)
- Signal SHORT contre-tendance DMI : ETH à 3000 → Trailing stop à 3045 (1.5%)
- Validation calcul prix selon formules correctes
- Vérification placement ordre BingX réussi

Données de Test :
- Position LONG BTC-USDT : 45000 USDT, signal tendance  
- Position SHORT ETH-USDT : 3000 USDT, signal contre-tendance
- Quantité : 0.1 BTC / 1 ETH
- Environnement : Demo VST

Définition de Fini (DoD) :
- Tests automatisés passent avec couverture 100%
- Logs structurés générés avec tous paramètres
- Métriques collectées : temps placement, taux succès
- Documentation API mise à jour
```

### 🔄 Story #2 : Ajustement Trailing Stop sur CCI Inverse

```
En tant que : Agent économique en position active
Je veux : Ajuster automatiquement mon trailing stop quand CCI entre en zone inverse
Pour que : Mes gains soient mieux sécurisés lors de retournements de marché

Critères d'Acceptation :
✅ Position LONG (CCI survente) → Ajustement si CCI > +100 (surachat)
✅ Position SHORT (CCI surachat) → Ajustement si CCI < -100 (survente)
✅ Ajustement selon grille profit : 5% → 1.5%, 10% → 1.0%, etc.
✅ Nouveau trailing stop plus serré que l'actuel uniquement
✅ Ancien ordre trailing stop annulé avant nouveau placement
✅ Monitoring continu tant que CCI reste en zone inverse
✅ Log événement : "CCI_INVERSE_ADJUSTMENT" avec profit %

Scénarios de Test :
- Position LONG BTC, profit 8% → CCI passe à +120 → Stop ajusté à 1.5%
- Position SHORT ETH, profit 12% → CCI passe à -130 → Stop ajusté à 1.0%  
- Position avec profit <5% → CCI inverse → Pas d'ajustement (maintien 2.0%)
- Nouveau stop moins serré → Pas d'ajustement (garde actuel)

Données de Test :
- Position BTC-USDT LONG : entrée 45000, prix actuel 48600 (8% profit)
- CCI passe de -120 (survente) à +120 (surachat)  
- Trailing stop actuel : 2.0%, nouveau calculé : 1.5%

Définition de Fini (DoD) :
- Détection CCI inverse temps réel fonctionnelle
- Calcul grille ajustement exact
- Replacement ordre trailing stop sans interruption
- Tests edge cases : profits limites, CCI volatil
```

### 🔍 Story #3 : Sortie Anticipée MACD Inverse

```
En tant que : Trader automatisé prudent
Je veux : Fermer ma position si MACD inverse avant trailing stop positif  
Pour que : J'évite les pertes lors de retournements précoces

Critères d'Acceptation :
✅ Position LONG + MACD croise baisse → Évaluation sortie anticipée
✅ Position SHORT + MACD croise hausse → Évaluation sortie anticipée
✅ Sortie SI trailing stop pas encore "positif" (prix > entrée)
✅ Maintien position SI trailing stop déjà "positif"
✅ Fermeture market instantanée en cas sortie anticipée
✅ Annulation trailing stop lors fermeture anticipée
✅ Log raison : "MACD_EARLY_EXIT" avec prix sortie

Scénarios de Test :
- LONG BTC entrée 45000, prix 44500, MACD inverse → Sortie (stop pas positif)
- LONG ETH entrée 3000, prix 3100, MACD inverse → Maintien (stop positif)
- SHORT SOL entrée 200, prix 195, MACD inverse → Sortie (stop pas positif)
- Validation vitesse exécution <3 secondes

Données de Test :
- Position BTC LONG : entrée 45000, prix actuel 44500
- Trailing stop actuel : 43560 (2%, pas encore positif)
- MACD croise de +0.5 vers -0.2 (signal inverse)

Définition de Fini (DoD) :
- Détection croisement MACD temps réel
- Calcul état "positif" trailing stop correct
- Fermeture position market sans slippage excessif
- Tests timing critique : MACD volatil, prix rapides
```

---

## 🚨 USER STORIES - CIRCUIT BREAKERS

### 🛑 Story #4 : Arrêt d'Urgence Perte Journalière

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

### 📊 Story #5 : Limite Mensuelle avec Retry

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

## 💰 USER STORIES - MONTANTS FIXES

### 🎯 Story #6 : Position Sizing Montant Fixe

```
En tant que : Trader avec stratégie simple
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

## 📈 USER STORIES - MONITORING PERFORMANCE

### 📊 Story #7 : Surveillance Métriques Temps Réel

```
En tant que : Superviseur système trading
Je veux : Monitorer toutes les métriques de performance en continu
Pour que : Je puisse détecter les anomalies et optimiser la stratégie

Critères d'Acceptation :
✅ PnL flottant mis à jour chaque seconde pour chaque position
✅ % profit calculé par rapport prix d'entrée en temps réel
✅ Drawdown maximum tracké depuis ouverture position
✅ Métriques globales : win rate, profit factor, PnL journalier
✅ Dashboard temps réel accessible via API/interface
✅ Alertes préventives : approche limites, drawdown excessif
✅ Persistence métriques historiques pour analyse

Scénarios de Test :
- Position BTC : entrée 45000, prix 46800 → PnL +4%, drawdown max -2%
- 10 positions fermées jour : 7 wins, 3 losses → Win rate 70%
- PnL approche -4.5% → Alerte préventive "APPROACHING_DAILY_LIMIT"
- Drawdown position >8% → Alerte "EXCESSIVE_DRAWDOWN"

Données de Test :
- Positions multiples avec PnL variés
- Historique trades journalier pour win rate
- Simulation approche limites risque

Définition de Fini (DoD) :
- Métriques temps réel <1 seconde de latence
- Calculs statistiques précis (win rate, profit factor)
- Système alertes configurable et fiable
- API métriques pour intégration dashboards externes
```

### 📝 Story #8 : Reporting Automatique Performance

```
En tant que : Analyste performance trading
Je veux : Recevoir des rapports automatiques de performance
Pour que : Je puisse analyser les résultats et ajuster la stratégie

Critères d'Acceptation :
✅ Rapport journalier automatique à 23h59 UTC
✅ Synthèse : PnL jour, nombre trades, win rate, profit factor
✅ Détail positions fermées : entrée, sortie, durée, PnL
✅ Respect limites risque : distance aux seuils journalier/mensuel
✅ Rapport hebdomadaire : performance 7 jours, tendances
✅ Format structuré : JSON + résumé texte lisible
✅ Envoi email/webhook configurable

Scénarios de Test :
- Fin journée : 15 trades, 9 wins, PnL +2.3% → Rapport positif
- Semaine : 5 jours de trading, évolution win rate 65%→72%
- Approche limite mensuelle → Recommandation prudence
- Format JSON valide + résumé texte <200 mots

Données de Test :
- Historique trades complet semaine
- Métriques performance calculées
- Configuration email/webhook test

Définition de Fini (DoD) :
- Génération rapports automatique fiable
- Calculs métriques avancées exactes
- Templates rapports lisibles et informatifs  
- Système notifications robuste (email/webhook)
```

---

## 🔧 USER STORIES - INTÉGRATION SYSTÈME

### ⚙️ Story #9 : Configuration Runtime Money Management

```
En tant que : Administrateur système trading
Je veux : Modifier les paramètres Money Management sans redémarrage
Pour que : Je puisse ajuster rapidement selon les conditions de marché

Critères d'Acceptation :
✅ Modification trailing stop % via API/config sans arrêt système
✅ Ajustement montants fixes en temps réel
✅ Modification limites risque (journalier/mensuel) dynamique
✅ Validation paramètres avant application (cohérence, limites)
✅ Application progressive : nouvelles positions utilisent nouveaux params
✅ Positions existantes gardent anciens params jusqu'à fermeture
✅ Log changements configuration avec timestamp et utilisateur

Scénarios de Test :
- Modification trailing stop 2.0% → 1.8% → Nouvelles positions à 1.8%
- Position existante garde 2.0% jusqu'à fermeture
- Changement limite journalière 5% → 4% → Application immédiate
- Paramètre invalide (trailing stop 15%) → Rejet avec erreur claire

Données de Test :
- Configuration actuelle complète Money Management
- Nouveaux paramètres avec validations à tester
- Position active pour test conservation paramètres

Définition de Fini (DoD) :
- API configuration runtime complète et sécurisée
- Validation paramètres exhaustive avant application
- Coexistence anciens/nouveaux params sans conflit
- Logs audit trail modifications configuration
```

### 🔄 Story #10 : Intégration Engine Temporal

```
En tant que : Développeur système intégré
Je veux : Synchroniser parfaitement Money Management avec Engine Temporal
Pour que : Toutes les décisions soient coordonnées et cohérentes

Critères d'Acceptation :
✅ Appel Money Management à chaque tick (1 seconde) sans latence
✅ Mise à jour trailing stops synchronisée avec prix temps réel
✅ Événements indicateurs (MACD/CCI/DMI) transmis immédiatement
✅ Circuit breakers intégrés dans boucle principale Engine
✅ Partage état positions cohérent entre composants
✅ Gestion erreurs Money Management n'interrompt pas Engine
✅ Métriques performance intégrées dans monitoring global

Scénarios de Test :
- Tick Engine 1Hz → Money Management appelé exactement 1Hz
- Signal MACD inverse → Transmission <100ms à Money Management
- Erreur trailing stop → Engine continue, erreur loggée
- Position fermée par MM → État synchronisé instantanément

Données de Test :
- Engine Temporal en fonctionnement normal
- Money Management avec positions actives
- Simulation erreurs diverses pour robustesse

Définition de Fini (DoD) :
- Synchronisation parfaite sans drift temporel
- Latence communication <100ms garantie
- Résilience aux erreurs sans impact Engine principal
- Tests intégration bout-en-bout 100% passants
```

---

## 📋 MATRICE PRIORITÉS USER STORIES

### 🎯 Classification par Criticité

#### **Priorité CRITIQUE (Must Have) :**
- Story #1 : Placement Trailing Stop Initial
- Story #4 : Arrêt Perte Journalière  
- Story #6 : Montants Fixes
- Story #10 : Intégration Engine

#### **Priorité HAUTE (Should Have) :**
- Story #2 : Ajustement CCI Inverse
- Story #3 : Sortie Anticipée MACD
- Story #5 : Limite Mensuelle

#### **Priorité MOYENNE (Could Have) :**
- Story #7 : Monitoring Temps Réel
- Story #8 : Reporting Automatique
- Story #9 : Configuration Runtime

### 🔄 Ordre d'Implémentation Recommandé

```
Sprint 1 (Fondations) : Stories #1, #6, #10
Sprint 2 (Protections) : Stories #4, #5  
Sprint 3 (Optimisations) : Stories #2, #3
Sprint 4 (Monitoring) : Stories #7, #8, #9
```

— Fin user stories Money Management —
