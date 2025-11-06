# MFI+Stoch Extended - Live Trading Deployment

## 📊 Stratégie

**MFI+Stoch Extended** est la stratégie **la plus performante** basée sur les backtests 2024:
- **Double Confluence**: MFI + Stochastic (pas de CCI)
- **MM Extended**: Re-entry après stop loss si conditions maintenues
- **Paramètres Optimisés**: MFI=16, Stoch=16,3,3

## 🎯 Performances Backtestées (SOLUSDT 2024)

| Timeframe | Return | Max DD | Trades | Win Rate | Status |
|-----------|--------|--------|--------|----------|--------|
| **15m** | **+6049%** | 6.27% | 4,764 | 85.4% | 🥇 **CHAMPION** |
| **1m** | **+3622%** | 7.95% | 52,378 | 77.2% | ✅ Excellent |
| **5m** | Non testé | - | - | - | ⚠️ À tester |

## 🚀 Déploiement

### 1. Compiler le binaire

```bash
cd backend
go build -o /root/data/backtest-optimizer/mfi_stoch_live_notifications cmd/mfi_stoch_live_notifications/main.go
```

### 2. Déployer tous les timeframes

```bash
./deploy/deploy_mfi_stoch_extended.sh
```

Ou déployer individuellement:

```bash
# 1m
nomad job run deploy/mfi-stoch-extended-1m.nomad

# 5m
nomad job run deploy/mfi-stoch-extended-5m.nomad

# 15m
nomad job run deploy/mfi-stoch-extended-15m.nomad
```

## 📱 Topics de Notification

Chaque timeframe a son **propre topic** ntfy:

| Timeframe | Topic | URL |
|-----------|-------|-----|
| 1m | `mfi-stoch-1m` | https://notifications.koyad.com/mfi-stoch-1m |
| 5m | `mfi-stoch-5m` | https://notifications.koyad.com/mfi-stoch-5m |
| 15m | `mfi-stoch-15m` | https://notifications.koyad.com/mfi-stoch-15m |

**S'abonner aux notifications:**
```bash
# Via mobile app
# Ajouter les topics: mfi-stoch-1m, mfi-stoch-5m, mfi-stoch-15m

# Via CLI
ntfy subscribe notifications.koyad.com/mfi-stoch-15m
```

## 📊 Configuration

### Paramètres Indicateurs
```json
{
  "mfi_period": 16,
  "mfi_extreme_high": 70.0,
  "mfi_extreme_low": 30.0,
  "stoch_k_period": 16,
  "stoch_smooth_k": 3,
  "stoch_d_period": 3,
  "stoch_extreme_high": 80.0,
  "stoch_extreme_low": 20.0
}
```

### Paramètres MM Extended
```json
{
  "mfi_extended_high": 60.0,
  "mfi_extended_low": 40.0,
  "stoch_extended_high": 70.0,
  "stoch_extended_low": 30.0,
  "max_reentries": 3,
  "use_extended_mm": true
}
```

### Trading
```json
{
  "leverage": 5.0,
  "trailing_stop_pct": 0.25,
  "take_profit_target_pct": 0.0025
}
```

## 🔍 Monitoring

### Vérifier le statut
```bash
nomad job status mfi-stoch-extended-1m
nomad job status mfi-stoch-extended-5m
nomad job status mfi-stoch-extended-15m
```

### Voir les logs
```bash
# Obtenir l'allocation ID
nomad job status mfi-stoch-extended-15m

# Suivre les logs
nomad alloc logs -f -task mfi-stoch-15m <alloc-id>
```

### Vérifier la base de données
Les signaux sont stockés dans ArangoDB:
- Collection: `mfi_stoch_1m_signals`
- Collection: `mfi_stoch_5m_signals`
- Collection: `mfi_stoch_15m_signals`

## 🛑 Arrêter les Déploiements

```bash
nomad job stop mfi-stoch-extended-1m
nomad job stop mfi-stoch-extended-5m
nomad job stop mfi-stoch-extended-15m
```

## 📝 State Files

Chaque timeframe maintient son propre état:
- `/root/data/backtest-optimizer/state/mfi_stoch_1m_extended.json`
- `/root/data/backtest-optimizer/state/mfi_stoch_5m_extended.json`
- `/root/data/backtest-optimizer/state/mfi_stoch_15m_extended.json`

## ⚠️ Notes Importantes

1. **1m est très actif**: ~143 trades/jour, haute fréquence de notifications
2. **15m recommandé**: Meilleur ratio performance/activité (+6049%, ~13 trades/jour)
3. **5m non testé**: Performances inconnues, déployer avec prudence
4. **MM Extended**: Re-entry automatique si conditions extrêmes maintenues (max 3 fois)

## 🆚 Comparaison avec CCI+MFI+Stoch

| Metric | CCI+MFI+Stoch 15m | MFI+Stoch 15m | Avantage |
|--------|-------------------|---------------|----------|
| Return | +3615% | **+6049%** | **MFI+Stoch +67%** 🏆 |
| Max DD | 4.01% | 6.27% | CCI+MFI+Stoch -2.26% |
| Win Rate | 85.5% | 85.4% | Égalité |
| Trades | 2,883 | 4,764 | MFI+Stoch +65% |

**Verdict**: MFI+Stoch Extended 15m est **la meilleure stratégie testée** !

## 📞 Support

- Logs: `nomad alloc logs`
- Database: ArangoDB Web UI (http://10.0.0.1:8529)
- Notifications: ntfy topics (voir ci-dessus)
