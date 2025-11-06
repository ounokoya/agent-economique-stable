# Plan de Tests Money Management Comportemental - Stratégie MACD/CCI/DMI

## 📋 Vue d'Ensemble

Plan de tests pour Money Management comportemental de la stratégie MACD/CCI/DMI : trailing stops adaptatifs, réactions événements indicateurs, sortie anticipée. Tests respectant contraintes Go (100% coverage, <500 lignes/fichier).

---

## 🎯 STRATÉGIE GLOBALE DE TESTS

### 🧪 Objectifs Qualité Money Management

#### Couverture et Critères :
- **Couverture code : 100%** (contrainte stricte Money Management)
- **Tests unitaires : Obligatoires** pour chaque fonction financière
- **Tests intégration : Complets** avec Engine Temporal  
- **Tests stress : Performance** sous charge (1000 positions)
- **Tests sécurité : Validation** paramètres financiers

#### Contraintes Spécifiques MM :
- **Précision calculs : ±0.0001%** pour PnL et trailing stops
- **Latence maximale : 100ms** pour circuit breakers
- **Résilience : 99.9%** uptime sous conditions normales
- **Audit trail : Complet** pour toutes décisions financières

---

## 🔧 TESTS UNITAIRES MONEY MANAGEMENT

### 📊 Tests Module Trailing Stop

#### **TestTrailingStopInitialPlacement**
```go
Objectif : Valider placement trailing stop selon type signal DMI
Fonctions testées : 
- PlaceInitialTrailingStop()
- CalculateTrailingStopPrice() 
- DetermineStopTypeFromSignal()

Cas de Tests :
✅ Signal tendance DMI → 2.0% trailing stop correctement placé
✅ Signal contre-tendance DMI → 1.5% trailing stop correctement placé
✅ Position LONG → Prix stop = entrée * (1 - percent/100)
✅ Position SHORT → Prix stop = entrée * (1 + percent/100)
✅ Précision calcul ±0.0001% pour tous prix
✅ Gestion erreur placement échoué
✅ Validation paramètres invalides (percent > 100%, négatif)

Données de Test :
- Position LONG BTC : entrée 45000.0, tendance → stop 44100.0
- Position SHORT ETH : entrée 3000.0, contre-tendance → stop 3045.0
- Cas limites : prix 0.00001, prix 99999999.99

Métriques Attendues :
- Temps exécution : <50ms par placement
- Précision : ±0.0001% sur tous calculs
- Couverture : 100% branches conditionnelles
```

#### **TestTrailingStopDynamicAdjustment** 
```go
Objectif : Valider ajustements selon grille profit
Fonctions testées :
- AdjustTrailingStop()
- CalculateProfitPercent()
- ApplyAdjustmentGrid()
- IsNewStopTighter()

Cas de Tests :
✅ Profit 0-5% → Stop maintenu à valeur initiale
✅ Profit 5-10% → Stop ajusté à 1.5% (plus serré)
✅ Profit 10-20% → Stop ajusté à 1.0% (encore plus serré)
✅ Profit 20%+ → Stop ajusté à 0.5% (maximum serré)
✅ Nouveau stop moins serré → Pas d'ajustement (garde actuel)
✅ Position SHORT → Calculs profit inversés correctement
✅ Grille personnalisée → Application selon config

Test Data :
- Position LONG : entrée 1000, prix 1080 (8% profit) → stop 1.5%
- Position SHORT : entrée 1000, prix 890 (11% profit) → stop 1.0%  
- Edge case : profit exactement 5.0000% → test bordure grille

Performance :
- Latence : <10ms par ajustement
- Précision profit : ±0.01%
- Atomicité : ajustement complet ou rollback
```

#### **TestEarlyExitMACD**
```go
Objectif : Valider sortie anticipée MACD avant trailing stop positif
Fonctions testées :
- EvaluateEarlyExitMACD()
- IsTrailingStopPositive()
- ExecuteEarlyExit()

Cas de Tests :
✅ LONG + MACD inverse + stop pas positif → Sortie immédiate
✅ LONG + MACD inverse + stop déjà positif → Maintien position
✅ SHORT + MACD inverse + stop pas positif → Sortie immédiate  
✅ SHORT + MACD inverse + stop déjà positif → Maintien position
✅ Fermeture market réussie en <5 secondes
✅ Annulation trailing stop lors sortie anticipée
✅ Logs raison : "MACD_EARLY_EXIT" avec détails

Test Scenarios :
- LONG BTC entrée 45000, prix 44800, stop 44100 → Pas positif → EXIT
- LONG ETH entrée 3000, prix 3200, stop 3100 → Positif → KEEP
- Timing critique : MACD inverse pendant volatilité prix élevée

Edge Cases :
- Prix exactement = prix entrée → Test limite "positif"
- MACD oscillation rapide → Pas de double sortie
- Position fermée pendant évaluation → Gestion état race
```

### 💰 Tests Module Circuit Breakers

#### **TestDailyLimitBreaker**
```go
Objectif : Valider arrêt automatique perte journalière -5%
Fonctions testées :
- CalculateDailyPnL()
- CheckDailyLimits() 
- ExecuteEmergencyStop()
- HaltTradingUntilMidnight()

Cas de Tests :
✅ PnL journalier -4.99% → Pas d'arrêt, surveillance continue
✅ PnL journalier -5.00% → Déclenchement immédiat circuit breaker
✅ PnL journalier -5.01% → Déclenchement + fermeture toutes positions
✅ Calcul PnL correct : toutes positions fermées du jour UTC
✅ Fermeture simultanée multiple positions en <30 secondes
✅ Blocage nouveaux trades effectif jusqu'à 00h00 UTC+1
✅ Notification urgence : "DAILY_LIMIT_BREACH" + détails complets

Test Data :
- Capital début jour : 10000.0 USDT
- Positions fermées : -200, +50, -351 USDT = -501 USDT (-5.01%)
- Positions ouvertes : 3 à fermer simultanément
- Heure test : 15h30 UTC → Blocage jusqu'à 00h00 UTC lendemain

Performance Critique :
- Détection limite : <1 seconde après dépassement
- Fermeture positions : <30 secondes toutes positions
- Blocage trades : 100% effectif, 0 faux positif
```

#### **TestMonthlyLimitBreaker**
```go 
Objectif : Valider gestion limite mensuelle avec retry automatique
Fonctions testées :
- CalculateMonthlyPnL()
- CheckMonthlyLimits()
- ScheduleRetryNextDay()

Cas de Tests :
✅ Calcul PnL 30 jours glissants (pas calendaire fixe)
✅ PnL -14.99% → Surveillance, pas d'action
✅ PnL -15.00% → Déclenchement limite mensuelle  
✅ Fermeture toutes positions + arrêt trading jour courant
✅ Réactivation automatique 00h00 UTC jour suivant
✅ Persistence état entre redémarrages système
✅ Historique mensuel sauvegardé pour compliance

Test Scenarios :
- 30 derniers jours : PnL -1501 USDT sur capital 10000 (-15.01%)
- Déclenchement 15h30 → Arrêt → Retry 00h00 jour+1
- Redémarrage système pendant arrêt → État preserved
- Nouveau mois → Reset compteurs limite mensuelle

Data Integrity :
- Calcul glissant précis au jour près
- Persistence état sans corruption
- Recovery après crash système
```

### 🎯 Tests Module Position Sizing

#### **TestFixedAmountPositionSizing**
```go
Objectif : Valider calculs quantité avec montants fixes
Fonctions testées :
- CalculateSpotQuantity()
- CalculateFuturesQuantity() 
- ValidateMinimumAmounts()
- AdjustPrecisionBySymbol()

Cas de Tests :
✅ Spot : 1000 USDT à 45000 USD/BTC → 0.02222 BTC (8 décimales)
✅ Futures : 500 USDT × 10 levier à 3000 USD/ETH → 1.667 ETH
✅ Respect minimums exchange : BTC 0.00001, ETH 0.001
✅ Ajustement précision : BTC 8 dec, ETH 3 dec, SOL 2 dec
✅ Validation solde suffisant avant calcul quantité
✅ Gestion prix extrêmes : très élevé → quantité très petite
✅ Erreurs paramètres : montant négatif, prix zéro

Test Matrix :
| Asset | Prix | Montant | Levier | Quantité Attendue | Précision |
|-------|------|---------|--------|------------------|-----------|
| BTC | 45000 | 1000 | 1x | 0.02222222 | 8 dec |
| ETH | 3000 | 500 | 10x | 1.667 | 3 dec |  
| SOL | 200 | 1000 | 1x | 5.00 | 2 dec |
| ADA | 0.5 | 500 | 5x | 5000.0 | 1 dec |

Edge Cases :
- Prix Bitcoin 100M USD → Quantité 0.00001 BTC (minimum)
- Solde 999 USDT, montant fixe 1000 → Erreur solde insuffisant
- Levier futures 0 → Erreur paramètre invalide
```

---

## 🧪 TESTS INTÉGRATION

### 🔄 Tests Intégration Engine Temporal

#### **TestMoneyManagementEngineSync**
```go
Objectif : Valider synchronisation parfaite MM avec Engine Temporal
Composants testés :
- Cycle principal Engine (1Hz)
- Money Management updates 
- État positions partagé
- Communication événements

Scénarios Intégration :
✅ Engine tick 1Hz → MM appelé exactement 1Hz (±1ms)
✅ Signal trading Engine → MM placement trailing stop <100ms
✅ Événement MACD/CCI/DMI → MM ajustement <50ms
✅ Circuit breaker MM → Engine informé immédiatement
✅ Erreur MM → Engine continue sans interruption
✅ Position fermée MM → État Engine synchronisé instantanément
✅ Redémarrage Engine → MM restaure état correctement

Performance Benchmarks :
- Latence communication : <10ms moyenne, <100ms P99
- Throughput : >1000 événements/seconde sans dégradation
- Synchronisation : 0% drift sur 24h de fonctionnement
- Résilience : 99.9% uptime même avec erreurs MM

Test Endurance :
- 24h fonctionnement continu avec 500 positions simulées
- Injection erreurs aléatoires → Recovery automatique
- Simulation crash/redémarrage → État cohérent restauré
```

#### **TestRealTimeMetricsIntegration**
```go
Objectif : Valider métriques temps réel intégrées
Fonctions testées :
- MetricsCollector.Update()
- RealTimeDashboard.Refresh()
- AlertSystem.Evaluate()

Test Scenarios :
✅ PnL position mis à jour chaque tick (1Hz) sans latence
✅ Métriques globales calculées en continu (win rate, profit factor)
✅ Alertes préventives déclenchées aux bons seuils
✅ Dashboard temps réel <1 seconde de lag derrière Engine
✅ Persistence métriques historiques sans perte de données
✅ API métriques répond en <100ms sous charge normale

Load Testing :
- 1000 positions simultanées → Métriques à jour <1 seconde
- 10000 ticks/seconde → Pas de queue buildup
- Dashboard 50 utilisateurs concurrents → Responsif
```

---

## 🚨 TESTS STRESS ET PERFORMANCE

### ⚡ Tests Performance Haute Charge

#### **TestHighFrequencyTrailingStopUpdates**
```go
Objectif : Valider performance sous mise à jour intensive trailing stops
Conditions de Test :
- 1000 positions actives simultanément
- Ajustements trailing stop 100/seconde
- Volatilité prix élevée (updates 10Hz)

Métriques Performance :
✅ Latence moyenne ajustement : <50ms
✅ Latence P99 ajustement : <200ms  
✅ Throughput : >100 ajustements/seconde soutenus
✅ Memory usage : <500MB pour 1000 positions
✅ CPU usage : <70% sous pic de charge
✅ Aucune perte d'ordre trailing stop
✅ Cohérence état positions : 100%

Scenarios Stress :
- Spike soudain 1000 ajustements simultanés → Recovery <10 secondes
- Volatilité Bitcoin flash crash → Tous trailing stops suivent
- Panne réseau temporaire → Queuing + replay à la reconnexion
```

#### **TestCircuitBreakerUnderLoad**
```go
Objectif : Valider circuit breakers sous charge système élevée
Test Conditions :
- 500 positions ouvertes
- Système CPU 90% utilisé
- Réseau avec latence 500ms

Critical Requirements :
✅ Détection limite journalière : <5 secondes même sous charge
✅ Fermeture 500 positions : <120 secondes maximum
✅ Aucune position "oubliée" lors fermeture masse
✅ Circuit breaker priority : pause autres opérations si nécessaire
✅ Logs complets même sous charge extrême
✅ Recovery système après circuit breaker : <60 secondes

Failure Scenarios :
- 50% positions ferment avec erreur réseau → Retry automatique
- Crash système pendant fermeture masse → Recovery coherent state
- API BingX temporairement indisponible → Queue + retry logic
```

---

## 🔒 TESTS SÉCURITÉ FINANCIÈRE

### 🛡️ Tests Validation Paramètres

#### **TestParameterValidationSecurity**
```go
Objectif : Valider robustesse validation paramètres financiers
Attack Vectors :
- Injection paramètres malformés
- Valeurs extrêmes/overflow
- Tentatives bypass validation

Security Tests :
✅ Trailing stop >100% → Rejet avec erreur sécurisée
✅ Montant négatif → Rejet sans crash système
✅ Prix zéro/négatif → Gestion propre sans division par zéro
✅ Overflow float64 → Detection + gestion gracieuse
✅ Injection SQL dans logs → Sanitization complète
✅ Race conditions multi-thread → Locks appropriés
✅ Memory corruption protection → Bounds checking

Edge Cases Malveillants :
- Trailing stop 999999% → Rejet + log tentative
- Montant NaN/Infinity → Conversion sécurisée
- Paramètres concurrents contradictoires → Coherence locks
```

#### **TestAuditTrailSecurity**
```go
Objectif : Valider intégrité audit trail decisions financières
Security Requirements :
✅ Toute decision MM loggée avec timestamp précis
✅ Logs tamper-proof (hash chain ou signature)
✅ Aucune information sensible en plain text (API keys)
✅ Rotation logs automatique sans perte
✅ Backup audit trail sur stockage séparé
✅ Accès logs restreint + authentification
✅ Compliance réglementaire (GDPR, SOX si applicable)

Audit Coverage :
- Placement/modification trailing stop → Log complet
- Circuit breaker activation → Log avec cause détaillée  
- Modification paramètres → Log qui/quand/quoi/pourquoi
- Accès métriques sensibles → Log consultation
```

---

## 📊 TESTS DONNÉES ET EDGE CASES

### 🎯 Tests Précision Calculs Financiers

#### **TestFinancialCalculationPrecision**
```go
Objectif : Valider précision calculs monétaires critiques
Precision Requirements :
- PnL calculations : ±0.0001 USDT
- Percentage calculations : ±0.0001%  
- Price calculations : ±0.00000001 BTC
- Rounding : Banker's rounding consistent

Test Cases :
✅ Calcul PnL : 0.12345678 BTC × 45123.45678912 USD → Précision 8 décimales
✅ Profit % : (46123.45 - 45123.67) / 45123.67 → ±0.0001% précision
✅ Trailing stop : 45123.45 × (1 - 0.02) → Arrondi cohérent
✅ Accumulation erreurs : 1000 calculs successifs → Drift <0.01%
✅ Conversion devises : USD↔EUR↔BTC → Précision préservée
✅ Overflow protection : Montants > float64 max → Gestion gracieuse

Edge Cases Numériques :
- Très petites valeurs : 0.00000001 BTC calculations
- Très grandes valeurs : 99999999.99999999 USD calculations  
- Division par zéro : Prix marché = 0 → Gestion d'erreur
- Underflow/Overflow : Detection + mitigation
```

### 📈 Tests Données Historiques

#### **TestHistoricalDataConsistency**
```go
Objectif : Valider cohérence données historiques MM
Data Integrity :
✅ PnL historique cohérent avec positions fermées
✅ Métriques calculées identiques en temps réel vs batch
✅ Win rate historique = comptage manuel positions
✅ Profit factor = (total gains) / (total pertes)
✅ Drawdown maximum correct sur période
✅ Pas de gaps temporels dans historique

Test Scenarios :
- Reconstruction métriques depuis logs → Identique à temps réel
- Migration données historiques → Intégrité préservée  
- Purge données anciennes → Métriques affectées correctement
- Backup/restore → État cohérent après restauration

Performance Historique :
- Query 1 an d'historique : <10 secondes
- Calcul métriques 1 mois : <5 secondes
- Indexation données : Optimal query performance
```

---

## 🧪 TESTS AUTOMATISATION ET CI/CD

### 🤖 Pipeline Tests Automatisés

#### **Stratégie CI/CD Money Management**
```yaml
# Pipeline Configuration
stages:
  - unit_tests          # Tests unitaires 100% coverage
  - integration_tests   # Tests intégration Engine
  - performance_tests   # Tests charge + latence  
  - security_tests      # Tests sécurité + validation
  - end_to_end_tests   # Tests bout-en-bout complets

quality_gates:
  code_coverage: 100%           # Pas de compromis MM
  performance_latency: <100ms   # Circuit breakers critiques
  security_score: A+           # Aucune vulnérabilité
  financial_precision: ±0.0001% # Calculs exacts

environments:
  - unit: go test -race -cover ./...
  - integration: docker-compose integration stack  
  - performance: k6 load testing + monitoring
  - security: gosec + custom financial validators
```

#### **TestAutomationBenchmarks**
```go
Objectif : Benchmark performance automatisé dans CI/CD
Benchmarks Critiques :
- BenchmarkTrailingStopUpdate: <50ms/op target
- BenchmarkCircuitBreakerCheck: <10ms/op target  
- BenchmarkPnLCalculation: <1ms/op target
- BenchmarkMemoryUsage: <500MB pour 1000 positions

Regression Testing :
✅ Performance ne dégrade jamais >20% entre versions
✅ Memory usage stable ±10% entre releases
✅ Latence P99 reste <seuils critiques
✅ Throughput maintenu sous charge

Automated Alerts :
- Performance régression → Block deploy + alert équipe
- Coverage <100% → Build failed automatiquement
- Security scan failed → Deploy impossible  
- Financial precision tests failed → Escalation immédiate
```

---

## 📋 MATRICE EXÉCUTION TESTS

### 🎯 Planning Exécution par Phase

#### **Phase 1 : Tests Unitaires (Semaine 1)**
```
Priority: CRITICAL
- TestTrailingStopInitialPlacement ✅ 
- TestFixedAmountPositionSizing ✅
- TestParameterValidationSecurity ✅
- TestFinancialCalculationPrecision ✅

Coverage Target: 100% functions money management core
Success Criteria: Tous tests passent + benchmarks dans limites
```

#### **Phase 2 : Tests Circuit Breakers (Semaine 2)**  
```
Priority: CRITICAL
- TestDailyLimitBreaker ✅
- TestMonthlyLimitBreaker ✅ 
- TestCircuitBreakerUnderLoad ✅

Coverage Target: 100% risk management + emergency procedures  
Success Criteria: Circuit breakers <5sec reaction + 100% position closure
```

#### **Phase 3 : Tests Intégration (Semaine 3)**
```
Priority: HIGH  
- TestMoneyManagementEngineSync ✅
- TestRealTimeMetricsIntegration ✅
- TestHighFrequencyTrailingStopUpdates ✅

Coverage Target: 100% integration points avec Engine Temporal
Success Criteria: <100ms latency + 99.9% uptime + sync parfaite
```

#### **Phase 4 : Tests Performance + Sécurité (Semaine 4)**
```
Priority: HIGH
- Performance benchmarks suite complète
- Security penetration testing  
- End-to-end system validation

Coverage Target: Système complet sous charge + attack scenarios
Success Criteria: Production-ready performance + security validated
```

### 📊 Métriques Succès Global

#### **Acceptance Criteria Money Management**
```yaml
code_quality:
  coverage: 100%              # Aucune exception MM
  complexity: <10 par fonction # Lisibilité + maintenabilité  
  documentation: 100%         # Chaque fonction documentée

performance:
  trailing_stop_latency: <50ms     # Réactivité market
  circuit_breaker_latency: <5s     # Protection rapide
  throughput: >100 ops/sec         # Scalabilité
  memory_usage: <500MB/1k pos      # Efficacité ressources

reliability:
  uptime: 99.9%                    # Haute disponibilité
  data_integrity: 100%             # Zéro corruption
  financial_precision: ±0.0001%    # Exactitude calculs
  recovery_time: <60s              # Résilience

security:
  vulnerability_score: 0           # Aucune faille
  audit_trail: 100%               # Traçabilité complète  
  parameter_validation: 100%       # Robustesse inputs
  access_control: Role-based       # Sécurité accès
```

— Fin plan de tests Money Management —
