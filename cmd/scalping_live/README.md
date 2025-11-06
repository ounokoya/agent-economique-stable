# 🚀 Scalping Live - Trading Production

Module de trading en **TEMPS RÉEL** pour le trading **LIVE** (argent réel).

---

## ⚠️ ATTENTION - MODE PRODUCTION

Ce module lance le trading avec de **l'argent réel**.  
**Confirmation obligatoire** : Tu dois taper `CONFIRM` au démarrage.

---

## 🎯 Différences avec scalping_paper

| Aspect | scalping_paper | scalping_live |
|--------|----------------|---------------|
| **Mode par défaut** | paper (testnet) | live (production) |
| **Argument `-mode`** | Requis (`-mode paper/live`) | Forcé à `live` |
| **Confirmation** | Non requise | **Obligatoire** (taper "CONFIRM") |
| **Endpoint** | Testnet Binance | **Production Binance** |
| **Notifications** | Topic `scalping-paper` | Topic `scalping-live` |
| **API Keys** | Pas nécessaires (lecture publique) | **OBLIGATOIRES** |

---

## 🚀 Utilisation

### **Lancement Standard**

```bash
cd cmd/scalping_live
go run .
```

ou depuis la racine :

```bash
go run ./cmd/scalping_live
```

### **Avec Configuration Custom**

```bash
go run ./cmd/scalping_live -config custom_config.yaml
```

### **Override Symbole**

```bash
go run ./cmd/scalping_live -symbol ETHUSDT
```

---

## 📊 Processus de Lancement

```
1. Affichage warning MODE LIVE
2. Demande de confirmation : Taper "CONFIRM"
3. Chargement configuration
4. Affichage paramètres
5. Connexion API Binance Production
6. Envoi notification démarrage (ntfy)
7. Chargement 300 dernières klines
8. Démarrage loop 10 secondes
9. Trading actif
```

---

## 🔐 Sécurité

### **API Keys Binance**

Les clés API doivent être configurées dans le code avec les permissions :
- ✅ **Lecture** données marché
- ✅ **Trading** (ordres spot)
- ❌ **Retrait** (INTERDIT pour sécurité)

### **Confirmation Obligatoire**

```
⚠️  MODE LIVE ACTIVÉ - TRADING RÉEL ⚠️

🔴 ATTENTION : Vous êtes sur le point de lancer le trading LIVE (argent réel)
Tapez 'CONFIRM' pour continuer: █
```

Si tu tapes autre chose que `CONFIRM`, le programme s'arrête immédiatement.

---

## 📱 Notifications

Les notifications sont envoyées sur le topic **`scalping-live`** :

- 🚀 Démarrage
- 🎯 Signaux (LONG/SHORT)
- ⚠️ Erreurs
- 🛑 Arrêt

**S'abonner :**
```
App ntfy → Ajouter topic → scalping-live
Serveur: https://notifications.koyad.com
```

---

## 🛠️ Configuration

**Fichier** : `config/config.yaml` (par défaut)

```yaml
binance_data:
  symbols:
    - "SOLUSDT"

strategy:
  name: "SCALPING"
  scalping:
    timeframe: "5m"
    cci_surachat: 100.0
    cci_survente: -100.0
    mfi_surachat: 60.0
    mfi_survente: 40.0
    stoch_surachat: 70.0
    stoch_survente: 30.0
    validation_window: 3
```

---

## 🔄 Loop Temporelle

**Fréquence** : Tick toutes les **10 secondes** (synchronisé sur :00, :10, :20, :30, :40, :50)

### **Actions par tick :**
1. Fetch 10 dernières klines
2. **Si position ouverte** : Update trailing stop
3. **Si bougie fermée** : Calcul indicateurs + détection signaux

---

## 🎯 Détection Signaux

### **Conditions LONG :**
- CCI < -100 (survente)
- MFI < 40 (survente)
- Stoch < 30 (survente)
- Croisement Stoch: K passe AU-DESSUS de D
- Validation : Bougie verte dans les 3 suivantes

### **Conditions SHORT :**
- CCI > 100 (surachat)
- MFI > 60 (surachat)
- Stoch > 70 (surachat)
- Croisement Stoch: K passe SOUS D
- Validation : Bougie rouge dans les 3 suivantes

---

## 🛑 Arrêt

**Graceful shutdown** :
```bash
Ctrl+C
```

Le programme :
1. Arrête le ticker
2. Ferme les positions ouvertes (TODO)
3. Envoie notification arrêt
4. Sort proprement

---

## 📝 Logs

### **Exemple de logs normaux :**

```
🎯 SCALPING LIVE - Trading Production
========================================

⚠️  MODE LIVE ACTIVÉ - TRADING RÉEL ⚠️

🔴 ATTENTION : Vous êtes sur le point de lancer le trading LIVE (argent réel)
Tapez 'CONFIRM' pour continuer: CONFIRM

📋 Chargement configuration: config/config.yaml

📊 Paramètres Trading:
   - Mode: live
   - Stratégie: SCALPING
   - Symbole: SOLUSDT
   - Timeframe: 5m
   - Endpoint: PRODUCTION BINANCE

🚀 Démarrage LIVE trading...

📂 Chargement historique initial...
✅ 95 klines initiales chargées

🔄 Démarrage loop trading...
⏱️  Synchronisation sur multiples de 10s...

[21:55:00] 🔔 Synchronisé!
[21:55:00] 🔄 Tick...
⏱️  Loop active (tick toutes les 10s)

[21:55:10] 🔄 Tick...
[21:55:20] 🔄 Tick...
```

### **Exemple détection signal :**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🕐 21:55:00 | MARQUEUR 5M DÉTECTÉ
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Indicateurs calculés: CCI=96, MFI=96, StochK=96, StochD=96

📊 INDICATEURS CALCULÉS:
   CCI(N-1): -105.2 | MFI(N-1): 18.3
   Stoch K(N-1): 15.7 D(N-1): 22.1

🔍 DÉTECTION SIGNAUX:
[DEBUG] Triple extrême: N-2=true, N-1=true
[DEBUG] 🎯 Triple extrême DÉTECTÉ!
[DEBUG] Croisement stochastique: type=LONG
[DEBUG] ✅ CROISEMENT DÉTECTÉ: LONG
[DEBUG] ✅ SIGNAL VALIDÉ dans window!

   🎯 1 signal(aux) détecté(s)!
      → LONG à 185.43 (CCI=-105.2, MFI=18.3, K=15.7)
      ✅ Notification envoyée
```

---

## ⚙️ Arguments CLI

| Argument | Valeur par défaut | Description |
|----------|-------------------|-------------|
| `-config` | `config/config.yaml` | Chemin fichier configuration |
| `-symbol` | (de config) | Override symbole (ex: `SOLUSDT`) |

**Note** : Pas d'argument `-mode`, il est forcé à `live`.

---

## 📚 Voir Aussi

- **Scalping Paper (Testnet)** : `cmd/scalping_paper/`
- **Scalping Engine (Backtest)** : `cmd/scalping_engine/`
- **Configuration** : `config/config.yaml`
- **Architecture** : `cmd/scalping_live/ARCHITECTURE.md`

---

## ⚠️ TODO - À Implémenter

- [ ] Position Management (ouverture/fermeture)
- [ ] Trailing Stop dynamique
- [ ] Money Management
- [ ] API Keys configuration
- [ ] Risk Management
- [ ] Métriques performance temps réel
