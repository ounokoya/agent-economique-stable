// Package tests provides tests for configuration functionality
package tests

import (
	"os"
	"testing"
	"time"

	"agent-economique/internal/shared"
)

// TestConfigLoading tests YAML configuration loading
func TestConfigLoading(t *testing.T) {
	// Test loading existing config
	config, err := shared.LoadConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate loaded configuration
	if config.BinanceData.CacheRoot != "data/binance" {
		t.Errorf("Expected cache_root 'data/binance', got '%s'", config.BinanceData.CacheRoot)
	}

	expectedSymbols := []string{"SOLUSDT", "SUIUSDT", "ETHUSDT"}
	if len(config.BinanceData.Symbols) != len(expectedSymbols) {
		t.Errorf("Expected %d symbols, got %d", len(expectedSymbols), len(config.BinanceData.Symbols))
	}

	for i, expected := range expectedSymbols {
		if i < len(config.BinanceData.Symbols) && config.BinanceData.Symbols[i] != expected {
			t.Errorf("Expected symbol %s, got %s", expected, config.BinanceData.Symbols[i])
		}
	}

	// Test downloader config conversion
	downloadConfig, err := config.BinanceData.Downloader.ToDownloadConfig()
	if err != nil {
		t.Errorf("Failed to convert download config: %v", err)
	}

	if downloadConfig.BaseURL != "https://data.binance.vision" {
		t.Errorf("Expected BaseURL 'https://data.binance.vision', got '%s'", downloadConfig.BaseURL)
	}

	if downloadConfig.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", downloadConfig.MaxRetries)
	}

	if downloadConfig.Timeout != 10*time.Minute {
		t.Errorf("Expected Timeout 10m, got %v", downloadConfig.Timeout)
	}

	// Test streaming config conversion
	streamingConfig := config.BinanceData.Streaming.ToStreamingConfig()
	if streamingConfig.BufferSize != 65536 {
		t.Errorf("Expected BufferSize 65536, got %d", streamingConfig.BufferSize)
	}

	if streamingConfig.MaxMemoryMB != 100 {
		t.Errorf("Expected MaxMemoryMB 100, got %d", streamingConfig.MaxMemoryMB)
	}

	// Test strategy configuration
	if config.Strategy.Name != "STOCH_MFI_CCI" {
		t.Errorf("Expected strategy 'STOCH_MFI_CCI', got '%s'", config.Strategy.Name)
	}
	
	// Test STOCH/MFI/CCI indicator configuration
	if config.Strategy.Indicators.Stochastic.PeriodK != 14 {
		t.Errorf("Expected Stochastic PeriodK 14, got %d", config.Strategy.Indicators.Stochastic.PeriodK)
	}
	
	if config.Strategy.Indicators.MFI.Period != 14 {
		t.Errorf("Expected MFI Period 14, got %d", config.Strategy.Indicators.MFI.Period)
	}
	
	// Test signal generation configuration
	if config.Strategy.SignalGeneration.MinConfidence != 0.7 {
		t.Errorf("Expected MinConfidence 0.7, got %f", config.Strategy.SignalGeneration.MinConfidence)
	}
	
	// Test position management configuration
	if !config.Strategy.PositionManagement.EnableDynamicAdjustments {
		t.Error("Expected EnableDynamicAdjustments to be true")
	}

	t.Logf("✅ Configuration YAML chargée et validée correctement")
}

// TestConfigFileNotFound tests behavior when config file doesn't exist
func TestConfigFileNotFound(t *testing.T) {
	_, err := shared.LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}

	t.Logf("✅ Gestion d'erreur pour fichier config inexistant: %v", err)
}

// TestInvalidConfigYAML tests behavior with invalid YAML
func TestInvalidConfigYAML(t *testing.T) {
	// Create temporary invalid YAML file
	tempFile, err := os.CreateTemp("", "invalid_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write invalid YAML
	invalidYAML := `
binance_data:
  cache_root: "test"
  symbols: [
    - "SOLUSDT"
    - "ETHUSDT"
  # Missing closing bracket
`
	if _, err := tempFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Try to load invalid config
	_, err = shared.LoadConfig(tempFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}

	t.Logf("✅ Gestion d'erreur pour YAML invalide: %v", err)
}

// TestConfigDefaults tests that modules still work with partial configuration
func TestConfigDefaults(t *testing.T) {
	// Create minimal valid config
	tempFile, err := os.CreateTemp("", "minimal_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	minimalYAML := `
binance_data:
  cache_root: "test_cache"
  symbols: ["SOLUSDT"]
  timeframes: ["5m"]
  downloader:
    base_url: "https://test.example.com"
  streaming:
    buffer_size: 1024
  validation:
    max_price_deviation: 5.0
strategy:
  name: "TEST"
data_period:
  start_date: "2023-01-01"
  end_date: "2023-12-31"
`
	if _, err := tempFile.WriteString(minimalYAML); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Load minimal config
	config, err := shared.LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load minimal config: %v", err)
	}

	// Test conversion with defaults
	downloadConfig, err := config.BinanceData.Downloader.ToDownloadConfig()
	if err != nil {
		t.Errorf("Failed to convert minimal download config: %v", err)
	}

	// Should have default values for unspecified fields
	if downloadConfig.MaxRetries == 0 {
		t.Error("Expected default MaxRetries to be set")
	}

	if downloadConfig.BaseURL != "https://test.example.com" {
		t.Errorf("Expected BaseURL from config, got %s", downloadConfig.BaseURL)
	}

	t.Logf("✅ Configuration minimale avec valeurs par défaut fonctionne")
}
