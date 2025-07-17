package ucases

import (
	"encoding/json"
	"fmt"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// ifEmail checks if the identifierType is email and returns the value if true
func ifEmail(identifierType, val string) string {
	if identifierType == constants.IdentifierEmail.String() {
		return val
	}
	return ""
}

// ifUsername checks if the identifierType is username and returns the value if true
func ifPhone(identifierType, val string) string {
	if identifierType == constants.IdentifierPhone.String() {
		return val
	}
	return ""
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
func extractUserFromTraits(traits interface{}, fallbackEmail, fallbackPhone string) (dto.IdentityUserDTO, error) {
	traitsMap, ok := safeExtractTraits(traits)
	if !ok {
		return dto.IdentityUserDTO{}, fmt.Errorf("unable to extract traits from interface{}")
	}

	return dto.IdentityUserDTO{
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
