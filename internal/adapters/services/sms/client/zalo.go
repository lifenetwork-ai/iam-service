package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type ZaloClient struct {
	client       *http.Client
	tokenRepo    domainrepo.ZaloTokenRepository
	baseURL      string
	secretKey    string
	appID        string
	accessToken  string
	refreshToken string
}

// ZaloTemplateRequest represents the request payload for sending template messages
type ZaloTemplateRequest struct {
	Phone        string                 `json:"phone"`
	TemplateID   int                    `json:"template_id"`
	TemplateData map[string]interface{} `json:"template_data"`
}

// ZaloTemplateResponse represents the response from Zalo API
type ZaloTemplateResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    struct {
		SentTime    string `json:"sent_time"`
		SendingMode string `json:"sending_mode"`
		Quota       struct {
			RemainingQuota string `json:"remainingQuota"`
			DailyQuota     string `json:"dailyQuota"`
		} `json:"quota"`
		MsgID string `json:"msg_id"`
	} `json:"data"`
}

// ZaloError represents an error response from Zalo API
type ZaloError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

// ZaloTokenRefreshRequest represents the request payload for refreshing tokens
type ZaloTokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	AppID        string `json:"app_id"`
	GrantType    string `json:"grant_type"`
}

// ZaloTokenRefreshResponse represents the response from Zalo OAuth API
type ZaloTokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

func NewZaloClient(ctx context.Context, baseURL, secretKey, appID, accessToken, refreshToken string) (*ZaloClient, error) {
	return &ZaloClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:      baseURL,
		secretKey:    secretKey,
		appID:        appID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

// SendTemplateMessage sends a template message via Zalo API
func (c *ZaloClient) SendTemplateMessage(ctx context.Context, phone string, templateID int, templateData map[string]interface{}) (*ZaloTemplateResponse, error) {
	// Prepare the request payload
	payload := ZaloTemplateRequest{
		Phone:        phone,
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	// Marshal the payload to JSON
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/message/template", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("access_token", c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var zaloResp ZaloTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&zaloResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if zaloResp.Error != 0 {
		return &zaloResp, fmt.Errorf("zalo API error: %d - %s", zaloResp.Error, zaloResp.Message)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return &zaloResp, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return &zaloResp, nil
}

// SendOTP sends an OTP message using a template
func (c *ZaloClient) SendOTP(ctx context.Context, phone, otp string, templateID int) (*ZaloTemplateResponse, error) {
	templateData := map[string]interface{}{
		"otp": otp,
	}

	return c.SendTemplateMessage(ctx, phone, templateID, templateData)
}

// Update RefreshAccessToken to persist new tokens in DB
func (c *ZaloClient) RefreshAccessToken(ctx context.Context) (*ZaloTokenRefreshResponse, error) {
	data := url.Values{}
	data.Set("refresh_token", c.refreshToken)
	data.Set("app_id", c.appID)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth.zaloapp.com/v4/oa/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}

	req.Header.Set("secret_key", c.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send refresh token request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp ZaloTokenRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error during token refresh: %s", resp.Status)
	}

	return &tokenResp, nil
}

func (c *ZaloClient) UpdateTokens(ctx context.Context, tokenResp *ZaloTokenRefreshResponse) error {
	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken
	return nil
}
