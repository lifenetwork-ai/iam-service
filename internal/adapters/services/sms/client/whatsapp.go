package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var MessagingProduct = "whatsapp"

type WhatsAppClient struct {
	AuthToken string
	PhoneID   string
	BaseURL   string
}

type MessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	}
	Messages []struct {
		ID     string `json:"id"`
		Status string `json:"message_status"`
	}
}

func NewWhatsAppClient(authToken, phoneID, baseURL string) *WhatsAppClient {
	return &WhatsAppClient{
		AuthToken: authToken,
		PhoneID:   phoneID,
		BaseURL:   baseURL,
	}
}

func (c *WhatsAppClient) messageEndpoint(phoneID string) string {
	return fmt.Sprintf("%s/%s/messages", c.BaseURL, phoneID)
}

func (c *WhatsAppClient) SendMessage(tenantName, to, message string) (*MessageResponse, error) {
	data := url.Values{}
	data.Set("messaging_product", MessagingProduct)
	data.Set("recipient_type", "individual")
	data.Set("to", to)
	data.Set("type", "text")

	// Fix: text parameter must be a JSON object with "body" field
	textObj := map[string]string{"body": message}
	textJSON, err := json.Marshal(textObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal text object: %w", err)
	}
	data.Set("text", string(textJSON))

	req, err := http.NewRequest("POST", c.messageEndpoint(c.PhoneID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read the response body for error details
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	var messageResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messageResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &messageResp, nil
}

func (c *WhatsAppClient) RefreshAccessToken(ctx context.Context) error {
	return nil
}
