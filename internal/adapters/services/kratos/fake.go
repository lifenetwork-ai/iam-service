package kratos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	kratos "github.com/ory/kratos-client-go"
)

// FakeKratosService is a stateful in-memory fake for tests.
type FakeKratosService struct {
	mu         sync.Mutex
	identities map[uuid.UUID]map[string]*kratos.Identity // tenantID → identifier → identity
	sessions   map[string]*kratos.Session                // sessionToken → session
}

func NewFakeKratosService() *FakeKratosService {
	return &FakeKratosService{
		identities: make(map[uuid.UUID]map[string]*kratos.Identity),
		sessions:   make(map[string]*kratos.Session),
	}
}

// --- Registration flow ---

func (f *FakeKratosService) InitializeRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error) {
	return &kratos.RegistrationFlow{Id: uuid.NewString()}, nil
}

func (f *FakeKratosService) SubmitRegistrationFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.RegistrationFlow,
	method string,
	traits map[string]interface{},
) (*kratos.SuccessfulNativeRegistration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := uuid.NewString()
	identity := &kratos.Identity{Id: id, Traits: traits}

	if f.identities[tenantID] == nil {
		f.identities[tenantID] = make(map[string]*kratos.Identity)
	}
	for k, v := range traits {
		if k == constants.IdentifierEmail.String() || k == constants.IdentifierPhone.String() {
			f.identities[tenantID][v.(string)] = identity
		}
	}

	session := &kratos.Session{
		Id:              uuid.NewString(),
		Active:          ptr(true),
		Identity:        identity,
		AuthenticatedAt: ptr(time.Now()),
		ExpiresAt:       ptr(time.Now().Add(30 * time.Minute)),
	}
	token := uuid.NewString()
	f.sessions[token] = session

	return &kratos.SuccessfulNativeRegistration{Session: session, SessionToken: &token}, nil
}

func (f *FakeKratosService) GetRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.RegistrationFlow, error) {
	return &kratos.RegistrationFlow{Id: flowID}, nil
}

//	{
//	    "session_token": "ory_st_YJ87MAIry3Cny2jLMfD1IYQ2ikSw3mo0",
//	    "session": {
//	        "id": "bae64228-e59b-4522-8225-965b4d75533e",
//	        "active": true,
//	        "expires_at": "2025-09-10T08:15:48.79483081Z",
//	        "authenticated_at": "2025-09-10T08:10:48.79483081Z",
//	        "authenticator_assurance_level": "aal1",
//	        "authentication_methods": [
//	            {
//	                "method": "code",
//	                "aal": "aal1",
//	                "completed_at": "2025-09-10T08:10:48.7948301Z"
//	            }
//	        ],
//	        "issued_at": "2025-09-10T08:10:48.79483081Z",
//	        "identity": {
//	            "id": "bc60c801-5d1f-4f16-a195-777b6546c295",
//	            "schema_id": "default",
//	            "schema_url": "https://auth.develop.lifenetwork.ai/schemas/ZGVmYXVsdA",
//	            "state": "active",
//	            "state_changed_at": "2025-09-10T08:10:48.76121218Z",
//	            "traits": {
//	                "phone_number": "+84346840621",
//	                "tenant": "life_ai"
//	            },
//	            "verifiable_addresses": [
//	                {
//	                    "id": "9db67cae-3549-46aa-86cd-752e642c73a1",
//	                    "value": "+84346840621",
//	                    "verified": true,
//	                    "via": "sms",
//	                    "status": "completed",
//	                    "verified_at": "2025-09-10T08:10:48.753381758Z",
//	                    "created_at": "2025-09-10T08:10:48.77096Z",
//	                    "updated_at": "2025-09-10T08:10:48.77096Z"
//	                }
//	            ],
//	            "metadata_public": null,
//	            "created_at": "2025-09-10T08:10:48.76513Z",
//	            "updated_at": "2025-09-10T08:10:48.76513Z",
//	            "organization_id": null
//	        },
//	        "devices": [
//	            {
//	                "id": "dacc57aa-f061-4367-9406-882b8da0434d",
//	                "ip_address": "2401:d800:147:d2b9:cc20:bf09:6b41:47db",
//	                "user_agent": "PostmanRuntime/7.45.0",
//	                "location": ""
//	            }
//	        ]
//	    },
//	    "identity": {
//	        "id": "bc60c801-5d1f-4f16-a195-777b6546c295",
//	        "schema_id": "default",
//	        "schema_url": "https://auth.develop.lifenetwork.ai/schemas/ZGVmYXVsdA",
//	        "state": "active",
//	        "state_changed_at": "2025-09-10T08:10:48.76121218Z",
//	        "traits": {
//	            "phone_number": "+84346840621",
//	            "tenant": "life_ai"
//	        },
//	        "verifiable_addresses": [
//	            {
//	                "id": "9db67cae-3549-46aa-86cd-752e642c73a1",
//	                "value": "+84346840621",
//	                "verified": true,
//	                "via": "sms",
//	                "status": "completed",
//	                "verified_at": "2025-09-10T08:10:48.753381758Z",
//	                "created_at": "2025-09-10T08:10:48.77096Z",
//	                "updated_at": "2025-09-10T08:10:48.77096Z"
//	            }
//	        ],
//	        "metadata_public": null,
//	        "created_at": "2025-09-10T08:10:48.76513Z",
//	        "updated_at": "2025-09-10T08:10:48.76513Z",
//	        "organization_id": null
//	    },
//	    "continue_with": [
//	        {
//	            "action": "set_ory_session_token",
//	            "ory_session_token": "ory_st_YJ87MAIry3Cny2jLMfD1IYQ2ikSw3mo0"
//	        }
//	    ]
//	}
func (f *FakeKratosService) SubmitRegistrationFlowWithCode(ctx context.Context, tenantID uuid.UUID, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error) {
	// For simplicity, reuse SubmitRegistrationFlow logic
	return &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:              uuid.NewString(),
			Active:          ptr(true),
			Identity:        &kratos.Identity{Id: uuid.NewString(), Traits: map[string]interface{}{"tenant": "tenant-name"}},
			AuthenticatedAt: ptr(time.Now()),
			ExpiresAt:       ptr(time.Now().Add(30 * time.Minute)),
		},
		SessionToken: ptr(uuid.NewString()),
	}, nil
}

// --- Login flow ---

func (f *FakeKratosService) InitializeLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error) {
	return &kratos.LoginFlow{Id: uuid.NewString()}, nil
}

func (f *FakeKratosService) SubmitLoginFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.LoginFlow,
	method string,
	identifier, password, code *string,
) (*kratos.SuccessfulNativeLogin, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if identifier == nil {
		return nil, fmt.Errorf("identifier required")
	}

	idMap := f.identities[tenantID]
	if idMap == nil {
		return nil, fmt.Errorf("tenant not found")
	}
	identity, ok := idMap[*identifier]
	if !ok {
		return nil, fmt.Errorf("identity not found")
	}

	session := &kratos.Session{
		Id:              uuid.NewString(),
		Active:          ptr(true),
		Identity:        identity,
		AuthenticatedAt: ptr(time.Now()),
		ExpiresAt:       ptr(time.Now().Add(30 * time.Minute)),
	}
	token := uuid.NewString()
	f.sessions[token] = session

	return &kratos.SuccessfulNativeLogin{Session: *session, SessionToken: &token}, nil
}

func (f *FakeKratosService) GetLoginFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.LoginFlow, error) {
	return &kratos.LoginFlow{Id: flowID}, nil
}

// --- Verification flow (stubbed minimal) ---

func (f *FakeKratosService) InitializeVerificationFlow(ctx context.Context, tenantID uuid.UUID) (string, error) {
	return uuid.NewString(), nil
}

func (f *FakeKratosService) GetVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.VerificationFlow, error) {
	return &kratos.VerificationFlow{Id: flowID}, nil
}

func (f *FakeKratosService) SubmitVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string, identifier *string, identifierType constants.IdentifierType, code *string) (*kratos.VerificationFlow, error) {
	return &kratos.VerificationFlow{Id: flowID, State: "passed_challenge"}, nil
}

// --- Logout and session ---

func (f *FakeKratosService) Logout(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	delete(f.sessions, sessionToken)
	return nil
}

func (f *FakeKratosService) GetSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	sess, ok := f.sessions[sessionToken]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return sess, nil
}

func (f *FakeKratosService) RevokeSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	delete(f.sessions, sessionToken)
	return nil
}

func (f *FakeKratosService) WhoAmI(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	return f.GetSession(ctx, tenantID, sessionToken)
}

// --- Settings flow (stubbed) ---

func (f *FakeKratosService) InitializeSettingsFlow(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.SettingsFlow, error) {
	return &kratos.SettingsFlow{Id: uuid.NewString()}, nil
}

func (f *FakeKratosService) SubmitSettingsFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.SettingsFlow, sessionToken, method string, traits map[string]interface{}) (*kratos.SettingsFlow, error) {
	return flow, nil
}

func (f *FakeKratosService) GetSettingsFlow(ctx context.Context, tenantID uuid.UUID, flowID, sessionToken string) (*kratos.SettingsFlow, error) {
	return &kratos.SettingsFlow{Id: flowID}, nil
}

// --- Admin API ---

func (f *FakeKratosService) GetIdentity(ctx context.Context, tenantID, identityID uuid.UUID) (*kratos.Identity, error) {
	// naive search
	for _, idMap := range f.identities {
		for _, ident := range idMap {
			if ident.Id == identityID.String() {
				return ident, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

func (f *FakeKratosService) UpdateIdentifierTraitAdmin(ctx context.Context, tenantID, identityID uuid.UUID, traits map[string]interface{}) error {
	// just overwrite traits in memory
	for _, idMap := range f.identities {
		for _, ident := range idMap {
			if ident.Id == identityID.String() {
				ident.Traits = traits
				return nil
			}
		}
	}
	return fmt.Errorf("identity not found")
}

func (f *FakeKratosService) DeleteIdentifierAdmin(ctx context.Context, tenantID, identityID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for idVal, ident := range f.identities[tenantID] {
		if ident.Id == identityID.String() {
			delete(f.identities[tenantID], idVal)
		}
	}
	return nil
}

// helper
func ptr[T any](v T) *T { return &v }
