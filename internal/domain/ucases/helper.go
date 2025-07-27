package ucases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// extractSessionToken extracts and validates the session token from context
func extractSessionToken(ctx context.Context) (string, *domainerrors.DomainError) {
	sessionTokenVal := ctx.Value(constants.SessionTokenKey)
	if sessionTokenVal == nil {
		return "", domainerrors.NewUnauthorizedError("MSG_UNAUTHORIZED", "Unauthorized").WithDetails([]interface{}{
			map[string]string{"field": "session_token", "error": "Session token not found"},
		})
	}

	sessionToken, ok := sessionTokenVal.(string)
	if !ok || sessionToken == "" {
		return "", domainerrors.NewUnauthorizedError("MSG_UNAUTHORIZED", "Unauthorized").WithDetails([]interface{}{
			map[string]string{"field": "session_token", "error": "Invalid session token format"},
		})
	}

	return sessionToken, nil
}

// safeExtractTraits safely converts interface{} to map[string]interface{}
// Returns the map and a boolean indicating success
func safeExtractTraits(traits interface{}) (map[string]interface{}, bool) {
	if traits == nil {
		return make(map[string]interface{}), false
	}

	// Direct type assertion (most common case)
	if traitsMap, ok := traits.(map[string]interface{}); ok {
		return traitsMap, true
	}

	// Fallback: JSON marshal/unmarshal for complex cases
	jsonBytes, err := json.Marshal(traits)
	if err != nil {
		logger.GetLogger().Errorf("Failed to marshal traits: %v", err)
		return make(map[string]interface{}), false
	}

	var traitsMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &traitsMap); err != nil {
		logger.GetLogger().Errorf("Failed to unmarshal traits: %v", err)
		return make(map[string]interface{}), false
	}

	return traitsMap, true
}

// extractUserFromTraits safely extracts user data from traits
func extractUserFromTraits(traits interface{}, fallbackEmail, fallbackPhone string) (types.IdentityUserResponse, error) {
	traitsMap, ok := safeExtractTraits(traits)
	if !ok {
		return types.IdentityUserResponse{}, fmt.Errorf("unable to extract traits from interface{}")
	}

	return types.IdentityUserResponse{
		UserName: extractStringFromTraits(traitsMap, constants.IdentifierUsername.String(), ""),
		Email:    extractStringFromTraits(traitsMap, constants.IdentifierEmail.String(), fallbackEmail),
		Phone:    extractStringFromTraits(traitsMap, constants.IdentifierPhone.String(), fallbackPhone),
		Tenant:   extractStringFromTraits(traitsMap, constants.IdentifierTenant.String(), ""),
	}, nil
}

// extractStringFromTraits extracts a string value from traits map
// If the value is a pointer to a string, it dereferences it
// If the value is nil, it returns the default value
// If the value is not a string, it returns the default value
func extractStringFromTraits(traits map[string]interface{}, key, defaultValue string) string {
	if traits == nil {
		return defaultValue
	}

	value, exists := traits[key]
	if !exists {
		return defaultValue
	}

	// Handle different types that might be stored
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return defaultValue
	case nil:
		return defaultValue
	default:
		// Convert other types to string as fallback
		return fmt.Sprintf("%v", v)
	}
}

// extractTenantNameFromBody extracts the tenant name from the message body
func extractTenantNameFromBody(body string) string {
	// Eg: [genetica] Your login code is: 123456...
	if len(body) < 3 || body[0] != '[' {
		return ""
	}
	end := strings.Index(body, "]")
	if end <= 1 {
		return ""
	}
	return strings.ToLower(body[1:end]) // normalize tenant name
}

// TODO: refactor this later
// mockWebhookURL is the URL to send mock messages to
var mockWebhookURL = os.Getenv("MOCK_WEBHOOK_URL")

type otpMessage struct {
	Body string `json:"Body"`
	To   string `json:"To"`
}

// sendViaProvider simulates sending OTP via the specified channel.
func sendViaProvider(ctx context.Context, channel, receiver, message string) error {
	// Check if mock webhook URL is configured
	if mockWebhookURL == "" {
		logger.GetLogger().Warnf("MOCK_WEBHOOK_URL environment variable is not set, skipping mock message delivery to %s", receiver)
		return fmt.Errorf("mock webhook URL not configured")
	}

	logger.GetLogger().Infof("Sending mock message to %s via %s: %s", receiver, channel, message)
	payload := otpMessage{
		Body: message,
		To:   receiver,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal mock payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", mockWebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create mock HTTP request: %w", err)
	}
	req.Header.Set(constants.HeaderKeyContentType, constants.HeaderContentTypeJson)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send mock HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("mock webhook returned status: %s", resp.Status)
	}

	return nil
}
