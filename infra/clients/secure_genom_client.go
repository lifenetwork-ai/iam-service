package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SecureGenomClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewSecureGenomClient(endpoint string) *SecureGenomClient {
	return &SecureGenomClient{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

// StoreReencryptionKeys sends the re-encryption keys to the Secure Genom endpoint.
func (c *SecureGenomClient) StoreReencryptionKeys(
	ctx context.Context,
	header http.Header,
	payload StoreReencryptionKeysRequest,
) (*StoreReencryptionKeysResponse, error) {
	url := fmt.Sprintf("%s/api/v1/dataowner/reencryption/store-reencryption-keys", c.endpoint)

	// Serialize the payload
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

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
	var response StoreReencryptionKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &response, nil
}
