// Package testutils provides mock implementations for testing
package testutils

import (
	"fmt"
	"time"

	"agent-economique/internal/shared"
)

// MockCacheManager implements the CacheManager interface for testing
type MockCacheManager struct {
	files     map[string]*shared.FileMetadata
	stats     *shared.CacheStatistics
	corrupted []string
}

// NewMockCacheManager creates a new mock cache manager
func NewMockCacheManager() *MockCacheManager {
	return &MockCacheManager{
		files: make(map[string]*shared.FileMetadata),
		stats: &shared.CacheStatistics{
			TotalFiles:     0,
			TotalSizeMB:    0,
			HitRate:        1.0,
			CorruptedFiles: 0,
			LastUpdateTime: time.Now(),
		},
		corrupted: []string{},
	}
}

// FileExists implements CacheManager interface
func (m *MockCacheManager) FileExists(symbol, dataType, date string, timeframe ...string) bool {
	key := m.generateKey(symbol, dataType, date, timeframe...)
	_, exists := m.files[key]
	return exists
}

// GetFilePath implements CacheManager interface
func (m *MockCacheManager) GetFilePath(symbol, dataType, date string, timeframe ...string) string {
	if dataType == "klines" && len(timeframe) > 0 {
		return fmt.Sprintf("mock/path/%s-%s-%s.zip", symbol, timeframe[0], date)
	} else if dataType == "trades" {
		return fmt.Sprintf("mock/path/%s-trades-%s.zip", symbol, date)
	}
	return ""
}

// UpdateIndex implements CacheManager interface
func (m *MockCacheManager) UpdateIndex(fileInfo shared.FileMetadata) error {
	if fileInfo.Symbol == "" || fileInfo.DataType == "" || fileInfo.Date == "" {
		return fmt.Errorf("invalid file metadata")
	}

	var key string
	if fileInfo.Timeframe != "" {
		key = m.generateKey(fileInfo.Symbol, fileInfo.DataType, fileInfo.Date, fileInfo.Timeframe)
	} else {
		key = m.generateKey(fileInfo.Symbol, fileInfo.DataType, fileInfo.Date)
	}

	m.files[key] = &fileInfo
	m.updateStats()
	return nil
}

// IsFileCorrupted implements CacheManager interface
func (m *MockCacheManager) IsFileCorrupted(filePath string) (bool, error) {
	for _, corruptedFile := range m.corrupted {
		if corruptedFile == filePath {
			return true, nil
		}
	}
	return false, nil
}

// GetCacheStats implements CacheManager interface
func (m *MockCacheManager) GetCacheStats() *shared.CacheStatistics {
	return &shared.CacheStatistics{
		TotalFiles:     m.stats.TotalFiles,
		TotalSizeMB:    m.stats.TotalSizeMB,
		HitRate:        m.stats.HitRate,
		CorruptedFiles: m.stats.CorruptedFiles,
		LastUpdateTime: m.stats.LastUpdateTime,
	}
}

// CleanupCorrupted implements CacheManager interface
func (m *MockCacheManager) CleanupCorrupted() error {
	var toDelete []string
	for key, metadata := range m.files {
		for _, corruptedFile := range m.corrupted {
			if metadata.FilePath == corruptedFile {
				toDelete = append(toDelete, key)
				break
			}
		}
	}

	for _, key := range toDelete {
		delete(m.files, key)
	}

	m.corrupted = []string{}
	m.updateStats()
	return nil
}

// Mock-specific methods for testing

// AddCorruptedFile marks a file as corrupted for testing
func (m *MockCacheManager) AddCorruptedFile(filePath string) {
	m.corrupted = append(m.corrupted, filePath)
	m.updateStats()
}

// SetHitRate sets the cache hit rate for testing
func (m *MockCacheManager) SetHitRate(rate float64) {
	m.stats.HitRate = rate
}

// Helper methods
func (m *MockCacheManager) generateKey(symbol, dataType, date string, timeframe ...string) string {
	if len(timeframe) > 0 {
		return fmt.Sprintf("%s_%s_%s_%s", symbol, dataType, date, timeframe[0])
	}
	return fmt.Sprintf("%s_%s_%s", symbol, dataType, date)
}

func (m *MockCacheManager) updateStats() {
	m.stats.TotalFiles = len(m.files)
	m.stats.CorruptedFiles = len(m.corrupted)

	var totalSize int64
	for _, metadata := range m.files {
		totalSize += metadata.FileSize
	}
	m.stats.TotalSizeMB = float64(totalSize) / (1024 * 1024)
	m.stats.LastUpdateTime = time.Now()
}

// MockDownloader implements the Downloader interface for testing
type MockDownloader struct {
	downloads map[string]*shared.DownloadResult
	failures  map[string]error
}

// NewMockDownloader creates a new mock downloader
func NewMockDownloader() *MockDownloader {
	return &MockDownloader{
		downloads: make(map[string]*shared.DownloadResult),
		failures:  make(map[string]error),
	}
}

// DownloadFile implements Downloader interface
func (m *MockDownloader) DownloadFile(request shared.DownloadRequest) (*shared.DownloadResult, error) {
	key := fmt.Sprintf("%s_%s_%s_%s", request.Symbol, request.DataType, request.Date, request.Timeframe)
	
	if err, exists := m.failures[key]; exists {
		return &shared.DownloadResult{
			Request: request,
			Success: false,
			Error:   err.Error(),
		}, err
	}

	if result, exists := m.downloads[key]; exists {
		return result, nil
	}

	// Default success result
	return &shared.DownloadResult{
		Request:   request,
		Success:   true,
		FilePath:  fmt.Sprintf("mock/downloads/%s", key),
		FileSize:  1000,
		Duration:  time.Millisecond * 100,
	}, nil
}

// CheckFileExists implements Downloader interface
func (m *MockDownloader) CheckFileExists(request shared.DownloadRequest) (bool, string, error) {
	key := fmt.Sprintf("%s_%s_%s_%s", request.Symbol, request.DataType, request.Date, request.Timeframe)
	
	if result, exists := m.downloads[key]; exists {
		return true, result.FilePath, nil
	}
	
	return false, "", nil
}

// GetDownloadURL implements Downloader interface
func (m *MockDownloader) GetDownloadURL(request shared.DownloadRequest) (string, error) {
	return fmt.Sprintf("https://mock.binance.vision/%s/%s/%s", request.DataType, request.Symbol, request.Date), nil
}

// ValidateChecksumFile implements Downloader interface
func (m *MockDownloader) ValidateChecksumFile(filePath string) (bool, error) {
	return false, nil // Mock: not corrupted
}

// GetDownloadStats implements Downloader interface
func (m *MockDownloader) GetDownloadStats() *shared.CacheStatistics {
	return &shared.CacheStatistics{
		TotalFiles:     len(m.downloads),
		TotalSizeMB:    float64(len(m.downloads) * 1000) / (1024 * 1024),
		LastUpdateTime: time.Now(),
	}
}

// CleanupCorruptedFiles implements Downloader interface
func (m *MockDownloader) CleanupCorruptedFiles() error {
	return nil
}

// Mock-specific methods for testing

// SetDownloadResult sets a predefined result for a request
func (m *MockDownloader) SetDownloadResult(request shared.DownloadRequest, result *shared.DownloadResult) {
	key := fmt.Sprintf("%s_%s_%s_%s", request.Symbol, request.DataType, request.Date, request.Timeframe)
	m.downloads[key] = result
}

// SetDownloadFailure sets a failure for a request
func (m *MockDownloader) SetDownloadFailure(request shared.DownloadRequest, err error) {
	key := fmt.Sprintf("%s_%s_%s_%s", request.Symbol, request.DataType, request.Date, request.Timeframe)
	m.failures[key] = err
}

// MockStreamingReader implements the StreamingReader interface for testing
type MockStreamingReader struct {
	klinesData []shared.KlineData
	tradesData []shared.TradeData
	metrics    *shared.MemoryMetrics
}

// NewMockStreamingReader creates a new mock streaming reader
func NewMockStreamingReader() *MockStreamingReader {
	return &MockStreamingReader{
		klinesData: []shared.KlineData{},
		tradesData: []shared.TradeData{},
		metrics: &shared.MemoryMetrics{
			CurrentUsageMB: 10.5,
			PeakUsageMB:    15.2,
			BuffersActive:  2,
			LastUpdateTime: time.Now(),
		},
	}
}

// StreamKlines implements StreamingReader interface
func (m *MockStreamingReader) StreamKlines(filePath string, callback func(shared.KlineData) error) error {
	if filePath == "" {
		return fmt.Errorf("empty file path")
	}

	for _, kline := range m.klinesData {
		if err := callback(kline); err != nil {
			return err
		}
	}
	return nil
}

// StreamTrades implements StreamingReader interface
func (m *MockStreamingReader) StreamTrades(filePath string, callback func(shared.TradeData) error) error {
	if filePath == "" {
		return fmt.Errorf("empty file path")
	}

	for _, trade := range m.tradesData {
		if err := callback(trade); err != nil {
			return err
		}
	}
	return nil
}

// GetMemoryMetrics implements StreamingReader interface
func (m *MockStreamingReader) GetMemoryMetrics() *shared.MemoryMetrics {
	return &shared.MemoryMetrics{
		CurrentUsageMB: m.metrics.CurrentUsageMB,
		PeakUsageMB:    m.metrics.PeakUsageMB,
		BuffersActive:  m.metrics.BuffersActive,
		LastUpdateTime: m.metrics.LastUpdateTime,
	}
}

// ValidateMemoryConstraints implements StreamingReader interface
func (m *MockStreamingReader) ValidateMemoryConstraints() (bool, error) {
	return m.metrics.CurrentUsageMB < 100.0, nil // Mock constraint
}

// Mock-specific methods for testing

// SetKlinesData sets the klines data to be returned by StreamKlines
func (m *MockStreamingReader) SetKlinesData(klines []shared.KlineData) {
	m.klinesData = klines
}

// SetTradesData sets the trades data to be returned by StreamTrades
func (m *MockStreamingReader) SetTradesData(trades []shared.TradeData) {
	m.tradesData = trades
}

// SetMemoryUsage sets the memory usage metrics
func (m *MockStreamingReader) SetMemoryUsage(currentMB, peakMB float64, buffers int) {
	m.metrics.CurrentUsageMB = currentMB
	m.metrics.PeakUsageMB = peakMB
	m.metrics.BuffersActive = buffers
	m.metrics.LastUpdateTime = time.Now()
}
