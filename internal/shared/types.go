// Package shared provides common types and structures for the trading agent
package shared

import "time"

// FileMetadata contains metadata information for cached files
type FileMetadata struct {
	Symbol     string    `json:"symbol"`
	DataType   string    `json:"data_type"`   // "klines" or "trades"
	Date       string    `json:"date"`        // Format: YYYY-MM-DD
	Timeframe  string    `json:"timeframe"`   // "5m", "15m", "1h", "4h" (empty for trades)
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	Checksum   string    `json:"checksum"`    // SHA256
	Downloaded time.Time `json:"downloaded"`
	Verified   bool      `json:"verified"`
}

// CacheStatistics holds cache performance metrics
type CacheStatistics struct {
	TotalFiles      int     `json:"total_files"`
	TotalSizeMB     float64 `json:"total_size_mb"`
	HitRate         float64 `json:"hit_rate"`
	CorruptedFiles  int     `json:"corrupted_files"`
	LastUpdateTime  time.Time `json:"last_update_time"`
}

// DownloadRequest represents a request to download Binance data
type DownloadRequest struct {
	Symbol     string `json:"symbol"`
	DataType   string `json:"data_type"`   // "klines" or "trades"
	Date       string `json:"date"`        // Format: YYYY-MM-DD
	Timeframe  string `json:"timeframe"`   // "5m", "15m", "1h", "4h" (empty for trades)
}

// DownloadResult contains the result of a download operation
type DownloadResult struct {
	Request     DownloadRequest `json:"request"`
	Success     bool           `json:"success"`
	FilePath    string         `json:"file_path"`
	FileSize    int64          `json:"file_size"`
	Checksum    string         `json:"checksum"`
	Duration    time.Duration  `json:"duration"`
	Error       string         `json:"error,omitempty"`
}

// KlineData represents a single kline record
type KlineData struct {
	OpenTime                 int64   `json:"open_time"`
	Open                     float64 `json:"open,string"`
	High                     float64 `json:"high,string"`
	Low                      float64 `json:"low,string"`
	Close                    float64 `json:"close,string"`
	Volume                   float64 `json:"volume,string"`
	CloseTime                int64   `json:"close_time"`
	QuoteAssetVolume        float64 `json:"quote_asset_volume,string"`
	NumberOfTrades          int64   `json:"number_of_trades"`
	TakerBuyBaseAssetVolume float64 `json:"taker_buy_base_asset_volume,string"`
	TakerBuyQuoteAssetVolume float64 `json:"taker_buy_quote_asset_volume,string"`
	Ignore                   string  `json:"ignore"`
}

// TradeData represents a single trade record
type TradeData struct {
	ID           int64   `json:"id"`
	Price        float64 `json:"price,string"`
	Quantity     float64 `json:"qty,string"`
	QuoteQty     float64 `json:"quoteQty,string"`
	Time         int64   `json:"time"`
	IsBuyerMaker bool    `json:"isBuyerMaker"`
}

// MemoryMetrics tracks memory usage during streaming
type MemoryMetrics struct {
	CurrentUsageMB  float64   `json:"current_usage_mb"`
	PeakUsageMB     float64   `json:"peak_usage_mb"`
	BuffersActive   int       `json:"buffers_active"`
	LastUpdateTime  time.Time `json:"last_update_time"`
}

// ParsedDataBatch represents a batch of parsed data with metadata
type ParsedDataBatch struct {
	Symbol       string      `json:"symbol"`
	DataType     string      `json:"data_type"`     // "klines" or "trades"
	Timeframe    string      `json:"timeframe"`     // "5m", "15m", "1h", "4h" (empty for trades)
	Date         string      `json:"date"`          // YYYY-MM-DD
	RecordCount  int         `json:"record_count"`
	StartTime    int64       `json:"start_time"`    // Unix timestamp
	EndTime      int64       `json:"end_time"`      // Unix timestamp
	KlinesData   []KlineData `json:"klines_data,omitempty"`
	TradesData   []TradeData `json:"trades_data,omitempty"`
	ProcessedAt  time.Time   `json:"processed_at"`
}

// DataValidationResult contains validation results for a data batch
type DataValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	ErrorCount   int      `json:"error_count"`
	WarningCount int      `json:"warning_count"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
}
