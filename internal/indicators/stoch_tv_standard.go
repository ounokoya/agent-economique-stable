package indicators

import (
	"math"
)

// StochTVStandard - Implémentation Stochastic TradingView Standard
// Composants: %K (14), %K Smoothed (3), %D (3)
type StochTVStandard struct {
	periodK  int
	smoothK  int
	periodD  int
}

// NewStochTVStandard crée une nouvelle instance Stochastic TV Standard
func NewStochTVStandard(periodK, smoothK, periodD int) *StochTVStandard {
	return &StochTVStandard{
		periodK: periodK,
		smoothK: smoothK,
		periodD: periodD,
	}
}

// Calculate calcule le Stochastic selon la formule TradingView standard
func (stoch *StochTVStandard) Calculate(high, low, close []float64) (k, d []float64) {
	n := len(high)
	if n != len(low) || n != len(close) {
		return nil, nil
	}

	// Calculer %K brut
	kRaw := stoch.calculateKRaw(high, low, close)

	// Lisser %K (slow stochastic)
	kSmoothed := stoch.calculateSMA(kRaw, stoch.smoothK)

	// Calculer %D (signal line)
	d = stoch.calculateSMA(kSmoothed, stoch.periodD)

	return kSmoothed, d
}

// calculateKRaw calcule %K brut (méthode TradingView précise)
func (stoch *StochTVStandard) calculateKRaw(high, low, close []float64) []float64 {
	n := len(high)
	kRaw := make([]float64, n)

	for i := 0; i < n; i++ {
		if i < stoch.periodK-1 {
			kRaw[i] = math.NaN()  // TradingView retourne na pour les premières barres
		} else {
			// Trouver Highest High et Lowest Low sur la période EXACTE
			highestHigh := high[i-stoch.periodK+1]  // ✅ Initialisation correcte
			lowestLow := low[i-stoch.periodK+1]     // ✅ Initialisation correcte
			
			for j := i - stoch.periodK + 1; j <= i; j++ {
				if high[j] > highestHigh {
					highestHigh = high[j]
				}
				if low[j] < lowestLow {
					lowestLow = low[j]
				}
			}
			
			// Calculer %K brut: 100 × (Close - Lowest Low) / (Highest High - Lowest Low)
			// ✅ Cas particulier TradingView : division par zéro = 50
			if highestHigh == lowestLow {
				kRaw[i] = 50.0  // Valeur neutre exacte TradingView
			} else {
				kRaw[i] = 100.0 * (close[i] - lowestLow) / (highestHigh - lowestLow)
			}
		}
	}

	return kRaw
}

// calculateSMA calcule Simple Moving Average (implémentation TradingView précise)
func (stoch *StochTVStandard) calculateSMA(values []float64, period int) []float64 {
	sma := NewSMATVStandard(period)
	return sma.Calculate(values)
}

// GetLastValues retourne les dernières valeurs valides
func (stoch *StochTVStandard) GetLastValues(k, d []float64) (float64, float64) {
	getLastValid := func(values []float64) float64 {
		for i := len(values) - 1; i >= 0; i-- {
			if !math.IsNaN(values[i]) {
				return values[i]
			}
		}
		return math.NaN()
	}

	return getLastValid(k), getLastValid(d)
}

// GetZone retourne la zone actuelle du Stochastic
func (stoch *StochTVStandard) GetZone(k, d float64) string {
	if math.IsNaN(k) || math.IsNaN(d) {
		return "Inconnue"
	}
	
	if k > 80 && d > 80 {
		return "Surachat Fort"
	} else if k > 80 {
		return "Surachat %K"
	} else if d > 80 {
		return "Surachat %D"
	} else if k < 20 && d < 20 {
		return "Survente Forte"
	} else if k < 20 {
		return "Survente %K"
	} else if d < 20 {
		return "Survente %D"
	} else if k > 50 && d > 50 {
		return "Haussière"
	} else if k < 50 && d < 50 {
		return "Baissière"
	} else {
		return "Neutre"
	}
}

// GetSignal retourne le signal de trading actuel
func (stoch *StochTVStandard) GetSignal(k, d float64) string {
	if math.IsNaN(k) || math.IsNaN(d) {
		return "Inconnu"
	}
	
	if k > 80 && d > 80 {
		return "SURACHAT"
	} else if k < 20 && d < 20 {
		return "SURVENTE"
	} else if k > d && k > 50 {
		return "ACHAT"
	} else if d > k && d < 50 {
		return "VENTE"
	} else if k > d {
		return "HAUSSIER"
	} else if d > k {
		return "BAISSIER"
	} else {
		return "NEUTRE"
	}
}

// IsOverbought vérifie si le Stochastic est en surachat
func (stoch *StochTVStandard) IsOverbought(k, d float64, threshold float64) bool {
	return !math.IsNaN(k) && !math.IsNaN(d) && k > threshold && d > threshold
}

// IsOversold vérifie si le Stochastic est en survente
func (stoch *StochTVStandard) IsOversold(k, d float64, threshold float64) bool {
	return !math.IsNaN(k) && !math.IsNaN(d) && k < threshold && d < threshold
}

// IsBullishCrossover vérifie croisement haussier %K > %D
func (stoch *StochTVStandard) IsBullishCrossover(k, d []float64, index int) bool {
	if index <= 0 || index >= len(k) || index >= len(d) {
		return false
	}
	
	prevK := k[index-1]
	currK := k[index]
	prevD := d[index-1]
	currD := d[index]
	
	return prevK <= prevD && currK > currD
}

// IsBearishCrossover vérifie croisement baissier %K < %D
func (stoch *StochTVStandard) IsBearishCrossover(k, d []float64, index int) bool {
	if index <= 0 || index >= len(k) || index >= len(d) {
		return false
	}
	
	prevK := k[index-1]
	currK := k[index]
	prevD := d[index-1]
	currD := d[index]
	
	return prevK >= prevD && currK < currD
}

// GetMomentumStrength retourne la force du momentum
func (stoch *StochTVStandard) GetMomentumStrength(k float64) string {
	if math.IsNaN(k) {
		return "Inconnu"
	}
	
	if k > 90 {
		return "Extrême Haussier"
	} else if k > 80 {
		return "Très Haussier"
	} else if k > 70 {
		return "Haussier"
	} else if k > 60 {
		return "Modéré Haussier"
	} else if k > 40 {
		return "Neutre"
	} else if k > 30 {
		return "Modéré Baissier"
	} else if k > 20 {
		return "Baissier"
	} else if k > 10 {
		return "Très Baissier"
	} else {
		return "Extrême Baissier"
	}
}

// GetDivergenceType détecte les divergences potentielles
func (stoch *StochTVStandard) GetDivergenceType(prices, k []float64, lookback int) string {
	if lookback <= 0 || len(prices) < lookback || len(k) < lookback {
		return "Insuffisant"
	}
	
	// Divergence baissière: prix plus haut mais Stochastic plus bas
	if prices[len(prices)-1] > prices[len(prices)-lookback] && 
	   k[len(k)-1] < k[len(k)-lookback] {
		return "Baissière"
	}
	
	// Divergence haussière: prix plus bas mais Stochastic plus haut
	if prices[len(prices)-1] < prices[len(prices)-lookback] && 
	   k[len(k)-1] > k[len(k)-lookback] {
		return "Haussière"
	}
	
	return "Aucune"
}

// GetPositionInRange retourne la position dans le range 0-100
func (stoch *StochTVStandard) GetPositionInRange(k float64) string {
	if math.IsNaN(k) {
		return "Inconnue"
	}
	
	if k >= 80 {
		return "Zone Supérieure (80-100)"
	} else if k >= 60 {
		return "Zone Haute (60-80)"
	} else if k >= 40 {
		return "Zone Moyenne (40-60)"
	} else if k >= 20 {
		return "Zone Basse (20-40)"
	} else {
		return "Zone Inférieure (0-20)"
	}
}
