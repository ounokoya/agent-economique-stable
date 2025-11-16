# Operations Specifiques par Indicateur

## 📋 Vue d'ensemble

Ce document regroupe les operations analytiques atomiques **appliquees specifiquement a chaque indicateur** utilise dans les strategies detaillée et simplifiée.

**Principe** : Montrer comment composer les operations generiques pour chaque indicateur.

**Indicateurs des strategies** :
1. **VWMA** (Volume Weighted Moving Average)
2. **DMI** (Directional Movement Index - DI+, DI-, DX, ADX)
3. **MFI** (Money Flow Index)
4. **CHOP** (Choppiness Index)
5. **ATR%** (Average True Range Percentage)

---

## 🟢 **VWMA (Volume Weighted Moving Average)**

### Configuration
```go
vwma6 := indicators.NewVWMATVStandard(6)
vwma20 := indicators.NewVWMATVStandard(20)
vwma30 := indicators.NewVWMATVStandard(30)
```

### Operations applicables

#### 1. **Detection croisement VWMA**
```go
// Croisement VWMA rapide/lent
cross, direction := detecterCroisement(vwma6_values, vwma20_values, index)
// Retourne: (true, "HAUSSIER") ou (true, "BAISSIER") ou (false, "")
```

#### 2. **Validation gap gamma**
```go
// Ecart minimal pour validation
gap := calculerEcart(vwma6_values[index], vwma20_values[index])
gap_valide := gap >= (gamma_gap * atr_values[index])
```

#### 3. **Croisement valide complet**
```go
// Croisement + gap en une operation
cross, direction := detecterCroisementValide(vwma6_values, vwma20_values, index, gamma_gap)
```

#### 4. **Position relative tendance**
```go
// Direction actuelle
position := positionRelative(vwma6_values[index], vwma20_values[index])
// Retourne: "AU-DESSUS", "EN-DESSOUS", "EGAL"

direction := determinerDirection(vwma6_values[index], vwma20_values[index])
// Retourne: "LONG", "SHORT", "NEUTRE"
```

#### 5. **Position maintenue (tendance EN COURS)**
```go
// Verifier si position stable sur N periodes
tendance_maintenue := positionMaintenue(vwma6_values, vwma20_values, index, 5, "AU-DESSUS")
// Retourne: true si VWMA6 reste au-dessus de VWMA20 sur 5 bougies
```

#### 6. **Ecart relatif (%)**
```go
// Distance en pourcentage
distance_pct := calculerEcartRelatif(vwma6_values[index], vwma20_values[index])
```

#### 7. **Stop suiveur VWMA**
```go
// Calcul distance stop dynamique
atr_pct := normaliser(atr_values[index], prix_actuel)
distance_stop := clip(k * atr_pct, p_min, p_max)
stop_long := vwma30_values[index] * (1 - distance_stop/100)
stop_short := vwma30_values[index] * (1 + distance_stop/100)
```

---

## 🔵 **DMI (Directional Movement Index)**

### Configuration
```go
dmi := indicators.NewDMITVStandard(14, 3)
```

### Operations applicables

#### 1. **Croisement DI+ / DI-**
```go
// Detection croisement directionnel
cross_di, direction := detecterCroisement(di_plus_values, di_minus_values, index)
// Direction: "HAUSSIER" (DI+ > DI-) ou "BAISSIER" (DI- > DI+)
```

#### 2. **Validation gap DI**
```go
// Ecart DI au moment du croisement
gap_di := calculerEcart(di_plus_values[index], di_minus_values[index])
gap_di_valide := gap_di >= gamma_gap
```

#### 3. **Dominance directionnelle**
```go
// Verifier DI dominant avec ecart minimal
dominant_long := estDominant(di_plus_values[index], di_minus_values[index], gamma_gap)
dominant_short := estDominant(di_minus_values[index], di_plus_values[index], gamma_gap)
```

#### 4. **Dominance maintenue sur periode**
```go
// DI+ reste dominant sur N periodes
dominance_maintenue := positionMaintenue(di_plus_values, di_minus_values, index, 3, "AU-DESSUS")
```

#### 5. **Croisement DX / ADX**
```go
// Detection acceleration/ralentissement
cross_dx, type_momentum := detecterCroisement(dx_values, adx_values, index)
// type_momentum: "HAUSSIER" (DX > ADX = acceleration) ou "BAISSIER" (DX < ADX = ralentissement)
```

#### 6. **Overshoot DX/ADX**
```go
// Ecart au croisement
overshoot := calculerOvershoot(dx_values, adx_values, index)
overshoot_valide := overshoot >= gamma_gap
```

#### 7. **Seuil DX minimum**
```go
// Force directionnelle suffisante
dx_suffisant := depasseSeuil(dx_values[index], seuil_dx_min, "SUPERIEUR_OU_EGAL")
```

#### 8. **Position relative DX/ADX**
```go
// Savoir si en acceleration ou ralentissement
position := positionRelative(dx_values[index], adx_values[index])
// "AU-DESSUS" = en acceleration, "EN-DESSOUS" = en ralentissement
```

#### 9. **DMI Mode Tendance (complet)**
```go
// Validation mode tendance
cross_di, _ := detecterCroisement(di_plus_values, di_minus_values, index)
gap_di_ok := calculerEcart(di_plus_values[index], di_minus_values[index]) >= gamma_gap
cross_dx, _ := detecterCroisement(dx_values, adx_values, index)
overshoot_ok := calculerOvershoot(dx_values, adx_values, index) >= gamma_gap
dominance_ok := positionMaintenue(di_plus_values, di_minus_values, index, W, "AU-DESSUS")

mode_tendance_valide := cross_di && gap_di_ok && cross_dx && overshoot_ok && dominance_ok
```

#### 10. **DMI Mode Momentum (simplifie)**
```go
// Validation mode momentum
cross_dx, type_mom := detecterCroisement(dx_values, adx_values, index)
overshoot_ok := calculerOvershoot(dx_values, adx_values, index) >= gamma_gap
dx_suffisant := depasseSeuil(dx_values[index], seuil_dx_min, "SUPERIEUR_OU_EGAL")
dominance_ok := estDominant(di_plus_values[index], di_minus_values[index], gamma_gap)

mode_momentum_valide := cross_dx && overshoot_ok && dx_suffisant && dominance_ok
```

---

## 🟡 **CHOP (Choppiness Index)**

### Configuration
```go
chop := indicators.NewCHOPTVStandard(14)
```

### Operations applicables

#### 1. **Calcul pente CHOP**
```go
// Pente sur 3 periodes
pente_chop := calculerPente(chop_values, index, 3)
```

#### 2. **Validation trending (pente <= 0)**
```go
// Regime trending si pente negative ou nulle
trending := pente_chop <= tau_slope
ranging := pente_chop > tau_slope
```

#### 3. **Veto si croissance excessive**
```go
// Refus si CHOP augmente trop
veto := pente_chop > tau_slope
```

#### 4. **Detection zones Fibonacci**
```go
// Classification par seuils
choppy := depasseSeuil(chop_values[index], 61.8, "SUPERIEUR")
trending := depasseSeuil(chop_values[index], 38.2, "INFERIEUR")
neutre := dansZone(chop_values[index], 38.2, 61.8)
```

#### 5. **Transition de zone**
```go
// Sortie zone choppy
sortie_choppy := detecterTransitionSeuil(chop_values, index, 61.8, "BAISSIER")

// Entree zone trending
entree_trending := detecterTransitionSeuil(chop_values, index, 38.2, "BAISSIER")
```

#### 6. **Sens variation CHOP**
```go
// Croissant/Decroissant/Stable
sens := sensVariation(chop_values, index, 3)
// Retourne: "CROISSANT", "DECROISSANT", "STABLE"
```

#### 7. **Stabilite CHOP**
```go
// Variation faible (< 2 points)
stable := estStable(chop_values, index, 3, 2.0)
```

#### 8. **CHOP monotone decroissant**
```go
// Strictement decroissant sur N periodes
decroissant_strict := estMonotone(chop_values, index, 3, "DECROISSANT")
```

---

## 🔴 **MFI (Money Flow Index)**

### Configuration
```go
mfi := indicators.NewMFITVStandard(14)
```

### Operations applicables

#### 1. **Detection zones extremes**
```go
// Zones surachat/survente
surachat := depasseSeuil(mfi_values[index], 80, "SUPERIEUR_OU_EGAL")
survente := depasseSeuil(mfi_values[index], 20, "INFERIEUR_OU_EGAL")
zone_neutre := dansZone(mfi_values[index], 40, 60)
```

#### 2. **Calcul pente MFI**
```go
// Pente sur 3 periodes
pente_mfi := calculerPente(mfi_values, index, 3)
```

#### 3. **Validation pente favorable**
```go
// Pente suffisante
pente_favorable := pente_mfi >= tau_slope
pente_defavorable := pente_mfi <= -tau_slope
```

#### 4. **Sens variation MFI**
```go
// Direction MFI
sens := sensVariation(mfi_values, index, 3)
croissant := sens == "CROISSANT"
decroissant := sens == "DECROISSANT"
```

#### 5. **Constance MFI**
```go
// Variation faible (< 2 points)
constance := estStable(mfi_values, index, 3, 2.0)
```

#### 6. **Transition zones MFI**
```go
// Sortie zone survente
sortie_survente := detecterTransitionSeuil(mfi_values, index, 20, "HAUSSIER")

// Entree zone surachat
entree_surachat := detecterTransitionSeuil(mfi_values, index, 80, "HAUSSIER")

// Sortie zone surachat
sortie_surachat := detecterTransitionSeuil(mfi_values, index, 80, "BAISSIER")

// Entree zone survente
entree_survente := detecterTransitionSeuil(mfi_values, index, 20, "BAISSIER")
```

#### 7. **Synchronisation MFI avec autres indicateurs**
```go
// MFI et CHOP varient dans sens oppose (MFI up = CHOP down)
sync_mfi_chop := !memeSensVariation(mfi_values, chop_values, index, 3)

// MFI et pente VWMA dans meme sens
pente_vwma := calculerPente(vwma6_values, index, 3)
pente_mfi := calculerPente(mfi_values, index, 3)
sync_mfi_vwma := (pente_mfi > 0) == (pente_vwma > 0)
```

#### 8. **MFI monotone**
```go
// MFI strictement croissant
mfi_croissant := estMonotone(mfi_values, index, 3, "CROISSANT")

// MFI strictement decroissant
mfi_decroissant := estMonotone(mfi_values, index, 3, "DECROISSANT")
```

#### 9. **Ajustement seuils MFI par contexte DMI**
```go
// Modulation dynamique selon force DI
gap_di := calculerEcart(di_plus_values[index], di_minus_values[index])
di_fort := gap_di > seuil_di_fort

if di_fort {
    seuil_surachat_ajuste := 80 + ajustement_strict
    seuil_survente_ajuste := 20 - ajustement_strict
} else {
    seuil_surachat_ajuste := 80
    seuil_survente_ajuste := 20
}

surachat := depasseSeuil(mfi_values[index], seuil_surachat_ajuste, "SUPERIEUR_OU_EGAL")
```

---

## 🟠 **ATR% (Average True Range Percentage)**

### Configuration
```go
atr := indicators.NewATRTVStandard(14)
```

### Operations applicables

#### 1. **Calcul ATR en pourcentage**
```go
// Normalisation par prix
atr_pct := normaliser(atr_values[index], prix_actuel)
// ou
atr_pct := calculerEcartRelatif(atr_values[index], prix_actuel)
```

#### 2. **Validation volatilite minimale**
```go
// Volatilite suffisante pour trading
vol_ok := depasseSeuil(atr_pct, seuil_min_vol, "SUPERIEUR_OU_EGAL")
```

#### 3. **Classification volatilite**
```go
// Categories de volatilite
faible := depasseSeuil(atr_pct, 0.15, "INFERIEUR")
moderee := dansZone(atr_pct, 0.15, 0.30)
elevee := depasseSeuil(atr_pct, 0.30, "SUPERIEUR_OU_EGAL")
```

#### 4. **Sizing stop dynamique**
```go
// Dimensionnement stop selon ATR
distance_stop := clip(k * atr_pct, p_min, p_max)
```

#### 5. **Tendance volatilite**
```go
// Pente ATR
pente_atr := calculerPente(atr_values, index, 5)

// Sens variation
sens := sensVariation(atr_values, index, 5)
expansion := sens == "CROISSANT"
contraction := sens == "DECROISSANT"
```

#### 6. **ATR monotone**
```go
// Volatilite en expansion continue
expansion_continue := estMonotone(atr_values, index, 5, "CROISSANT")

// Volatilite en contraction continue
contraction_continue := estMonotone(atr_values, index, 5, "DECROISSANT")
```

#### 7. **Extremum ATR**
```go
// ATR max sur periode
atr_max := trouverExtremum(atr_values, index, 20, "MAX")

// ATR min sur periode
atr_min := trouverExtremum(atr_values, index, 20, "MIN")
```

---

## 🎯 **OPERATIONS COMBINEES MULTI-INDICATEURS**

### 1. **Alignement directionnel**
```go
// Verifier coherence directionnelle
vwma_pos := positionRelative(vwma6_values[index], vwma20_values[index])
di_pos := positionRelative(di_plus_values[index], di_minus_values[index])
mfi_pos := sensVariation(mfi_values, index, 3)

aligne_haussier := vwma_pos == "AU-DESSUS" && di_pos == "AU-DESSUS" && mfi_pos == "CROISSANT"
aligne_baissier := vwma_pos == "EN-DESSOUS" && di_pos == "EN-DESSOUS" && mfi_pos == "DECROISSANT"
```

### 2. **Score force tendance composite**
```go
// Force composite de tendance (0-100)
gap_vwma := calculerEcart(vwma6_values[index], vwma20_values[index])
gap_di := calculerEcart(di_plus_values[index], di_minus_values[index])

score_vwma := clip(gap_vwma / max_gap_vwma * 30, 0, 30)
score_di := clip(gap_di / max_gap_di * 30, 0, 30)
score_dx := clip(dx_values[index] / max_dx * 20, 0, 20)
score_mfi := clip(mfi_values[index] / 100 * 20, 0, 20)

score_total := score_vwma + score_di + score_dx + score_mfi
```

### 3. **Validation setup complet**
```go
// Setup de trading complet
vwma_cross := detecterCroisementValide(vwma6_values, vwma20_values, index, gamma_gap)
dmi_valide := /* validation DMI complete */
mfi_extreme := depasseSeuil(mfi_values[index], 80, "SUPERIEUR_OU_EGAL")
chop_trending := calculerPente(chop_values, index, 3) <= tau_slope
atr_ok := depasseSeuil(calculerEcartRelatif(atr_values[index], prix), 0.15, "SUPERIEUR_OU_EGAL")

setup_complet := vwma_cross && dmi_valide && mfi_extreme && chop_trending && atr_ok
```

### 4. **Validation fenetre W**
```go
// Valider toutes conditions dans fenetre
conditions := []bool{
    vwma_cross,
    dmi_valide,
    mfi_extreme,
    chop_trending,
    atr_ok,
}

valide_fenetre := validerDansFenetre(conditions, index_debut, W)
nb_validations := compterValidations(conditions, index_debut, W)
```

---

## 📊 **RECAPITULATIF PAR INDICATEUR**

**Note importante** : Ce document ne liste QUE les indicateurs utilises dans les strategies detaillée et simplifiée.

### **VWMA - 7 operations**
1. Croisement
2. Gap validation
3. Position relative
4. Position maintenue
5. Ecart relatif
6. Stop suiveur
7. Croisement valide

### **DMI - 10 operations**
1. Croisement DI
2. Gap DI
3. Dominance
4. Dominance maintenue
5. Croisement DX/ADX
6. Overshoot
7. Seuil DX
8. Position DX/ADX
9. Mode tendance
10. Mode momentum

### **CHOP - 8 operations**
1. Pente
2. Validation trending
3. Veto croissance
4. Zones Fibonacci
5. Transition zone
6. Sens variation
7. Stabilite
8. Monotonie

### **MFI - 9 operations**
1. Zones extremes
2. Pente
3. Pente favorable
4. Sens variation
5. Constance
6. Transition zones
7. Synchronisation
8. Monotonie
9. Ajustement seuils

### **ATR% - 7 operations**
1. Calcul %
2. Validation volatilite
3. Classification
4. Sizing stop
5. Tendance volatilite
6. Monotonie
7. Extremum

### **COMBINEES - 4 operations**
1. Alignement directionnel
2. Score force
3. Setup complet
4. Validation fenetre

---

## 🚀 **USAGE**

**Chaque indicateur utilise les memes operations atomiques de base, appliquees specifiquement a ses valeurs.**

```go
// Exemple: Meme fonction pour tous
pente_mfi := calculerPente(mfi_values, index, 3)
pente_chop := calculerPente(chop_values, index, 3)
pente_atr := calculerPente(atr_values, index, 5)

// Meme fonction croisement pour tous
cross_vwma, _ := detecterCroisement(vwma6, vwma20, index)
cross_di, _ := detecterCroisement(di_plus, di_minus, index)
cross_dx_adx, _ := detecterCroisement(dx, adx, index)
```

**Total: 45 operations specifiques (5 indicateurs + operations combinees) reutilisant 25 operations atomiques generiques.**

**Indicateurs utilises :**
- ✅ VWMA (7 ops)
- ✅ DMI (10 ops)
- ✅ CHOP (8 ops)
- ✅ MFI (9 ops)
- ✅ ATR% (7 ops)
- ✅ Combinees (4 ops)
