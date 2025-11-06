// Package binance provides tests for cache functionality
package binance

import (
	"os"
	"path/filepath"
	"testing"

	"agent-economique/internal/shared"
)

// Test InitializeCache function - Test 1: Initialisation cache vide  
func TestInitializeCache_EmptyCache(t *testing.T) {
	tempDir := t.TempDir()

	// Test with valid path - should create root directory
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	if cache == nil {
		t.Fatal("Cache instance should not be nil")
	}

	// Verify root directory creation (not nested structure)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Root directory not created at %s", tempDir)
	}
	
	// Test basic functionality
	exists := cache.FileExists("SOLUSDT", "klines", "2023-06-01", "5m")
	if exists {
		t.Error("File should not exist in empty cache")
	}
}

// Test InitializeCache function - Test 2: Initialisation cache existant
func TestInitializeCache_ExistingCache(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create cache first time
	cache1, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache first time: %v", err)
	}

	// Initialize again - should load existing
	cache2, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize existing cache: %v", err)
	}

	if cache1 == nil || cache2 == nil {
		t.Fatal("Both cache instances should be valid")
	}
}

// Test InitializeCache error handling
func TestInitializeCache_ErrorHandling(t *testing.T) {
	// Test with empty path should return error
	_, err := InitializeCache("")
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	// Test with invalid permissions (if possible)
	invalidPath := "/root/no_permission"
	_, err = InitializeCache(invalidPath)
	if err == nil {
		t.Log("Permission test skipped - may not be applicable in test environment")
	}
}

// Test FileExists function
func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test FileExists for non-existent file
	exists := cache.FileExists("SOLUSDT", "klines", "2023-06-01", "5m")
	if exists {
		t.Error("File should not exist initially")
	}
}

// Test GetFilePath function - VRAIE FONCTION
func TestGetFilePath_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test klines path generation (vraie fonction)
	klinesPath := cache.GetFilePath("SOLUSDT", "klines", "2023-06-01", "5m")
	if klinesPath == "" {
		t.Error("Expected non-empty klines path")
	}

	// Vérifier structure selon code réel: binance/futures_um/klines/SYMBOL/TIMEFRAME/
	expectedComponents := []string{"binance", "futures_um", "klines", "SOLUSDT", "5m", "SOLUSDT-5m-2023-06-01.zip"}
	for _, component := range expectedComponents {
		if !containsString(klinesPath, component) {
			t.Errorf("Path missing component '%s': %s", component, klinesPath)
		}
	}

	// Test trades path generation (vraie fonction)
	tradesPath := cache.GetFilePath("ETHUSDT", "trades", "2023-06-02")
	if tradesPath == "" {
		t.Error("Expected non-empty trades path")
	}

	// Structure trades: binance/futures_um/trades/SYMBOL/SYMBOL-trades-DATE.zip
	expectedTradesComponents := []string{"binance", "futures_um", "trades", "ETHUSDT", "ETHUSDT-trades-2023-06-02.zip"}
	for _, component := range expectedTradesComponents {
		if !containsString(tradesPath, component) {
			t.Errorf("Trades path missing component '%s': %s", component, tradesPath)
		}
	}
}

// Test UpdateIndex function - VRAIE FONCTION
func TestUpdateIndex_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test avec métadonnées valides (vraie fonction)
	metadata := shared.FileMetadata{
		Symbol:   "SOLUSDT",
		DataType: "klines", 
		Date:     "2023-06-01",
		Timeframe: "5m",
		FilePath: "/tmp/test.zip",
		FileSize: 1024,
	}

	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Errorf("UpdateIndex failed: %v", err)
	}

	// Vérifier que l'index a été mis à jour
	exists := cache.FileExists("SOLUSDT", "klines", "2023-06-01", "5m")
	if !exists {
		t.Error("File should exist in index after UpdateIndex")
	}
}

// Test GetCacheStats function - VRAIE FONCTION  
func TestGetCacheStats_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test vraie fonction GetCacheStats
	stats := cache.GetCacheStats()
	if stats == nil {
		t.Error("Expected cache stats, got nil")
	}

	// Validation structure selon code réel
	if stats.TotalFiles < 0 {
		t.Error("TotalFiles should not be negative")
	}

	if stats.TotalSizeMB < 0 {
		t.Error("TotalSizeMB should not be negative") 
	}
}

// Test IsFileCorrupted function - VRAIE FONCTION
func TestIsFileCorrupted_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Créer fichier de test
	testFile := filepath.Join(tempDir, "test.zip")
	testContent := []byte("test zip content for corruption check")
	err = os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Ajouter à l'index avec métadonnées correctes
	metadata := shared.FileMetadata{
		Symbol:   "TESTUSDT",
		DataType: "klines",
		Date:     "2023-06-01", 
		FilePath: testFile,
		FileSize: int64(len(testContent)),
		Checksum: "", // Pas de checksum pour test simple
	}
	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Fatalf("Failed to update index: %v", err)
	}

	// TEST VRAIE FONCTION IsFileCorrupted
	corrupted, err := cache.IsFileCorrupted(testFile)
	if err != nil {
		t.Errorf("IsFileCorrupted failed: %v", err)
	}

	// Fichier valide ne doit pas être corrompu
	if corrupted {
		t.Error("Valid file detected as corrupted")
	}

	// Test avec taille modifiée (simuler corruption)
	corruptedContent := append(testContent, []byte("extra content")...)
	err = os.WriteFile(testFile, corruptedContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted content: %v", err)
	}

	// Maintenant doit détecter corruption
	corrupted, err = cache.IsFileCorrupted(testFile)
	if err != nil {
		t.Errorf("IsFileCorrupted failed on size mismatch: %v", err)
	}

	if !corrupted {
		t.Error("File with wrong size not detected as corrupted")
	}
}

// Test CleanupCorrupted function - VRAIE FONCTION
func TestCleanupCorrupted_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// TEST VRAIE FONCTION CleanupCorrupted sur cache vide
	err = cache.CleanupCorrupted()
	if err != nil {
		t.Errorf("CleanupCorrupted should not fail on empty cache: %v", err)
	}
}

// Helper function to check if string contains all substrings
func containsAll(str string, substrings []string) bool {
	for _, substr := range substrings {
		if !containsString(str, substr) {
			return false
		}
	}
	return true
}

// Helper function for string contains check
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (substr == "" || findSubstring(str, substr))
}

// Simple substring finder
func findSubstring(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(str) < len(substr) {
		return false
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
