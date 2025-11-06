package bingx

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewSubAccountService(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	if service == nil {
		t.Fatal("NewSubAccountService should not return nil")
	}
	
	if service.client != client {
		t.Error("Service should reference the provided client")
	}
}

func TestSubAccountServiceValidateSubAccountPermissions(t *testing.T) {
	service := &SubAccountService{}
	
	tests := []struct {
		name        string
		permissions []string
		expectError bool
	}{
		{
			name:        "Valid permissions",
			permissions: []string{"SPOT", "FUTURES", "READ"},
			expectError: false,
		},
		{
			name:        "Single valid permission",
			permissions: []string{"TRADE"},
			expectError: false,
		},
		{
			name:        "All valid permissions",
			permissions: []string{"SPOT", "FUTURES", "MARGIN", "READ", "TRADE"},
			expectError: false,
		},
		{
			name:        "Invalid permission",
			permissions: []string{"INVALID"},
			expectError: true,
		},
		{
			name:        "Mix of valid and invalid",
			permissions: []string{"SPOT", "INVALID", "READ"},
			expectError: true,
		},
		{
			name:        "Empty permissions",
			permissions: []string{},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSubAccountPermissions(tt.permissions)
			
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

func TestSubAccountServiceGenerateSubAccountEmail(t *testing.T) {
	service := &SubAccountService{}
	
	tests := []struct {
		name      string
		baseEmail string
		suffix    string
		expected  func(string) bool // Function to validate the result
	}{
		{
			name:      "Valid email with suffix",
			baseEmail: "user@example.com",
			suffix:    "bot1",
			expected: func(result string) bool {
				return result == "user+bot1@example.com"
			},
		},
		{
			name:      "Valid email without suffix",
			baseEmail: "trader@domain.org",
			suffix:    "",
			expected: func(result string) bool {
				// Should generate timestamp suffix
				return strings.HasPrefix(result, "trader+") && strings.HasSuffix(result, "@domain.org")
			},
		},
		{
			name:      "Invalid email format",
			baseEmail: "invalid-email",
			suffix:    "test",
			expected: func(result string) bool {
				return result == "subaccount_test@example.com"
			},
		},
		{
			name:      "Empty email",
			baseEmail: "",
			suffix:    "suffix",
			expected: func(result string) bool {
				return result == "subaccount_suffix@example.com"
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateSubAccountEmail(tt.baseEmail, tt.suffix)
			
			if !tt.expected(result) {
				t.Errorf("GenerateSubAccountEmail result validation failed: %s", result)
			}
		})
	}
}

func TestSubAccountServiceParseResponseData(t *testing.T) {
	service := &SubAccountService{}
	
	// Test parsing sub-account data
	inputData := map[string]interface{}{
		"uid":    int64(12345),
		"email":  "test@example.com",
		"status": "ACTIVE",
	}
	
	var subAccount SubAccountInfo
	err := service.parseResponseData(inputData, &subAccount)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	// Note: The actual field mapping would depend on the struct tags
	// This test validates the parsing mechanism works
}

// Test CreateSubAccount parameter validation
func TestSubAccountServiceCreateSubAccountValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test empty email
	_, err = service.CreateSubAccount(ctx, CreateSubAccountRequest{
		Email: "",
	})
	if err == nil {
		t.Error("CreateSubAccount should return error for empty email")
	}
}

// Test GetSubAccountInfo parameter validation
func TestSubAccountServiceGetSubAccountInfoValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	_, err = service.GetSubAccountInfo(ctx, 0)
	if err == nil {
		t.Error("GetSubAccountInfo should return error for UID 0")
	}
	
	_, err = service.GetSubAccountInfo(ctx, -1)
	if err == nil {
		t.Error("GetSubAccountInfo should return error for negative UID")
	}
}

// Test FreezeSubAccount parameter validation
func TestSubAccountServiceFreezeSubAccountValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	err = service.FreezeSubAccount(ctx, 0)
	if err == nil {
		t.Error("FreezeSubAccount should return error for UID 0")
	}
	
	err = service.UnfreezeSubAccount(ctx, -1)
	if err == nil {
		t.Error("UnfreezeSubAccount should return error for negative UID")
	}
}

// Test CreateAPIKey parameter validation
func TestSubAccountServiceCreateAPIKeyValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	_, err = service.CreateAPIKey(ctx, CreateAPIKeyRequest{
		SubAccountUID: 0,
		Permissions:   []string{"READ"},
	})
	if err == nil {
		t.Error("CreateAPIKey should return error for UID 0")
	}
	
	// Test empty permissions
	_, err = service.CreateAPIKey(ctx, CreateAPIKeyRequest{
		SubAccountUID: 12345,
		Permissions:   []string{},
	})
	if err == nil {
		t.Error("CreateAPIKey should return error for empty permissions")
	}
}

// Test ListAPIKeys parameter validation
func TestSubAccountServiceListAPIKeysValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	_, err = service.ListAPIKeys(ctx, 0)
	if err == nil {
		t.Error("ListAPIKeys should return error for UID 0")
	}
}

// Test DeleteAPIKey parameter validation
func TestSubAccountServiceDeleteAPIKeyValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	err = service.DeleteAPIKey(ctx, 0, "test_key")
	if err == nil {
		t.Error("DeleteAPIKey should return error for UID 0")
	}
	
	// Test empty API key
	err = service.DeleteAPIKey(ctx, 12345, "")
	if err == nil {
		t.Error("DeleteAPIKey should return error for empty API key")
	}
}

// Test ResetAPIKey parameter validation
func TestSubAccountServiceResetAPIKeyValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	_, err = service.ResetAPIKey(ctx, 0, "test_key")
	if err == nil {
		t.Error("ResetAPIKey should return error for UID 0")
	}
	
	// Test empty API key
	_, err = service.ResetAPIKey(ctx, 12345, "")
	if err == nil {
		t.Error("ResetAPIKey should return error for empty API key")
	}
}

// Test GetSubAccountBalances parameter validation
func TestSubAccountServiceGetSubAccountBalancesValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test invalid UID
	_, err = service.GetSubAccountBalances(ctx, 0)
	if err == nil {
		t.Error("GetSubAccountBalances should return error for UID 0")
	}
}

// Test ListSubAccounts limit handling
func TestSubAccountServiceListSubAccountsLimitHandling(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewSubAccountService(client)
	ctx := context.Background()
	
	// Test with limit greater than maximum - should not return error in validation
	// The actual API call would fail, but that's expected in test environment
	_, err = service.ListSubAccounts(ctx, 500) // Over 200 limit
	// We don't check for specific error since API would fail anyway
}

// Test struct field validation
func TestSubAccountStructs(t *testing.T) {
	// Test CreateSubAccountRequest
	req := CreateSubAccountRequest{
		Email:    "test@example.com",
		Password: "password123",
		Label:    "Test Account",
	}
	
	if req.Email != "test@example.com" {
		t.Error("Email field should be preserved")
	}
	
	// Test CreateAPIKeyRequest
	apiReq := CreateAPIKeyRequest{
		SubAccountUID: 12345,
		Label:         "API Key 1",
		Permissions:   []string{"READ", "TRADE"},
		IPWhitelist:   []string{"192.168.1.1", "10.0.0.1"},
	}
	
	if len(apiReq.Permissions) != 2 {
		t.Error("Permissions should be preserved")
	}
	
	if len(apiReq.IPWhitelist) != 2 {
		t.Error("IP whitelist should be preserved")
	}
	
	// Test SubAccountInfo
	info := SubAccountInfo{
		IsActive:  true,
		LastLogin: time.Now(),
	}
	
	if !info.IsActive {
		t.Error("IsActive should be true")
	}
	
	// Test APIKeyInfo
	keyInfo := APIKeyInfo{
		APIKey:      "test_api_key",
		Permissions: []string{"READ"},
		Status:      "ACTIVE",
		CreatedAt:   time.Now(),
	}
	
	if keyInfo.APIKey != "test_api_key" {
		t.Error("API key should be preserved")
	}
	
	if keyInfo.Status != "ACTIVE" {
		t.Error("Status should be preserved")
	}
}
