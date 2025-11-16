package trend

import (
	"fmt"
	"time"

	"agent-economique/internal/execution"
	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

// TrendGenerator générateur de signaux basé sur VWMA + DMI
type TrendGenerator struct {
	config signals.GeneratorConfig

	// Paramètres spécifiques
	vwmaRapide          int
	vwmaLent            int
	dmiPeriode          int
	dmiSmooth           int
	atrPeriode          int
	gammaGapVWMA        float64
	gammaGapDI          float64
	gammaGapDX          float64
	volatiliteMin       float64
	windowGammaValidate int
	windowW             int

	// Filtres de bougie initiale
	bodyPctMin             float64
	bodyATRMin             float64
	enforceCandleDirection bool

	// État interne
	vwmaRapideValues []float64
	vwmaLentValues   []float64
	diPlus           []float64
	diMinus          []float64
	dx               []float64
	adx              []float64
	atr              []float64
	lastProcessedIdx int
	openPositions    map[int]*OpenPosition // index → position
	nextPositionID   int
	metrics          signals.GeneratorMetrics

	// Sorties indépendantes
	enableExitVWMA     bool
	enableExitTrailing bool
	trailingCoeff      float64
	trailingCap        float64
	trailings          map[int]*execution.Trailing // posID -> trailing
}

// OpenPosition représente une position ouverte en attente de sortie
type OpenPosition struct {
	ID         int
	EntryIndex int
	EntryTime  time.Time
	EntryPrice float64
	Type       signals.SignalType
}

// Config configuration spécifique Trend
type Config struct {
	VwmaRapide          int
	VwmaLent            int
	DmiPeriode          int
	DmiSmooth           int
	AtrPeriode          int
	GammaGapVWMA        float64
	GammaGapDI          float64
	GammaGapDX          float64
	VolatiliteMin       float64
	WindowGammaValidate int
	WindowW             int

	// Filtres de bougie initiale
	BodyPctMin             float64
	BodyATRMin             float64
	EnforceCandleDirection bool

	// Sorties indépendantes
	EnableExitVWMA      bool
	EnableExitTrailing  bool
	TrailingATRCoeff    float64
	TrailingCapPct      float64
}

// SignalVWMA signal VWMA détecté
type SignalVWMA struct {
	Index             int
	Timestamp         time.Time
	Direction         string
	Prix              float64
	Vwma6             float64
	Vwma24            float64
	Gap               float64
	AtrPct            float64
	GapValide         bool
	GapValideBougie   int
	VolatiliteOK      bool
	Valide            bool
}

// SignalDMI signal DMI détecté
type SignalDMI struct {
	Index                int
	Timestamp            time.Time
	Direction            string
	DiPlus               float64
	DiMinus              float64
	Dx                   float64
	Adx                  float64
	GapDI                float64
	GapDIValide          bool
	GapDIValideBougie    int
	GapDXADX             float64
	GapDXADXValide       bool
	GapDXADXValideBougie int
	Valide               bool
}

// SignalTrend signal trend matché
type SignalTrend struct {
	Index        int
	Timestamp    time.Time
	Direction    string
	Prix         float64
	Vwma6        float64
	DiPlus       float64
	DiMinus      float64
	DistanceBars int
	Motif        string
}

// NewTrendGenerator crée un nouveau générateur Trend
func NewTrendGenerator(config Config) *TrendGenerator {
	return &TrendGenerator{
		vwmaRapide:          config.VwmaRapide,
		vwmaLent:            config.VwmaLent,
		dmiPeriode:          config.DmiPeriode,
		dmiSmooth:           config.DmiSmooth,
		atrPeriode:          config.AtrPeriode,
		gammaGapVWMA:        config.GammaGapVWMA,
		gammaGapDI:          config.GammaGapDI,
		gammaGapDX:          config.GammaGapDX,
		volatiliteMin:       config.VolatiliteMin,
		windowGammaValidate: config.WindowGammaValidate,
		windowW:             config.WindowW,
		lastProcessedIdx:    -1,
		openPositions:       make(map[int]*OpenPosition),
		nextPositionID:      1,
		enableExitVWMA:      config.EnableExitVWMA,
		enableExitTrailing:  config.EnableExitTrailing,
		trailingCoeff:       func() float64 { if config.TrailingATRCoeff > 0 { return config.TrailingATRCoeff }; return 1.0 }(),
		trailingCap:         func() float64 { if config.TrailingCapPct > 0 { return config.TrailingCapPct }; return 0.003 }(),
		trailings:           make(map[int]*execution.Trailing),
		bodyPctMin:             config.BodyPctMin,
		bodyATRMin:             config.BodyATRMin,
		enforceCandleDirection: config.EnforceCandleDirection,
	}
}

// Name retourne le nom du générateur
func (g *TrendGenerator) Name() string {
	return "TrendVWMADMI"
}

// Initialize initialise le générateur
func (g *TrendGenerator) Initialize(config signals.GeneratorConfig) error {
	g.config = config
	return nil
}

// CalculateIndicators calcule VWMA, ATR et DMI
func (g *TrendGenerator) CalculateIndicators(klines []signals.Kline) error {
	if len(klines) < g.vwmaLent {
		return fmt.Errorf("pas assez de klines: %d < %d", len(klines), g.vwmaLent)
	}

	// Tri chronologique (comme autres générateurs)
	for i := 0; i < len(klines); i++ {
		for j := i + 1; j < len(klines); j++ {
			if klines[j].OpenTime.Before(klines[i].OpenTime) {
				klines[i], klines[j] = klines[j], klines[i]
			}
		}
	}

	// Extraire données
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

	// VWMA
	vwmaRapideInd := indicators.NewVWMATVStandard(g.vwmaRapide)
	vwmaLentInd := indicators.NewVWMATVStandard(g.vwmaLent)
	g.vwmaRapideValues = vwmaRapideInd.Calculate(closes, volumes)
	g.vwmaLentValues = vwmaLentInd.Calculate(closes, volumes)

	// ATR
	atrInd := indicators.NewATRTVStandard(g.atrPeriode)
	g.atr = atrInd.Calculate(highs, lows, closes)

	// DMI
	dmiInd := indicators.NewDMITVStandard(g.dmiPeriode)
	g.diPlus, g.diMinus, g.adx = dmiInd.Calculate(highs, lows, closes)

	// DX
	g.dx = make([]float64, len(g.diPlus))
	for i := range g.diPlus {
		sum := g.diPlus[i] + g.diMinus[i]
		if sum != 0 {
			g.dx[i] = indicators.CalculerEcart(g.diPlus[i], g.diMinus[i]) / sum * 100
		}
	}

	return nil
}

// DetectSignals détecte signaux ENTRY (matching VWMA+DMI) et EXIT (croisement inverse VWMA)
func (g *TrendGenerator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var newSignals []signals.Signal

	// IMPORTANT: len(klines)-1 = bougie EN COURS (ignorer)
	//            len(klines)-2 = dernière bougie FERMÉE
	lastClosedIdx := len(klines) - 2
	if lastClosedIdx < 0 {
		return newSignals, nil
	}

	// Mode demo: traiter TOUTES les klines si c'est le premier appel
	startIdx := g.lastProcessedIdx + 1
	if g.lastProcessedIdx == -1 {
		// Premier appel : traiter toutes les klines historiques
		startIdx = g.vwmaLent + 5
	}
	
	if startIdx < g.vwmaLent+5 {
		startIdx = g.vwmaLent + 5
	}

	// Ne pas dépasser lastClosedIdx
	if startIdx > lastClosedIdx {
		return newSignals, nil
	}

	// 1️⃣ Détecter signaux ENTRY (VWMA + DMI matchés)
	entrySignals := g.detectEntrySignals(klines, startIdx, lastClosedIdx)
	newSignals = append(newSignals, entrySignals...)

	// 2️⃣ Détecter signaux EXIT (croisement inverse VWMA pour positions ouvertes)
	exitSignals := g.detectExitSignals(klines, startIdx, lastClosedIdx)
	newSignals = append(newSignals, exitSignals...)

	// Mettre à jour avec la dernière bougie fermée traitée
	g.lastProcessedIdx = lastClosedIdx
	g.metrics.TotalSignals += len(newSignals)

	return newSignals, nil
}

// detectEntrySignals détecte les signaux d'entrée (VWMA + DMI matchés)
func (g *TrendGenerator) detectEntrySignals(klines []signals.Kline, startIdx int, lastClosedIdx int) []signals.Signal {
	var entrySignals []signals.Signal

	// Détecter croisements VWMA et DMI
	vwmaSignals := g.detectVWMASignals(klines, startIdx, lastClosedIdx)
	dmiSignals := g.detectDMISignals(klines, startIdx, lastClosedIdx)

	// Matcher VWMA + DMI
	matchedSignals := g.matchSignals(klines, vwmaSignals, dmiSignals)

	// Convertir en signaux unifiés ENTRY (avec filtres de bougie)
	for _, ms := range matchedSignals {
		sigType := signals.SignalTypeLong
		if ms.Direction == "BAISSIER" {
			sigType = signals.SignalTypeShort
		}

		// Appliquer filtres de bougie initiale
		i := ms.Index
		if i < 0 || i >= len(klines) { continue }
		k := klines[i]
		rangeHL := k.High - k.Low
		if rangeHL <= 0 { continue }
		body := k.Close - k.Open
		bodyAbs := body
		if bodyAbs < 0 { bodyAbs = -bodyAbs }
		bodyPct := bodyAbs / rangeHL
		atrVal := 0.0
		if i < len(g.atr) { atrVal = g.atr[i] }
		// Seuils
		if g.bodyPctMin > 0 && bodyPct < g.bodyPctMin { continue }
		if g.bodyATRMin > 0 && atrVal > 0 && bodyAbs < g.bodyATRMin*atrVal { continue }
		// Direction de bougie cohérente
		if g.enforceCandleDirection {
			if sigType == signals.SignalTypeLong && !(k.Close > k.Open) { continue }
			if sigType == signals.SignalTypeShort && !(k.Close < k.Open) { continue }
		}

		sig := signals.Signal{
			Timestamp:  ms.Timestamp,
			Action:     signals.SignalActionEntry,
			Type:       sigType,
			Price:      ms.Prix,
			Confidence: g.calculateMatchConfidence(ms.DistanceBars),
			Metadata: map[string]interface{}{
				"generator":     "trend",
				"motif":         ms.Motif,
				"distance_bars": ms.DistanceBars,
				"vwma6":         ms.Vwma6,
				"di_plus":       ms.DiPlus,
				"di_minus":      ms.DiMinus,
				"body_pct":      bodyPct,
				"body_to_atr":   func() float64 { if atrVal>0 { return bodyAbs/atrVal }; return 0 }(),
			},
		}

		entrySignals = append(entrySignals, sig)

		// Enregistrer position ouverte
		posID := g.nextPositionID
		g.nextPositionID++
		g.openPositions[posID] = &OpenPosition{
			ID:         posID,
			EntryIndex: ms.Index,
			EntryTime:  ms.Timestamp,
			EntryPrice: ms.Prix,
			Type:       sigType,
		}

		// Créer trailing si activé
		if g.enableExitTrailing {
			atrAtEntry := 0.0
			if ms.Index >= 0 && ms.Index < len(g.atr) { atrAtEntry = g.atr[ms.Index] }
			side := execution.SideShort
			if sigType == signals.SignalTypeLong { side = execution.SideLong }
			effATR := atrAtEntry * g.trailingCoeff
			g.trailings[posID] = execution.NewTrailing(side, ms.Prix, effATR, g.trailingCap)
		}

		g.metrics.EntrySignals++
		if sigType == signals.SignalTypeLong {
			g.metrics.LongSignals++
		} else {
			g.metrics.ShortSignals++
		}
	}

	return entrySignals
}

// detectExitSignals détecte les sorties par croisement inverse VWMA
func (g *TrendGenerator) detectExitSignals(klines []signals.Kline, startIdx int, lastClosedIdx int) []signals.Signal {
	var exitSignals []signals.Signal

	// Traiter jusqu'à lastClosedIdx (pas la bougie en cours)
	for i := startIdx; i <= lastClosedIdx; i++ {
		// 1) Trailing (si activé): priorité si hit
		if g.enableExitTrailing {
			for posID, pos := range g.openPositions {
				t, ok := g.trailings[posID]
				if !ok || t == nil { continue }
				closePrice := klines[i].Close
				t.Update(closePrice)
				if hit, hitPrice := t.Hit(closePrice); hit {
					variation := (hitPrice - pos.EntryPrice) / pos.EntryPrice * 100
					if pos.Type == signals.SignalTypeShort {
						variation = (pos.EntryPrice - hitPrice) / pos.EntryPrice * 100
					}
					exitSig := signals.Signal{
						Timestamp:  klines[i].OpenTime,
						Action:     signals.SignalActionExit,
						Type:       pos.Type,
						Price:      hitPrice,
						Confidence: 0.7,
						EntryPrice: &pos.EntryPrice,
						EntryTime:  &pos.EntryTime,
						Metadata: map[string]interface{}{
							"generator":     "trend",
							"exit_reason":   "trailing",
							"duration_bars": i - pos.EntryIndex,
							"variation_pct": variation,
						},
					}
					exitSignals = append(exitSignals, exitSig)
					delete(g.openPositions, posID)
					delete(g.trailings, posID)
					g.metrics.ExitSignals++
				}
			}
		}

		// 2) VWMA inverse (si activé)
		if g.enableExitVWMA {
			cross, direction := indicators.DetecterCroisement(g.vwmaRapideValues, g.vwmaLentValues, i)
			if cross {
				for posID, pos := range g.openPositions {
					shouldExit := false
					if pos.Type == signals.SignalTypeLong && direction == "BAISSIER" { shouldExit = true }
					if pos.Type == signals.SignalTypeShort && direction == "HAUSSIER" { shouldExit = true }
					if shouldExit {
						variation := (klines[i].Close - pos.EntryPrice) / pos.EntryPrice * 100
						if pos.Type == signals.SignalTypeShort {
							variation = (pos.EntryPrice - klines[i].Close) / pos.EntryPrice * 100
						}
						exitSig := signals.Signal{
							Timestamp:  klines[i].OpenTime,
							Action:     signals.SignalActionExit,
							Type:       pos.Type,
							Price:      klines[i].Close,
							Confidence: 0.8,
							EntryPrice: &pos.EntryPrice,
							EntryTime:  &pos.EntryTime,
							Metadata: map[string]interface{}{
								"generator":     "trend",
								"exit_reason":   "vwma_inverse_cross",
								"duration_bars": i - pos.EntryIndex,
								"variation_pct": variation,
							},
						}
						exitSignals = append(exitSignals, exitSig)
						delete(g.openPositions, posID)
						delete(g.trailings, posID)
						g.metrics.ExitSignals++
					}
				}
			}
		}
	}

	return exitSignals
}

// detectVWMASignals détecte croisements VWMA avec validation gamma
func (g *TrendGenerator) detectVWMASignals(klines []signals.Kline, startIdx int, lastClosedIdx int) []SignalVWMA {
	var vwmaSignalsResult []SignalVWMA

	// Traiter jusqu'à lastClosedIdx
	for i := startIdx; i <= lastClosedIdx; i++ {
		cross, direction := indicators.DetecterCroisement(g.vwmaRapideValues, g.vwmaLentValues, i)
		if !cross {
			continue
		}

		signal := SignalVWMA{
			Index:     i,
			Timestamp: klines[i].OpenTime,
			Direction: direction,
			Prix:      klines[i].Close,
			Vwma6:     g.vwmaRapideValues[i],
			Vwma24:    g.vwmaLentValues[i],
		}

		// Validation gamma gap avec fenêtre différée
		gapInitial := indicators.CalculerEcart(g.vwmaRapideValues[i], g.vwmaLentValues[i])
		signal.Gap = gapInitial
		gammaGapValue := g.gammaGapVWMA * g.atr[i]
		signal.GapValide = gapInitial >= gammaGapValue
		signal.GapValideBougie = -1

		if signal.GapValide {
			signal.GapValideBougie = 0
		} else {
			for w := 1; w <= g.windowGammaValidate; w++ {
				futureIdx := i + w
				if futureIdx >= len(klines)-1 {
					break
				}
				gapFuture := indicators.CalculerEcart(g.vwmaRapideValues[futureIdx], g.vwmaLentValues[futureIdx])
				gammaFuture := g.gammaGapVWMA * g.atr[futureIdx]
				if gapFuture >= gammaFuture {
					signal.GapValide = true
					signal.GapValideBougie = w
					break
				}
			}
		}

		// Validation volatilité
		atrPct := indicators.Normaliser(g.atr[i], signal.Prix)
		signal.AtrPct = atrPct
		signal.VolatiliteOK = atrPct >= g.volatiliteMin

		signal.Valide = signal.GapValide && signal.VolatiliteOK

		if signal.Valide {
			vwmaSignalsResult = append(vwmaSignalsResult, signal)
		}
	}

	return vwmaSignalsResult
}

// detectDMISignals détecte croisements DMI avec validation
func (g *TrendGenerator) detectDMISignals(klines []signals.Kline, startIdx int, lastClosedIdx int) []SignalDMI {
	var dmiSignalsResult []SignalDMI

	// Traiter jusqu'à lastClosedIdx
	for i := startIdx; i <= lastClosedIdx; i++ {
		crossDI, directionDI := indicators.DetecterCroisement(g.diPlus, g.diMinus, i)
		if !crossDI {
			continue
		}

		signal := SignalDMI{
			Index:     i,
			Timestamp: klines[i].OpenTime,
			Direction: directionDI,
			DiPlus:    g.diPlus[i],
			DiMinus:   g.diMinus[i],
			Dx:        g.dx[i],
			Adx:       g.adx[i],
		}

		// Validation gamma gap DI
		gapDIInitial := indicators.CalculerEcart(g.diPlus[i], g.diMinus[i])
		signal.GapDI = gapDIInitial
		signal.GapDIValide = gapDIInitial >= g.gammaGapDI
		signal.GapDIValideBougie = -1

		if signal.GapDIValide {
			signal.GapDIValideBougie = 0
		} else {
			for w := 1; w <= g.windowGammaValidate; w++ {
				futureIdx := i + w
				// Ne pas dépasser lastClosedIdx
				if futureIdx > lastClosedIdx {
					break
				}
				gapDIFuture := indicators.CalculerEcart(g.diPlus[futureIdx], g.diMinus[futureIdx])
				if gapDIFuture >= g.gammaGapDI {
					signal.GapDIValide = true
					signal.GapDIValideBougie = w
					break
				}
			}
		}

		// Validation croisement DX/ADX
		signal.GapDXADXValide = false
		signal.GapDXADXValideBougie = -1

		for w := 0; w <= g.windowGammaValidate; w++ {
			futureIdx := i + w
			// Ne pas dépasser lastClosedIdx
			if futureIdx > lastClosedIdx {
				break
			}

			crossDXADX, directionDXADX := indicators.DetecterCroisement(g.dx, g.adx, futureIdx)
			if crossDXADX && directionDXADX == "HAUSSIER" {
				gapDXADX := indicators.CalculerEcart(g.dx[futureIdx], g.adx[futureIdx])
				signal.GapDXADX = gapDXADX

				if gapDXADX >= g.gammaGapDX {
					signal.GapDXADXValide = true
					signal.GapDXADXValideBougie = w
					break
				}
			}
		}

		signal.Valide = signal.GapDIValide && signal.GapDXADXValide

		if signal.Valide {
			dmiSignalsResult = append(dmiSignalsResult, signal)
		}
	}

	return dmiSignalsResult
}

// matchSignals matche signaux VWMA + DMI dans fenêtre W
func (g *TrendGenerator) matchSignals(klines []signals.Kline, vwmaSignals []SignalVWMA, dmiSignals []SignalDMI) []SignalTrend {
	var trendSignals []SignalTrend

	for _, vwma := range vwmaSignals {
		var bestMatch *SignalDMI
		minDistance := g.windowW + 1

		for _, dmi := range dmiSignals {
			if vwma.Direction != dmi.Direction {
				continue
			}

			distance := vwma.Index - dmi.Index
			if distance < 0 {
				distance = -distance
			}

			if distance <= g.windowW && distance < minDistance {
				bestMatch = &dmi
				minDistance = distance
			}
		}

		if bestMatch != nil {
			motif := ""
			sigTimestamp := vwma.Timestamp
			sigIndex := vwma.Index

			if vwma.Index < bestMatch.Index {
				motif = fmt.Sprintf("VWMA→DMI (+%d bars)", bestMatch.Index-vwma.Index)
			} else if vwma.Index > bestMatch.Index {
				motif = fmt.Sprintf("DMI→VWMA (+%d bars)", vwma.Index-bestMatch.Index)
				sigTimestamp = bestMatch.Timestamp
				sigIndex = bestMatch.Index
			} else {
				motif = "VWMA+DMI simultané"
			}

			trendSignals = append(trendSignals, SignalTrend{
				Index:        sigIndex,
				Timestamp:    sigTimestamp,
				Direction:    vwma.Direction,
				Prix:         vwma.Prix,
				Vwma6:        vwma.Vwma6,
				DiPlus:       bestMatch.DiPlus,
				DiMinus:      bestMatch.DiMinus,
				DistanceBars: minDistance,
				Motif:        motif,
			})
		}
	}

	return trendSignals
}

// calculateMatchConfidence calcule la confiance basée sur distance matching
func (g *TrendGenerator) calculateMatchConfidence(distance int) float64 {
	baseConf := 0.8

	if distance == 0 {
		return 0.95 // Simultané = excellent
	} else if distance <= 3 {
		return 0.85 // Très proche
	} else if distance <= 5 {
		return 0.75 // Proche
	}

	return baseConf - (float64(distance) * 0.05)
}

// GetMetrics retourne les métriques
func (g *TrendGenerator) GetMetrics() signals.GeneratorMetrics {
	return g.metrics
}
