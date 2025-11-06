# Contraintes de développement Go

**Version:** 0.1  
**Statut:** Standards techniques obligatoires  
**Scope:** Contraintes architecturales et normes de développement Go

## Contraintes Architecturales

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

## Structure de modules Go standard

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

## Principes de décomposition
- **Séparation des préoccupations**: un module = une responsabilité
- **Interfaces claires**: contrats explicites entre modules
- **Réutilisabilité**: fonctions communes externalisées
- **Maintenabilité**: code lisible et documenté

## Standards de code
- **Nommage**: Go conventions (CamelCase, package names)
- **Documentation**: Godoc pour toutes les fonctions publiques
- **Gestion d'erreurs**: Toujours gérer les erreurs explicitement
- **Interfaces**: Préférer les petites interfaces spécialisées
- **Concurrence**: Utiliser channels et goroutines de façon idiomatique
