// Package testutils provides test utilities and fixtures
package testutils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/shared"
)

// CreateMockZipFile creates a mock ZIP file for testing
func CreateMockZipFile(zipPath, csvFileName, csvContent string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	csvWriter, err := zipWriter.Create(csvFileName)
	if err != nil {
		return err
	}

	_, err = csvWriter.Write([]byte(csvContent))
	return err
}

// MockKlineData returns sample klines data for testing
func MockKlineData() []shared.KlineData {
	baseTime := int64(1623024000000) // 2021-06-07 00:00:00 UTC
	return []shared.KlineData{
		{
			OpenTime:                 baseTime,
			Open:                     100.0,
			High:                     105.0,
			Low:                      99.0,
			Close:                    102.0,
			Volume:                   1000.0,
			CloseTime:                baseTime + 5*60*1000 - 1,
			QuoteAssetVolume:        102000.0,
			NumberOfTrades:          50,
			TakerBuyBaseAssetVolume: 500.0,
			TakerBuyQuoteAssetVolume: 51000.0,
			Ignore:                   "0",
		},
		{
			OpenTime:                 baseTime + 5*60*1000,
			Open:                     102.0,
			High:                     107.0,
			Low:                      101.0,
			Close:                    104.0,
			Volume:                   1200.0,
			CloseTime:                baseTime + 10*60*1000 - 1,
			QuoteAssetVolume:        124800.0,
			NumberOfTrades:          60,
			TakerBuyBaseAssetVolume: 600.0,
			TakerBuyQuoteAssetVolume: 62400.0,
			Ignore:                   "0",
		},
		{
			OpenTime:                 baseTime + 10*60*1000,
			Open:                     104.0,
			High:                     106.0,
			Low:                      103.0,
			Close:                    105.0,
			Volume:                   800.0,
			CloseTime:                baseTime + 15*60*1000 - 1,
			QuoteAssetVolume:        84000.0,
			NumberOfTrades:          40,
			TakerBuyBaseAssetVolume: 400.0,
			TakerBuyQuoteAssetVolume: 42000.0,
			Ignore:                   "0",
		},
	}
}

// MockTradeData returns sample trades data for testing
func MockTradeData() []shared.TradeData {
	baseTime := int64(1623024000000) // 2021-06-07 00:00:00 UTC
	return []shared.TradeData{
		{
			ID:           1001,
			Price:        100.0,
			Quantity:     10.0,
			QuoteQty:     1000.0,
			Time:         baseTime,
			IsBuyerMaker: true,
		},
		{
			ID:           1002,
			Price:        105.0,
			Quantity:     5.0,
			QuoteQty:     525.0,
			Time:         baseTime + 60*1000,
			IsBuyerMaker: false,
		},
		{
			ID:           1003,
			Price:        103.0,
			Quantity:     8.0,
			QuoteQty:     824.0,
			Time:         baseTime + 120*1000,
			IsBuyerMaker: true,
		},
	}
}

// MockFileMetadata returns sample file metadata for testing
func MockFileMetadata(tempDir string) shared.FileMetadata {
	return shared.FileMetadata{
		Symbol:     "SOLUSDT",
		DataType:   "klines",
		Date:       "2023-06-01",
		Timeframe:  "5m",
		FilePath:   filepath.Join(tempDir, "test.zip"),
		FileSize:   1000,
		Checksum:   "abc123",
		Downloaded: time.Now(),
		Verified:   true,
	}
}

// MockDownloadRequest returns sample download request for testing
func MockDownloadRequest() shared.DownloadRequest {
	return shared.DownloadRequest{
		Symbol:    "SOLUSDT",
		DataType:  "klines",
		Date:      "2023-06-01",
		Timeframe: "5m",
	}
}

// MockParsedDataBatch returns sample parsed data batch for testing
func MockParsedDataBatch() *shared.ParsedDataBatch {
	return &shared.ParsedDataBatch{
		Symbol:      "SOLUSDT",
		DataType:    "klines",
		Timeframe:   "5m",
		Date:        "2023-06-01",
		RecordCount: 3,
		StartTime:   1623024000000,
		EndTime:     1623024899999,
		KlinesData:  MockKlineData(),
		ProcessedAt: time.Now(),
	}
}

// MockConfigs returns sample configurations for testing
func MockConfigs() (shared.CacheConfig, shared.DownloadConfig, shared.StreamingConfig, shared.AggregationConfig) {
	cache := shared.CacheConfig{
		RootPath:           "./test_cache",
		Symbols:            []string{"SOLUSDT", "ETHUSDT"},
		Timeframes:         []string{"5m", "15m", "1h"},
		ChecksumValidation: true,
	}

	download := shared.DownloadConfig{
		BaseURL:        "https://test.binance.vision",
		MaxRetries:     2,
		RetryDelay:     time.Second * 2,
		Timeout:        time.Minute * 5,
		MaxConcurrent:  3,
		ChecksumVerify: false,
	}

	streaming := shared.StreamingConfig{
		BufferSize:    32 * 1024, // 32KB for tests
		MaxMemoryMB:   50,        // 50MB for tests
		EnableMetrics: true,
	}

	aggregation := shared.AggregationConfig{
		SourceTimeframes: []string{"5m"},
		TargetTimeframes: []string{"15m", "1h"},
		ValidationRules: shared.ValidationConfig{
			MaxPriceDeviation:    10.0,
			MaxVolumeDeviation:   100.0,
			RequireMonotonicTime: false,
			MaxTimeGap:          60000, // 1 minute for tests
		},
	}

	return cache, download, streaming, aggregation
}

// CreateMockKlineCSV returns CSV content for klines testing
func CreateMockKlineCSV() string {
	return `1623024000000,100.0,105.0,99.0,102.0,1000.0,1623024299999,102000.0,50,500.0,51000.0,0
1623024300000,102.0,107.0,101.0,104.0,1200.0,1623024599999,124800.0,60,600.0,62400.0,0
1623024600000,104.0,106.0,103.0,105.0,800.0,1623024899999,84000.0,40,400.0,42000.0,0`
}

// CreateMockTradeCSV returns CSV content for trades testing
func CreateMockTradeCSV() string {
	return `1001,100.0,10.0,1000.0,1623024000000,true
1002,105.0,5.0,525.0,1623024060000,false
1003,103.0,8.0,824.0,1623024120000,true`
}

// CleanupTestFiles removes test files and directories
func CleanupTestFiles(paths []string) {
	for _, path := range paths {
		os.RemoveAll(path)
	}
}
