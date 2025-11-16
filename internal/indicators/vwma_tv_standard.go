package indicators

import (
	"math"
)

// VWMATVStandard - Implémentation Volume-Weighted Moving Average TradingView Standard
// Basé sur la documentation vwma_tradingview_research.md
// Formule: VWMA = Σ(Close × Volume) / Σ(Volume)
type VWMATVStandard struct {
	period int
}

// NewVWMATVStandard crée une nouvelle instance VWMA TV Standard
func NewVWMATVStandard(period int) *VWMATVStandard {
	return &VWMATVStandard{
		period: period,
	}
}

// Calculate calcule le VWMA selon les spécifications TradingView
func (vwma *VWMATVStandard) Calculate(close, volume []float64) []float64 {
	n := len(close)
	if n != len(volume) {
		return nil
	}

	// Initialiser avec NaN (TradingView standard)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	if vwma.period <= 0 || n == 0 || vwma.period > n {
		return result
	}

	// Calculer VWMA pour chaque période
	for i := vwma.period - 1; i < n; i++ {
		var sumWeightedPrice, sumVolume float64
		
		// Calculer les sommes sur la période [i-period+1 .. i]
		for j := i - vwma.period + 1; j <= i; j++ {
			sumWeightedPrice += close[j] * volume[j]
			sumVolume += volume[j]
		}
		
		// Calculer VWMA avec gestion division par zéro
		if sumVolume != 0 {
			result[i] = sumWeightedPrice / sumVolume
		} else {
			result[i] = math.NaN()
		}
	}

	return result
}

// CalculateFromKlines calcule VWMA depuis les klines (méthode utilitaire)
func (vwma *VWMATVStandard) CalculateFromKlines(klines []Kline) []float64 {
	if len(klines) == 0 {
		return nil
	}

	// Extraire les données
	close := make([]float64, len(klines))
	volume := make([]float64, len(klines))

	for i, k := range klines {
		close[i] = k.Close
		volume[i] = k.Volume
	}

	return vwma.Calculate(close, volume)
}

// GetLastValue retourne la dernière valeur valide du VWMA
func (vwma *VWMATVStandard) GetLastValue(vwmaValues []float64) float64 {
	for i := len(vwmaValues) - 1; i >= 0; i-- {
		if !math.IsNaN(vwmaValues[i]) {
			return vwmaValues[i]
		}
	}
	return math.NaN()
}

// GetSignal retourne le signal de trading basé sur le VWMA
func (vwma *VWMATVStandard) GetSignal(vwmaValue, currentPrice float64) string {
	if math.IsNaN(vwmaValue) || math.IsNaN(currentPrice) {
		return "Inconnu"
	}

	if currentPrice > vwmaValue {
		return "AU-DESSUS" // Prix au-dessus du VWMA (haussier)
	} else if currentPrice < vwmaValue {
		return "EN-DESSOUS" // Prix en-dessous du VWMA (baissier)
	} else {
		return "SUR VWMA" // Prix égal au VWMA
	}
}

// IsAbove vérifie si le prix est au-dessus du VWMA
func (vwma *VWMATVStandard) IsAbove(vwmaValue, currentPrice float64) bool {
	return !math.IsNaN(vwmaValue) && !math.IsNaN(currentPrice) && currentPrice > vwmaValue
}

// IsBelow vérifie si le prix est en-dessous du VWMA
func (vwma *VWMATVStandard) IsBelow(vwmaValue, currentPrice float64) bool {
	return !math.IsNaN(vwmaValue) && !math.IsNaN(currentPrice) && currentPrice < vwmaValue
}

// IsCrossoverAbove vérifie le croisement haussier prix/VWMA
func (vwma *VWMATVStandard) IsCrossoverAbove(vwmaValues, prices []float64, index int) bool {
	if index <= 0 || index >= len(vwmaValues) || index >= len(prices) {
		return false
	}

	prev := vwmaValues[index-1]
	curr := vwmaValues[index]
	prevPrice := prices[index-1]
	currPrice := prices[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		!math.IsNaN(prevPrice) && !math.IsNaN(currPrice) &&
		prevPrice <= prev && currPrice > curr
}

// IsCrossoverBelow vérifie le croisement baissier prix/VWMA
func (vwma *VWMATVStandard) IsCrossoverBelow(vwmaValues, prices []float64, index int) bool {
	if index <= 0 || index >= len(vwmaValues) || index >= len(prices) {
		return false
	}

	prev := vwmaValues[index-1]
	curr := vwmaValues[index]
	prevPrice := prices[index-1]
	currPrice := prices[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		!math.IsNaN(prevPrice) && !math.IsNaN(currPrice) &&
		prevPrice >= prev && currPrice < curr
}

// GetDeviation calcule la déviation en pourcentage du prix par rapport au VWMA
func (vwma *VWMATVStandard) GetDeviation(vwmaValue, currentPrice float64) float64 {
	if math.IsNaN(vwmaValue) || math.IsNaN(currentPrice) || vwmaValue == 0 {
		return math.NaN()
	}
	return (currentPrice - vwmaValue) / vwmaValue * 100.0
}

// GetTrendDirection détermine la direction de la tendance du VWMA
func (vwma *VWMATVStandard) GetTrendDirection(vwmaValues []float64, lookback int) string {
	if lookback <= 1 || len(vwmaValues) < lookback {
		return "Insuffisant"
	}

	// Analyser les dernières 'lookback' périodes
	recentVWMA := vwmaValues[len(vwmaValues)-lookback:]
	
	// Vérifier si nous avons des valeurs valides
	validVWMA := make([]float64, 0)
	for _, v := range recentVWMA {
		if !math.IsNaN(v) {
			validVWMA = append(validVWMA, v)
		}
	}

	if len(validVWMA) < 2 {
		return "Insuffisant"
	}

	// Tendance haussière: VWMA monte
	vwmaUp := validVWMA[len(validVWMA)-1] > validVWMA[0]
	
	// Tendance baissière: VWMA descend
	vwmaDown := validVWMA[len(validVWMA)-1] < validVWMA[0]

	if vwmaUp {
		return "Haussière"
	} else if vwmaDown {
		return "Baissière"
	} else {
		return "Neutre"
	}
}
