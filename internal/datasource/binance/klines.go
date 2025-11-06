// Package binance provides streaming functionality for Binance klines data
package binance

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"time"

	"agent-economique/internal/shared"
)

// StreamingReader provides streaming access to Binance ZIP files
type StreamingReader struct {
	cache   *CacheManager
	config  shared.StreamingConfig
	metrics *shared.MemoryMetrics
}

// NewStreamingReader creates a new StreamingReader instance
func NewStreamingReader(cache *CacheManager, config shared.StreamingConfig) (*StreamingReader, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}

	// Set default configuration
	if config.BufferSize == 0 {
		config.BufferSize = 64 * 1024 // 64KB buffer
	}
	if config.MaxMemoryMB == 0 {
		config.MaxMemoryMB = 100 // 100MB limit
	}

	metrics := &shared.MemoryMetrics{
		CurrentUsageMB: 0,
		PeakUsageMB:    0,
		BuffersActive:  0,
	}

	return &StreamingReader{
		cache:   cache,
		config:  config,
		metrics: metrics,
	}, nil
}

// StreamKlines streams klines data from a ZIP file without loading everything in memory
func (sr *StreamingReader) StreamKlines(filePath string, callback func(shared.KlineData) error) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Open ZIP file
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer zipReader.Close()

	// Update memory metrics
	sr.updateMemoryMetrics()

	// Find CSV file inside ZIP
	var csvFile *zip.File
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, ".csv") {
			csvFile = file
			break
		}
	}

	if csvFile == nil {
		return fmt.Errorf("no CSV file found in ZIP archive")
	}

	// Open CSV file from ZIP
	csvReader, err := csvFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open CSV from ZIP: %w", err)
	}
	defer csvReader.Close()

	// Create buffered reader with limited buffer size
	bufferedReader := bufio.NewReaderSize(csvReader, sr.config.BufferSize)
	csvParser := csv.NewReader(bufferedReader)

	sr.metrics.BuffersActive++
	defer func() { sr.metrics.BuffersActive-- }()

	lineCount := 0
	for {
		// Check memory usage before processing each line
		if sr.config.EnableMetrics {
			sr.updateMemoryMetrics()
			if sr.metrics.CurrentUsageMB > float64(sr.config.MaxMemoryMB) {
				return fmt.Errorf("memory usage exceeded limit: %.2f MB > %d MB", 
					sr.metrics.CurrentUsageMB, sr.config.MaxMemoryMB)
			}
		}

		record, err := csvParser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV line %d: %w", lineCount, err)
		}

		lineCount++

		// Skip header line (first line)
		if lineCount == 1 {
			continue
		}

		// Parse kline data from CSV record
		kline, err := sr.parseKlineRecord(record)
		if err != nil {
			return fmt.Errorf("failed to parse kline at line %d: %w", lineCount, err)
		}

		// Call callback with parsed data
		if err := callback(*kline); err != nil {
			return fmt.Errorf("callback failed at line %d: %w", lineCount, err)
		}

		// Trigger GC periodically to keep memory usage low
		if lineCount%1000 == 0 {
			runtime.GC()
		}
	}

	return nil
}

// parseKlineRecord parses a CSV record into KlineData
func (sr *StreamingReader) parseKlineRecord(record []string) (*shared.KlineData, error) {
	if len(record) < 12 {
		return nil, fmt.Errorf("invalid kline record: expected 12 fields, got %d", len(record))
	}

	kline := &shared.KlineData{}
	var err error

	if kline.OpenTime, err = strconv.ParseInt(record[0], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid open_time: %w", err)
	}
	if kline.Open, err = strconv.ParseFloat(record[1], 64); err != nil {
		return nil, fmt.Errorf("invalid open: %w", err)
	}
	if kline.High, err = strconv.ParseFloat(record[2], 64); err != nil {
		return nil, fmt.Errorf("invalid high: %w", err)
	}
	if kline.Low, err = strconv.ParseFloat(record[3], 64); err != nil {
		return nil, fmt.Errorf("invalid low: %w", err)
	}
	if kline.Close, err = strconv.ParseFloat(record[4], 64); err != nil {
		return nil, fmt.Errorf("invalid close: %w", err)
	}
	if kline.Volume, err = strconv.ParseFloat(record[5], 64); err != nil {
		return nil, fmt.Errorf("invalid volume: %w", err)
	}
	if kline.CloseTime, err = strconv.ParseInt(record[6], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid close_time: %w", err)
	}
	if kline.QuoteAssetVolume, err = strconv.ParseFloat(record[7], 64); err != nil {
		return nil, fmt.Errorf("invalid quote_asset_volume: %w", err)
	}
	if kline.NumberOfTrades, err = strconv.ParseInt(record[8], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid number_of_trades: %w", err)
	}
	if kline.TakerBuyBaseAssetVolume, err = strconv.ParseFloat(record[9], 64); err != nil {
		return nil, fmt.Errorf("invalid taker_buy_base_asset_volume: %w", err)
	}
	if kline.TakerBuyQuoteAssetVolume, err = strconv.ParseFloat(record[10], 64); err != nil {
		return nil, fmt.Errorf("invalid taker_buy_quote_asset_volume: %w", err)
	}
	kline.Ignore = record[11]

	return kline, nil
}

// GetMemoryMetrics returns current memory usage metrics
func (sr *StreamingReader) GetMemoryMetrics() *shared.MemoryMetrics {
	sr.updateMemoryMetrics()
	
	// Return a copy to prevent concurrent access issues
	return &shared.MemoryMetrics{
		CurrentUsageMB: sr.metrics.CurrentUsageMB,
		PeakUsageMB:    sr.metrics.PeakUsageMB,
		BuffersActive:  sr.metrics.BuffersActive,
		LastUpdateTime: sr.metrics.LastUpdateTime,
	}
}

// ValidateMemoryConstraints checks if current memory usage is within limits
func (sr *StreamingReader) ValidateMemoryConstraints() (bool, error) {
	sr.updateMemoryMetrics()
	
	if sr.metrics.CurrentUsageMB > float64(sr.config.MaxMemoryMB) {
		return false, fmt.Errorf("memory constraint violation: %.2f MB > %d MB", 
			sr.metrics.CurrentUsageMB, sr.config.MaxMemoryMB)
	}
	
	return true, nil
}

// ResetMetrics resets memory tracking metrics
func (sr *StreamingReader) ResetMetrics() {
	sr.metrics.CurrentUsageMB = 0
	sr.metrics.PeakUsageMB = 0
	sr.metrics.BuffersActive = 0
	sr.updateMemoryMetrics()
}

// updateMemoryMetrics updates current memory usage statistics
func (sr *StreamingReader) updateMemoryMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	currentMB := float64(memStats.Alloc) / (1024 * 1024)
	
	sr.metrics.CurrentUsageMB = currentMB
	if currentMB > sr.metrics.PeakUsageMB {
		sr.metrics.PeakUsageMB = currentMB
	}
	sr.metrics.LastUpdateTime = time.Now()
}
