package bingx

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test credentials for unit tests
var testCredentials = APICredentials{
	APIKey:    "Zsm4DcrHBTewmVaElrdwA67PmivPv6VDK6JAkiECZ9QfcUnmn67qjCOgvRuZVOzU",
	SecretKey: "UuGuyEGt6ZEkpUObCYCmIfh0elYsZVh80jlYwpJuRZEw70t6vomMH7Sjmf94ztSI",
}

func TestNewAuthManager(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	if auth == nil {
		t.Fatal("NewAuthManager should not return nil")
	}
	
	creds := auth.GetCredentials()
	if creds.APIKey != testCredentials.APIKey {
		t.Errorf("Expected API key %s, got %s", testCredentials.APIKey, creds.APIKey)
	}
	
	if creds.SecretKey != testCredentials.SecretKey {
		t.Errorf("Expected secret key %s, got %s", testCredentials.SecretKey, creds.SecretKey)
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name        string
		credentials APICredentials
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid credentials",
			credentials: testCredentials,
			expectError: false,
		},
		{
			name: "Empty API key",
			credentials: APICredentials{
				APIKey:    "",
				SecretKey: testCredentials.SecretKey,
			},
			expectError: true,
			errorMsg:    "API key cannot be empty",
		},
		{
			name: "Empty secret key",
			credentials: APICredentials{
				APIKey:    testCredentials.APIKey,
				SecretKey: "",
			},
			expectError: true,
			errorMsg:    "secret key cannot be empty",
		},
		{
			name: "Short API key",
			credentials: APICredentials{
				APIKey:    "short",
				SecretKey: testCredentials.SecretKey,
			},
			expectError: true,
			errorMsg:    "API key appears to be invalid (too short)",
		},
		{
			name: "Short secret key",
			credentials: APICredentials{
				APIKey:    testCredentials.APIKey,
				SecretKey: "short",
			},
			expectError: true,
			errorMsg:    "secret key appears to be invalid (too short)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAuthManager(tt.credentials)
			err := auth.ValidateCredentials()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGenerateSignature(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	// Test with known parameters from BingX documentation
	params := "quoteOrderQty=20&side=BUY&symbol=ETHUSDT&timestamp=1649404670162&type=MARKET"
	expectedSignature := "428a3c383bde514baff0d10d3c20e5adfaacaf799e324546dafe5ccc480dd827"
	
	signature := auth.GenerateSignature(params)
	
	if signature != expectedSignature {
		t.Errorf("Expected signature %s, got %s", expectedSignature, signature)
	}
}

func TestGenerateSignatureEmpty(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	signature := auth.GenerateSignature("")
	
	if signature == "" {
		t.Error("Signature should not be empty even for empty input")
	}
	
	// Should be consistent
	signature2 := auth.GenerateSignature("")
	if signature != signature2 {
		t.Error("Signature should be consistent for same input")
	}
}

func TestBuildQueryString(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	tests := []struct {
		name     string
		params   map[string]string
		expected string
	}{
		{
			name:     "Empty params",
			params:   map[string]string{},
			expected: "",
		},
		{
			name: "Single param",
			params: map[string]string{
				"symbol": "BTCUSDT",
			},
			expected: "symbol=BTCUSDT",
		},
		{
			name: "Multiple params (sorted)",
			params: map[string]string{
				"symbol":    "BTCUSDT",
				"side":      "BUY",
				"timestamp": "1649404670162",
			},
			expected: "side=BUY&symbol=BTCUSDT&timestamp=1649404670162",
		},
		{
			name: "Params with special characters",
			params: map[string]string{
				"symbol": "BTC-USDT",
				"note":   "test order",
			},
			expected: "note=test+order&symbol=BTC-USDT",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.buildQueryString(tt.params)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildAuthenticatedURL(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	baseURL := "https://open-api.bingx.com"
	endpoint := "/openApi/spot/v1/trade/order"
	params := map[string]string{
		"symbol":    "BTCUSDT",
		"side":      "BUY",
		"type":      "MARKET",
		"quantity":  "0.001",
		"timestamp": "1649404670162",
	}
	
	url, err := auth.BuildAuthenticatedURL(baseURL, endpoint, params)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !strings.HasPrefix(url, baseURL+endpoint) {
		t.Errorf("URL should start with %s%s", baseURL, endpoint)
	}
	
	if !strings.Contains(url, "signature=") {
		t.Error("URL should contain signature parameter")
	}
	
	if !strings.Contains(url, "symbol=BTCUSDT") {
		t.Error("URL should contain original parameters")
	}
}

func TestBuildAuthenticatedURLInvalidCredentials(t *testing.T) {
	invalidCreds := APICredentials{
		APIKey:    "",
		SecretKey: "valid_secret",
	}
	auth := NewAuthManager(invalidCreds)
	
	_, err := auth.BuildAuthenticatedURL("https://api.test.com", "/test", map[string]string{})
	
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

func TestGetAuthHeaders(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	headers := auth.GetAuthHeaders()
	
	expectedHeaders := map[string]string{
		"X-BX-APIKEY":  testCredentials.APIKey,
		"Content-Type": "application/json",
	}
	
	for key, expectedValue := range expectedHeaders {
		if value, exists := headers[key]; !exists {
			t.Errorf("Missing header: %s", key)
		} else if value != expectedValue {
			t.Errorf("Expected header %s to be %s, got %s", key, expectedValue, value)
		}
	}
}

func TestBuildSignedParams(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	params := map[string]string{
		"symbol": "BTCUSDT",
		"side":   "BUY",
	}
	
	signedParams, err := auth.BuildSignedParams(params)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should contain original params
	for key, value := range params {
		if signedValue, exists := signedParams[key]; !exists {
			t.Errorf("Missing parameter: %s", key)
		} else if signedValue != value {
			t.Errorf("Parameter %s changed from %s to %s", key, value, signedValue)
		}
	}
	
	// Should contain signature
	if _, exists := signedParams["signature"]; !exists {
		t.Error("Missing signature parameter")
	}
	
	// Should contain timestamp
	if _, exists := signedParams["timestamp"]; !exists {
		t.Error("Missing timestamp parameter")
	}
}

func TestBuildSignedParamsInvalidCredentials(t *testing.T) {
	invalidCreds := APICredentials{
		APIKey:    "short",
		SecretKey: "also_short",
	}
	auth := NewAuthManager(invalidCreds)
	
	_, err := auth.BuildSignedParams(map[string]string{"test": "value"})
	
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

func TestVerifySignature(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	params := map[string]string{
		"symbol":    "BTCUSDT",
		"side":      "BUY",
		"timestamp": "1649404670162",
	}
	
	// Generate signature
	queryString := auth.buildQueryString(params)
	expectedSignature := auth.GenerateSignature(queryString)
	
	// Verify with correct signature
	params["signature"] = expectedSignature
	if !auth.VerifySignature(params, expectedSignature) {
		t.Error("Should verify correct signature")
	}
	
	// Verify with incorrect signature
	if auth.VerifySignature(params, "wrong_signature") {
		t.Error("Should not verify incorrect signature")
	}
}

func TestGetTimestamp(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	timestamp1 := auth.GetTimestamp()
	time.Sleep(time.Millisecond)
	timestamp2 := auth.GetTimestamp()
	
	if timestamp1 == timestamp2 {
		t.Error("Timestamps should be different")
	}
	
	// Should be valid integers
	_, err1 := strconv.ParseInt(timestamp1, 10, 64)
	_, err2 := strconv.ParseInt(timestamp2, 10, 64)
	
	if err1 != nil || err2 != nil {
		t.Error("Timestamps should be valid integers")
	}
}

func TestIsTimestampValid(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	// Current timestamp should be valid
	currentTimestamp := auth.GetTimestamp()
	if !auth.IsTimestampValid(currentTimestamp) {
		t.Error("Current timestamp should be valid")
	}
	
	// Future timestamp (within 5 minutes) should be invalid
	futureTime := time.Now().Add(10 * time.Minute).UnixMilli()
	futureTimestamp := strconv.FormatInt(futureTime, 10)
	if auth.IsTimestampValid(futureTimestamp) {
		t.Error("Future timestamp should be invalid")
	}
	
	// Old timestamp (more than 5 minutes) should be invalid
	oldTime := time.Now().Add(-10 * time.Minute).UnixMilli()
	oldTimestamp := strconv.FormatInt(oldTime, 10)
	if auth.IsTimestampValid(oldTimestamp) {
		t.Error("Old timestamp should be invalid")
	}
	
	// Invalid format should be invalid
	if auth.IsTimestampValid("invalid") {
		t.Error("Invalid timestamp format should be invalid")
	}
}

func TestSanitizeParams(t *testing.T) {
	auth := NewAuthManager(testCredentials)
	
	params := map[string]string{
		"symbol":       "BTCUSDT",
		"side":         " BUY ",
		"empty_value":  "",
		"whitespace":   "  ",
		"normal":       "value",
	}
	
	sanitized := auth.SanitizeParams(params)
	
	expected := map[string]string{
		"symbol": "BTCUSDT",
		"side":   "BUY",
		"normal": "value",
	}
	
	if len(sanitized) != len(expected) {
		t.Errorf("Expected %d parameters, got %d", len(expected), len(sanitized))
	}
	
	for key, expectedValue := range expected {
		if value, exists := sanitized[key]; !exists {
			t.Errorf("Missing parameter: %s", key)
		} else if value != expectedValue {
			t.Errorf("Expected %s to be '%s', got '%s'", key, expectedValue, value)
		}
	}
	
	// Should not contain empty or whitespace-only values
	for key := range sanitized {
		if key == "empty_value" || key == "whitespace" {
			t.Errorf("Should not contain empty parameter: %s", key)
		}
	}
}
