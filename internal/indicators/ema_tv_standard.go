package indicators

import (
	"math"
	"sort"
)

// EMATVStandard - Implémentation EMA TradingView Standard
// Basé sur la documentation ema_tradingview_research.md
// Composants: Alpha, Seed SMA, Lazy seeding, Recursive formula
type EMATVStandard struct {
	period int
}

// NewEMATVStandard crée une nouvelle instance EMA TV Standard
func NewEMATVStandard(period int) *EMATVStandard {
	return &EMATVStandard{
		period: period,
	}
}

// Calculate calcule l'EMA selon les spécifications TradingView
func (ema *EMATVStandard) Calculate(src []float64) []float64 {
	n := len(src)
	out := make([]float64, n)
	
	// Initialiser avec NaN (TradingView standard)
	for i := range out {
		out[i] = math.NaN()
	}
	
	if ema.period <= 0 || n == 0 || ema.period > n {
		return out
	}

	// Précalculer SMA pour seed (TradingView lazy seeding)
	smaTV := NewSMATVStandard(ema.period)
	sma := smaTV.Calculate(src)
	alpha := 2.0 / (float64(ema.period) + 1.0)

	// Lazy seed + re-seeding forward pass
	seeded := false
	for i := ema.period - 1; i < n; i++ {
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
		
		// Formule EMA récursive TradingView
		out[i] = alpha*v + (1.0-alpha)*prev
	}
	
	return out
}

// GetLastValue retourne la dernière valeur valide de l'EMA
func (ema *EMATVStandard) GetLastValue(emaValues []float64) float64 {
	for i := len(emaValues) - 1; i >= 0; i-- {
		if !math.IsNaN(emaValues[i]) {
			return emaValues[i]
		}
	}
	return math.NaN()
}

// GetAlpha retourne le coefficient alpha utilisé
func (ema *EMATVStandard) GetAlpha() float64 {
	return 2.0 / (float64(ema.period) + 1.0)
}

// IsRising vérifie si l'EMA est en hausse
func (ema *EMATVStandard) IsRising(emaValues []float64, index int) bool {
	if index <= 0 || index >= len(emaValues) {
		return false
	}

	prev := emaValues[index-1]
	curr := emaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr > prev
}

// IsFalling vérifie si l'EMA est en baisse
func (ema *EMATVStandard) IsFalling(emaValues []float64, index int) bool {
	if index <= 0 || index >= len(emaValues) {
		return false
	}

	prev := emaValues[index-1]
	curr := emaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr < prev
}

// GetSlope calcule la pente de l'EMA
func (ema *EMATVStandard) GetSlope(emaValues []float64, index int) float64 {
	if index <= 0 || index >= len(emaValues) {
		return math.NaN()
	}

	prev := emaValues[index-1]
	curr := emaValues[index]

	if math.IsNaN(prev) || math.IsNaN(curr) {
		return math.NaN()
	}

	return curr - prev
}

// GetDirection retourne la direction de l'EMA
func (ema *EMATVStandard) GetDirection(emaValues []float64, index int) string {
	slope := ema.GetSlope(emaValues, index)
	
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

// IsAbovePrice vérifie si l'EMA est au-dessus du prix
func (ema *EMATVStandard) IsAbovePrice(emaValues []float64, prices []float64, index int) bool {
	if index >= len(emaValues) || index >= len(prices) {
		return false
	}

	emaVal := emaValues[index]
	price := prices[index]

	return !math.IsNaN(emaVal) && !math.IsNaN(price) && emaVal > price
}

// IsBelowPrice vérifie si l'EMA est en dessous du prix
func (ema *EMATVStandard) IsBelowPrice(emaValues []float64, prices []float64, index int) bool {
	if index >= len(emaValues) || index >= len(prices) {
		return false
	}

	emaVal := emaValues[index]
	price := prices[index]

	return !math.IsNaN(emaVal) && !math.IsNaN(price) && emaVal < price
}

// GetCrossoverType détecte le type de croisement avec le prix
func (ema *EMATVStandard) GetCrossoverType(emaValues []float64, prices []float64, index int) string {
	if index <= 0 || index >= len(emaValues) || index >= len(prices) {
		return "Inconnu"
	}

	prevEMA := emaValues[index-1]
	currEMA := emaValues[index]
	prevPrice := prices[index-1]
	currPrice := prices[index]

	if math.IsNaN(prevEMA) || math.IsNaN(currEMA) || 
	   math.IsNaN(prevPrice) || math.IsNaN(currPrice) {
		return "Inconnu"
	}

	// Croisement haussier: EMA passe sous prix à dessus prix
	if prevEMA <= prevPrice && currEMA > currPrice {
		return "Haussier"
	}

	// Croisement baissier: EMA passe dessus prix à sous prix
	if prevEMA >= prevPrice && currEMA < currPrice {
		return "Baissier"
	}

	return "Aucun"
}

// GetDistanceFromPrice calcule la distance de l'EMA au prix en pourcentage
func (ema *EMATVStandard) GetDistanceFromPrice(emaValues []float64, prices []float64, index int) float64 {
	if index >= len(emaValues) || index >= len(prices) {
		return math.NaN()
	}

	emaVal := emaValues[index]
	price := prices[index]

	if math.IsNaN(emaVal) || math.IsNaN(price) || price == 0 {
		return math.NaN()
	}

	return (emaVal - price) / price * 100.0
}

// IsConverging vérifie si l'EMA converge vers le prix
func (ema *EMATVStandard) IsConverging(emaValues []float64, prices []float64, index int, lookback int) bool {
	if lookback <= 1 || index < lookback || 
	   index >= len(emaValues) || index >= len(prices) {
		return false
	}

	// Calculer la distance actuelle vs précédente
	currentDist := math.Abs(ema.GetDistanceFromPrice(emaValues, prices, index))
	prevDist := math.Abs(ema.GetDistanceFromPrice(emaValues, prices, index-1))

	if math.IsNaN(currentDist) || math.IsNaN(prevDist) {
		return false
	}

	return currentDist < prevDist
}

// CalculateFromKlines calcule EMA depuis les klines (méthode utilitaire)
func (ema *EMATVStandard) CalculateFromKlines(klines []Kline, src func(Kline) float64) []float64 {
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

	return ema.Calculate(values)
}
