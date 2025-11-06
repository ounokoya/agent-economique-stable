// Package shared provides common interfaces for the trading agent
package shared

import "time"

// CacheManager interface defines cache operations
type CacheManager interface {
	FileExists(symbol, dataType, date string, timeframe ...string) bool
	GetFilePath(symbol, dataType, date string, timeframe ...string) string
	UpdateIndex(fileInfo FileMetadata) error
	IsFileCorrupted(filePath string) (bool, error)
	GetCacheStats() *CacheStatistics
	CleanupCorrupted() error
}

// Downloader interface defines download operations
type Downloader interface {
	DownloadFile(request DownloadRequest) (*DownloadResult, error)
	CheckFileExists(request DownloadRequest) (bool, string, error)
	GetDownloadURL(request DownloadRequest) (string, error)
	ValidateChecksumFile(filePath string) (bool, error)
	GetDownloadStats() *CacheStatistics
	CleanupCorruptedFiles() error
}

// StreamingReader interface defines streaming operations
type StreamingReader interface {
	StreamKlines(filePath string, callback func(KlineData) error) error
	StreamTrades(filePath string, callback func(TradeData) error) error
	GetMemoryMetrics() *MemoryMetrics
	ValidateMemoryConstraints() (bool, error)
}

// DataProcessor interface defines data processing operations
type DataProcessor interface {
	ParseKlinesBatch(filePath, symbol, timeframe, date string) (*ParsedDataBatch, error)
	ParseTradesBatch(filePath, symbol, date string) (*ParsedDataBatch, error)
	ValidateDataBatch(batch *ParsedDataBatch) (*DataValidationResult, error)
	GetBatchStatistics(batch *ParsedDataBatch) (map[string]interface{}, error)
}

// TimeframeAggregator interface defines aggregation operations
type TimeframeAggregator interface {
	AggregateKlinesToTimeframe(sourceKlines []KlineData, targetTimeframe string) ([]KlineData, error)
	AggregateTradestoKlines(trades []TradeData, timeframe string) ([]KlineData, error)
	ValidateTimeframeContinuity(klines []KlineData, timeframe string) error
	SynchronizeMultiTimeframes(klinesMap map[string][]KlineData) (map[string][]KlineData, error)
}

// Config interfaces

// CacheConfig holds configuration for cache initialization
type CacheConfig struct {
	RootPath           string   `yaml:"root_path"`
	Symbols            []string `yaml:"symbols"`
	Timeframes         []string `yaml:"timeframes"`
	ChecksumValidation bool     `yaml:"checksum_validation"`
}

// DownloadConfig holds configuration for downloader
type DownloadConfig struct {
	BaseURL         string        `yaml:"base_url"`
	MaxRetries      int           `yaml:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxConcurrent   int           `yaml:"max_concurrent"`
	ChecksumVerify  bool          `yaml:"checksum_verify"`
}

// StreamingConfig holds configuration for streaming ZIP reader
type StreamingConfig struct {
	BufferSize     int  `yaml:"buffer_size"`     // Buffer size for reading
	MaxMemoryMB    int  `yaml:"max_memory_mb"`   // Maximum memory usage in MB
	EnableMetrics  bool `yaml:"enable_metrics"`  // Enable memory monitoring
}

// AggregationConfig holds configuration for multi-timeframe aggregation
type AggregationConfig struct {
	SourceTimeframes []string `yaml:"source_timeframes"` // Source timeframes to aggregate from
	TargetTimeframes []string `yaml:"target_timeframes"` // Target timeframes to create
	ValidationRules  ValidationConfig `yaml:"validation_rules"`
}

// ValidationConfig holds data validation rules
type ValidationConfig struct {
	MaxPriceDeviation    float64 `yaml:"max_price_deviation"`     // Max % price change between records
	MaxVolumeDeviation   float64 `yaml:"max_volume_deviation"`    // Max % volume change
	RequireMonotonicTime bool    `yaml:"require_monotonic_time"`  // Time must be strictly increasing
	MaxTimeGap          int64   `yaml:"max_time_gap"`            // Max gap in milliseconds
}
