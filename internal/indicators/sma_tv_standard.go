package indicators

import (
	"math"
	"sort"
)

// SMATVStandard - Implémentation SMA TradingView Standard
// Basé sur la documentation sma_tradingview_research.md
// Composants: Simple Moving Average, Sliding window, NaN handling
type SMATVStandard struct {
	period int
}

// NewSMATVStandard crée une nouvelle instance SMA TV Standard
func NewSMATVStandard(period int) *SMATVStandard {
	return &SMATVStandard{
		period: period,
	}
}

// Calculate calcule le SMA selon les spécifications TradingView
func (sma *SMATVStandard) Calculate(src []float64) []float64 {
	n := len(src)
	out := make([]float64, n)
	
	// Initialiser avec NaN (TradingView standard)
	for i := range out {
		out[i] = math.NaN()
	}
	
	if n == 0 || sma.period <= 0 || sma.period > n {
		return out
	}

	// Helper pour marquer les valeurs invalides
	isBad := func(x float64) bool { 
		return math.IsNaN(x) || math.IsInf(x, 0) 
	}

	var sum float64
	var count int

	// Fenêtre glissante optimisée
	for i := 0; i < n; i++ {
		val := src[i]
		
		if isBad(val) {
			// Valeur invalide: réinitialiser la fenêtre
			sum = 0
			count = 0
			out[i] = math.NaN()
			continue
		}

		// Ajouter la valeur actuelle
		sum += val
		count++

		// Retirer la valeur qui sort de la fenêtre
		if i >= sma.period {
			oldVal := src[i-sma.period]
			if !isBad(oldVal) {
				sum -= oldVal
				count--
			}
		}

		// Calculer SMA seulement si la fenêtre est complète et valide
		if i >= sma.period-1 && count == sma.period {
			out[i] = sum / float64(sma.period)
		} else {
			out[i] = math.NaN()
		}
	}
	
	return out
}

// GetLastValue retourne la dernière valeur valide du SMA
func (sma *SMATVStandard) GetLastValue(smaValues []float64) float64 {
	for i := len(smaValues) - 1; i >= 0; i-- {
		if !math.IsNaN(smaValues[i]) {
			return smaValues[i]
		}
	}
	return math.NaN()
}

// IsRising vérifie si le SMA est en hausse
func (sma *SMATVStandard) IsRising(smaValues []float64, index int) bool {
	if index <= 0 || index >= len(smaValues) {
		return false
	}

	prev := smaValues[index-1]
	curr := smaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr > prev
}

// IsFalling vérifie si le SMA est en baisse
func (sma *SMATVStandard) IsFalling(smaValues []float64, index int) bool {
	if index <= 0 || index >= len(smaValues) {
		return false
	}

	prev := smaValues[index-1]
	curr := smaValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) && curr < prev
}

// GetSlope calcule la pente du SMA
func (sma *SMATVStandard) GetSlope(smaValues []float64, index int) float64 {
	if index <= 0 || index >= len(smaValues) {
		return math.NaN()
	}

	prev := smaValues[index-1]
	curr := smaValues[index]

	if math.IsNaN(prev) || math.IsNaN(curr) {
		return math.NaN()
	}

	return curr - prev
}

// GetDirection retourne la direction du SMA
func (sma *SMATVStandard) GetDirection(smaValues []float64, index int) string {
	slope := sma.GetSlope(smaValues, index)
	
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

// IsAbovePrice vérifie si le SMA est au-dessus du prix
func (sma *SMATVStandard) IsAbovePrice(smaValues []float64, prices []float64, index int) bool {
	if index >= len(smaValues) || index >= len(prices) {
		return false
	}

	smaVal := smaValues[index]
	price := prices[index]

	return !math.IsNaN(smaVal) && !math.IsNaN(price) && smaVal > price
}

// IsBelowPrice vérifie si le SMA est en dessous du prix
func (sma *SMATVStandard) IsBelowPrice(smaValues []float64, prices []float64, index int) bool {
	if index >= len(smaValues) || index >= len(prices) {
		return false
	}

	smaVal := smaValues[index]
	price := prices[index]

	return !math.IsNaN(smaVal) && !math.IsNaN(price) && smaVal < price
}

// GetCrossoverType détecte le type de croisement avec le prix
func (sma *SMATVStandard) GetCrossoverType(smaValues []float64, prices []float64, index int) string {
	if index <= 0 || index >= len(smaValues) || index >= len(prices) {
		return "Inconnu"
	}

	prevSMA := smaValues[index-1]
	currSMA := smaValues[index]
	prevPrice := prices[index-1]
	currPrice := prices[index]

	if math.IsNaN(prevSMA) || math.IsNaN(currSMA) || 
	   math.IsNaN(prevPrice) || math.IsNaN(currPrice) {
		return "Inconnu"
	}

	// Croisement haussier: SMA passe sous prix à dessus prix
	if prevSMA <= prevPrice && currSMA > currPrice {
		return "Haussier"
	}

	// Croisement baissier: SMA passe dessus prix à sous prix
	if prevSMA >= prevPrice && currSMA < currPrice {
		return "Baissier"
	}

	return "Aucun"
}

// GetDistanceFromPrice calcule la distance du SMA au prix en pourcentage
func (sma *SMATVStandard) GetDistanceFromPrice(smaValues []float64, prices []float64, index int) float64 {
	if index >= len(smaValues) || index >= len(prices) {
		return math.NaN()
	}

	smaVal := smaValues[index]
	price := prices[index]

	if math.IsNaN(smaVal) || math.IsNaN(price) || price == 0 {
		return math.NaN()
	}

	return (smaVal - price) / price * 100.0
}

// IsConverging vérifie si le SMA converge vers le prix
func (sma *SMATVStandard) IsConverging(smaValues []float64, prices []float64, index int, lookback int) bool {
	if lookback <= 1 || index < lookback || 
	   index >= len(smaValues) || index >= len(prices) {
		return false
	}

	// Calculer la distance actuelle vs précédente
	currentDist := math.Abs(sma.GetDistanceFromPrice(smaValues, prices, index))
	prevDist := math.Abs(sma.GetDistanceFromPrice(smaValues, prices, index-1))

	if math.IsNaN(currentDist) || math.IsNaN(prevDist) {
		return false
	}

	return currentDist < prevDist
}

// GetVariance calcule la variance sur une période
func (sma *SMATVStandard) GetVariance(smaValues []float64, index int, lookback int) float64 {
	if lookback <= 1 || index < lookback-1 || index >= len(smaValues) {
		return math.NaN()
	}

	// Extraire les valeurs valides
	values := make([]float64, 0, lookback)
	for i := index - lookback + 1; i <= index; i++ {
		if !math.IsNaN(smaValues[i]) {
			values = append(values, smaValues[i])
		}
	}

	if len(values) < 2 {
		return math.NaN()
	}

	// Calculer la moyenne
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	// Calculer la variance
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))

	return variance
}

// GetStandardDeviation calcule l'écart-type sur une période
func (sma *SMATVStandard) GetStandardDeviation(smaValues []float64, index int, lookback int) float64 {
	variance := sma.GetVariance(smaValues, index, lookback)
	
	if math.IsNaN(variance) || variance < 0 {
		return math.NaN()
	}

	return math.Sqrt(variance)
}

// IsInTrend vérifie si le SMA est en tendance (basé sur la pente)
func (sma *SMATVStandard) IsInTrend(smaValues []float64, index int, lookback int, threshold float64) bool {
	if lookback <= 1 || index < lookback || index >= len(smaValues) {
		return false
	}

	// Calculer la pente moyenne sur la période
	startVal := smaValues[index-lookback]
	endVal := smaValues[index]

	if math.IsNaN(startVal) || math.IsNaN(endVal) {
		return false
	}

	avgSlope := (endVal - startVal) / float64(lookback)

	return math.Abs(avgSlope) > threshold
}

// CalculateFromKlines calcule SMA depuis les klines (méthode utilitaire)
func (sma *SMATVStandard) CalculateFromKlines(klines []Kline, src func(Kline) float64) []float64 {
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

	return sma.Calculate(values)
}
