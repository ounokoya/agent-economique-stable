// Package binance provides tests for downloader functionality
package binance

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agent-economique/internal/shared"
)

// Test NewDownloader function - Configuration valide
func TestNewDownloader_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Test with empty config (should use defaults)
	config := shared.DownloadConfig{}
	
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if downloader == nil {
		t.Fatal("Expected downloader instance, got nil")
	}
	
	// Test that defaults are set
	if downloader.config.BaseURL == "" {
		t.Error("Expected default BaseURL to be set")
	}
}

// Test NewDownloader error handling
func TestNewDownloader_ErrorHandling(t *testing.T) {
	config := shared.DownloadConfig{}

	// Test with nil cache should return error
	_, err := NewDownloader(nil, config)
	if err == nil {
		t.Error("Expected error for nil cache, got nil")
	}
}

// Test DownloadFile - VRAIE FONCTION avec mock server
func TestDownloadFile_RealFunction_Success(t *testing.T) {
	// Mock HTTP server simulant Binance Vision
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simuler réponse Binance avec ZIP valide
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", "24")
		w.WriteHeader(http.StatusOK)
		
		// Contenu ZIP minimal valide
		zipContent := []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00test_content_here")
		w.Write(zipContent)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Configuration avec mock server
	config := shared.DownloadConfig{
		BaseURL:        server.URL,
		MaxRetries:     2,
		Timeout:        time.Second * 3,
		ChecksumVerify: false, // Désactiver pour test simple
	}

	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Requête avec données statiques
	request := shared.DownloadRequest{
		Symbol:    "SOLUSDT",
		DataType:  "klines", 
		Date:      "2023-06-01",
		Timeframe: "5m",
	}

	// TEST VRAIE FONCTION DownloadFile()
	result, err := downloader.DownloadFile(request)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	// Validation résultat selon code réel
	if result == nil {
		t.Fatal("Expected DownloadResult, got nil")
	}

	if !result.Success {
		t.Errorf("Expected Success=true, got false. Error: %s", result.Error)
	}

	if result.FilePath == "" {
		t.Error("Expected FilePath to be set")
	}

	if result.FileSize <= 0 {
		t.Error("Expected FileSize > 0")
	}

	// Vérifier que fichier a été créé
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist: %s", result.FilePath)
	}
}

// Test DownloadFile - Fichier déjà en cache
func TestDownloadFile_CacheHit(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Créer fichier mock en cache d'abord
	filePath := cache.GetFilePath("ETHUSDT", "klines", "2023-06-01", "5m")
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	
	// Écrire fichier mock
	mockContent := []byte("mock zip content for cache test")
	err = os.WriteFile(filePath, mockContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write mock file: %v", err)
	}

	// Mettre à jour index cache
	metadata := shared.FileMetadata{
		Symbol:    "ETHUSDT",
		DataType:  "klines",
		Date:      "2023-06-01", 
		Timeframe: "5m",
		FilePath:  filePath,
		FileSize:  int64(len(mockContent)),
	}
	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Fatalf("Failed to update cache index: %v", err)
	}

	// Configuration downloader
	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Requête pour fichier en cache
	request := shared.DownloadRequest{
		Symbol:    "ETHUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "5m",
	}

	// TEST CACHE HIT avec vraie fonction
	result, err := downloader.DownloadFile(request)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	// Validation cache hit
	if !result.Success {
		t.Error("Expected cache hit to be successful")
	}

	if result.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, result.FilePath)
	}

	if result.FileSize != int64(len(mockContent)) {
		t.Errorf("Expected FileSize %d, got %d", len(mockContent), result.FileSize)
	}
}

// Test DownloadFile - Erreur réseau
func TestDownloadFile_NetworkError(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Configuration avec URL invalide pour simuler erreur réseau
	config := shared.DownloadConfig{
		BaseURL:    "http://invalid-domain-that-does-not-exist-12345.com",
		MaxRetries: 1, // Une seule tentative pour test rapide
		Timeout:    time.Millisecond * 100, // Timeout court
	}

	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	request := shared.DownloadRequest{
		Symbol:    "BTCUSDT",
		DataType:  "trades",
		Date:      "2023-06-01",
	}

	// TEST GESTION ERREUR avec vraie fonction
	result, err := downloader.DownloadFile(request)
	
	// Doit retourner erreur
	if err == nil {
		t.Error("Expected network error, got nil")
	}

	if result.Success {
		t.Error("Expected Success=false for network error")
	}

	if result.Error == "" {
		t.Error("Expected Error message to be set")
	}
}

// Test CheckFileExists function - VRAIE FONCTION
func TestCheckFileExists_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test fichier inexistant
	request := shared.DownloadRequest{
		Symbol:    "SOLUSDT",
		DataType:  "klines", 
		Date:      "2023-06-01",
		Timeframe: "5m",
	}

	exists, filePath, err := downloader.CheckFileExists(request)
	if err != nil {
		t.Fatalf("CheckFileExists failed: %v", err)
	}

	if exists {
		t.Error("File should not exist initially")
	}

	if filePath != "" {
		t.Error("FilePath should be empty for non-existent file")
	}

	// Créer fichier et tester existence
	expectedPath := cache.GetFilePath("SOLUSDT", "klines", "2023-06-01", "5m")
	err = os.MkdirAll(filepath.Dir(expectedPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	err = os.WriteFile(expectedPath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Mettre à jour cache
	metadata := shared.FileMetadata{
		Symbol:    "SOLUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "5m",
		FilePath:  expectedPath,
		FileSize:  12,
	}
	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Fatalf("Failed to update cache: %v", err)
	}

	// Test fichier existant
	exists, filePath, err = downloader.CheckFileExists(request)
	if err != nil {
		t.Fatalf("CheckFileExists failed: %v", err)
	}

	if !exists {
		t.Error("File should exist now")
	}

	if filePath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, filePath)
	}
}

// Test GetDownloadURL function - VRAIE FONCTION
func TestGetDownloadURL_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{
		BaseURL: "https://data.binance.vision",
	}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test klines URL generation
	request := shared.DownloadRequest{
		Symbol:    "SOLUSDT",
		DataType:  "klines",
		Date:      "2023-06-01", 
		Timeframe: "5m",
	}

	url, err := downloader.GetDownloadURL(request)
	if err != nil {
		t.Fatalf("GetDownloadURL failed: %v", err)
	}

	// Vérifier que l'URL contient les composants attendus
	expectedComponents := []string{
		"https://data.binance.vision",
		"SOLUSDT",
		"klines",
		"2023-06-01",
		"5m",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(url, component) {
			t.Errorf("URL missing component '%s': %s", component, url)
		}
	}

	// Test trades URL generation
	tradesRequest := shared.DownloadRequest{
		Symbol:   "ETHUSDT",
		DataType: "trades",
		Date:     "2023-06-02",
	}

	tradesURL, err := downloader.GetDownloadURL(tradesRequest)
	if err != nil {
		t.Fatalf("GetDownloadURL failed for trades: %v", err)
	}

	expectedTradesComponents := []string{
		"https://data.binance.vision",
		"ETHUSDT",
		"trades",
		"2023-06-02",
	}

	for _, component := range expectedTradesComponents {
		if !strings.Contains(tradesURL, component) {
			t.Errorf("Trades URL missing component '%s': %s", component, tradesURL)
		}
	}
}

// Test ValidateChecksumFile function - VRAIE FONCTION
func TestValidateChecksumFile_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Créer fichier de test
	testFile := filepath.Join(tempDir, "checksum_test.zip")
	testContent := []byte("test content for checksum validation")
	err = os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Ajouter au cache index
	metadata := shared.FileMetadata{
		Symbol:   "BTCUSDT",
		DataType: "klines",
		Date:     "2023-06-01",
		FilePath: testFile,
		FileSize: int64(len(testContent)),
	}
	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Fatalf("Failed to update cache: %v", err)
	}

	// TEST VRAIE FONCTION ValidateChecksumFile
	isCorrupted, err := downloader.ValidateChecksumFile(testFile)
	if err != nil {
		t.Errorf("ValidateChecksumFile failed: %v", err)
	}

	// Fichier valide ne doit pas être corrompu
	if isCorrupted {
		t.Error("Valid file detected as corrupted")
	}

	// Test avec fichier invalide (path vide)
	_, err = downloader.ValidateChecksumFile("")
	if err == nil {
		t.Error("Expected error for empty file path")
	}
}

// Test BatchExists function - VRAIE FONCTION
func TestBatchExists_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Créer un fichier existant pour test
	existingPath := cache.GetFilePath("SOLUSDT", "klines", "2023-06-01", "5m")
	err = os.MkdirAll(filepath.Dir(existingPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	err = os.WriteFile(existingPath, []byte("existing file"), 0644)
	if err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	// Mettre à jour cache pour fichier existant
	metadata := shared.FileMetadata{
		Symbol:    "SOLUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "5m",
		FilePath:  existingPath,
		FileSize:  13,
	}
	err = cache.UpdateIndex(metadata)
	if err != nil {
		t.Fatalf("Failed to update cache: %v", err)
	}

	// Préparer requêtes batch
	requests := []shared.DownloadRequest{
		{Symbol: "SOLUSDT", DataType: "klines", Date: "2023-06-01", Timeframe: "5m"}, // Existe
		{Symbol: "ETHUSDT", DataType: "klines", Date: "2023-06-01", Timeframe: "5m"}, // N'existe pas
		{Symbol: "BTCUSDT", DataType: "trades", Date: "2023-06-01"},                  // N'existe pas
	}

	// TEST VRAIE FONCTION BatchExists
	results, err := downloader.BatchExists(requests)
	if err != nil {
		t.Fatalf("BatchExists failed: %v", err)
	}

	if results == nil {
		t.Fatal("Expected results map, got nil")
	}

	expectedResultCount := len(requests)
	if len(results) != expectedResultCount {
		t.Errorf("Expected %d results, got %d", expectedResultCount, len(results))
	}

	// Vérifier résultats
	solusdt5mKey := "SOLUSDT_klines_2023-06-01_5m"
	if exists, found := results[solusdt5mKey]; !found {
		t.Errorf("Missing result for %s", solusdt5mKey)
	} else if !exists {
		t.Errorf("Expected %s to exist", solusdt5mKey)
	}

	ethusdt5mKey := "ETHUSDT_klines_2023-06-01_5m"
	if exists, found := results[ethusdt5mKey]; !found {
		t.Errorf("Missing result for %s", ethusdt5mKey)
	} else if exists {
		t.Errorf("Expected %s to not exist", ethusdt5mKey)
	}
}

// Test GetCachedFilePath function - VRAIE FONCTION
func TestGetCachedFilePath_RealFunction(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test requête klines
	request := shared.DownloadRequest{
		Symbol:    "ADAUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "15m",
	}

	// TEST VRAIE FONCTION GetCachedFilePath
	filePath, err := downloader.GetCachedFilePath(request)
	if err != nil {
		t.Fatalf("GetCachedFilePath failed: %v", err)
	}

	if filePath == "" {
		t.Error("Expected non-empty file path")
	}

	// Vérifier structure path selon cache.GetFilePath()
	expectedPath := cache.GetFilePath("ADAUSDT", "klines", "2023-06-01", "15m")
	if filePath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, filePath)
	}

	// Test requête trades
	tradesRequest := shared.DownloadRequest{
		Symbol:   "DOTUSDT",
		DataType: "trades",
		Date:     "2023-06-02",
	}

	tradesPath, err := downloader.GetCachedFilePath(tradesRequest)
	if err != nil {
		t.Fatalf("GetCachedFilePath failed for trades: %v", err)
	}

	expectedTradesPath := cache.GetFilePath("DOTUSDT", "trades", "2023-06-02")
	if tradesPath != expectedTradesPath {
		t.Errorf("Expected trades path %s, got %s", expectedTradesPath, tradesPath)
	}

	// Test validation erreur
	invalidRequest := shared.DownloadRequest{} // Symbol vide
	_, err = downloader.GetCachedFilePath(invalidRequest)
	if err == nil {
		t.Error("Expected error for invalid request (empty symbol)")
	}
}

// Test CheckFileExists function
func TestCheckFileExists(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test non-existent file
	request := shared.DownloadRequest{
		Symbol:    "SOLUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "5m",
	}

	exists, _, err := downloader.CheckFileExists(request)
	if err != nil {
		t.Fatalf("CheckFileExists failed: %v", err)
	}

	if exists {
		t.Error("File should not exist initially")
	}
}

// Test BatchExists function
func TestBatchExists(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := InitializeCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	config := shared.DownloadConfig{}
	downloader, err := NewDownloader(cache, config)
	if err != nil {
		t.Fatalf("Failed to create downloader: %v", err)
	}

	// Test with multiple requests
	requests := []shared.DownloadRequest{
		{Symbol: "SOLUSDT", DataType: "klines", Date: "2023-06-01", Timeframe: "5m"},
		{Symbol: "ETHUSDT", DataType: "klines", Date: "2023-06-01", Timeframe: "5m"},
	}

	results, err := downloader.BatchExists(requests)
	if err != nil {
		t.Fatalf("BatchExists failed: %v", err)
	}

	if results == nil {
		t.Fatal("Expected results map, got nil")
	}

	if len(results) != len(requests) {
		t.Errorf("Expected %d results, got %d", len(requests), len(results))
	}

	// All files should not exist initially
	for symbol, exists := range results {
		if exists {
			t.Errorf("File for %s should not exist initially", symbol)
		}
	}
}
