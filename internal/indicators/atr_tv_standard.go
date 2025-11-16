package indicators

import (
	"math"
)

// ATRTVStandard - Implémentation Average True Range TradingView Standard
// Basé sur la documentation atr_tradingview_research.md
// Formule: ATR = RMA(TR, length) où TR = MAX(H-L, |H-PrevClose|, |L-PrevClose|)
type ATRTVStandard struct {
	period int
}

// NewATRTVStandard crée une nouvelle instance ATR TV Standard
func NewATRTVStandard(period int) *ATRTVStandard {
	return &ATRTVStandard{
		period: period,
	}
}

// Calculate calcule l'ATR selon les spécifications TradingView
func (atr *ATRTVStandard) Calculate(high, low, close []float64) []float64 {
	n := len(high)
	if n != len(low) || n != len(close) {
		return nil
	}

	// Calculer True Range
	tr := atr.calculateTrueRange(high, low, close)

	// Appliquer RMA (Wilder's Smoothing)
	atrValues := atr.calculateRMA(tr, atr.period)

	return atrValues
}

// calculateTrueRange calcule le True Range (méthode TradingView)
func (atr *ATRTVStandard) calculateTrueRange(high, low, close []float64) []float64 {
	n := len(high)
	tr := make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			// Première bougie : pas de close précédent
			tr[i] = high[i] - low[i]
		} else {
			// TR = MAX(H-L, |H-PrevClose|, |L-PrevClose|)
			range1 := high[i] - low[i]
			range2 := math.Abs(high[i] - close[i-1])
			range3 := math.Abs(low[i] - close[i-1])
			tr[i] = math.Max(range1, math.Max(range2, range3))
		}
	}

	return tr
}

// calculateRMA calcule le RMA (Relative Moving Average - Wilder's Smoothing)
func (atr *ATRTVStandard) calculateRMA(values []float64, period int) []float64 {
	n := len(values)
	rma := make([]float64, n)

	// Initialiser avec NaN (TradingView standard)
	for i := range rma {
		rma[i] = math.NaN()
	}

	if period <= 0 || n == 0 || period > n {
		return rma
	}

	// Seed avec SMA (première valeur RMA)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	rma[period-1] = sum / float64(period)

	// Calcul RMA récursif (formule Wilder)
	for i := period; i < n; i++ {
		rma[i] = (rma[i-1]*float64(period-1) + values[i]) / float64(period)
	}

	return rma
}

// CalculateFromKlines calcule ATR depuis les klines (méthode utilitaire)
func (atr *ATRTVStandard) CalculateFromKlines(klines []Kline) []float64 {
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

	return atr.Calculate(high, low, close)
}

// GetLastValue retourne la dernière valeur valide de l'ATR
func (atr *ATRTVStandard) GetLastValue(atrValues []float64) float64 {
	for i := len(atrValues) - 1; i >= 0; i-- {
		if !math.IsNaN(atrValues[i]) {
			return atrValues[i]
		}
	}
	return math.NaN()
}

// GetSignal retourne le signal de trading basé sur l'ATR
func (atr *ATRTVStandard) GetSignal(atrValue, currentPrice float64) string {
	if math.IsNaN(atrValue) || math.IsNaN(currentPrice) {
		return "Inconnu"
	}

	// ATR en pourcentage du prix
	atrPercent := atrValue / currentPrice * 100

	if atrPercent > 3.0 {
		return "HAUTE VOLATILITÉ"
	} else if atrPercent > 1.5 {
		return "VOLATILITÉ MODÉRÉE"
	} else if atrPercent > 0.5 {
		return "FAIBLE VOLATILITÉ"
	} else {
		return "TRÈS FAIBLE VOLATILITÉ"
	}
}

// IsHighVolatility vérifie si la volatilité est élevée
func (atr *ATRTVStandard) IsHighVolatility(atrValue, currentPrice float64) bool {
	if math.IsNaN(atrValue) || math.IsNaN(currentPrice) || currentPrice == 0 {
		return false
	}
	atrPercent := atrValue / currentPrice * 100
	return atrPercent > 2.0
}

// IsLowVolatility vérifie si la volatilité est faible
func (atr *ATRTVStandard) IsLowVolatility(atrValue, currentPrice float64) bool {
	if math.IsNaN(atrValue) || math.IsNaN(currentPrice) || currentPrice == 0 {
		return false
	}
	atrPercent := atrValue / currentPrice * 100
	return atrPercent < 0.5
}

// GetStopLossDistance calcule la distance de stop-loss basée sur l'ATR
func (atr *ATRTVStandard) GetStopLossDistance(atrValue, multiplier float64) float64 {
	if math.IsNaN(atrValue) || multiplier <= 0 {
		return math.NaN()
	}
	return atrValue * multiplier
}

// GetTakeProfitDistance calcule la distance de take-profit basée sur l'ATR
func (atr *ATRTVStandard) GetTakeProfitDistance(atrValue, multiplier float64) float64 {
	if math.IsNaN(atrValue) || multiplier <= 0 {
		return math.NaN()
	}
	return atrValue * multiplier
}

// GetVolatilityTrend détermine la tendance de la volatilité
func (atr *ATRTVStandard) GetVolatilityTrend(atrValues []float64, lookback int) string {
	if lookback <= 1 || len(atrValues) < lookback {
		return "Insuffisant"
	}

	// Analyser les dernières 'lookback' périodes
	recentATR := atrValues[len(atrValues)-lookback:]
	
	// Vérifier si nous avons des valeurs valides
	validATR := make([]float64, 0)
	for _, v := range recentATR {
		if !math.IsNaN(v) {
			validATR = append(validATR, v)
		}
	}

	if len(validATR) < 2 {
		return "Insuffisant"
	}

	// Volatilité en expansion: ATR monte
	atrExpanding := validATR[len(validATR)-1] > validATR[0]
	
	// Volatilité en contraction: ATR descend
	atrContracting := validATR[len(validATR)-1] < validATR[0]

	if atrExpanding {
		return "Expansion"
	} else if atrContracting {
		return "Contraction"
	} else {
		return "Stable"
	}
}

// GetATRBands calcule les bandes ATR (supérieure et inférieure)
func (atr *ATRTVStandard) GetATRBands(currentPrice, atrValue, multiplier float64) (upperBand, lowerBand float64) {
	if math.IsNaN(currentPrice) || math.IsNaN(atrValue) || multiplier <= 0 {
		return math.NaN(), math.NaN()
	}

	upperBand = currentPrice + (atrValue * multiplier)
	lowerBand = currentPrice - (atrValue * multiplier)

	return upperBand, lowerBand
}

// IsPriceOutsideATRBands vérifie si le prix est en dehors des bandes ATR
func (atr *ATRTVStandard) IsPriceOutsideATRBands(currentPrice, atrValue, multiplier float64) (above, below bool) {
	upperBand, lowerBand := atr.GetATRBands(currentPrice, atrValue, multiplier)
	
	if math.IsNaN(upperBand) || math.IsNaN(lowerBand) {
		return false, false
	}

	above = currentPrice > upperBand
	below = currentPrice < lowerBand

	return above, below
}
