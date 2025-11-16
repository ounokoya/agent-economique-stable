# 📋 SPÉCIFICATION STRATÉGIE VWMA CROSS + DMI SIMPLE

**Date**: 2025-11-08  
**Version**: 1.0  
**Auteur**: Agent Économique Stable

---

## 🎯 PRINCIPE GÉNÉRAL

La stratégie **VWMA Cross + DMI Simple** combine :
1. **Croisement VWMA court/long** : Détection de changement de direction
2. **Position relative DI** : Force directionnelle (DI+ vs DI-)
3. **Croisement DX/ADX** : Évolution de la force (augmente ou diminue)

**RÈGLE ABSOLUE** : 
- Croisement VWMA dicte TOUJOURS la direction du signal (LONG ou SHORT)
- DI position et DX/ADX croisement ajoutent qualification (Tendance ou Contre-Tendance)
- **Pas de signals ENTRY/EXIT distincts** : seulement des signaux LONG et SHORT
- **Le signal est contextuel** : ENTRY si aucune position, EXIT si position inverse ouverte

---

## 📊 LES 4 COMBINAISONS DE BASE

| # | VWMA Cross | DI Position | DX/ADX Cross | Signal Généré |
|---|------------|-------------|--------------|---------------|
| **1** | Court > Long ↗ | DI+ > DI- | DX croise ADX ↑ | **LONG** |
| **2** | Court > Long ↗ | DI- > DI+ | DX croise ADX ↓ | **LONG** |
| **3** | Court < Long ↘ | DI- > DI+ | DX croise ADX ↑ | **SHORT** |
| **4** | Court < Long ↘ | DI+ > DI- | DX croise ADX ↓ | **SHORT** |

### Interprétation

**Tendance** (DX croise ADX ↑) :
- VWMA cross et DI position vont dans le même sens
- La force dans cette direction augmente
- Signal fort, confirmé

**Contre-Tendance** (DX croise ADX ↓) :
- VWMA cross commence une direction, mais DI encore dans l'ancienne
- La force de l'ancienne direction diminue
- Signal d'anticipation de retournement

### Comportement contextuel des signaux

**Logique automatique** :
- **Si aucune position ouverte** → Signal = **ENTRY**
- **Si position LONG ouverte** → Signal SHORT = **EXIT LONG**
- **Si position SHORT ouverte** → Signal LONG = **EXIT SHORT**

**Exemple** :
- Position = NONE → Signal généré = **LONG** → Action = **ENTRY LONG**
- Position = LONG → Signal généré = **SHORT** → Action = **EXIT LONG**
- Position = SHORT → Signal généré = **LONG** → Action = **EXIT SHORT**

---

## 🟢 SIGNAUX LONG

### Signal 1 : LONG (Tendance)
**Conditions** :
- ✅ Croisement VWMA court > long (détection du croisement)
- ✅ DI+ > DI- (position relative, force haussière domine)
- ✅ DX croise ADX vers le haut (force haussière augmente)

**Comportement** :
- Si aucune position → **ENTRY LONG**
- Si position SHORT → **EXIT SHORT**

### Signal 2 : LONG (Contre-Tendance)
**Conditions** :
- ✅ Croisement VWMA court > long (détection du croisement)
- ✅ DI- > DI+ (position relative, force baissière domine ENCORE)
- ✅ DX croise ADX vers le bas (force baissière diminue)

**Comportement** :
- Si aucune position → **ENTRY LONG**
- Si position SHORT → **EXIT SHORT**

---

## 🔴 SIGNAUX SHORT

### Signal 3 : SHORT (Tendance)
**Conditions** :
- ✅ Croisement VWMA court < long (détection du croisement)
- ✅ DI- > DI+ (position relative, force baissière domine)
- ✅ DX croise ADX vers le haut (force baissière augmente)

**Comportement** :
- Si aucune position → **ENTRY SHORT**
- Si position LONG → **EXIT LONG**

### Signal 4 : SHORT (Contre-Tendance)
**Conditions** :
- ✅ Croisement VWMA court < long (détection du croisement)
- ✅ DI+ > DI- (position relative, force haussière domine ENCORE)
- ✅ DX croise ADX vers le bas (force haussière diminue)

**Comportement** :
- Si aucune position → **ENTRY SHORT**
- Si position LONG → **EXIT LONG**


---

## 🔄 FENÊTRE DE MATCHING

### Règle de validation obligatoire

**Les 3 conditions doivent être validées dans une fenêtre de W bougies consécutives, peu importe l'ordre** :

1. ✅ **VWMA Cross** : Croisement court/long détecté
2. ✅ **DI Position** : Position relative DI+ vs DI- (simple comparaison)
3. ✅ **DX/ADX Cross** : Croisement DX vs ADX détecté

### Paramètre de fenêtre

```yaml
window_matching: 5  # Nombre de bougies pour matcher les 3 conditions (typ. 3-5)
```

### Logique de détection

```
POUR chaque bougie i:
  
  # Évaluer fenêtre W précédente [i-W+1, i]
  fenetre = [i-W+1, i-W+2, ..., i]
  
  # Chercher conditions dans la fenêtre
  vwma_cross_ok = false
  di_position_ok = false  
  dx_adx_cross_ok = false
  
  POUR j dans fenetre:
    SI VWMA cross détecté à j: vwma_cross_ok = true
    SI DI position valide à j: di_position_ok = true
    SI DX/ADX cross détecté à j: dx_adx_cross_ok = true
  
  # Générer signal si les 3 conditions réunies
  SI vwma_cross_ok ET di_position_ok ET dx_adx_cross_ok:
    signal = ClassifierSignal(vwma_direction, di_dominant, dx_cross_direction)
    GÉNÉRER Signal(timestamp=i, price=Close[i])
```

### Exemples

#### Signal LONG Tendance dans fenêtre W=5

**Bougie T0** : DX croise ADX ↑ ✅  
**Bougie T1** : Rien  
**Bougie T2** : VWMA court croise long vers haut ✅  
**Bougie T3** : Rien  
**Bougie T4** : DI+ > DI- (position) ✅  

**Résultat** : Signal LONG Tendance généré à **T4** (timestamp=T4, prix=Close[T4])

#### Conditions simultanées

**Bougie T0** : VWMA cross + DI position + DX/ADX cross tous validés ✅✅✅  
**Résultat** : Signal généré **immédiatement** à T0 (fenêtre = 1 bougie)

### Timestamp et prix du signal

- **Timestamp** : Toujours la dernière bougie de la fenêtre (quand 3ème condition validée)
- **Prix d'exécution** : Close de cette dernière bougie
- **Fenêtre référence** : Sauvegardée dans métadonnées pour debug

---

## 📐 PARAMÈTRES TECHNIQUES

### Paramètres VWMA
```yaml
vwma_short_period: 10        # Période VWMA court
vwma_long_period: 20         # Période VWMA long
```

### Paramètres DMI
```yaml
dmi_period: 14               # Période DMI standard
dmi_smooth: 14               # Lissage DMI
```

### Paramètres fenêtre
```yaml
window_matching: 5           # Fenêtre matching 3 conditions
```

---

## 🔍 DÉTECTION ET VALIDATION

### Étapes de détection

1. **Calculer indicateurs** :
   - VWMA court, VWMA long
   - DMI (DI+, DI-)
   - DX, ADX

2. **Détecter croisements/états sur chaque bougie** :
   - Croisement VWMA court vs long
   - Position relative DI+ vs DI-
   - Croisement DX vs ADX

3. **Appliquer fenêtre de matching** :
   - Vérifier si les 3 conditions sont présentes dans fenêtre W
   - Classifier le signal combiné

4. **Générer signal final** :
   - Toujours générer si conditions validées (pas de filtres)

---

## 🎨 MÉTADONNÉES DES SIGNAUX

```go
type VWMACrossDMISimpleSignal struct {
    // Identification
    Type        string  // "LONG" ou "SHORT"
    Mode        string  // "TREND" ou "COUNTER_TREND"
    
    // VWMA Cross
    VWMAShort      float64
    VWMALong       float64
    VWMACrossDetected bool
    VWMADirection  string  // "UP" ou "DOWN"
    
    // DI Position
    DIPlus         float64
    DIMinus        float64
    DIDominant     string  // "DI_PLUS" ou "DI_MINUS"
    
    // DX/ADX Cross
    DX             float64
    ADX            float64
    DXCrossADX     bool
    DXCrossDirection string  // "UP" ou "DOWN"
    
    // Contexte
    Confidence     float64
    Timestamp      time.Time
    Price          float64
    
    // Comportement contextuel
    CurrentPosition string // "NONE", "LONG", "SHORT"
    Action         string  // "ENTRY" ou "EXIT" (déterminé par contexte)
    
    // Méta
    WindowSize     int     // Taille fenêtre utilisée
    WindowDuration int     // Durée réelle matching (bougies)
}
```

---

## ⚠️ RÈGLES IMPORTANTES

### 1. VWMA Cross = Direction absolue
- VWMA court > long → LONG uniquement
- VWMA court < long → SHORT uniquement
- **JAMAIS** de LONG si VWMA court < long
- **JAMAIS** de SHORT si VWMA court > long

### 2. Logique contextuelle unique
- **Signal LONG** : ENTRY si aucune position, EXIT si position SHORT
- **Signal SHORT** : ENTRY si aucune position, EXIT si position LONG
- **Un seul signal actif** à la fois selon contexte de position

### 3. Priorité automatique
- **EXIT > ENTRY** implicite par contexte (un signal inverse ferme toujours)
- Si plusieurs signaux valides même direction : prendre priorité 1 (Tendance) puis 2 (Contre-Tendance)

### 4. Gestion positions
- **Une seule position à la fois** (par défaut)
- Fermer position existante avant ouvrir nouvelle

---

## 📊 EXEMPLE COMPLET

### Scénario : Retournement haussier

**T0 : Baisse en cours**
- VWMA court < long
- DI- > DI+ (position)
- DX croise ADX ↑ (force baisse augmente)
- → **Signal SHORT** → Si aucune position = **ENTRY SHORT**

**T1 : VWMA se retourne**
- VWMA **court > long** (croisement)
- DI- > DI+ **encore** (position)
- DX croise ADX ↓ (force baisse diminue)
- → **Signal LONG** → Si position SHORT = **EXIT SHORT**

**T2 : DMI confirme**
- VWMA court > long
- DI+ > DI- (position bascule)
- DX croise ADX ↑ (force hausse augmente)
- → **Signal LONG** → Si aucune position = **ENTRY LONG**

---

## 🚀 IMPLÉMENTATION

### Fichiers à créer

1. **`internal/signals/vwma_cross_dmi_simple/generator.go`**
   - Structure `VWMACrossDMISimpleGenerator`
   - Méthodes `Initialize()`, `CalculateIndicators()`, `DetectSignals()`
   - Logique de matching VWMA cross + DI position + DX/ADX cross

2. **`cmd/vwma_cross_dmi_simple_demo/main.go`**
   - Demo standalone
   - Test de la stratégie
   - Export résultats JSON

3. **`cmd/vwma_cross_dmi_simple_engine/`**
   - Moteur temporal backtest
   - Comme `direction_engine` mais avec VWMA cross + DMI simple

### Tests à effectuer

1. Vérifier que les 4 signaux sont correctement détectés
2. Tester avec différentes tailles de fenêtre (3, 5, 7 bougies)
3. Valider que Entry/Exit sont indépendants
4. Comparer performances vs Direction simple et Direction+DMI
5. Optimiser paramètres VWMA (short/long periods)

---

## 📚 RÉFÉRENCES

- **Direction Generator** : `internal/signals/direction/generator.go`
- **Direction+DMI Generator** : `internal/signals/direction_dmi/generator.go`
- **Trend Generator** : `internal/signals/trend/generator.go` (logique DMI/DX/ADX)

---

**Version finale validée le 2025-11-08**
**Cette spécification fait référence - Ne pas modifier sans discussion**
