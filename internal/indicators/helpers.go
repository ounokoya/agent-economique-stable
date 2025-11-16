package indicators

import "math"

// ============================================================================
// CATEGORIE 1 : CALCULS DE BASE
// ============================================================================

// CalculerPente calcule la pente entre 2 valeurs sur N periodes
func CalculerPente(serie []float64, indexActuel, periodes int) float64 {
	if indexActuel < periodes {
		return 0.0
	}
	return (serie[indexActuel] - serie[indexActuel-periodes]) / float64(periodes)
}

// CalculerEcart calcule la distance absolue entre 2 valeurs
func CalculerEcart(valeur1, valeur2 float64) float64 {
	return math.Abs(valeur1 - valeur2)
}

// CalculerEcartRelatif calcule la distance relative en pourcentage
func CalculerEcartRelatif(valeur, reference float64) float64 {
	if reference == 0 {
		return 0.0
	}
	return math.Abs(valeur-reference) / reference * 100
}

// Normaliser normalise une valeur par une reference
func Normaliser(valeur, reference float64) float64 {
	if reference == 0 {
		return 0.0
	}
	return valeur / reference * 100
}

// CalculerMoyenne calcule la moyenne mobile simple sur N periodes
func CalculerMoyenne(serie []float64, indexActuel, periodes int) float64 {
	if indexActuel < periodes-1 {
		return 0.0
	}
	somme := 0.0
	for i := indexActuel - periodes + 1; i <= indexActuel; i++ {
		somme += serie[i]
	}
	return somme / float64(periodes)
}

// CalculerVariation calcule la variation entre 2 points
func CalculerVariation(serie []float64, indexActuel, periodes int) float64 {
	if indexActuel < periodes {
		return 0.0
	}
	return serie[indexActuel] - serie[indexActuel-periodes]
}

// ============================================================================
// CATEGORIE 2 : DETECTION CROISEMENTS
// ============================================================================

// DetecterCroisement detecte un croisement entre 2 series
// Retourne (croisement detecte, direction "HAUSSIER" ou "BAISSIER")
// NOTE: utilise les barres FERMEES (index = n-1, index-1 = n-2)
func DetecterCroisement(rapide, lent []float64, index int) (bool, string) {
	// On détecte le croisement entre 2 barres FERMÉES
	// index = dernière barre fermée (n-1)
	// index-1 = avant-dernière barre fermée (n-2)
	if index < 1 {
		return false, ""
	}

	precRapide := rapide[index-1]  // n-2 (avant-dernière fermée)
	actuelRapide := rapide[index]  // n-1 (dernière fermée)
	precLent := lent[index-1]      // n-2
	actuelLent := lent[index]      // n-1

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

// DetecterCroisementValide detecte un croisement avec validation d'ecart minimal
func DetecterCroisementValide(rapide, lent []float64, index int, ecartMin float64) (bool, string) {
	croisement, direction := DetecterCroisement(rapide, lent, index)
	if !croisement {
		return false, ""
	}

	ecart := CalculerEcart(rapide[index], lent[index])
	if ecart >= ecartMin {
		return true, direction
	}

	return false, ""
}

// CalculerOvershoot calcule l'ecart au moment du croisement
func CalculerOvershoot(serie1, serie2 []float64, index int) float64 {
	return CalculerEcart(serie1[index], serie2[index])
}

// ============================================================================
// CATEGORIE 3 : COMPARAISONS ET POSITIONS
// ============================================================================

// PositionRelative compare 2 valeurs et retourne leur position relative
func PositionRelative(valeur1, valeur2 float64) string {
	if valeur1 > valeur2 {
		return "AU-DESSUS"
	} else if valeur1 < valeur2 {
		return "EN-DESSOUS"
	} else {
		return "EGAL"
	}
}

// DeterminerDirection determine la direction selon la position
func DeterminerDirection(valeur1, valeur2 float64) string {
	if valeur1 > valeur2 {
		return "LONG"
	} else if valeur1 < valeur2 {
		return "SHORT"
	} else {
		return "NEUTRE"
	}
}

// EstDominant verifie si valeur1 domine valeur2 avec un ecart minimal
func EstDominant(valeur1, valeur2, ecartMin float64) bool {
	return valeur1 > valeur2 && (valeur1-valeur2) >= ecartMin
}

// SensVariation determine si une serie est croissante/decroissante
func SensVariation(serie []float64, indexActuel, periodes int) string {
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

// ============================================================================
// CATEGORIE 4 : VALIDATION SEUILS
// ============================================================================

// DepasseSeuil teste si une valeur depasse un seuil
func DepasseSeuil(valeur, seuil float64, typeTest string) bool {
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

// DansZone verifie si une valeur est dans une zone entre 2 seuils
func DansZone(valeur, seuilBas, seuilHaut float64) bool {
	return valeur >= seuilBas && valeur <= seuilHaut
}

// DetecterTransitionSeuil detecte un franchissement de seuil
func DetecterTransitionSeuil(serie []float64, index int, seuil float64, direction string) bool {
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

// ============================================================================
// CATEGORIE 5 : TENDANCES ET MAINTIEN
// ============================================================================

// PositionMaintenue verifie qu'une position relative est stable sur N periodes
func PositionMaintenue(serie1, serie2 []float64, indexActuel, periodes int, position string) bool {
	if indexActuel < periodes {
		return false
	}

	for i := indexActuel - periodes + 1; i <= indexActuel; i++ {
		posActuelle := PositionRelative(serie1[i], serie2[i])
		if posActuelle != position {
			return false
		}
	}
	return true
}

// EstMonotone verifie si une serie est strictement monotone
func EstMonotone(serie []float64, indexActuel, periodes int, typeMonotonie string) bool {
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

// EstStable verifie si la variation est dans une limite
func EstStable(serie []float64, indexActuel, periodes int, variationMax float64) bool {
	if indexActuel < periodes {
		return false
	}

	variation := CalculerVariation(serie, indexActuel, periodes)
	return math.Abs(variation) <= variationMax
}

// ============================================================================
// CATEGORIE 6 : UTILITAIRES
// ============================================================================

// Clip contraint une valeur dans des bornes
func Clip(valeur, min, max float64) float64 {
	if valeur < min {
		return min
	} else if valeur > max {
		return max
	}
	return valeur
}

// ValiderDansFenetre verifie que toutes les conditions sont vraies dans une fenetre
func ValiderDansFenetre(conditions []bool, indexDebut, tailleW int) bool {
	for i := indexDebut; i < indexDebut+tailleW && i < len(conditions); i++ {
		if !conditions[i] {
			return false
		}
	}
	return true
}

// CompterValidations compte le nombre de validations dans une fenetre
func CompterValidations(conditions []bool, indexDebut, tailleW int) int {
	count := 0
	for i := indexDebut; i < indexDebut+tailleW && i < len(conditions); i++ {
		if conditions[i] {
			count++
		}
	}
	return count
}

// TrouverExtremum trouve le maximum ou minimum sur N periodes
func TrouverExtremum(serie []float64, indexActuel, periodes int, typeExtremum string) float64 {
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

// ============================================================================
// CATEGORIE 7 : SYNCHRONISATION ET COHERENCE
// ============================================================================

// MemeSensVariation verifie si 2 series varient dans le meme sens
func MemeSensVariation(serie1, serie2 []float64, index, periodes int) bool {
	if index < periodes {
		return false
	}

	sens1 := SensVariation(serie1, index, periodes)
	sens2 := SensVariation(serie2, index, periodes)

	return sens1 == sens2 && sens1 != "STABLE"
}

// AlignementMultiple verifie que plusieurs valeurs ont la meme position relative
func AlignementMultiple(valeurs []float64, reference float64, position string) bool {
	for _, val := range valeurs {
		posActuelle := PositionRelative(val, reference)
		if posActuelle != position {
			return false
		}
	}
	return true
}
