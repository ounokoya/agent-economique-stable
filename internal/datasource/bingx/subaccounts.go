package bingx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SubAccountService provides sub-account management operations for BingX
type SubAccountService struct {
	client *Client
}

// SubAccountInfo represents detailed sub-account information
type SubAccountInfo struct {
	SubAccount
	APIKeys    []APIKeyInfo `json:"apiKeys,omitempty"`
	Balances   []Balance    `json:"balances,omitempty"`
	CreatedBy  string       `json:"createdBy,omitempty"`
	IsActive   bool         `json:"isActive"`
	LastLogin  time.Time    `json:"lastLogin,omitempty"`
}

// APIKeyInfo represents API key information for sub-accounts
type APIKeyInfo struct {
	APIKey      string            `json:"apiKey"`
	SecretKey   string            `json:"secretKey,omitempty"` // Only returned on creation
	Permissions []string          `json:"permissions"`
	IPWhitelist []string          `json:"ipWhitelist,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateSubAccountRequest represents parameters for creating a sub-account
type CreateSubAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Label    string `json:"label,omitempty"`
}

// CreateAPIKeyRequest represents parameters for creating API keys
type CreateAPIKeyRequest struct {
	SubAccountUID int64    `json:"subAccountUid"`
	Label         string   `json:"label,omitempty"`
	Permissions   []string `json:"permissions"`
	IPWhitelist   []string `json:"ipWhitelist,omitempty"`
}

// NewSubAccountService creates a new sub-account service
func NewSubAccountService(client *Client) *SubAccountService {
	return &SubAccountService{
		client: client,
	}
}

// CreateSubAccount creates a new sub-account
func (s *SubAccountService) CreateSubAccount(ctx context.Context, req CreateSubAccountRequest) (*SubAccountInfo, error) {
	if req.Email == "" {
		return nil, fmt.Errorf("email is required for sub-account creation")
	}

	params := map[string]string{
		"email": req.Email,
	}

	if req.Password != "" {
		params["password"] = req.Password
	}

	if req.Label != "" {
		params["label"] = req.Label
	}

	resp, err := s.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/create", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-account: %w", err)
	}

	var subAccount SubAccountInfo
	if err := s.parseResponseData(resp.Data, &subAccount); err != nil {
		return nil, fmt.Errorf("failed to parse sub-account response: %w", err)
	}

	return &subAccount, nil
}

// ListSubAccounts retrieves all sub-accounts
func (s *SubAccountService) ListSubAccounts(ctx context.Context, limit int) ([]SubAccountInfo, error) {
	params := map[string]string{}
	
	if limit > 0 {
		if limit > 200 {
			limit = 200 // BingX maximum limit
		}
		params["limit"] = strconv.Itoa(limit)
	}

	resp, err := s.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/list", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to list sub-accounts: %w", err)
	}

	var subAccounts []SubAccountInfo
	if err := s.parseResponseData(resp.Data, &subAccounts); err != nil {
		return nil, fmt.Errorf("failed to parse sub-accounts response: %w", err)
	}

	return subAccounts, nil
}

// GetSubAccountInfo retrieves detailed information about a specific sub-account
func (s *SubAccountService) GetSubAccountInfo(ctx context.Context, uid int64) (*SubAccountInfo, error) {
	if uid <= 0 {
		return nil, fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	params := map[string]string{
		"uid": strconv.FormatInt(uid, 10),
	}

	resp, err := s.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/uid", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-account info: %w", err)
	}

	var subAccount SubAccountInfo
	if err := s.parseResponseData(resp.Data, &subAccount); err != nil {
		return nil, fmt.Errorf("failed to parse sub-account info response: %w", err)
	}

	return &subAccount, nil
}

// FreezeSubAccount freezes a sub-account
func (s *SubAccountService) FreezeSubAccount(ctx context.Context, uid int64) error {
	return s.setSubAccountStatus(ctx, uid, "FREEZE")
}

// UnfreezeSubAccount unfreezes a sub-account
func (s *SubAccountService) UnfreezeSubAccount(ctx context.Context, uid int64) error {
	return s.setSubAccountStatus(ctx, uid, "UNFREEZE")
}

// setSubAccountStatus changes the status of a sub-account
func (s *SubAccountService) setSubAccountStatus(ctx context.Context, uid int64, status string) error {
	if uid <= 0 {
		return fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	params := map[string]string{
		"uid":    strconv.FormatInt(uid, 10),
		"status": status,
	}

	_, err := s.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/freeze", params, EndpointTypeAccount)
	if err != nil {
		return fmt.Errorf("failed to set sub-account status to %s: %w", status, err)
	}

	return nil
}

// CreateAPIKey creates an API key for a sub-account
func (s *SubAccountService) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyInfo, error) {
	if req.SubAccountUID <= 0 {
		return nil, fmt.Errorf("invalid sub-account UID: %d", req.SubAccountUID)
	}

	if len(req.Permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(req.SubAccountUID, 10),
	}

	if req.Label != "" {
		params["label"] = req.Label
	}

	// Convert permissions to comma-separated string
	permissionsStr := ""
	for i, perm := range req.Permissions {
		if i > 0 {
			permissionsStr += ","
		}
		permissionsStr += perm
	}
	params["permissions"] = permissionsStr

	// Convert IP whitelist to comma-separated string
	if len(req.IPWhitelist) > 0 {
		ipStr := ""
		for i, ip := range req.IPWhitelist {
			if i > 0 {
				ipStr += ","
			}
			ipStr += ip
		}
		params["ipWhitelist"] = ipStr
	}

	resp, err := s.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/apikey/create", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	var apiKey APIKeyInfo
	if err := s.parseResponseData(resp.Data, &apiKey); err != nil {
		return nil, fmt.Errorf("failed to parse API key response: %w", err)
	}

	return &apiKey, nil
}

// ListAPIKeys retrieves all API keys for a sub-account
func (s *SubAccountService) ListAPIKeys(ctx context.Context, uid int64) ([]APIKeyInfo, error) {
	if uid <= 0 {
		return nil, fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(uid, 10),
	}

	resp, err := s.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/apikey/query", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	var apiKeys []APIKeyInfo
	if err := s.parseResponseData(resp.Data, &apiKeys); err != nil {
		return nil, fmt.Errorf("failed to parse API keys response: %w", err)
	}

	return apiKeys, nil
}

// DeleteAPIKey deletes an API key for a sub-account
func (s *SubAccountService) DeleteAPIKey(ctx context.Context, uid int64, apiKey string) error {
	if uid <= 0 {
		return fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(uid, 10),
		"apiKey":        apiKey,
	}

	_, err := s.client.DoRequest(ctx, http.MethodDelete, "/openApi/api/v3/sub-account/apikey/delete", params, EndpointTypeAccount)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// ResetAPIKey resets the secret key for an API key
func (s *SubAccountService) ResetAPIKey(ctx context.Context, uid int64, apiKey string) (*APIKeyInfo, error) {
	if uid <= 0 {
		return nil, fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(uid, 10),
		"apiKey":        apiKey,
	}

	resp, err := s.client.DoRequest(ctx, http.MethodPost, "/openApi/api/v3/sub-account/apikey/reset", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to reset API key: %w", err)
	}

	var resetApiKey APIKeyInfo
	if err := s.parseResponseData(resp.Data, &resetApiKey); err != nil {
		return nil, fmt.Errorf("failed to parse reset API key response: %w", err)
	}

	return &resetApiKey, nil
}

// GetSubAccountBalances retrieves spot balances for a sub-account
func (s *SubAccountService) GetSubAccountBalances(ctx context.Context, uid int64) ([]Balance, error) {
	if uid <= 0 {
		return nil, fmt.Errorf("invalid sub-account UID: %d", uid)
	}

	params := map[string]string{
		"subAccountUid": strconv.FormatInt(uid, 10),
	}

	resp, err := s.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/spot/assets", params, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-account balances: %w", err)
	}

	var balances []Balance
	if err := s.parseResponseData(resp.Data, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse balances response: %w", err)
	}

	return balances, nil
}

// GetAllSubAccountBalances retrieves balances for all sub-accounts
func (s *SubAccountService) GetAllSubAccountBalances(ctx context.Context) (map[int64][]Balance, error) {
	resp, err := s.client.DoRequest(ctx, http.MethodGet, "/openApi/api/v3/sub-account/balance", nil, EndpointTypeAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sub-account balances: %w", err)
	}

	// Parse response as map of UID to balances
	var responseData map[string]interface{}
	if err := s.parseResponseData(resp.Data, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse all balances response: %w", err)
	}

	result := make(map[int64][]Balance)
	for uidStr, balanceData := range responseData {
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			continue // Skip invalid UIDs
		}

		var balances []Balance
		if err := s.parseResponseData(balanceData, &balances); err != nil {
			continue // Skip invalid balance data
		}

		result[uid] = balances
	}

	return result, nil
}

// ValidateSubAccountPermissions validates that permissions are valid
func (s *SubAccountService) ValidateSubAccountPermissions(permissions []string) error {
	validPermissions := map[string]bool{
		"SPOT":    true,
		"FUTURES": true,
		"MARGIN":  true,
		"READ":    true,
		"TRADE":   true,
	}

	for _, perm := range permissions {
		if !validPermissions[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	return nil
}

// GenerateSubAccountEmail generates a unique email for sub-account creation
func (s *SubAccountService) GenerateSubAccountEmail(baseEmail, suffix string) string {
	if suffix == "" {
		suffix = strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	}

	// Extract local and domain parts
	parts := strings.Split(baseEmail, "@")
	if len(parts) != 2 {
		return fmt.Sprintf("subaccount_%s@example.com", suffix)
	}

	return fmt.Sprintf("%s+%s@%s", parts[0], suffix, parts[1])
}

// parseResponseData is a generic response parser
func (s *SubAccountService) parseResponseData(data interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}

