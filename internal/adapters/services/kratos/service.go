package kratos

import (
	"context"

	"github.com/google/uuid"
	kratos "github.com/ory/kratos-client-go"
)

// Service defines the interface for interacting with Ory Kratos
type Service interface {
	// Registration
	CreateRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error)
	UpdateRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateRegistrationFlowBody) (*kratos.SuccessfulNativeRegistration, error)

	// Login
	CreateLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error)
	UpdateLoginFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateLoginFlowBody) (*kratos.SuccessfulNativeLogin, error)

	// Verification
	CreateVerificationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.VerificationFlow, error)
	UpdateVerificationFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateVerificationFlowBody) (*kratos.VerificationFlow, error)
}

type service struct {
	client *Client
}

// NewService creates a new Kratos service
func NewService(client *Client) Service {
	return &service{
		client: client,
	}
}

// CreateRegistrationFlow initiates a new registration flow
func (s *service) CreateRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	return flow, err
}

// UpdateRegistrationFlow submits a registration flow
func (s *service) UpdateRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateRegistrationFlowBody) (*kratos.SuccessfulNativeRegistration, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	result, _, err := publicAPI.FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow).UpdateRegistrationFlowBody(body).Execute()
	return result, err
}

// CreateLoginFlow initiates a new login flow
func (s *service) CreateLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	return flow, err
}

// UpdateLoginFlow submits a login flow
func (s *service) UpdateLoginFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateLoginFlowBody) (*kratos.SuccessfulNativeLogin, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	result, _, err := publicAPI.FrontendAPI.UpdateLoginFlow(ctx).Flow(flow).UpdateLoginFlowBody(body).Execute()
	return result, err
}

// CreateVerificationFlow initiates a new verification flow
func (s *service) CreateVerificationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.VerificationFlow, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	flow, _, err := publicAPI.FrontendAPI.CreateNativeVerificationFlow(ctx).Execute()
	return flow, err
}

// UpdateVerificationFlow submits a verification flow
func (s *service) UpdateVerificationFlow(ctx context.Context, tenantID uuid.UUID, flow string, body kratos.UpdateVerificationFlowBody) (*kratos.VerificationFlow, error) {
	publicAPI, err := s.client.PublicAPI(tenantID)
	if err != nil {
		return nil, err
	}
	result, _, err := publicAPI.FrontendAPI.UpdateVerificationFlow(ctx).Flow(flow).UpdateVerificationFlowBody(body).Execute()
	return result, err
}
