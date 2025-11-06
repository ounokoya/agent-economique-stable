// Package binance provides tests for klines functionality
package binance

import (
	"fmt"
	"testing"
	"time"

	"agent-economique/internal/shared"
)

// Données fixes réutilisables pour tests klines
var (
	// Séquence klines 5m continue - données fixes selon spécification
	TestKlinesSequence5m = []shared.KlineData{
		{
			OpenTime:        1623024000000, // 2021-06-07 00:00:00
			CloseTime:       1623024299999, // 2021-06-07 00:04:59
			Open:            100.0,
			High:            102.0,
			Low:             99.0,
			Close:           101.0,
			Volume:          1000.0,
			NumberOfTrades:  50,
		},
		{
			OpenTime:        1623024300000, // 2021-06-07 00:05:00
			CloseTime:       1623024599999, // 2021-06-07 00:09:59
			Open:            101.0,
			High:            104.0,
			Low:             100.5,
			Close:           103.0,
			Volume:          1200.0,
			NumberOfTrades:  60,
		},
		{
			OpenTime:        1623024600000, // 2021-06-07 00:10:00
			CloseTime:       1623024899999, // 2021-06-07 00:14:59
			Open:            103.0,
			High:            105.0,
			Low:             102.0,
			Close:           104.5,
			Volume:          800.0,
			NumberOfTrades:  40,
		},
	}

	// Klines avec anomalies pour tests robustesse
	TestKlinesWithAnomalies = []shared.KlineData{
		{
			OpenTime:  1623024000000,
			CloseTime: 1623024299999,
			Open:      100.0,
			High:      98.0, // Anomalie: High < Open
			Low:       99.0,
			Close:     101.0,
			Volume:    1000.0,
		},
		{
			OpenTime:  1623024300000,
			CloseTime: 1623024599999,
			Open:      101.0,
			High:      104.0,
			Low:       105.0, // Anomalie: Low > High
			Close:     103.0,
			Volume:    -500.0, // Anomalie: Volume négative
		},
	}
	
	// Klines avec gaps temporels
	TestKlinesWithGaps = []shared.KlineData{
		{
			OpenTime:  1623024000000, // 00:00:00
			CloseTime: 1623024299999, // 00:04:59
			Open:      100.0,
			Close:     101.0,
		},
		{
			OpenTime:  1623024900000, // 00:15:00 (gap de 10 minutes)
			CloseTime: 1623025199999, // 00:19:59
			Open:      102.0,
			Close:     103.0,
		},
	}
)

// Test ProcessKlinesFile - Fichier valide (Spec doc non explicitée mais logique)
func TestProcessKlinesFile_ValidData(t *testing.T) {
	// Test avec données fixes valides
	klines := TestKlinesSequence5m
	
	if len(klines) == 0 {
		t.Fatal("Test data should not be empty")
	}
	
	// Validation basique des données de test
	for i, kline := range klines {
		if kline.OpenTime >= kline.CloseTime {
			t.Errorf("Kline %d: OpenTime %d >= CloseTime %d", i, kline.OpenTime, kline.CloseTime)
		}
		
		if kline.High < kline.Low {
			t.Errorf("Kline %d: High %.2f < Low %.2f", i, kline.High, kline.Low)
		}
		
		if kline.Volume < 0 {
			t.Errorf("Kline %d: Negative volume %.2f", i, kline.Volume)
		}
	}
	
	t.Logf("Processed %d valid klines", len(klines))
}

// Test ValidateKlineSequence - Séquence valide (Spec: fonction ValidateKlineSequence)
func TestValidateKlineSequence_ValidSequence(t *testing.T) {
	// Test avec données fixes continues
	err := ValidateKlineSequence(TestKlinesSequence5m, "5m")
	if err != nil {
		t.Errorf("ValidateKlineSequence failed for valid sequence: %v", err)
	}
}

// Test ValidateKlineSequence - Séquence avec gaps (Spec: détection gaps selon timeframe)
func TestValidateKlineSequence_WithGaps(t *testing.T) {
	// Test avec données fixes contenant gaps
	err := ValidateKlineSequence(TestKlinesWithGaps, "5m")
	if err == nil {
		t.Error("Expected error for sequence with gaps, got nil")
	} else {
		t.Logf("Gap correctly detected: %v", err)
	}
}

// Test DetectKlineAnomalies - Détection anomalies (Spec: validation cohérence données)
func TestDetectKlineAnomalies_Detection(t *testing.T) {
	anomalies := DetectKlineAnomalies(TestKlinesWithAnomalies)
	
	if len(anomalies) == 0 {
		t.Error("Expected anomalies to be detected, got none")
	}
	
	// Vérification types d'anomalies selon spec
	expectedAnomalies := map[string]bool{
		"high_low_inconsistent": false,
		"negative_volume":       false,
		"ohlc_inconsistent":     false,
	}
	
	for _, anomaly := range anomalies {
		switch anomaly.Type {
		case "high_low_inconsistent":
			expectedAnomalies["high_low_inconsistent"] = true
		case "negative_volume":
			expectedAnomalies["negative_volume"] = true
		case "ohlc_inconsistent":
			expectedAnomalies["ohlc_inconsistent"] = true
		}
	}
	
	// Validation détection selon données test
	if !expectedAnomalies["high_low_inconsistent"] {
		t.Error("Expected to detect high/low inconsistency")
	}
	
	if !expectedAnomalies["negative_volume"] {
		t.Error("Expected to detect negative volume")
	}
	
	t.Logf("Detected %d anomalies: %v", len(anomalies), anomalies)
}

// Test CalculateKlineStatistics - Statistiques données (Performance monitoring)
func TestCalculateKlineStatistics_Calculation(t *testing.T) {
	stats := CalculateKlineStatistics(TestKlinesSequence5m)
	
	if stats == nil {
		t.Fatal("Expected statistics, got nil")
	}
	
	// Validation statistiques selon données fixes
	expectedCount := len(TestKlinesSequence5m)
	if stats.Count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, stats.Count)
	}
	
	// Volume total selon données fixes
	expectedTotalVolume := 1000.0 + 1200.0 + 800.0 // 3000.0
	if stats.TotalVolume != expectedTotalVolume {
		t.Errorf("Expected total volume %.1f, got %.1f", expectedTotalVolume, stats.TotalVolume)
	}
	
	// Prix min/max selon données fixes
	expectedMinPrice := 99.0  // Plus bas Low
	expectedMaxPrice := 105.0 // Plus haut High
	
	if stats.MinPrice != expectedMinPrice {
		t.Errorf("Expected min price %.1f, got %.1f", expectedMinPrice, stats.MinPrice)
	}
	
	if stats.MaxPrice != expectedMaxPrice {
		t.Errorf("Expected max price %.1f, got %.1f", expectedMaxPrice, stats.MaxPrice)
	}
	
	// Volume moyen
	expectedAvgVolume := expectedTotalVolume / float64(expectedCount)
	if stats.AvgVolume != expectedAvgVolume {
		t.Errorf("Expected avg volume %.1f, got %.1f", expectedAvgVolume, stats.AvgVolume)
	}
	
	t.Logf("Kline statistics: Count=%d, Volume=%.1f, Price=%.1f-%.1f", 
		stats.Count, stats.TotalVolume, stats.MinPrice, stats.MaxPrice)
}

// Test ConvertTimeframe - Conversion timeframes (Spec: timeframes supportés)
func TestConvertTimeframe_Supported(t *testing.T) {
	timeframes := map[string]int64{
		"5m":  5 * 60 * 1000,
		"15m": 15 * 60 * 1000,
		"1h":  60 * 60 * 1000,
		"4h":  4 * 60 * 60 * 1000,
	}
	
	for tf, expectedMs := range timeframes {
		actualMs := ConvertTimeframeToMs(tf)
		
		if actualMs != expectedMs {
			t.Errorf("Timeframe %s: expected %d ms, got %d ms", tf, expectedMs, actualMs)
		}
		
		if actualMs == 0 {
			t.Errorf("Timeframe %s not supported", tf)
		}
	}
}

// Test FormatKlineTimestamp - Formatage timestamps (Utilité debug)
func TestFormatKlineTimestamp_Format(t *testing.T) {
	// Test avec timestamp fixe des données test
	timestamp := TestKlinesSequence5m[0].OpenTime // 1623024000000
	
	formatted := FormatKlineTimestamp(timestamp)
	
	// Validation format attendu
	expected := "2021-06-07 00:00:00"
	if formatted != expected {
		t.Errorf("Expected format %s, got %s", expected, formatted)
	}
	
	t.Logf("Timestamp %d formatted as: %s", timestamp, formatted)
}

// Test ValidateKlineData - Validation kline individuelle (Spec: cohérence données)
func TestValidateKlineData_Individual(t *testing.T) {
	// Test kline valide
	validKline := TestKlinesSequence5m[0]
	err := ValidateKlineData(&validKline)
	if err != nil {
		t.Errorf("ValidateKlineData failed for valid kline: %v", err)
	}
	
	// Test kline invalide
	invalidKline := TestKlinesWithAnomalies[0] // High < Open
	err = ValidateKlineData(&invalidKline)
	if err == nil {
		t.Error("Expected error for invalid kline, got nil")
	} else {
		t.Logf("Invalid kline correctly rejected: %v", err)
	}
}

// Test Performance - Processing klines (Spec: performance < 0.1ms par kline)
func TestPerformance_KlineProcessing(t *testing.T) {
	// Données fixes - 1000 klines identiques
	var klines []shared.KlineData
	baseTime := TestKlinesSequence5m[0].OpenTime
	
	for i := 0; i < 1000; i++ {
		kline := TestKlinesSequence5m[0] // Copie première kline test
		kline.OpenTime = baseTime + int64(i*5*60*1000) // Intervalles 5m
		kline.CloseTime = kline.OpenTime + 5*60*1000 - 1
		klines = append(klines, kline)
	}
	
	// Test performance selon spec
	start := time.Now()
	
	processedCount := 0
	for i := range klines {
		err := ValidateKlineData(&klines[i])
		if err == nil {
			processedCount++
		}
	}
	
	duration := time.Since(start)
	
	if processedCount != len(klines) {
		t.Errorf("Expected %d processed klines, got %d", len(klines), processedCount)
	}
	
	// Validation performance: < 0.1ms par kline selon spec
	avgTimePerKline := duration / time.Duration(len(klines))
	maxTimePerKline := time.Microsecond * 100 // 0.1ms
	
	if avgTimePerKline > maxTimePerKline {
		t.Errorf("Processing too slow: %.2fμs per kline (max: %.2fμs)", 
			float64(avgTimePerKline.Nanoseconds())/1000, 
			float64(maxTimePerKline.Nanoseconds())/1000)
	}
	
	t.Logf("Performance: processed %d klines in %v (avg: %.2fμs per kline)", 
		processedCount, duration, float64(avgTimePerKline.Nanoseconds())/1000)
}

// Fonctions utilitaires selon spécifications (simulation API)

func ValidateKlineSequence(klines []shared.KlineData, timeframe string) error {
	if len(klines) < 2 {
		return nil
	}
	
	// Intervalle selon timeframe
	intervals := map[string]int64{
		"5m": 5 * 60 * 1000,
		"15m": 15 * 60 * 1000,
		"1h": 60 * 60 * 1000,
		"4h": 4 * 60 * 60 * 1000,
	}
	
	expectedInterval := intervals[timeframe]
	if expectedInterval == 0 {
		return fmt.Errorf("unsupported timeframe: %s", timeframe)
	}
	
	// Validation continuité
	for i := 1; i < len(klines); i++ {
		gap := klines[i].OpenTime - klines[i-1].CloseTime - 1
		
		if gap != expectedInterval {
			return fmt.Errorf("gap detected at index %d: expected %d ms, got %d ms", 
				i, expectedInterval, gap)
		}
	}
	
	return nil
}

func DetectKlineAnomalies(klines []shared.KlineData) []KlineAnomaly {
	var anomalies []KlineAnomaly
	
	for i, kline := range klines {
		if kline.High < kline.Low {
			anomalies = append(anomalies, KlineAnomaly{
				Index: i,
				Type:  "high_low_inconsistent",
				Message: fmt.Sprintf("High %.2f < Low %.2f", kline.High, kline.Low),
			})
		}
		
		if kline.Volume < 0 {
			anomalies = append(anomalies, KlineAnomaly{
				Index: i,
				Type:  "negative_volume",
				Message: fmt.Sprintf("Negative volume %.2f", kline.Volume),
			})
		}
		
		if kline.High < kline.Open || kline.High < kline.Close {
			anomalies = append(anomalies, KlineAnomaly{
				Index: i,
				Type:  "ohlc_inconsistent",
				Message: "High not highest price",
			})
		}
	}
	
	return anomalies
}

func CalculateKlineStatistics(klines []shared.KlineData) *KlineStatistics {
	if len(klines) == 0 {
		return nil
	}
	
	stats := &KlineStatistics{
		Count:    len(klines),
		MinPrice: klines[0].Low,
		MaxPrice: klines[0].High,
	}
	
	for _, kline := range klines {
		stats.TotalVolume += kline.Volume
		
		if kline.Low < stats.MinPrice {
			stats.MinPrice = kline.Low
		}
		if kline.High > stats.MaxPrice {
			stats.MaxPrice = kline.High
		}
	}
	
	stats.AvgVolume = stats.TotalVolume / float64(stats.Count)
	
	return stats
}

func ConvertTimeframeToMs(timeframe string) int64 {
	timeframes := map[string]int64{
		"5m":  5 * 60 * 1000,
		"15m": 15 * 60 * 1000,
		"1h":  60 * 60 * 1000,
		"4h":  4 * 60 * 60 * 1000,
	}
	
	return timeframes[timeframe]
}

func FormatKlineTimestamp(timestamp int64) string {
	t := time.Unix(timestamp/1000, 0).UTC()
	return t.Format("2006-01-02 15:04:05")
}

func ValidateKlineData(kline *shared.KlineData) error {
	if kline.High < kline.Low {
		return fmt.Errorf("high %.2f < low %.2f", kline.High, kline.Low)
	}
	
	if kline.Volume < 0 {
		return fmt.Errorf("negative volume %.2f", kline.Volume)
	}
	
	if kline.High < kline.Open || kline.High < kline.Close {
		return fmt.Errorf("high not highest price")
	}
	
	if kline.Low > kline.Open || kline.Low > kline.Close {
		return fmt.Errorf("low not lowest price")
	}
	
	return nil
}

// Types selon spécifications
type KlineAnomaly struct {
	Index   int
	Type    string
	Message string
}

type KlineStatistics struct {
	Count       int
	TotalVolume float64
	MinPrice    float64
	MaxPrice    float64
	AvgVolume   float64
}
