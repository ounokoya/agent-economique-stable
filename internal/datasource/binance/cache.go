// Package binance provides cache management for Binance Vision data
package binance

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"agent-economique/internal/shared"
)

// CacheManager manages the local cache of Binance Vision data
type CacheManager struct {
	rootPath  string
	index     map[string]*shared.FileMetadata // Key: symbol_datatype_date_timeframe
	mutex     sync.RWMutex
	stats     *shared.CacheStatistics
}

// InitializeCache creates a new CacheManager instance and loads existing index
func InitializeCache(rootPath string) (*CacheManager, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("root path cannot be empty")
	}

	// Create root directory if not exists
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	cache := &CacheManager{
		rootPath: rootPath,
		index:    make(map[string]*shared.FileMetadata),
		stats:    &shared.CacheStatistics{},
	}

	// Load existing index
	if err := cache.loadIndex(); err != nil {
		// If index doesn't exist, create empty one
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load index: %w", err)
		}
	}

	// Update statistics
	cache.updateStatistics()

	return cache, nil
}

// FileExists checks if a file exists in cache for given parameters
func (c *CacheManager) FileExists(symbol, dataType, date string, timeframe ...string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := c.generateKey(symbol, dataType, date, timeframe...)
	metadata, exists := c.index[key]
	
	if !exists {
		return false
	}

	// Verify file actually exists on disk
	if _, err := os.Stat(metadata.FilePath); os.IsNotExist(err) {
		// File was deleted, remove from index
		go func() {
			c.mutex.Lock()
			delete(c.index, key)
			c.mutex.Unlock()
			c.saveIndex() // Best effort save
		}()
		return false
	}

	return true
}

// GetFilePath returns the expected file path for given parameters
func (c *CacheManager) GetFilePath(symbol, dataType, date string, timeframe ...string) string {
	basePath := filepath.Join(c.rootPath, "binance", "futures_um")
	
	if dataType == "klines" && len(timeframe) > 0 {
		// Structure: data/binance/futures_um/klines/SYMBOL/TIMEFRAME/SYMBOL-TIMEFRAME-DATE.zip
		fileName := fmt.Sprintf("%s-%s-%s.zip", symbol, timeframe[0], date)
		return filepath.Join(basePath, "klines", symbol, timeframe[0], fileName)
	} else if dataType == "trades" {
		// Structure: data/binance/futures_um/trades/SYMBOL/SYMBOL-trades-DATE.zip
		fileName := fmt.Sprintf("%s-trades-%s.zip", symbol, date)
		return filepath.Join(basePath, "trades", symbol, fileName)
	}
	
	return ""
}

// UpdateIndex updates the cache index with new file metadata
func (c *CacheManager) UpdateIndex(fileInfo shared.FileMetadata) error {
	if fileInfo.Symbol == "" || fileInfo.DataType == "" || fileInfo.Date == "" {
		return fmt.Errorf("invalid file metadata: symbol, dataType, and date are required")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var key string
	if fileInfo.Timeframe != "" {
		key = c.generateKey(fileInfo.Symbol, fileInfo.DataType, fileInfo.Date, fileInfo.Timeframe)
	} else {
		key = c.generateKey(fileInfo.Symbol, fileInfo.DataType, fileInfo.Date)
	}
	c.index[key] = &fileInfo
	
	// Update statistics
	c.updateStatisticsUnsafe()
	
	// Save index to disk
	return c.saveIndexUnsafe()
}

// IsFileCorrupted checks if a cached file is corrupted using checksum
func (c *CacheManager) IsFileCorrupted(filePath string) (bool, error) {
	if filePath == "" {
		return false, fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// Get metadata from index
	c.mutex.RLock()
	var metadata *shared.FileMetadata
	for _, meta := range c.index {
		if meta.FilePath == filePath {
			metadata = meta
			break
		}
	}
	c.mutex.RUnlock()

	if metadata == nil {
		return false, fmt.Errorf("file not found in cache index: %s", filePath)
	}

	// Check file size
	if fileInfo.Size() != metadata.FileSize {
		return true, nil
	}

	// Verify checksum if available
	if metadata.Checksum != "" {
		actualChecksum, err := c.calculateChecksum(filePath)
		if err != nil {
			return false, fmt.Errorf("failed to calculate checksum: %w", err)
		}
		
		if actualChecksum != metadata.Checksum {
			return true, nil
		}
	}

	return false, nil
}

// GetCacheStats returns current cache statistics
func (c *CacheManager) GetCacheStats() *shared.CacheStatistics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Return a copy to prevent concurrent access issues
	return &shared.CacheStatistics{
		TotalFiles:     c.stats.TotalFiles,
		TotalSizeMB:    c.stats.TotalSizeMB,
		HitRate:        c.stats.HitRate,
		CorruptedFiles: c.stats.CorruptedFiles,
		LastUpdateTime: c.stats.LastUpdateTime,
	}
}

// CleanupCorrupted removes corrupted files from cache and index
func (c *CacheManager) CleanupCorrupted() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var corruptedKeys []string
	
	for key, metadata := range c.index {
		// Check if file exists on disk first
		fileInfo, err := os.Stat(metadata.FilePath)
		if os.IsNotExist(err) {
			// File doesn't exist, mark for removal
			corruptedKeys = append(corruptedKeys, key)
			continue
		}
		if err != nil {
			// Other error, skip this file
			continue
		}

		// Check file corruption without acquiring locks (we already have the write lock)
		corrupted := false
		
		// Check file size
		if fileInfo.Size() != metadata.FileSize {
			corrupted = true
		}

		// Check checksum if available
		if !corrupted && metadata.Checksum != "" {
			actualChecksum, err := c.calculateChecksum(metadata.FilePath)
			if err == nil && actualChecksum != metadata.Checksum {
				corrupted = true
			}
		}
		
		if corrupted {
			// Remove corrupted file
			if err := os.Remove(metadata.FilePath); err != nil {
				return fmt.Errorf("failed to remove corrupted file %s: %w", metadata.FilePath, err)
			}
			corruptedKeys = append(corruptedKeys, key)
		}
	}

	// Remove from index
	for _, key := range corruptedKeys {
		delete(c.index, key)
	}

	// Update statistics and save
	c.updateStatisticsUnsafe()
	return c.saveIndexUnsafe()
}

// Helper functions

// generateKey creates a unique key for cache index
func (c *CacheManager) generateKey(symbol, dataType, date string, timeframe ...string) string {
	if len(timeframe) > 0 {
		return fmt.Sprintf("%s_%s_%s_%s", symbol, dataType, date, timeframe[0])
	}
	return fmt.Sprintf("%s_%s_%s", symbol, dataType, date)
}

// calculateChecksum computes SHA256 checksum of a file
func (c *CacheManager) calculateChecksum(filePath string) (string, error) {
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

// loadIndex loads the cache index from disk
func (c *CacheManager) loadIndex() error {
	indexPath := filepath.Join(c.rootPath, "index.json")
	
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index map[string]*shared.FileMetadata
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	c.index = index
	return nil
}

// saveIndex saves the cache index to disk
func (c *CacheManager) saveIndex() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.saveIndexUnsafe()
}

// saveIndexUnsafe saves index without locking (caller must hold lock)
func (c *CacheManager) saveIndexUnsafe() error {
	indexPath := filepath.Join(c.rootPath, "index.json")
	
	data, err := json.MarshalIndent(c.index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

// updateStatistics recalculates cache statistics
func (c *CacheManager) updateStatistics() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.updateStatisticsUnsafe()
}

// updateStatisticsUnsafe updates stats without locking (caller must hold lock)
func (c *CacheManager) updateStatisticsUnsafe() {
	totalFiles := len(c.index)
	var totalSize int64
	corruptedFiles := 0

	for _, metadata := range c.index {
		totalSize += metadata.FileSize
		
		if !metadata.Verified {
			// Check corruption without acquiring additional locks
			fileInfo, err := os.Stat(metadata.FilePath)
			if err != nil {
				corruptedFiles++
				continue
			}

			// Quick corruption check - size mismatch
			if fileInfo.Size() != metadata.FileSize {
				corruptedFiles++
			}
		}
	}

	c.stats = &shared.CacheStatistics{
		TotalFiles:     totalFiles,
		TotalSizeMB:    float64(totalSize) / (1024 * 1024),
		HitRate:        0.0, // Will be calculated during actual usage
		CorruptedFiles: corruptedFiles,
		LastUpdateTime: time.Now(),
	}
}
