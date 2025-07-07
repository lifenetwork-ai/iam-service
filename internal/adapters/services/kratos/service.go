package kratos

import (
	"context"

	kratos "github.com/ory/kratos-client-go"
)

// Service defines the interface for interacting with Ory Kratos
type Service interface {
	// Registration
	CreateRegistrationFlow(ctx context.Context) (*kratos.RegistrationFlow, error)
	UpdateRegistrationFlow(ctx context.Context, flow string, body kratos.UpdateRegistrationFlowBody) (*kratos.SuccessfulNativeRegistration, error)

	// Login
	CreateLoginFlow(ctx context.Context) (*kratos.LoginFlow, error)
	UpdateLoginFlow(ctx context.Context, flow string, body kratos.UpdateLoginFlowBody) (*kratos.SuccessfulNativeLogin, error)

	// Verification
	CreateVerificationFlow(ctx context.Context) (*kratos.VerificationFlow, error)
	UpdateVerificationFlow(ctx context.Context, flow string, body kratos.UpdateVerificationFlowBody) (*kratos.VerificationFlow, error)
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
func (s *service) CreateRegistrationFlow(ctx context.Context) (*kratos.RegistrationFlow, error) {
	flow, _, err := s.client.PublicAPI().FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	return flow, err
}

// UpdateRegistrationFlow submits a registration flow
func (s *service) UpdateRegistrationFlow(ctx context.Context, flow string, body kratos.UpdateRegistrationFlowBody) (*kratos.SuccessfulNativeRegistration, error) {
	result, _, err := s.client.PublicAPI().FrontendAPI.UpdateRegistrationFlow(ctx).Flow(flow).UpdateRegistrationFlowBody(body).Execute()
	return result, err
}

// CreateLoginFlow initiates a new login flow
func (s *service) CreateLoginFlow(ctx context.Context) (*kratos.LoginFlow, error) {
	flow, _, err := s.client.PublicAPI().FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	return flow, err
}

// UpdateLoginFlow submits a login flow
func (s *service) UpdateLoginFlow(ctx context.Context, flow string, body kratos.UpdateLoginFlowBody) (*kratos.SuccessfulNativeLogin, error) {
	result, _, err := s.client.PublicAPI().FrontendAPI.UpdateLoginFlow(ctx).Flow(flow).UpdateLoginFlowBody(body).Execute()
	return result, err
}

// CreateVerificationFlow initiates a new verification flow
func (s *service) CreateVerificationFlow(ctx context.Context) (*kratos.VerificationFlow, error) {
	flow, _, err := s.client.PublicAPI().FrontendAPI.CreateNativeVerificationFlow(ctx).Execute()
	return flow, err
}

// UpdateVerificationFlow submits a verification flow
func (s *service) UpdateVerificationFlow(ctx context.Context, flow string, body kratos.UpdateVerificationFlowBody) (*kratos.VerificationFlow, error) {
	result, _, err := s.client.PublicAPI().FrontendAPI.UpdateVerificationFlow(ctx).Flow(flow).UpdateVerificationFlowBody(body).Execute()
	return result, err
}
