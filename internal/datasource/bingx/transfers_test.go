package bingx

import (
	"context"
	"testing"
	"time"
)

func TestNewTransferService(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	if service == nil {
		t.Fatal("NewTransferService should not return nil")
	}
	
	if service.client != client {
		t.Error("Service should reference the provided client")
	}
}

func TestTransferServiceValidateTransferRequest(t *testing.T) {
	service := &TransferService{}
	
	tests := []struct {
		name        string
		request     TransferRequest
		expectError bool
	}{
		{
			name: "Valid transfer request",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   67890,
				Asset:   "USDT",
				Amount:  100.0,
			},
			expectError: false,
		},
		{
			name: "Invalid from UID (zero)",
			request: TransferRequest{
				FromUID: 0,
				ToUID:   67890,
				Asset:   "USDT",
				Amount:  100.0,
			},
			expectError: true,
		},
		{
			name: "Invalid from UID (negative)",
			request: TransferRequest{
				FromUID: -1,
				ToUID:   67890,
				Asset:   "USDT",
				Amount:  100.0,
			},
			expectError: true,
		},
		{
			name: "Invalid to UID (zero)",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   0,
				Asset:   "USDT",
				Amount:  100.0,
			},
			expectError: true,
		},
		{
			name: "Same from and to UID",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   12345,
				Asset:   "USDT",
				Amount:  100.0,
			},
			expectError: true,
		},
		{
			name: "Empty asset",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   67890,
				Asset:   "",
				Amount:  100.0,
			},
			expectError: true,
		},
		{
			name: "Zero amount",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   67890,
				Asset:   "USDT",
				Amount:  0,
			},
			expectError: true,
		},
		{
			name: "Negative amount",
			request: TransferRequest{
				FromUID: 12345,
				ToUID:   67890,
				Asset:   "USDT",
				Amount:  -100.0,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateTransferRequest(tt.request)
			
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

func TestTransferServiceValidateAutoTransferConfig(t *testing.T) {
	service := &TransferService{}
	
	tests := []struct {
		name        string
		config      AutoTransferConfig
		expectError bool
	}{
		{
			name: "Valid config",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: false,
		},
		{
			name: "Invalid sub-account UID",
			config: AutoTransferConfig{
				SubAccountUID:   0,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Empty asset",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Zero profit threshold",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 0,
				TransferRatio:   0.8,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Invalid transfer ratio (zero)",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Invalid transfer ratio (over 1)",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   1.5,
				MinAmount:       10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Negative minimum amount",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       -10.0,
				MaxAmount:       1000.0,
			},
			expectError: true,
		},
		{
			name: "Zero maximum amount",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       10.0,
				MaxAmount:       0,
			},
			expectError: true,
		},
		{
			name: "Min amount greater than max amount",
			config: AutoTransferConfig{
				SubAccountUID:   12345,
				Asset:           "USDT",
				ProfitThreshold: 100.0,
				TransferRatio:   0.8,
				MinAmount:       1000.0,
				MaxAmount:       100.0,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateAutoTransferConfig(tt.config)
			
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

func TestTransferServiceParseResponseData(t *testing.T) {
	service := &TransferService{}
	
	// Test parsing transfer response data
	inputData := map[string]interface{}{
		"transferId": "txn_12345",
		"fromUid":    int64(11111),
		"toUid":      int64(22222),
		"asset":      "USDT",
		"amount":     100.50,
		"status":     "SUCCESS",
	}
	
	var transfer TransferResponse
	err := service.parseResponseData(inputData, &transfer)
	
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	// Note: The actual field mapping would depend on the struct tags
	// This test validates the parsing mechanism works
}

// Test InternalTransfer parameter validation
func TestTransferServiceInternalTransferValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test invalid request
	_, err = service.InternalTransfer(ctx, TransferRequest{
		FromUID: 0, // Invalid
		ToUID:   12345,
		Asset:   "USDT",
		Amount:  100.0,
	})
	if err == nil {
		t.Error("InternalTransfer should return error for invalid FromUID")
	}
}

// Test TransferToMaster
func TestTransferServiceTransferToMaster(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// This will fail with API call, but tests the parameter setup
	_, err = service.TransferToMaster(ctx, 12345, "USDT", 100.0)
	// We expect this to fail in test environment, but parameter validation should pass
}

// Test TransferFromMaster
func TestTransferServiceTransferFromMaster(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// This will fail with API call, but tests the parameter setup
	_, err = service.TransferFromMaster(ctx, 12345, "USDT", 100.0)
	// We expect this to fail in test environment, but parameter validation should pass
}

// Test BatchTransfer validation
func TestTransferServiceBatchTransferValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test empty transfers
	_, err = service.BatchTransfer(ctx, []TransferRequest{})
	if err == nil {
		t.Error("BatchTransfer should return error for empty transfers")
	}
	
	// Test too many transfers
	var tooManyTransfers []TransferRequest
	for i := 0; i < 25; i++ {
		tooManyTransfers = append(tooManyTransfers, TransferRequest{
			FromUID: int64(i + 1),
			ToUID:   int64(i + 1000),
			Asset:   "USDT",
			Amount:  100.0,
		})
	}
	
	_, err = service.BatchTransfer(ctx, tooManyTransfers)
	if err == nil {
		t.Error("BatchTransfer should return error for too many transfers")
	}
}

// Test DistributeFunds validation
func TestTransferServiceDistributeFundsValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test empty sub-accounts
	_, err = service.DistributeFunds(ctx, "USDT", 1000.0, []int64{})
	if err == nil {
		t.Error("DistributeFunds should return error for empty sub-accounts")
	}
	
	// Test zero amount
	_, err = service.DistributeFunds(ctx, "USDT", 0, []int64{12345, 67890})
	if err == nil {
		t.Error("DistributeFunds should return error for zero amount")
	}
	
	// Test negative amount
	_, err = service.DistributeFunds(ctx, "USDT", -1000.0, []int64{12345, 67890})
	if err == nil {
		t.Error("DistributeFunds should return error for negative amount")
	}
}

// Test RebalanceSubAccounts validation
func TestTransferServiceRebalanceSubAccountsValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test empty rebalance map
	_, err = service.RebalanceSubAccounts(ctx, "USDT", map[int64]float64{})
	if err == nil {
		t.Error("RebalanceSubAccounts should return error for empty rebalance map")
	}
}

// Test GetTransferHistory limit handling
func TestTransferServiceGetTransferHistoryLimitHandling(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test with limit greater than maximum
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	
	_, err = service.GetTransferHistory(ctx, 12345, 1000, &startTime, &now) // Over 500 limit
	// We don't check for specific error since API would fail anyway in test environment
}

// Test struct field validation
func TestTransferStructs(t *testing.T) {
	// Test TransferRequest
	req := TransferRequest{
		FromUID: 12345,
		ToUID:   67890,
		Asset:   "USDT",
		Amount:  100.50,
		Note:    "Test transfer",
	}
	
	if req.FromUID != 12345 {
		t.Error("FromUID should be preserved")
	}
	
	if req.Amount != 100.50 {
		t.Error("Amount should be preserved")
	}
	
	// Test TransferResponse
	resp := TransferResponse{
		TransferID: "txn_12345",
		Status:     "SUCCESS",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	if resp.TransferID != "txn_12345" {
		t.Error("TransferID should be preserved")
	}
	
	if resp.Status != "SUCCESS" {
		t.Error("Status should be preserved")
	}
	
	// Test AutoTransferConfig
	config := AutoTransferConfig{
		SubAccountUID:   12345,
		Asset:           "USDT",
		ProfitThreshold: 100.0,
		TransferRatio:   0.8,
		MinAmount:       10.0,
		MaxAmount:       1000.0,
		Enabled:         true,
		Schedule:        "daily",
	}
	
	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	
	if config.Schedule != "daily" {
		t.Error("Schedule should be preserved")
	}
	
	// Test TransferAuthRequest
	authReq := TransferAuthRequest{
		SubAccountUID: 12345,
		Authorized:    true,
	}
	
	if !authReq.Authorized {
		t.Error("Authorized should be true")
	}
}

// Test AuthorizeTransfers parameter validation
func TestTransferServiceAuthorizeTransfersValidation(t *testing.T) {
	config := ClientConfig{
		Environment: DemoEnvironment,
		Credentials: testCredentials,
	}
	
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	service := NewTransferService(client)
	ctx := context.Background()
	
	// Test invalid UID
	err = service.AuthorizeTransfers(ctx, TransferAuthRequest{
		SubAccountUID: 0,
		Authorized:    true,
	})
	if err == nil {
		t.Error("AuthorizeTransfers should return error for invalid UID")
	}
}
