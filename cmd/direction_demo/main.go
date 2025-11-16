package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"agent-economique/internal/datasource/gateio"
	"agent-economique/internal/indicators"
)

// ============================================================================
// CONSTANTES
// ============================================================================

// Symbole et donnees
const (
	SYMBOL     = "SOL_USDT" // Format Gate.io
	TIMEFRAME  = "1d"
	NB_CANDLES = 1000
)

// Periodes
const (
	VWMA_RAPIDE = 3
)

// Calibrage pentes
const (
	PERIODE_PENTE    = 2   // Nombre de bougies pour calculer la variation
	SEUIL_PENTE_VWMA = 0.1 // Variation minimale pour VWMA6 en % (mode fixe)
	K_CONFIRMATION   = 2   // Nombre de bougies pour confirmer un changement de sens

	// Seuil dynamique basé sur ATR (optionnel)
	USE_DYNAMIC_THRESHOLD = true // true = ATR, false = seuil fixe
	ATR_PERIODE           = 14   // Période ATR
	ATR_COEFFICIENT       = 1    // Coefficient à multiplier à ATR pour obtenir le seuil
)

// ============================================================================
// STRUCTURES
// ============================================================================

type Config struct {
	Symbol              string
	Timeframe           string
	NbCandles           int
	VwmaRapide          int
	PeriodePente        int
	SeuilPenteVWMA      float64
	KConfirmation       int
	UseDynamicThreshold bool
	AtrPeriode          int
	AtrCoefficient      float64
}

// EtatDirection represente l'etat directionnel d'une bougie
type EtatDirection struct {
	Index         int
	Timestamp     time.Time
	SensVWMA6     string
	VariationVWMA float64
	VWMA6         float64
}

// Intervalle represente un intervalle directionnel VWMA6
type Intervalle struct {
	Numero          int
	Sens            string // "CROISSANT", "DECROISSANT", "STABLE"
	IndexDebut      int
	IndexFin        int
	DateDebut       time.Time
	DateFin         time.Time
	NbBougies       int
	ValeurMoyenne   float64
	PrixDebut       float64
	PrixFin         float64
	VariationCaptee float64 // En pourcentage
}

// ============================================================================
// ANALYSE DIRECTION
// ============================================================================

func analyzeDirection(config Config, klines []gateio.Kline, vwma6, atr []float64) []EtatDirection {
	etats := []EtatDirection{}

	startIdx := config.PeriodePente + 5

	for i := startIdx; i < len(klines)-1; i++ {
		etat := EtatDirection{
			Index:     i,
			Timestamp: klines[i].OpenTime,
			VWMA6:     vwma6[i],
		}

		// Calculer variation VWMA6 en pourcentage
		if vwma6[i-config.PeriodePente] != 0 {
			etat.VariationVWMA = (vwma6[i] - vwma6[i-config.PeriodePente]) / vwma6[i-config.PeriodePente] * 100
		}

		// Déterminer le seuil (fixe ou dynamique)
		var seuilVWMA float64
		if config.UseDynamicThreshold && atr != nil && i < len(atr) {
			// Seuil dynamique : ATR × Coefficient converti en %
			seuilAbsolu := atr[i] * config.AtrCoefficient
			if vwma6[i] != 0 {
				seuilVWMA = (seuilAbsolu / vwma6[i]) * 100
			} else {
				seuilVWMA = config.SeuilPenteVWMA // Fallback
			}
		} else {
			// Seuil fixe
			seuilVWMA = config.SeuilPenteVWMA
		}

		// Déterminer sens VWMA6
		if etat.VariationVWMA > seuilVWMA {
			etat.SensVWMA6 = "CROISSANT"
		} else if etat.VariationVWMA < -seuilVWMA {
			etat.SensVWMA6 = "DECROISSANT"
		} else {
			etat.SensVWMA6 = "STABLE"
		}

		etats = append(etats, etat)
	}

	return etats
}

// ============================================================================
// REGROUPEMENT INTERVALLES
// ============================================================================

func groupIntervallesVWMA6(etats []EtatDirection, klines []gateio.Kline) []Intervalle {
	if len(etats) == 0 {
		return []Intervalle{}
	}

	intervalles := []Intervalle{}

	// Trouver le premier sens non-STABLE
	var currentSens string
	indexDebut := 0
	for i := 0; i < len(etats); i++ {
		if etats[i].SensVWMA6 != "STABLE" {
			currentSens = etats[i].SensVWMA6
			indexDebut = i
			break
		}
	}

	if currentSens == "" {
		// Tous les états sont STABLE
		return intervalles
	}

	dateDebut := etats[indexDebut].Timestamp
	sumVWMA6 := 0.0
	count := 0

	// Buffer pour confirmer changement de sens
	bufferCount := 0
	candidatSens := ""
	bufferStartIdx := 0

	for i := indexDebut; i < len(etats); i++ {
		sens := etats[i].SensVWMA6

		// Ajouter à la somme de l'intervalle courant
		sumVWMA6 += etats[i].VWMA6
		count++

		if sens == "STABLE" {
			// STABLE n'interrompt pas l'intervalle
			continue
		}

		if sens == currentSens {
			// Même sens : réinitialiser le buffer
			bufferCount = 0
			candidatSens = ""
		} else {
			// Sens opposé
			if bufferCount == 0 {
				// Début d'un potentiel changement
				candidatSens = sens
				bufferCount = 1
				bufferStartIdx = i
			} else if sens == candidatSens {
				// Confirmation du changement
				bufferCount++

				if bufferCount >= K_CONFIRMATION {
					// Changement confirmé : fermer l'intervalle
					finIdx := bufferStartIdx - 1

					// Retirer les bougies du buffer de la somme
					for j := bufferStartIdx; j <= i; j++ {
						sumVWMA6 -= etats[j].VWMA6
						count--
					}

					if count > 0 {
						prixDebut := klines[etats[indexDebut].Index].Close
						prixFin := klines[etats[finIdx].Index].Close
						variationCaptee := 0.0
						if prixDebut != 0 {
							variationCaptee = (prixFin - prixDebut) / prixDebut * 100
						}

						intervalle := Intervalle{
							Numero:          len(intervalles) + 1,
							Sens:            currentSens,
							IndexDebut:      indexDebut,
							IndexFin:        etats[finIdx].Index,
							DateDebut:       dateDebut,
							DateFin:         etats[finIdx].Timestamp,
							NbBougies:       count,
							ValeurMoyenne:   sumVWMA6 / float64(count),
							PrixDebut:       prixDebut,
							PrixFin:         prixFin,
							VariationCaptee: variationCaptee,
						}
						intervalles = append(intervalles, intervalle)
					}

					// Nouveau intervalle avec le nouveau sens
					currentSens = candidatSens
					indexDebut = bufferStartIdx
					dateDebut = etats[bufferStartIdx].Timestamp
					sumVWMA6 = 0.0
					count = 0
					for j := bufferStartIdx; j <= i; j++ {
						sumVWMA6 += etats[j].VWMA6
						count++
					}

					bufferCount = 0
					candidatSens = ""
				}
			} else {
				// Retour au sens original : réinitialiser
				bufferCount = 0
				candidatSens = ""
			}
		}
	}

	// Fermer le dernier intervalle
	if count > 0 {
		prixDebut := klines[etats[indexDebut].Index].Close
		prixFin := klines[etats[len(etats)-1].Index].Close
		variationCaptee := 0.0
		if prixDebut != 0 {
			variationCaptee = (prixFin - prixDebut) / prixDebut * 100
		}

		intervalle := Intervalle{
			Numero:          len(intervalles) + 1,
			Sens:            currentSens,
			IndexDebut:      indexDebut,
			IndexFin:        etats[len(etats)-1].Index,
			DateDebut:       dateDebut,
			DateFin:         etats[len(etats)-1].Timestamp,
			NbBougies:       count,
			ValeurMoyenne:   sumVWMA6 / float64(count),
			PrixDebut:       prixDebut,
			PrixFin:         prixFin,
			VariationCaptee: variationCaptee,
		}
		intervalles = append(intervalles, intervalle)
	}

	return intervalles
}

// ============================================================================
// AFFICHAGE
// ============================================================================

func displayEtats(etats []EtatDirection, limit int) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("ETATS DIRECTIONNELS (Dernières bougies) - Pour calibrage")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-19s | %-10s | %-12s | %-12s\n",
		"Date/Heure", "VWMA6", "Var VWMA%", "Sens VWMA6")
	fmt.Println(strings.Repeat("-", 100))

	start := len(etats) - limit
	if start < 0 {
		start = 0
	}

	for i := start; i < len(etats); i++ {
		e := etats[i]

		sensVWMASymbol := "→"
		if e.SensVWMA6 == "CROISSANT" {
			sensVWMASymbol = "↗"
		} else if e.SensVWMA6 == "DECROISSANT" {
			sensVWMASymbol = "↘"
		}

		fmt.Printf("%s | %10.2f | %+11.2f%% | %-12s\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.VWMA6, e.VariationVWMA, sensVWMASymbol)
	}

	fmt.Println(strings.Repeat("=", 100))
	if USE_DYNAMIC_THRESHOLD {
		fmt.Printf("\nSeuil: DYNAMIQUE (ATR × %.2f) | Période pente: %d bougies\n", float64(ATR_COEFFICIENT), PERIODE_PENTE)
	} else {
		fmt.Printf("\nSeuil: FIXE %.2f%% | Période pente: %d bougies\n", SEUIL_PENTE_VWMA, PERIODE_PENTE)
	}
}

func displayIntervalles(intervalles []Intervalle, titre string) {
	fmt.Println("\n" + strings.Repeat("=", 140))
	fmt.Println(titre)
	fmt.Println(strings.Repeat("=", 140))
	fmt.Printf("%-4s | %-9s | %-14s | %-19s | %-19s | %-8s | %-12s | %-12s\n",
		"#", "Var %", "Sens", "Date Début (Open)", "Date Fin (Open)", "Bougies", "Prix Début", "Prix Fin")
	fmt.Println(strings.Repeat("-", 140))

	for _, inter := range intervalles {
		sensDisplay := inter.Sens
		if inter.Sens == "CROISSANT" {
			sensDisplay = "↗ CROISSANT"
		} else if inter.Sens == "DECROISSANT" {
			sensDisplay = "↘ DÉCROISSANT"
		} else if inter.Sens == "STABLE" {
			sensDisplay = "→ STABLE"
		}

		fmt.Printf("%-4d | %+8.2f%% | %-14s | %-19s | %-19s | %8d | %12.2f | %12.2f\n",
			inter.Numero,
			inter.VariationCaptee,
			sensDisplay,
			inter.DateDebut.Format("2006-01-02 15:04:05"),
			inter.DateFin.Format("2006-01-02 15:04:05"),
			inter.NbBougies,
			inter.PrixDebut,
			inter.PrixFin)
	}

	fmt.Println(strings.Repeat("=", 140))
}

func displayStatisticsIntervalles(intervallesVWMA6 []Intervalle) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("STATISTIQUES INTERVALLES VWMA6")
	fmt.Println(strings.Repeat("=", 100))

	// Stats VWMA6
	totalVWMACroissant := 0
	totalVWMADecroissant := 0
	totalVWMAStable := 0
	bougiesVWMACroissant := 0
	bougiesVWMADecroissant := 0
	bougiesVWMAStable := 0
	variationTotaleCroissant := 0.0
	variationTotaleDecroissant := 0.0

	for _, inter := range intervallesVWMA6 {
		switch inter.Sens {
		case "CROISSANT":
			totalVWMACroissant++
			bougiesVWMACroissant += inter.NbBougies
			variationTotaleCroissant += inter.VariationCaptee
		case "DECROISSANT":
			totalVWMADecroissant++
			bougiesVWMADecroissant += inter.NbBougies
			variationTotaleDecroissant += inter.VariationCaptee
		case "STABLE":
			totalVWMAStable++
			bougiesVWMAStable += inter.NbBougies
		}
	}

	fmt.Println("\nINTERVALLES VWMA6:")
	fmt.Printf("  Total intervalles    : %d\n", len(intervallesVWMA6))
	fmt.Printf("  - Croissant (↗)      : %d intervalles (%d bougies)\n", totalVWMACroissant, bougiesVWMACroissant)
	fmt.Printf("  - Décroissant (↘)    : %d intervalles (%d bougies)\n", totalVWMADecroissant, bougiesVWMADecroissant)
	fmt.Printf("  - Stable (→)         : %d intervalles (%d bougies)\n", totalVWMAStable, bougiesVWMAStable)

	fmt.Println("\nVARIATIONS CAPTÉES:")
	if totalVWMACroissant > 0 {
		fmt.Printf("  - Croissant (↗)      : %+.2f%% total, %+.2f%% moyen par intervalle\n",
			variationTotaleCroissant, variationTotaleCroissant/float64(totalVWMACroissant))
	}
	if totalVWMADecroissant > 0 {
		fmt.Printf("  - Décroissant (↘)    : %+.2f%% total, %+.2f%% moyen par intervalle\n",
			variationTotaleDecroissant, variationTotaleDecroissant/float64(totalVWMADecroissant))
	}

	// Total capté = CROISSANT + (DÉCROISSANT × -1)
	// Les variations DÉCROISSANT profitables sont négatives, donc on inverse
	totalCapte := variationTotaleCroissant - variationTotaleDecroissant

	fmt.Printf("  - TOTAL CAPTÉ        : %.2f%% (bidirectionnel LONG+SHORT)\n",
		totalCapte)

	fmt.Println(strings.Repeat("=", 100))
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	config := Config{
		Symbol:              SYMBOL,
		Timeframe:           TIMEFRAME,
		NbCandles:           NB_CANDLES,
		VwmaRapide:          VWMA_RAPIDE,
		PeriodePente:        PERIODE_PENTE,
		SeuilPenteVWMA:      SEUIL_PENTE_VWMA,
		KConfirmation:       K_CONFIRMATION,
		UseDynamicThreshold: USE_DYNAMIC_THRESHOLD,
		AtrPeriode:          ATR_PERIODE,
		AtrCoefficient:      ATR_COEFFICIENT,
	}

	fmt.Println("=== DEMO ANALYSE DIRECTIONNELLE (VWMA6 uniquement) ===")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("Exchange           : Gate.io\n")
	fmt.Printf("Symbole            : %s\n", config.Symbol)
	fmt.Printf("Timeframe          : %s\n", config.Timeframe)
	fmt.Printf("VWMA rapide        : %d\n", config.VwmaRapide)
	fmt.Printf("Periode pente      : %d bougies\n", config.PeriodePente)
	if config.UseDynamicThreshold {
		fmt.Printf("Seuil pente        : DYNAMIQUE (ATR × %.2f)\n", config.AtrCoefficient)
		fmt.Printf("ATR periode        : %d\n", config.AtrPeriode)
	} else {
		fmt.Printf("Seuil pente VWMA6  : %.2f%% (fixe)\n", config.SeuilPenteVWMA)
	}
	fmt.Printf("K confirmation     : %d bougies\n", config.KConfirmation)

	// Récupération des données
	fmt.Printf("\nRecuperation des klines depuis Gate.io...\n")
	ctx := context.Background()
	client := gateio.NewClient()

	klines, err := client.GetKlines(ctx, config.Symbol, config.Timeframe, config.NbCandles)
	if err != nil {
		log.Fatalf("Erreur recuperation klines: %v", err)
	}

	// Tri chronologique
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].OpenTime.Before(klines[j].OpenTime)
	})

	fmt.Printf("Klines recuperees: %d\n", len(klines))

	// Préparation données
	fmt.Println("\nCalcul des indicateurs...")

	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	volumes := make([]float64, len(klines))

	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
		volumes[i] = k.Volume
	}

	// Calcul VWMA6
	vwmaRapideIndicator := indicators.NewVWMATVStandard(config.VwmaRapide)
	vwma6 := vwmaRapideIndicator.Calculate(closes, volumes)

	// Calcul ATR (si mode dynamique)
	var atr []float64
	if config.UseDynamicThreshold {
		atrIndicator := indicators.NewATRTVStandard(config.AtrPeriode)
		atr = atrIndicator.Calculate(highs, lows, closes)
	}

	// Analyse direction
	fmt.Println("Analyse des directions...")
	etats := analyzeDirection(config, klines, vwma6, atr)

	fmt.Println("Regroupement en intervalles...")
	intervallesVWMA6 := groupIntervallesVWMA6(etats, klines)

	// Affichage
	displayEtats(etats, 30) // Afficher les 30 dernières bougies pour calibrage
	displayIntervalles(intervallesVWMA6, "INTERVALLES VWMA6 (Direction Prix)")
	displayStatisticsIntervalles(intervallesVWMA6)

	fmt.Println("\n=== FIN DEMO ===")
}
