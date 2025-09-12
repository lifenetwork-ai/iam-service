package kratos

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
	kratos_types "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos/types"
	"github.com/pkg/errors"
)

// normalizeTraitsIdentifiers normalizes email and phone identifiers in traits
func normalizeTraitsIdentifiers(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	traits := make(map[string]interface{}, len(in))
	maps.Copy(traits, in)

	// email (singular)
	if raw, ok := traits["email"]; ok {
		if s, ok := raw.(string); ok {
			traits["email"] = strings.ToLower(strings.TrimSpace(s))
		}
	}

	// phone (singular)
	if raw, ok := traits["phone"]; ok {
		if s, ok := raw.(string); ok {
			traits["phone"] = strings.TrimSpace(s)
		}
	}

	return traits
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
