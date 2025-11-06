package indicators

import (
	"math"
	"sort"
)

// RMATVStandard - Implémentation RMA TradingView Standard
// Basé sur la documentation rma_tradingview_research.md
// Composants: Wilder's Smoothing, Alpha = 1/length, Seed SMA, Recursive formula
type RMATVStandard struct {
	period int
}

// NewRMATVStandard crée une nouvelle instance RMA TV Standard
func NewRMATVStandard(period int) *RMATVStandard {
	return &RMATVStandard{
		period: period,
	}
}

// Calculate calcule le RMA selon les spécifications TradingView
func (rma *RMATVStandard) Calculate(src []float64) []float64 {
	n := len(src)
	out := make([]float64, n)
	
	// Initialiser avec NaN (TradingView standard)
	for i := range out {
		out[i] = math.NaN()
	}
	
	if rma.period <= 0 || n == 0 || rma.period > n {
		return out
	}

	// Cas spécial: length == 1 (comme TradingView)
	if rma.period == 1 {
		for i := 0; i < n; i++ {
			out[i] = src[i]
		}
		return out
	}

	// Précalculer SMA pour seed (TradingView lazy seeding)
	smaTV := NewSMATVStandard(rma.period)
	sma := smaTV.Calculate(src)

	// Lazy seed + re-seeding forward pass
	seeded := false
	for i := rma.period - 1; i < n; i++ {
		if !seeded {
			// Seed avec SMA si disponible
			if !math.IsNaN(sma[i]) && !math.IsInf(sma[i], 0) {
				out[i] = sma[i]
				seeded = true
			} else {
				out[i] = math.NaN()
			}
			continue
		}

		prev := out[i-1]
		v := src[i]
		
		// Si continuité brisée ou src invalide
		if math.IsNaN(prev) || math.IsInf(prev, 0) || 
		   math.IsNaN(v) || math.IsInf(v, 0) {
			out[i] = math.NaN()
			seeded = false
			continue
		}
		
		// Formule RMA récursive TradingView (Wilder's Smoothing)
		// out[i] = (out[i-1]*(period-1) + src[i]) / period
		out[i] = (prev*float64(rma.period-1) + v) / float64(rma.period)
	}
	
	return out
}

// GetLastValue retourne la dernière valeur valide du RMA
func (rma *RMATVStandard) GetLastValue(rmaValues []float64) float64 {
	for i := len(rmaValues) - 1; i >= 0; i-- {
		if !math.IsNaN(rmaValues[i]) {
			return rmaValues[i]
		}
	}
	return math.NaN()
}

// GetAlpha retourne le coefficient alpha utilisé (Wilder's)
func (rma *RMATVStandard) GetAlpha() float64 {
	return 1.0 / float64(rma.period)
}

// GetSmoothingFactor retourne le facteur de lissage
func (rma *RMATVStandard) GetSmoothingFactor() float64 {
	return float64(rma.period - 1) / float64(rma.period)
}

// IsRising vérifie si le RMA est en hausse
func (rma *RMATVStandard) IsRising(rmaValues []float64, index int) bool {
	if index <= 0 || index >= len(rmaValues) {
		return false
	}

	prev := rmaValues[index-1]
	curr := rmaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr > prev
}

// IsFalling vérifie si le RMA est en baisse
func (rma *RMATVStandard) IsFalling(rmaValues []float64, index int) bool {
	if index <= 0 || index >= len(rmaValues) {
		return false
	}

	prev := rmaValues[index-1]
	curr := rmaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr < prev
}

// GetSlope calcule la pente du RMA
func (rma *RMATVStandard) GetSlope(rmaValues []float64, index int) float64 {
	if index <= 0 || index >= len(rmaValues) {
		return math.NaN()
	}

	prev := rmaValues[index-1]
	curr := rmaValues[index]

	if math.IsNaN(prev) || math.IsNaN(curr) {
		return math.NaN()
	}

	return curr - prev
}

// GetDirection retourne la direction du RMA
func (rma *RMATVStandard) GetDirection(rmaValues []float64, index int) string {
	slope := rma.GetSlope(rmaValues, index)
	
	if math.IsNaN(slope) {
		return "Inconnue"
	}
	
	if slope > 0 {
		return "Haussière"
	} else if slope < 0 {
		return "Baissière"
	} else {
		return "Plate"
	}
}

// IsAbovePrice vérifie si le RMA est au-dessus du prix
func (rma *RMATVStandard) IsAbovePrice(rmaValues []float64, prices []float64, index int) bool {
	if index >= len(rmaValues) || index >= len(prices) {
		return false
	}

	rmaVal := rmaValues[index]
	price := prices[index]

	return !math.IsNaN(rmaVal) && !math.IsNaN(price) && rmaVal > price
}

// IsBelowPrice vérifie si le RMA est en dessous du prix
func (rma *RMATVStandard) IsBelowPrice(rmaValues []float64, prices []float64, index int) bool {
	if index >= len(rmaValues) || index >= len(prices) {
		return false
	}

	rmaVal := rmaValues[index]
	price := prices[index]

	return !math.IsNaN(rmaVal) && !math.IsNaN(price) && rmaVal < price
}

// GetCrossoverType détecte le type de croisement avec le prix
func (rma *RMATVStandard) GetCrossoverType(rmaValues []float64, prices []float64, index int) string {
	if index <= 0 || index >= len(rmaValues) || index >= len(prices) {
		return "Inconnu"
	}

	prevRMA := rmaValues[index-1]
	currRMA := rmaValues[index]
	prevPrice := prices[index-1]
	currPrice := prices[index]

	if math.IsNaN(prevRMA) || math.IsNaN(currRMA) || 
	   math.IsNaN(prevPrice) || math.IsNaN(currPrice) {
		return "Inconnu"
	}

	// Croisement haussier: RMA passe sous prix à dessus prix
	if prevRMA <= prevPrice && currRMA > currPrice {
		return "Haussier"
	}

	// Croisement baissier: RMA passe dessus prix à sous prix
	if prevRMA >= prevPrice && currRMA < currPrice {
		return "Baissier"
	}

	return "Aucun"
}

// GetDistanceFromPrice calcule la distance du RMA au prix en pourcentage
func (rma *RMATVStandard) GetDistanceFromPrice(rmaValues []float64, prices []float64, index int) float64 {
	if index >= len(rmaValues) || index >= len(prices) {
		return math.NaN()
	}

	rmaVal := rmaValues[index]
	price := prices[index]

	if math.IsNaN(rmaVal) || math.IsNaN(price) || price == 0 {
		return math.NaN()
	}

	return (rmaVal - price) / price * 100.0
}

// IsConverging vérifie si le RMA converge vers le prix
func (rma *RMATVStandard) IsConverging(rmaValues []float64, prices []float64, index int, lookback int) bool {
	if lookback <= 1 || index < lookback || 
	   index >= len(rmaValues) || index >= len(prices) {
		return false
	}

	// Calculer la distance actuelle vs précédente
	currentDist := math.Abs(rma.GetDistanceFromPrice(rmaValues, prices, index))
	prevDist := math.Abs(rma.GetDistanceFromPrice(rmaValues, prices, index-1))

	if math.IsNaN(currentDist) || math.IsNaN(prevDist) {
		return false
	}

	return currentDist < prevDist
}

// GetMomentum calcule le momentum du RMA
func (rma *RMATVStandard) GetMomentum(rmaValues []float64, index int, lookback int) float64 {
	if lookback <= 0 || index < lookback || index >= len(rmaValues) {
		return math.NaN()
	}

	current := rmaValues[index]
	previous := rmaValues[index-lookback]

	if math.IsNaN(current) || math.IsNaN(previous) {
		return math.NaN()
	}

	return (current - previous) / previous * 100.0
}

// IsStable vérifie si le RMA est stable (faible variation)
func (rma *RMATVStandard) IsStable(rmaValues []float64, index int, lookback int, threshold float64) bool {
	if lookback <= 1 || index < lookback || index >= len(rmaValues) {
		return false
	}

	momentum := math.Abs(rma.GetMomentum(rmaValues, index, lookback))
	
	return !math.IsNaN(momentum) && momentum < threshold
}

// GetWilderSmoothingFactor retourne le facteur de lissage Wilder's
func (rma *RMATVStandard) GetWilderSmoothingFactor() float64 {
	return float64(rma.period - 1) / float64(rma.period)
}

// CompareWithSMA compare le RMA avec le SMA
func (rma *RMATVStandard) CompareWithSMA(rmaValues, smaValues []float64, index int) string {
	if index >= len(rmaValues) || index >= len(smaValues) {
		return "Inconnu"
	}

	rmaVal := rmaValues[index]
	smaVal := smaValues[index]

	if math.IsNaN(rmaVal) || math.IsNaN(smaVal) {
		return "Inconnu"
	}

	diff := math.Abs(rmaVal - smaVal)
	
	if diff < 0.01 {
		return "Identique"
	} else if rmaVal > smaVal {
		return "RMA > SMA"
	} else {
		return "RMA < SMA"
	}
}

// CalculateFromKlines calcule RMA depuis les klines (méthode utilitaire)
func (rma *RMATVStandard) CalculateFromKlines(klines []Kline, src func(Kline) float64) []float64 {
	if len(klines) == 0 {
		return nil
	}

	// Trier chronologiquement (défensif)
	if len(klines) > 1 {
		cpy := make([]Kline, len(klines))
		copy(cpy, klines)
		sort.Slice(cpy, func(i, j int) bool { 
			return cpy[i].Timestamp < cpy[j].Timestamp 
		})
		klines = cpy
	}

	// Extraire les données source
	values := make([]float64, len(klines))
	for i, k := range klines {
		values[i] = src(k)
	}

	return rma.Calculate(values)
}
