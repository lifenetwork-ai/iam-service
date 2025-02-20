package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LifeAIClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewLifeAIClient(endpoint string) *LifeAIClient {
	return &LifeAIClient{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

func (c *LifeAIClient) GetProfile(
	ctx context.Context,
	authHeader string,
) (*LifeAIProfile, error) {
	url := fmt.Sprintf("%s/api/v1/user-profile/", c.endpoint)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Version-Management", "1.0.20|web")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, error: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response body
	var response LifeAIProfile
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &response, nil
}
