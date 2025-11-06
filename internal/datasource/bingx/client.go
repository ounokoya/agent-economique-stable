package bingx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// EndpointType represents different types of API endpoints for rate limiting
type EndpointType int

const (
	EndpointTypeMarketData EndpointType = iota
	EndpointTypeTrading
	EndpointTypeAccount
)

// RateLimiter manages rate limiting for different endpoint types
type RateLimiter struct {
	marketDataLimiter *rate.Limiter
	tradingLimiter    *rate.Limiter
	accountLimiter    *rate.Limiter
	globalLimiter     *rate.Limiter
	mutex             sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with BingX limits
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		marketDataLimiter: rate.NewLimiter(rate.Limit(MarketDataRateLimit), MarketDataRateLimit*2), // burst allowance
		tradingLimiter:    rate.NewLimiter(rate.Limit(TradingRateLimit), TradingRateLimit*2),
		accountLimiter:    rate.NewLimiter(rate.Limit(AccountRateLimit), AccountRateLimit*2),
		globalLimiter:     rate.NewLimiter(rate.Limit(MarketDataRateLimit+TradingRateLimit), 50), // global limit
	}
}

// Wait waits for permission to make a request based on endpoint type
func (rl *RateLimiter) Wait(ctx context.Context, endpointType EndpointType) error {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	// Wait for specific endpoint limiter
	switch endpointType {
	case EndpointTypeMarketData:
		if err := rl.marketDataLimiter.Wait(ctx); err != nil {
			return err
		}
	case EndpointTypeTrading:
		if err := rl.tradingLimiter.Wait(ctx); err != nil {
			return err
		}
	case EndpointTypeAccount:
		if err := rl.accountLimiter.Wait(ctx); err != nil {
			return err
		}
	}

	// Wait for global limiter
	return rl.globalLimiter.Wait(ctx)
}

// Client represents a BingX API client
type Client struct {
	config      ClientConfig
	authManager *AuthManager
	rateLimiter *RateLimiter
	httpClient  *http.Client
	mutex       sync.RWMutex
}

// NewClient creates a new BingX API client
func NewClient(config ClientConfig) (*Client, error) {
	// Validate configuration
	if err := validateClientConfig(config); err != nil {
		return nil, fmt.Errorf("invalid client configuration: %w", err)
	}

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Set base URL based on environment if not specified
	if config.BaseURL == "" {
		switch config.Environment {
		case DemoEnvironment:
			config.BaseURL = DemoBaseURL
		case LiveEnvironment:
			config.BaseURL = LiveBaseURL
		default:
			return nil, fmt.Errorf("unknown environment: %s", config.Environment)
		}
	}

	authManager := NewAuthManager(config.Credentials)
	if err := authManager.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	return &Client{
		config:      config,
		authManager: authManager,
		rateLimiter: NewRateLimiter(),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// DoRequest performs an authenticated HTTP request with rate limiting
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, params map[string]string, endpointType EndpointType) (*APIResponse, error) {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx, endpointType); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		req, err = c.buildGETRequest(endpoint, params)
	case http.MethodPost:
		req, err = c.buildPOSTRequest(endpoint, params)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// buildGETRequest builds an authenticated GET request
func (c *Client) buildGETRequest(endpoint string, params map[string]string) (*http.Request, error) {
	// Sanitize parameters
	cleanParams := c.authManager.SanitizeParams(params)

	// Build authenticated URL
	url, err := c.authManager.BuildAuthenticatedURL(c.config.BaseURL, endpoint, cleanParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.authManager.GetAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// buildPOSTRequest builds an authenticated POST request
func (c *Client) buildPOSTRequest(endpoint string, params map[string]string) (*http.Request, error) {
	// Sanitize parameters
	cleanParams := c.authManager.SanitizeParams(params)

	// Build signed parameters
	signedParams, err := c.authManager.BuildSignedParams(cleanParams)
	if err != nil {
		return nil, err
	}

	// Convert to JSON
	jsonData, err := json.Marshal(signedParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	url := c.config.BaseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.authManager.GetAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// parseResponse parses the HTTP response into an APIResponse
func (c *Client) parseResponse(resp *http.Response) (*APIResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, errorResp
		}
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API-level errors
	if apiResp.Code != 0 {
		return nil, ErrorResponse{
			Code: apiResp.Code,
			Msg:  apiResp.Msg,
		}
	}

	return &apiResp, nil
}

// GetConfig returns a copy of the client configuration
func (c *Client) GetConfig() ClientConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config
}

// UpdateCredentials updates the API credentials
func (c *Client) UpdateCredentials(newCredentials APICredentials) error {
	newAuthManager := NewAuthManager(newCredentials)
	if err := newAuthManager.ValidateCredentials(); err != nil {
		return fmt.Errorf("invalid new credentials: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.config.Credentials = newCredentials
	c.authManager = newAuthManager

	return nil
}

// Ping performs a connectivity test to BingX API
func (c *Client) Ping(ctx context.Context) error {
	// Use server time endpoint for ping test
	_, err := c.DoRequest(ctx, http.MethodGet, "/openApi/swap/v2/server/time", nil, EndpointTypeMarketData)
	return err
}

// GetRateLimitStatus returns current rate limit usage information
func (c *Client) GetRateLimitStatus() map[string]interface{} {
	c.rateLimiter.mutex.RLock()
	defer c.rateLimiter.mutex.RUnlock()

	return map[string]interface{}{
		"market_data_tokens": c.rateLimiter.marketDataLimiter.Tokens(),
		"trading_tokens":     c.rateLimiter.tradingLimiter.Tokens(),
		"account_tokens":     c.rateLimiter.accountLimiter.Tokens(),
		"global_tokens":      c.rateLimiter.globalLimiter.Tokens(),
		"market_data_limit":  c.rateLimiter.marketDataLimiter.Limit(),
		"trading_limit":      c.rateLimiter.tradingLimiter.Limit(),
		"account_limit":      c.rateLimiter.accountLimiter.Limit(),
	}
}

// validateClientConfig validates the client configuration
func validateClientConfig(config ClientConfig) error {
	if config.Environment != DemoEnvironment && config.Environment != LiveEnvironment {
		return fmt.Errorf("environment must be either '%s' or '%s'", DemoEnvironment, LiveEnvironment)
	}

	if config.Credentials.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if config.Credentials.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	return nil
}
