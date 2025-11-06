package bingx

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a single cache entry with TTL
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache provides thread-safe caching with TTL support
type Cache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	
	// Default TTL for cache entries
	defaultTTL time.Duration
	
	// Cleanup ticker for expired entries
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewCache creates a new cache with specified default TTL
func NewCache(defaultTTL time.Duration) *Cache {
	cache := &Cache{
		entries:    make(map[string]*CacheEntry),
		defaultTTL: defaultTTL,
		stopCleanup: make(chan bool, 1),
	}
	
	// Start cleanup goroutine
	cache.startCleanup()
	
	return cache
}

// Set stores a value in cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	
	if entry.IsExpired() {
		// Remove expired entry
		delete(c.entries, key)
		return nil, false
	}
	
	return entry.Value, true
}

// GetString retrieves a string value from cache
func (c *Cache) GetString(key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}
	
	strValue, ok := value.(string)
	return strValue, ok
}

// GetFloat64 retrieves a float64 value from cache
func (c *Cache) GetFloat64(key string) (float64, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0, false
	}
	
	floatValue, ok := value.(float64)
	return floatValue, ok
}

// GetTicker retrieves a Ticker from cache
func (c *Cache) GetTicker(key string) (*Ticker, bool) {
	value, exists := c.Get(key)
	if !exists {
		return nil, false
	}
	
	ticker, ok := value.(*Ticker)
	return ticker, ok
}

// GetKlines retrieves Klines from cache
func (c *Cache) GetKlines(key string) ([]Kline, bool) {
	value, exists := c.Get(key)
	if !exists {
		return nil, false
	}
	
	klines, ok := value.([]Kline)
	return klines, ok
}

// Delete removes a key from cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.entries, key)
}

// Has checks if a key exists and is not expired
func (c *Cache) Has(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.entries = make(map[string]*CacheEntry)
}

// Size returns the number of entries in cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return len(c.entries)
}

// Keys returns all non-expired keys in cache
func (c *Cache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	var keys []string
	now := time.Now()
	
	for key, entry := range c.entries {
		if !now.After(entry.ExpiresAt) {
			keys = append(keys, key)
		}
	}
	
	return keys
}

// GetTTL returns the remaining TTL for a key
func (c *Cache) GetTTL(key string) (time.Duration, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return 0, false
	}
	
	now := time.Now()
	if now.After(entry.ExpiresAt) {
		return 0, false
	}
	
	return entry.ExpiresAt.Sub(now), true
}

// Extend extends the TTL of an existing cache entry
func (c *Cache) Extend(key string, additionalTTL time.Duration) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return false
	}
	
	if entry.IsExpired() {
		delete(c.entries, key)
		return false
	}
	
	entry.ExpiresAt = entry.ExpiresAt.Add(additionalTTL)
	return true
}

// Refresh updates the expiration time of an entry to now + default TTL
func (c *Cache) Refresh(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return false
	}
	
	entry.ExpiresAt = time.Now().Add(c.defaultTTL)
	return true
}

// GetWithRefresh gets a value and refreshes its TTL if found
func (c *Cache) GetWithRefresh(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	
	if entry.IsExpired() {
		delete(c.entries, key)
		return nil, false
	}
	
	// Refresh TTL
	entry.ExpiresAt = time.Now().Add(c.defaultTTL)
	return entry.Value, true
}

// startCleanup starts the background cleanup goroutine
func (c *Cache) startCleanup() {
	// Run cleanup every minute
	c.cleanupTicker = time.NewTicker(1 * time.Minute)
	
	go func() {
		for {
			select {
			case <-c.cleanupTicker.C:
				c.cleanupExpired()
			case <-c.stopCleanup:
				c.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanupExpired removes all expired entries from cache
func (c *Cache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Stop stops the cache cleanup goroutine
func (c *Cache) Stop() {
	select {
	case c.stopCleanup <- true:
	default:
		// Channel is full or closed, cleanup already stopped
	}
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	totalEntries := len(c.entries)
	expiredEntries := 0
	now := time.Now()
	
	for _, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expiredEntries++
		}
	}
	
	return map[string]interface{}{
		"total_entries":   totalEntries,
		"active_entries":  totalEntries - expiredEntries,
		"expired_entries": expiredEntries,
		"default_ttl":     c.defaultTTL.String(),
	}
}

// CachedMarketDataService wraps MarketDataService with caching
type CachedMarketDataService struct {
	service *MarketDataService
	cache   *Cache
	
	// TTL configurations for different data types
	priceTTL  time.Duration
	klinesTTL time.Duration
	tickerTTL time.Duration
}

// NewCachedMarketDataService creates a cached market data service
func NewCachedMarketDataService(service *MarketDataService) *CachedMarketDataService {
	return &CachedMarketDataService{
		service:   service,
		cache:     NewCache(10 * time.Second), // Default 10 seconds TTL
		priceTTL:  5 * time.Second,            // Price cache for 5 seconds
		klinesTTL: 30 * time.Second,           // Klines cache for 30 seconds
		tickerTTL: 10 * time.Second,           // Ticker cache for 10 seconds
	}
}

// GetPriceWithCache retrieves price with caching
func (c *CachedMarketDataService) GetPriceWithCache(ctx context.Context, symbol string) (*Ticker, error) {
	cacheKey := "price:" + symbol
	
	if cachedTicker, exists := c.cache.GetTicker(cacheKey); exists {
		return cachedTicker, nil
	}
	
	ticker, err := c.service.GetPrice(ctx, symbol)
	if err != nil {
		return nil, err
	}
	
	c.cache.SetWithTTL(cacheKey, ticker, c.priceTTL)
	return ticker, nil
}

// GetKlinesWithCache retrieves klines with caching
func (c *CachedMarketDataService) GetKlinesWithCache(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	cacheKey := fmt.Sprintf("klines:%s:%s:%d", symbol, interval, limit)
	
	if cachedKlines, exists := c.cache.GetKlines(cacheKey); exists {
		return cachedKlines, nil
	}
	
	klines, err := c.service.GetKlines(ctx, symbol, interval, limit, nil, nil)
	if err != nil {
		return nil, err
	}
	
	c.cache.SetWithTTL(cacheKey, klines, c.klinesTTL)
	return klines, nil
}

// GetTicker24hrWithCache retrieves 24hr ticker with caching
func (c *CachedMarketDataService) GetTicker24hrWithCache(ctx context.Context, symbol string) (*Ticker, error) {
	cacheKey := "ticker24hr:" + symbol
	
	if cachedTicker, exists := c.cache.GetTicker(cacheKey); exists {
		return cachedTicker, nil
	}
	
	ticker, err := c.service.GetTicker24hr(ctx, symbol)
	if err != nil {
		return nil, err
	}
	
	c.cache.SetWithTTL(cacheKey, ticker, c.tickerTTL)
	return ticker, nil
}

// InvalidateSymbol removes all cached data for a symbol
func (c *CachedMarketDataService) InvalidateSymbol(symbol string) {
	keys := c.cache.Keys()
	for _, key := range keys {
		if contains(key, symbol) {
			c.cache.Delete(key)
		}
	}
}

// GetCacheStats returns cache statistics
func (c *CachedMarketDataService) GetCacheStats() map[string]interface{} {
	return c.cache.GetStats()
}

// Stop stops the cached service
func (c *CachedMarketDataService) Stop() {
	c.cache.Stop()
}

// Helper function for string contains check (simple implementation)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
