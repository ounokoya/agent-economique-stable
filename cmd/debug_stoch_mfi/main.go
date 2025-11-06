package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/indicators"
)

// Copie des structures et fonctions nécessaires du fichier main.go
type Signal struct {
	Timestamp  time.Time `json:"timestamp"`
	Strategy   string    `json:"strategy"`
	Direction  string    `json:"direction"`
	Price      float64   `json:"price"`
	Confidence float64   `json:"confidence"`
	Conditions []string  `json:"conditions"`
}

type ScalpingStrategy struct {
	config         StrategyConfig
	klines         []indicators.Kline
	cciValues      []float64
	mfiValues      []float64
	stochKValues   []float64
	stochDValues   []float64
	volumeAnalyzer *VolumeAnalyzer
	ExtremeConditions int
}

type StrategyConfig struct {
	CCIPeriod     int
	MFIPeriod     int
	StochKPeriod  int
	StochDPeriod  int
	VolumePeriodeAnalyse   int
	VolumeSeuilPourcentage float64
	VolumeMaxExtension     int
	CCISurachat   float64
	CCISurvente   float64
	MFISurachat   float64
	MFISurvente   float64
	StochSurachat float64
	StochSurvente float64
	VolumeConditionEnabled bool
}

type VolumeAnalyzer struct {
	klines       []indicators.Kline
	periodeAnalyse int
	seuilPourcentage float64
	maxExtension     int
}

func DefaultStrategyConfig() StrategyConfig {
	return StrategyConfig{
		CCIPeriod:     20,
		MFIPeriod:     14,
		StochKPeriod:  14,
		StochDPeriod:  3,
		VolumePeriodeAnalyse:   5,
		VolumeSeuilPourcentage: 25.0,
		VolumeMaxExtension:     100,
		VolumeConditionEnabled: false,
		CCISurachat:   100,
		CCISurvente:   -100,
		MFISurachat:   80,
		MFISurvente:   20,
		StochSurachat: 80,
		StochSurvente: 20,
	}
}

func NewScalpingStrategy(klines []indicators.Kline) *ScalpingStrategy {
	config := DefaultStrategyConfig()
	
	strategy := &ScalpingStrategy{
		config: config,
		klines: klines,
		volumeAnalyzer: NewVolumeAnalyzer(klines, config.VolumePeriodeAnalyse, config.VolumeSeuilPourcentage, config.VolumeMaxExtension),
	}
	
	strategy.calculateIndicators()
	return strategy
}

func NewVolumeAnalyzer(klines []indicators.Kline, periodeAnalyse int, seuilPourcentage float64, maxExtension int) *VolumeAnalyzer {
	return &VolumeAnalyzer{
		klines:       klines,
		periodeAnalyse: periodeAnalyse,
		seuilPourcentage: seuilPourcentage,
		maxExtension:     maxExtension,
	}
}

func (s *ScalpingStrategy) calculateIndicators() {
	high := make([]float64, len(s.klines))
	low := make([]float64, len(s.klines))
	close := make([]float64, len(s.klines))
	volume := make([]float64, len(s.klines))
	
	for i, k := range s.klines {
		high[i] = k.High
		low[i] = k.Low
		close[i] = k.Close
		volume[i] = k.Volume
	}
	
	cciTV := indicators.NewCCITVStandard(s.config.CCIPeriod)
	s.cciValues = cciTV.Calculate(high, low, close)
	
	mfiTV := indicators.NewMFITVStandard(s.config.MFIPeriod)
	s.mfiValues = mfiTV.Calculate(high, low, close, volume)
	
	stochTV := indicators.NewStochTVStandard(s.config.StochKPeriod, s.config.StochDPeriod, 3)
	s.stochKValues, s.stochDValues = stochTV.Calculate(high, low, close)
}

func convertToIndicatorsKlines(klines []binance.Kline) []indicators.Kline {
	indicatorsKlines := make([]indicators.Kline, len(klines))
	for i, k := range klines {
		indicatorsKlines[i] = indicators.Kline{
			Timestamp: k.OpenTime.Unix(),
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}
	return indicatorsKlines
}

// Fonction debug principale
func debugStochMFICoincidence(scalpingStrategy *ScalpingStrategy) {
	fmt.Println("🔍 DEBUG: Stoch + MFI coincidence sur 100 dernières bougies")
	fmt.Println("===========================================================")
	
	var coincidences []struct {
		index int
		time string
		stochK float64
		stochD float64
		mfi float64
		zone string
	}
	
	// Analyser les 100 dernières bougies
	startIndex := len(scalpingStrategy.klines) - 100
	for i := startIndex; i < len(scalpingStrategy.klines); i++ {
		if i >= len(scalpingStrategy.stochKValues) || i >= len(scalpingStrategy.stochDValues) || i >= len(scalpingStrategy.mfiValues) {
			continue
		}
		
		stochK := scalpingStrategy.stochKValues[i]
		stochD := scalpingStrategy.stochDValues[i]
		mfi := scalpingStrategy.mfiValues[i]
		
		// Vérifier NaN
		if math.IsNaN(stochK) || math.IsNaN(stochD) || math.IsNaN(mfi) {
			continue
		}
		
		// Zones extrêmes
		stochOverbought := stochK > 80 && stochD > 80
		stochOversold := stochK < 20 && stochD < 20
		mfiOverbought := mfi > 80
		mfiOversold := mfi < 20
		
		var zone string
		isCoincidence := false
		
		if stochOverbought && mfiOverbought {
			zone = "SURACHAT"
			isCoincidence = true
		} else if stochOversold && mfiOversold {
			zone = "SURVENTE"
			isCoincidence = true
		}
		
		if isCoincidence {
			coincidences = append(coincidences, struct {
				index int
				time string
				stochK float64
				stochD float64
				mfi float64
				zone string
			}{
				index:  i,
				time:   time.Unix(scalpingStrategy.klines[i].Timestamp, 0).Format("15:04"),
				stochK: stochK,
				stochD: stochD,
				mfi:   mfi,
				zone:  zone,
			})
		}
	}
	
	// Afficher les résultats
	fmt.Printf("📊 Résultats: %d coincidences Stoch+MFI trouvées\n\n", len(coincidences))
	
	if len(coincidences) > 0 {
		fmt.Println("🎯 DÉTAIL DES COINCIDENCES:")
		fmt.Println("┌──────┬────────┬─────────────┬─────────────┬──────────┬──────────┐")
		fmt.Println("│ Index│ Heure  │   Stoch K   │   Stoch D   │   MFI    │  Zone    │")
		fmt.Println("├──────┼────────┼─────────────┼─────────────┼──────────┼──────────┤")
		
		for _, coin := range coincidences {
			fmt.Printf("│ %4d │ %6s │ %11.1f │ %11.1f │ %8.1f │ %8s │\n",
				coin.index, coin.time, coin.stochK, coin.stochD, coin.mfi, coin.zone)
		}
		
		fmt.Println("└──────┴────────┴─────────────┴─────────────┴──────────┴──────────┘")
		
		// Afficher début et fin
		fmt.Printf("\n📍 PÉRIODE: %s → %s\n", 
			coincidences[0].time, coincidences[len(coincidences)-1].time)
		fmt.Printf("📍 INDICES: %d → %d\n", 
			coincidences[0].index, coincidences[len(coincidences)-1].index)
	} else {
		fmt.Println("❌ Aucune coincidence Stoch+MFI trouvée")
	}
	
	fmt.Println("===========================================================")
}

func main() {
	fmt.Println("🔍 DEBUG SEUL: Stoch + MFI coincidence")
	fmt.Println("=======================================")

	// Configuration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connexion Binance Futures
	fmt.Println("📡 Connexion Binance Futures...")
	client := binance.NewFuturesClient()

	// Récupération des 300 dernières bougies
	fmt.Println("📊 Récupération des 300 dernières klines...")
	futuresKlines, err := client.GetKlines(ctx, "SOLUSDT", "5m", 300)
	if err != nil {
		log.Fatalf("❌ Erreur récupération klines: %v", err)
	}

	klines := client.ConvertToStandardKline(futuresKlines)
	fmt.Printf("✅ %d klines récupérées de %s à %s\n", 
		len(klines), 
		klines[0].OpenTime.Format("2006-01-02 15:04"), 
		klines[len(klines)-1].OpenTime.Format("15:04"))

	// Conversion pour indicateurs
	indicatorsKlines := convertToIndicatorsKlines(klines)

	// Initialisation stratégie SEULEMENT pour debug
	scalpingStrategy := NewScalpingStrategy(indicatorsKlines)

	// LANCEMENT SEULEMENT de la fonction debug
	debugStochMFICoincidence(scalpingStrategy)
}
