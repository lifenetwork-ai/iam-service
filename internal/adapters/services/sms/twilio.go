package sms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TwilioClient struct {
	AccountSID string
	AuthToken  string
	BaseURL    string
}

type SMSResponse struct {
	SID          string `json:"sid"`
	Status       any    `json:"status"`
	Body         string `json:"body"`
	From         string `json:"from"`
	To           string `json:"to"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewTwilioClient creates a new Twilio client
func NewTwilioClient(accountSID, authToken string) *TwilioClient {
	return &TwilioClient{
		AccountSID: accountSID,
		AuthToken:  authToken,
		BaseURL:    "https://api.twilio.com/2010-04-01",
	}
}

// SendSMS sends an SMS message via Twilio
func (c *TwilioClient) SendSMS(_, from, to, message string) (*SMSResponse, error) {
	// Prepare the API endpoint
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", c.BaseURL, c.AccountSID)

	// Prepare form data
	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	data.Set("Body", message)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.AccountSID, c.AuthToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusCreated {
		fmt.Println(resp)
		return &smsResp, fmt.Errorf("twilio API error: %s - %s", smsResp.ErrorCode, smsResp.ErrorMessage)
	}

	return &smsResp, nil
}
