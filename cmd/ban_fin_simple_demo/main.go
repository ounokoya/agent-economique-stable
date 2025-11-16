package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"agent-economique/internal/baranalysis"
	"agent-economique/internal/datasource/bybit"
	"agent-economique/internal/indicators"
	"agent-economique/internal/setups/dmiopen"
    "agent-economique/internal/setups/dmiresp"
)

const (
	DEMO_SYMBOL   = "SOLUSDT"
	DEMO_INTERVAL = "5m" // 1m | 5m | 15m | 30m | 1h | 2h | 4h | 1d
	DEMO_LIMIT    = 1000 // nombre de dernières bougies à récupérer

	// Indicateur périodes (constantes de démo)
	VWMA_P1    = 10
	VWMA_P2    = 60
	VWMA_P3    = 240
	DMI_PERIOD = 30
	ADX_PERIOD = 4 // période d'ADX distincte de DMI
	MFI_PERIOD = 14
	ATR_PERIOD = 10

	// Option d'activation du log de base (lignes d'indicateurs [INDIC])
	BASE_LOG_ENABLED = false

	// Filtre du "premier signal de base": n'imprimer l'analyse de barre
	// que si Body/ATR dépasse ce seuil
	BASE_FILTER_ENABLED      = false
	BASE_FILTER_MIN_BODY_ATR = 1.0

	// Filtre 2: lister les bougies où
	//  - ratio bar (Body/ATR) > 0
	//  - |ratio 3B (avg signed body/atr)| > 0
	//  - signe ratio 3B cohérent avec le type de la bougie (VERT=>+, ROUGE=>-)
	FILTER2_ENABLED      = false
	FILTER2_MIN_BODY_ATR = 0.4

	// Options d'affichage par filtre: si false, n'imprime que l'heure d'ouverture
	BASE_FILTER_LOG_VERBOSE = false
	FILTER2_LOG_VERBOSE     = false

	// Filtre 3 (indépendant): croisement VWMA(5) et VWMA(10)
	FILTER_VWMA_ENABLED     = false
	FILTER_VWMA_LOG_VERBOSE = false
	FILTER_VWMA_P1          = 60
	FILTER_VWMA_P2          = 240
	FILTER_VWMA_CROSS_TYPE  = "ANY" // ANY | GOLDEN | DEAD

	// Setup DMI: croisement ouverture tendance (LONG/SHORT)
	DMI_OPEN_ENABLED     = true
	DMI_OPEN_LOG_VERBOSE = false
	// Toggles d'activation des seuils
	DMI_OPEN_USE_MIN_DX  = true
	DMI_OPEN_USE_MIN_ADX = true
	DMI_OPEN_USE_MAX_ADX = true
	// Toggles d'activation des règles relatives
	DMI_OPEN_USE_REL_ADX_UNDER_DI_SUP = true
	DMI_OPEN_USE_REL_DX_OVER_DI_INF   = false
	// Filtre de distance minimale DX/ADX
	DMI_OPEN_USE_MIN_DX_ADX_GAP         = true
	DMI_OPEN_MIN_DX_ADX_GAP     float64 = 5.0
	// Re-vérification 1 bougie si invalidé à i
	DMI_OPEN_USE_ONE_BAR_RECHECK = true
	// Re-vérification 2 et 3 bougies (indépendantes)
	DMI_OPEN_USE_TWO_BAR_RECHECK   = true
	DMI_OPEN_USE_THREE_BAR_RECHECK = true
	// Re-vérification 4,5,6 bougies (indépendantes)
	DMI_OPEN_USE_FOUR_BAR_RECHECK         = true
	DMI_OPEN_USE_FIVE_BAR_RECHECK         = true
	DMI_OPEN_USE_SIX_BAR_RECHECK          = false
	DMI_OPEN_MIN_DX               float64 = 10.0
	DMI_OPEN_MIN_ADX              float64 = 10.0
	DMI_OPEN_MAX_ADX              float64 = 25.0

	// Setup DMI: respiration (confirmation / retour sous ADX)
	DMI_RESP_ENABLED      = false
	DMI_RESP_LOG_VERBOSE  = false
	DMI_RESP_USE_MIN_DX   = false
	DMI_RESP_MIN_DX float64 = 0.0
	DMI_RESP_USE_MIN_ADX  = false
	DMI_RESP_MIN_ADX float64 = 0.0
	DMI_RESP_USE_MAX_ADX  = true
	DMI_RESP_MAX_ADX float64 = 25.0
	DMI_RESP_GAP    float64 = 5.0 // seuil_resp
	DMI_RESP_USE_CONFIRM1 = true  // attendre i+1 pour cas C
	DMI_RESP_REQUIRE_DX_DOWN = true
	DMI_RESP_REQUIRE_ADX_DOWN = true
	DMI_RESP_EPS float64 = 0
)

const (
	tagAnalyse    = "\x1b[36m[ANALYSE]\x1b[0m"
	tagAnalyse3B  = "\x1b[35m[ANALYSE_3B]\x1b[0m"
	tagIndic      = "\x1b[34m[INDIC]\x1b[0m"
	tagBaseSig    = "\x1b[32m[BASE_SIGNAL]\x1b[0m"
	tagFilter2    = "\x1b[96m[FILTER2]\x1b[0m"
	tagFilterVWMA = "\x1b[95m[FILTER_VWMA]\x1b[0m"
	tagDmiOpen    = "\x1b[93m[DMI_OPEN]\x1b[0m"
	tagDmiResp    = "\x1b[92m[DMI_RESP]\x1b[0m"
	tagSummary    = "\x1b[33m[SUMMARY]\x1b[0m"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1) Fetch N dernières klines Bybit
	client := bybit.NewClient()
	kl, err := client.GetKlines(ctx, DEMO_SYMBOL, DEMO_INTERVAL, DEMO_LIMIT)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bybit get klines error: %v\n", err)
		os.Exit(1)
	}
	if len(kl) == 0 {
		fmt.Fprintln(os.Stderr, "no klines returned")
		os.Exit(1)
	}

	// 2) Pré-traitements: trier, dédupliquer, exclure bougies incomplètes
	processed := preprocessKlines(kl)
	if len(processed) == 0 {
		fmt.Fprintf(os.Stdout, "%s 0 klines after preprocessing\n", tagSummary)
		return
	}

	// 3) Construire séries de base
	n := len(processed)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	volumes := make([]float64, n)
	for i := 0; i < n; i++ {
		highs[i] = processed[i].High
		lows[i] = processed[i].Low
		closes[i] = processed[i].Close
		volumes[i] = processed[i].Volume
	}

	// 4) Calculer indicateurs (VWMA 5/10/50, DMI(+/−,ADX), DX, MFI)
	vwma5 := indicators.NewVWMATVStandard(VWMA_P1).Calculate(closes, volumes)
	vwma10 := indicators.NewVWMATVStandard(VWMA_P2).Calculate(closes, volumes)
	vwma50 := indicators.NewVWMATVStandard(VWMA_P3).Calculate(closes, volumes)

	// VWMA dynamiques pour le filtre de croisement (périodes configurables)
	vwmaP1Vals := indicators.NewVWMATVStandard(FILTER_VWMA_P1).Calculate(closes, volumes)
	vwmaP2Vals := indicators.NewVWMATVStandard(FILTER_VWMA_P2).Calculate(closes, volumes)

	// DMI/ADX avec périodes distinctes: DI avec DMI_PERIOD, ADX avec ADX_PERIOD
	dmiInd := indicators.NewDMITVStandardWithPeriods(DMI_PERIOD, ADX_PERIOD)
	diPlus, diMinus, adx := dmiInd.Calculate(highs, lows, closes)
	dx := dmiInd.CalculateDX(diPlus, diMinus)

	mfi := indicators.NewMFITVStandard(MFI_PERIOD).Calculate(highs, lows, closes, volumes)

	atr10 := indicators.NewATRTVStandard(ATR_PERIOD).Calculate(highs, lows, closes)

	// Configurer le module réutilisable DMI_OPEN
	cfgOpen := dmiopen.Config{
		UseRelAdxUnderDiSup:         DMI_OPEN_USE_REL_ADX_UNDER_DI_SUP,
		UseRelDxOverDiInf:           DMI_OPEN_USE_REL_DX_OVER_DI_INF,
		UseMinDX:                    DMI_OPEN_USE_MIN_DX,
		MinDX:                       DMI_OPEN_MIN_DX,
		UseMinADX:                   DMI_OPEN_USE_MIN_ADX,
		MinADX:                      DMI_OPEN_MIN_ADX,
		UseMaxADX:                   DMI_OPEN_USE_MAX_ADX,
		MaxADX:                      DMI_OPEN_MAX_ADX,
		UseMinGapDXADX:              DMI_OPEN_USE_MIN_DX_ADX_GAP,
		MinGapDXADX:                 DMI_OPEN_MIN_DX_ADX_GAP,
		UseRecheck1:                 DMI_OPEN_USE_ONE_BAR_RECHECK,
		UseRecheck2:                 DMI_OPEN_USE_TWO_BAR_RECHECK,
		UseRecheck3:                 DMI_OPEN_USE_THREE_BAR_RECHECK,
		UseRecheck4:                 DMI_OPEN_USE_FOUR_BAR_RECHECK,
		UseRecheck5:                 DMI_OPEN_USE_FIVE_BAR_RECHECK,
		UseRecheck6:                 DMI_OPEN_USE_SIX_BAR_RECHECK,
		RecheckRequireContextSide:   true,
		RecheckRequireDXAboveADX:    true,
		RecheckCancelSiblingsOnFlip: true,
		RequireDXUp:                 true,
		RequireADXUp:                true,
		Eps:                         0,
		LogLevel:                    0,
	}
	modOpen := dmiopen.New(cfgOpen)

	// Configurer le module réutilisable DMI_RESP
	cfgResp := dmiresp.Config{
		UseMinDX:     DMI_RESP_USE_MIN_DX,
		MinDX:        DMI_RESP_MIN_DX,
		UseMinADX:    DMI_RESP_USE_MIN_ADX,
		MinADX:       DMI_RESP_MIN_ADX,
		UseMaxADX:    DMI_RESP_USE_MAX_ADX,
		MaxADX:       DMI_RESP_MAX_ADX,
		RespGap:      DMI_RESP_GAP,
		UseConfirm1:  DMI_RESP_USE_CONFIRM1,
		RequireDXDown: DMI_RESP_REQUIRE_DX_DOWN,
		RequireADXDown: DMI_RESP_REQUIRE_ADX_DOWN,
		Eps:           DMI_RESP_EPS,
		LogLevel:      0,
	}
	modResp := dmiresp.New(cfgResp)

	// 5) Parcourir et logger à chaque date d'ouverture
	filteredAnalyses := make([]baranalysis.BarAnalysis, 0, 3)
	allAnalyses := make([]baranalysis.BarAnalysis, 0, 3)
	for i := 0; i < n; i++ {
		k := processed[i]

		// Analyse de barre et log dédié
		analysis := baranalysis.AnalyzeBar(k.Open, k.High, k.Low, k.Close, atr10[i], k.OpenTime)

		// Maintenir une fenêtre 3B non filtrée (pour FILTER2)
		allAnalyses = append(allAnalyses, analysis)
		if len(allAnalyses) > 3 {
			allAnalyses = allAnalyses[1:]
		}

		// Préparer les flags de passage des filtres
		barRatio := analysis.BodyATRRatio
		threeBRatio := math.NaN()
		passF2 := false
		if FILTER2_ENABLED && len(allAnalyses) == 3 {
			aggAll := baranalysis.AggregateAnalyses(allAnalyses)
			threeBRatio = aggAll.AvgSignedBodyATRRatio
			if !math.IsNaN(barRatio) && barRatio > 0 && !math.IsNaN(threeBRatio) && math.Abs(threeBRatio) > FILTER2_MIN_BODY_ATR {
				sameSign := (analysis.Type == "VERT" && threeBRatio > 0) || (analysis.Type == "ROUGE" && threeBRatio < 0)
				if sameSign {
					passF2 = true
				}
			}
		}

		// Filtre VWMA(P1/P2) indépendant: détection de croisement edge-triggered (i-1 -> i)
		passVWMA := false
		crossType := ""
		if FILTER_VWMA_ENABLED && i > 0 {
			prevP1, prevP2 := vwmaP1Vals[i-1], vwmaP2Vals[i-1]
			curP1, curP2 := vwmaP1Vals[i], vwmaP2Vals[i]
			if !math.IsNaN(prevP1) && !math.IsNaN(prevP2) && !math.IsNaN(curP1) && !math.IsNaN(curP2) {
				golden := prevP1 <= prevP2 && curP1 > curP2
				dead := prevP1 >= prevP2 && curP1 < curP2
				if golden {
					crossType = "GOLDEN"
				} else if dead {
					crossType = "DEAD"
				}
				if crossType != "" {
					switch FILTER_VWMA_CROSS_TYPE {
					case "GOLDEN":
						passVWMA = crossType == "GOLDEN"
					case "DEAD":
						passVWMA = crossType == "DEAD"
					default: // ANY
						passVWMA = golden || dead
					}
				}
			}
		}

		// Setup DMI "ouverture tendance" (LONG/SHORT)
		passDmiOpen := false
		dmiOpenSide := ""
		dmiOpenLag := -1
		if DMI_OPEN_ENABLED && i > 0 {
			evt, ok := modOpen.Step(dmiopen.Inputs{
				Index:   i,
				DIpPrev: diPlus[i-1], DImPrev: diMinus[i-1],
				DXPrev: dx[i-1], ADXPrev: adx[i-1],
				DIp: diPlus[i], DIm: diMinus[i], DX: dx[i], ADX: adx[i],
			})
			if ok {
				passDmiOpen = true
				dmiOpenSide = string(evt.Side)
				dmiOpenLag = evt.Lag
			}
		}

		// Setup DMI "respiration tendance" (LONG/SHORT)
		passDmiResp := false
		dmiRespSide := ""
		dmiRespLag := -1
		if DMI_RESP_ENABLED && i > 0 {
			evt, ok := modResp.Step(dmiresp.Inputs{
				Index:   i,
				DIpPrev: diPlus[i-1], DImPrev: diMinus[i-1],
				DXPrev:  dx[i-1], ADXPrev: adx[i-1],
				DIp:     diPlus[i], DIm: diMinus[i], DX: dx[i], ADX: adx[i],
			})
			if ok {
				passDmiResp = true
				dmiRespSide = string(evt.Side)
				dmiRespLag = evt.Lag
			}
		}

		passBase := false
		if BASE_FILTER_ENABLED {
			passBase = !math.IsNaN(analysis.BodyATRRatio) && analysis.BodyATRRatio > BASE_FILTER_MIN_BODY_ATR
		}

		allowed := BASE_LOG_ENABLED || passBase || passF2 || passVWMA || passDmiOpen || passDmiResp

		if allowed {
			// Détail si le master est ON ou si un filtre actif a sa verbosité ON; sinon, heure d'open uniquement
			detailRequested := BASE_LOG_ENABLED || (passBase && BASE_FILTER_LOG_VERBOSE) || (passF2 && FILTER2_LOG_VERBOSE) || (passVWMA && FILTER_VWMA_LOG_VERBOSE) || (passDmiOpen && DMI_OPEN_LOG_VERBOSE) || (passDmiResp && DMI_RESP_LOG_VERBOSE)
			if detailRequested {
				// INDIC pour la bougie courante
				fmt.Printf(
					"%s %s close=%.6f vwma5=%.6f vwma10=%.6f vwma50=%.6f di+=%.6f di-=%.6f dx=%.6f adx=%.6f mfi=%.6f atr10=%.6f\n",
					tagIndic, k.OpenTime.Format(time.RFC3339), k.Close,
					vwma5[i], vwma10[i], vwma50[i],
					diPlus[i], diMinus[i], dx[i], adx[i], mfi[i], atr10[i],
				)

				// ANALYSE détaillée de la bougie courante
				fmt.Printf(
					"%s %s type=%s occ%%=%.2f atr10=%.6f body=%.6f non_body=%.6f body/atr=%.6f non_body/atr=%.6f\n",
					tagAnalyse,
					analysis.Time.Format(time.RFC3339), analysis.Type, analysis.OccupationPercent,
					analysis.ATR, analysis.Body, analysis.NonBody, analysis.BodyATRRatio, analysis.NonBodyATRRatio,
				)

				// Base signal: log only bars with body/atr > 1
				if analysis.BodyATRRatio > 1.0 {
					fmt.Printf(
						"%s %s type=%s close=%.6f body/atr=%.6f body=%.6f atr10=%.6f\n",
						tagBaseSig,
						k.OpenTime.Format(time.RFC3339), analysis.Type, k.Close, analysis.BodyATRRatio, analysis.Body, atr10[i],
					)
				}

				// Fenêtre 3 barres des bougies effectivement loggées
				filteredAnalyses = append(filteredAnalyses, analysis)
				if len(filteredAnalyses) > 3 {
					filteredAnalyses = filteredAnalyses[1:]
				}
				if passF2 && len(allAnalyses) == 3 {
					// Aligner l'analyse 3B avec la fenêtre utilisée par FILTER2 (3 dernières barres réelles)
					agg := baranalysis.AggregateAnalyses(allAnalyses)
					fmt.Printf(
						"%s %s count=%d verts=%d rouges=%d avg_occ%%=%.2f avg_atr=%.6f avg_signed_body=%.6f avg_signed_non_body=%.6f avg_signed_body/atr=%.6f avg_signed_non_body/atr=%.6f sum_body/atr=%.6f sum_non_body/atr=%.6f\n",
						tagAnalyse3B,
						analysis.Time.Format(time.RFC3339),
						agg.Count, agg.CountVert, agg.CountRouge,
						agg.AvgOccupationPct, agg.AvgATR,
						agg.AvgSignedBody, agg.AvgSignedNonBody,
						agg.AvgSignedBodyATRRatio, agg.AvgSignedNonBodyATRRatio,
						agg.SumBodyATRRatio, agg.SumNonBodyATRRatio,
					)
				} else if len(filteredAnalyses) == 3 {
					agg := baranalysis.AggregateAnalyses(filteredAnalyses)
					fmt.Printf(
						"%s %s count=%d verts=%d rouges=%d avg_occ%%=%.2f avg_atr=%.6f avg_signed_body=%.6f avg_signed_non_body=%.6f avg_signed_body/atr=%.6f avg_signed_non_body/atr=%.6f sum_body/atr=%.6f sum_non_body/atr=%.6f\n",
						tagAnalyse3B,
						analysis.Time.Format(time.RFC3339),
						agg.Count, agg.CountVert, agg.CountRouge,
						agg.AvgOccupationPct, agg.AvgATR,
						agg.AvgSignedBody, agg.AvgSignedNonBody,
						agg.AvgSignedBodyATRRatio, agg.AvgSignedNonBodyATRRatio,
						agg.SumBodyATRRatio, agg.SumNonBodyATRRatio,
					)
				}

				// Log spécifique FILTER2 si déclenché et verbosité active
				if passF2 && FILTER2_LOG_VERBOSE {
					fmt.Printf(
						"%s %s type=%s close=%.6f body/atr=%.6f 3b_avg_signed_body/atr=%.6f\n",
						tagFilter2,
						k.OpenTime.Format(time.RFC3339), analysis.Type, k.Close, barRatio, threeBRatio,
					)
				}

				// Log spécifique FILTER_VWMA si déclenché et verbosité active
				if passVWMA && FILTER_VWMA_LOG_VERBOSE {
					fmt.Printf(
						"%s %s cross=%s vwmaP1=%.6f vwmaP2=%.6f p1=%d p2=%d\n",
						tagFilterVWMA,
						k.OpenTime.Format(time.RFC3339), crossType, vwmaP1Vals[i], vwmaP2Vals[i], FILTER_VWMA_P1, FILTER_VWMA_P2,
					)
				}

				// Log spécifique DMI_OPEN si déclenché et verbosité active
				if passDmiOpen && DMI_OPEN_LOG_VERBOSE {
					fmt.Printf(
						"%s %s side=%s lag=%d di+=%.6f di-=%.6f dx=%.6f adx=%.6f gap=%.6f min_dx=%.2f min_adx=%.2f max_adx=%.2f min_gap=%.2f\n",
						tagDmiOpen,
						k.OpenTime.Format(time.RFC3339), dmiOpenSide, dmiOpenLag, diPlus[i], diMinus[i], dx[i], adx[i], math.Abs(dx[i]-adx[i]), DMI_OPEN_MIN_DX, DMI_OPEN_MIN_ADX, DMI_OPEN_MAX_ADX, DMI_OPEN_MIN_DX_ADX_GAP,
					)
				}

				// Log spécifique DMI_RESP si déclenché et verbosité active
				if passDmiResp && DMI_RESP_LOG_VERBOSE {
					fmt.Printf(
						"%s %s side=%s lag=%d di+=%.6f di-=%.6f dx=%.6f adx=%.6f gap=%.6f min_dx=%.2f min_adx=%.2f max_adx=%.2f resp_gap=%.2f\n",
						tagDmiResp,
						k.OpenTime.Format(time.RFC3339), dmiRespSide, dmiRespLag, diPlus[i], diMinus[i], dx[i], adx[i], math.Abs(dx[i]-adx[i]), DMI_RESP_MIN_DX, DMI_RESP_MIN_ADX, DMI_RESP_MAX_ADX, DMI_RESP_GAP,
					)
				}
			} else {
				// Verbosité désactivée pour les filtres actifs: imprimer uniquement l'heure d'ouverture
				// Orientation priorisée par le filtre déclencheur, sinon signe du ratio 3B
				orien := ""
				// 1) Si DMI_OPEN a déclenché, utiliser son côté
				if passDmiOpen && dmiOpenSide != "" {
					orien = dmiOpenSide
				} else if passDmiResp && dmiRespSide != "" { // 2) Sinon DMI_RESP
					orien = dmiRespSide
				} else if passVWMA && crossType != "" { // 3) Sinon croisement VWMA: GOLDEN=LONG, DEAD=SHORT
					if crossType == "GOLDEN" {
						orien = "LONG"
					} else if crossType == "DEAD" {
						orien = "SHORT"
					}
				} else if passF2 { // 4) Sinon Filter2: utiliser le signe du ratio 3B
					ratio := math.NaN()
					if len(allAnalyses) == 3 {
						aggTmp := baranalysis.AggregateAnalyses(allAnalyses)
						ratio = aggTmp.AvgSignedBodyATRRatio
					} else if !math.IsNaN(threeBRatio) {
						ratio = threeBRatio
					}
					if !math.IsNaN(ratio) {
						if ratio > 0 {
							orien = "LONG"
						} else if ratio < 0 {
							orien = "SHORT"
						}
					}
				}
				if orien != "" {
					fmt.Printf("%s %s\n", k.OpenTime.Format(time.RFC3339), orien)
				} else {
					fmt.Println(k.OpenTime.Format(time.RFC3339))
				}
			}
		}

	}

	fmt.Fprintf(os.Stdout, "%s printed %d kline rows with indicators (symbol=%s, interval=%s)\n", tagSummary, n, DEMO_SYMBOL, DEMO_INTERVAL)
}

// preprocessKlines trie, déduplique par OpenTime, et exclut les bougies incomplètes
func preprocessKlines(in []bybit.Kline) []bybit.Kline {
	if len(in) == 0 {
		return in
	}

	// Trier par OpenTime asc (défensif même si déjà trié)
	cp := make([]bybit.Kline, len(in))
	copy(cp, in)
	sort.Slice(cp, func(i, j int) bool { return cp[i].OpenTime.Before(cp[j].OpenTime) })

	now := time.Now()
	out := make([]bybit.Kline, 0, len(cp))
	var last time.Time
	for _, k := range cp {
		// Dédup par OpenTime exact
		if !last.IsZero() && k.OpenTime.Equal(last) {
			continue
		}
		// Exclure bougie incomplète (closeTime dans le futur)
		if k.CloseTime.After(now) {
			continue
		}
		out = append(out, k)
		last = k.OpenTime
	}
	return out
}

// Les fonctions d'analyse de bougie sont désormais dans internal/baranalysis
