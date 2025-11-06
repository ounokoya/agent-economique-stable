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

// ============================================================
// 🎯 PARAMÈTRES STRATÉGIE - MODIFIABLES
// ============================================================

// Marché et Données
const (
	SYMBOL        = "SOLUSDT" // Paire à trader
	TIMEFRAME     = "30m"     // Timeframe des bougies
	CANDLES_COUNT = 300       // Nombre de bougies à analyser
)

// Seuils extrêmes
const (
	CCI_SURACHAT   = 100.0  // Surachat CCI
	CCI_SURVENTE   = -100.0 // Survente CCI
	MFI_SURACHAT   = 60.0   // Surachat MFI
	MFI_SURVENTE   = 40.0   // Survente MFI
	STOCH_SURACHAT = 80.0   // Surachat Stochastique
	STOCH_SURVENTE = 20.0   // Survente Stochastique
)

// Fenêtre de validation
const VALIDATION_WINDOW = 3

// Synchronisation tendance (croissant/décroissant) au moment du croisement
const (
	SYNC_MFI_WITH_K = true // MFI doit suivre le même sens que K (croissant/décroissant)
	SYNC_CCI_WITH_K = true // CCI doit suivre le même sens que K (croissant/décroissant)
)

// Volume
const (
	VOLUME_ENABLED   = true // Activer vérification volume
	VOLUME_THRESHOLD = 0.25 // 25% du volume moyen bougies inverses
	VOLUME_PERIOD    = 5    // Période de base pour volume
	VOLUME_MAX_EXT   = 100  // Extension max pour recherche volume
)

// ============================================================

// Configuration
type Config struct {
	Symbol           string
	Timeframe        string
	CandlesCount     int
	ValidationWindow int // Fenêtre de validation N (défaut: 3)

	// Seuils extrêmes
	CCISurachat   float64
	CCISurvente   float64
	MFISurachat   float64
	MFISurvente   float64
	StochSurachat float64
	StochSurvente float64

	// Synchronisation tendance
	SyncMFIWithK bool // MFI doit suivre même sens que K
	SyncCCIWithK bool // CCI doit suivre même sens que K

	// Volume
	VolumeEnabled      bool
	VolumeThreshold    float64 // 0.25 = 25%
	VolumePeriod       int
	VolumeMaxExtension int
}

func DefaultConfig() Config {
	return Config{
		Symbol:           SYMBOL,
		Timeframe:        TIMEFRAME,
		CandlesCount:     CANDLES_COUNT,
		ValidationWindow: VALIDATION_WINDOW,

		CCISurachat:   CCI_SURACHAT,
		CCISurvente:   CCI_SURVENTE,
		MFISurachat:   MFI_SURACHAT,
		MFISurvente:   MFI_SURVENTE,
		StochSurachat: STOCH_SURACHAT,
		StochSurvente: STOCH_SURVENTE,

		SyncMFIWithK: SYNC_MFI_WITH_K,
		SyncCCIWithK: SYNC_CCI_WITH_K,

		VolumeEnabled:      VOLUME_ENABLED,
		VolumeThreshold:    VOLUME_THRESHOLD,
		VolumePeriod:       VOLUME_PERIOD,
		VolumeMaxExtension: VOLUME_MAX_EXT,
	}
}

// Kline simple pour scalping
type Kline struct {
	Timestamp        int64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64 // Volume en SOL (base asset)
	QuoteAssetVolume float64 // Volume en USDT (quote asset)
}

// Signal de trading
type Signal struct {
	Index            int
	Time             time.Time // Heure du signal (validation)
	TimeN1           time.Time // Heure de la barre N-1 (dernière fermée)
	TimeN2           time.Time // Heure de la barre N-2 (avant-dernière fermée)
	Type             string
	Price            float64
	Volume           float64 // Volume SOL de la barre de validation
	QuoteAssetVolume float64 // Volume USDT de la barre de validation
	Confidence       int
	CCI              float64 // CCI N-1 (dernière fermée)
	MFI              float64 // MFI N-1 (dernière fermée)
	StochK           float64 // Stoch K N-1 (dernière fermée)
	StochD           float64 // Stoch D N-1 (dernière fermée)
	CCIPrev          float64 // CCI N-2 (avant-dernière fermée)
	MFIPrev          float64 // MFI N-2 (avant-dernière fermée)
	StochKPrev       float64 // Stoch K N-2 (avant-dernière fermée)
	StochDPrev       float64 // Stoch D N-2 (avant-dernière fermée)
}

// ScalpingStrategy
type ScalpingStrategy struct {
	config       Config
	klines       []Kline
	cciValues    []float64
	mfiValues    []float64
	stochKValues []float64
	stochDValues []float64
}

func NewScalpingStrategy(klines []Kline, config Config) *ScalpingStrategy {
	s := &ScalpingStrategy{
		config: config,
		klines: klines,
	}
	s.calculateIndicators()
	return s
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

	// CCI TV Standard (20)
	cciTV := indicators.NewCCITVStandard(20)
	s.cciValues = cciTV.Calculate(high, low, close)

	// MFI TV Standard (14)
	mfiTV := indicators.NewMFITVStandard(14)
	s.mfiValues = mfiTV.Calculate(high, low, close, volume)

	// Stochastique TV Standard (14,3,3)
	stochTV := indicators.NewStochTVStandard(14, 3, 3)
	s.stochKValues, s.stochDValues = stochTV.Calculate(high, low, close)

	fmt.Printf("✅ Indicateurs calculés: CCI=%d, MFI=%d, StochK=%d, StochD=%d\n",
		len(s.cciValues), len(s.mfiValues), len(s.stochKValues), len(s.stochDValues))
}

func (s *ScalpingStrategy) DetectSignals() []Signal {
	var signals []Signal

	fmt.Println("\n🔍 DÉTECTION SIGNAUX SCALPING")
	fmt.Println("=" + string(make([]byte, 50)))

	// Calculer l'index de départ dynamiquement
	// Périodes des indicateurs : CCI=20, MFI=14, Stoch=(14,3,3)
	cciPeriod := 20
	mfiPeriod := 14
	stochPeriod := 17 // 14 + 3

	// Trouver la période maximale
	maxPeriod := cciPeriod
	if mfiPeriod > maxPeriod {
		maxPeriod = mfiPeriod
	}
	if stochPeriod > maxPeriod {
		maxPeriod = stochPeriod
	}

	// + 2 pour avoir i-2 et i-1 disponibles (croisement)
	// + ValidationWindow pour la fenêtre de validation
	startIndex := maxPeriod + 2 + s.config.ValidationWindow

	fmt.Printf("   Période max indicateurs: %d, Index départ: %d\n\n", maxPeriod, startIndex)

	// 🔍 DEBUG: Afficher les indicateurs autour de 03/11 20h et 04/11 00h
	fmt.Println("\n🔍 DEBUG: VALEURS INDICATEURS AUTOUR 03/11-04/11")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i := 290; i < 300 && i < len(s.klines); i++ {
		t := time.Unix(s.klines[i].Timestamp, 0)
		fmt.Printf("[%d] %s | CCI: %7.1f | MFI: %7.1f | K: %7.1f | D: %7.1f\n",
			i, t.Format("02/01 15:04"),
			s.cciValues[i], s.mfiValues[i],
			s.stochKValues[i], s.stochDValues[i])
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Analyser chaque bougie (besoin de i-2, i-1, et fenêtre validation)
	for i := startIndex; i < len(s.klines)-1; i++ {
		// Étape 1 : Triple extrême sur dernière bougie FERMÉE (i-1)
		if !s.isTripleExtreme(i - 1) {
			continue
		}

		// Étape 2 : Croisement Stochastique (comparer i-2 vs i-1)
		crossingType := s.detectStochCrossing(i - 1)
		if crossingType == "" {
			continue
		}

		// Étape 3 : Fenêtre de validation à partir de i-1
		signal := s.validateInWindow(i-1, crossingType)
		if signal != nil {
			signals = append(signals, *signal)
			fmt.Printf("\n🎯 SIGNAL #%d VALIDÉ: %s à %s - Prix: %.2f\n\n",
				len(signals), signal.Type, signal.Time.Format("15:04"), signal.Price)
		}
	}

	return signals
}

func (s *ScalpingStrategy) isTripleExtreme(index int) bool {
	if index >= len(s.cciValues) || index >= len(s.mfiValues) ||
		index >= len(s.stochKValues) || index >= len(s.stochDValues) {
		return false
	}

	cci := s.cciValues[index]
	mfi := s.mfiValues[index]
	k := s.stochKValues[index]
	d := s.stochDValues[index]

	if math.IsNaN(cci) || math.IsNaN(mfi) || math.IsNaN(k) || math.IsNaN(d) {
		return false
	}

	// SURACHAT
	isOverbought := cci > s.config.CCISurachat && mfi > s.config.MFISurachat && (k >= s.config.StochSurachat || d >= s.config.StochSurachat)

	// SURVENTE
	isOversold := cci < s.config.CCISurvente && mfi < s.config.MFISurvente && (k <= s.config.StochSurvente || d <= s.config.StochSurvente)

	if isOverbought {
		fmt.Printf("\n🔴 TRIPLE EXTRÊME SURACHAT [%d] %s\n", index, time.Unix(s.klines[index].Timestamp, 0).Format("15:04"))
		fmt.Printf("   CCI: %.1f (>%.0f) ✅\n", cci, s.config.CCISurachat)
		fmt.Printf("   MFI: %.1f (>%.0f) ✅\n", mfi, s.config.MFISurachat)
		fmt.Printf("   Stoch K: %.1f, D: %.1f (≥%.0f) ✅\n", k, d, s.config.StochSurachat)
	} else if isOversold {
		fmt.Printf("\n🟢 TRIPLE EXTRÊME SURVENTE [%d] %s\n", index, time.Unix(s.klines[index].Timestamp, 0).Format("15:04"))
		fmt.Printf("   CCI: %.1f (<%.0f) ✅\n", cci, s.config.CCISurvente)
		fmt.Printf("   MFI: %.1f (<%.0f) ✅\n", mfi, s.config.MFISurvente)
		fmt.Printf("   Stoch K: %.1f, D: %.1f (≤%.0f) ✅\n", k, d, s.config.StochSurvente)
	}

	return isOverbought || isOversold
}

func (s *ScalpingStrategy) detectStochCrossing(index int) string {
	// index = N-1 (dernière fermée)
	// On compare N-2 vs N-1
	if index < 1 || index >= len(s.stochKValues) || index >= len(s.stochDValues) {
		return ""
	}

	// N-1 (dernière fermée)
	kN1 := s.stochKValues[index]
	dN1 := s.stochDValues[index]

	// N-2 (avant-dernière fermée)
	kN2 := s.stochKValues[index-1]
	dN2 := s.stochDValues[index-1]

	if math.IsNaN(kN1) || math.IsNaN(dN1) || math.IsNaN(kN2) || math.IsNaN(dN2) {
		return ""
	}

	cciN1 := s.cciValues[index]
	mfiN1 := s.mfiValues[index]
	cciN2 := s.cciValues[index-1]
	mfiN2 := s.mfiValues[index-1]

	if math.IsNaN(cciN1) || math.IsNaN(mfiN1) || math.IsNaN(cciN2) || math.IsNaN(mfiN2) {
		return ""
	}

	isOverbought := cciN1 > s.config.CCISurachat && mfiN1 > s.config.MFISurachat
	isOversold := cciN1 < s.config.CCISurvente && mfiN1 < s.config.MFISurvente

	// SURACHAT → croisement BAISSIER → SHORT
	// N-2: K ≥ D,  N-1: K < D (K décroissant)
	if isOverbought && kN2 >= dN2 && kN1 < dN1 {
		// Vérifier synchronisation tendance si activée
		syncOK := true

		if s.config.SyncMFIWithK {
			mfiDecroissant := mfiN1 < mfiN2
			if !mfiDecroissant {
				fmt.Printf("   ❌ MFI non synchronisé: MFI croissant (%.1f→%.1f) alors que K décroissant\n", mfiN2, mfiN1)
				syncOK = false
			} else {
				fmt.Printf("   ✅ MFI synchronisé: décroissant (%.1f→%.1f) ✓\n", mfiN2, mfiN1)
			}
		}

		if s.config.SyncCCIWithK {
			cciDecroissant := cciN1 < cciN2
			if !cciDecroissant {
				fmt.Printf("   ❌ CCI non synchronisé: CCI croissant (%.1f→%.1f) alors que K décroissant\n", cciN2, cciN1)
				syncOK = false
			} else {
				fmt.Printf("   ✅ CCI synchronisé: décroissant (%.1f→%.1f) ✓\n", cciN2, cciN1)
			}
		}

		if !syncOK {
			return ""
		}

		fmt.Printf("   ✅ Croisement BAISSIER [N-2→N-1]: K %.2f<D %.2f (était K %.2f≥D %.2f) → SHORT\n",
			kN1, dN1, kN2, dN2)
		return "SHORT"
	}

	// SURVENTE → croisement HAUSSIER → LONG
	// N-2: K ≤ D,  N-1: K > D (K croissant)
	if isOversold && kN2 <= dN2 && kN1 > dN1 {
		// Vérifier synchronisation tendance si activée
		syncOK := true

		if s.config.SyncMFIWithK {
			mfiCroissant := mfiN1 > mfiN2
			if !mfiCroissant {
				fmt.Printf("   ❌ MFI non synchronisé: MFI décroissant (%.1f→%.1f) alors que K croissant\n", mfiN2, mfiN1)
				syncOK = false
			} else {
				fmt.Printf("   ✅ MFI synchronisé: croissant (%.1f→%.1f) ✓\n", mfiN2, mfiN1)
			}
		}

		if s.config.SyncCCIWithK {
			cciCroissant := cciN1 > cciN2
			if !cciCroissant {
				fmt.Printf("   ❌ CCI non synchronisé: CCI décroissant (%.1f→%.1f) alors que K croissant\n", cciN2, cciN1)
				syncOK = false
			} else {
				fmt.Printf("   ✅ CCI synchronisé: croissant (%.1f→%.1f) ✓\n", cciN2, cciN1)
			}
		}

		if !syncOK {
			return ""
		}

		fmt.Printf("   ✅ Croisement HAUSSIER [N-2→N-1]: K %.2f>D %.2f (était K %.2f≤D %.2f) → LONG\n",
			kN1, dN1, kN2, dN2)
		return "LONG"
	}

	return ""
}

func (s *ScalpingStrategy) validateInWindow(crossingIndex int, signalType string) *Signal {
	fmt.Printf("   🔍 Fenêtre validation [%d → %d] (N=%d)\n",
		crossingIndex, crossingIndex+s.config.ValidationWindow-1, s.config.ValidationWindow)

	for i := crossingIndex; i < crossingIndex+s.config.ValidationWindow && i < len(s.klines); i++ {
		fmt.Printf("      Bougie %d/%d [%s]: ", i-crossingIndex+1, s.config.ValidationWindow,
			time.Unix(s.klines[i].Timestamp, 0).Format("15:04"))

		// Vérifier bougie inverse
		bougiValide := false
		if signalType == "SHORT" {
			bougiValide = s.klines[i].Close < s.klines[i].Open // rouge
			if bougiValide {
				fmt.Printf("ROUGE ✅ ")
			} else {
				fmt.Printf("VERTE ❌ ")
			}
		} else {
			bougiValide = s.klines[i].Close > s.klines[i].Open // verte
			if bougiValide {
				fmt.Printf("VERTE ✅ ")
			} else {
				fmt.Printf("ROUGE ❌ ")
			}
		}

		if !bougiValide {
			fmt.Println()
			continue
		}

		// Vérifier volume
		if s.config.VolumeEnabled {
			volumeOK := s.checkVolume(i, signalType)
			if !volumeOK {
				fmt.Printf("Volume ❌\n")
				continue
			}
			fmt.Printf("Volume ✅\n")
		} else {
			fmt.Printf("Volume SKIP\n")
		}

		// VALIDATION COMPLÈTE
		// Passer crossingIndex (pas i) pour afficher les indicateurs du croisement
		fmt.Printf("      ✅✅✅ VALIDATION BOUGIE %d ✅✅✅\n", i-crossingIndex+1)
		return s.createSignal(crossingIndex, i, signalType)
	}

	fmt.Printf("   ❌ Fenêtre expirée - Signal perdu\n")
	return nil
}

func (s *ScalpingStrategy) checkVolume(index int, signalType string) bool {
	if index < s.config.VolumePeriod {
		return false
	}

	volumeActuel := s.klines[index].Volume

	// Chercher bougies inverses
	volumeSum := 0.0
	count := 0
	extension := 1

	for extension <= 20 { // max 20 extensions
		periodes := s.config.VolumePeriod * extension
		startIdx := max(0, index-periodes)

		volumeSum = 0.0
		count = 0

		for i := startIdx; i < index; i++ {
			isInverse := false
			if signalType == "SHORT" {
				isInverse = s.klines[i].Close > s.klines[i].Open // verte (inverse de rouge)
			} else {
				isInverse = s.klines[i].Close < s.klines[i].Open // rouge (inverse de verte)
			}

			if isInverse {
				volumeSum += s.klines[i].Volume
				count++
			}
		}

		if count > 0 {
			volumeMoyen := volumeSum / float64(count)
			seuil := volumeMoyen * s.config.VolumeThreshold
			return volumeActuel > seuil
		}

		extension *= 2
		if periodes > s.config.VolumeMaxExtension {
			break
		}
	}

	return false
}

func (s *ScalpingStrategy) createSignal(crossingIndex int, validationIndex int, signalType string) *Signal {
	// crossingIndex = index du croisement (N-1 dans la logique)
	// validationIndex = index de la bougie de validation
	// On affiche les indicateurs du croisement, pas de la validation !

	// N-1 (dernière fermée) = crossingIndex
	// N-2 (avant-dernière fermée) = crossingIndex - 1

	var cciN1, mfiN1, stochKN1, stochDN1 float64
	var cciN2, mfiN2, stochKN2, stochDN2 float64
	var timeN1, timeN2 time.Time

	// Capturer N-1 (crossingIndex)
	if crossingIndex < len(s.cciValues) && crossingIndex < len(s.mfiValues) &&
		crossingIndex < len(s.stochKValues) && crossingIndex < len(s.stochDValues) &&
		crossingIndex < len(s.klines) {
		cciN1 = s.cciValues[crossingIndex]
		mfiN1 = s.mfiValues[crossingIndex]
		stochKN1 = s.stochKValues[crossingIndex]
		stochDN1 = s.stochDValues[crossingIndex]
		timeN1 = time.Unix(s.klines[crossingIndex].Timestamp, 0)
	}

	// Capturer N-2 (crossingIndex - 1)
	if crossingIndex > 0 && crossingIndex-1 < len(s.cciValues) && crossingIndex-1 < len(s.mfiValues) &&
		crossingIndex-1 < len(s.stochKValues) && crossingIndex-1 < len(s.stochDValues) &&
		crossingIndex-1 < len(s.klines) {
		cciN2 = s.cciValues[crossingIndex-1]
		mfiN2 = s.mfiValues[crossingIndex-1]
		stochKN2 = s.stochKValues[crossingIndex-1]
		stochDN2 = s.stochDValues[crossingIndex-1]
		timeN2 = time.Unix(s.klines[crossingIndex-1].Timestamp, 0)
	}

	return &Signal{
		Index:            validationIndex,
		Time:             time.Unix(s.klines[validationIndex].Timestamp, 0), // Heure validation
		TimeN1:           timeN1,                                            // Heure N-1 (croisement)
		TimeN2:           timeN2,                                            // Heure N-2
		Type:             signalType,
		Price:            s.klines[validationIndex].Close,            // Prix de validation
		Volume:           s.klines[validationIndex].Volume,           // Volume SOL
		QuoteAssetVolume: s.klines[validationIndex].QuoteAssetVolume, // Volume USDT
		Confidence:       75,
		CCI:              cciN1,    // CCI N-1 (croisement)
		MFI:              mfiN1,    // MFI N-1 (croisement)
		StochK:           stochKN1, // Stoch K N-1 (croisement)
		StochD:           stochDN1, // Stoch D N-1 (croisement)
		CCIPrev:          cciN2,    // CCI N-2
		MFIPrev:          mfiN2,    // MFI N-2
		StochKPrev:       stochKN2, // Stoch K N-2
		StochDPrev:       stochDN2, // Stoch D N-2
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println("🎯 DEMO STRATÉGIE SCALPING - Triple Extrême")
	fmt.Println("============================================\n")

	config := DefaultConfig()

	// Connexion Binance
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("📡 Connexion Binance Futures...")
	client := binance.NewFuturesClient()

	// Récupération données
	fmt.Printf("📊 Récupération %d bougies %s sur %s...\n", config.CandlesCount, config.Timeframe, config.Symbol)
	futuresKlines, err := client.GetKlines(ctx, config.Symbol, config.Timeframe, config.CandlesCount)
	if err != nil {
		log.Fatalf("❌ Erreur: %v", err)
	}

	// Conversion
	klines := make([]Kline, len(futuresKlines))
	for i, fk := range futuresKlines {
		klines[i] = Kline{
			Timestamp:        fk.OpenTime.Unix(),
			Open:             fk.Open,
			High:             fk.High,
			Low:              fk.Low,
			Close:            fk.Close,
			Volume:           fk.Volume,
			QuoteAssetVolume: fk.QuoteAssetVolume,
		}
	}

	fmt.Printf("✅ %d bougies récupérées: %s → %s\n\n",
		len(klines),
		time.Unix(klines[0].Timestamp, 0).Format("2006-01-02 15:04"),
		time.Unix(klines[len(klines)-1].Timestamp, 0).Format("15:04"))

	// Stratégie
	strategy := NewScalpingStrategy(klines, config)
	signals := strategy.DetectSignals()

	// Résultats
	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Printf("\n📊 RÉSULTATS - %d SIGNAUX DÉTECTÉS\n", len(signals))
	fmt.Println("=" + string(make([]byte, 80)))

	if len(signals) == 0 {
		fmt.Println("Aucun signal détecté sur la période.")
		return
	}

	longCount := 0
	shortCount := 0

	fmt.Println("\n┌────────┬──────────────┬──────────────┬──────────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬────────┬──────────┬──────────┬──────────┐")
	fmt.Println("│  Type  │  Barre N-2   │  Barre N-1   │    Signal    │ CCI N-2 │ CCI N-1 │ MFI N-2 │ MFI N-1 │  K N-2  │  D N-2  │  K N-1  │  D N-1  │  Prix  │  Vol SOL │ Vol USDT │   Conf   │")
	fmt.Println("├────────┼──────────────┼──────────────┼──────────────┼─────────┼─────────┼─────────┼─────────┼─────────┼─────────┼─────────┼─────────┼────────┼──────────┼──────────┼──────────┤")

	for _, sig := range signals {
		fmt.Printf("│ %-6s │ %s │ %s │ %s │ %7.1f │ %7.1f │ %7.1f │ %7.1f │ %7.1f │ %7.1f │ %7.1f │ %7.1f │ %6.2f │ %8.0f │ %8.0f │    %d%%   │\n",
			sig.Type,
			sig.TimeN2.Format("02/01 15:04"),
			sig.TimeN1.Format("02/01 15:04"),
			sig.Time.Format("02/01 15:04"),
			sig.CCIPrev,
			sig.CCI,
			sig.MFIPrev,
			sig.MFI,
			sig.StochKPrev,
			sig.StochDPrev,
			sig.StochK,
			sig.StochD,
			sig.Price,
			sig.Volume,
			sig.QuoteAssetVolume,
			sig.Confidence)

		if sig.Type == "LONG" {
			longCount++
		} else {
			shortCount++
		}
	}

	fmt.Println("└────────┴──────────────┴──────────────┴──────────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴────────┴──────────┴──────────┴──────────┘")

	fmt.Printf("\n📈 STATISTIQUES:\n")
	fmt.Printf("   - LONG:  %d signaux (%.1f%%)\n", longCount, float64(longCount)/float64(len(signals))*100)
	fmt.Printf("   - SHORT: %d signaux (%.1f%%)\n", shortCount, float64(shortCount)/float64(len(signals))*100)
	fmt.Printf("   - Sélectivité: 1 signal toutes les ~%.1f bougies\n", float64(config.CandlesCount)/float64(len(signals)))

	fmt.Println("\n✅ Analyse terminée!")
}
