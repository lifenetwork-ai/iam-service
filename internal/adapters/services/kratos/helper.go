package kratos

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_types "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos/types"
	"github.com/pkg/errors"
)

// normalizeTraitsIdentifiers normalizes email and phone identifiers in traits
func normalizeTraitsIdentifiers(in map[string]interface{}) (map[string]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("traits cannot be nil")
	}

	traits := make(map[string]interface{}, len(in))
	maps.Copy(traits, in)

	// tenant (required)
	if v, ok := traits["tenant"]; !ok || v == nil || strings.TrimSpace(fmt.Sprint(v)) == "" {
		return nil, fmt.Errorf("traits.tenant is required")
	}

	// email (optional but may be identifier)
	if raw, ok := traits["email"]; ok {
		if s, ok := raw.(string); ok {
			traits["email"] = strings.ToLower(strings.TrimSpace(s))
		}
	}

	// phone_number (optional but may be identifier)
	if raw, ok := traits["phone_number"]; ok {
		if s, ok := raw.(string); ok {
			traits["phone_number"] = strings.TrimSpace(s)
		}
	}

	// Must have at least email or phone_number (per schema anyOf)
	if _, hasEmail := traits["email"]; !hasEmail {
		if _, hasPhone := traits["phone_number"]; !hasPhone {
			return nil, fmt.Errorf("either traits.email or traits.phone_number is required")
		}
	}
	return traits, nil
}

// parseKratosErrorResponse parses error response from Kratos and returns appropriate error
func parseKratosErrorResponse(resp *http.Response, defaultErr error) error {
	if resp == nil {
		return defaultErr
	}

	var kratosResp kratos_types.KratosErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err != nil {
		return defaultErr
	}

	errMsgs := kratosResp.GetErrorMessages()
	if len(errMsgs) > 0 {
		return fmt.Errorf("error occurred while submitting flow: %s", strings.Join(errMsgs, "; "))
	}

	// Handle different states if no explicit error messages
	switch kratosResp.State {
	case constants.StateSentEmail:
		return nil
	case constants.StateChooseMethod:
		return errors.New("error occurred while submitting flow")
	default:
		return defaultErr
	}
}

// extractSixDigitCode parses the first 6-digit substring it finds.
func extractSixDigitCode(s string) (string, error) {
	re := regexp.MustCompile(`\b(\d{6})\b`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return "", fmt.Errorf("no 6-digit code found")
	}
	return m[1], nil
}
