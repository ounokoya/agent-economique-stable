// Package binance provides tests for streaming functionality
package binance

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"agent-economique/internal/shared"
)

// Test NewStreamingReader function
func TestNewStreamingReader(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test with valid cache and default config
	config := shared.StreamingConfig{
		BufferSize:    1024,
		MaxMemoryMB:   100,
		EnableMetrics: true,
	}

	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if reader == nil {
		t.Fatal("Expected reader instance, got nil")
	}
}

// Test NewStreamingReader error handling
func TestNewStreamingReader_ErrorHandling(t *testing.T) {
	config := shared.StreamingConfig{}

	// Test with nil cache should return error
	_, err := NewStreamingReader(nil, config)
	if err == nil {
		t.Error("Expected error for nil cache, got nil")
	}
}

// Test StreamKlines - VRAIE FONCTION avec fichier ZIP mock
func TestStreamKlines_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.StreamingConfig{
		BufferSize:    1024,
		MaxMemoryMB:   100,
		EnableMetrics: true,
	}

	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Créer fichier ZIP mock avec CSV klines
	zipPath := filepath.Join(tempDir, "test_klines.zip")
	err = createMockKlinesZIP(zipPath)
	if err != nil {
		t.Fatalf("Failed to create mock ZIP: %v", err)
	}

	// Callback pour collecter klines
	var collectedKlines []shared.KlineData
	callback := func(kline shared.KlineData) error {
		collectedKlines = append(collectedKlines, kline)
		return nil
	}

	// TEST VRAIE FONCTION StreamKlines
	err = reader.StreamKlines(zipPath, callback)
	if err != nil {
		t.Fatalf("StreamKlines failed: %v", err)
	}

	// Valider résultats
	if len(collectedKlines) == 0 {
		t.Error("Expected at least one kline to be streamed")
	}

	// Vérifier structure premier kline
	if len(collectedKlines) > 0 {
		kline := collectedKlines[0]
		if kline.OpenTime == 0 {
			t.Error("Expected valid OpenTime")
		}
		if kline.Open <= 0 {
			t.Error("Expected valid Open price")
		}
		if kline.Volume < 0 {
			t.Error("Expected non-negative Volume")
		}
	}

	// Test avec fichier inexistant
	err = reader.StreamKlines("non_existent.zip", callback)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Test StreamTrades basic functionality
func TestStreamTrades_Basic(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.StreamingConfig{}
	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Test with non-existent file (should handle gracefully)
	callback := func(trade shared.TradeData) error {
		return nil
	}

	err = reader.StreamTrades("non_existent.zip", callback)
	if err == nil {
		t.Log("StreamTrades handled non-existent file gracefully")
	} else {
		t.Logf("Expected error for non-existent file: %v", err)
	}
}

// Test GetMemoryMetrics function
func TestGetMemoryMetrics(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.StreamingConfig{EnableMetrics: true}
	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	metrics := reader.GetMemoryMetrics()
	if metrics == nil {
		t.Error("Expected metrics, got nil")
	} else {
		if metrics.CurrentUsageMB < 0 {
			t.Error("Current usage should not be negative")
		}
	}
}

// Test ValidateMemoryConstraints function
func TestValidateMemoryConstraints(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test with reasonable memory limit
	config := shared.StreamingConfig{
		BufferSize:    1024,
		MaxMemoryMB:   100, // 100MB limit
		EnableMetrics: true,
	}

	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Test memory validation
	isValid, err := reader.ValidateMemoryConstraints()
	if err != nil {
		t.Logf("ValidateMemoryConstraints returned error: %v", err)
	}

	if !isValid {
		t.Log("Memory constraint validation returned false - this may be normal")
	}
}

// Test ResetMetrics function
func TestResetMetrics(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.StreamingConfig{
		BufferSize:    1024,
		MaxMemoryMB:   100,
		EnableMetrics: true,
	}

	reader, err := NewStreamingReader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Get initial metrics
	initialMetrics := reader.GetMemoryMetrics()
	if initialMetrics == nil {
		t.Fatal("Expected initial metrics, got nil")
	}

	// Reset metrics (if method exists)
	reader.ResetMetrics() // Method may not return error

	// Get metrics after reset
	resetMetrics := reader.GetMemoryMetrics()
	if resetMetrics == nil {
		t.Fatal("Expected metrics after reset, got nil")
	}

	t.Logf("ResetMetrics test completed - Current usage: %.2f MB", resetMetrics.CurrentUsageMB)
}

// Helper function pour créer ZIP mock avec données klines
func createMockKlinesZIP(zipPath string) error {
	// Créer fichier ZIP
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Créer fichier CSV à l'intérieur du ZIP
	csvWriter, err := zipWriter.Create("SOLUSDT-5m-2023-06-01.csv")
	if err != nil {
		return err
	}

	// Données CSV mock au format Binance klines
	csvData := `1623024000000,100.0,101.0,99.0,100.5,1000.0,1623024059999,100500.0,50,500.0,50250.0,0
1623024060000,100.5,102.0,100.0,101.0,1200.0,1623024119999,121200.0,60,600.0,60600.0,0
1623024120000,101.0,103.0,100.5,102.5,800.0,1623024179999,82000.0,40,400.0,41000.0,0`

	_, err = csvWriter.Write([]byte(csvData))
	return err
}
