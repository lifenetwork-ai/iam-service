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
)

const (
	TokenInvalidError = -124
)

type ZaloClient struct {
	client       *http.Client
	baseURL      string
	secretKey    string
	appID        string
	accessToken  string
	refreshToken string
	oauthBaseURL string
}

type ZaloTemplateInfo struct {
	TemplateID      int    `json:"templateId"`
	TemplateName    string `json:"templateName"`
	CreatedTime     int    `json:"createdTime"`
	Status          string `json:"status"`
	TemplateQuality string `json:"templateQuality"`
}

type ZaloGetAllTemplatesResponse struct {
	Error    any                `json:"error"`
	Message  string             `json:"message"`
	Data     []ZaloTemplateInfo `json:"data"`
	Metadata struct {
		Total int `json:"total"`
	} `json:"metadata"`
}

// ZaloSendNotificationRequest represents the request payload for sending templatenotification messages
type ZaloSendNotificationRequest struct {
	Phone        string                 `json:"phone"`
	TemplateID   int                    `json:"template_id"`
	TemplateData map[string]interface{} `json:"template_data"`
}

// ZaloSendNotificationResponse represents the response from Zalo API
type ZaloSendNotificationResponse struct {
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
	Message      string `json:"message"`
	Error        int    `json:"error"`
}

func NewZaloClient(ctx context.Context, baseURL, secretKey, appID, accessToken, refreshToken string) (*ZaloClient, error) {
	return NewZaloClientWithHTTP(ctx, baseURL, secretKey, appID, accessToken, refreshToken, nil, "")
}

// NewZaloClientWithHTTP allows injecting http.Client and OAuth base URL for testing
func NewZaloClientWithHTTP(ctx context.Context, baseURL, secretKey, appID, accessToken, refreshToken string, httpClient *http.Client, oauthBaseURL string) (*ZaloClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	// Capture the current default transport to avoid cross-test interference when
	// other tests mutate http.DefaultTransport in parallel. If a custom client is
	// provided without a Transport, mirror the default at construction time.
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	if oauthBaseURL == "" {
		oauthBaseURL = "https://oauth.zaloapp.com"
	}
	return &ZaloClient{
		client:       httpClient,
		baseURL:      baseURL,
		secretKey:    secretKey,
		appID:        appID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		oauthBaseURL: oauthBaseURL,
	}, nil
}

// SendTemplateMessage sends a template message via Zalo API
func (c *ZaloClient) SendTemplateMessage(ctx context.Context, phone string, templateID int, templateData map[string]interface{}) (*ZaloSendNotificationResponse, error) {
	// Prepare the request payload
	payload := ZaloSendNotificationRequest{
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
	var zaloResp ZaloSendNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&zaloResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If Zalo responded with a non-zero error code, surface the response to the caller
	// and let higher layers decide whether it is retryable. Do not return a Go error here
	// so callers can inspect resp.Error and handle backoff semantics correctly.
	if zaloResp.Error != 0 {
		return &zaloResp, nil
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return &zaloResp, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return &zaloResp, nil
}

// SendOTP sends an OTP message using a template
func (c *ZaloClient) SendOTP(ctx context.Context, phone, otp string, templateID int) (*ZaloSendNotificationResponse, error) {
	templateData := map[string]interface{}{
		"otp": otp,
	}

	return c.SendTemplateMessage(ctx, phone, templateID, templateData)
}

// Update RefreshAccessToken to retrieve new tokens from Zalo OA API
func (c *ZaloClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*ZaloTokenRefreshResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("app_id", c.appID)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, "POST", c.oauthBaseURL+"/v4/oa/access_token", strings.NewReader(data.Encode()))
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
	if tokenResp.Error != 0 {
		return nil, fmt.Errorf("failed to refresh token: %s", tokenResp.Message)
	}

	return &tokenResp, nil
}

func (c *ZaloClient) UpdateTokens(ctx context.Context, tokenResp *ZaloTokenRefreshResponse) error {
	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken
	return nil
}

func (c *ZaloClient) GetAllTemplates(ctx context.Context) (*ZaloGetAllTemplatesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/template/all?offset=0&limit=100", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get all templates request: %w", err)
	}

	req.Header.Set("access_token", c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send get all templates request: %w", err)
	}
	defer resp.Body.Close()

	var zaloResp ZaloGetAllTemplatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&zaloResp); err != nil {
		return nil, fmt.Errorf("failed to decode get all templates response: %w", err)
	}

	return &zaloResp, nil
}

func (c *ZaloClient) GetAccessToken(ctx context.Context) string {
	return c.accessToken
}

func (c *ZaloClient) GetRefreshToken(ctx context.Context) string {
	return c.refreshToken
}
