package indicators

import (
	"math"
)

// DMITVStandard - Implémentation DMI TradingView Standard
// Composants: ADX (force), +DI (direction haussière), -DI (direction baissière)
type DMITVStandard struct {
	period   int
	periodDI int
}

// NewDMITVStandard crée une nouvelle instance DMI TV Standard
func NewDMITVStandard(period int) *DMITVStandard {
	return &DMITVStandard{
		period:   period,   // Pour ADX
		periodDI: period,   // Pour +DI et -DI (généralement identique)
	}
}

// Calculate calcule le DMI selon la formule TradingView standard
func (dmi *DMITVStandard) Calculate(high, low, close []float64) (plusDI, minusDI, adx []float64) {
	n := len(high)
	if n != len(low) || n != len(close) {
		return nil, nil, nil
	}

	// Calculer True Range
	tr := dmi.calculateTR(high, low, close)

	// Calculer Directional Movement
	plusDM, minusDM := dmi.calculateDM(high, low)

	// Appliquer Wilder's Smoothing
	atr := dmi.calculateWilderSmoothing(tr, dmi.periodDI)
	smoothedPlusDM := dmi.calculateWilderSmoothing(plusDM, dmi.periodDI)
	smoothedMinusDM := dmi.calculateWilderSmoothing(minusDM, dmi.periodDI)

	// Calculer +DI et -DI
	plusDI = make([]float64, n)
	minusDI = make([]float64, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(atr[i]) && atr[i] != 0 {
			plusDI[i] = 100 * smoothedPlusDM[i] / atr[i]
			minusDI[i] = 100 * smoothedMinusDM[i] / atr[i]
		} else {
			plusDI[i] = math.NaN()
			minusDI[i] = math.NaN()
		}
	}

	// Calculer DX puis ADX
	dx := dmi.calculateDX(plusDI, minusDI)
	adx = dmi.calculateWilderSmoothing(dx, dmi.period)

	return plusDI, minusDI, adx
}

// calculateTR calcule True Range (méthode TradingView)
func (dmi *DMITVStandard) calculateTR(high, low, close []float64) []float64 {
	n := len(high)
	tr := make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			tr[i] = high[i] - low[i]
		} else {
			highLow := high[i] - low[i]
			highClosePrev := math.Abs(high[i] - close[i-1])
			lowClosePrev := math.Abs(low[i] - close[i-1])
			
			tr[i] = math.Max(highLow, math.Max(highClosePrev, lowClosePrev))
		}
	}

	return tr
}

// calculateDM calcule Directional Movement (méthode TradingView)
func (dmi *DMITVStandard) calculateDM(high, low []float64) (plusDM, minusDM []float64) {
	n := len(high)
	plusDM = make([]float64, n)
	minusDM = make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			plusDM[i] = 0
			minusDM[i] = 0
		} else {
			upMove := high[i] - high[i-1]
			downMove := low[i-1] - low[i]

			if upMove > downMove && upMove > 0 {
				plusDM[i] = upMove
			} else {
				plusDM[i] = 0
			}

			if downMove > upMove && downMove > 0 {
				minusDM[i] = downMove
			} else {
				minusDM[i] = 0
			}
		}
	}

	return plusDM, minusDM
}

// calculateWilderSmoothing applique Wilder's Smoothing (méthode TradingView)
func (dmi *DMITVStandard) calculateWilderSmoothing(values []float64, period int) []float64 {
	rma := NewRMATVStandard(period)
	return rma.Calculate(values)
}

// CalculateDX calcule le Directional Index (DX) - méthode publique
func (dmi *DMITVStandard) CalculateDX(plusDI, minusDI []float64) []float64 {
	return dmi.calculateDX(plusDI, minusDI)
}

// calculateDX calcule Directional Index (méthode TradingView)
func (dmi *DMITVStandard) calculateDX(plusDI, minusDI []float64) []float64 {
	n := len(plusDI)
	dx := make([]float64, n)

	for i := 0; i < n; i++ {
		if !math.IsNaN(plusDI[i]) && !math.IsNaN(minusDI[i]) {
			sum := plusDI[i] + minusDI[i]
			if sum != 0 {
				dx[i] = 100 * math.Abs(plusDI[i]-minusDI[i]) / sum
			} else {
				dx[i] = 0
			}
		} else {
			dx[i] = math.NaN()
		}
	}

	return dx
}

// GetLastValues retourne les dernières valeurs valides
func (dmi *DMITVStandard) GetLastValues(plusDI, minusDI, adx []float64) (float64, float64, float64) {
	getLastValid := func(values []float64) float64 {
		for i := len(values) - 1; i >= 0; i-- {
			if !math.IsNaN(values[i]) {
				return values[i]
			}
		}
		return math.NaN()
	}

	return getLastValid(plusDI), getLastValid(minusDI), getLastValid(adx)
}

// GetTrendStrength retourne la force de la tendance
func (dmi *DMITVStandard) GetTrendStrength(adx float64) string {
	if math.IsNaN(adx) {
		return "Inconnu"
	}
	
	if adx > 25 {
		return "Forte"
	} else if adx > 20 {
		return "Modérée"
	} else if adx > 15 {
		return "Faible"
	} else {
		return "Nulle"
	}
}

// GetTrendDirection retourne la direction de la tendance
func (dmi *DMITVStandard) GetTrendDirection(plusDI, minusDI float64) string {
	if math.IsNaN(plusDI) || math.IsNaN(minusDI) {
		return "Inconnue"
	}
	
	if plusDI > minusDI {
		return "Haussière"
	} else if minusDI > plusDI {
		return "Baissière"
	} else {
		return "Neutre"
	}
}

// IsStrongTrend vérifie si la tendance est forte (ADX > 25)
func (dmi *DMITVStandard) IsStrongTrend(adx float64) bool {
	return !math.IsNaN(adx) && adx > 25
}

// IsBullish vérifie si la tendance est haussière
func (dmi *DMITVStandard) IsBullish(plusDI, minusDI float64) bool {
	return !math.IsNaN(plusDI) && !math.IsNaN(minusDI) && plusDI > minusDI
}

// IsBearish vérifie si la tendance est baissière
func (dmi *DMITVStandard) IsBearish(plusDI, minusDI float64) bool {
	return !math.IsNaN(plusDI) && !math.IsNaN(minusDI) && minusDI > plusDI
}

// GetSignal retourne le signal de trading actuel
func (dmi *DMITVStandard) GetSignal(plusDI, minusDI, adx float64) string {
	if !dmi.IsStrongTrend(adx) {
		return "Pas de trend"
	}
	
	if dmi.IsBullish(plusDI, minusDI) {
		return "ACHAT"
	} else if dmi.IsBearish(plusDI, minusDI) {
		return "VENTE"
	} else {
		return "NEUTRE"
	}
}
