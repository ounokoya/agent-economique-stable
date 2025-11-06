# User Stories 7: Module Indicateurs - Interface & Signal Generator

**Epic:** Interface Engine ↔ Indicateurs + Signal Generator pur  
**Priorité:** Critique (P0) - **RÉVISÉ SANS REDONDANCES**
**Estimation:** 8 points (réduit - pas de redondance Engine)  
**Sprint:** 3

## Description générale

> **En tant que** module Indicateurs stateless  
> **Je veux** fournir interface claire + signal generator pur  
> **Afin de** servir l'Engine Temporel sans duplication de responsabilités

## User Stories principales (révisées)

### US-IND-001: Interface Communication Engine ↔ Indicateurs  
**Priorité:** P0 - Critique  
**Points:** 3

> **En tant qu'** interface de communication  
> **Je veux** structures CalculationRequest/Response standardisées  
> **Afin de** permettre échange fluide avec Engine Temporel

#### Critères d'acceptation (pas de redondance)
- **✅ Structures interface** : CalculationRequest/Response définies
- **✅ Pas de validation** : Engine fournit données déjà propres (≥35 candles)
- **✅ Utilise types Engine** : ZoneEvent, PositionContext existants
- **✅ Performance** : < 5ms assemblage request/response

#### Scénarios de test (révisés)
```gherkin
Scenario: Interface Request/Response
  Given CalculationRequest avec CandleWindow[300] propre (Engine validé)
  And PositionContext avec position LONG ouverte
  When module Indicateurs traite request
  Then CalculationResponse.Success = true
  And CalculationResponse.Results contient MACD/CCI/DMI
  And CalculationResponse.Signals contient signaux générés
  And CalculationResponse.ZoneEvents utilise types Engine existants

Scenario: Pas de validation redondante
  Given Engine fournit déjà données >= 35 candles validées
  When module reçoit CalculationRequest
  Then aucune validation supplémentaire effectuée
  And calculs démarrent directement sur CandleWindow
```

#### DoD (Definition of Done)
- [ ] CalculationRequest/Response structures définies
- [ ] Interface utilise types Engine existants (ZoneEvent, PositionContext)
- [ ] Aucune duplication de validation/window management
- [ ] Performance < 5ms assemblage request/response

---

### US-IND-002: Signal Generator Stratégie MACD/CCI/DMI  
**Priorité:** P0 - Critique  
**Points:** 3

> **En tant que** générateur signaux purs  
> **Je veux** implémenter logique stratégie MACD/CCI/DMI exacte  
> **Afin de** fournir signaux conformes à la mémoire utilisateur

#### Critères d'acceptation (logique pure)
- **✅ Stratégie exacte** : MACD croisements + CCI zones + DMI conditions
- **✅ Signaux LONG/SHORT** : Tendance/contre-tendance selon DI+/DI-
- **✅ Pas de gestion position** : Engine s'occupe ouverture/fermeture
- **✅ Calculs stateless** : Utilise indicateurs déjà calculés

#### Scénarios de test
```gherkin
Scenario: Signal LONG tendance (mémoire utilisateur)
  Given MACD croise à la hausse (CrossUp)
  And CCI en survente (< -100)  
  And DI+ > DI- (tendance haussière)
  When generateStrategySignals() appelé
  Then signal LONG tendance généré
  And Direction = LongSignal, Type = TrendSignal

Scenario: Signal SHORT contre-tendance (mémoire utilisateur)
  Given MACD croise à la baisse (CrossDown)
  And CCI en surachat (> +100)
  And DI+ > DI- (mais contre-tendance SHORT)
  When generateStrategySignals() appelé  
  Then signal SHORT contre-tendance généré
  And Direction = ShortSignal, Type = CounterTrendSignal
  Then CCIValues.Zone = OVERBOUGHT
  And CCIValues.IsExtreme = true
  And CCIValues.SignalType = COUNTER_TREND

Scenario: Événement entrée zone inverse
  Given position LONG ouverte sur CCI oversold (-110)
  And CCI passe à +120 (overbought)
  When je détecte les événements zones
  Then ZoneEvent généré: CCI_ZONE_ENTERED
  And event.IsInverse = true
  And event.CurrentZone = OVERBOUGHT
```

#### DoD (Definition of Done)
- [ ] CCICalculator avec seuils configurables
- [ ] Détection zones précise selon type signal
- [ ] Génération événements pour Engine Temporel
- [ ] Tests tous seuils de configuration

---

### US-IND-003: DMI tendance et contre-tendance
**Priorité:** P0 - Critique  
**Points:** 4

> **En tant qu'** analyseur de tendance  
> **Je veux** calculer DI+/DI-/ADX et déterminer direction/force  
> **Afin de** filtrer les signaux selon contexte tendance

#### Critères d'acceptation
- **✅ Calcul DMI(14)** : DI+, DI-, DX, ADX précis
- **✅ Direction tendance** : BULLISH, BEARISH, SIDEWAYS selon DI+/DI-
- **✅ Force tendance** : WEAK/MODERATE/STRONG selon ADX
- **✅ Croisements DI** : Détection changements direction

#### Scénarios de test
```gherkin
Scenario: Calcul DMI précis
  Given candles avec High, Low, Close
  And période DMI = 14
  When je calcule DMI
  Then +DI = 100 * SMA(+DM, 14) / SMA(TR, 14)
  And -DI = 100 * SMA(-DM, 14) / SMA(TR, 14)  
  And DX = 100 * abs(+DI - -DI) / (+DI + -DI)
  And ADX = SMA(DX, 14)

Scenario: Détection tendance bullish forte
  Given DI+ = 35, DI- = 15, ADX = 40
  When j'analyse la tendance
  Then DMIValues.TrendDirection = BULLISH
  And DMIValues.TrendStrength = STRONG
  And justification: DI+ > DI- ET ADX > 25

Scenario: Détection croisement DI
  Given DI+ précédent = 20, DI- précédent = 25
  And DI+ courant = 28, DI- courant = 22
  When je détecte les croisements
  Then DMIValues.DIsCrossed = true
  And DMIValues.CrossDirection = UP
  And événement DI_CROSS généré
```

#### DoD (Definition of Done)
- [ ] DMICalculator avec calculs standard validés
- [ ] Classification tendance/force automatique
- [ ] Détection croisements DI robuste
- [ ] Tests edge cases (valeurs extrêmes)

---

### US-IND-004: Génération signaux stratégie
**Priorité:** P0 - Critique  
**Points:** 6

> **En tant que** générateur de signaux  
> **Je veux** produire signaux LONG/SHORT selon combinaisons MACD+CCI+DMI  
> **Afin de** respecter strictement les règles de la stratégie

#### Critères d'acceptation
- **✅ Règles LONG** : MACD croise ↗ + CCI zone appropriée + DMI direction
- **✅ Règles SHORT** : MACD croise ↘ + CCI zone appropriée + DMI direction  
- **✅ Filtres optionnels** : MACD même signe, DX/ADX selon config
- **✅ Score confiance** : Calcul selon qualité signaux

#### Scénarios de test
```gherkin
Scenario: Signal LONG tendance valide
  Given MACD croise à la hausse (CrossedUp = true)
  And CCI = -110 (oversold tendance, seuil -100)
  And DI+ = 30, DI- = 20 (DI+ > DI- pour tendance)
  And filtres: macd_same_sign_filter = false
  When je génère les signaux
  Then StrategySignal créé avec Type = LONG_ENTRY
  And Confidence >= 80%
  And TriggerReason = "MACD cross up + CCI oversold + DI bullish trend"

Scenario: Signal SHORT contre-tendance valide
  Given MACD croise à la baisse (CrossedDown = true)
  And CCI = +190 (overbought contre-tendance, seuil +180)
  And DI+ = 25, DI- = 15 (DI+ > DI- pour contre-tendance)
  And filtres activés selon config
  When je génère les signaux
  Then StrategySignal créé avec Type = SHORT_ENTRY
  And métadonnées complètes des indicateurs incluses

Scenario: Signal rejeté par filtres
  Given conditions MACD + CCI validés
  And DMI ne respecte pas direction requise
  When je génère les signaux
  Then aucun signal généré
  And log "signal rejected: DMI direction mismatch"
  And FiltersBlocked = ["dmi_direction"]

Scenario: Score confiance dégradé
  Given signal LONG valide
  And MACD amplitude faible
  And CCI proche seuil (-105 vs -100)
  And ADX = 20 (faible < 25)
  When je calcule le score
  Then Confidence = 100% - 10% - 15% - 20% = 55%
  And détail pénalités dans métadonnées
```

#### DoD (Definition of Done)
- [ ] SignalGenerator avec règles stratégie complètes
- [ ] Filtres optionnels configurables
- [ ] Calcul score confiance avec pénalités
- [ ] Tests tous cas edge et rejets

---

### US-IND-005: API standardisée et extensible
**Priorité:** P1 - Important  
**Points:** 3

> **En tant que** module calculatoire  
> **Je veux** fournir une API claire et extensible  
> **Afin de** faciliter l'intégration avec Engine Temporel et futurs indicateurs

#### Critères d'acceptation
- **✅ Interface commune** : IndicatorCalculator pour tous types
- **✅ Request/Response** : Structures standardisées
- **✅ Validation entrée** : Données suffisantes, format correct
- **✅ Extensibilité** : Nouveaux indicateurs facilement ajoutables

#### Scénarios de test
```gherkin
Scenario: Request standardisé
  Given CalculationRequest avec symbol, timeframe, candles
  And configuration indicateurs valide
  When j'appelle ProcessCalculationRequest()
  Then CalculationResponse retourné avec résultats complets
  And métadonnées performance incluses
  And pas d'état persistant dans le calculateur

Scenario: Validation données insuffisantes
  Given request avec seulement 20 candles
  And MACD requiert minimum 26 + 9 = 35 candles
  When j'appelle Calculate()
  Then error ERROR_INSUFFICIENT_DATA retourné
  And message explicite: "need 35 candles, got 20"

Scenario: Extensibilité nouvel indicateur
  Given interface IndicatorCalculator
  When j'implémente RSICalculator
  Then même pattern: Calculate(), ValidateInput(), GetMinimumPeriod()
  And intégration transparente dans SignalGenerator
  And tests identiques applicable
```

#### DoD (Definition of Done)
- [ ] Interface IndicatorCalculator bien définie
- [ ] Structures Request/Response complètes
- [ ] Validation robuste entrées
- [ ] Documentation API pour extensibilité

---

## Epic Summary

### Architecture technique
```
CalculationRequest → [MACDCalculator] → MACDValues
                  → [CCICalculator]  → CCIValues + ZoneEvents  
                  → [DMICalculator]  → DMIValues
                                   ↓
                     [SignalGenerator] → StrategySignals
                                   ↓
                    CalculationResponse → Engine Temporel
```

### Dépendances
- **Dépend de:** Structures données (Kline, configuration)
- **Bloque:** Engine Temporel (interface requise)
- **Interface avec:** Configuration YAML, tests validation

### Métriques de succès
- **Précision** : < 0.001% erreur vs références externes
- **Performance** : < 50ms calcul complet (MACD+CCI+DMI+signaux)
- **Fiabilité** : 100% conformité règles stratégie sur 10k+ signaux
- **Couverture** : > 95% tests unitaires

### Critères acceptation Epic
- [ ] Tous indicateurs (MACD/CCI/DMI) calculés avec précision
- [ ] Génération signaux conforme strategy_macd_cci_dmi_pure.md
- [ ] Événements zones détectés pour Engine Temporel
- [ ] API standardisée documentée et testée
- [ ] Performance validée sur gros volumes
- [ ] Extensibilité démontrée avec nouvel indicateur test

### Configuration référence
```yaml
indicators:
  macd:
    fast_period: 12
    slow_period: 26
    signal_period: 9
    
  cci:
    period: 14
    long_trend_oversold: -100
    long_trend_overbought: 100
    long_counter_trend_oversold: -150
    long_counter_trend_overbought: 150
    short_trend_oversold: -120
    short_trend_overbought: 120
    short_counter_trend_oversold: -180
    short_counter_trend_overbought: 180
    
  dmi:
    period: 14
    adx_period: 14
    trend_threshold: 25

signal_generation:
  filters:
    macd_same_sign_filter: false
    dmi_trend_signals_enabled: true
    dmi_counter_trend_signals_enabled: false
    dx_adx_filter_enabled: false
```

### Risques identifiés
- **Précision** : Différences algorithmes EMA entre références → validation extensive requise
- **Performance** : Calculs sur 300+ candles répétitifs → optimisation cache nécessaire  
- **Edge cases** : Données manquantes/corrompues → robustesse validation critique

---

*Version 1.0 - Indicateurs Techniques : Calculs précis MACD/CCI/DMI et génération signaux conformes*
