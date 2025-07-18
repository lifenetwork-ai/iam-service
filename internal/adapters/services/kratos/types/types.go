package types

import (
	"context"

	"github.com/google/uuid"
	kratos "github.com/ory/kratos-client-go"
)

type KratosTraits = map[string]interface{}

// KratosService defines the interface for interacting with Ory Kratos
type KratosService interface {
	// Registration flow
	InitializeRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error)
	SubmitRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.RegistrationFlow, method string, traits map[string]interface{}) (*kratos.SuccessfulNativeRegistration, error)
	GetRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.RegistrationFlow, error)
	SubmitRegistrationFlowWithCode(ctx context.Context, tenantID uuid.UUID, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error)

	// Login flow
	InitializeLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error)
	SubmitLoginFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.LoginFlow, method string, identifier, password, code *string) (*kratos.SuccessfulNativeLogin, error)
	GetLoginFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.LoginFlow, error)

	// Verification flow
	InitializeVerificationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.VerificationFlow, error)
	GetVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.VerificationFlow, error)
	SubmitVerificationFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.VerificationFlow, code string) (*kratos.VerificationFlow, error)

	// Logout flow
	Logout(ctx context.Context, tenantID uuid.UUID, sessionToken string) error

	// Session management
	GetSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error)
	RevokeSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) error
	WhoAmI(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error)

	// Update identifier trait
	UpdateIdentifierTrait(ctx context.Context, tenantID uuid.UUID, identityID, identifierType, newIdentifier string) error
}

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
