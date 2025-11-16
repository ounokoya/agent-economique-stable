package direction_dmi

import (
	"fmt"
	"sort"
	"time"

	"agent-economique/internal/indicators"
	"agent-economique/internal/signals"
)

// DirectionDMIGenerator générateur de signaux basé sur Direction (VWMA) + DMI/DX/ADX
// Implémentation selon spécification docs/SPEC_DIRECTION_DMI.md
type DirectionDMIGenerator struct {
	config signals.GeneratorConfig

	// Paramètres VWMA (Direction)
	vwmaPeriod          int
	slopePeriod         int
	kConfirmation       int
	useDynamicThreshold bool
	atrPeriod           int
	atrCoefficient      float64
	fixedThreshold      float64

	// Paramètres DMI
	dmiPeriod            int
	dmiSmooth            int
	gammaGapDI           float64
	gammaGapDX           float64
	windowGammaValidate  int
	windowMatching       int

	// Validations optionnelles (nouveaux)
	requireDICrossover   bool    // Exiger croisement DI obligatoire
	requireDXValidation  bool    // Exiger validation DX/ADX
	requireGapValidation bool    // Exiger validation gaps

	// Flags d'activation
	enableEntryTrend        bool
	enableEntryCounterTrend bool
	enableExitTrend         bool
	enableExitCounterTrend  bool

	// État interne - Indicateurs
	vwmaValues    []float64
	slopeValues   []float64
	atrValues     []float64
	diPlus        []float64
	diMinus       []float64
	dx            []float64
	adx           []float64

	// État interne - Positions
	lastProcessedIdx int
	openPositions    map[int]*OpenPosition
	nextPositionID   int
	metrics          signals.GeneratorMetrics
	
	// Tracking des signaux (debug)
	trackedSignals   []TrackedSignal
}

// OpenPosition représente une position ouverte
type OpenPosition struct {
	ID         int
	EntryIndex int
	EntryTime  time.Time
	EntryPrice float64
	Type       signals.SignalType
	Mode       string // "TREND" ou "COUNTER_TREND"
}

// Config configuration spécifique Direction+DMI
type Config struct {
	// Paramètres VWMA (Direction)
	VWMAPeriod          int
	SlopePeriod         int
	KConfirmation       int
	UseDynamicThreshold bool
	ATRPeriod           int
	ATRCoefficient      float64
	FixedThreshold      float64

	// Paramètres DMI
	DMIPeriod            int
	DMISmooth            int
	GammaGapDI           float64
	GammaGapDX           float64
	WindowGammaValidate  int
	WindowMatching       int

	// Flags validation optionnelle (nouveaux)
	RequireDICrossover   bool    // Exiger croisement DI obligatoire
	RequireDXValidation  bool    // Exiger validation DX/ADX
	RequireGapValidation bool    // Exiger validation gaps

	// Flags d'activation
	EnableEntryTrend        bool
	EnableEntryCounterTrend bool
	EnableExitTrend         bool
	EnableExitCounterTrend  bool
}

// SignalVWMA représente un signal VWMA détecté dans une bougie
type SignalVWMA struct {
	Index     int
	Timestamp time.Time
	Direction string  // "RISING" ou "FALLING"
	Slope     float64
	Valid     bool
}

// SignalDMI représente un signal DMI détecté dans une fenêtre
type SignalDMI struct {
	Index       int
	Timestamp   time.Time
	Direction   string  // "DI_PLUS_DOMINANT" ou "DI_MINUS_DOMINANT"
	DIPlus      float64
	DIMinus     float64
	GapDI       float64
	GapDIValid  bool
	Valid       bool
	IsCrossover bool    // true = croisement (TENDANCE), false = position relative (CONTRE-TENDANCE)
}

// SignalDXADX représente un signal DX/ADX détecté dans une bougie
type SignalDXADX struct {
	Index        int
	Timestamp    time.Time
	Direction    string  // "DX_RISING" ou "DX_FALLING"
	DX           float64
	ADX          float64
	GapDX        float64
	GapDXValid   bool
	Valid        bool
}

// CombinedSignal représente un signal combiné validé dans la fenêtre de matching
type CombinedSignal struct {
	Index         int
	Timestamp     time.Time
	Price         float64
	Action        string // "ENTRY" ou "EXIT"
	Type          string // "LONG" ou "SHORT"
	Mode          string // "TREND" ou "COUNTER_TREND"
	
	// Signaux constituants
	VWMASignal    SignalVWMA
	DMISignal     SignalDMI
	DXADXSignal   SignalDXADX
	
	// Fenêtre de matching
	WindowStart   int
	WindowEnd     int
	
	Confidence    float64
}

// TrackedSignal représente un signal avec son statut de validation
type TrackedSignal struct {
	Index       int
	Timestamp   time.Time
	Price       float64
	
	// État des 3 conditions
	VWMAFound   bool
	DMIFound    bool
	DXADXFound  bool
	
	// Signaux constituants (si trouvés)
	VWMASignal  *SignalVWMA
	DMISignal   *SignalDMI
	DXADXSignal *SignalDXADX
	
	// Statut de validation
	IsValid     bool
	Status      string // "VALID", "INCOMPLETE_WINDOW", "FLAG_DISABLED", "GAP_INSUFFICIENT", etc.
	Reason      string // Raison détaillée de l'invalidation
	
	// Classification (si valide)
	Action      string // "ENTRY" ou "EXIT"  
	Type        string // "LONG" ou "SHORT"
	Mode        string // "TREND" ou "COUNTER_TREND"
	
	// Fenêtre de matching
	WindowStart int
	WindowEnd   int
}

// NewDirectionDMIGenerator crée un nouveau générateur Direction+DMI
func NewDirectionDMIGenerator(config signals.GeneratorConfig, dmiConfig Config) *DirectionDMIGenerator {
	return &DirectionDMIGenerator{
		config: config,

		// Paramètres VWMA
		vwmaPeriod:          dmiConfig.VWMAPeriod,
		slopePeriod:         dmiConfig.SlopePeriod,
		kConfirmation:       dmiConfig.KConfirmation,
		useDynamicThreshold: dmiConfig.UseDynamicThreshold,
		atrPeriod:           dmiConfig.ATRPeriod,
		atrCoefficient:      dmiConfig.ATRCoefficient,
		fixedThreshold:      dmiConfig.FixedThreshold,

		// Paramètres DMI
		dmiPeriod:           dmiConfig.DMIPeriod,
		dmiSmooth:           dmiConfig.DMISmooth,
		gammaGapDI:          dmiConfig.GammaGapDI,
		gammaGapDX:          dmiConfig.GammaGapDX,
		windowGammaValidate: dmiConfig.WindowGammaValidate,
		windowMatching:      dmiConfig.WindowMatching,

		// Validations optionnelles (nouveaux)
		requireDICrossover:   dmiConfig.RequireDICrossover,
		requireDXValidation:  dmiConfig.RequireDXValidation,
		requireGapValidation: dmiConfig.RequireGapValidation,

		// Flags
		enableEntryTrend:        dmiConfig.EnableEntryTrend,
		enableEntryCounterTrend: dmiConfig.EnableEntryCounterTrend,
		enableExitTrend:         dmiConfig.EnableExitTrend,
		enableExitCounterTrend:  dmiConfig.EnableExitCounterTrend,

		// État interne
		openPositions:   make(map[int]*OpenPosition),
		nextPositionID:  1,
		lastProcessedIdx: -1,
	}
}

// Name retourne le nom du générateur
func (g *DirectionDMIGenerator) Name() string {
	return "DirectionDMI"
}

// Initialize initialise le générateur
func (g *DirectionDMIGenerator) Initialize() error {
	// Validation des paramètres
	if g.vwmaPeriod < 1 {
		return fmt.Errorf("VWMAPeriod doit être > 0: %d", g.vwmaPeriod)
	}
	if g.slopePeriod < 1 {
		return fmt.Errorf("SlopePeriod doit être > 0: %d", g.slopePeriod)
	}
	if g.dmiPeriod < 1 {
		return fmt.Errorf("DMIPeriod doit être > 0: %d", g.dmiPeriod)
	}
	if g.windowMatching < 1 {
		return fmt.Errorf("WindowMatching doit être > 0: %d", g.windowMatching)
	}

	// Validation des flags (au moins un exit doit être actif)
	if !g.enableExitTrend && !g.enableExitCounterTrend {
		return fmt.Errorf("au moins un flag exit doit être activé")
	}

	return nil
}

// CalculateIndicators calcule VWMA, pente, ATR, DMI, DX, ADX
func (g *DirectionDMIGenerator) CalculateIndicators(klines []signals.Kline) error {
	if len(klines) < g.vwmaPeriod {
		return fmt.Errorf("pas assez de klines pour VWMA: %d < %d", len(klines), g.vwmaPeriod)
	}
	if len(klines) < g.dmiPeriod {
		return fmt.Errorf("pas assez de klines pour DMI: %d < %d", len(klines), g.dmiPeriod)
	}

	// Préparation des données
	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	volumes := make([]float64, len(klines))

	for i, kline := range klines {
		closes[i] = kline.Close
		highs[i] = kline.High
		lows[i] = kline.Low
		volumes[i] = kline.Volume
	}

	// VWMA
	vwmaInd := indicators.NewVWMATVStandard(g.vwmaPeriod)
	g.vwmaValues = vwmaInd.Calculate(closes, volumes)

	// Pente VWMA
	g.slopeValues = make([]float64, len(g.vwmaValues))
	for i := g.slopePeriod; i < len(g.vwmaValues); i++ {
		slope := (g.vwmaValues[i] - g.vwmaValues[i-g.slopePeriod]) / g.vwmaValues[i-g.slopePeriod] * 100
		g.slopeValues[i] = slope
	}

	// ATR (pour seuil dynamique)
	if g.useDynamicThreshold {
		atrInd := indicators.NewATRTVStandard(g.atrPeriod)
		g.atrValues = atrInd.Calculate(highs, lows, closes)
	}

	// DMI (DI+, DI-)
	dmiInd := indicators.NewDMITVStandard(g.dmiPeriod)
	g.diPlus, g.diMinus, g.adx = dmiInd.Calculate(highs, lows, closes)

	// DX (manuellement car pas forcément dans DMI standard)
	g.dx = make([]float64, len(g.diPlus))
	for i := range g.diPlus {
		if g.diPlus[i]+g.diMinus[i] != 0 {
			g.dx[i] = (abs(g.diPlus[i]-g.diMinus[i]) / (g.diPlus[i] + g.diMinus[i])) * 100
		}
	}

	return nil
}

// abs retourne la valeur absolue
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// DetectSignals détecte les signaux selon spécification avec fenêtre de matching
func (g *DirectionDMIGenerator) DetectSignals(klines []signals.Kline) ([]signals.Signal, error) {
	var newSignals []signals.Signal

	// Déterminer la plage à traiter
	startIdx := g.lastProcessedIdx + 1
	if startIdx < 0 {
		startIdx = 0
	}

	// Dans un moteur temporel, on ne traite que les bougies fermées
	lastClosedIdx := len(klines) - 2
	if lastClosedIdx < startIdx {
		return newSignals, nil
	}

	// Détecter signaux combinés avec fenêtre de matching
	combinedSignals := g.detectCombinedSignals(klines, startIdx, lastClosedIdx)

	// Prioriser : EXIT avant ENTRY
	exitSignals := []CombinedSignal{}
	entrySignals := []CombinedSignal{}

	for _, sig := range combinedSignals {
		if sig.Action == "EXIT" {
			exitSignals = append(exitSignals, sig)
		} else {
			entrySignals = append(entrySignals, sig)
		}
	}

	// Traiter les sorties d'abord
	for _, exitSig := range exitSignals {
		signal := g.convertToUnifiedSignal(exitSig, klines)
		if signal != nil {
			newSignals = append(newSignals, *signal)
			// Fermer la position correspondante
			g.closePosition(exitSig)
		}
	}

	// Puis traiter les entrées
	for _, entrySig := range entrySignals {
		signal := g.convertToUnifiedSignal(entrySig, klines)
		if signal != nil {
			newSignals = append(newSignals, *signal)
			// Ouvrir nouvelle position
			g.openPosition(entrySig)
		}
	}

	// Mettre à jour métriques
	g.metrics.TotalSignals += len(newSignals)
	g.lastProcessedIdx = lastClosedIdx

	return newSignals, nil
}

// detectCombinedSignals détecte les signaux combinés avec fenêtre de matching et tracking
func (g *DirectionDMIGenerator) detectCombinedSignals(klines []signals.Kline, startIdx, endIdx int) []CombinedSignal {
	var combinedSignals []CombinedSignal

	// 1. Collecter tous les événements réels d'abord
	vwmaEvents := g.detectVWMAEvents(klines, startIdx, endIdx)
	dmiEvents := g.detectDMIEvents(klines, startIdx, endIdx)  
	dxadxEvents := g.detectDXADXEvents(klines, startIdx, endIdx)

	// 2. Pour chaque événement, tenter le matching avec fenêtre
	allEvents := make(map[int]bool) // Index des bougies avec au moins un événement
	
	for _, event := range vwmaEvents {
		allEvents[event.Index] = true
	}
	for _, event := range dmiEvents {
		allEvents[event.Index] = true
	}
	for _, event := range dxadxEvents {
		allEvents[event.Index] = true
	}

	// 3. Convertir map en slice triée pour ordre chronologique
	var eventIndices []int
	for eventIdx := range allEvents {
		if eventIdx >= startIdx && eventIdx <= endIdx {
			eventIndices = append(eventIndices, eventIdx)
		}
	}
	
	// Trier par index chronologique
	sort.Ints(eventIndices)
	
	// 4. Analyser les événements dans l'ordre chronologique
	for _, eventIdx := range eventIndices {
		
		windowStart := max(0, eventIdx-g.windowMatching+1)
		windowEnd := eventIdx

		// Chercher les 3 conditions dans la fenêtre
		vwmaSignal := g.findVWMASignalInWindow(klines, windowStart, windowEnd)
		dmiSignal := g.findDMISignalInWindow(klines, windowStart, windowEnd)
		dxadxSignal := g.findDXADXSignalInWindow(klines, windowStart, windowEnd)

		// Créer TrackedSignal pour debug
		tracked := TrackedSignal{
			Index:       eventIdx,
			Timestamp:   klines[eventIdx].OpenTime,
			Price:       klines[eventIdx].Close,
			VWMAFound:   vwmaSignal.Valid,
			DMIFound:    dmiSignal.Valid,
			DXADXFound:  dxadxSignal.Valid,
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
		}

		// Ajouter signaux constituants si trouvés
		if vwmaSignal.Valid {
			tracked.VWMASignal = &vwmaSignal
		}
		if dmiSignal.Valid {
			tracked.DMISignal = &dmiSignal
		}
		if dxadxSignal.Valid {
			tracked.DXADXSignal = &dxadxSignal
		}

		// Vérifier si fenêtre complète
		if vwmaSignal.Valid && dmiSignal.Valid && dxadxSignal.Valid {
			// Classifier le signal combiné
			combined := g.classifyeCombinedSignal(klines, eventIdx, vwmaSignal, dmiSignal, dxadxSignal)
			if combined.Action != "" { // Signal valide (flag activé)
				tracked.IsValid = true
				tracked.Status = "VALID"
				tracked.Reason = "Toutes les conditions remplies"
				tracked.Action = combined.Action
				tracked.Type = combined.Type
				tracked.Mode = combined.Mode
				
				combined.WindowStart = windowStart
				combined.WindowEnd = windowEnd
				combinedSignals = append(combinedSignals, combined)
			} else {
				// Fenêtre complète mais flag désactivé
				tracked.IsValid = false
				tracked.Status = "FLAG_DISABLED"
				tracked.Reason = g.getDisabledFlagReason(vwmaSignal, dmiSignal, dxadxSignal)
			}
		} else {
			// Fenêtre incomplète mais avec au moins un événement réel
			tracked.IsValid = false
			tracked.Status = "INCOMPLETE_WINDOW"  
			tracked.Reason = g.getMissingConditionsReason(vwmaSignal.Valid, dmiSignal.Valid, dxadxSignal.Valid)
		}

		// Ajouter au tracking
		g.trackedSignals = append(g.trackedSignals, tracked)
	}

	return combinedSignals
}

// max retourne le maximum de deux entiers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetMetrics retourne les métriques du générateur
func (g *DirectionDMIGenerator) GetMetrics() signals.GeneratorMetrics {
	return g.metrics
}

// closePosition ferme une position
func (g *DirectionDMIGenerator) closePosition(exitSignal CombinedSignal) {
	// Trouver position correspondante au type
	for id, pos := range g.openPositions {
		if (exitSignal.Type == "LONG" && pos.Type == signals.SignalTypeLong) ||
		   (exitSignal.Type == "SHORT" && pos.Type == signals.SignalTypeShort) {
			delete(g.openPositions, id)
			break
		}
	}
}

// openPosition ouvre une nouvelle position
func (g *DirectionDMIGenerator) openPosition(entrySignal CombinedSignal) {
	var posType signals.SignalType
	if entrySignal.Type == "LONG" {
		posType = signals.SignalTypeLong
	} else {
		posType = signals.SignalTypeShort
	}

	pos := &OpenPosition{
		ID:         g.nextPositionID,
		EntryIndex: entrySignal.Index,
		EntryTime:  entrySignal.Timestamp,
		EntryPrice: entrySignal.Price,
		Type:       posType,
		Mode:       entrySignal.Mode,
	}

	g.openPositions[g.nextPositionID] = pos
	g.nextPositionID++
}

// convertToUnifiedSignal convertit un signal combiné en signal unifié
func (g *DirectionDMIGenerator) convertToUnifiedSignal(combined CombinedSignal, klines []signals.Kline) *signals.Signal {
	var sigType signals.SignalType
	var sigAction signals.SignalAction

	// Type
	if combined.Type == "LONG" {
		sigType = signals.SignalTypeLong
	} else {
		sigType = signals.SignalTypeShort
	}

	// Action
	if combined.Action == "ENTRY" {
		sigAction = signals.SignalActionEntry
	} else {
		sigAction = signals.SignalActionExit
	}

	// Métadonnées
	metadata := map[string]interface{}{
		"mode":              combined.Mode,
		"vwma_slope":        combined.VWMASignal.Slope,
		"vwma_direction":    combined.VWMASignal.Direction,
		"dmi_di_plus":       combined.DMISignal.DIPlus,
		"dmi_di_minus":      combined.DMISignal.DIMinus,
		"dmi_gap":           combined.DMISignal.GapDI,
		"dx":                combined.DXADXSignal.DX,
		"adx":               combined.DXADXSignal.ADX,
		"dx_gap":            combined.DXADXSignal.GapDX,
		"window_start":      combined.WindowStart,
		"window_end":        combined.WindowEnd,
		"generator":         "DirectionDMI",
	}

	return &signals.Signal{
		Timestamp:  combined.Timestamp,
		Type:       sigType,
		Action:     sigAction,
		Price:      combined.Price,
		Confidence: combined.Confidence,
		Metadata:   metadata,
	}
}

//
// MÉTHODES DE DÉTECTION DANS FENÊTRE
//

// findVWMASignalInWindow cherche un signal VWMA valide dans la fenêtre [start, end]
func (g *DirectionDMIGenerator) findVWMASignalInWindow(klines []signals.Kline, start, end int) SignalVWMA {
	for i := start; i <= end; i++ {
		if i >= g.slopePeriod && i < len(g.slopeValues) {
			slope := g.slopeValues[i]
			
			// Calculer seuil
			threshold := g.fixedThreshold
			if g.useDynamicThreshold && i < len(g.atrValues) {
				atrPct := (g.atrValues[i] / klines[i].Close) * 100
				threshold = atrPct * g.atrCoefficient
			}
			
			// Vérifier si pente significative
			if abs(slope) >= threshold {
				direction := "RISING"
				if slope < 0 {
					direction = "FALLING"
				}
				
				// Vérification instantanée seulement (sens VWMA au moment du croisement)
				confirmed := true  // Pas de K-confirmation, juste direction instantanée
				
				if confirmed {
					return SignalVWMA{
						Index:     i,
						Timestamp: klines[i].OpenTime,
						Direction: direction,
						Slope:     slope,
						Valid:     true,
					}
				}
			}
		}
	}
	
	return SignalVWMA{Valid: false}
}

// isVWMATrendConfirmed vérifie la K-confirmation de la tendance VWMA
func (g *DirectionDMIGenerator) isVWMATrendConfirmed(index int, direction string) bool {
	if index < g.kConfirmation {
		return false
	}
	
	for k := 1; k <= g.kConfirmation; k++ {
		prevIdx := index - k
		if prevIdx < g.slopePeriod || prevIdx >= len(g.slopeValues) {
			return false
		}
		
		prevSlope := g.slopeValues[prevIdx]
		if direction == "RISING" && prevSlope <= 0 {
			return false
		}
		if direction == "FALLING" && prevSlope >= 0 {
			return false
		}
	}
	
	return true
}

// findDMISignalInWindow cherche un signal DMI valide dans la fenêtre [start, end]
// La logique diffère selon si on cherche des signaux TENDANCE (croisement obligatoire) ou CONTRE-TENDANCE (position relative suffit)
func (g *DirectionDMIGenerator) findDMISignalInWindow(klines []signals.Kline, start, end int) SignalDMI {
	// D'abord chercher les croisements (pour signaux TENDANCE)
	for i := start; i <= end; i++ {
		if i >= g.dmiPeriod && i < len(g.diPlus) && i < len(g.diMinus) {
			// Vérifier croisements DMI
			cross, _ := indicators.DetecterCroisement(g.diPlus, g.diMinus, i)
			if cross {
				diPlus := g.diPlus[i]
				diMinus := g.diMinus[i]
				gapDI := abs(diPlus - diMinus)
				
				// Valider gap dans fenêtre
				gapValid := g.validateGapInWindow(gapDI, g.gammaGapDI, i, g.windowGammaValidate)
				
				if gapValid {
					dmiDirection := "DI_PLUS_DOMINANT"
					if diMinus > diPlus {
						dmiDirection = "DI_MINUS_DOMINANT"
					}
					
					return SignalDMI{
						Index:      i,
						Timestamp:  klines[i].OpenTime,
						Direction:  dmiDirection,
						DIPlus:     diPlus,
						DIMinus:    diMinus,
						GapDI:      gapDI,
						GapDIValid: true,
						Valid:      true,
						IsCrossover: true, // Marqueur croisement
					}
				}
			}
		}
	}
	
	// Si pas de croisement trouvé, chercher position relative (pour signaux CONTRE-TENDANCE)
	for i := start; i <= end; i++ {
		if i >= g.dmiPeriod && i < len(g.diPlus) && i < len(g.diMinus) {
			diPlus := g.diPlus[i]
			diMinus := g.diMinus[i]
			gapDI := abs(diPlus - diMinus)
			
			// Vérifier si gap DMI suffisant (position relative)
			if gapDI >= g.gammaGapDI {
				// Valider gap dans fenêtre
				gapValid := g.validateGapInWindow(gapDI, g.gammaGapDI, i, g.windowGammaValidate)
				
				if gapValid {
					dmiDirection := "DI_PLUS_DOMINANT"
					if diMinus > diPlus {
						dmiDirection = "DI_MINUS_DOMINANT"
					}
					
					return SignalDMI{
						Index:      i,
						Timestamp:  klines[i].OpenTime,
						Direction:  dmiDirection,
						DIPlus:     diPlus,
						DIMinus:    diMinus,
						GapDI:      gapDI,
						GapDIValid: true,
						Valid:      true,
						IsCrossover: false, // Pas de croisement, juste position
					}
				}
			}
		}
	}
	
	return SignalDMI{Valid: false}
}

// findDXADXSignalInWindow cherche un signal DX/ADX valide dans la fenêtre [start, end]
func (g *DirectionDMIGenerator) findDXADXSignalInWindow(klines []signals.Kline, start, end int) SignalDXADX {
	for i := start; i <= end; i++ {
		if i >= g.dmiPeriod && i < len(g.dx) && i < len(g.adx) {
			// Vérifier croisements DX/ADX
			cross, direction := indicators.DetecterCroisement(g.dx, g.adx, i)
			if cross {
				dx := g.dx[i]
				adx := g.adx[i]
				gapDX := abs(dx - adx)
				
				// Valider gap dans fenêtre
				gapValid := g.validateGapInWindow(gapDX, g.gammaGapDX, i, g.windowGammaValidate)
				
				if gapValid {
					dxDirection := "DX_RISING"
					if direction == "BAISSIER" {
						dxDirection = "DX_FALLING"
					}
					
					return SignalDXADX{
						Index:      i,
						Timestamp:  klines[i].OpenTime,
						Direction:  dxDirection,
						DX:         dx,
						ADX:        adx,
						GapDX:      gapDX,
						GapDXValid: true,
						Valid:      true,
					}
				}
			}
		}
	}
	
	return SignalDXADX{Valid: false}
}

// validateGapInWindow valide qu'un gap est suffisant dans une fenêtre
func (g *DirectionDMIGenerator) validateGapInWindow(gap, minGap float64, index, window int) bool {
	if gap >= minGap {
		return true
	}
	
	// Chercher dans fenêtre de validation
	for w := 1; w <= window; w++ {
		futureIdx := index + w
		if futureIdx >= len(g.diPlus) || futureIdx >= len(g.dx) {
			break
		}
		
		// Re-calculer gaps aux indices futurs
		gapDIFuture := abs(g.diPlus[futureIdx] - g.diMinus[futureIdx])
		gapDXFuture := abs(g.dx[futureIdx] - g.adx[futureIdx])
		
		if gapDIFuture >= minGap || gapDXFuture >= minGap {
			return true
		}
	}
	
	return false
}

// classifyeCombinedSignal classifie un signal combiné selon la matrice des 4 types
func (g *DirectionDMIGenerator) classifyeCombinedSignal(klines []signals.Kline, index int, vwma SignalVWMA, dmi SignalDMI, dxadx SignalDXADX) CombinedSignal {
	combined := CombinedSignal{
		Index:       index,
		Timestamp:   klines[index].OpenTime,
		Price:       klines[index].Close,
		VWMASignal:  vwma,
		DMISignal:   dmi,
		DXADXSignal: dxadx,
		Confidence:  0.0,
	}
	
	// Classifier selon la matrice avec validation du type DMI (croisement vs position)
	if vwma.Direction == "RISING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		// Signal 1: LONG Tendance (OBLIGATOIRE: croisement DMI)
		if g.enableEntryTrend && dmi.IsCrossover {
			combined.Action = "ENTRY"
			combined.Type = "LONG"
			combined.Mode = "TREND"
			combined.Confidence = 0.9
		}
	} else if vwma.Direction == "RISING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		// Signal 2: LONG Contre-Tendance (position relative suffit)
		if g.enableEntryCounterTrend {
			combined.Action = "ENTRY"
			combined.Type = "LONG"
			combined.Mode = "COUNTER_TREND"
			combined.Confidence = 0.7
		}
	} else if vwma.Direction == "FALLING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		// Signal 3: SHORT Tendance (OBLIGATOIRE: croisement DMI)
		if g.enableEntryTrend && dmi.IsCrossover {
			combined.Action = "ENTRY"
			combined.Type = "SHORT"
			combined.Mode = "TREND"
			combined.Confidence = 0.9
		}
	} else if vwma.Direction == "FALLING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		// Signal 4: SHORT Contre-Tendance (position relative suffit)
		if g.enableEntryCounterTrend {
			combined.Action = "ENTRY"
			combined.Type = "SHORT"
			combined.Mode = "COUNTER_TREND"
			combined.Confidence = 0.7
		}
	}
	
	// COMBINAISONS SUPPLÉMENTAIRES (non-trade mais à signaler comme invalides)
	// Ces combinaisons existent mais ne correspondent pas aux stratégies définies
	// Elles sont marquées comme invalides mais trackées pour debugging
	
	// Vérifier sorties pour positions ouvertes
	for _, pos := range g.openPositions {
		if pos.Type == signals.SignalTypeLong {
			// Sorties LONG
			if vwma.Direction == "FALLING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_RISING" && g.enableExitTrend {
				// Exit LONG Tendance
				combined.Action = "EXIT"
				combined.Type = "LONG"
				combined.Mode = "TREND"
				combined.Confidence = 0.8
				break
			} else if vwma.Direction == "FALLING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_FALLING" && g.enableExitCounterTrend {
				// Exit LONG Contre-Tendance
				combined.Action = "EXIT"
				combined.Type = "LONG"
				combined.Mode = "COUNTER_TREND"
				combined.Confidence = 0.6
				break
			}
		} else if pos.Type == signals.SignalTypeShort {
			// Sorties SHORT
			if vwma.Direction == "RISING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_RISING" && g.enableExitTrend {
				// Exit SHORT Tendance
				combined.Action = "EXIT"
				combined.Type = "SHORT"
				combined.Mode = "TREND"
				combined.Confidence = 0.8
				break
			} else if vwma.Direction == "RISING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_FALLING" && g.enableExitCounterTrend {
				// Exit SHORT Contre-Tendance
				combined.Action = "EXIT"
				combined.Type = "SHORT"
				combined.Mode = "COUNTER_TREND"
				combined.Confidence = 0.6
				break
			}
		}
	}
	
	return combined
}

//
// MÉTHODES DE DEBUG ET TRACKING
//

// GetTrackedSignals retourne tous les signaux trackés (debug)
func (g *DirectionDMIGenerator) GetTrackedSignals() []TrackedSignal {
	return g.trackedSignals
}

// GetOpenPositionsCount retourne le nombre de positions ouvertes (debug)
func (g *DirectionDMIGenerator) GetOpenPositionsCount() int {
	return len(g.openPositions)
}

// getDisabledFlagReason retourne la raison pour laquelle un signal avec fenêtre complète est désactivé
func (g *DirectionDMIGenerator) getDisabledFlagReason(vwma SignalVWMA, dmi SignalDMI, dxadx SignalDXADX) string {
	// Classifier le signal pour savoir quel flag devrait être actif
	if vwma.Direction == "RISING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		if !dmi.IsCrossover {
			return "Signal LONG Tendance: Croisement DMI requis mais seulement position relative détectée"
		}
		if !g.enableEntryTrend {
			return "Flag 'enable_entry_trend' désactivé pour LONG Tendance"
		}
	} else if vwma.Direction == "RISING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		if !g.enableEntryCounterTrend {
			return "Flag 'enable_entry_counter_trend' désactivé pour LONG Contre-Tendance"
		}
	} else if vwma.Direction == "FALLING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		if !dmi.IsCrossover {
			return "Signal SHORT Tendance: Croisement DMI requis mais seulement position relative détectée"
		}
		if !g.enableEntryTrend {
			return "Flag 'enable_entry_trend' désactivé pour SHORT Tendance"
		}
	} else if vwma.Direction == "FALLING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		if !g.enableEntryCounterTrend {
			return "Flag 'enable_entry_counter_trend' désactivé pour SHORT Contre-Tendance"
		}
	}
	
	// Combinaisons invalides non-stratégiques
	if vwma.Direction == "RISING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		return "Combinaison non-stratégique: VWMA↗ + DI+>DI- + DX↓ (signal mixte non défini)"
	}
	if vwma.Direction == "RISING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		return "Combinaison non-stratégique: VWMA↗ + DI-<DI+ + DX↑ (signal mixte non défini)"
	}
	if vwma.Direction == "FALLING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_RISING" {
		return "Combinaison non-stratégique: VWMA↘ + DI+>DI- + DX↑ (signal mixte non défini)"
	}
	if vwma.Direction == "FALLING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_FALLING" {
		return "Combinaison non-stratégique: VWMA↘ + DI-<DI+ + DX↓ (signal mixte non défini)"
	}
	
	// Vérifier flags de sortie si positions ouvertes
	for _, pos := range g.openPositions {
		if pos.Type == signals.SignalTypeLong {
			if vwma.Direction == "FALLING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_RISING" && !g.enableExitTrend {
				return "Flag 'enable_exit_trend' désactivé pour Exit LONG Tendance"
			}
			if vwma.Direction == "FALLING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_FALLING" && !g.enableExitCounterTrend {
				return "Flag 'enable_exit_counter_trend' désactivé pour Exit LONG Contre-Tendance"
			}
		} else if pos.Type == signals.SignalTypeShort {
			if vwma.Direction == "RISING" && dmi.Direction == "DI_PLUS_DOMINANT" && dxadx.Direction == "DX_RISING" && !g.enableExitTrend {
				return "Flag 'enable_exit_trend' désactivé pour Exit SHORT Tendance"
			}
			if vwma.Direction == "RISING" && dmi.Direction == "DI_MINUS_DOMINANT" && dxadx.Direction == "DX_FALLING" && !g.enableExitCounterTrend {
				return "Flag 'enable_exit_counter_trend' désactivé pour Exit SHORT Contre-Tendance"
			}
		}
	}
	
	return "Aucune position ouverte pour signal de sortie"
}

// getMissingConditionsReason retourne la raison pour laquelle la fenêtre est incomplète
func (g *DirectionDMIGenerator) getMissingConditionsReason(vwmaValid, dmiValid, dxadxValid bool) string {
	missing := []string{}
	
	if !vwmaValid {
		missing = append(missing, "VWMA (pente insuffisante ou K-confirmation échouée)")
	}
	if !dmiValid {
		missing = append(missing, "DMI (pas de croisement DI ou gap insuffisant)")
	}
	if !dxadxValid {
		missing = append(missing, "DX/ADX (pas de croisement ou gap insuffisant)")
	}
	
	if len(missing) == 0 {
		return "Erreur: toutes les conditions semblent présentes"
	}
	
	result := "Conditions manquantes: "
	for i, m := range missing {
		if i > 0 {
			result += ", "
		}
		result += m
	}
	
	return result
}

// detectVWMAEvents détecte tous les événements VWMA réels (changements de pente significatifs)
func (g *DirectionDMIGenerator) detectVWMAEvents(klines []signals.Kline, startIdx, endIdx int) []SignalVWMA {
	var events []SignalVWMA
	
	for i := startIdx; i <= endIdx; i++ {
		if i >= g.slopePeriod && i < len(g.slopeValues) {
			slope := g.slopeValues[i]
			
			// Calculer seuil
			threshold := g.fixedThreshold
			if g.useDynamicThreshold && i < len(g.atrValues) {
				atrPct := (g.atrValues[i] / klines[i].Close) * 100
				threshold = atrPct * g.atrCoefficient
			}
			
			// Vérifier si pente significative
			if abs(slope) >= threshold {
				direction := "RISING"
				if slope < 0 {
					direction = "FALLING"
				}
				
				// Vérification instantanée seulement (sens VWMA au moment du croisement)
				confirmed := true  // Pas de K-confirmation, juste direction instantanée
				
				if confirmed {
					events = append(events, SignalVWMA{
						Index:     i,
						Timestamp: klines[i].OpenTime,
						Direction: direction,
						Slope:     slope,
						Valid:     true,
					})
				}
			}
		}
	}
	
	return events
}

// detectDMIEvents détecte tous les événements DMI réels (changements de position DI+ vs DI-)
func (g *DirectionDMIGenerator) detectDMIEvents(klines []signals.Kline, startIdx, endIdx int) []SignalDMI {
	var events []SignalDMI
	
	for i := startIdx; i <= endIdx; i++ {
		if i >= g.dmiPeriod && i < len(g.diPlus) && i < len(g.diMinus) {
			diPlus := g.diPlus[i]
			diMinus := g.diMinus[i]
			gapDI := abs(diPlus - diMinus)
			
			// Chercher changement de dominance OU gap suffisant nouveau
			shouldDetect := false
			
			// 1. Vérifier si croisement (changement de dominance)
			if i > 0 && i-1 < len(g.diPlus) && i-1 < len(g.diMinus) {
				prevDIPlus := g.diPlus[i-1]
				prevDIMinus := g.diMinus[i-1]
				
				// Changement de dominance
				if (prevDIPlus > prevDIMinus && diMinus > diPlus) || 
				   (prevDIMinus > prevDIPlus && diPlus > diMinus) {
					shouldDetect = true
				}
			}
			
			// 2. OU si gap DMI devient suffisant pour la première fois
			if gapDI >= g.gammaGapDI {
				if i > 0 && i-1 < len(g.diPlus) && i-1 < len(g.diMinus) {
					prevGap := abs(g.diPlus[i-1] - g.diMinus[i-1])
					if prevGap < g.gammaGapDI {
						shouldDetect = true // Gap devient suffisant
					}
				} else {
					shouldDetect = true // Premier point avec gap suffisant
				}
			}
			
			if shouldDetect && gapDI >= g.gammaGapDI {
				// Valider gap dans fenêtre
				gapValid := g.validateGapInWindow(gapDI, g.gammaGapDI, i, g.windowGammaValidate)
				
				if gapValid {
					dmiDirection := "DI_PLUS_DOMINANT"
					if diMinus > diPlus {
						dmiDirection = "DI_MINUS_DOMINANT"
					}
					
					events = append(events, SignalDMI{
						Index:      i,
						Timestamp:  klines[i].OpenTime,
						Direction:  dmiDirection,
						DIPlus:     diPlus,
						DIMinus:    diMinus,
						GapDI:      gapDI,
						GapDIValid: true,
						Valid:      true,
					})
				}
			}
		}
	}
	
	return events
}

// detectDXADXEvents détecte tous les événements DX/ADX réels (croisements DX vs ADX)
func (g *DirectionDMIGenerator) detectDXADXEvents(klines []signals.Kline, startIdx, endIdx int) []SignalDXADX {
	var events []SignalDXADX
	
	for i := startIdx; i <= endIdx; i++ {
		if i >= g.dmiPeriod && i < len(g.dx) && i < len(g.adx) {
			// Vérifier croisements DX/ADX
			cross, direction := indicators.DetecterCroisement(g.dx, g.adx, i)
			if cross {
				dx := g.dx[i]
				adx := g.adx[i]
				gapDX := abs(dx - adx)
				
				// Valider gap dans fenêtre
				gapValid := g.validateGapInWindow(gapDX, g.gammaGapDX, i, g.windowGammaValidate)
				
				if gapValid {
					dxDirection := "DX_RISING"
					if direction == "BAISSIER" {
						dxDirection = "DX_FALLING"
					}
					
					events = append(events, SignalDXADX{
						Index:      i,
						Timestamp:  klines[i].OpenTime,
						Direction:  dxDirection,
						DX:         dx,
						ADX:        adx,
						GapDX:      gapDX,
						GapDXValid: true,
						Valid:      true,
					})
				}
			}
		}
	}
	
	return events
}

// GetDMIValues expose les valeurs DMI calculées (pour debug)
func (g *DirectionDMIGenerator) GetDMIValues() ([]float64, []float64, []float64) {
	return g.diPlus, g.diMinus, g.adx
}
