// Package cli - Workflow execution implementation
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/shared"
)

// Components structure to hold initialized components
type Components struct {
	Cache      *binance.CacheManager
	Downloader *binance.Downloader
	Streaming  *binance.StreamingReader
	Parser     *binance.ParsedDataProcessor
}

// initializeComponents initializes all required components
func (app *CLIApp) initializeComponents() (*Components, error) {
	// Initialize cache
	cache, err := binance.InitializeCache(app.config.BinanceData.CacheRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize downloader
	downloadConfig, err := app.config.BinanceData.Downloader.ToDownloadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert download config: %w", err)
	}

	downloader, err := binance.NewDownloader(cache, downloadConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}

	// Initialize streaming reader
	streamingConfig := app.config.BinanceData.Streaming.ToStreamingConfig()
	streaming, err := binance.NewStreamingReader(cache, streamingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming reader: %w", err)
	}

	// Initialize parser
	aggregationConfig := app.config.ToAggregationConfig()
	parser, err := binance.NewParsedDataProcessor(cache, streaming, aggregationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	return &Components{
		Cache:      cache,
		Downloader: downloader,
		Streaming:  streaming,
		Parser:     parser,
	}, nil
}

// executeDownloadStage executes only the download stage
func (app *CLIApp) executeDownloadStage(components *Components, result *WorkflowResult) error {
	// Get symbols and timeframes from args or config
	symbols := app.getEffectiveSymbols()
	timeframes := app.getEffectiveTimeframes()

	// Parse date range
	startDate, err := time.Parse("2006-01-02", app.config.DataPeriod.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}
	
	endDate, err := time.Parse("2006-01-02", app.config.DataPeriod.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	// Get data types from config
	dataTypes := app.getEffectiveDataTypes()

	// Download files for each symbol, data type, timeframe, and date
	for _, symbol := range symbols {
		for _, dataType := range dataTypes {
			if dataType == "klines" {
				// Klines: iterate through timeframes
				for _, timeframe := range timeframes {
					// If force redownload, clean existing files for this symbol/timeframe
					if app.args.ForceRedownload {
						err := app.cleanCacheForSymbolTimeframe(components.Cache, symbol, timeframe, startDate, endDate)
						if err != nil {
							result.AddWarning(fmt.Sprintf("Failed to clean cache for %s %s: %v", symbol, timeframe, err))
						}
					}
					
					// Iterate through all dates in the range
					for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
						dateStr := currentDate.Format("2006-01-02")
						
						request := shared.DownloadRequest{
							Symbol:    symbol,
							DataType:  dataType,
							Date:      dateStr,
							Timeframe: timeframe,
						}

						downloadResult, err := components.Downloader.DownloadFile(request)
						if err != nil {
							result.AddError(fmt.Errorf("klines download failed for %s %s %s: %w", symbol, timeframe, dateStr, err))
							continue
						}

						if downloadResult.Success {
							result.FilesDownloaded++
						}
					}
					result.TimeframesDownloaded++
				}
			} else if dataType == "trades" {
				// Trades: no timeframes, just symbol and date
				if app.args.ForceRedownload {
					err := app.cleanCacheForSymbolTrades(components.Cache, symbol, startDate, endDate)
					if err != nil {
						result.AddWarning(fmt.Sprintf("Failed to clean trades cache for %s: %v", symbol, err))
					}
				}
				
				// Iterate through all dates in the range
				for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
					dateStr := currentDate.Format("2006-01-02")
					
					request := shared.DownloadRequest{
						Symbol:    symbol,
						DataType:  dataType,
						Date:      dateStr,
						Timeframe: "", // No timeframe for trades
					}

					downloadResult, err := components.Downloader.DownloadFile(request)
					if err != nil {
						result.AddError(fmt.Errorf("trades download failed for %s %s: %w", symbol, dateStr, err))
						continue
					}

					if downloadResult.Success {
						result.FilesDownloaded++
					}
				}
			}
		}
		
		result.SymbolsProcessed = append(result.SymbolsProcessed, symbol)
	}

	return nil
}

// executeStreamingMode executes streaming mode workflow
func (app *CLIApp) executeStreamingMode(components *Components, result *WorkflowResult) error {
	// Validate memory constraints first
	valid, err := components.Streaming.ValidateMemoryConstraints()
	if err != nil {
		return fmt.Errorf("memory constraint validation failed: %w", err)
	}

	if !valid {
		result.AddWarning("Memory constraints validation failed, proceeding with caution")
	}

	// Execute download stage with streaming
	return app.executeDownloadStage(components, result)
}

// executeFullWorkflow executes the complete workflow
func (app *CLIApp) executeFullWorkflow(components *Components, result *WorkflowResult) error {
	// Stage 1: Download
	if err := app.executeDownloadStage(components, result); err != nil {
		return fmt.Errorf("download stage failed: %w", err)
	}

	// Stage 2: Processing and statistics
	for _, symbol := range result.SymbolsProcessed {
		stats := map[string]interface{}{
			"symbol":           symbol,
			"files_processed":  result.FilesDownloaded,
			"timeframes":       len(app.config.BinanceData.Timeframes),
		}
		
		result.Statistics[symbol] = stats
		result.FilesProcessed = result.FilesDownloaded
	}

	return nil
}

// ExportData exports data in the specified format
func (app *CLIApp) ExportData(result *WorkflowResult, format ExportFormat, outputPath string) error {
	if result == nil {
		return fmt.Errorf("workflow result is nil")
	}

	switch format {
	case FormatCSV:
		return app.exportCSV(result, outputPath)
	case FormatJSON:
		return app.exportJSON(result, outputPath)
	case FormatParquet:
		return app.exportParquet(result, outputPath)
	default:
		return fmt.Errorf("unsupported export format: %v", format)
	}
}

// exportCSV exports workflow results to CSV format
func (app *CLIApp) exportCSV(result *WorkflowResult, outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write CSV header and data (simplified)
	file.WriteString("Symbol,FilesDownloaded,TimeframesDownloaded,Success\n")
	for _, symbol := range result.SymbolsProcessed {
		file.WriteString(fmt.Sprintf("%s,%d,%d,%t\n", 
			symbol, result.FilesDownloaded, result.TimeframesDownloaded, result.Success))
	}

	return nil
}

// exportJSON exports workflow results to JSON format
func (app *CLIApp) exportJSON(result *WorkflowResult, outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file and write JSON (simplified)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Write simplified JSON
	file.WriteString(fmt.Sprintf(`{
	"success": %t,
	"filesDownloaded": %d,
	"timeframesDownloaded": %d,
	"symbolsProcessed": %d,
	"executionTime": "%v"
}`, result.Success, result.FilesDownloaded, result.TimeframesDownloaded, 
		len(result.SymbolsProcessed), result.ExecutionTime))

	return nil
}

// exportParquet exports workflow results to Parquet format
func (app *CLIApp) exportParquet(result *WorkflowResult, outputPath string) error {
	// Placeholder implementation - in real implementation would use parquet library
	return fmt.Errorf("parquet export not implemented in MVP")
}

// getEffectiveSymbols returns symbols from command line args or config
func (app *CLIApp) getEffectiveSymbols() []string {
	if len(app.args.Symbols) > 0 {
		return app.args.Symbols
	}
	
	if app.config != nil {
		return app.config.BinanceData.Symbols
	}
	
	return []string{}
}

// getEffectiveTimeframes returns timeframes from command line args or config
func (app *CLIApp) getEffectiveTimeframes() []string {
	if len(app.args.Timeframes) > 0 {
		return app.args.Timeframes
	}
	
	if app.config != nil {
		return app.config.BinanceData.Timeframes
	}
	
	return []string{}
}

// getEffectiveDataTypes returns data types from config
func (app *CLIApp) getEffectiveDataTypes() []string {
	if app.config != nil && len(app.config.BinanceData.DataTypes) > 0 {
		return app.config.BinanceData.DataTypes
	}
	
	// Default to klines only for backward compatibility
	return []string{"klines"}
}

// cleanCacheForSymbolTimeframe removes cached files for a specific symbol/timeframe/date range
func (app *CLIApp) cleanCacheForSymbolTimeframe(cache *binance.CacheManager, symbol, timeframe string, startDate, endDate time.Time) error {
	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		dateStr := currentDate.Format("2006-01-02")
		
		// Get file path and remove if exists
		filePath := cache.GetFilePath(symbol, "klines", dateStr, timeframe)
		if _, err := os.Stat(filePath); err == nil {
			// File exists, remove it
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove cached file %s: %w", filePath, err)
			}
		}
	}
	
	return nil
}

// cleanCacheForSymbolTrades removes cached trades files for a specific symbol/date range
func (app *CLIApp) cleanCacheForSymbolTrades(cache *binance.CacheManager, symbol string, startDate, endDate time.Time) error {
	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		dateStr := currentDate.Format("2006-01-02")
		
		// Get file path and remove if exists (trades don't have timeframes)
		filePath := cache.GetFilePath(symbol, "trades", dateStr)
		if _, err := os.Stat(filePath); err == nil {
			// File exists, remove it
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove cached trades file %s: %w", filePath, err)
			}
		}
	}
	
	return nil
}
