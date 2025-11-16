package direction

import (
	"fmt"
	"time"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

// DirectionGenerator générateur de signaux basé sur direction VWMA6
type DirectionGenerator struct {
	config signals.GeneratorConfig

	// Paramètres spécifiques
	vwmaPeriod          int
	slopePeriod         int
	kConfirmation       int
	useDynamicThreshold bool
	atrPeriod           int
	atrCoefficient      float64
	fixedThreshold      float64

	// État interne
	vwma6            []float64
	atr              []float64
	currentInterval  *Interval
	lastProcessedIdx int
	metrics          signals.GeneratorMetrics
}

// Interval représente un intervalle directionnel en cours
type Interval struct {
	StartIndex int
	Direction  signals.SignalType // LONG (croissant) ou SHORT (décroissant)
	StartPrice float64
	StartTime  time.Time
}

// Config configuration spécifique Direction
type Config struct {
	VWMAPeriod          int
	SlopePeriod         int
	KConfirmation       int
	UseDynamicThreshold bool
	ATRPeriod           int
	ATRCoefficient      float64
	FixedThreshold      float64
}

// NewDirectionGenerator crée un nouveau générateur Direction
func NewDirectionGenerator(config Config) *DirectionGenerator {
	return &DirectionGenerator{
		vwmaPeriod:          config.VWMAPeriod,
		slopePeriod:         config.SlopePeriod,
		kConfirmation:       config.KConfirmation,
		useDynamicThreshold: config.UseDynamicThreshold,
		atrPeriod:           config.ATRPeriod,
		atrCoefficient:      config.ATRCoefficient,
		fixedThreshold:      config.FixedThreshold,
		lastProcessedIdx:    -1,
	}
}

// Name retourne le nom du générateur
func (g *DirectionGenerator) Name() string {
	return "DirectionVWMA"
}

// Initialize initialise le générateur avec la config commune
func (g *DirectionGenerator) Initialize(config signals.GeneratorConfig) error {
	g.config = config
	return nil
}

// CalculateIndicators calcule VWMA6 et ATR
func (g *DirectionGenerator) CalculateIndicators(klines []signals.Kline) error {
	if len(klines) < g.vwmaPeriod {
		return fmt.Errorf("pas assez de klines: %d < %d", len(klines), g.vwmaPeriod)
	}

	// Extraire données
	closes := make([]float64, len(klines))
	volumes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))

	for i, k := range klines {
		closes[i] = k.Close
		volumes[i] = k.Volume
		highs[i] = k.High
		lows[i] = k.Low
	}

	// Calculer VWMA
	vwmaIndicator := indicators.NewVWMATVStandard(g.vwmaPeriod)
	g.vwma6 = vwmaIndicator.Calculate(closes, volumes)

	// Calculer ATR si mode dynamique
	if g.useDynamicThreshold {
		atrIndicator := indicators.NewATRTVStandard(g.atrPeriod)
		g.atr = atrIndicator.Calculate(highs, lows, closes)
	}

	return nil
}

// DetectSignals détecte nouveaux signaux ENTRY/EXIT
func (g *DirectionGenerator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var newSignals []signals.Signal

	// IMPORTANT: len(klines)-1 = bougie EN COURS (ignorer)
	//            len(klines)-2 = dernière bougie FERMÉE
	lastClosedIdx := len(klines) - 2
	if lastClosedIdx < 0 {
		return newSignals, nil
	}

	// Démarrer après la période de warmup
	startIdx := g.lastProcessedIdx + 1
	if startIdx < g.vwmaPeriod+g.slopePeriod+g.kConfirmation {
		startIdx = g.vwmaPeriod + g.slopePeriod + g.kConfirmation
	}

	// Traiter jusqu'à lastClosedIdx (inclus), pas jusqu'à la fin
	for i := startIdx; i <= lastClosedIdx; i++ {
		// Calculer variation de pente sur N bougies
		slopeVariation := (g.vwma6[i] - g.vwma6[i-g.slopePeriod]) / g.vwma6[i-g.slopePeriod] * 100

		// Calculer seuil (fixe ou dynamique)
		threshold := g.fixedThreshold
		if g.useDynamicThreshold && len(g.atr) > i {
			threshold = (g.atr[i] / klines[i].Close) * 100 * g.atrCoefficient
		}

		// Déterminer direction
		var newDirection signals.SignalType
		var isDirectional bool

		if slopeVariation > threshold {
			newDirection = signals.SignalTypeLong // Croissant
			isDirectional = true
		} else if slopeVariation < -threshold {
			newDirection = signals.SignalTypeShort // Décroissant
			isDirectional = true
		} else {
			isDirectional = false // Stable
		}

		// Gérer changement de direction
		if isDirectional {
			// Si pas d'intervalle en cours OU changement de direction
			if g.currentInterval == nil || g.currentInterval.Direction != newDirection {
				// Clôturer intervalle précédent (signal EXIT)
				if g.currentInterval != nil {
					exitSignal := g.createExitSignal(g.currentInterval, klines[i], i)
					newSignals = append(newSignals, exitSignal)
					g.metrics.ExitSignals++
				}

				// Ouvrir nouvel intervalle (signal ENTRY)
				g.currentInterval = &Interval{
					StartIndex: i,
					Direction:  newDirection,
					StartPrice: klines[i].Close,
					StartTime:  klines[i].OpenTime,
				}

				entrySignal := g.createEntrySignal(g.currentInterval, klines[i])
				newSignals = append(newSignals, entrySignal)
				g.metrics.EntrySignals++

				if newDirection == signals.SignalTypeLong {
					g.metrics.LongSignals++
				} else {
					g.metrics.ShortSignals++
				}
			}
		}
	}

	// Mettre à jour avec la dernière bougie fermée traitée
	g.lastProcessedIdx = lastClosedIdx
	g.metrics.TotalSignals += len(newSignals)

	// Calculer confiance moyenne
	if len(newSignals) > 0 {
		totalConf := 0.0
		for _, sig := range newSignals {
			totalConf += sig.Confidence
		}
		g.metrics.AvgConfidence = totalConf / float64(len(newSignals))
		g.metrics.LastSignalTime = newSignals[len(newSignals)-1].Timestamp
	}

	return newSignals, nil
}

// createEntrySignal crée un signal d'entrée
func (g *DirectionGenerator) createEntrySignal(interval *Interval, kline signals.Kline) signals.Signal {
	return signals.Signal{
		Timestamp:  kline.OpenTime,
		Action:     signals.SignalActionEntry,
		Type:       interval.Direction,
		Price:      kline.Close,
		Confidence: 0.7, // Confiance initiale moyenne
		Metadata: map[string]interface{}{
			"generator":   "direction",
			"start_index": interval.StartIndex,
			"vwma6":       g.vwma6[interval.StartIndex],
		},
	}
}

// createExitSignal crée un signal de sortie
func (g *DirectionGenerator) createExitSignal(interval *Interval, kline signals.Kline, currentIdx int) signals.Signal {
	duration := currentIdx - interval.StartIndex
	variation := (kline.Close - interval.StartPrice) / interval.StartPrice * 100

	// Calculer confiance basée sur durée et variation
	confidence := g.calculateExitConfidence(duration, variation)

	return signals.Signal{
		Timestamp:  kline.OpenTime,
		Action:     signals.SignalActionExit,
		Type:       interval.Direction,
		Price:      kline.Close,
		Confidence: confidence,
		EntryPrice: &interval.StartPrice,
		EntryTime:  &interval.StartTime,
		Metadata: map[string]interface{}{
			"generator":      "direction",
			"start_index":    interval.StartIndex,
			"end_index":      currentIdx,
			"duration_bars":  duration,
			"variation_pct":  variation,
			"entry_vwma6":    g.vwma6[interval.StartIndex],
			"exit_vwma6":     g.vwma6[currentIdx],
		},
	}
}

// calculateExitConfidence calcule la confiance du signal EXIT
func (g *DirectionGenerator) calculateExitConfidence(duration int, variation float64) float64 {
	// Confiance basée sur durée et variation captée
	baseConf := 0.5

	// Bonus pour longue durée (trend fort)
	if duration >= 50 {
		baseConf = 0.9
	} else if duration >= 20 {
		baseConf = 0.7
	} else if duration >= 10 {
		baseConf = 0.6
	}

	// Bonus pour grosse variation
	absVariation := variation
	if absVariation < 0 {
		absVariation = -absVariation
	}

	if absVariation >= 5.0 {
		baseConf += 0.1
	} else if absVariation >= 2.0 {
		baseConf += 0.05
	}

	// Limiter entre 0.3 et 1.0
	if baseConf > 1.0 {
		baseConf = 1.0
	}
	if baseConf < 0.3 {
		baseConf = 0.3
	}

	return baseConf
}

// GetMetrics retourne les métriques du générateur
func (g *DirectionGenerator) GetMetrics() signals.GeneratorMetrics {
	return g.metrics
}
