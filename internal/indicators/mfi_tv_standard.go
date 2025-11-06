package indicators

import (
	"math"
	"sort"
)

// MFITVStandard - Implémentation Money Flow Index TradingView Standard
// Basé sur la documentation mfi_tradingview_research.md
// Composants: Typical Price, Money Flow, Positive/Negative Money Flow, MFI
type MFITVStandard struct {
	period int
}

// NewMFITVStandard crée une nouvelle instance MFI TV Standard
func NewMFITVStandard(period int) *MFITVStandard {
	return &MFITVStandard{
		period: period,
	}
}

// Calculate calcule le MFI selon les spécifications TradingView
func (mfi *MFITVStandard) Calculate(high, low, close, volume []float64) []float64 {
	n := len(high)
	if n != len(low) || n != len(close) || n != len(volume) {
		return nil
	}

	// Calculer Typical Price: TP = (High + Low + Close) / 3
	typicalPrice := mfi.calculateTypicalPrice(high, low, close)

	// Calculer Raw Money Flow: MF = TP × Volume
	rawMoneyFlow := mfi.calculateRawMoneyFlow(typicalPrice, volume)

	// Classifier Positive/Negative Money Flow
	positiveMF, negativeMF := mfi.classifyMoneyFlow(typicalPrice, rawMoneyFlow)

	// Calculer MFI final
	mfiValues := mfi.calculateMFIFinal(positiveMF, negativeMF)

	return mfiValues
}

// calculateTypicalPrice calcule Typical Price (méthode TradingView)
func (mfi *MFITVStandard) calculateTypicalPrice(high, low, close []float64) []float64 {
	n := len(high)
	tp := make([]float64, n)

	for i := 0; i < n; i++ {
		tp[i] = (high[i] + low[i] + close[i]) / 3.0
	}

	return tp
}

// calculateRawMoneyFlow calcule Raw Money Flow (méthode TradingView)
func (mfi *MFITVStandard) calculateRawMoneyFlow(typicalPrice, volume []float64) []float64 {
	n := len(typicalPrice)
	mf := make([]float64, n)

	for i := 0; i < n; i++ {
		mf[i] = typicalPrice[i] * volume[i]
	}

	return mf
}

// classifyMoneyFlow classifie Positive/Negative Money Flow (méthode TradingView)
func (mfi *MFITVStandard) classifyMoneyFlow(typicalPrice, rawMoneyFlow []float64) (positiveMF, negativeMF []float64) {
	n := len(typicalPrice)
	positiveMF = make([]float64, n)
	negativeMF = make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			// Première barre: pas de comparaison possible
			positiveMF[i] = 0
			negativeMF[i] = 0
		} else {
			// Classification selon TradingView
			if typicalPrice[i] > typicalPrice[i-1] {
				positiveMF[i] = rawMoneyFlow[i]
				negativeMF[i] = 0
			} else if typicalPrice[i] < typicalPrice[i-1] {
				positiveMF[i] = 0
				negativeMF[i] = rawMoneyFlow[i]
			} else {
				// TP[i] == TP[i-1]
				positiveMF[i] = 0
				negativeMF[i] = 0
			}
		}
	}

	return positiveMF, negativeMF
}

// calculateMFIFinal calcule le MFI final (méthode TradingView)
func (mfi *MFITVStandard) calculateMFIFinal(positiveMF, negativeMF []float64) []float64 {
	n := len(positiveMF)
	mfiValues := make([]float64, n)

	// Initialiser avec NaN (TradingView standard)
	for i := range mfiValues {
		mfiValues[i] = math.NaN()
	}

	if mfi.period <= 0 || n == 0 {
		return mfiValues
	}

	// Calculer les sommes glissantes
	for i := mfi.period; i < n; i++ {
		// Somme sur la période [i-period+1 .. i]
		sumPositive := 0.0
		sumNegative := 0.0
		validCount := 0

		for j := i - mfi.period + 1; j <= i; j++ {
			if !math.IsNaN(positiveMF[j]) && !math.IsNaN(negativeMF[j]) {
				sumPositive += positiveMF[j]
				sumNegative += negativeMF[j]
				validCount++
			}
		}

		// Calculer MFI seulement si période complète
		if validCount == mfi.period {
			mfiValues[i] = mfi.calculateMFIValue(sumPositive, sumNegative)
		}
	}

	// NOTE: Code désactivé - uniformisation avec CCI/Stoch qui ne forcent pas NaN
	// TradingView: exclure la dernière barre (in-flight)
	// if n > 0 {
	// 	mfiValues[n-1] = math.NaN()
	// }

	return mfiValues
}

// calculateMFIValue calcule la valeur MFI selon les formules TradingView
func (mfi *MFITVStandard) calculateMFIValue(sumPositive, sumNegative float64) float64 {
	// Cas particuliers TradingView
	if sumPositive > 0 && sumNegative == 0 {
		return 100.0
	}
	if sumPositive == 0 && sumNegative > 0 {
		return 0.0
	}
	if sumPositive == 0 && sumNegative == 0 {
		return 50.0
	}

	// Formule standard: MFI = 100 - (100 / (1 + MoneyFlowRatio))
	moneyFlowRatio := sumPositive / sumNegative
	mfiValue := 100.0 - (100.0 / (1.0 + moneyFlowRatio))

	return mfiValue
}

// GetLastValue retourne la dernière valeur valide du MFI
func (mfi *MFITVStandard) GetLastValue(mfiValues []float64) float64 {
	for i := len(mfiValues) - 1; i >= 0; i-- {
		if !math.IsNaN(mfiValues[i]) {
			return mfiValues[i]
		}
	}
	return math.NaN()
}

// GetSignal retourne le signal de trading basé sur le MFI
func (mfi *MFITVStandard) GetSignal(mfiValue float64) string {
	if math.IsNaN(mfiValue) {
		return "Inconnu"
	}

	if mfiValue > 80 {
		return "SURACHAT"
	} else if mfiValue > 70 {
		return "Zone Haute"
	} else if mfiValue < 20 {
		return "SURVENTE"
	} else if mfiValue < 30 {
		return "Zone Basse"
	} else {
		return "NEUTRE"
	}
}

// IsOverbought vérifie si le MFI est en zone de surachat
func (mfi *MFITVStandard) IsOverbought(mfiValue float64) bool {
	return !math.IsNaN(mfiValue) && mfiValue > 80
}

// IsOversold vérifie si le MFI est en zone de survente
func (mfi *MFITVStandard) IsOversold(mfiValue float64) bool {
	return !math.IsNaN(mfiValue) && mfiValue < 20
}

// IsExitingOverbought vérifie la sortie de zone de surachat
func (mfi *MFITVStandard) IsExitingOverbought(mfiValues []float64, index int) bool {
	if index <= 0 || index >= len(mfiValues) {
		return false
	}

	prev := mfiValues[index-1]
	curr := mfiValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		prev > 80 && curr <= 80
}

// IsExitingOversold vérifie la sortie de zone de survente
func (mfi *MFITVStandard) IsExitingOversold(mfiValues []float64, index int) bool {
	if index <= 0 || index >= len(mfiValues) {
		return false
	}

	prev := mfiValues[index-1]
	curr := mfiValues[index]

	return !math.IsNaN(prev) && !math.IsNaN(curr) &&
		prev < 20 && curr >= 20
}

// GetDivergenceType détecte les divergences prix/MFI
func (mfi *MFITVStandard) GetDivergenceType(prices, mfiValues []float64, lookback int) string {
	if lookback <= 1 || len(prices) < lookback || len(mfiValues) < lookback {
		return "Insuffisant"
	}

	// Analyser les dernières 'lookback' périodes
	recentPrices := prices[len(prices)-lookback:]
	recentMFI := mfiValues[len(mfiValues)-lookback:]

	// Vérifier si nous avons des valeurs valides
	validMFI := make([]float64, 0)
	for _, v := range recentMFI {
		if !math.IsNaN(v) {
			validMFI = append(validMFI, v)
		}
	}

	if len(validMFI) < 2 {
		return "Insuffisant"
	}

	// Divergence baissière: prix monte, MFI descend
	priceUp := recentPrices[len(recentPrices)-1] > recentPrices[0]
	mfiDown := validMFI[len(validMFI)-1] < validMFI[0]

	if priceUp && mfiDown {
		return "Baissière"
	}

	// Divergence haussière: prix descend, MFI monte
	priceDown := recentPrices[len(recentPrices)-1] < recentPrices[0]
	mfiUp := validMFI[len(validMFI)-1] > validMFI[0]

	if priceDown && mfiUp {
		return "Haussière"
	}

	return "Aucune"
}

// CalculateFromKlines calcule MFI depuis les klines (méthode utilitaire)
func (mfi *MFITVStandard) CalculateFromKlines(klines []Kline) []float64 {
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

	// Extraire les données
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	volume := make([]float64, len(klines))

	for i, k := range klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volume[i] = k.Volume
	}

	return mfi.Calculate(high, low, close, volume)
}
