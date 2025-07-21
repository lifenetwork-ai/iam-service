package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type KratosTraits = map[string]interface{}

// KratosResponse represents the structured response from Kratos API
type KratosResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	ExpiresAt    string `json:"expires_at"`
	IssuedAt     string `json:"issued_at"`
	RequestURL   string `json:"request_url"`
	Active       string `json:"active"`
	State        string `json:"state"`
	SessionToken string `json:"session_token"`
	UI           struct {
		Action   string `json:"action"`
		Method   string `json:"method"`
		Messages []struct {
			ID   int64  `json:"id"`
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"messages"`
		Nodes []struct {
			Type       string `json:"type"`
			Group      string `json:"group"`
			Attributes struct {
				Name     string `json:"name"`
				Type     string `json:"type"`
				Value    string `json:"value"`
				Required bool   `json:"required"`
			} `json:"attributes"`
		} `json:"nodes"`
	} `json:"ui"`
}

// KratosFlowResponse represents the structured response from Kratos API
type KratosFlowResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	UI   struct {
		Nodes []struct {
			Type       string `json:"type"`
			Group      string `json:"group"`
			Attributes struct {
				Name  string      `json:"name"`
				Value interface{} `json:"value"`
				Type  string      `json:"type"`
			} `json:"attributes"`
		} `json:"nodes"`
	} `json:"ui"`
}

// FromHttpResp parses the response body into a KratosFlowResponse
func FromHttpResp(resp *http.Response) (*KratosFlowResponse, error) {
	var flowData KratosFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flowData); err != nil {
		return nil, fmt.Errorf("failed to parse flow data: %w", err)
	}
	return &flowData, nil
}

// GetTraits extracts traits from the flow response
func (r *KratosFlowResponse) GetTraits() map[string]any {
	// Extract traits from nodes
	traits := make(map[string]any)
	for _, node := range r.UI.Nodes {
		if node.Group != "default" || node.Type != "input" {
			continue
		}

		name := node.Attributes.Name
		if !strings.HasPrefix(name, "traits.") {
			continue
		}

		traitKey := strings.TrimPrefix(name, "traits.")
		if node.Attributes.Value != nil {
			// Check for empty string values
			if strVal, ok := node.Attributes.Value.(string); ok {
				if strVal != "" {
					traits[traitKey] = strVal
				}
			} else {
				// For non-string values, add them as is
				traits[traitKey] = node.Attributes.Value
			}
		}

	}

	return traits
}

// KratosErrorResponse represents the error response structure from Kratos API
type KratosErrorResponse struct {
	State string `json:"state"`
	UI    struct {
		Messages []struct {
			ID   int64  `json:"id"`
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"messages"`
		Nodes []struct {
			Messages []struct {
				ID      int64  `json:"id"`
				Text    string `json:"text"`
				Type    string `json:"type"`
				Context struct {
					Reason string `json:"reason"`
				} `json:"context"`
			} `json:"messages"`
		} `json:"nodes"`
	} `json:"ui"`
}

// GetValidationErrors extracts validation error messages from the response
func (r *KratosErrorResponse) GetValidationErrors() []string {
	var validationErrors []string

	// Check messages in UI nodes
	for _, node := range r.UI.Nodes {
		for _, msg := range node.Messages {
			if msg.Type == "error" {
				validationErrors = append(validationErrors, msg.Text)
			}
		}
	}

	return validationErrors
}

// GetErrorMessages returns all error messages from both UI.Messages and validation errors
func (r *KratosErrorResponse) GetErrorMessages() []string {
	var errMsgs []string

	// Collect error messages from top-level UI.Messages
	for _, msg := range r.UI.Messages {
		if msg.Type == "error" && msg.Text != "" {
			errMsgs = append(errMsgs, msg.Text)
		}
	}

	// Also collect from validation errors in nodes
	for _, ve := range r.GetValidationErrors() {
		if ve != "" {
			errMsgs = append(errMsgs, ve)
		}
	}

	return errMsgs
}
