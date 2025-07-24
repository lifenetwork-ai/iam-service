package domain

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	kratos "github.com/ory/kratos-client-go"
)

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
	InitializeVerificationFlow(ctx context.Context, tenantID uuid.UUID) (string, error)
	GetVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.VerificationFlow, error)
	SubmitVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string, code *string) (*kratos.VerificationFlow, error)

	// Logout flow
	Logout(ctx context.Context, tenantID uuid.UUID, sessionToken string) error

	// Session management
	GetSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error)
	RevokeSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) error
	WhoAmI(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error)

	// Update identifier trait
	UpdateIdentifierTrait(ctx context.Context, tenantID uuid.UUID, identityID, identifierType, newIdentifier string) error

	// Settings flow
	InitializeSettingsFlow(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.SettingsFlow, error)
	SubmitSettingsFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.SettingsFlow, sessionToken, method string, traits map[string]interface{}) (*kratos.SettingsFlow, error)
	GetSettingsFlow(ctx context.Context, tenantID uuid.UUID, flowID, sessionToken string) (*kratos.SettingsFlow, error)
}

type KetoService interface {
	CheckPermission(ctx context.Context, request types.CheckPermissionRequest) (bool, *domainerrors.DomainError)
	// BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, *domainerrors.DomainError)
	CreateRelationTuple(ctx context.Context, request types.CreateRelationTupleRequest) *domainerrors.DomainError
}
