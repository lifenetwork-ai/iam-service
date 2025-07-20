package kratos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	repo_types "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	kratos_types "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos/types"
	kratos "github.com/ory/kratos-client-go"
	"github.com/pkg/errors"
)

type kratosServiceImpl struct {
	client *Client
}

// NewKratosService creates a new instance of KratosService
func NewKratosService(tenantRepo repo_types.TenantRepository) kratos_types.KratosService {
	config := conf.GetKratosConfig()
	client := NewClient(config, tenantRepo)
	return &kratosServiceImpl{
		client: client,
	}
}

// InitializeRegistrationFlow initiates a new registration flow
func (k *kratosServiceImpl) InitializeRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registration flow: %w", err)
	}
	return flow, nil
}

// SubmitRegistrationFlow submits a registration flow
func (k *kratosServiceImpl) SubmitRegistrationFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.RegistrationFlow,
	method string,
	traits map[string]interface{},
) (*kratos.SuccessfulNativeRegistration, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	submitFlow := publicAPI.FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow.Id)

	var body kratos.UpdateRegistrationFlowBody
	switch method {
	case constants.MethodTypeCode.String():
		body.UpdateRegistrationFlowWithCodeMethod = &kratos.UpdateRegistrationFlowWithCodeMethod{
			Method: constants.MethodTypeCode.String(),
			Traits: traits,
		}

		// This method always return 400 error, even if the registration is successful
		result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
		if err != nil {
			if resp != nil && resp.StatusCode == 400 {
				if err := parseKratosErrorResponse(resp, fmt.Errorf("registration failed: %w", err)); err != nil {
					return nil, err
				}
				return &kratos.SuccessfulNativeRegistration{}, nil
			}
			return nil, fmt.Errorf("registration failed: %w", err)
		}
		return result, nil

	case constants.MethodTypePassword.String():
		body.UpdateRegistrationFlowWithPasswordMethod = &kratos.UpdateRegistrationFlowWithPasswordMethod{
			Method: constants.MethodTypePassword.String(),
			Traits: traits,
		}
		result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
		if err != nil {
			if resp != nil && resp.StatusCode == 400 {
				if err := parseKratosErrorResponse(resp, err); err != nil {
					return nil, err
				}
				return &kratos.SuccessfulNativeRegistration{}, nil
			}
			return nil, err
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// GetRegistrationFlow gets a registration flow
func (k *kratosServiceImpl) GetRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.RegistrationFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.GetRegistrationFlow(ctx).Id(flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get registration flow: %w", err)
	}
	return flow, nil
}

// SubmitRegistrationFlowWithCode submits a registration flow with code
func (k *kratosServiceImpl) SubmitRegistrationFlowWithCode(ctx context.Context, tenantID uuid.UUID, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	// Get the flow again to extract traits
	_, resp, err := publicAPI.FrontendAPI.GetRegistrationFlow(ctx).Id(flow.Id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get registration flow: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from registration flow")
	}
	defer resp.Body.Close()

	flowData, err := kratos_types.FromHttpResp(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flow data: %w", err)
	}

	traits := flowData.GetTraits()
	if len(traits) == 0 {
		return nil, fmt.Errorf("no traits found in registration flow")
	}

	// Submit the flow with code
	submitFlow := publicAPI.FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow.Id)
	body := kratos.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithCodeMethod: &kratos.UpdateRegistrationFlowWithCodeMethod{
			Method: constants.MethodTypeCode.String(),
			Code:   &code,
			Traits: traits,
		},
	}

	result, resp, err := submitFlow.UpdateRegistrationFlowBody(body).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 400 {
			if err := parseKratosErrorResponse(resp, fmt.Errorf("registration failed: %w", err)); err != nil {
				return nil, err
			}
			return &kratos.SuccessfulNativeRegistration{}, nil
		}
		return nil, fmt.Errorf("failed to submit registration flow with code: %w", err)
	}

	return result, nil
}

// InitializeLoginFlow initiates a new login flow
func (k *kratosServiceImpl) InitializeLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize login flow: %w", err)
	}
	return flow, nil
}

// GetLoginFlow gets a login flow
func (k *kratosServiceImpl) GetLoginFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.LoginFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.GetLoginFlow(ctx).Id(flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get login flow: %w", err)
	}
	return flow, nil
}

// SubmitLoginFlow submits a login flow
func (k *kratosServiceImpl) SubmitLoginFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.LoginFlow, method string, identifier, password, code *string) (*kratos.SuccessfulNativeLogin, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	submitFlow := publicAPI.FrontendAPI.UpdateLoginFlow(ctx).Flow(flow.Id)

	var body kratos.UpdateLoginFlowBody
	switch method {
	case constants.MethodTypeCode.String():
		body.UpdateLoginFlowWithCodeMethod = &kratos.UpdateLoginFlowWithCodeMethod{
			Method:     constants.MethodTypeCode.String(),
			Identifier: identifier,
			Code:       code,
		}
	case constants.MethodTypePassword.String():
		body.UpdateLoginFlowWithPasswordMethod = &kratos.UpdateLoginFlowWithPasswordMethod{
			Method:     constants.MethodTypePassword.String(),
			Password:   *password,
			Identifier: *identifier,
		}
	default:
		return nil, fmt.Errorf("unsupported login method: %s", method)
	}

	result, resp, err := submitFlow.UpdateLoginFlowBody(body).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 400 {
			if err := parseKratosErrorResponse(resp, fmt.Errorf("login failed: %w", err)); err != nil {
				return nil, err
			}
			return &kratos.SuccessfulNativeLogin{}, nil
		}
		return nil, fmt.Errorf("failed to submit login flow: %w", err)
	}

	return result, nil
}

// InitializeVerificationFlow initiates a new verification flow
func (k *kratosServiceImpl) InitializeVerificationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.VerificationFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeVerificationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize verification flow: %w", err)
	}
	return flow, nil
}

// GetVerificationFlow gets a verification flow
func (k *kratosServiceImpl) GetVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.VerificationFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}
	flow, _, err := publicAPI.FrontendAPI.GetVerificationFlow(ctx).Id(flowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get verification flow: %w", err)
	}
	return flow, nil
}

// SubmitVerificationFlow submits a verification flow
func (k *kratosServiceImpl) SubmitVerificationFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.VerificationFlow, code string) (*kratos.VerificationFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	codePtr := code
	body := kratos.UpdateVerificationFlowBody{
		UpdateVerificationFlowWithCodeMethod: &kratos.UpdateVerificationFlowWithCodeMethod{
			Method: constants.MethodTypeCode.String(),
			Code:   &codePtr,
		},
	}

	result, resp, err := publicAPI.FrontendAPI.UpdateVerificationFlow(ctx).
		Flow(flow.Id).
		UpdateVerificationFlowBody(body).
		Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 400 {
			if err := parseKratosErrorResponse(resp, fmt.Errorf("verification failed: %w", err)); err != nil {
				return nil, err
			}
			return flow, nil
		}
		return nil, fmt.Errorf("failed to submit verification flow: %w", err)
	}

	return result, nil
}

// Logout logs out the user
func (k *kratosServiceImpl) Logout(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get public API client: %w", err)
	}

	_, err = publicAPI.FrontendAPI.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(kratos.PerformNativeLogoutBody{
			SessionToken: sessionToken,
		}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to perform native logout: %w", err)
	}
	return nil
}

// GetSession gets a session
func (k *kratosServiceImpl) GetSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	session, _, err := publicAPI.FrontendAPI.ToSession(ctx).XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// RevokeSession revokes a session
func (k *kratosServiceImpl) RevokeSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get public API client: %w", err)
	}

	_, err = publicAPI.FrontendAPI.DisableMySession(ctx, sessionToken).XSessionToken(sessionToken).Execute()
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// WhoAmI gets the current session
func (k *kratosServiceImpl) WhoAmI(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	session, _, err := publicAPI.FrontendAPI.ToSession(ctx).XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get whoami session: %w", err)
	}
	return session, nil
}

// UpdateIdentifierTrait updates the identifier trait (email or phone number) in the identity
func (k *kratosServiceImpl) UpdateIdentifierTrait(
	ctx context.Context,
	tenantID uuid.UUID,
	identityID, identifierType, newIdentifier string,
) error {
	adminAPI, err := k.client.AdminAPI(tenantID)
	if err != nil {
		return fmt.Errorf("get admin API failed: %w", err)
	}

	identity, _, err := adminAPI.IdentityAPI.GetIdentity(ctx, identityID).Execute()
	if err != nil {
		return fmt.Errorf("fetch identity failed: %w", err)
	}

	traits, ok := identity.Traits.(map[string]interface{})
	if !ok {
		return fmt.Errorf("identity traits cast failed")
	}
	traits[identifierType] = newIdentifier

	update := kratos.UpdateIdentityBody{Traits: traits}
	_, _, err = adminAPI.IdentityAPI.UpdateIdentity(ctx, identityID).
		UpdateIdentityBody(update).Execute()
	if err != nil {
		return fmt.Errorf("update identity failed: %w", err)
	}

	return nil
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
	case "sent_email":
		return nil
	case "choose_method":
		return errors.New("error occurred while submitting flow")
	default:
		return defaultErr
	}
}

// InitializeSettingsFlow initializes a settings flow for the user
func (k *kratosServiceImpl) InitializeSettingsFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	sessionToken string,
) (*kratos.SettingsFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	flow, _, err := publicAPI.FrontendAPI.CreateNativeSettingsFlow(ctx).
		XSessionToken(sessionToken).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize settings flow: %w", err)
	}

	return flow, nil
}

// GetSettingsFlow retrieves a settings flow by its ID
func (k *kratosServiceImpl) GetSettingsFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	sessionToken string,
) (*kratos.SettingsFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	flow, _, err := publicAPI.FrontendAPI.GetSettingsFlow(ctx).
		Id(flowID).
		XSessionToken(sessionToken).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings flow: %w", err)
	}

	return flow, nil
}

func (k *kratosServiceImpl) SubmitSettingsFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.SettingsFlow,
	sessionToken string,
	method string,
	traits map[string]interface{},
) (*kratos.SettingsFlow, error) {
	publicAPI, err := k.client.PublicAPI(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public API client: %w", err)
	}

	submit := publicAPI.FrontendAPI.UpdateSettingsFlow(ctx).Flow(flow.Id).XSessionToken(sessionToken)

	var body kratos.UpdateSettingsFlowBody
	switch method {
	case constants.MethodTypeProfile.String():
		body.UpdateSettingsFlowWithProfileMethod = &kratos.UpdateSettingsFlowWithProfileMethod{
			Method: method,
			Traits: traits,
		}
	default:
		return nil, fmt.Errorf("unsupported settings method: %s", method)
	}

	result, resp, err := submit.UpdateSettingsFlowBody(body).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 400 {
			if err := k.parseKratosErrorResponse(resp, fmt.Errorf("settings update failed: %w", err)); err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("failed to submit settings flow: %w", err)
	}

	return result, nil
}
