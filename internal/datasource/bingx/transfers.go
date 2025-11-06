package bingx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// TransferService provides internal transfer operations for BingX
type TransferService struct {
	client *Client
}

// TransferRequest represents parameters for internal transfers
type TransferRequest struct {
	FromUID int64   `json:"fromUid"`
	ToUID   int64   `json:"toUid"`
	Asset   string  `json:"asset"`
	Amount  float64 `json:"amount"`
	Note    string  `json:"note,omitempty"`
}

// TransferResponse represents the response from a transfer operation
type TransferResponse struct {
	TransferID string    `json:"transferId"`
	FromUID    int64     `json:"fromUid"`
	ToUID      int64     `json:"toUid"`
	Asset      string    `json:"asset"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	Note       string    `json:"note,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TransferHistory represents historical transfer data
type TransferHistory struct {
	Transfers []TransferResponse `json:"transfers"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
}

// TransferAuthRequest represents transfer authorization settings
type TransferAuthRequest struct {
	SubAccountUID int64 `json:"subAccountUid"`
	Authorized    bool  `json:"authorized"`
}

// AutoTransferConfig represents automatic transfer configuration
type AutoTransferConfig struct {
	SubAccountUID   int64   `json:"subAccountUid"`
	Asset           string  `json:"asset"`
	ProfitThreshold float64 `json:"profitThreshold"`
	TransferRatio   float64 `json:"transferRatio"`
	MinAmount       float64 `json:"minAmount"`
	MaxAmount       float64 `json:"maxAmount"`
	Enabled         bool    `json:"enabled"`
	Schedule        string  `json:"schedule,omitempty"` // e.g., "daily", "weekly"
}

// NewTransferService creates a new transfer service
func NewTransferService(client *Client) *TransferService {
	return &TransferService{
		client: client,
	}
}

// InternalTransfer performs an internal transfer between accounts
func (t *TransferService) InternalTransfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	if err := t.validateTransferRequest(req); err != nil {
		return nil, fmt.Errorf("invalid transfer request: %w", err)
	}

	params := map[string]string{
		"fromUid": strconv.FormatInt(req.FromUID, 10),
		"toUid":   strconv.FormatInt(req.ToUID, 10),
		"asset":   req.Asset,
		"amount":  strconv.FormatFloat(req.Amount, 'f', -1, 64),
	}

	if req.Note != "" {
		params["note"] = req.Note
	}

	resp, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/transfer/internal", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to execute internal transfer: %w", err)
	}

	var transfer TransferResponse
	if err := t.parseResponseData(resp.Data, &transfer); err != nil {
		return nil, fmt.Errorf("failed to parse transfer response: %w", err)
	}

	return &transfer, nil
}

// TransferToMaster transfers funds from sub-account to master account
func (t *TransferService) TransferToMaster(ctx context.Context, fromUID int64, asset string, amount float64) (*TransferResponse, error) {
	// Master account typically has UID 0 or specific master UID
	return t.InternalTransfer(ctx, TransferRequest{
		FromUID: fromUID,
		ToUID:   0, // Assuming master account UID is 0
		Asset:   asset,
		Amount:  amount,
		Note:    "Transfer to master account",
	})
}

// TransferFromMaster transfers funds from master account to sub-account
func (t *TransferService) TransferFromMaster(ctx context.Context, toUID int64, asset string, amount float64) (*TransferResponse, error) {
	return t.InternalTransfer(ctx, TransferRequest{
		FromUID: 0, // Assuming master account UID is 0
		ToUID:   toUID,
		Asset:   asset,
		Amount:  amount,
		Note:    "Transfer from master account",
	})
}

// GetTransferHistory retrieves transfer history for an account
func (t *TransferService) GetTransferHistory(ctx context.Context, uid int64, limit int, startTime, endTime *time.Time) (*TransferHistory, error) {
	params := map[string]string{}

	if uid > 0 {
		params["uid"] = strconv.FormatInt(uid, 10)
	}

	if limit > 0 {
		if limit > 500 {
			limit = 500 // BingX maximum limit
		}
		params["limit"] = strconv.Itoa(limit)
	}

	if startTime != nil {
		params["startTime"] = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	if endTime != nil {
		params["endTime"] = strconv.FormatInt(endTime.UnixMilli(), 10)
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/transfer/history", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer history: %w", err)
	}

	var history TransferHistory
	if err := t.parseResponseData(resp.Data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse transfer history response: %w", err)
	}

	return &history, nil
}

// AuthorizeTransfers enables or disables transfer permissions for a sub-account
func (t *TransferService) AuthorizeTransfers(ctx context.Context, req TransferAuthRequest) error {
	if req.SubAccountUID <= 0 {
		return fmt.Errorf("invalid sub-account UID: %d", req.SubAccountUID)
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(req.SubAccountUID, 10),
		"authorized":    strconv.FormatBool(req.Authorized),
	}

	_, err := t.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/transfer/authorize", params, EndpointTypeAccount)
	if err != nil {
		return fmt.Errorf("failed to authorize transfers: %w", err)
	}

	return nil
}

// BatchTransfer performs multiple transfers in a single operation
func (t *TransferService) BatchTransfer(ctx context.Context, transfers []TransferRequest) ([]TransferResponse, error) {
	if len(transfers) == 0 {
		return nil, fmt.Errorf("no transfers specified")
	}

	if len(transfers) > 20 {
		return nil, fmt.Errorf("too many transfers in batch (max 20)")
	}

	var results []TransferResponse
	var errors []error

	for i, transfer := range transfers {
		result, err := t.InternalTransfer(ctx, transfer)
		if err != nil {
			errors = append(errors, fmt.Errorf("transfer %d failed: %w", i, err))
			continue
		}
		results = append(results, *result)
	}

	if len(errors) > 0 {
		// Return partial results with error
		return results, fmt.Errorf("batch transfer completed with %d errors", len(errors))
	}

	return results, nil
}

// AutoProfitTransfer automatically transfers profits from sub-account to master
func (t *TransferService) AutoProfitTransfer(ctx context.Context, config AutoTransferConfig) (*TransferResponse, error) {
	if err := t.validateAutoTransferConfig(config); err != nil {
		return nil, fmt.Errorf("invalid auto transfer config: %w", err)
	}

	// Get sub-account balances
	balances, err := t.getSubAccountBalance(ctx, config.SubAccountUID, config.Asset)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-account balance: %w", err)
	}

	availableBalance := balances.Free
	if availableBalance < config.ProfitThreshold {
		return nil, fmt.Errorf("balance %.8f below profit threshold %.8f", availableBalance, config.ProfitThreshold)
	}

	// Calculate transfer amount
	transferAmount := availableBalance * config.TransferRatio
	if transferAmount > config.MaxAmount {
		transferAmount = config.MaxAmount
	}
	if transferAmount < config.MinAmount {
		return nil, fmt.Errorf("calculated transfer amount %.8f below minimum %.8f", transferAmount, config.MinAmount)
	}

	// Execute transfer
	return t.TransferToMaster(ctx, config.SubAccountUID, config.Asset, transferAmount)
}

// DistributeFunds distributes funds from master account to multiple sub-accounts
func (t *TransferService) DistributeFunds(ctx context.Context, asset string, totalAmount float64, subAccountUIDs []int64) ([]TransferResponse, error) {
	if len(subAccountUIDs) == 0 {
		return nil, fmt.Errorf("no sub-accounts specified")
	}

	if totalAmount <= 0 {
		return nil, fmt.Errorf("invalid total amount: %f", totalAmount)
	}

	// Calculate amount per sub-account
	amountPerAccount := totalAmount / float64(len(subAccountUIDs))

	var transfers []TransferRequest
	for _, uid := range subAccountUIDs {
		transfers = append(transfers, TransferRequest{
			FromUID: 0, // Master account
			ToUID:   uid,
			Asset:   asset,
			Amount:  amountPerAccount,
			Note:    fmt.Sprintf("Distribution of %s", asset),
		})
	}

	return t.BatchTransfer(ctx, transfers)
}

// RebalanceSubAccounts rebalances funds between sub-accounts based on performance
func (t *TransferService) RebalanceSubAccounts(ctx context.Context, asset string, rebalanceMap map[int64]float64) ([]TransferResponse, error) {
	if len(rebalanceMap) == 0 {
		return nil, fmt.Errorf("no rebalance targets specified")
	}

	var transfers []TransferRequest

	// Create transfers based on rebalance map
	// Positive amounts = receive funds, negative = send funds
	var senders []int64
	var receivers []int64

	for uid, amount := range rebalanceMap {
		if amount > 0 {
			receivers = append(receivers, uid)
		} else if amount < 0 {
			senders = append(senders, uid)
		}
	}

	// Simple rebalancing: transfer from senders to receivers
	for i, sender := range senders {
		if i >= len(receivers) {
			break
		}
		receiver := receivers[i]
		amount := -rebalanceMap[sender] // Convert negative to positive

		transfers = append(transfers, TransferRequest{
			FromUID: sender,
			ToUID:   receiver,
			Asset:   asset,
			Amount:  amount,
			Note:    "Rebalancing transfer",
		})
	}

	return t.BatchTransfer(ctx, transfers)
}

// validateTransferRequest validates transfer request parameters
func (t *TransferService) validateTransferRequest(req TransferRequest) error {
	if req.FromUID <= 0 {
		return fmt.Errorf("invalid from UID: %d", req.FromUID)
	}

	if req.ToUID <= 0 {
		return fmt.Errorf("invalid to UID: %d", req.ToUID)
	}

	if req.FromUID == req.ToUID {
		return fmt.Errorf("from UID and to UID cannot be the same")
	}

	if req.Asset == "" {
		return fmt.Errorf("asset is required")
	}

	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive: %f", req.Amount)
	}

	return nil
}

// validateAutoTransferConfig validates auto transfer configuration
func (t *TransferService) validateAutoTransferConfig(config AutoTransferConfig) error {
	if config.SubAccountUID <= 0 {
		return fmt.Errorf("invalid sub-account UID: %d", config.SubAccountUID)
	}

	if config.Asset == "" {
		return fmt.Errorf("asset is required")
	}

	if config.ProfitThreshold <= 0 {
		return fmt.Errorf("profit threshold must be positive: %f", config.ProfitThreshold)
	}

	if config.TransferRatio <= 0 || config.TransferRatio > 1 {
		return fmt.Errorf("transfer ratio must be between 0 and 1: %f", config.TransferRatio)
	}

	if config.MinAmount < 0 {
		return fmt.Errorf("minimum amount cannot be negative: %f", config.MinAmount)
	}

	if config.MaxAmount <= 0 {
		return fmt.Errorf("maximum amount must be positive: %f", config.MaxAmount)
	}

	if config.MinAmount > config.MaxAmount {
		return fmt.Errorf("minimum amount cannot be greater than maximum amount")
	}

	return nil
}

// getSubAccountBalance retrieves balance for a specific asset in a sub-account
func (t *TransferService) getSubAccountBalance(ctx context.Context, uid int64, asset string) (*Balance, error) {
	// This would typically call the sub-account service
	// For now, we'll use a placeholder implementation
	params := map[string]string{
		"subAccountUid": strconv.FormatInt(uid, 10),
		"asset":         asset,
	}

	resp, err := t.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/spot/assets", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-account balance: %w", err)
	}

	var balances []Balance
	if err := t.parseResponseData(resp.Data, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse balance response: %w", err)
	}

	// Find the specific asset balance
	for _, balance := range balances {
		if balance.Asset == asset {
			return &balance, nil
		}
	}

	return nil, fmt.Errorf("asset %s not found in sub-account balances", asset)
}

// parseResponseData is a generic response parser
func (t *TransferService) parseResponseData(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}

// GetTransferStats returns transfer statistics for monitoring
func (t *TransferService) GetTransferStats(ctx context.Context, uid int64, period time.Duration) (map[string]interface{}, error) {
	endTime := time.Now()
	startTime := endTime.Add(-period)

	history, err := t.GetTransferHistory(ctx, uid, 1000, &startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer history: %w", err)
	}

	stats := map[string]interface{}{
		"total_transfers":  len(history.Transfers),
		"total_amount":     0.0,
		"successful":       0,
		"failed":           0,
		"assets":           make(map[string]float64),
		"start_time":       startTime,
		"end_time":         endTime,
	}

	var totalAmount float64
	successful := 0
	failed := 0
	assetAmounts := make(map[string]float64)

	for _, transfer := range history.Transfers {
		totalAmount += transfer.Amount

		if transfer.Status == "SUCCESS" || transfer.Status == "COMPLETED" {
			successful++
		} else {
			failed++
		}

		assetAmounts[transfer.Asset] += transfer.Amount
	}

	stats["total_amount"] = totalAmount
	stats["successful"] = successful
	stats["failed"] = failed
	stats["assets"] = assetAmounts

	return stats, nil
}
