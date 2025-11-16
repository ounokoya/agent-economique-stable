package vwma_cross_dmi_simple

import (
	"fmt"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

// VWMACrossDMISimpleGenerator générateur de signaux basé sur VWMA Cross + DMI Simple
type VWMACrossDMISimpleGenerator struct {
	config signals.GeneratorConfig

	// Paramètres spécifiques
	vwmaShortPeriod int
	vwmaLongPeriod  int
	dmiPeriod        int
	dmiSmooth        int
	windowMatching   int

	// État interne
	vwmaShort      []float64
	vwmaLong       []float64
	diPlus         []float64
	diMinus        []float64
	dx             []float64
	adx            []float64
	currentPosition signals.SignalType // LONG ou SHORT (pas de NONE)
	lastProcessedIdx int
	metrics          signals.GeneratorMetrics
}

// Config configuration spécifique VWMA Cross + DMI Simple
type Config struct {
	VWMAShortPeriod int
	VWMALongPeriod  int
	DMIPeriod        int
	DMISmooth        int
	WindowMatching   int
}

// NewVWMACrossDMISimpleGenerator crée un nouveau générateur VWMA Cross + DMI Simple
func NewVWMACrossDMISimpleGenerator(config Config) *VWMACrossDMISimpleGenerator {
	return &VWMACrossDMISimpleGenerator{
		vwmaShortPeriod: config.VWMAShortPeriod,
		vwmaLongPeriod:  config.VWMALongPeriod,
		dmiPeriod:        config.DMIPeriod,
		dmiSmooth:        config.DMISmooth,
		windowMatching:   config.WindowMatching,
		currentPosition:  signals.SignalTypeLong, // Initialiser à LONG par défaut
		lastProcessedIdx: -1,
	}
}

// Name retourne le nom du générateur
func (g *VWMACrossDMISimpleGenerator) Name() string {
	return "VWMACrossDMISimple"
}

// Initialize initialise le générateur avec la config commune
func (g *VWMACrossDMISimpleGenerator) Initialize(config signals.GeneratorConfig) error {
	g.config = config
	return nil
}

// CalculateIndicators calcule VWMA court/long et DMI
func (g *VWMACrossDMISimpleGenerator) CalculateIndicators(klines []signals.Kline) error {
	if len(klines) < g.vwmaLongPeriod {
		return fmt.Errorf("pas assez de klines: %d < %d", len(klines), g.vwmaLongPeriod)
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

	// Calculer VWMA court et long
	vwmaShortIndicator := indicators.NewVWMATVStandard(g.vwmaShortPeriod)
	g.vwmaShort = vwmaShortIndicator.Calculate(closes, volumes)

	vwmaLongIndicator := indicators.NewVWMATVStandard(g.vwmaLongPeriod)
	g.vwmaLong = vwmaLongIndicator.Calculate(closes, volumes)

	// DMI
	dmiInd := indicators.NewDMITVStandard(g.dmiPeriod)
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

// DetectSignals détecte nouveaux signaux LONG/SHORT contextuels
func (g *VWMACrossDMISimpleGenerator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var newSignals []signals.Signal

	// IMPORTANT: len(klines)-1 = bougie EN COURS (ignorer)
	//            len(klines)-2 = dernière bougie FERMÉE
	lastClosedIdx := len(klines) - 2
	if lastClosedIdx < 0 {
		return newSignals, nil
	}

	// Démarrer après la période de warmup
	startIdx := g.lastProcessedIdx + 1
	maxPeriod := max(g.vwmaLongPeriod, g.dmiPeriod+g.dmiSmooth)
	warmupIdx := maxPeriod + g.windowMatching
	if startIdx < warmupIdx {
		startIdx = warmupIdx
	}

	// Traiter jusqu'à lastClosedIdx (inclus)
	for i := startIdx; i <= lastClosedIdx; i++ {
		// Appliquer fenêtre de matching
		signal := g.detectSignalInWindow(klines, i)
		if signal != nil {
			newSignals = append(newSignals, *signal)
			g.metrics.TotalSignals++

			// Mettre à jour position courante
			if signal.Type == signals.SignalTypeLong {
				g.currentPosition = signals.SignalTypeLong
				g.metrics.LongSignals++
			} else if signal.Type == signals.SignalTypeShort {
				g.currentPosition = signals.SignalTypeShort
				g.metrics.ShortSignals++
			}

			// Comptabiliser entry/exit selon contexte
			if signal.Action == signals.SignalActionEntry {
				g.metrics.EntrySignals++
			} else {
				g.metrics.ExitSignals++
			}
		}
	}

	// Mettre à jour avec la dernière bougie fermée traitée
	g.lastProcessedIdx = lastClosedIdx

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

// detectSignalInWindow détecte un signal dans la fenêtre de matching
func (g *VWMACrossDMISimpleGenerator) detectSignalInWindow(klines []signals.Kline, currentIdx int) *signals.Signal {
	// Définir fenêtre de recherche
	windowStart := max(g.vwmaLongPeriod, currentIdx-g.windowMatching+1)
	if windowStart < 0 {
		windowStart = 0
	}

	// Chercher les 3 conditions dans la fenêtre
	vwmaCrossOk := false
	diPositionOk := false
	dxAdxCrossOk := false

	var vwmaDirection string
	var diDominant string
	var dxCrossDirection string

	for i := windowStart; i <= currentIdx; i++ {
		// 1. Détecter croisement VWMA court/long
		if !vwmaCrossOk && g.detectVWMACross(i) {
			vwmaCrossOk = true
			if g.vwmaShort[i] > g.vwmaLong[i] {
				vwmaDirection = "UP"
			} else {
				vwmaDirection = "DOWN"
			}
		}

		// 2. Détecter position relative DI
		if !diPositionOk && g.detectDIPosition(i) {
			diPositionOk = true
			if g.diPlus[i] > g.diMinus[i] {
				diDominant = "DI_PLUS"
			} else {
				diDominant = "DI_MINUS"
			}
		}

		// 3. Détecter croisement DX/ADX
		if !dxAdxCrossOk && g.detectDXADXCross(i) {
			dxAdxCrossOk = true
			if g.dx[i] > g.adx[i] {
				dxCrossDirection = "UP"
			} else {
				dxCrossDirection = "DOWN"
			}
		}

		// Sortir si les 3 conditions sont trouvées
		if vwmaCrossOk && diPositionOk && dxAdxCrossOk {
			break
		}
	}

	// Générer signal si les 3 conditions sont réunies
	if vwmaCrossOk && diPositionOk && dxAdxCrossOk {
		return g.createSignal(klines[currentIdx], currentIdx, vwmaDirection, diDominant, dxCrossDirection)
	}

	return nil
}

// detectVWMACross détecte un croisement VWMA court/long
func (g *VWMACrossDMISimpleGenerator) detectVWMACross(idx int) bool {
	if idx <= 0 {
		return false
	}

	// Croisement vers le haut (court passe au-dessus du long)
	if g.vwmaShort[idx-1] <= g.vwmaLong[idx-1] && g.vwmaShort[idx] > g.vwmaLong[idx] {
		return true
	}

	// Croisement vers le bas (court passe en dessous du long)
	if g.vwmaShort[idx-1] >= g.vwmaLong[idx-1] && g.vwmaShort[idx] < g.vwmaLong[idx] {
		return true
	}

	return false
}

// detectDIPosition détecte la position relative DI+ vs DI-
func (g *VWMACrossDMISimpleGenerator) detectDIPosition(idx int) bool {
	// Simple position relative (pas de croisement)
	return g.diPlus[idx] != g.diMinus[idx]
}

// detectDXADXCross détecte un croisement DX vs ADX
func (g *VWMACrossDMISimpleGenerator) detectDXADXCross(idx int) bool {
	if idx <= 0 {
		return false
	}

	// Croisement vers le haut (DX passe au-dessus de ADX)
	if g.dx[idx-1] <= g.adx[idx-1] && g.dx[idx] > g.adx[idx] {
		return true
	}

	// Croisement vers le bas (DX passe en dessous de ADX)
	if g.dx[idx-1] >= g.adx[idx-1] && g.dx[idx] < g.adx[idx] {
		return true
	}

	return false
}

// createSignal crée un signal selon les 3 conditions
func (g *VWMACrossDMISimpleGenerator) createSignal(kline signals.Kline, idx int, vwmaDirection, diDominant, dxCrossDirection string) *signals.Signal {
	var signalType signals.SignalType
	var mode string

	// Déterminer le type de signal (LONG ou SHORT) selon VWMA
	if vwmaDirection == "UP" {
		signalType = signals.SignalTypeLong
	} else {
		signalType = signals.SignalTypeShort
	}

	// Déterminer le mode (TREND ou COUNTER_TREND)
	if (vwmaDirection == "UP" && diDominant == "DI_PLUS" && dxCrossDirection == "UP") ||
		(vwmaDirection == "DOWN" && diDominant == "DI_MINUS" && dxCrossDirection == "UP") {
		mode = "TREND"
	} else {
		mode = "COUNTER_TREND"
	}

	// Déterminer l'action contextuelle (ENTRY ou EXIT)
	var action signals.SignalAction
	
	// Si pas de position ouverte = ENTRY
	if (g.currentPosition == signals.SignalTypeLong && signalType == signals.SignalTypeShort) ||
	   (g.currentPosition == signals.SignalTypeShort && signalType == signals.SignalTypeLong) {
		action = signals.SignalActionExit
	} else {
		action = signals.SignalActionEntry
	}

	// Calculer confiance
	confidence := g.calculateConfidence(mode)

	return &signals.Signal{
		Timestamp: kline.OpenTime,
		Action:    action,
		Type:      signalType,
		Price:     kline.Close,
		Confidence: confidence,
		Metadata: map[string]interface{}{
			"generator":          "vwma_cross_dmi_simple",
			"mode":               mode,
			"vwma_direction":     vwmaDirection,
			"di_dominant":        diDominant,
			"dx_cross_direction": dxCrossDirection,
			"current_position":   string(g.currentPosition),
			"vwma_short":         g.vwmaShort[idx],
			"vwma_long":          g.vwmaLong[idx],
			"di_plus":            g.diPlus[idx],
			"di_minus":           g.diMinus[idx],
			"dx":                 g.dx[idx],
			"adx":                g.adx[idx],
		},
	}
}

// calculateConfidence calcule la confiance du signal
func (g *VWMACrossDMISimpleGenerator) calculateConfidence(mode string) float64 {
	baseConf := 0.7

	if mode == "TREND" {
		baseConf = 0.8 // Signaux tendance plus fiables
	} else {
		baseConf = 0.6 // Signaux contre-tendance moins fiables
	}

	return baseConf
}

// GetMetrics retourne les métriques du générateur
func (g *VWMACrossDMISimpleGenerator) GetMetrics() signals.GeneratorMetrics {
	return g.metrics
}

// SetPosition permet de définir la position courante (pour backtest)
func (g *VWMACrossDMISimpleGenerator) SetPosition(position signals.SignalType) {
	g.currentPosition = position
}

// GetCurrentPosition retourne la position courante
func (g *VWMACrossDMISimpleGenerator) GetCurrentPosition() signals.SignalType {
	return g.currentPosition
}

// ResetPosition réinitialise la position (pour début de backtest)
func (g *VWMACrossDMISimpleGenerator) ResetPosition() {
	g.currentPosition = signals.SignalTypeLong // Forcer première entrée
}

// max fonction utilitaire
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
