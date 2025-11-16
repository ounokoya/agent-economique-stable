package indicators

import (
	"math"
)

// HMATVStandard - Hull Moving Average conforme TradingView
// Basé sur la formule : Integer(SquareRoot(Period)) WMA [2 x Integer(Period/2) WMA(Price) - Period WMA(Price)]
// Source : https://alanhull.com/hull-moving-average
type HMATVStandard struct {
	period int
}

// NewHMATVStandard crée une nouvelle instance HMA TV Standard
func NewHMATVStandard(period int) *HMATVStandard {
	return &HMATVStandard{period: period}
}

// Calculate calcule l'HMA selon la formule TradingView
func (hma *HMATVStandard) Calculate(prices []float64) []float64 {
	n := len(prices)
	result := make([]float64, n)
	
	if n < hma.period {
		return result
	}
	
	// Calculer n/2 arrondi à l'entier inférieur (selon formule Alan Hull)
	halfPeriod := hma.period / 2
	
	// Calculer sqrt(n) arrondi à l'entier (selon formule Alan Hull)
	sqrtPeriod := int(math.Sqrt(float64(hma.period)))
	
	// Calculer WMA sur n/2
	wmaHalf := hma.calculateWMA(prices, halfPeriod)
	
	// Calculer WMA sur n
	wmaFull := hma.calculateWMA(prices, hma.period)
	
	// Calculer la série intermédiaire : (2 × WMA(n/2)) - WMA(n)
	intermediate := make([]float64, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(wmaHalf[i]) && !math.IsNaN(wmaFull[i]) {
			intermediate[i] = (2 * wmaHalf[i]) - wmaFull[i]
		}
	}
	
	// Calculer HMA final sur sqrt(n)
	result = hma.calculateWMA(intermediate, sqrtPeriod)
	
	return result
}

// calculateWMA calcule la Weighted Moving Average
// Formule : (P1 + 2*P2 + 3*P3 + ... + n*Pn) / K où K = n(n+1)/2 et Pn est le plus récent
func (hma *HMATVStandard) calculateWMA(prices []float64, period int) []float64 {
	n := len(prices)
	result := make([]float64, n)
	
	// Calculer K = n(n+1)/2 une seule fois
	k := float64(period * (period + 1)) / 2.0
	
	for i := period - 1; i < n; i++ {
		var sum float64
		
		// Calculer la somme pondérée comme Python pyti
		for j := 0; j < period; j++ {
			// data[idx + period_idx] * (period_idx + 1)
			price := prices[i - period + 1 + j]
			weight := float64(j + 1)  // 1, 2, 3, ..., period
			sum += price * weight
		}
		
		result[i] = sum / k
	}
	
	return result
}

// GetLastValue retourne la dernière valeur valide de HMA
func (hma *HMATVStandard) GetLastValue(values []float64) float64 {
	for i := len(values) - 1; i >= 0; i-- {
		if !math.IsNaN(values[i]) {
			return values[i]
		}
	}
	return math.NaN()
}

// GetSignal détermine le signal de trading basé sur HMA
func (hma *HMATVStandard) GetSignal(prices, hmaValues []float64, index int) string {
	if index >= len(prices) || index >= len(hmaValues) || math.IsNaN(hmaValues[index]) {
		return "Inconnu"
	}
	
	currentPrice := prices[index]
	currentHMA := hmaValues[index]
	
	// Signaux de croisement
	if index > 0 && !math.IsNaN(hmaValues[index-1]) {
		prevPrice := prices[index-1]
		prevHMA := hmaValues[index-1]
		
		// Croisement haussier : prix passe au-dessus de HMA
		if prevPrice <= prevHMA && currentPrice > currentHMA {
			return "ACHAT"
		}
		
		// Croisement baissier : prix passe en dessous de HMA
		if prevPrice >= prevHMA && currentPrice < currentHMA {
			return "VENTE"
		}
	}
	
	// Position par rapport à HMA
	if currentPrice > currentHMA {
		return "Haussier"
	} else if currentPrice < currentHMA {
		return "Baissier"
	} else {
		return "Neutre"
	}
}

// IsTrendingUp vérifie si la tendance est haussière
func (hma *HMATVStandard) IsTrendingUp(hmaValues []float64, index int) bool {
	if index < 1 || index >= len(hmaValues) {
		return false
	}
	
	return hmaValues[index] > hmaValues[index-1] && !math.IsNaN(hmaValues[index]) && !math.IsNaN(hmaValues[index-1])
}

// IsTrendingDown vérifie si la tendance est baissière
func (hma *HMATVStandard) IsTrendingDown(hmaValues []float64, index int) bool {
	if index < 1 || index >= len(hmaValues) {
		return false
	}
	
	return hmaValues[index] < hmaValues[index-1] && !math.IsNaN(hmaValues[index]) && !math.IsNaN(hmaValues[index-1])
}

// GetSlope calcule la pente de HMA
func (hma *HMATVStandard) GetSlope(hmaValues []float64, index int) float64 {
	if index < 1 || index >= len(hmaValues) {
		return 0
	}
	
	if math.IsNaN(hmaValues[index]) || math.IsNaN(hmaValues[index-1]) {
		return 0
	}
	
	return hmaValues[index] - hmaValues[index-1]
}

// GetSlopeAngle calcule l'angle de la pente en degrés
func (hma *HMATVStandard) GetSlopeAngle(hmaValues []float64, index int) float64 {
	slope := hma.GetSlope(hmaValues, index)
	return math.Atan(slope) * 180 / math.Pi
}

// IsCrossoverUp détecte un croisement haussier prix/HMA
func (hma *HMATVStandard) IsCrossoverUp(prices, hmaValues []float64, index int) bool {
	if index < 1 || index >= len(prices) || index >= len(hmaValues) {
		return false
	}
	
	prevPrice := prices[index-1]
	currentPrice := prices[index]
	prevHMA := hmaValues[index-1]
	currentHMA := hmaValues[index]
	
	return prevPrice <= prevHMA && currentPrice > currentHMA &&
		!math.IsNaN(prevHMA) && !math.IsNaN(currentHMA)
}

// IsCrossoverDown détecte un croisement baissier prix/HMA
func (hma *HMATVStandard) IsCrossoverDown(prices, hmaValues []float64, index int) bool {
	if index < 1 || index >= len(prices) || index >= len(hmaValues) {
		return false
	}
	
	prevPrice := prices[index-1]
	currentPrice := prices[index]
	prevHMA := hmaValues[index-1]
	currentHMA := hmaValues[index]
	
	return prevPrice >= prevHMA && currentPrice < currentHMA &&
		!math.IsNaN(prevHMA) && !math.IsNaN(currentHMA)
}

// GetDivergenceType détecte les divergences prix/HMA
func (hma *HMATVStandard) GetDivergenceType(prices, hmaValues []float64, lookback int) string {
	if lookback < 2 || len(prices) < lookback+1 || len(hmaValues) < lookback+1 {
		return "Aucune"
	}
	
	currentIdx := len(prices) - 1
	lookbackIdx := currentIdx - lookback
	
	// Divergence haussière : prix plus bas mais HMA plus haut
	if prices[currentIdx] < prices[lookbackIdx] && hmaValues[currentIdx] > hmaValues[lookbackIdx] {
		return "Haussière"
	}
	
	// Divergence baissière : prix plus haut mais HMA plus bas
	if prices[currentIdx] > prices[lookbackIdx] && hmaValues[currentIdx] < hmaValues[lookbackIdx] {
		return "Baissière"
	}
	
	return "Aucune"
}
