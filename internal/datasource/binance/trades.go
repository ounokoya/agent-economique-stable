// Package binance provides streaming functionality for Binance trades data
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

	"agent-economique/internal/shared"
)

// StreamTrades streams trades data from a ZIP file without loading everything in memory
func (sr *StreamingReader) StreamTrades(filePath string, callback func(shared.TradeData) error) error {
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

		// Parse trade data from CSV record
		trade, err := sr.parseTradeRecord(record)
		if err != nil {
			return fmt.Errorf("failed to parse trade at line %d: %w", lineCount, err)
		}

		// Call callback with parsed data
		if err := callback(*trade); err != nil {
			return fmt.Errorf("callback failed at line %d: %w", lineCount, err)
		}

		// Trigger GC periodically to keep memory usage low
		if lineCount%1000 == 0 {
			runtime.GC()
		}
	}

	return nil
}

// parseTradeRecord parses a CSV record into TradeData
func (sr *StreamingReader) parseTradeRecord(record []string) (*shared.TradeData, error) {
	if len(record) < 6 {
		return nil, fmt.Errorf("invalid trade record: expected 6 fields, got %d", len(record))
	}

	trade := &shared.TradeData{}
	var err error

	if trade.ID, err = strconv.ParseInt(record[0], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}
	if trade.Price, err = strconv.ParseFloat(record[1], 64); err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}
	if trade.Quantity, err = strconv.ParseFloat(record[2], 64); err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}
	if trade.QuoteQty, err = strconv.ParseFloat(record[3], 64); err != nil {
		return nil, fmt.Errorf("invalid quoteQty: %w", err)
	}
	if trade.Time, err = strconv.ParseInt(record[4], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid time: %w", err)
	}
	if trade.IsBuyerMaker, err = strconv.ParseBool(record[5]); err != nil {
		return nil, fmt.Errorf("invalid isBuyerMaker: %w", err)
	}

	return trade, nil
}
