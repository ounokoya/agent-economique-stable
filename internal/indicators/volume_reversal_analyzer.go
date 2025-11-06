package indicators

import (
	"fmt"
	"math"
)

// VolumeReversalAnalyzer - Analyse les volumes sur inversions de tendance
type VolumeReversalAnalyzer struct {
	defaultLookback       int
	defaultVolumeThreshold float64
}

// NewVolumeReversalAnalyzer crée une nouvelle instance
func NewVolumeReversalAnalyzer() *VolumeReversalAnalyzer {
	return &VolumeReversalAnalyzer{
		defaultLookback:        5,
		defaultVolumeThreshold: 1.5, // 150% par défaut
	}
}

// AnalyzeVolumeReversalResult - Résultat de l'analyse
type AnalyzeVolumeReversalResult struct {
	IsReversalWithHighVolume bool     // True si condition vérifiée
	CurrentCandleColor       string   // "verte" ou "rouge"
	CurrentVolume            float64  // Volume de la bougie actuelle
	AverageInverseVolume     float64  // Moyenne volume bougies inverses
	VolumeRatio              float64  // Ratio volume actuel / moyenne inverse
	LookbackUsed             int      // N utilisé pour l'analyse
	Message                  string   // Message détaillé
}

// AnalyzeVolumeReversal analyse si la bougie actuelle a un volume significatif
// sur une inversion par rapport aux N dernières bougies de couleur inverse
func (vra *VolumeReversalAnalyzer) AnalyzeVolumeReversal(
	klines []Kline,
	currentIndex int,
	lookback int,
	volumeThreshold float64,
) AnalyzeVolumeReversalResult {
	
	// Valeurs par défaut
	if lookback <= 0 {
		lookback = vra.defaultLookback
	}
	if volumeThreshold <= 0 {
		volumeThreshold = vra.defaultVolumeThreshold
	}
	
	// Validation des indices
	if currentIndex < 0 || currentIndex >= len(klines) {
		return AnalyzeVolumeReversalResult{
			IsReversalWithHighVolume: false,
			Message:                  "Indice invalide",
		}
	}
	
	currentKline := klines[currentIndex]
	
	// Déterminer la couleur de la bougie actuelle
	var currentColor string
	var isGreen bool
	if currentKline.Close > currentKline.Open {
		currentColor = "verte"
		isGreen = true
	} else if currentKline.Close < currentKline.Open {
		currentColor = "rouge"
		isGreen = false
	} else {
		// Bougie doji (close = open)
		return AnalyzeVolumeReversalResult{
			IsReversalWithHighVolume: false,
			CurrentCandleColor:       "doji",
			CurrentVolume:            currentKline.Volume,
			Message:                  "Bougie doji (close = open)",
		}
	}
	
	// Chercher les bougies inverses avec expansion progressive
	searchLookback := lookback
	maxSearchLookback := lookback * 8 // Limite raisonnable
	
	for searchLookback <= maxSearchLookback {
		result := vra.analyzeWithLookback(klines, currentIndex, searchLookback, volumeThreshold, isGreen)
		if result.AverageInverseVolume > 0 {
			// On a trouvé des bougies inverses
			result.LookbackUsed = searchLookback
			return result
		}
		// Multiplier par 2 pour la prochaine recherche
		searchLookback *= 2
	}
	
	// Aucune bougie inverse trouvée même après expansion
	return AnalyzeVolumeReversalResult{
		IsReversalWithHighVolume: false,
		CurrentCandleColor:       currentColor,
		CurrentVolume:            currentKline.Volume,
		LookbackUsed:             maxSearchLookback,
		Message:                  fmt.Sprintf("Aucune bougie %s trouvée sur les %d dernières bougies", 
			vra.getInverseColor(currentColor), maxSearchLookback),
	}
}

// analyzeWithLookback analyse avec un lookback spécifique
func (vra *VolumeReversalAnalyzer) analyzeWithLookback(
	klines []Kline,
	currentIndex int,
	lookback int,
	volumeThreshold float64,
	isGreen bool,
) AnalyzeVolumeReversalResult {
	
	currentKline := klines[currentIndex]
	currentColor := "verte"
	if !isGreen {
		currentColor = "rouge"
	}
	
	// Collecter les volumes des bougies inverses
	var inverseVolumes []float64
	startIndex := currentIndex - lookback
	if startIndex < 0 {
		startIndex = 0
	}
	
	for i := startIndex; i < currentIndex; i++ {
		kline := klines[i]
		
		// Vérifier si c'est une bougie inverse
		isInverse := false
		if isGreen && kline.Close < kline.Open {
			isInverse = true // Bougie rouge inverse
		} else if !isGreen && kline.Close > kline.Open {
			isInverse = true // Bougie verte inverse
		}
		
		if isInverse && !math.IsNaN(kline.Volume) && kline.Volume > 0 {
			inverseVolumes = append(inverseVolumes, kline.Volume)
		}
	}
	
	// Si pas de bougies inverses
	if len(inverseVolumes) == 0 {
		return AnalyzeVolumeReversalResult{
			IsReversalWithHighVolume: false,
			CurrentCandleColor:       currentColor,
			CurrentVolume:            currentKline.Volume,
			AverageInverseVolume:     0,
		}
	}
	
	// Calculer la moyenne des volumes inverses
	sumInverseVolume := 0.0
	for _, vol := range inverseVolumes {
		sumInverseVolume += vol
	}
	averageInverseVolume := sumInverseVolume / float64(len(inverseVolumes))
	
	// Calculer le ratio
	volumeRatio := currentKline.Volume / averageInverseVolume
	
	// Vérifier la condition
	isReversalWithHighVolume := volumeRatio >= volumeThreshold
	
	// Créer le message
	message := fmt.Sprintf("Bougie %s - Volume: %.2f, Moyenne %s (n=%d): %.2f, Ratio: %.2fx",
		currentColor,
		currentKline.Volume,
		vra.getInverseColor(currentColor),
		len(inverseVolumes),
		averageInverseVolume,
		volumeRatio,
	)
	
	if isReversalWithHighVolume {
		message += fmt.Sprintf(" ✅ Volume significatif (> %.0fx)", volumeThreshold)
	} else {
		message += fmt.Sprintf(" ❌ Volume faible (< %.0fx)", volumeThreshold)
	}
	
	return AnalyzeVolumeReversalResult{
		IsReversalWithHighVolume: isReversalWithHighVolume,
		CurrentCandleColor:       currentColor,
		CurrentVolume:            currentKline.Volume,
		AverageInverseVolume:     averageInverseVolume,
		VolumeRatio:              volumeRatio,
		Message:                  message,
	}
}

// getInverseColor retourne la couleur inverse
func (vra *VolumeReversalAnalyzer) getInverseColor(color string) string {
	if color == "verte" {
		return "rouge"
	}
	return "verte"
}

// IsReversalWithHighVolume méthode simplifiée qui retourne juste le booléen
func (vra *VolumeReversalAnalyzer) IsReversalWithHighVolume(
	klines []Kline,
	currentIndex int,
	lookback int,
	volumeThreshold float64,
) bool {
	
	result := vra.AnalyzeVolumeReversal(klines, currentIndex, lookback, volumeThreshold)
	return result.IsReversalWithHighVolume
}

// GetReversalInfo retourne des informations détaillées sur l'analyse
func (vra *VolumeReversalAnalyzer) GetReversalInfo(
	klines []Kline,
	currentIndex int,
	lookback int,
	volumeThreshold float64,
) string {
	
	result := vra.AnalyzeVolumeReversal(klines, currentIndex, lookback, volumeThreshold)
	return result.Message
}
