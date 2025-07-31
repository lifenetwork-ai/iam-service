package ucases

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

// extractTenantNameFromBody extracts the tenant name from the first [] in the message body.
// E.g. [LIFE AI]
func extractTenantNameFromBody(body string) string {
	// Regex to find [TENANT]
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(body)

	if len(matches) < 2 {
		return ""
	}

	return normalizeTenant(matches[1])
}

// normalizeTenant standardizes tenant display names.
func normalizeTenant(t string) string {
	switch strings.ToLower(t) {
	case "life_ai", "lifeai", "life ai":
		return constants.TenantLifeAI
	case "genetica":
		return constants.TenantGenetica
	default:
		return t
	}
}

// extractOTPFromBody extracts the OTP from the body
// TODO: make this more robust to handle different OTP formats
func extractOTPFromBody(body string) string {
	// Regex to find OTP
	re := regexp.MustCompile(`\d{6}`)
	matches := re.FindStringSubmatch(body)
	return matches[0]
}
