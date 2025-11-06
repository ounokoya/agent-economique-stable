package bingx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuthManager handles BingX API authentication using HMAC SHA256
type AuthManager struct {
	credentials APICredentials
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(credentials APICredentials) *AuthManager {
	return &AuthManager{
		credentials: credentials,
	}
}

// ValidateCredentials checks if API credentials are properly formatted
func (a *AuthManager) ValidateCredentials() error {
	if len(a.credentials.APIKey) == 0 {
		return fmt.Errorf("API key cannot be empty")
	}
	
	if len(a.credentials.SecretKey) == 0 {
		return fmt.Errorf("secret key cannot be empty")
	}
	
	// BingX API key should be a valid format (basic validation)
	if len(a.credentials.APIKey) < 32 {
		return fmt.Errorf("API key appears to be invalid (too short)")
	}
	
	if len(a.credentials.SecretKey) < 32 {
		return fmt.Errorf("secret key appears to be invalid (too short)")
	}
	
	return nil
}

// GenerateSignature creates HMAC SHA256 signature for BingX API request
func (a *AuthManager) GenerateSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(a.credentials.SecretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// BuildAuthenticatedURL constructs a complete authenticated URL with signature
func (a *AuthManager) BuildAuthenticatedURL(baseURL, endpoint string, params map[string]string) (string, error) {
	// Validate credentials first
	if err := a.ValidateCredentials(); err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}
	
	// Add timestamp if not present
	if _, exists := params["timestamp"]; !exists {
		params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	
	// Build query string
	queryString := a.buildQueryString(params)
	
	// Generate signature
	signature := a.GenerateSignature(queryString)
	
	// Add signature to parameters
	params["signature"] = signature
	
	// Build final URL
	finalQueryString := a.buildQueryString(params)
	return fmt.Sprintf("%s%s?%s", baseURL, endpoint, finalQueryString), nil
}

// GetAuthHeaders returns the required headers for BingX API authentication
func (a *AuthManager) GetAuthHeaders() map[string]string {
	return map[string]string{
		"X-BX-APIKEY": a.credentials.APIKey,
		"Content-Type": "application/json",
	}
}

// BuildSignedParams creates signed parameters for POST requests
func (a *AuthManager) BuildSignedParams(params map[string]string) (map[string]string, error) {
	if err := a.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	
	// Add timestamp if not present
	if _, exists := params["timestamp"]; !exists {
		params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	
	// Build query string for signature
	queryString := a.buildQueryString(params)
	
	// Generate signature
	signature := a.GenerateSignature(queryString)
	
	// Create result params with signature
	result := make(map[string]string)
	for k, v := range params {
		result[k] = v
	}
	result["signature"] = signature
	
	return result, nil
}

// VerifySignature verifies if a signature is correct for given parameters
func (a *AuthManager) VerifySignature(params map[string]string, expectedSignature string) bool {
	// Remove signature from params for verification
	verifyParams := make(map[string]string)
	for k, v := range params {
		if k != "signature" {
			verifyParams[k] = v
		}
	}
	
	queryString := a.buildQueryString(verifyParams)
	actualSignature := a.GenerateSignature(queryString)
	
	return actualSignature == expectedSignature
}

// GetTimestamp returns current timestamp in milliseconds (BingX format)
func (a *AuthManager) GetTimestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// IsTimestampValid checks if timestamp is within acceptable range (5 minutes)
func (a *AuthManager) IsTimestampValid(timestampStr string) bool {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}
	
	now := time.Now().UnixMilli()
	diff := now - timestamp
	
	// Accept timestamps within 5 minutes (300,000 milliseconds)
	return diff >= 0 && diff <= 300000
}

// buildQueryString creates a sorted query string from parameters
func (a *AuthManager) buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	
	// Sort keys for consistent signature generation
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build query string
	var parts []string
	for _, key := range keys {
		value := params[key]
		// URL encode the value
		encodedValue := url.QueryEscape(value)
		parts = append(parts, fmt.Sprintf("%s=%s", key, encodedValue))
	}
	
	return strings.Join(parts, "&")
}

// SanitizeParams removes empty values and trims whitespace
func (a *AuthManager) SanitizeParams(params map[string]string) map[string]string {
	sanitized := make(map[string]string)
	
	for key, value := range params {
		// Trim whitespace
		trimmedValue := strings.TrimSpace(value)
		
		// Only include non-empty values
		if trimmedValue != "" {
			sanitized[key] = trimmedValue
		}
	}
	
	return sanitized
}

// GetCredentials returns a copy of the stored credentials (for testing)
func (a *AuthManager) GetCredentials() APICredentials {
	return APICredentials{
		APIKey:    a.credentials.APIKey,
		SecretKey: a.credentials.SecretKey,
	}
}
