# Operations Analytiques Generiques - Boite a Outils Reutilisable

## 📋 Vue d'ensemble

Ce document liste toutes les **operations mathematiques et analytiques atomiques** reutilisables pour n'importe quel indicateur ou strategie.

**Principe** : Fonctions generiques et composables, pas specialisees par indicateur.

---

## 🔢 **CATEGORIE 1 : CALCULS DE BASE**

### 1.1 **Calcul Pente**
```go
// Operation atomique : Pente entre 2 valeurs sur N periodes
func calculerPente(serie []float64, indexActuel, periodes int) float64 {
    if indexActuel < periodes {
        return 0.0
    }
    return (serie[indexActuel] - serie[indexActuel-periodes]) / float64(periodes)
}

// Usage universel :
// - pente_mfi = calculerPente(mfi, index, 3)
// - pente_chop = calculerPente(chop, index, 3)
// - pente_atr = calculerPente(atr, index, 5)
// - pente_vwma = calculerPente(vwma, index, 10)
```

### 1.2 **Calcul Ecart Absolu**
```go
// Operation atomique : Distance entre 2 valeurs
func calculerEcart(valeur1, valeur2 float64) float64 {
    return math.Abs(valeur1 - valeur2)
}

// Usage universel :
// - gap_vwma = calculerEcart(vwma6, vwma20)
// - gap_di = calculerEcart(di_plus, di_moins)
// - gap_dx_adx = calculerEcart(dx, adx)
// - ecart_k_d = calculerEcart(k, d)
```

### 1.3 **Calcul Ecart Relatif (%)**
```go
// Operation atomique : Distance relative en pourcentage
func calculerEcartRelatif(valeur, reference float64) float64 {
    if reference == 0 {
        return 0.0
    }
    return math.Abs(valeur - reference) / reference * 100
}

// Usage universel :
// - distance_pct = calculerEcartRelatif(vwma6, vwma20)
// - atr_pct = calculerEcartRelatif(atr, prix)
// - deviation_mfi = calculerEcartRelatif(mfi_actuel, mfi_moyen)
```

### 1.4 **Normalisation Par Valeur**
```go
// Operation atomique : Normaliser valeur par reference
func normaliser(valeur, reference float64) float64 {
    if reference == 0 {
        return 0.0
    }
    return valeur / reference * 100
}

// Usage universel :
// - atr_pct = normaliser(atr, prix)
// - volume_relatif = normaliser(volume_actuel, volume_moyen)
```

### 1.5 **Moyenne sur N Periodes**
```go
// Operation atomique : Moyenne mobile simple
func calculerMoyenne(serie []float64, indexActuel, periodes int) float64 {
    if indexActuel < periodes-1 {
        return 0.0
    }
    somme := 0.0
    for i := indexActuel - periodes + 1; i <= indexActuel; i++ {
        somme += serie[i]
    }
    return somme / float64(periodes)
}

// Usage universel :
// - moy_mfi = calculerMoyenne(mfi, index, 5)
// - moy_volume = calculerMoyenne(volume, index, 20)
// - moy_atr = calculerMoyenne(atr, index, 10)
```

### 1.6 **Variation Absolue**
```go
// Operation atomique : Variation entre 2 points
func calculerVariation(serie []float64, indexActuel, periodes int) float64 {
    if indexActuel < periodes {
        return 0.0
    }
    return serie[indexActuel] - serie[indexActuel-periodes]
}

// Usage universel :
// - var_mfi = calculerVariation(mfi, index, 3)
// - var_prix = calculerVariation(prix, index, 1)
```

---

## 🔄 **CATEGORIE 2 : DETECTION CROISEMENTS**

### 2.1 **Croisement Haussier/Baissier**
```go
// Operation atomique : Detecter croisement entre 2 series
func detecterCroisement(rapide, lent []float64, index int) (bool, string) {
    if index < 1 {
        return false, ""
    }
    
    precRapide := rapide[index-1]
    actuelRapide := rapide[index]
    precLent := lent[index-1]
    actuelLent := lent[index]
    
    // Croisement haussier
    if precRapide <= precLent && actuelRapide > actuelLent {
        return true, "HAUSSIER"
    }
    
    // Croisement baissier
    if precRapide >= precLent && actuelRapide < actuelLent {
        return true, "BAISSIER"
    }
    
    return false, ""
}

// Usage universel :
// - cross, dir = detecterCroisement(vwma6, vwma20, index)
// - cross, dir = detecterCroisement(di_plus, di_moins, index)
// - cross, dir = detecterCroisement(dx, adx, index)
// - cross, dir = detecterCroisement(k, d, index)
// - cross, dir = detecterCroisement(macd, signal, index)
```

### 2.2 **Croisement Avec Validation Ecart**
```go
// Operation atomique : Croisement valide par ecart minimal
func detecterCroisementValide(rapide, lent []float64, index int, ecartMin float64) (bool, string) {
    croisement, direction := detecterCroisement(rapide, lent, index)
    if !croisement {
        return false, ""
    }
    
    ecart := calculerEcart(rapide[index], lent[index])
    if ecart >= ecartMin {
        return true, direction
    }
    
    return false, ""
}

// Usage universel :
// - cross, dir = detecterCroisementValide(vwma6, vwma20, index, gamma_gap)
// - cross, dir = detecterCroisementValide(di_plus, di_moins, index, gamma_gap)
// - cross, dir = detecterCroisementValide(dx, adx, index, gamma_gap)
```

### 2.3 **Overshoot/Undershoot**
```go
// Operation atomique : Ecart au moment du croisement
func calculerOvershoot(serie1, serie2 []float64, index int) float64 {
    return calculerEcart(serie1[index], serie2[index])
}

// Usage universel :
// - overshoot = calculerOvershoot(dx, adx, index)
// - ecart_croisement = calculerOvershoot(vwma6, vwma20, index)
```

---

## 📊 **CATEGORIE 3 : COMPARAISONS ET POSITIONS**

### 3.1 **Position Relative**
```go
// Operation atomique : Comparer 2 valeurs
func positionRelative(valeur1, valeur2 float64) string {
    if valeur1 > valeur2 {
        return "AU-DESSUS"
    } else if valeur1 < valeur2 {
        return "EN-DESSOUS"
    } else {
        return "EGAL"
    }
}

// Usage universel :
// - pos = positionRelative(vwma6, vwma20)
// - pos = positionRelative(di_plus, di_moins)
// - pos = positionRelative(prix, vwma)
```

### 3.2 **Direction Basee sur Position**
```go
// Operation atomique : Direction selon position
func determinerDirection(valeur1, valeur2 float64) string {
    if valeur1 > valeur2 {
        return "LONG"
    } else if valeur1 < valeur2 {
        return "SHORT"
    } else {
        return "NEUTRE"
    }
}

// Usage universel :
// - direction = determinerDirection(vwma6, vwma20)
// - direction = determinerDirection(di_plus, di_moins)
// - direction = determinerDirection(prix, vwma96)
```

### 3.3 **Dominance Avec Ecart Minimal**
```go
// Operation atomique : Dominance avec ecart valide
func estDominant(valeur1, valeur2, ecartMin float64) bool {
    return valeur1 > valeur2 && (valeur1 - valeur2) >= ecartMin
}

// Usage universel :
// - dominant = estDominant(di_plus, di_moins, gamma_gap)
// - gap_valide = estDominant(vwma6, vwma20, gamma_gap)
// - force_suffisante = estDominant(dx, seuil_min, 0)
```

### 3.4 **Sens de Variation**
```go
// Operation atomique : Determiner si croissant/decroissant
func sensVariation(serie []float64, indexActuel, periodes int) string {
    if indexActuel < periodes {
        return "INDETERMINE"
    }
    
    variation := serie[indexActuel] - serie[indexActuel-periodes]
    if variation > 0 {
        return "CROISSANT"
    } else if variation < 0 {
        return "DECROISSANT"
    } else {
        return "STABLE"
    }
}

// Usage universel :
// - sens = sensVariation(mfi, index, 3)
// - sens = sensVariation(chop, index, 3)
// - sens = sensVariation(k, index, 2)
```

---

## 🎯 **CATEGORIE 4 : VALIDATION SEUILS**

### 4.1 **Test Seuil Unique**
```go
// Operation atomique : Valeur depasse seuil
func depasseSeuil(valeur, seuil float64, typeTest string) bool {
    switch typeTest {
    case "SUPERIEUR":
        return valeur > seuil
    case "INFERIEUR":
        return valeur < seuil
    case "SUPERIEUR_OU_EGAL":
        return valeur >= seuil
    case "INFERIEUR_OU_EGAL":
        return valeur <= seuil
    default:
        return false
    }
}

// Usage universel :
// - surachat = depasseSeuil(mfi, 80, "SUPERIEUR_OU_EGAL")
// - survente = depasseSeuil(mfi, 20, "INFERIEUR_OU_EGAL")
// - vol_ok = depasseSeuil(atr_pct, 0.15, "SUPERIEUR_OU_EGAL")
// - choppy = depasseSeuil(chop, 61.8, "SUPERIEUR")
```

### 4.2 **Test Zone (entre 2 seuils)**
```go
// Operation atomique : Valeur dans zone
func dansZone(valeur, seuilBas, seuilHaut float64) bool {
    return valeur >= seuilBas && valeur <= seuilHaut
}

// Usage universel :
// - zone_neutre_mfi = dansZone(mfi, 40, 60)
// - zone_neutre_chop = dansZone(chop, 38.2, 61.8)
// - zone_normale_cci = dansZone(cci, -100, 100)
```

### 4.3 **Transition de Seuil**
```go
// Operation atomique : Detecter franchissement de seuil
func detecterTransitionSeuil(serie []float64, index int, seuil float64, direction string) bool {
    if index < 1 {
        return false
    }
    
    valeurPrec := serie[index-1]
    valeurActuelle := serie[index]
    
    if direction == "HAUSSIER" {
        // Passage au-dessus du seuil
        return valeurPrec <= seuil && valeurActuelle > seuil
    } else if direction == "BAISSIER" {
        // Passage en-dessous du seuil
        return valeurPrec >= seuil && valeurActuelle < seuil
    }
    
    return false
}

// Usage universel :
// - sortie_survente = detecterTransitionSeuil(mfi, index, 20, "HAUSSIER")
// - entree_surachat = detecterTransitionSeuil(mfi, index, 80, "HAUSSIER")
// - sortie_choppy = detecterTransitionSeuil(chop, index, 61.8, "BAISSIER")
```

---

## 📈 **CATEGORIE 5 : TENDANCES ET MAINTIEN**

### 5.1 **Position Maintenue sur N Periodes**
```go
// Operation atomique : Verifier position stable
func positionMaintenue(serie1, serie2 []float64, indexActuel, periodes int, position string) bool {
    if indexActuel < periodes {
        return false
    }
    
    for i := indexActuel - periodes + 1; i <= indexActuel; i++ {
        posActuelle := positionRelative(serie1[i], serie2[i])
        if posActuelle != position {
            return false
        }
    }
    return true
}

// Usage universel :
// - maintenue = positionMaintenue(vwma6, vwma20, index, 5, "AU-DESSUS")
// - dominance = positionMaintenue(di_plus, di_moins, index, 3, "AU-DESSUS")
```

### 5.2 **Monotonie (toujours croissant/decroissant)**
```go
// Operation atomique : Serie strictement monotone
func estMonotone(serie []float64, indexActuel, periodes int, typeMonotonie string) bool {
    if indexActuel < periodes {
        return false
    }
    
    for i := indexActuel - periodes + 1; i < indexActuel; i++ {
        if typeMonotonie == "CROISSANT" && serie[i+1] <= serie[i] {
            return false
        } else if typeMonotonie == "DECROISSANT" && serie[i+1] >= serie[i] {
            return false
        }
    }
    return true
}

// Usage universel :
// - mfi_croissant = estMonotone(mfi, index, 3, "CROISSANT")
// - chop_decroissant = estMonotone(chop, index, 3, "DECROISSANT")
```

### 5.3 **Stabilite (variation faible)**
```go
// Operation atomique : Variation dans limite
func estStable(serie []float64, indexActuel, periodes int, variationMax float64) bool {
    if indexActuel < periodes {
        return false
    }
    
    variation := calculerVariation(serie, indexActuel, periodes)
    return math.Abs(variation) <= variationMax
}

// Usage universel :
// - mfi_stable = estStable(mfi, index, 3, 2.0)
// - prix_stable = estStable(prix, index, 5, 0.5)
```

---

## 🔧 **CATEGORIE 6 : UTILITAIRES**

### 6.1 **Clipping (limitation valeur)**
```go
// Operation atomique : Contraindre valeur dans bornes
func clip(valeur, min, max float64) float64 {
    if valeur < min {
        return min
    } else if valeur > max {
        return max
    }
    return valeur
}

// Usage universel :
// - distance_stop = clip(k * atr_pct, p_min, p_max)
// - ajustement = clip(modulation, -10, 10)
```

### 6.2 **Fenetre de Validation**
```go
// Operation atomique : Verifier conditions dans fenetre
func validerDansFenetre(conditions []bool, indexDebut, tailleW int) bool {
    for i := indexDebut; i < indexDebut+tailleW && i < len(conditions); i++ {
        if !conditions[i] {
            return false
        }
    }
    return true
}

// Usage universel :
// - valide = validerDansFenetre(conditions, index, W)
```

### 6.3 **Comptage Occurrences**
```go
// Operation atomique : Compter validations
func compterValidations(conditions []bool, indexDebut, tailleW int) int {
    count := 0
    for i := indexDebut; i < indexDebut+tailleW && i < len(conditions); i++ {
        if conditions[i] {
            count++
        }
    }
    return count
}

// Usage universel :
// - nb_validations = compterValidations(conditions, index, W)
```

### 6.4 **Maximum/Minimum sur Periode**
```go
// Operation atomique : Extremum sur N periodes
func trouverExtremum(serie []float64, indexActuel, periodes int, typeExtremum string) float64 {
    if indexActuel < periodes-1 {
        return 0.0
    }
    
    extremum := serie[indexActuel-periodes+1]
    for i := indexActuel - periodes + 2; i <= indexActuel; i++ {
        if typeExtremum == "MAX" && serie[i] > extremum {
            extremum = serie[i]
        } else if typeExtremum == "MIN" && serie[i] < extremum {
            extremum = serie[i]
        }
    }
    return extremum
}

// Usage universel :
// - max_prix = trouverExtremum(prix, index, 14, "MAX")
// - min_prix = trouverExtremum(prix, index, 14, "MIN")
```

---

## 🎯 **CATEGORIE 7 : SYNCHRONISATION ET COHERENCE**

### 7.1 **Meme Sens de Variation**
```go
// Operation atomique : 2 series varient dans meme sens
func memeSensVariation(serie1, serie2 []float64, index, periodes int) bool {
    if index < periodes {
        return false
    }
    
    sens1 := sensVariation(serie1, index, periodes)
    sens2 := sensVariation(serie2, index, periodes)
    
    return sens1 == sens2 && sens1 != "STABLE"
}

// Usage universel :
// - sync = memeSensVariation(mfi, k_stoch, index, 3)
// - sync = memeSensVariation(cci, mfi, index, 2)
```

### 7.2 **Alignement Multiple**
```go
// Operation atomique : Verifier meme position pour N series
func alignementMultiple(valeurs []float64, reference float64, position string) bool {
    for _, val := range valeurs {
        posActuelle := positionRelative(val, reference)
        if posActuelle != position {
            return false
        }
    }
    return true
}

// Usage universel :
// - aligne = alignementMultiple([]float64{vwma6, prix, di_plus}, vwma20, "AU-DESSUS")
```

---

## 📋 **RECAPITULATIF DES OPERATIONS**

### **Calculs de base (6 ops)**
1. Pente
2. Ecart absolu
3. Ecart relatif %
4. Normalisation
5. Moyenne
6. Variation

### **Croisements (3 ops)**
1. Detection croisement
2. Croisement valide
3. Overshoot/Undershoot

### **Comparaisons (4 ops)**
1. Position relative
2. Direction
3. Dominance
4. Sens variation

### **Seuils (3 ops)**
1. Test seuil unique
2. Test zone
3. Transition seuil

### **Tendances (3 ops)**
1. Position maintenue
2. Monotonie
3. Stabilite

### **Utilitaires (4 ops)**
1. Clipping
2. Fenetre validation
3. Comptage
4. Extremum

### **Synchronisation (2 ops)**
1. Meme sens
2. Alignement multiple

---

## 🚀 **PRINCIPE D'UTILISATION**

**Toutes ces operations sont ATOMIQUES et COMPOSABLES**

Exemple composition :
```go
// Validation VWMA avec gamma_gap
cross, dir := detecterCroisement(vwma6, vwma20, index)
gap := calculerEcart(vwma6[index], vwma20[index])
gap_valide := gap >= gamma_gap * atr[index]
signal_vwma := cross && gap_valide

// Validation DMI mode tendance
cross_di, dir_di := detecterCroisement(di_plus, di_moins, index)
gap_di := calculerEcart(di_plus[index], di_moins[index])
gap_di_valide := gap_di >= gamma_gap
cross_dx, dir_dx := detecterCroisement(dx, adx, index)
overshoot := calculerOvershoot(dx, adx, index)
overshoot_valide := overshoot >= gamma_gap
dmi_valide := cross_di && gap_di_valide && cross_dx && overshoot_valide

// MFI zone + pente
surachat := depasseSeuil(mfi[index], 80, "SUPERIEUR_OU_EGAL")
pente_mfi := calculerPente(mfi, index, 3)
pente_favorable := pente_mfi >= tau_slope
mfi_valide := surachat && pente_favorable

// CHOP trending
pente_chop := calculerPente(chop, index, 3)
trending := pente_chop <= tau_slope

// Signal final
signal := signal_vwma && dmi_valide && mfi_valide && trending
```

**Chaque operation est reutilisable pour N'IMPORTE QUEL indicateur ou strategie.**
