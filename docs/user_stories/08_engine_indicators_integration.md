# User Stories 8: Intégration Engine-Indicateurs (Révisé)

**Epic:** Orchestration Engine ↔ Module Indicateurs (sans redondances)  
**Priorité:** Critique (P0) - **RÉVISÉ FOCUS INTÉGRATION**
**Estimation:** 5 points (réduit - Engine gère déjà beaucoup)  
**Sprint:** 4

## Description générale

> **En tant qu'** Engine Temporel orchestrateur  
> **Je veux** appeler module Indicateurs de façon optimale  
> **Afin d'** obtenir signaux/événements sans duplication de responsabilités

## User Stories principales (révisées sans redondances)

### US-INT-001: Appel module Indicateurs depuis Engine
**Priorité:** P0 - Critique  
**Points:** 2

> **En tant qu'** Engine Temporel  
> **Je veux** appeler Calculate() du module Indicateurs aux marqueurs  
> **Afin d'** obtenir signaux stratégie et événements zones

#### Critères d'acceptation (pas de redondance validation)
- **✅ Appel optimal** : Engine prépare CalculationRequest sans validation redondante
- **✅ Window management** : Engine fournit window déjà validée (≥35 candles)
- **✅ Contexte position** : Engine fournit PositionContext courant
- **✅ Traitement response** : Engine traite CalculationResponse pour actions

#### Scénarios de test
```gherkin
Scenario: Request valide avec contexte position
  Given Engine Temporel au marqueur 10:15:00
  And position LONG ouverte entry_price=100, CCI_zone=OVERSOLD  
  And window 300 candles SOLUSDT 5m
  When je construis CalculationRequest
  Then request.Symbol = "SOLUSDT"
  And request.CurrentTime = 10:15:00 timestamp
  And request.CandleWindow contient exactement 300 candles ≤ 10:15:00
  And request.PositionContext.IsOpen = true
  And request.PositionContext.EntryCCIZone = OVERSOLD

Scenario: Response complète avec signaux
  Given CalculationRequest traité par Indicateurs
  When calculs MACD/CCI/DMI terminés
  Then CalculationResponse.Results contient valeurs complètes
  And CalculationResponse.Signals contient signaux générés
  And CalculationResponse.ZoneEvents contient événements détectés
  And CalculationResponse.CalculationTime < 50ms
  And CalculationResponse.RequestID correspond à request

Scenario: Gestion erreur données insuffisantes
  Given request avec seulement 20 candles
  When Indicateurs traite la request
  Then CalculationResponse.Success = false
  And CalculationResponse.Error.Type = ERROR_INSUFFICIENT_DATA
  And Engine Temporel applique stratégie recovery
  And retry automatique après accumulation plus de données
```

#### DoD (Definition of Done)
- [ ] Structures Request/Response implémentées et documentées
- [ ] Validation exhaustive côté Indicateurs
- [ ] Traçabilité complète des échanges
- [ ] Tests error handling tous scénarios

---

### US-INT-002: Orchestration cycles temporels
**Priorité:** P0 - Critique  
**Points:** 5

> **En tant qu'** Engine Temporel  
> **Je veux** déclencher les calculs indicateurs aux moments précis  
> **Afin de** maintenir synchronisation parfaite avec cycles d'exécution

#### Critères d'acceptation
- **✅ Déclenchement marqueurs** : Calculs seulement aux 00:00:00 bougies
- **✅ Synchronisation multi-TF** : Cohérence 5m/15m/1h/4h sur même timestamp
- **✅ Performance cycles** : Latence totale < 200ms par cycle
- **✅ Fallback robuste** : Gestion échecs calculs sans arrêt cycle

#### Scenarios de test
```gherkin
Scenario: Déclenchement calcul au marqueur bougie
  Given Engine en cycle trade 09:59:58, 09:59:59, 10:00:00
  When trade 10:00:00 traité (marqueur détecté)
  Then CalculationRequest envoyé aux Indicateurs
  And request.CurrentTime = 10:00:00
  And calculs MACD/CCI/DMI déclenchés
  And pas de calcul sur trades 09:59:58, 09:59:59

Scenario: Synchronisation multi-timeframes
  Given marqueur 10:00:00 (début bougie 5m)
  And aussi début bougie 15m (10:00:00 % 15min == 0)
  When calculs déclenchés
  Then request séparée pour chaque TF actif
  And même CurrentTime pour tous : 10:00:00
  And window adaptée par TF (300 candles 5m vs 300 candles 15m)
  And synchronisation résultats avant traitement signaux

Scenario: Performance cycle complet
  Given marqueur bougie détecté
  When cycle: Request → Calculs → Response → Traitement
  Then latence totale < 200ms
  And répartition: 10ms Request, 50ms Calculs, 10ms Response, 130ms Traitement
  And pas d'impact sur cycle suivant

Scenario: Fallback échec calcul
  Given calculs indicateurs échouent (timeout)
  When Engine reçoit erreur
  Then utilisation valeurs précédentes en cache
  And cycle continue avec signaux dégradés
  And retry calcul au prochain marqueur
  And log warning détaillé
```

#### DoD (Definition of Done)
- [ ] Déclenchement calculs synchronisé avec marqueurs
- [ ] Gestion multi-timeframes cohérente
- [ ] Performance cycle validée < 200ms
- [ ] Stratégies fallback robustes implémentées

---

### US-INT-003: Traitement événements zones en temps réel
**Priorité:** P0 - Critique  
**Points:** 6

> **En tant qu'** coordinateur zones actives  
> **Je veux** traiter les événements zones depuis Indicateurs vers Engine  
> **Afin d'** activer/désactiver surveillance selon règles métier

#### Critères d'acceptation
- **✅ Détection événements** : CCI zone inverse, MACD inverse, DI counter
- **✅ Activation zones** : Passage état inactive → active automatique
- **✅ Conditions profit** : Événements MACD/DI seulement si profit suffisant
- **✅ Monitoring continu** : Zones actives vérifiées à chaque trade

#### Scénarios de test
```gherkin
Scenario: Activation zone CCI inverse
  Given position LONG ouverte sur CCI oversold
  And calculs retournent CCI = +120 (overbought)
  When Engine traite ZoneEvent CCI_ZONE_ENTERED
  Then ZoneMonitor.activeZones["CCI_INVERSE"] = active
  And ActiveZone.Type = CCI_INVERSE
  And ActiveZone.EntryTime = current timestamp
  And surveillance déclenchée pour trades suivants

Scenario: Événement MACD inverse avec profit
  Given position avec profit courant = 8%
  And calculs détectent MACD croisement inverse
  And configuration profit_threshold = 5%
  When Engine traite ZoneEvent MACD_INVERSE_CROSS  
  Then événement MACD traité (profit > 5%)
  And ZoneMonitor déclenche ajustement si conditions grille
  And log "MACD inverse event processed: 8% profit > 5% threshold"

Scenario: Événement DI rejeté profit insuffisant
  Given position avec profit = 3%
  And calculs détectent DI contre-tendance
  And configuration profit_threshold = 5%
  When Engine traite ZoneEvent DI_COUNTER_CROSS
  Then événement ignoré (profit < 5%)
  And aucun ajustement stop déclenché
  And log "DI counter event ignored: 3% profit < 5% threshold"

Scenario: Désactivation zone sortie CCI
  Given zone CCI_INVERSE active (CCI en overbought)
  And calculs retournent CCI = +90 (normal)
  When Engine traite ZoneEvent CCI_ZONE_EXITED
  Then ZoneMonitor.activeZones["CCI_INVERSE"] = inactive
  And arrêt surveillance continue
  And log "CCI zone exited, monitoring stopped"
```

#### DoD (Definition of Done)
- [ ] Traitement tous types ZoneEvents
- [ ] Logique activation/désactivation zones complète
- [ ] Validation conditions profit pour événements
- [ ] Tests scenarios complexes multi-zones

---

### US-INT-004: Cache intelligent et optimisation
**Priorité:** P1 - Important  
**Points:** 4

> **En tant qu'** optimiseur de performance  
> **Je veux** éviter recalculs inutiles d'indicateurs  
> **Afin d'** améliorer performance sans compromettre précision

#### Critères d'acceptation
- **✅ Cache conditionnel** : Réutilisation si mêmes données + config
- **✅ Invalidation intelligente** : Cache invalide si nouvelles données
- **✅ Performance améliorée** : Gain 50%+ sur cycles sans nouvelles données
- **✅ Traçabilité cache** : Hits/misses dans métriques

#### Scénarios de test
```gherkin
Scenario: Cache hit données identiques
  Given CalculationRequest précédente avec window 300 candles
  And nouvelle request avec mêmes candles + config
  And CurrentTime inchangé depuis dernier calcul
  When je vérifie le cache
  Then cache hit détecté
  And CalculationResponse retournée depuis cache < 5ms
  And pas de recalcul MACD/CCI/DMI

Scenario: Cache miss nouvelles données
  Given cache contient calculs pour timestamp 10:00:00
  And nouvelle request pour timestamp 10:05:00
  When je vérifie le cache
  Then cache miss détecté (nouvelles données)
  And calculs complets MACD/CCI/DMI relancés
  And cache mis à jour avec nouveaux résultats

Scenario: Cache invalide changement config
  Given cache valide pour config CCI period=14
  And nouvelle request avec config CCI period=20
  When je vérifie le cache
  Then cache invalide (configuration différente)
  And recalcul complet avec nouvelle config
  And métrique cache_invalidation_config++
```

#### DoD (Definition of Done)
- [ ] Système cache avec validation intelligente
- [ ] Métriques hits/misses/invalidations
- [ ] Gain performance > 50% scenarios cache hit
- [ ] Tests cache tous edge cases

---

### US-INT-005: Validation end-to-end stratégie complète
**Priorité:** P0 - Critique  
**Points:** 4

> **En tant que** système complet  
> **Je veux** valider l'exécution stratégie de bout en bout  
> **Afin de** garantir conformité règles métier dans conditions réelles

#### Critères d'acceptation
- **✅ Scénario complet** : Données → Calculs → Signaux → Position → Zones → Ajustements
- **✅ Conformité stratégie** : 100% respect strategy_macd_cci_dmi_pure.md
- **✅ Performance réaliste** : Validation sur données historiques complètes
- **✅ Robustesse** : Gestion edge cases et conditions dégradées

#### Scénarios de test
```gherkin
Scenario: Exécution stratégie LONG complète
  Given données SOLUSDT 5m du 2023-06-01
  And Engine en mode backtest
  When cycle complet exécuté
  Then détection signal LONG: MACD↗ + CCI oversold + DI+ > DI-
  And position ouverte avec stops configurés
  And surveillance zones activée
  And ajustements stops selon grille quand profit + zones
  And fermeture position stop hit ou sortie anticipée
  And log complet traçable de A à Z

Scenario: Performance données réelles
  Given 30 jours données SOLUSDT/SUIUSDT/ETHUSDT
  And tous timeframes 5m/15m/1h/4h
  When stratégie exécutée mode backtest
  Then latence moyenne < 200ms par cycle
  And mémoire stable < 500MB
  And 0 look-ahead détecté
  And signaux générés conformes règles

Scenario: Robustesse conditions dégradées
  Given données avec gaps occasionnels
  And échecs calculs simulés (5% des cycles)
  When stratégie exécutée
  Then fallback automatique vers cache/valeurs précédentes
  And continuation cycle sans interruption
  And signaux dégradés mais cohérents
  And recovery automatique cycles suivants
```

#### DoD (Definition of Done)
- [ ] Tests end-to-end avec données réelles multi-symboles
- [ ] Validation conformité stratégie 100%
- [ ] Métriques performance en conditions réelles
- [ ] Robustesse validée scenarios edge

---

## Epic Summary

### Architecture intégration
```
Engine Temporel:
├── Cycle principal (trade/loop)
├── Marqueur bougie détecté → CalculationRequest
├── Response reçue → Traitement signaux + zones
├── Position management + zones actives
└── Monitoring continu ajustements

Indicateurs Techniques:
├── Receive CalculationRequest
├── Validate données + config
├── Calculate MACD/CCI/DMI
├── Generate signaux + zone events
└── Return CalculationResponse

Integration Layer:
├── Request/Response structures
├── Cache intelligent
├── Error handling + recovery
└── Performance monitoring
```

### Flux de données complet
```
Données historiques → Engine Temporel → CalculationRequest
                                    ↓
                              Indicateurs Techniques
                                    ↓
                    CalculationResponse ← Engine Temporel
                         ↓
              Signaux + ZoneEvents traités
                         ↓
         Position management + Zones actives
                         ↓
              Ajustements stops + cycles
```

### Dépendances
- **Dépend de:** US-ENG-001 à US-ENG-004 (Engine Temporel)
- **Dépend de:** US-IND-001 à US-IND-005 (Indicateurs)
- **Bloque:** Déploiement stratégie complète
- **Interface avec:** CLI, configuration, monitoring

### Métriques de succès
- **Performance** : < 200ms latence end-to-end
- **Cache efficiency** : > 50% hits en conditions normales
- **Robustesse** : < 0.1% échecs critiques sur 1M+ cycles
- **Conformité** : 100% respect règles stratégie

### Critères acceptation Epic
- [ ] Communication Engine ↔ Indicateurs opérationnelle
- [ ] Orchestration cycles temporels synchronisée
- [ ] Événements zones traités en temps réel
- [ ] Cache intelligent fonctionnel
- [ ] Validation end-to-end stratégie complète
- [ ] Performance + robustesse validées données réelles
- [ ] Documentation intégration complète

### Configuration intégration
```yaml
integration:
  # Communication
  request_timeout: 5000ms
  max_retries: 3
  
  # Cache
  cache_enabled: true
  cache_ttl: 1000ms
  cache_max_size: 100
  
  # Performance
  parallel_symbols: true
  max_goroutines: 10
  
  # Monitoring  
  metrics_enabled: true
  detailed_logging: false
  trace_requests: true
```

### Risques identifiés
- **Performance** : Latence cumulative cycles → optimisation cache critique
- **Synchronisation** : Multi-TF complex → tests exhaustifs requis
- **Robustesse** : Échecs cascades → fallback strategies multiples nécessaires

---

*Version 1.0 - Intégration Engine-Indicateurs : Orchestration fluide et performance optimisée*
