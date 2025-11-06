# Architecture Scalping Paper/Live

## 🎯 Objectif

Module de trading en temps réel pour modes **paper** (testnet) et **live** (production).

---

## ⏱️ Cycle Principal : LOOP 10 SECONDES SYNCHRONISÉE

**Référence :** `docs/workflow/04_engine_temporal.md` ligne 92

```yaml
InitPaperLiveMode:
  Inputs:
    - loop_interval: 10 secondes  ← SPÉCIFICATION OFFICIELLE
```

### **Synchronisation Critique**

Les ticks doivent tomber **exactement sur :00, :10, :20, :30, :40, :50 secondes** pour que :
- Le tick **:00** coïncide avec les clôtures de bougies (ex: 19:40:00)
- Pas de décalage cumulatif
- Précision maximale

**Au démarrage**, le programme calcule le délai jusqu'au prochain multiple de 10s, puis démarre un ticker synchronisé.

### **Pourquoi 10 secondes ?**

Le temporal engine a **DEUX responsabilités distinctes** :

1. **Mise à jour trailing stops** (tick-by-tick, toutes les 10s)
   - Si position ouverte
   - Ajustements dynamiques basés sur indicateurs
   - Vérification stop hit
   - **Nécessite cycle rapide (10s) pour réactivité**

2. **Détection signaux** (sur clôture de bougie)
   - Calcul indicateurs (CCI, MFI, Stochastic)
   - Triple extreme detection
   - Validation window
   - **Nécessite bougies fermées (timeframe: 5m, 15m, etc.)**

---

## 📊 Architecture du Tick

```
SYNCHRONISATION INITIALE:
┌────────────────────────────────────────┐
│ Démarrage à HH:MM:SS                   │
│ Calcul: (10 - SS%10) secondes         │
│ Attente...                             │
│ Premier tick à HH:MM:X0  ← Multiple 10│
└────────────────────────────────────────┘

PUIS TOUTES LES 10 SECONDES (X0, X0, X0...):
┌────────────────────────────────────────┐
│ [HH:MM:X0] 🔄 Tick...                 │
├────────────────────────────────────────┤
│ 1️⃣ Fetch dernières klines (API)      │
│                                        │
│ 2️⃣ Position ouverte ?                │
│    ├─ OUI → Update trailing stop      │
│    │         Check stop hit            │
│    └─ NON → Rien                       │
│                                        │
│ 3️⃣ Bougie fermée détectée ?          │
│    ├─ OUI → Calcul indicateurs        │
│    │         Détection signaux         │
│    │         Update zones              │
│    └─ NON → Rien                       │
└────────────────────────────────────────┘
      ⬇ 10 secondes
┌────────────────────────────────────────┐
│ [HH:MM:Y0] 🔄 Tick...                 │
│ (Y0 = X0 + 10, ex: :20, :30, :40...)  │
└────────────────────────────────────────┘
```

---

## 🔄 Synchronisation au Démarrage

### **Exemple : Démarrage à 19:35:07**

```
⏱️  Synchronisation sur multiples de 10s...
   Heure actuelle: 19:35:07
   Prochain tick: 19:35:10 (dans 3s)
   Timeframe bougie: 5m

[Attente 3 secondes...]

[19:35:10] 🔔 Synchronisé!
⏱️  Loop active (tick toutes les 10s)
```

**Algorithme :**
```go
currentSecond = 7
secondsUntilNext = 10 - (7 % 10) = 3
// Attendre 3s → Premier tick à :10
```

---

## 🔄 Exemple Complet (Timeframe 5m)

### **Scénario : Démarrage avant clôture bougie**

```
[Démarrage à 19:34:53]
⏱️  Synchronisation sur multiples de 10s...
   Attente: 7s

19:35:00 [Tick] - Bougie 19:30-19:35 fermée ← CLÔTURE
   → Calcul indicateurs
   → Signal LONG détecté
   → Position ouverte

19:35:10 [Tick] - Bougie 19:35-19:40 en construction
   → Update trailing stop (prix actuel)
   → Pas de signal (bougie non fermée)

19:35:20 [Tick] - Bougie 19:35-19:40 en construction
   → Update trailing stop (prix actuel)
   → Pas de signal (bougie non fermée)

19:35:30 [Tick] - Bougie 19:35-19:40 en construction
   → Update trailing stop (prix actuel)
   → Pas de signal (bougie non fermée)

19:35:40 [Tick] - Bougie 19:35-19:40 en construction
   → Update trailing stop (prix actuel)
   → Stop hit → Position fermée !

19:35:50 [Tick] - Bougie 19:35-19:40 en construction
   → Pas de position
   → Pas de signal (bougie non fermée)

19:36:00 [Tick] - Bougie 19:35-19:40 en construction
   → Pas de position
   → Pas de signal (bougie non fermée)

... (30 ticks pour 1 bougie de 5 minutes)

19:40:00 [Tick] - Bougie 19:35-19:40 fermée
   → Calcul indicateurs
   → Détection signaux
   → Signal SHORT détecté
   → Position ouverte
```

---

## 🆚 Différences avec Backtest

| Aspect | Backtest | Paper/Live |
|--------|----------|------------|
| **Cycle** | Trade par trade (ms) | Loop 10 secondes |
| **Données** | Fichiers historiques | API REST temps réel |
| **Timestamp** | trade.timestamp | time.Now() |
| **Trailing Stop** | Update à chaque trade | Update toutes les 10s |
| **Indicateurs** | Sur marqueur bougie | Sur bougie fermée |
| **Granularité** | Ultra-fine | Macro (10s) |

---

## 📋 Implémentation Actuelle

### **Fichier : `app_paper.go`**

```go
func (app *ScalpingPaperApp) runTimerLoop(ctx context.Context) error {
    loopInterval := 10 * time.Second  // ← FIXE 10 SECONDES
    ticker := time.NewTicker(loopInterval)
    
    for {
        case <-ticker.C:
            app.processTimerTick()  // Toutes les 10s
    }
}

func (app *ScalpingPaperApp) processTimerTick() error {
    // 1. Fetch klines
    newKlines := app.fetchLatestKlines()
    
    // 2. Update trailing stop (si position ouverte)
    if app.hasOpenPosition() {
        app.updateTrailingStop(newKlines)
        app.checkStopHit()
    }
    
    // 3. Détecter bougies fermées
    completedCandles := app.detectNewCompletedCandles(newKlines)
    
    // 4. Pour chaque bougie fermée → calcul indicateurs
    for _, timestamp := range completedCandles {
        app.processMarker(timestamp)  // Indicateurs + signaux
    }
}
```

---

## ⚠️ Points Critiques

### **NE PAS CONFONDRE**

- ❌ **Intervalle loop (10s)** ≠ Timeframe bougie (5m)
- ❌ **Update stop (10s)** ≠ Calcul indicateurs (clôture)
- ❌ **Tick du timer** ≠ Trade Binance

### **TOUJOURS RESPECTER**

- ✅ Loop = 10 secondes (fixe, non configurable)
- ✅ Indicateurs = bougies fermées uniquement
- ✅ Trailing stop = à chaque tick si position ouverte

---

## 🔧 Configuration

```yaml
# config/config.yaml
strategy:
  scalping:
    timeframe: "5m"  # ← Fréquence CALCUL INDICATEURS
    # Loop 10s est HARDCODÉ dans le code (non configurable)
```

**IMPORTANT :** Le `timeframe` configure uniquement la fréquence de calcul des indicateurs, PAS la fréquence de la loop.

---

## 📚 Références

- `docs/workflow/04_engine_temporal.md` - Spécification loop 10s
- `docs/user_stories/06_engine_temporal_backtest.md` - Différences modes
- `cmd/scalping_engine/app.go` - Implémentation backtest (trade-by-trade)
- `cmd/scalping_paper/app_paper.go` - Implémentation paper/live (loop 10s)

---

## 🎯 Résumé

```
┌─────────────────────────────────────┐
│ SCALPING PAPER/LIVE                 │
├─────────────────────────────────────┤
│ Loop : 10 SECONDES (fixe)           │
│                                     │
│ Chaque tick (10s) :                 │
│ ✅ Fetch klines                     │
│ ✅ Update trailing stop             │
│ ✅ Check bougie fermée              │
│    └─ Si oui → Indicateurs + signaux│
└─────────────────────────────────────┘
```

**DEUX fonctions en UNE loop de 10 secondes !**
