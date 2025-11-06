# User Stories 6: Engine Temporel et Gestion Position

**Epic:** Engine Temporel - Orchestration cycles et positions  
**Priorité:** Critique (P0)  
**Estimation:** 21 points  
**Sprint:** 4-5

## Description générale

> **En tant qu'** agent économique de trading  
> **Je veux** un engine temporel qui simule parfaitement le présent  
> **Afin de** exécuter ma stratégie sans biais look-ahead dans tous les modes

## User Stories principales

### US-ENG-001: Présent artificiel cohérent
**Priorité:** P0 - Critique  
**Points:** 5

> **En tant que** simulateur de backtest  
> **Je veux** maintenir un Current_Timestamp cohérent dans tous les modes  
> **Afin de** simuler parfaitement le présent sans look-ahead bias

#### Critères d'acceptation
- **✅ Current_Timestamp unique** pour tous les composants
- **✅ Anti-look-ahead automatique** : jamais de données > Current_Timestamp
- **✅ Historiques complets** avec accès window configuré (300 candles)
- **✅ Cohérence backtest/paper/live** : même logique temporelle

#### Scénarios de test
```gherkin
Scenario: Maintien cohérence temporelle en backtest
  Given un engine temporel en mode backtest
  And des données historiques du 2023-06-01 au 2023-06-30
  When je démarre le cycle sur trade timestamp 2023-06-01 10:15:30
  Then Current_Timestamp = 2023-06-01 10:15:30
  And aucune donnée > 10:15:30 n'est accessible
  And historique contient toutes données ≤ 10:15:30

Scenario: Validation anti-look-ahead automatique
  Given Current_Timestamp = 2023-06-01 14:30:00
  When je demande des données klines
  Then seules les klines ≤ 14:30:00 sont retournées
  And une erreur est levée si tentative d'accès données futures

Scenario: Synchronisation modes backtest/live
  Given même configuration temporelle
  When j'exécute en mode backtest puis live
  Then la logique de gestion timestamp est identique
  And les décisions prises sont cohérentes
```

#### DoD (Definition of Done)
- [ ] TemporalEngine implémenté avec Current_Timestamp centralisé
- [ ] Validation anti-look-ahead sur tous accès données
- [ ] Tests unitaires couvrant tous modes temporels
- [ ] Performance < 10ms par mise à jour timestamp

---

### US-ENG-002: Cycles d'exécution adaptatifs
**Priorité:** P0 - Critique  
**Points:** 8

> **En tant que** système multi-environnements  
> **Je veux** adapter les cycles d'exécution selon le mode  
> **Afin d'** optimiser performance vs réalisme selon le contexte

#### Critères d'acceptation
- **✅ Mode Backtest** : Cycle trade par trade avec marqueurs bougies 00:00:00
- **✅ Mode Paper/Live** : Loop 10 secondes avec détection nouvelles bougies
- **✅ Marqueurs intelligents** : Détection précise débuts de bougies
- **✅ Performance adaptée** : Optimisation selon mode

#### Scénarios de test
```gherkin
Scenario: Cycle backtest trade par trade
  Given engine en mode backtest
  And trades chronologiques SOLUSDT du 2023-06-01
  When je démarre le cycle
  Then chaque trade est traité séquentiellement
  And Current_Timestamp = trade.timestamp à chaque cycle
  And marqueur détecté à 10:00:00, 10:05:00, 10:10:00 (5m)

Scenario: Cycle live avec loop 10 secondes
  Given engine en mode live
  When je démarre le cycle
  Then une boucle s'exécute toutes les 10 secondes
  And Current_Timestamp = NOW à chaque iteration
  And détection automatique nouvelles bougies depuis dernière loop

Scenario: Marqueurs bougies précis
  Given trades: 09:59:58, 09:59:59, 10:00:00, 10:00:01
  When le cycle traite ces trades
  Then marqueur détecté seulement à 10:00:00
  And calculs indicateurs déclenchés au marqueur
  And pas de recalcul sur autres trades
```

#### DoD (Definition of Done)
- [ ] ExecutionModes implémentés (BacktestMode, LiveMode)
- [ ] Détection marqueurs bougies avec précision milliseconde
- [ ] Performance : < 1ms par trade (backtest), < 100ms per loop (live)
- [ ] Tests stress avec 100k+ trades

---

### US-ENG-003: Position unique avec états complets
**Priorité:** P0 - Critique  
**Points:** 5

> **En tant qu'** algorithme de trading  
> **Je veux** gérer une seule position avec tous ses états  
> **Afin de** simplifier la logique et éviter les conflits de positions

#### Critères d'acceptation
- **✅ Position unique** : Maximum 1 position ouverte simultanément
- **✅ État complet** : Price, time, direction, stops, contexte CCI
- **✅ Ouverture contrôlée** : Seulement aux signaux d'indicateurs
- **✅ Fermeture automatique** : Stop hit ou sortie anticipée MACD

#### Scénarios de test
```gherkin
Scenario: Ouverture position unique
  Given position fermée
  And signal LONG généré aux indicateurs
  When je traite le signal d'ouverture
  Then position s'ouvre avec état complet
  And Position.IsOpen = true
  And Position.Direction = LONG
  And Position.EntryPrice, EntryTime correctement définis

Scenario: Blocage multi-positions
  Given position LONG ouverte
  And nouveau signal SHORT généré
  When je traite le nouveau signal
  Then signal SHORT est ignoré
  And position LONG reste ouverte inchangée
  And log "position already open, signal ignored"

Scenario: Fermeture automatique stop
  Given position LONG ouverte entry_price = 100, stop = 98
  And prix courant = 97.5 (stop hit)
  When je vérifie les conditions fermeture
  Then position se ferme automatiquement
  And Position.IsOpen = false
  And log de fermeture avec détails P&L
```

#### DoD (Definition of Done)
- [ ] PositionManager avec état Position complet
- [ ] Validation blocage multi-positions
- [ ] Calculs P&L précis
- [ ] Tests edge cases (gaps, prix extrêmes)

---

### US-ENG-004: Zones actives extensibles
**Priorité:** P0 - Critique  
**Points:** 8

> **En tant que** système de monitoring avancé  
> **Je veux** surveiller des zones d'indicateurs activées  
> **Afin d'** déclencher des ajustements de stops à chaque trade

#### Critères d'acceptation
- **✅ Zones configurables** : CCI inverse, MACD inverse, DI contre-tendance
- **✅ Activation dynamique** : Selon événements indicateurs
- **✅ Monitoring continu** : Vérification à chaque trade si zone active
- **✅ Extensibilité** : Nouveaux types zones facilement ajoutables

#### Scénarios de test
```gherkin
Scenario: Activation zone CCI inverse
  Given position LONG ouverte sur CCI survente (-110)
  And CCI passe en surachat (+120)
  When l'événement CCI zone inverse est détecté
  Then zone CCI_INVERSE devient active
  And ZoneMonitor.activeZones["CCI_INVERSE"].Active = true
  And surveillance déclenchée à chaque trade

Scenario: Monitoring continu zone active
  Given zone CCI_INVERSE active
  And position avec profit 8%
  When nouveau trade arrive
  Then vérification: CCI toujours en zone surachat?
  And calcul % profit courant
  And SI conditions grille atteintes → ajustement stop
  And logs détaillés du monitoring

Scenario: Désactivation zone
  Given zone CCI_INVERSE active (CCI en surachat)
  And CCI redescend à +80 (sort de zone surachat)
  When la sortie de zone est détectée
  Then zone CCI_INVERSE devient inactive
  And arrêt du monitoring continu
  And log "CCI zone exited, monitoring stopped"

Scenario: Extensibilité nouveaux types
  Given architecture ZoneMonitor
  When j'ajoute un nouveau type RSI_DIVERGENCE
  Then même pattern activation/monitoring/désactivation
  And configuration via YAML
  And tests identiques à CCI
```

#### DoD (Definition of Done)
- [ ] ZoneMonitor avec pattern extensible
- [ ] Types zones : CCI_INVERSE, MACD_INVERSE, DI_COUNTER
- [ ] Configuration complète via YAML
- [ ] Framework tests pour nouveaux types zones

---

### US-ENG-005: Ajustements stops sophistiqués
**Priorité:** P1 - Important  
**Points:** 5

> **En tant que** gestionnaire de risque avancé  
> **Je veux** ajuster dynamiquement les stops selon les indicateurs  
> **Afin d'** optimiser la capture de profits selon les règles métier

#### Critères d'acceptation
- **✅ Grille d'ajustement** : % profit → % nouveau trailing stop
- **✅ Application conditionnelle** : Seulement si zone active + conditions
- **✅ Stops plus serrés** : Valeurs décroissantes selon profit
- **✅ Validation métier** : Conforme doc strategy_macd_cci_dmi_pure.md

#### Scénarios de test
```gherkin
Scenario: Application grille ajustement
  Given position LONG entry_price = 100, profit 12%
  And zone CCI_INVERSE active
  And grille config: 10-20% profit → trailing stop 1.0%
  When conditions d'ajustement sont vérifiées
  Then trailing stop passe de 2.0% à 1.0%
  And nouveau stop = 112 * (1 - 0.01) = 110.88
  And log "trailing stop adjusted: 2.0% → 1.0% (12% profit)"

Scenario: Validation seulement si zone active
  Given position avec profit 15%
  And aucune zone active
  When je vérifie conditions ajustement
  Then aucun ajustement appliqué
  And trailing stop reste inchangé
  And log "no active zones, standard trailing stop maintained"
```

#### DoD (Definition of Done)
- [ ] Implémentation grille trailing_stop_adjustment_grid
- [ ] Intégration avec ZoneMonitor
- [ ] Validation conformité règles métier
- [ ] Tests profit/loss scenarios complets

---

## Epic Summary

### Dépendances
- **Bloque:** Module Indicateurs (calculs MACD/CCI/DMI)
- **Dépend de:** Infrastructure données (parsers, cache)
- **Interface avec:** CLI workflow, configuration

### Métriques de succès
- **Performance** : < 50ms cycle backtest, < 200ms cycle live
- **Fiabilité** : 0 look-ahead détecté sur 1M+ trades
- **Précision** : 100% conformité règles position/stops
- **Extensibilité** : Nouveau type zone ajouté < 2h dev

### Critères acceptation Epic
- [ ] Engine temporel fonctionnel tous modes
- [ ] Position unique avec gestion complète
- [ ] Zones actives extensibles opérationnelles
- [ ] Performance validée environnement cible
- [ ] Tests end-to-end avec données réelles
- [ ] Documentation technique complète

### Risques identifiés
- **Performance** : Monitoring zones très fréquent → optimisation requise
- **Complexité** : Interaction zones multiples → tests exhaustifs nécessaires
- **Temporel** : Synchronisation multi-TF → validation précision critique

---

*Version 1.0 - Engine Temporel : Orchestration temporelle réaliste et gestion position unique avancée*
