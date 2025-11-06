# 🎯 Scalping Paper/Live Trading

Application de trading en temps réel pour la stratégie scalping (triple extrême: CCI + MFI + Stochastique).

## 📋 Modes d'Exécution

### **Paper Trading (Testnet)**
Trading réel sur Binance Testnet avec argent virtuel.

```bash
cd cmd/scalping_paper
go run . -mode paper -config ../../config/config.yaml
```

### **Live Trading (Production)**
Trading réel sur Binance Production avec argent RÉEL.

```bash
cd cmd/scalping_paper
go run . -mode live -config ../../config/config.yaml
```

⚠️ **ATTENTION** : Le mode `live` nécessite une confirmation `CONFIRM` avant de démarrer.

---

## 🔧 Fonctionnement

### **Cycle d'Exécution**

1. **Chargement initial** : Récupère les 300 dernières klines via REST API
2. **Loop 10 secondes** : Polling toutes les 10 secondes
3. **Détection bougies fermées** : Compare avec l'historique local
4. **Calcul indicateurs** : CCI, MFI, Stochastique (TV Standard)
5. **Détection signaux** : Triple extrême + croisement + validation

### **Endpoints Binance**

- **Paper** : `https://testnet.binance.vision/api`
- **Live** : `https://api.binance.com/api`

---

## 📊 Logique de Détection

Identique à `scalping_engine` (backtest) :

1. **Triple extrême** (CCI + MFI + Stoch) sur N-1
2. **Croisement Stochastique** (N-2 vs N-1)
3. **Fenêtre validation** (6 bougies par défaut)
4. **Type bougie** inverse au signal
5. **Volume** > 25% moyenne bougies inverses

---

## 🛠️ Configuration

**Fichier** : `config/config.yaml`

```yaml
strategy:
  name: "SCALPING"
  scalping:
    timeframe: "5m"
    
    # Seuils extrêmes
    cci_surachat: 100.0
    cci_survente: -100.0
    mfi_surachat: 80.0
    mfi_survente: 20.0
    stoch_surachat: 80.0
    stoch_survente: 20.0
    
    # Validation
    validation_window: 6
    volume_threshold: 0.25
    volume_period: 5
    volume_max_ext: 100

binance_data:
  symbols: ["SOLUSDT"]
```

---

## 🚀 Compilation

```bash
# Paper trading
go build -o scalping_paper .

# Exécution
./scalping_paper -mode paper

# Ou directement
go run . -mode paper
```

---

## 📝 Arguments CLI

| Argument | Valeur par défaut | Description |
|----------|-------------------|-------------|
| `-config` | `config/config.yaml` | Chemin fichier configuration |
| `-mode` | `paper` | Mode: `paper` ou `live` |
| `-symbol` | (de config) | Override symbole (ex: `SOLUSDT`) |

### **Exemples**

```bash
# Paper avec symbole custom
go run . -mode paper -symbol ETHUSDT

# Live trading
go run . -mode live

# Config custom
go run . -mode paper -config /path/to/config.yaml
```

---

## 🔍 Logs

```
🎯 SCALPING PAPER/LIVE - Trading Temps Réel
============================================

📋 Chargement configuration: config/config.yaml
✅ Configuration chargée

📊 Paramètres Trading:
   - Mode: paper
   - Stratégie: SCALPING
   - Symbole: SOLUSDT
   - Timeframe: 5m
   - Endpoint: https://testnet.binance.vision

📂 Chargement historique initial...
✅ 300 klines initiales chargées

🔄 Démarrage loop trading (10 secondes)...
⏱️  Loop 10 secondes démarrée

[14:35:10] 🔄 Tick...
[14:35:20] 🔄 Tick...
   📊 1 nouvelle(s) bougie(s) fermée(s)
   🔔 Marqueur détecté: 2024-11-05 14:35
   🎯 1 signal(aux) détecté(s)!
      → LONG à 185.43 (CCI=-105.2, MFI=18.3, K=15.7)
```

---

## ⚠️ Limitations Actuelles

1. **Pas de gestion position** : Signaux détectés uniquement (pas d'ordres passés)
2. **Pas de trailing stop** : À implémenter
3. **Pas de money management** : À implémenter
4. **Detection simplifiée** : `DetectSignals()` retourne vide (TODO)

---

## 🔜 Prochaines Étapes

1. ✅ Compléter `DetectSignals()` (copier logique de `scalping_engine`)
2. ⏳ Passer ordres via REST API Binance
3. ⏳ Gérer positions ouvertes
4. ⏳ Implémenter trailing stop
5. ⏳ Ajouter money management
6. ⏳ Export JSON des signaux

---

## 🔗 Voir Aussi

- **Scalping Engine (Backtest)** : `cmd/scalping_engine/`
- **Logique Détection** : `cmd/scalping_engine/LOGIQUE_DETECTION_SIGNAUX.md`
- **Configuration** : `config/config.yaml`
