package bingx

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	ttl := 10 * time.Second
	cache := NewCache(ttl)
	
	if cache == nil {
		t.Fatal("NewCache should not return nil")
	}
	
	if cache.defaultTTL != ttl {
		t.Errorf("Expected default TTL %v, got %v", ttl, cache.defaultTTL)
	}
	
	if cache.entries == nil {
		t.Error("Cache entries map should be initialized")
	}
	
	// Stop the cache to clean up goroutine
	cache.Stop()
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "test_key"
	value := "test_value"
	
	// Set value
	cache.Set(key, value)
	
	// Get value
	retrievedValue, exists := cache.Get(key)
	if !exists {
		t.Error("Value should exist in cache")
	}
	
	if retrievedValue != value {
		t.Errorf("Expected %s, got %s", value, retrievedValue)
	}
}

func TestCacheSetWithTTL(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "test_key"
	value := "test_value"
	shortTTL := 50 * time.Millisecond
	
	// Set value with short TTL
	cache.SetWithTTL(key, value, shortTTL)
	
	// Should exist immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Error("Value should exist immediately after setting")
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Should not exist after expiration
	_, exists = cache.Get(key)
	if exists {
		t.Error("Value should not exist after TTL expiration")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "expire_test"
	value := "will_expire"
	
	// Set with very short TTL
	cache.SetWithTTL(key, value, 1*time.Millisecond)
	
	// Wait for expiration
	time.Sleep(10 * time.Millisecond)
	
	// Should be expired
	_, exists := cache.Get(key)
	if exists {
		t.Error("Expired entry should not be retrievable")
	}
}

func TestCacheGetString(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "string_key"
	stringValue := "string_value"
	intValue := 123
	
	// Set string value
	cache.Set(key, stringValue)
	retrieved, exists := cache.GetString(key)
	if !exists || retrieved != stringValue {
		t.Errorf("Expected string %s, got %s (exists: %t)", stringValue, retrieved, exists)
	}
	
	// Set non-string value
	cache.Set("int_key", intValue)
	_, exists = cache.GetString("int_key")
	if exists {
		t.Error("GetString should return false for non-string values")
	}
}

func TestCacheGetFloat64(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "float_key"
	floatValue := 123.456
	stringValue := "not_a_float"
	
	// Set float value
	cache.Set(key, floatValue)
	retrieved, exists := cache.GetFloat64(key)
	if !exists || retrieved != floatValue {
		t.Errorf("Expected float %f, got %f (exists: %t)", floatValue, retrieved, exists)
	}
	
	// Set non-float value
	cache.Set("string_key", stringValue)
	_, exists = cache.GetFloat64("string_key")
	if exists {
		t.Error("GetFloat64 should return false for non-float values")
	}
}

func TestCacheGetTicker(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "ticker_key"
	ticker := &Ticker{
		Symbol: "BTC-USDT",
		Price:  45000.0,
	}
	
	// Set ticker
	cache.Set(key, ticker)
	retrieved, exists := cache.GetTicker(key)
	if !exists {
		t.Error("Ticker should exist in cache")
	}
	
	if retrieved.Symbol != ticker.Symbol || retrieved.Price != ticker.Price {
		t.Error("Retrieved ticker should match original")
	}
	
	// Test with non-ticker value
	cache.Set("non_ticker", "string")
	_, exists = cache.GetTicker("non_ticker")
	if exists {
		t.Error("GetTicker should return false for non-ticker values")
	}
}

func TestCacheGetKlines(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "klines_key"
	klines := []Kline{
		{Open: 45000, High: 45100, Low: 44900, Close: 45050, Volume: 100},
		{Open: 45050, High: 45200, Low: 45000, Close: 45150, Volume: 150},
	}
	
	// Set klines
	cache.Set(key, klines)
	retrieved, exists := cache.GetKlines(key)
	if !exists {
		t.Error("Klines should exist in cache")
	}
	
	if len(retrieved) != len(klines) {
		t.Errorf("Expected %d klines, got %d", len(klines), len(retrieved))
	}
	
	// Test with non-klines value
	cache.Set("non_klines", "string")
	_, exists = cache.GetKlines("non_klines")
	if exists {
		t.Error("GetKlines should return false for non-klines values")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "delete_test"
	value := "to_be_deleted"
	
	// Set and verify exists
	cache.Set(key, value)
	if !cache.Has(key) {
		t.Error("Key should exist before deletion")
	}
	
	// Delete
	cache.Delete(key)
	
	// Should not exist
	if cache.Has(key) {
		t.Error("Key should not exist after deletion")
	}
}

func TestCacheHas(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "has_test"
	value := "exists"
	
	// Should not exist initially
	if cache.Has(key) {
		t.Error("Key should not exist initially")
	}
	
	// Set value
	cache.Set(key, value)
	
	// Should exist now
	if !cache.Has(key) {
		t.Error("Key should exist after setting")
	}
	
	// Test with expired entry
	expiredKey := "expired"
	cache.SetWithTTL(expiredKey, value, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	
	if cache.Has(expiredKey) {
		t.Error("Expired key should not exist")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	// Add multiple entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}
	
	// Clear cache
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
	
	// Verify entries are gone
	if cache.Has("key1") || cache.Has("key2") || cache.Has("key3") {
		t.Error("No keys should exist after clear")
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	if cache.Size() != 0 {
		t.Errorf("Initial size should be 0, got %d", cache.Size())
	}
	
	// Add entries
	cache.Set("key1", "value1")
	if cache.Size() != 1 {
		t.Errorf("Size should be 1, got %d", cache.Size())
	}
	
	cache.Set("key2", "value2")
	if cache.Size() != 2 {
		t.Errorf("Size should be 2, got %d", cache.Size())
	}
	
	// Delete entry
	cache.Delete("key1")
	if cache.Size() != 1 {
		t.Errorf("Size should be 1 after deletion, got %d", cache.Size())
	}
}

func TestCacheKeys(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	// Add entries with different TTLs
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.SetWithTTL("expired_key", "value", 1*time.Millisecond)
	
	// Wait for one to expire
	time.Sleep(10 * time.Millisecond)
	
	keys := cache.Keys()
	
	// Should only return non-expired keys
	expectedKeys := []string{"key1", "key2"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}
	
	// Check keys exist
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key %s not found", expectedKey)
		}
	}
}

func TestCacheGetTTL(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "ttl_test"
	value := "test"
	ttl := 5 * time.Second
	
	// Set with custom TTL
	cache.SetWithTTL(key, value, ttl)
	
	// Get TTL immediately
	remainingTTL, exists := cache.GetTTL(key)
	if !exists {
		t.Error("TTL should exist for valid key")
	}
	
	if remainingTTL > ttl {
		t.Errorf("Remaining TTL %v should not exceed original TTL %v", remainingTTL, ttl)
	}
	
	// Test non-existent key
	_, exists = cache.GetTTL("non_existent")
	if exists {
		t.Error("TTL should not exist for non-existent key")
	}
}

func TestCacheExtend(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	key := "extend_test"
	value := "test"
	initialTTL := 100 * time.Millisecond
	extensionTTL := 1 * time.Second
	
	// Set with short TTL
	cache.SetWithTTL(key, value, initialTTL)
	
	// Extend TTL
	success := cache.Extend(key, extensionTTL)
	if !success {
		t.Error("Extend should succeed for existing key")
	}
	
	// Should still exist after initial TTL would have expired
	time.Sleep(200 * time.Millisecond)
	if !cache.Has(key) {
		t.Error("Key should still exist after extending TTL")
	}
	
	// Test extending non-existent key
	success = cache.Extend("non_existent", extensionTTL)
	if success {
		t.Error("Extend should fail for non-existent key")
	}
}

func TestCacheRefresh(t *testing.T) {
	cache := NewCache(1 * time.Second)
	defer cache.Stop()
	
	key := "refresh_test"
	value := "test"
	shortTTL := 50 * time.Millisecond
	
	// Set with short TTL
	cache.SetWithTTL(key, value, shortTTL)
	
	// Wait a bit then refresh
	time.Sleep(25 * time.Millisecond)
	success := cache.Refresh(key)
	if !success {
		t.Error("Refresh should succeed for existing key")
	}
	
	// Should still exist after original TTL would have expired
	time.Sleep(50 * time.Millisecond)
	if !cache.Has(key) {
		t.Error("Key should still exist after refresh")
	}
	
	// Test refreshing non-existent key
	success = cache.Refresh("non_existent")
	if success {
		t.Error("Refresh should fail for non-existent key")
	}
}

func TestCacheGetWithRefresh(t *testing.T) {
	cache := NewCache(1 * time.Second)
	defer cache.Stop()
	
	key := "refresh_get_test"
	value := "test"
	shortTTL := 50 * time.Millisecond
	
	// Set with short TTL
	cache.SetWithTTL(key, value, shortTTL)
	
	// Wait a bit then get with refresh
	time.Sleep(25 * time.Millisecond)
	retrievedValue, exists := cache.GetWithRefresh(key)
	if !exists {
		t.Error("GetWithRefresh should succeed for existing key")
	}
	
	if retrievedValue != value {
		t.Errorf("Expected %s, got %s", value, retrievedValue)
	}
	
	// Should still exist after original TTL would have expired
	time.Sleep(50 * time.Millisecond)
	if !cache.Has(key) {
		t.Error("Key should still exist after GetWithRefresh")
	}
}

func TestCacheGetStats(t *testing.T) {
	cache := NewCache(10 * time.Second)
	defer cache.Stop()
	
	// Add some entries
	cache.Set("active1", "value1")
	cache.Set("active2", "value2")
	cache.SetWithTTL("expired", "value", 1*time.Millisecond)
	
	// Wait for expiration
	time.Sleep(10 * time.Millisecond)
	
	stats := cache.GetStats()
	
	// Check required fields exist
	requiredFields := []string{"total_entries", "active_entries", "expired_entries", "default_ttl"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats should contain field: %s", field)
		}
	}
	
	totalEntries := stats["total_entries"].(int)
	if totalEntries < 2 {
		t.Errorf("Should have at least 2 total entries, got %d", totalEntries)
	}
}

func TestNewCachedMarketDataService(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	cachedService := NewCachedMarketDataService(service)
	defer cachedService.Stop()
	
	if cachedService == nil {
		t.Fatal("NewCachedMarketDataService should not return nil")
	}
	
	if cachedService.service != service {
		t.Error("CachedMarketDataService should reference the provided service")
	}
	
	if cachedService.cache == nil {
		t.Error("CachedMarketDataService should have a cache")
	}
}

func TestCachedMarketDataServiceInvalidateSymbol(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	cachedService := NewCachedMarketDataService(service)
	defer cachedService.Stop()
	
	// Add some cache entries
	cachedService.cache.Set("price:BTC-USDT", &Ticker{Symbol: "BTC-USDT"})
	cachedService.cache.Set("klines:BTC-USDT:5m:100", []Kline{})
	cachedService.cache.Set("ticker24hr:ETH-USDT", &Ticker{Symbol: "ETH-USDT"})
	
	// Invalidate BTC-USDT
	cachedService.InvalidateSymbol("BTC-USDT")
	
	// BTC-USDT entries should be gone
	if cachedService.cache.Has("price:BTC-USDT") {
		t.Error("BTC-USDT price cache should be invalidated")
	}
	
	if cachedService.cache.Has("klines:BTC-USDT:5m:100") {
		t.Error("BTC-USDT klines cache should be invalidated")
	}
	
	// ETH-USDT entry should still exist
	if !cachedService.cache.Has("ticker24hr:ETH-USDT") {
		t.Error("ETH-USDT ticker cache should not be affected")
	}
}

func TestCachedMarketDataServiceGetCacheStats(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewMarketDataService(client)
	cachedService := NewCachedMarketDataService(service)
	defer cachedService.Stop()
	
	stats := cachedService.GetCacheStats()
	
	// Should have cache stats
	if stats == nil {
		t.Error("GetCacheStats should return stats")
	}
	
	// Check for required fields
	if _, exists := stats["total_entries"]; !exists {
		t.Error("Stats should contain total_entries")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test contains function
	tests := []struct {
		s       string
		substr  string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"test", "test", true},
		{"", "test", false},
		{"test", "", true},
	}
	
	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %t, expected %t", tt.s, tt.substr, result, tt.expected)
		}
	}
	
	// Test indexOf function
	indexTests := []struct {
		s       string
		substr  string
		expected int
	}{
		{"hello world", "world", 6},
		{"hello world", "foo", -1},
		{"test", "test", 0},
		{"", "test", -1},
	}
	
	for _, tt := range indexTests {
		result := indexOf(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("indexOf(%q, %q) = %d, expected %d", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestCacheEntryIsExpired(t *testing.T) {
	// Test non-expired entry
	entry := &CacheEntry{
		Value:     "test",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	
	if entry.IsExpired() {
		t.Error("Entry should not be expired")
	}
	
	// Test expired entry
	expiredEntry := &CacheEntry{
		Value:     "test",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	
	if !expiredEntry.IsExpired() {
		t.Error("Entry should be expired")
	}
}
