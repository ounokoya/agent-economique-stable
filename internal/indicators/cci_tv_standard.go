package indicators

import (
	"math"
)

// CCITVStandard - Implémentation CCI TradingView Standard
// Formule exacte : CCI = (TP - SMA_TP) / (0.015 × Mean Deviation)
type CCITVStandard struct {
	period   int
	constant float64
}

// NewCCITVStandard crée une nouvelle instance CCI TV Standard
func NewCCITVStandard(period int) *CCITVStandard {
	return &CCITVStandard{
		period:   period,
		constant: 0.015, // Constante TradingView officielle
	}
}

// Calculate calcule le CCI selon la formule TradingView standard
func (cci *CCITVStandard) Calculate(high, low, close []float64) []float64 {
	n := len(high)
	if n != len(low) || n != len(close) {
		return nil
	}

	// Calculer Typical Price (TP) = (High + Low + Close) / 3
	tp := make([]float64, n)
	for i := 0; i < n; i++ {
		tp[i] = (high[i] + low[i] + close[i]) / 3.0
	}

	// Calculer SMA du TP
	sma := cci.calculateSMA(tp)

	// Calculer Mean Deviation
	meanDev := cci.calculateMeanDeviation(tp, sma)

	// Calculer CCI final
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		if i < cci.period-1 {
			result[i] = math.NaN() // TradingView retourne na pour les premières barres
		} else if meanDev[i] == 0 {
			result[i] = 0.0 // Éviter division par zéro
		} else {
			result[i] = (tp[i] - sma[i]) / (cci.constant * meanDev[i])
		}
	}

	return result
}

// calculateSMA calcule Simple Moving Average (méthode TradingView)
func (cci *CCITVStandard) calculateSMA(values []float64) []float64 {
	sma := NewSMATVStandard(cci.period)
	return sma.Calculate(values)
}

// calculateMeanDeviation calcule la déviation moyenne (méthode TradingView)
func (cci *CCITVStandard) calculateMeanDeviation(values, sma []float64) []float64 {
	n := len(values)
	meanDev := make([]float64, n)

	for i := 0; i < n; i++ {
		if i < cci.period-1 {
			meanDev[i] = math.NaN()
		} else {
			sum := 0.0
			for j := i - cci.period + 1; j <= i; j++ {
				sum += math.Abs(values[j] - sma[i])
			}
			meanDev[i] = sum / float64(cci.period)
		}
	}

	return meanDev
}

// GetLastValue retourne la dernière valeur CCI valide
func (cci *CCITVStandard) GetLastValue(cciValues []float64) float64 {
	for i := len(cciValues) - 1; i >= 0; i-- {
		if !math.IsNaN(cciValues[i]) {
			return cciValues[i]
		}
	}
	return math.NaN()
}

// GetZone retourne la zone actuelle (Surachat/Survente/Neutre)
func (cci *CCITVStandard) GetZone(value float64) string {
	if math.IsNaN(value) {
		return "Inconnu"
	}
	
	if value > 100 {
		return "Surachat"
	} else if value > 50 {
		return "Haussier"
	} else if value > -50 {
		return "Neutre"
	} else if value > -100 {
		return "Baissier"
	} else {
		return "Survente"
	}
}

// IsOverbought vérifie si le CCI est en zone de surachat
func (cci *CCITVStandard) IsOverbought(value float64) bool {
	return !math.IsNaN(value) && value > 100
}

// IsOversold vérifie si le CCI est en zone de survente
func (cci *CCITVStandard) IsOversold(value float64) bool {
	return !math.IsNaN(value) && value < -100
}

// IsBullish vérifie si le CCI est en tendance haussière
func (cci *CCITVStandard) IsBullish(value float64) bool {
	return !math.IsNaN(value) && value > 0
}

// IsBearish vérifie si le CCI est en tendance baissière
func (cci *CCITVStandard) IsBearish(value float64) bool {
	return !math.IsNaN(value) && value < 0
}
