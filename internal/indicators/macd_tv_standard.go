package indicators

import (
	"math"
)

// MACDTVStandard - Implémentation MACD TradingView Standard
// Composants: MACD Line (12-26), Signal Line (9), Histogram
type MACDTVStandard struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int
}

// NewMACDTVStandard crée une nouvelle instance MACD TV Standard
func NewMACDTVStandard(fast, slow, signal int) *MACDTVStandard {
	return &MACDTVStandard{
		fastPeriod:   fast,
		slowPeriod:   slow,
		signalPeriod: signal,
	}
}

// Calculate calcule le MACD selon la formule TradingView standard
func (macd *MACDTVStandard) Calculate(prices []float64) (macdLine, signalLine, histogram []float64) {
	n := len(prices)
	if n == 0 {
		return nil, nil, nil
	}

	// Calculer EMAs
	fastEMA := macd.calculateEMA(prices, macd.fastPeriod)
	slowEMA := macd.calculateEMA(prices, macd.slowPeriod)

	// Calculer MACD Line
	macdLine = make([]float64, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(fastEMA[i]) && !math.IsNaN(slowEMA[i]) {
			macdLine[i] = fastEMA[i] - slowEMA[i]
		} else {
			macdLine[i] = math.NaN()
		}
	}

	// Calculer Signal Line (EMA du MACD Line)
	signalLine = macd.calculateEMA(macdLine, macd.signalPeriod)

	// Calculer Histogram
	histogram = make([]float64, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(macdLine[i]) && !math.IsNaN(signalLine[i]) {
			histogram[i] = macdLine[i] - signalLine[i]
		} else {
			histogram[i] = math.NaN()
		}
	}

	return macdLine, signalLine, histogram
}

// calculateEMA calcule Exponential Moving Average (méthode TradingView)
func (macd *MACDTVStandard) calculateEMA(values []float64, period int) []float64 {
	ema := NewEMATVStandard(period)
	return ema.Calculate(values)
}

// GetLastValues retourne les dernières valeurs valides
func (macd *MACDTVStandard) GetLastValues(macdLine, signalLine, histogram []float64) (float64, float64, float64) {
	getLastValid := func(values []float64) float64 {
		for i := len(values) - 1; i >= 0; i-- {
			if !math.IsNaN(values[i]) {
				return values[i]
			}
		}
		return math.NaN()
	}

	return getLastValid(macdLine), getLastValid(signalLine), getLastValid(histogram)
}

// GetMACDSignal retourne le signal MACD actuel
func (macd *MACDTVStandard) GetMACDSignal(macdLine, signalLine float64) string {
	if math.IsNaN(macdLine) || math.IsNaN(signalLine) {
		return "Inconnu"
	}
	
	if macdLine > signalLine {
		return "Haussier"
	} else if macdLine < signalLine {
		return "Baissier"
	} else {
		return "Neutre"
	}
}

// GetHistogramSignal retourne le signal de l'histogramme
func (macd *MACDTVStandard) GetHistogramSignal(histogram float64) string {
	if math.IsNaN(histogram) {
		return "Inconnu"
	}
	
	if histogram > 0 {
		return "Positif"
	} else if histogram < 0 {
		return "Négatif"
	} else {
		return "Neutre"
	}
}

// IsBullishCrossover vérifie croisement haussier MACD > Signal
func (macd *MACDTVStandard) IsBullishCrossover(macdLine, signalLine []float64, index int) bool {
	if index <= 0 || index >= len(macdLine) || index >= len(signalLine) {
		return false
	}
	
	prevMACD := macdLine[index-1]
	currMACD := macdLine[index]
	prevSignal := signalLine[index-1]
	currSignal := signalLine[index]
	
	// Croisement haussier: MACD passe sous Signal à dessus de Signal
	return prevMACD <= prevSignal && currMACD > currSignal
}

// IsBearishCrossover vérifie croisement baissier MACD < Signal
func (macd *MACDTVStandard) IsBearishCrossover(macdLine, signalLine []float64, index int) bool {
	if index <= 0 || index >= len(macdLine) || index >= len(signalLine) {
		return false
	}
	
	prevMACD := macdLine[index-1]
	currMACD := macdLine[index]
	prevSignal := signalLine[index-1]
	currSignal := signalLine[index]
	
	// Croisement baissier: MACD passe au-dessus de Signal à sous Signal
	return prevMACD >= prevSignal && currMACD < currSignal
}

// IsZeroLineCrossoverUp vérifie croisement ligne zéro vers le haut
func (macd *MACDTVStandard) IsZeroLineCrossoverUp(macdLine []float64, index int) bool {
	if index <= 0 || index >= len(macdLine) {
		return false
	}
	
	return macdLine[index-1] <= 0 && macdLine[index] > 0
}

// IsZeroLineCrossoverDown vérifie croisement ligne zéro vers le bas
func (macd *MACDTVStandard) IsZeroLineCrossoverDown(macdLine []float64, index int) bool {
	if index <= 0 || index >= len(macdLine) {
		return false
	}
	
	return macdLine[index-1] >= 0 && macdLine[index] < 0
}

// GetMomentumStrength retourne la force du momentum
func (macd *MACDTVStandard) GetMomentumStrength(histogram float64) string {
	if math.IsNaN(histogram) {
		return "Inconnu"
	}
	
	abs := math.Abs(histogram)
	if abs > 1.0 {
		return "Très Fort"
	} else if abs > 0.5 {
		return "Fort"
	} else if abs > 0.1 {
		return "Modéré"
	} else if abs > 0.01 {
		return "Faible"
	} else {
		return "Très Faible"
	}
}

// GetDivergenceType détecte les divergences potentielles
func (macd *MACDTVStandard) GetDivergenceType(prices, macdLine []float64, lookback int) string {
	if lookback <= 0 || len(prices) < lookback || len(macdLine) < lookback {
		return "Insuffisant"
	}
	
	// Divergence baissière: prix plus haut mais MACD plus bas
	if prices[len(prices)-1] > prices[len(prices)-lookback] && 
	   macdLine[len(macdLine)-1] < macdLine[len(macdLine)-lookback] {
		return "Baissière"
	}
	
	// Divergence haussière: prix plus bas mais MACD plus haut
	if prices[len(prices)-1] < prices[len(prices)-lookback] && 
	   macdLine[len(macdLine)-1] > macdLine[len(macdLine)-lookback] {
		return "Haussière"
	}
	
	return "Aucune"
}
