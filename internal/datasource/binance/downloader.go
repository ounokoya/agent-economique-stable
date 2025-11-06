// Package binance provides download functionality for Binance Vision data
package binance

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/shared"
)

// Downloader manages downloads of Binance Vision data
type Downloader struct {
	cache  *CacheManager
	config shared.DownloadConfig
}

// NewDownloader creates a new Downloader instance
func NewDownloader(cache *CacheManager, config shared.DownloadConfig) (*Downloader, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}

	// Set default configuration if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://data.binance.vision"
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second * 5
	}
	if config.Timeout == 0 {
		config.Timeout = time.Minute * 10
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 5
	}

	return &Downloader{
		cache:  cache,
		config: config,
	}, nil
}

// DownloadFile downloads a single file if not already cached
func (d *Downloader) DownloadFile(request shared.DownloadRequest) (*shared.DownloadResult, error) {
	startTime := time.Now()
	
	result := &shared.DownloadResult{
		Request: request,
		Success: false,
	}

	// Validate request
	if err := d.validateRequest(request); err != nil {
		result.Error = fmt.Sprintf("invalid request: %v", err)
		return result, err
	}

	// Check if file already exists in cache
	var timeframe []string
	if request.Timeframe != "" {
		timeframe = []string{request.Timeframe}
	}
	
	if d.cache.FileExists(request.Symbol, request.DataType, request.Date, timeframe...) {
		// File exists, get path and size
		filePath := d.cache.GetFilePath(request.Symbol, request.DataType, request.Date, timeframe...)
		fileInfo, err := os.Stat(filePath)
		if err == nil {
			result.Success = true
			result.FilePath = filePath
			result.FileSize = fileInfo.Size()
			result.Duration = time.Since(startTime)
			return result, nil
		}
	}

	// File not in cache or corrupted, download it
	url := d.buildURL(request)
	filePath := d.cache.GetFilePath(request.Symbol, request.DataType, request.Date, timeframe...)
	
	// Create directory if not exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create directory: %v", err)
		return result, err
	}

	// Download with retries
	var lastErr error
	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(d.config.RetryDelay * time.Duration(attempt))
		}

		err := d.downloadFileWithTimeout(url, filePath)
		if err == nil {
			// Download successful
			break
		}
		
		lastErr = err
		if attempt < d.config.MaxRetries {
			// Remove partially downloaded file before retry
			os.Remove(filePath)
		}
	}

	if lastErr != nil {
		result.Error = fmt.Sprintf("download failed after %d attempts: %v", d.config.MaxRetries+1, lastErr)
		return result, lastErr
	}

	// Verify downloaded file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to stat downloaded file: %v", err)
		return result, err
	}

	// Calculate checksum
	checksum := ""
	if d.config.ChecksumVerify {
		checksum, err = d.calculateChecksum(filePath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to calculate checksum: %v", err)
			return result, err
		}
	}

	// Update cache index
	metadata := shared.FileMetadata{
		Symbol:     request.Symbol,
		DataType:   request.DataType,
		Date:       request.Date,
		Timeframe:  request.Timeframe,
		FilePath:   filePath,
		FileSize:   fileInfo.Size(),
		Checksum:   checksum,
		Downloaded: time.Now(),
		Verified:   d.config.ChecksumVerify,
	}

	if err := d.cache.UpdateIndex(metadata); err != nil {
		// Don't fail the download, just log the cache update error
		result.Error = fmt.Sprintf("warning: failed to update cache index: %v", err)
	}

	result.Success = true
	result.FilePath = filePath
	result.FileSize = fileInfo.Size()
	result.Checksum = checksum
	result.Duration = time.Since(startTime)

	return result, nil
}

// CheckFileExists verifies if a file exists locally (uses cache)
func (d *Downloader) CheckFileExists(request shared.DownloadRequest) (bool, string, error) {
	if err := d.validateRequest(request); err != nil {
		return false, "", err
	}

	var timeframe []string
	if request.Timeframe != "" {
		timeframe = []string{request.Timeframe}
	}

	exists := d.cache.FileExists(request.Symbol, request.DataType, request.Date, timeframe...)
	filePath := ""
	if exists {
		filePath = d.cache.GetFilePath(request.Symbol, request.DataType, request.Date, timeframe...)
	}

	return exists, filePath, nil
}

// GetDownloadURL returns the Binance Vision URL for a given request
func (d *Downloader) GetDownloadURL(request shared.DownloadRequest) (string, error) {
	if err := d.validateRequest(request); err != nil {
		return "", err
	}
	return d.buildURL(request), nil
}

// ValidateChecksumFile validates the checksum of a cached file
func (d *Downloader) ValidateChecksumFile(filePath string) (bool, error) {
	if filePath == "" {
		return false, fmt.Errorf("file path cannot be empty")
	}

	return d.cache.IsFileCorrupted(filePath)
}

// GetDownloadStats returns download statistics from cache
func (d *Downloader) GetDownloadStats() *shared.CacheStatistics {
	return d.cache.GetCacheStats()
}

// CleanupCorruptedFiles removes corrupted files using cache manager
func (d *Downloader) CleanupCorruptedFiles() error {
	return d.cache.CleanupCorrupted()
}

// BatchExists checks existence of multiple files efficiently
func (d *Downloader) BatchExists(requests []shared.DownloadRequest) (map[string]bool, error) {
	result := make(map[string]bool)
	
	for _, request := range requests {
		key := fmt.Sprintf("%s_%s_%s_%s", request.Symbol, request.DataType, request.Date, request.Timeframe)
		
		var timeframe []string
		if request.Timeframe != "" {
			timeframe = []string{request.Timeframe}
		}
		
		exists := d.cache.FileExists(request.Symbol, request.DataType, request.Date, timeframe...)
		result[key] = exists
	}
	
	return result, nil
}

// GetCachedFilePath returns the local path for a file (whether it exists or not)
func (d *Downloader) GetCachedFilePath(request shared.DownloadRequest) (string, error) {
	if err := d.validateRequest(request); err != nil {
		return "", err
	}

	var timeframe []string
	if request.Timeframe != "" {
		timeframe = []string{request.Timeframe}
	}

	return d.cache.GetFilePath(request.Symbol, request.DataType, request.Date, timeframe...), nil
}

// Helper functions

// validateRequest validates a download request
func (d *Downloader) validateRequest(request shared.DownloadRequest) error {
	if request.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if request.DataType != "klines" && request.DataType != "trades" {
		return fmt.Errorf("data type must be 'klines' or 'trades', got: %s", request.DataType)
	}
	if request.Date == "" {
		return fmt.Errorf("date cannot be empty")
	}
	if request.DataType == "klines" && request.Timeframe == "" {
		return fmt.Errorf("timeframe is required for klines data")
	}
	if request.DataType == "trades" && request.Timeframe != "" {
		return fmt.Errorf("timeframe should be empty for trades data")
	}
	
	return nil
}

// buildURL constructs the Binance Vision download URL
func (d *Downloader) buildURL(request shared.DownloadRequest) string {
	if request.DataType == "klines" {
		// https://data.binance.vision/data/futures/um/daily/klines/SOLUSDT/5m/SOLUSDT-5m-2023-06-01.zip
		return fmt.Sprintf("%s/data/futures/um/daily/klines/%s/%s/%s-%s-%s.zip",
			d.config.BaseURL, request.Symbol, request.Timeframe,
			request.Symbol, request.Timeframe, request.Date)
	} else {
		// https://data.binance.vision/data/futures/um/daily/trades/SOLUSDT/SOLUSDT-trades-2023-06-01.zip
		return fmt.Sprintf("%s/data/futures/um/daily/trades/%s/%s-trades-%s.zip",
			d.config.BaseURL, request.Symbol, request.Symbol, request.Date)
	}
}

// downloadFileWithTimeout downloads a file with timeout
func (d *Downloader) downloadFileWithTimeout(url, filePath string) error {
	client := &http.Client{
		Timeout: d.config.Timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy data
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of a file
func (d *Downloader) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
