# Agent économique de trading — Compréhension et cadrage initial (à valider)

- Version: 0.1
- Statut: En attente de validation

## Vision
- **Vision**: Un « agent économique » modulaire d’exécution et de décision de trading, piloté par un contexte unifié, capable d’opérer en backtest, notification, paper et live, avec diagnostics désactivables, logs multi-niveaux et optimisation de paramètres.

## Composants clés
- **Marché**
  - **Kline Engine**: calcule les indicateurs à partir de bougies (multi-timeframes). Source: bougies natives ou reconstruites depuis les ticks. Sortie: `IndicatorSnapshot`.
  - **Tick Engine**: calcule tout le reste en temps réel (ex. microstructure, volatilité intraday, order flow/imbalance, détection d’événements) à partir des ticks/L2. Sortie: `TickAnalytics`.
- **Agrégateur de Contexte**
  - Fusionne `IndicatorSnapshot` + `TickAnalytics` + signaux + métadonnées d’environnement → `Context` versionné et horodaté.
- **Money Management (MM)**
  - Reçoit `Context` (tendance, signaux, paliers, objectifs, contraintes). Décide tailles, pyramiding, stops/TP, hedge. Émet des `ExecutionIntents`.
- **Exécution**
  - Traduit `ExecutionIntents` en ordres selon l’environnement (backtest/paper/live) via des adaptateurs de courtage/exchange.
- **Configuration**
  - Paramètres, objectifs (ex. Sharpe/DD), contraintes (levier/exposition), sélection d’environnement, variantes par environnement, toggles runtime (diagnostics/logs).
- **Diagnostics**
  - Traces étape-par-étape par composant (réplicables) avec snapshots d’état, fichiers structurés (ex. JSONL/CSV), identifiant de run, seed, hash de config. Activable/désactivable à chaud.
- **Logs**
  - Multi-niveaux (trace/debug/info/warn/error), structurés (JSON) avec corrélation (run-id, component-id), sampling possible en live.
- **Optimisation**
  - Recherche de paramètres optimaux sous contraintes via backtests: random/grid/Bayesian/évolutionnaire; objectifs définissables; reproductibilité (seed) et parallélisation.

## Variantes par environnement
- **Backtest**: sources historiques, exécution simulée, métriques complètes, diagnostics intensifs.
- **Notification de signaux**: pas d’ordres, émission d’alertes + logs/diagnostics légers.
- **Paper**: ordres simulés via broker sim, temps réel, latences réalistes.
- **Live**: connecteurs réels, contraintes de latence/fiabilité, logs/diagnostics ajustés.

## Flux principaux
1. Ingestion → `Kline Engine` → `IndicatorSnapshot` → `Context`.
2. Ingestion → `Tick Engine` → `TickAnalytics` → `Context`.
3. `Context` (+ signaux) → `MoneyManagement` → `ExecutionIntents` → `Execution` → ordres/trades.
4. Diagnostics + logs transverses à chaque étape.

## Contrats d’interface (proposés à valider)
- **MarketEvent**: tick/trade/L2 normalisé (timestamp, symbol, side, price, size, level/book).
- **IndicatorSnapshot**: timeframe, indicateur, valeur(s), window/meta.
- **TickAnalytics**: mesures temps-réel (volatilité realized, imbalance, microprice, events).
- **Context**: horodaté, agrégé, versionné; sections: trend, signaux, paliers/objectifs, risque, environnement.
- **ExecutionIntent**: action (open/close/scale/hedge), qty, limites, urgence, time-in-force.
- **OrderEvent/TradeEvent**: état d’ordres/trades normalisé.
- **BacktestReport**: KPIs, courbes, diagnostics, paramètres.
## Architecture (principes)
- **Ports & Adapters + DI**: échange des variantes par environnement sans toucher au cœur.
- **Event-driven**: topics pour `MarketEvent`, `IndicatorSnapshot`, `TickAnalytics`, `Context`, `ExecutionIntent`, `OrderEvent`.
- **Stockage**: time-series pour kline/indicateurs, data lake (parquet/feather) pour ticks; artefacts diagnostics versionnés par run.
- **Observabilité**: métriques, healthchecks, heartbeats; seuils d’alerte configurables.
- **Résilience**: idempotence, reprise sur crash, snapshots d'état pour backtests, time-travel.

## Contraintes Architecturales et Standards de Code

### Limites de fichiers
- **500 lignes maximum** par fichier source
- **Décomposition obligatoire** des fichiers volumineux en modules
- **Organisation par dossiers** pour les ensembles cohérents (ex: logs, diagnostics)

### Paradigmes de programmation
- **Éviter les pointeurs** sauf si absolument nécessaire pour les performances
- **Préférer les fonctions** pures et les structures de données immutables
- **Objectif unique** par fonction (principe de responsabilité unique)
- **Fonctions réutilisables** privilégiées lors de la décomposition

### Tests obligatoires
- **Test unitaire** pour chaque fonction
  - Vérification logique (comportement attendu)
  - Vérification syntaxique (types, interfaces)
- **Tests utilitaires** pour les modules complexes
- **Couverture** des cas d'erreur et cas limites

### Structure de modules complète (exemple Go)
```
internal/
├── market/
│   ├── kline_engine.go      # Calcul indicateurs depuis bougies
│   ├── tick_engine.go       # Microstructure, volatilité, order flow
│   └── interfaces.go        # MarketEvent, IndicatorSnapshot, TickAnalytics
├── context/
│   ├── aggregator.go        # Fusion IndicatorSnapshot + TickAnalytics + signaux
│   ├── types.go             # Context, versioning, horodatage
│   └── validator.go         # Validation cohérence du contexte
├── money_management/
│   ├── core.go              # Logic MM principale (tailles, stops, TP)
│   ├── risk.go              # Contraintes risque, levier, exposition
│   └── intent.go            # ExecutionIntent, pyramiding, hedge
├── execution/
│   ├── engine.go            # Traduction ExecutionIntent → ordres
│   ├── adapters/            # Variants par environnement
│   │   ├── backtest.go      # Simulation ordres/trades
│   │   ├── paper.go         # Broker simulé temps réel
│   │   ├── live.go          # Connecteurs exchanges réels
│   │   └── notification.go  # Émission alertes uniquement
│   └── types.go             # OrderEvent, TradeEvent, interfaces
├── config/
│   ├── agent.go             # Paramètres, objectifs, contraintes agent
│   ├── environment.go       # Sélection backtest/paper/live/notification
│   ├── optimization.go      # Config recherche paramètres optimaux
│   └── validation.go        # Validation configuration complète
├── diagnostics/
│   ├── tracer.go            # Traces étape-par-étape par composant
│   ├── snapshots.go         # Snapshots d'état, reproductibilité
│   ├── reporters.go         # Génération fichiers diagnostics
│   └── toggle.go            # Activation/désactivation à chaud
├── logging/
│   ├── core.go              # Logger multi-niveaux structuré JSON
│   ├── correlation.go       # Run-id, component-id, sampling
│   ├── handlers.go          # File, console, remote handlers
│   └── formatters.go        # Formatage spécifique par environnement
├── optimization/
│   ├── engine.go            # Moteur recherche paramètres optimaux
│   ├── strategies.go        # Random, grid, Bayesian, évolutionnaire
│   ├── objectives.go        # Sharpe, DD, turnover, multi-objectif
│   └── parallelization.go   # Backtests parallélisés, reproductibilité
├── data/
│   ├── binance/
│   │   ├── downloader.go    # Téléchargement Binance Data Vision
│   │   ├── cache.go         # Gestion cache local journalier
│   │   ├── klines.go        # Parsing klines, streaming ZIP
│   │   └── trades.go        # Parsing trades, décompression flux
│   └── types.go             # Structures communes données market
└── shared/
    ├── interfaces.go        # Interfaces communes entre composants
    ├── errors.go            # Types d'erreurs spécialisées
    └── utils.go             # Utilitaires réutilisables

cmd/
├── agent/                   # Point d'entrée principal
│   └── main.go
├── backtest/                # CLI backtest
│   └── main.go  
└── optimizer/               # CLI optimisation
    └── main.go

tests/
├── market_test.go
├── context_test.go
├── money_management_test.go
├── execution_test.go
├── diagnostics_test.go
├── integration/             # Tests d'intégration
│   ├── backtest_test.go
│   └── live_simulation_test.go
└── testutils/
    ├── fixtures.go          # Données de test
    ├── mocks.go             # Mocks des interfaces
    └── assertions.go        # Assertions spécialisées trading
```
### Principes de décomposition
- **Séparation des préoccupations**: un module = une responsabilité
- **Interfaces claires**: contrats explicites entre modules
- **Réutilisabilité**: fonctions communes externalisées
- **Maintenabilité**: code lisible et documenté

## Points à clarifier
1. Données tick: trade prints uniquement ou aussi order book L2/L3? Débit/latence cibles?
2. Bougies: multi-timeframes? Reconstruites depuis ticks ou prises de l’exchange? Liste initiale d’indicateurs?
3. Origine des signaux: produits par kline/tick internes ou injectés (externes)? Priorité en cas de conflit?
5. Règles MM: séparées des signaux (préféré) ou couplées? Contraintes strictes (max risk, max DD, max leverage) à quel niveau?
6. Objectifs d’optimisation: mono-objectif (ex. Sharpe) ou multi-objectif (Sharpe + DD + turnover)? Budget de calcul?
7. Environnements: « notification de signaux » = messages uniquement (email/queue/webhook) ou stockage + dashboard?
9. Logs: JSON structuré et agrégation (ELK/Loki/autre)? Niveaux par environnement?
10. Marchés: spot/futures/options? Multi-exchange? Symboles/périmètre initial?
11. Latence live: SLO de bout en bout (ingestion → ordre)? Stratégies sensibles à la latence?
12. Contraintes techniques: langage/stack imposés, OS/infra, délais et jalons?

## Spécifications Backtest — Binance Data Vision

### Données sources
- **Source**: Binance Data Vision uniquement
- **Type de futures**: USDT-M (USD-M futures)
- **Paires cibles**: `SOLUSDT`, `SUIUSDT`, `ETHUSDT`
- **Fuseau horaire**: UTC
- **Format des archives**: ZIP

### Klines (pour indicateurs)
- **Timeframes**: 5m, 15m, 1h, 4h
- **Période**: 01/06/2023 au 30/06/2025
- **Usage**: calcul d'indicateurs techniques uniquement

### Trades/Ticks (pour autres calculs)
- **Dataset**: trades des futures perpétuels USDT-M
- **Même périmètre**: paires et période identiques
- **Usage**: microstructure, volatilité réalisée, order flow, détection d'événements

### Stockage et cache local
- **Répertoire racine**: `data/`
- **Structure proposée**:
  ```
  data/
  ├── binance/
  │   └── futures_um/
  │       ├── klines/
  │       │   ├── SOLUSDT/
  │       │   │   ├── 5m/
  │       │   │   │   └── 2023/
  │       │   │   │       └── 06/
  │       │   │   │           ├── 01/
  │       │   │   │           │   └── SOLUSDT-5m-2023-06-01.zip
  │       │   │   │           └── ...
  │       │   │   ├── 15m/ ...
  │       │   │   ├── 1h/ ...
  │       │   │   └── 4h/ ...
  │       │   ├── SUIUSDT/ ...
  │       │   └── ETHUSDT/ ...
  │       └── trades/
  │           ├── SOLUSDT/
  │           │   └── 2023/
  │           │       └── 06/
  │           │           ├── 01/
  │           │           │   └── SOLUSDT-trades-2023-06-01.zip
  │           │           └── ...
  │           ├── SUIUSDT/ ...
  │           └── ETHUSDT/ ...
  ```

### Stratégie de téléchargement et lecture
- **Cache intelligent**: vérification d'existance avant téléchargement
- **Organisation journalière**: un fichier ZIP par jour et par paire
- **Lecture streaming**: décompression à la volée, pas de full-load en mémoire
- **Buffer glissant**: fenêtre minimale par timeframe pour les calculs
- **Vérification d'intégrité**: checksum si disponible dans Binance Data Vision
- **Reprise de téléchargement**: gestion des interruptions et reprises

### Pipeline de traitement backtest
1. **Index local**: manifeste des fichiers disponibles par date/paire/timeframe
2. **Téléchargement conditionnel**: si absent ou corrompu uniquement  
3. **Décompression streamée**: parsing en flux des ZIP
4. **Agrégation multi-timeframe**: reconstruction des bougies sans stockage complet
5. **Génération contexte**: fusion klines + trades pour alimenter le pipeline de décision

## Étape suivante proposée
- Si cette compréhension est validée, préparer:
  - C4 Niveau 1 (contexte) et 2 (conteneurs) du système.
  - La première matrice « Composant × Environnement » avec variantes.
  - Les brouillons de schémas d'interface (`Context`, `ExecutionIntent`, `MarketEvent`) pour validation rapide.
