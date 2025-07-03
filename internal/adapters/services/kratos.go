package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lifenetwork-ai/iam-service/conf"
	kratos "github.com/ory/kratos-client-go"
)

// KratosService defines the interface for interacting with Ory Kratos
type KratosService interface {
	// Registration flow
	InitializeRegistrationFlow(ctx context.Context) (*kratos.RegistrationFlow, error)
	SubmitRegistrationFlow(ctx context.Context, flow *kratos.RegistrationFlow, method string, traits map[string]interface{}) (*kratos.SuccessfulNativeRegistration, error)
	GetRegistrationFlow(ctx context.Context, flowID string) (*kratos.RegistrationFlow, error)
	SubmitRegistrationFlowWithCode(ctx context.Context, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error)

	// Login flow
	InitializeLoginFlow(ctx context.Context) (*kratos.LoginFlow, error)
	SubmitLoginFlow(ctx context.Context, flow *kratos.LoginFlow, method string, identifier *string, password *string) (*kratos.SuccessfulNativeLogin, error)

	// Verification flow
	InitializeVerificationFlow(ctx context.Context) (*kratos.VerificationFlow, error)
	GetVerificationFlow(ctx context.Context, flowID string) (*kratos.VerificationFlow, error)
	SubmitVerificationFlow(ctx context.Context, flow *kratos.VerificationFlow, code string) (*kratos.VerificationFlow, error)

	// Session management
	GetSession(ctx context.Context, sessionToken string) (*kratos.Session, error)
	RevokeSession(ctx context.Context, sessionToken string) error
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

type kratosServiceImpl struct {
	client *kratos.APIClient
}

// NewKratosService creates a new instance of KratosService
func NewKratosService() KratosService {
	config := conf.GetKratosConfig()
	configuration := kratos.NewConfiguration()
	configuration.Servers = []kratos.ServerConfiguration{
		{
			URL: config.PublicEndpoint,
		},
	}

	client := kratos.NewAPIClient(configuration)
	return &kratosServiceImpl{
		client: client,
	}
}

func (k *kratosServiceImpl) InitializeRegistrationFlow(ctx context.Context) (*kratos.RegistrationFlow, error) {
	flow, _, err := k.client.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registration flow: %w", err)
	}
	return flow, nil
}

func (k *kratosServiceImpl) SubmitRegistrationFlow(
	ctx context.Context,
	flow *kratos.RegistrationFlow,
	method string,
	traits map[string]any,
) (*kratos.SuccessfulNativeRegistration, error) {
	submitFlow := k.client.FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow.Id)

	var body kratos.UpdateRegistrationFlowBody
	switch method {
	case "code":
		body.UpdateRegistrationFlowWithCodeMethod = &kratos.UpdateRegistrationFlowWithCodeMethod{
			Method: "code",
			Traits: traits,
		}

		result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
		if err != nil {
			if resp != nil && resp.StatusCode == 400 {
				var kratosResp KratosResponse
				if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err == nil {
					// Check for validation errors
					var validationErrors []string
					for _, msg := range kratosResp.UI.Messages {
						if msg.Type == "error" {
							validationErrors = append(validationErrors, msg.Text)
						}
					}

					if len(validationErrors) > 0 {
						return nil, fmt.Errorf("validation errors: %v", validationErrors)
					}

					// If state is sent_email, this means the code was sent successfully
					if kratosResp.State == "sent_email" {
						// Return empty successful registration since code was sent
						return &kratos.SuccessfulNativeRegistration{}, nil
					}
				}
			}
			return nil, fmt.Errorf("registration failed: %w", err)
		}

		return result, nil
	case "password":
		body.UpdateRegistrationFlowWithPasswordMethod = &kratos.UpdateRegistrationFlowWithPasswordMethod{
			Method: "password",
			Traits: traits,
		}
		result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
		if err != nil {
			// Check if this is a 400 response
			if resp != nil && resp.StatusCode == 400 {
				// Parse the response to check the flow state
				var kratosResp struct {
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
				if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err == nil {
					// Check for validation errors
					var validationErrors []string
					for _, node := range kratosResp.UI.Nodes {
						for _, msg := range node.Messages {
							if msg.Type == "error" {
								validationErrors = append(validationErrors, msg.Text)
							}
						}
					}

					// Return validation errors if any
					if len(validationErrors) > 0 {
						return nil, fmt.Errorf("validation errors: %v", validationErrors)
					}

					// Handle different states
					switch kratosResp.State {
					case "sent_email":
						// This is a successful response
						return &kratos.SuccessfulNativeRegistration{}, nil
					case "choose_method":
						// This means we need to select a registration method
						return nil, fmt.Errorf("registration method selection required")
					default:
						// For any other state, return the original error
						return nil, err
					}
				}
			}
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (k *kratosServiceImpl) SubmitRegistrationFlowWithCode(ctx context.Context, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error) {
	submitFlow := k.client.FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow.Id)

	// Get the flow again to extract traits
	_, resp, err := k.client.FrontendAPI.GetRegistrationFlow(ctx).Id(flow.Id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get registration flow: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from registration flow")
	}
	defer resp.Body.Close()

	// Parse the response to extract traits
	var flowData KratosFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flowData); err != nil {
		return nil, fmt.Errorf("failed to parse flow data: %w", err)
	}

	// Extract traits from nodes
	traits := make(map[string]interface{})
	for _, node := range flowData.UI.Nodes {
		if node.Group == "default" && node.Type == "input" {
			name := node.Attributes.Name
			if strings.HasPrefix(name, "traits.") {
				traitKey := strings.TrimPrefix(name, "traits.")
				traits[traitKey] = node.Attributes.Value
			}
		}
	}

	// Ensure we have the required traits
	if len(traits) == 0 {
		return nil, fmt.Errorf("no traits found in registration flow")
	}

	body := kratos.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithCodeMethod: &kratos.UpdateRegistrationFlowWithCodeMethod{
			Method: "code",
			Code:   &code,
			Traits: traits,
		},
	}
	result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to submit registration flow with code: %w", err)
	}
	return result, nil
}

func (k *kratosServiceImpl) GetRegistrationFlow(ctx context.Context, flowID string) (*kratos.RegistrationFlow, error) {
	flow, _, err := k.client.FrontendAPI.GetRegistrationFlow(ctx).Id(flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get registration flow: %w", err)
	}
	return flow, nil
}

func (k *kratosServiceImpl) InitializeLoginFlow(ctx context.Context) (*kratos.LoginFlow, error) {
	flow, _, err := k.client.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize login flow: %w", err)
	}
	return flow, nil
}

func (k *kratosServiceImpl) SubmitLoginFlow(
	ctx context.Context,
	flow *kratos.LoginFlow,
	method string,
	identifier *string,
	password *string,
) (*kratos.SuccessfulNativeLogin, error) {
	submitFlow := k.client.FrontendAPI.UpdateLoginFlow(ctx).Flow(flow.Id)

	var body kratos.UpdateLoginFlowBody
	switch method {
	case "code":
		body.UpdateLoginFlowWithCodeMethod = &kratos.UpdateLoginFlowWithCodeMethod{
			Method:     "code",
			Identifier: identifier,
		}
	case "password":
		body.UpdateLoginFlowWithPasswordMethod = &kratos.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Password:   *password,
			Identifier: *identifier,
		}
	default:
		return nil, fmt.Errorf("unsupported login method: %s", method)
	}

	result, resp, err := submitFlow.UpdateLoginFlowBody(body).Execute()
	if err != nil {
		// Check if this is a 400 response
		if resp != nil && resp.StatusCode == 400 {
			// Parse the response to check the flow state
			var kratosResp struct {
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
			if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err == nil {
				// Check for validation errors
				var validationErrors []string
				for _, node := range kratosResp.UI.Nodes {
					for _, msg := range node.Messages {
						if msg.Type == "error" {
							validationErrors = append(validationErrors, msg.Text)
						}
					}
				}

				// Return validation errors if any
				if len(validationErrors) > 0 {
					return nil, fmt.Errorf("validation errors: %v", validationErrors)
				}

				// Handle different states
				switch kratosResp.State {
				case "sent_email":
					// This is a successful response
					return &kratos.SuccessfulNativeLogin{}, nil
				case "choose_method":
					// This means we need to select a login method
					return nil, fmt.Errorf("login method selection required")
				default:
					// For any other state, return the original error
					return nil, err
				}
			}
		}
		return nil, fmt.Errorf("failed to submit login flow: %w", err)
	}

	return result, nil
}

func (k *kratosServiceImpl) InitializeVerificationFlow(ctx context.Context) (*kratos.VerificationFlow, error) {
	flow, _, err := k.client.FrontendAPI.CreateNativeVerificationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize verification flow: %w", err)
	}
	return flow, nil
}

func (k *kratosServiceImpl) GetVerificationFlow(ctx context.Context, flowID string) (*kratos.VerificationFlow, error) {
	flow, _, err := k.client.FrontendAPI.GetVerificationFlow(ctx).Id(flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get verification flow: %w", err)
	}
	return flow, nil
}

func (k *kratosServiceImpl) SubmitVerificationFlow(
	ctx context.Context,
	flow *kratos.VerificationFlow,
	code string,
) (*kratos.VerificationFlow, error) {
	codePtr := code
	body := kratos.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &kratos.UpdateVerificationFlowWithCodeMethod{
			Method: "code",
			Code:   &codePtr,
		},
	}

	result, resp, err := k.client.FrontendAPI.UpdateVerificationFlow(ctx).
		Flow(flow.Id).
		UpdateVerificationFlowBody(body).
		Execute()
	if err != nil {
		// Check if this is a 400 response
		if resp != nil && resp.StatusCode == 400 {
			// Parse the response to check the flow state
			var kratosResp struct {
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
			if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err == nil {
				// Check for validation errors
				var validationErrors []string
				for _, node := range kratosResp.UI.Nodes {
					for _, msg := range node.Messages {
						if msg.Type == "error" {
							validationErrors = append(validationErrors, msg.Text)
						}
					}
				}

				// Return validation errors if any
				if len(validationErrors) > 0 {
					return nil, fmt.Errorf("validation errors: %v", validationErrors)
				}

				// Handle different states
				switch kratosResp.State {
				case "sent_email":
					// This is a successful response for sending verification email
					return flow, nil
				case "passed_challenge":
					// This is a successful response for verification
					return flow, nil
				default:
					// For any other state, return the original error
					return nil, err
				}
			}
		}
		return nil, fmt.Errorf("failed to submit verification flow: %w", err)
	}

	return result, nil
}

func (k *kratosServiceImpl) GetSession(ctx context.Context, sessionToken string) (*kratos.Session, error) {
	session, _, err := k.client.FrontendAPI.ToSession(ctx).
		Cookie(fmt.Sprintf("ory_kratos_session=%s", sessionToken)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (k *kratosServiceImpl) RevokeSession(ctx context.Context, sessionToken string) error {
	_, err := k.client.FrontendAPI.DisableMySession(ctx, sessionToken).Execute()
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}
