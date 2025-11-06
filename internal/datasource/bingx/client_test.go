package bingx

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter()
	
	if rl == nil {
		t.Fatal("NewRateLimiter should not return nil")
	}
	
	if rl.marketDataLimiter == nil {
		t.Error("marketDataLimiter should not be nil")
	}
	
	if rl.tradingLimiter == nil {
		t.Error("tradingLimiter should not be nil")
	}
	
	if rl.accountLimiter == nil {
		t.Error("accountLimiter should not be nil")
	}
	
	if rl.globalLimiter == nil {
		t.Error("globalLimiter should not be nil")
	}
}

func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter()
	ctx := context.Background()
	
	// Test different endpoint types
	endpointTypes := []EndpointType{
		EndpointTypeMarketData,
		EndpointTypeTrading,
		EndpointTypeAccount,
	}
	
	for _, et := range endpointTypes {
		err := rl.Wait(ctx, et)
		if err != nil {
			t.Errorf("Wait should not return error for endpoint type %d: %v", et, err)
		}
	}
}

func TestRateLimiterWaitWithTimeout(t *testing.T) {
	rl := NewRateLimiter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Exhaust the rate limiter
	for i := 0; i < MarketDataRateLimit*2+10; i++ {
		rl.Wait(context.Background(), EndpointTypeMarketData)
	}
	
	// This should timeout
	err := rl.Wait(ctx, EndpointTypeMarketData)
	if err == nil {
		t.Error("Wait should return error when context times out")
	}
}

func TestNewClient(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	
	if err != nil {
		t.Fatalf("NewClient should not return error: %v", err)
	}
	
	if client == nil {
		t.Fatal("NewClient should not return nil")
	}
	
	clientConfig := client.GetConfig()
	if clientConfig.Environment != DemoEnvironment {
		t.Errorf("Expected environment %s, got %s", DemoEnvironment, clientConfig.Environment)
	}
	
	if clientConfig.BaseURL != DemoBaseURL {
		t.Errorf("Expected base URL %s, got %s", DemoBaseURL, clientConfig.BaseURL)
	}
}

func TestNewClientWithCustomBaseURL(t *testing.T) {
	customURL := "https://custom.api.com"
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
		BaseURL:     customURL,
	}
	
	client, err := NewClient(config)
	
	if err != nil {
		t.Fatalf("NewClient should not return error: %v", err)
	}
	
	clientConfig := client.GetConfig()
	if clientConfig.BaseURL != customURL {
		t.Errorf("Expected custom base URL %s, got %s", customURL, clientConfig.BaseURL)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 60 * time.Second
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
		Timeout:     timeout,
	}
	
	client, err := NewClient(config)
	
	if err != nil {
		t.Fatalf("NewClient should not return error: %v", err)
	}
	
	clientConfig := client.GetConfig()
	if clientConfig.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, clientConfig.Timeout)
	}
}

func TestNewClientInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config ClientConfig
	}{
		{
			name: "Invalid environment",
			config: ClientConfig{
				Environment: Environment("invalid"),
				Credentials: testCredentials,
			},
		},
		{
			name: "Empty API key",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: APICredentials{
					APIKey:    "",
					SecretKey: testCredentials.SecretKey,
				},
			},
		},
		{
			name: "Empty secret key",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: APICredentials{
					APIKey:    testCredentials.APIKey,
					SecretKey: "",
				},
			},
		},
		{
			name: "Negative timeout",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: testCredentials,
				Timeout:     -1 * time.Second,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			if err == nil {
				t.Error("NewClient should return error for invalid config")
			}
		})
	}
}

func TestClientUpdateCredentials(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	newCredentials := APICredentials{
		APIKey:    "new_api_key_12345678901234567890123456789012",
		SecretKey: "new_secret_key_12345678901234567890123456789012",
	}
	
	err = client.UpdateCredentials(newCredentials)
	if err != nil {
		t.Errorf("UpdateCredentials should not return error: %v", err)
	}
	
	updatedConfig := client.GetConfig()
	if updatedConfig.Credentials.APIKey != newCredentials.APIKey {
		t.Errorf("API key not updated correctly")
	}
	
	if updatedConfig.Credentials.SecretKey != newCredentials.SecretKey {
		t.Errorf("Secret key not updated correctly")
	}
}

func TestClientUpdateCredentialsInvalid(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	invalidCredentials := APICredentials{
		APIKey:    "short",
		SecretKey: "also_short",
	}
	
	err = client.UpdateCredentials(invalidCredentials)
	if err == nil {
		t.Error("UpdateCredentials should return error for invalid credentials")
	}
	
	// Original credentials should remain unchanged
	originalConfig := client.GetConfig()
	if originalConfig.Credentials.APIKey != testCredentials.APIKey {
		t.Error("Original API key should remain unchanged after failed update")
	}
}

func TestClientGetRateLimitStatus(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	status := client.GetRateLimitStatus()
	
	expectedKeys := []string{
		"market_data_tokens",
		"trading_tokens",
		"account_tokens",
		"global_tokens",
		"market_data_limit",
		"trading_limit",
		"account_limit",
	}
	
	for _, key := range expectedKeys {
		if _, exists := status[key]; !exists {
			t.Errorf("Missing key in rate limit status: %s", key)
		}
	}
}

func TestValidateClientConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      ClientConfig
		expectError bool
	}{
		{
			name: "Valid config",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: testCredentials,
			},
			expectError: false,
		},
		{
			name: "Invalid environment",
			config: ClientConfig{
				Environment: Environment("invalid"),
				Credentials: testCredentials,
			},
			expectError: true,
		},
		{
			name: "Empty API key",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: APICredentials{
					APIKey:    "",
					SecretKey: testCredentials.SecretKey,
				},
			},
			expectError: true,
		},
		{
			name: "Empty secret key",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: APICredentials{
					APIKey:    testCredentials.APIKey,
					SecretKey: "",
				},
			},
			expectError: true,
		},
		{
			name: "Negative timeout",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: testCredentials,
				Timeout:     -1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Zero timeout (valid)",
			config: ClientConfig{
				Environment: DemoEnvironment,
				Credentials: testCredentials,
				Timeout:     0,
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClientConfig(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestClientBuildGETRequest(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	params := map[string]string{
		"symbol": "BTCUSDT",
		"limit":  "100",
	}
	
	req, err := client.buildGETRequest("/test/endpoint", params)
	if err != nil {
		t.Fatalf("buildGETRequest failed: %v", err)
	}
	
	if req.Method != http.MethodGet {
		t.Errorf("Expected GET method, got %s", req.Method)
	}
	
	if req.Header.Get("X-BX-APIKEY") != testCredentials.APIKey {
		t.Error("Missing or incorrect X-BX-APIKEY header")
	}
	
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Missing or incorrect Content-Type header")
	}
	
	if !containsSubstring(req.URL.String(), "signature=") {
		t.Error("URL should contain signature parameter")
	}
}

func TestClientBuildPOSTRequest(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	params := map[string]string{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"quantity": "0.001",
	}
	
	req, err := client.buildPOSTRequest("/test/endpoint", params)
	if err != nil {
		t.Fatalf("buildPOSTRequest failed: %v", err)
	}
	
	if req.Method != http.MethodPost {
		t.Errorf("Expected POST method, got %s", req.Method)
	}
	
	if req.Header.Get("X-BX-APIKEY") != testCredentials.APIKey {
		t.Error("Missing or incorrect X-BX-APIKEY header")
	}
	
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Missing or incorrect Content-Type header")
	}
	
	if req.Body == nil {
		t.Error("POST request should have a body")
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(hasPrefix(s, substr) || hasSuffix(s, substr) || containsInMiddle(s, substr)))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
