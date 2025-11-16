# 📋 SPÉCIFICATION STRATÉGIE DIRECTION + DMI

**Date**: 2025-11-08  
**Version**: 1.0  
**Auteur**: Agent Économique Stable

---

## 🎯 PRINCIPE GÉNÉRAL

La stratégie **Direction+DMI** combine :
1. **VWMA** : Détection de la direction du marché (pente)
2. **DMI** (DI+, DI-) : Force directionnelle
3. **DX/ADX** : Évolution de la force (augmente ou diminue)

**RÈGLE ABSOLUE** : 
- VWMA dicte TOUJOURS la direction de la position (LONG ou SHORT)
- DMI/DX/ADX ajoutent qualification (Tendance ou Contre-Tendance)
- Entry et Exit sont INDÉPENDANTS (n'importe quelle entrée peut avoir n'importe quelle sortie)

---

## 📊 LES 4 COMBINAISONS DE BASE

| # | VWMA | DI Dominant | DX vs ADX | Nom Signal |
|---|------|-------------|-----------|------------|
| **1** | Croissante ↗ | DI+ > DI- | DX > ADX ↑ | **LONG Tendance** |
| **2** | Croissante ↗ | DI- > DI+ | DX < ADX ↓ | **LONG Contre-Tendance** |
| **3** | Décroissante ↘ | DI- > DI+ | DX > ADX ↑ | **SHORT Tendance** |
| **4** | Décroissante ↘ | DI+ > DI- | DX < ADX ↓ | **SHORT Contre-Tendance** |

### Interprétation

**Tendance** (DX > ADX ↑) :
- VWMA et DMI vont dans le même sens
- La force dans cette direction **augmente**
- Signal fort, confirmé

**Contre-Tendance** (DX < ADX ↓) :
- VWMA commence une direction, mais DMI encore dans l'ancienne
- La force de l'ancienne direction **diminue**
- Signal d'anticipation de retournement

---

## 🟢 ENTRÉES POSITION LONG

### Signal 1 : LONG Tendance
**Conditions** :
- ✅ VWMA pente croissante (confirmée K bougies)
- ✅ DI+ > DI- (force haussière domine)
- ✅ DX > ADX ↑ (force haussière augmente)
- ✅ Gap DI suffisant (DI+ - DI- ≥ gammaGapDI)
- ✅ Gap DX suffisant (DX - ADX ≥ gammaGapDX)

**Flag** : `enable_entry_trend = true`

### Signal 2 : LONG Contre-Tendance
**Conditions** :
- ✅ VWMA pente croissante (confirmée K bougies)
- ✅ DI- > DI+ (force baissière domine ENCORE)
- ✅ DX < ADX ↓ (force baissière diminue)
- ✅ Gap DI suffisant (DI- - DI+ ≥ gammaGapDI)
- ✅ Gap DX suffisant (ADX - DX ≥ gammaGapDX)

**Flag** : `enable_entry_counter_trend = true`

---

## 🔴 ENTRÉES POSITION SHORT

### Signal 3 : SHORT Tendance
**Conditions** :
- ✅ VWMA pente décroissante (confirmée K bougies)
- ✅ DI- > DI+ (force baissière domine)
- ✅ DX > ADX ↑ (force baissière augmente)
- ✅ Gap DI suffisant (DI- - DI+ ≥ gammaGapDI)
- ✅ Gap DX suffisant (DX - ADX ≥ gammaGapDX)

**Flag** : `enable_entry_trend = true`

### Signal 4 : SHORT Contre-Tendance
**Conditions** :
- ✅ VWMA pente décroissante (confirmée K bougies)
- ✅ DI+ > DI- (force haussière domine ENCORE)
- ✅ DX < ADX ↓ (force haussière diminue)
- ✅ Gap DI suffisant (DI+ - DI- ≥ gammaGapDI)
- ✅ Gap DX suffisant (ADX - DX ≥ gammaGapDX)

**Flag** : `enable_entry_counter_trend = true`

---

## 🔓 SORTIES POSITION LONG

**IMPORTANT** : Peu importe le type d'entrée (Tendance ou Contre-Tendance), une position LONG peut sortir par n'importe quelle sortie selon les flags activés.

### Sortie 1 : Exit LONG Tendance
**Conditions** :
- ✅ VWMA pente **décroissante** (inversion)
- ✅ DI- > DI+ (force baissière domine)
- ✅ DX > ADX ↑ (force baissière augmente)
- ✅ Gap DI suffisant
- ✅ Gap DX suffisant

**Flag** : `enable_exit_trend = true`

**Interprétation** : Retournement baissier fort confirmé

### Sortie 2 : Exit LONG Contre-Tendance
**Conditions** :
- ✅ VWMA pente **décroissante** (inversion)
- ✅ DI+ > DI- (force haussière domine encore)
- ✅ DX < ADX ↓ (force haussière diminue)
- ✅ Gap DI suffisant
- ✅ Gap DX suffisant

**Flag** : `enable_exit_counter_trend = true`

**Interprétation** : Début d'inversion, force haussière faiblit

---

## 🔒 SORTIES POSITION SHORT

**IMPORTANT** : Peu importe le type d'entrée (Tendance ou Contre-Tendance), une position SHORT peut sortir par n'importe quelle sortie selon les flags activés.

### Sortie 1 : Exit SHORT Tendance
**Conditions** :
- ✅ VWMA pente **croissante** (inversion)
- ✅ DI+ > DI- (force haussière domine)
- ✅ DX > ADX ↑ (force haussière augmente)
- ✅ Gap DI suffisant
- ✅ Gap DX suffisant

**Flag** : `enable_exit_trend = true`

**Interprétation** : Retournement haussier fort confirmé

### Sortie 2 : Exit SHORT Contre-Tendance
**Conditions** :
- ✅ VWMA pente **croissante** (inversion)
- ✅ DI- > DI+ (force baissière domine encore)
- ✅ DX < ADX ↓ (force baissière diminue)
- ✅ Gap DI suffisant
- ✅ Gap DX suffisant

**Flag** : `enable_exit_counter_trend = true`

**Interprétation** : Début d'inversion, force baissière faiblit

---

## 🎛️ FLAGS DE CONFIGURATION

```yaml
direction_dmi:
  # Activation signaux ENTRY
  enable_entry_trend: true          # Activer entrées Tendance (signaux 1 & 3)
  enable_entry_counter_trend: false # Activer entrées Contre-Tendance (signaux 2 & 4)
  
  # Activation signaux EXIT
  enable_exit_trend: true           # Activer sorties Tendance (force inverse augmente)
  enable_exit_counter_trend: true   # Activer sorties Contre-Tendance (force actuelle diminue)
```

### Comportement des flags

**Entry** :
- Si `enable_entry_trend = false` → Signaux 1 & 3 ignorés
- Si `enable_entry_counter_trend = false` → Signaux 2 & 4 ignorés

**Exit** :
- Si `enable_exit_trend = false` → Sorties tendance non détectées
- Si `enable_exit_counter_trend = false` → Sorties contre-tendance non détectées
- Si **AUCUN flag exit actif** → Position reste ouverte indéfiniment (DANGEREUX)

---

## 🔧 CONFIGURATIONS PRÉDÉFINIES

### Config 1 : CONSERVATEUR
```yaml
enable_entry_trend: true
enable_entry_counter_trend: false
enable_exit_trend: true
enable_exit_counter_trend: false
```
**Signaux** : Entrées tendance uniquement, sorties sur retournement fort

### Config 2 : AGRESSIF
```yaml
enable_entry_trend: true
enable_entry_counter_trend: true
enable_exit_trend: true
enable_exit_counter_trend: true
```
**Signaux** : Tous les signaux, maximum de trades

### Config 3 : ENTRY ANTICIPÉE
```yaml
enable_entry_trend: true
enable_entry_counter_trend: true
enable_exit_trend: true
enable_exit_counter_trend: false
```
**Signaux** : Entrées anticipées, sorties sur retournement fort uniquement

### Config 4 : EXIT ANTICIPÉE
```yaml
enable_entry_trend: true
enable_entry_counter_trend: false
enable_exit_trend: true
enable_exit_counter_trend: true
```
**Signaux** : Entrées conservatrices, sorties optimisées (anticipées)

---

## 📐 PARAMÈTRES TECHNIQUES

### Paramètres VWMA (hérités de Direction)
```yaml
vwma_period: 20              # Période VWMA (optimal 5m: 12-20)
slope_period: 6              # Période calcul pente (optimal: 4-6)
k_confirmation: 2            # Nombre bougies confirmation pente
use_dynamic_threshold: true  # Seuil ATR dynamique
atr_period: 8                # Période ATR
atr_coefficient: 0.25        # Coefficient ATR (optimal 5m: 0.25-0.50)
fixed_threshold: 0.1         # Seuil fixe (si dynamic = false)
```

### Paramètres DMI (nouveaux)
```yaml
dmi_period: 14               # Période DMI standard
dmi_smooth: 14               # Lissage DMI
gamma_gap_di: 2.0            # Gap minimum DI+ vs DI- (%)
gamma_gap_dx: 2.0            # Gap minimum DX vs ADX (%)
window_gamma_validate: 5     # Fenêtre validation gap (bougies)
window_matching: 5           # Fenêtre matching 3 conditions (VWMA+DMI+DX/ADX)
```

---

## 🔄 FENÊTRE DE MATCHING

### Règle de validation obligatoire

**Les 3 conditions doivent être validées dans une fenêtre de W bougies consécutives, peu importe l'ordre** :

1. ✅ **VWMA** : Pente confirmée (croissante ou décroissante avec K-confirmation)
2. ✅ **DMI** : Croisement DI+ vs DI- avec gap suffisant (≥ gammaGapDI)
3. ✅ **DX/ADX** : Croisement DX vs ADX avec gap suffisant (≥ gammaGapDX)

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
  vwma_ok = false
  dmi_ok = false  
  dx_ok = false
  
  POUR j dans fenetre:
    SI VWMA pente confirmée à j: vwma_ok = true
    SI DMI croisement valide à j: dmi_ok = true
    SI DX/ADX croisement valide à j: dx_ok = true
  
  # Générer signal si les 3 conditions réunies
  SI vwma_ok ET dmi_ok ET dx_ok:
    signal = ClassifierSignal(vwma_dir, dmi_dominant, dx_direction)
    GÉNÉRER Signal(timestamp=i, price=Close[i])
```

### Exemples

#### Signal LONG Tendance dans fenêtre W=5

**Bougie T0** : DX croise ADX ↑ (gap OK) ✅  
**Bougie T1** : Rien  
**Bougie T2** : VWMA devient croissante ✅  
**Bougie T3** : Rien  
**Bougie T4** : DI+ croise DI- (gap OK) ✅  

**Résultat** : Signal LONG Tendance généré à **T4** (timestamp=T4, prix=Close[T4])

#### Conditions simultanées

**Bougie T0** : VWMA + DMI + DX/ADX tous validés simultanément ✅✅✅  
**Résultat** : Signal généré **immédiatement** à T0 (fenêtre = 1 bougie)

### Cas limites

- **Fenêtre trop petite** : Si conditions étalées sur plus de W bougies → Pas de signal
- **Signaux chevauchants** : Plusieurs signaux possibles dans fenêtres qui se chevauchent
- **Priorité EXIT > ENTRY** : Si conflit entre sortie et entrée simultanées

### Timestamp et prix du signal

- **Timestamp** : Toujours la dernière bougie de la fenêtre (quand 3ème condition validée)
- **Prix d'exécution** : Close de cette dernière bougie
- **Fenêtre référence** : Sauvegardée dans métadonnées pour debug

---

## 🔍 DÉTECTION ET VALIDATION

### Étapes de détection

1. **Calculer indicateurs** :
   - VWMA, pente VWMA (avec K-confirmation)
   - DMI (DI+, DI-)
   - DX, ADX

2. **Détecter croisements/états sur chaque bougie** :
   - Croisement DI+ vs DI-
   - Croisement DX vs ADX
   - Direction pente VWMA

3. **Valider gaps sur chaque bougie** :
   - Gap DI suffisant (≥ gammaGapDI)
   - Gap DX suffisant (≥ gammaGapDX)

4. **Appliquer fenêtre de matching** :
   - Vérifier si les 3 conditions sont présentes dans fenêtre W
   - Classifier le signal combiné

5. **Filtrer par flags** :
   - Vérifier flag correspondant activé
   - Générer signal final si validé

---

## 🎨 MÉTADONNÉES DES SIGNAUX

```go
type DirectionDMISignal struct {
    // Identification
    Action      string  // "ENTRY" ou "EXIT"
    Type        string  // "LONG" ou "SHORT"
    Mode        string  // "TREND" ou "COUNTER_TREND"
    
    // VWMA
    VWMASlope      float64
    VWMASlopeDir   string  // "RISING" ou "FALLING"
    
    // DMI
    DIPlus         float64
    DIMinus        float64
    DIDominant     string  // "DI_PLUS" ou "DI_MINUS"
    GapDI          float64
    GapDIValid     bool
    
    // DX/ADX
    DX             float64
    ADX            float64
    DXDirection    string  // "RISING" ou "FALLING"
    GapDX          float64
    GapDXValid     bool
    
    // Contexte
    Confidence     float64
    Timestamp      time.Time
    Price          float64
    
    // Méta
    FlagsUsed      map[string]bool  // Quels flags ont permis ce signal
}
```

---

## ⚠️ RÈGLES IMPORTANTES

### 1. VWMA = Direction absolue
- VWMA croissante → LONG uniquement
- VWMA décroissante → SHORT uniquement
- **JAMAIS** de LONG si VWMA décroissante
- **JAMAIS** de SHORT si VWMA croissante

### 2. Indépendance Entry/Exit
- Une position LONG Tendance peut sortir en Exit Contre-Tendance
- Une position LONG Contre-Tendance peut sortir en Exit Tendance
- Le type d'entrée N'AFFECTE PAS le type de sortie possible

### 3. Flags obligatoires
- Au moins **un flag entry** doit être actif (sinon aucune position)
- Au moins **un flag exit** doit être actif (sinon position bloquée)

### 4. Priorités si plusieurs signaux
- **EXIT > ENTRY** (fermer avant ouvrir)
- Si plusieurs exits valides : prendre priorité 1 (Tendance) puis 2 (Contre-Tendance)

### 5. Gestion positions
- **Une seule position à la fois** (par défaut)
- Fermer position existante avant ouvrir nouvelle
- Ou autoriser plusieurs positions (paramètre à définir)

---

## 📊 EXEMPLE COMPLET

### Scénario : Retournement haussier

**T0 : Baisse en cours**
- VWMA décroissante
- DI- > DI+ (gap = 8%)
- DX > ADX ↑ (force baisse augmente)
- → **Entry SHORT Tendance** (si `enable_entry_trend = true`)

**T1 : VWMA se retourne**
- VWMA **croissante** (changement)
- DI- > DI+ **encore** (gap = 6%)
- DX < ADX ↓ (force baisse diminue)
- → **Entry LONG Contre-Tendance** (si `enable_entry_counter_trend = true`)
- → **Exit SHORT Contre-Tendance** (si `enable_exit_counter_trend = true`)

**T2 : DMI confirme**
- VWMA croissante
- DI+ > DI- (basculement, gap = 4%)
- DX > ADX ↑ (force hausse augmente)
- → **Entry LONG Tendance** (si `enable_entry_trend = true`)
- → **Exit SHORT Tendance** (si `enable_exit_trend = true`)

---

## 🚀 IMPLÉMENTATION

### Fichiers à créer

1. **`internal/signals/direction_dmi/generator.go`**
   - Structure `DirectionDMIGenerator`
   - Méthodes `Initialize()`, `CalculateIndicators()`, `DetectSignals()`
   - Logique de matching VWMA + DMI

2. **`cmd/direction_dmi_generator_demo/main.go`**
   - Demo standalone
   - Test de la stratégie
   - Export résultats JSON

3. **`cmd/direction_dmi_engine/`**
   - Moteur temporal backtest
   - Comme `direction_engine` mais avec DMI

### Tests à effectuer

1. Vérifier que les 4 signaux sont correctement détectés
2. Tester avec différentes combinaisons de flags
3. Valider que Entry/Exit sont indépendants
4. Comparer performances vs Direction simple
5. Optimiser paramètres (gammaGapDI, gammaGapDX, etc.)

---

## 📚 RÉFÉRENCES

- **Direction Generator** : `internal/signals/direction/generator.go`
- **Trend Generator** : `internal/signals/trend/generator.go` (logique DMI/DX/ADX)
- **Paramètres optimaux Direction** : `docs/RESUME_ANALYSE_DIRECTION.md`

---

**Version finale validée le 2025-11-08**
**Cette spécification fait référence - Ne pas modifier sans discussion**
