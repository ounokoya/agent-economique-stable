package indicators

import (
	"math"
)

// CHOPTVStandard - Implémentation Choppiness Index TradingView Standard
// Basé sur la documentation chop_tradingview_research.md
// Formule: CHOP = 100 * LOG10(SUM(ATR1,n) / (MaxHi - MinLo)) / LOG10(n)
type CHOPTVStandard struct {
	period int
}

// NewCHOPTVStandard crée une nouvelle instance CHOP TV Standard
func NewCHOPTVStandard(period int) *CHOPTVStandard {
	return &CHOPTVStandard{
		period: period,
	}
}

// Calculate calcule le CHOP selon les spécifications TradingView
func (chop *CHOPTVStandard) Calculate(high, low, close []float64) []float64 {
	n := len(high)
	if n != len(low) || n != len(close) {
		return nil
	}

	// Calculer ATR(1) (True Range)
	atr1 := chop.calculateATR1(high, low, close)

	// Calculer CHOP final
	chopValues := chop.calculateCHOPFinal(high, low, atr1)

	return chopValues
}

// calculateATR1 calcule l'ATR(1) ou True Range (méthode TradingView)
func (chop *CHOPTVStandard) calculateATR1(high, low, close []float64) []float64 {
	n := len(high)
	atr1 := make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			// Première bougie : pas de close précédent
			atr1[i] = high[i] - low[i]
		} else {
			// TR = MAX(H-L, |H-PrevClose|, |L-PrevClose|)
			range1 := high[i] - low[i]
			range2 := math.Abs(high[i] - close[i-1])
			range3 := math.Abs(low[i] - close[i-1])
			atr1[i] = math.Max(range1, math.Max(range2, range3))
		}
	}

	return atr1
}

// calculateCHOPFinal calcule le CHOP final (méthode TradingView)
func (chop *CHOPTVStandard) calculateCHOPFinal(high, low, atr1 []float64) []float64 {
	n := len(high)
	chopValues := make([]float64, n)

	// Initialiser avec NaN (TradingView standard)
	for i := range chopValues {
		chopValues[i] = math.NaN()
	}

	if chop.period <= 0 || n == 0 || chop.period > n {
		return chopValues
	}

	// Calculer CHOP pour chaque période
	for i := chop.period - 1; i < n; i++ {
		// Somme des ATR(1) sur la période [i-period+1 .. i]
		sumATR := 0.0
		for j := i - chop.period + 1; j <= i; j++ {
			sumATR += atr1[j]
		}

		// Range de prix sur la période [i-period+1 .. i]
		maxHigh := high[i]
		minLow := low[i]
		for j := i - chop.period + 1; j <= i; j++ {
			if high[j] > maxHigh {
				maxHigh = high[j]
			}
			if low[j] < minLow {
				minLow = low[j]
			}
		}
		priceRange := maxHigh - minLow

		// Calculer CHOP avec LOG10 base 10
		if priceRange != 0 && sumATR > 0 {
			ratio := sumATR / priceRange
			if ratio > 0 {
				chopValues[i] = 100.0 * math.Log10(ratio) / math.Log10(float64(chop.period))
			} else {
				chopValues[i] = 0.0
			}
		} else {
			chopValues[i] = 0.0
		}
	}

	return chopValues
}

// CalculateFromKlines calcule CHOP depuis les klines (méthode utilitaire)
func (chop *CHOPTVStandard) CalculateFromKlines(klines []Kline) []float64 {
	if len(klines) == 0 {
		return nil
	}

	// Extraire les données
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
	}

	return chop.Calculate(high, low, close)
}

// GetLastValue retourne la dernière valeur valide du CHOP
func (chop *CHOPTVStandard) GetLastValue(chopValues []float64) float64 {
	for i := len(chopValues) - 1; i >= 0; i-- {
		if !math.IsNaN(chopValues[i]) {
			return chopValues[i]
		}
	}
	return math.NaN()
}

// GetSignal retourne le signal de trading basé sur le CHOP
func (chop *CHOPTVStandard) GetSignal(chopValue float64) string {
	if math.IsNaN(chopValue) {
		return "Inconnu"
	}

	if chopValue > 61.8 {
		return "CHOPPY" // Marché sideways (seuil Fibonacci haut)
	} else if chopValue < 38.2 {
		return "TRENDING" // Marché en tendance (seuil Fibonacci bas)
	} else {
		return "NEUTRAL" // Zone de transition
	}
}

// IsChoppy vérifie si le marché est en phase choppy
func (chop *CHOPTVStandard) IsChoppy(chopValue float64) bool {
	return !math.IsNaN(chopValue) && chopValue > 61.8
}

// IsTrending vérifie si le marché est en phase de tendance
func (chop *CHOPTVStandard) IsTrending(chopValue float64) bool {
	return !math.IsNaN(chopValue) && chopValue < 38.2
}

// IsExitingChoppy vérifie la sortie de phase choppy
func (chop *CHOPTVStandard) IsExitingChoppy(chopValues []float64, index int) bool {
	if index <= 0 || index >= len(chopValues) {
		return false
	}

	prev := chopValues[index-1]
	curr := chopValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		prev > 61.8 && curr <= 61.8
}

// IsEnteringTrending vérifie l'entrée en phase de tendance
func (chop *CHOPTVStandard) IsEnteringTrending(chopValues []float64, index int) bool {
	if index <= 0 || index >= len(chopValues) {
		return false
	}

	prev := chopValues[index-1]
	curr := chopValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		prev >= 38.2 && curr < 38.2
}

// GetZone détermine la zone CHOP actuelle
func (chop *CHOPTVStandard) GetZone(chopValue float64) string {
	if math.IsNaN(chopValue) {
		return "Inconnue"
	}

	if chopValue > 80 {
		return "Très Choppy"
	} else if chopValue > 61.8 {
		return "Choppy"
	} else if chopValue > 50 {
		return "Neutre-Haut"
	} else if chopValue > 38.2 {
		return "Neutre-Bas"
	} else if chopValue > 20 {
		return "Trending"
	} else {
		return "Très Trending"
	}
}

// GetRegimeChange détecte les changements de régime (choppy ↔ trending)
func (chop *CHOPTVStandard) GetRegimeChange(chopValues []float64, lookback int) string {
	if lookback <= 1 || len(chopValues) < lookback {
		return "Insuffisant"
	}

	// Analyser les dernières 'lookback' périodes
	recentCHOP := chopValues[len(chopValues)-lookback:]
	
	// Vérifier si nous avons des valeurs valides
	validCHOP := make([]float64, 0)
	for _, v := range recentCHOP {
		if !math.IsNaN(v) {
			validCHOP = append(validCHOP, v)
		}
	}

	if len(validCHOP) < 2 {
		return "Insuffisant"
	}

	// Transition choppy → trending
	wasChoppy := validCHOP[0] > 61.8
	isTrending := validCHOP[len(validCHOP)-1] < 38.2

	// Transition trending → choppy
	wasTrending := validCHOP[0] < 38.2
	isChoppy := validCHOP[len(validCHOP)-1] > 61.8

	if wasChoppy && isTrending {
		return "Choppy → Trending"
	} else if wasTrending && isChoppy {
		return "Trending → Choppy"
	} else {
		return "Stable"
	}
}

// GetStrength mesure l'intensité du régime actuel
func (chop *CHOPTVStandard) GetStrength(chopValue float64) float64 {
	if math.IsNaN(chopValue) {
		return math.NaN()
	}

	if chopValue > 61.8 {
		// Force du choppy: 0-100 (plus c'est élevé, plus c'est choppy)
		return math.Min(100.0, (chopValue-61.8)/38.2*100)
	} else if chopValue < 38.2 {
		// Force de la tendance: 0-100 (plus c'est bas, plus c'est trendy)
		return math.Min(100.0, (38.2-chopValue)/38.2*100)
	} else {
		// Zone neutre: force faible
		return 0.0
	}
}

// GetCustomThresholds permet d'utiliser des seuils personnalisés
func (chop *CHOPTVStandard) GetCustomThresholds(chopValue, upperThreshold, lowerThreshold float64) string {
	if math.IsNaN(chopValue) {
		return "Inconnu"
	}

	if chopValue > upperThreshold {
		return "CHOPPY (Custom)"
	} else if chopValue < lowerThreshold {
		return "TRENDING (Custom)"
	} else {
		return "NEUTRAL (Custom)"
	}
}
