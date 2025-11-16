# Direction Engine - Moteur Temporel Backtest

## 🎯 Objectif

Backtester la stratégie **Direction** (VWMA6 + K-Confirmation) sur données Binance Vision avec un moteur temporel tick-par-tick, similaire au `scalping_engine`.

## 📋 Architecture

### Stratégie Direction

La stratégie détecte les **changements de direction du marché** basés sur:
- **VWMA6** : Volume Weighted Moving Average (période 3)
- **Pente VWMA** : Variation sur 2 bougies
- **Seuil dynamique** : ATR × coefficient (1.0)
- **K-Confirmation** : 2 bougies de confirmation

### Signaux

- **ENTRY LONG** : Pente VWMA > seuil (marché croissant)
- **ENTRY SHORT** : Pente VWMA < -seuil (marché décroissant)
- **EXIT** : Changement de direction détecté

### Cycle Moteur Temporel

```
Pour chaque trade dans l'ordre chronologique :
    ↓
1. Maintenir buffer trades (300 derniers)
    ↓
2. Détecter marqueur 5m (10:00:00, 10:05:00, etc.) ?
    ↓
   OUI → MARQUEUR DÉTECTÉ
    ↓
   a) Récupérer window klines (300 dernières)
    ↓
   b) Calculer VWMA6 + ATR
    ↓
   c) Détecter signaux via DirectionGenerator
    ↓
   d) Si signal ENTRY → Ouvrir position
    ↓
   e) Si signal EXIT → Fermer position
    ↓
3. NOTE: Direction n'a PAS de trailing stop
   - Position fermée UNIQUEMENT sur signal EXIT
   - Pas de gestion entre les marqueurs
    ↓
4. Continuer au trade suivant
```

## 🔧 Données Requises

### Klines (bougies 5m)
Pré-chargées en mémoire depuis fichiers Binance Vision:
```
Timestamp, Open, High, Low, Close, Volume
```

### Trades (tick-par-tick)
Streamés depuis fichiers Binance Vision:
```
Timestamp, Price, Quantity
```

## ⚙️ Configuration

### Paramètres Direction (hardcodés)
```go
VWMA_PERIOD           = 3
SLOPE_PERIOD          = 2
K_CONFIRMATION        = 2
USE_DYNAMIC_THRESHOLD = true
ATR_PERIOD            = 14
ATR_COEFFICIENT       = 1.0
```

### Gestion Positions
- **Ouverture** : Signal ENTRY du générateur
- **Fermeture** : Signal EXIT du générateur
- **Pas de trailing stop** : Positions tenues jusqu'au signal de sortie

## 🚀 Utilisation

### Lancer le backtest

```bash
go run cmd/direction_engine/main.go \
  -config config/config.yaml \
  -start 2024-11-01 \
  -end 2024-11-07 \
  -symbol SOLUSDT
```

### Structure config.yaml

```yaml
binance_data:
  cache_root: "./data/binance_vision"
  symbols: ["SOLUSDT"]
  
data_period:
  start_date: "2024-11-01"
  end_date: "2024-11-07"

backtest:
  window_size: 300
  trades_history_size: 300
  export_json: true
  export_path: "backtest_results"
  
  logging:
    enable_marker_logs: true
    enable_signal_logs: true
    enable_progress_logs: true
    enable_summary_logs: true
```

## 📊 Outputs

### Console

```
═══════════════════════════════════════════════════
  DIRECTION ENGINE - Moteur Temporel + Binance
═══════════════════════════════════════════════════

📝 Chargement configuration: config/config.yaml

⚙️  Paramètres Backtest:
   • Symbole: SOLUSDT
   • Période: 2024-11-01 → 2024-11-07
   • Timeframe: 5m (hardcodé)
   • VWMA: 3 (hardcodé)
   • K-Confirmation: 2 (hardcodé)
   • Cache: ./data/binance_vision
   • Jours à traiter: 7

🚀 Démarrage backtest - traitement trade par trade...

📂 Chargement klines...
✅ 2016 klines chargées

⚙️  Initialisation générateur direction...

🔄 Traitement trades en streaming...

📅 Date 1/7: 2024-11-01

🕐 10:00:00 | MARQUEUR DÉTECTÉ
   🎯 ENTRY LONG @ 161.50 (conf: 0.70)

🕐 14:35:00 | MARQUEUR DÉTECTÉ
   🎯 EXIT LONG @ 165.20 (conf: 0.75)

...

✅ Traitement terminé:
   • Trades: 1,250,000
   • Marqueurs: 2,016
   • Signaux: 85
   • Positions fermées: 42

═══════════════════════════════════════════════════
  RÉSULTATS BACKTEST DIRECTION
═══════════════════════════════════════════════════

📊 SIGNAUX:
   • Total: 85
   • ENTRY: 43
   • EXIT: 42
   • LONG: 21
   • SHORT: 22

💼 POSITIONS:
   • Fermées: 42
   • Gagnantes: 28 (66.7%)
   • Perdantes: 14

💰 VARIATIONS CAPTÉES:
   • LONG (↗)  : +2.33% total, +0.11% moyen
   • SHORT (↘) : -4.00% total, -0.18% moyen
   • TOTAL CAPTÉ: 6.33% (bidirectionnel)

📈 PERFORMANCE:
   • Max Win: +5.89%
   • Max Loss: -1.83%

═══════════════════════════════════════════════════

✅ Backtest terminé!
```

### Export JSON

Si `export_json: true`, génère:
```
backtest_results/direction_signals_20241108_150405.json
```

Contenu:
```json
{
  "timestamp": "2024-11-08T15:04:05Z",
  "signals": [
    {
      "Timestamp": "2024-11-01T10:00:00Z",
      "Type": "LONG",
      "Action": "ENTRY",
      "Price": 161.50,
      "Confidence": 0.70,
      "VWMA6": 161.20,
      "ATR": 1.45
    }
  ],
  "positions": [
    {
      "ID": 1,
      "Type": "LONG",
      "EntryTime": "2024-11-01T10:00:00Z",
      "EntryPrice": 161.50,
      "ExitTime": "2024-11-01T14:35:00Z",
      "ExitPrice": 165.20,
      "Duration": "4h35m",
      "PnLPercent": 2.29
    }
  ]
}
```

## 📝 Notes Importantes

1. **Anti-look-ahead** : Le générateur ne voit que les données passées
2. **Marqueurs précis** : Calculs uniquement aux timestamps 00:00 alignés sur 5m
3. **Window klines** : 300 dernières klines pour les indicateurs
4. **Trailing stop** : Mis à jour trade par trade (pas bougie par bougie)
5. **Volume SOL** : Utilisé pour VWMA (base asset)

## 🔗 Références

- **Générateur production** : `internal/signals/direction/generator.go`
- **Demo standalone** : `cmd/direction_generator_demo/main.go`
- **Architecture scalping** : `cmd/scalping_engine/` (modèle de référence)

## ⚠️ Différences avec direction_generator_demo

**direction_generator_demo :**
- Charge toutes les klines d'un coup
- Calcule tous les indicateurs
- Détecte tous les signaux
- Affichage résultats seulement

**direction_engine :**
- Traite trades tick-par-tick
- Calcule indicateurs aux marqueurs
- Ouvre/ferme positions sur signaux uniquement
- **PAS de trailing stop** - fermeture sur signal EXIT

## 🎯 Comparaison avec scalping_engine

| Aspect | scalping_engine | direction_engine |
|--------|----------------|------------------|
| **Stratégie** | Triple extrême (CCI+MFI+STOCH) | Direction VWMA6 |
| **Validation** | Multi-étapes (N-2 → N+2) | Immédiate (K-confirmation) |
| **Trailing Stop** | OUI (trade-par-trade) | NON (fermeture sur signal) |
| **Complexité** | Élevée (pending analyses) | Moyenne |
| **Timeframe** | 5m | 5m |
| **Signaux** | Rares, haute précision | Fréquents, capture vagues |

## TODO

- [ ] Ajouter `DirectionConfig` dans `internal/shared/config.go`
- [ ] Tester avec données réelles Binance Vision
- [ ] Ajouter métriques détaillées (drawdown, Sharpe ratio)
- [ ] Support multi-symboles
- [ ] Optimisation paramètres (VWMA period, K-confirmation)
