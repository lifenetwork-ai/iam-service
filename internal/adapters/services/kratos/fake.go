package kratos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
	kratos "github.com/ory/kratos-client-go"
)

// Fault flags
type Faults struct {
	FailRegistration       bool
	FailLogin              bool
	FailVerification       bool
	FailUpdate             bool
	FailDelete             bool
	NetworkError           bool
	SimulateDuplicateError bool
	ExpireFlowsAfter       time.Duration
	RejectOTP              bool
}

// FakeKratosService with fault injection.
type FakeKratosService struct {
	mu         sync.Mutex
	identities map[uuid.UUID]map[string]*kratos.Identity
	sessions   map[string]*kratos.Session
	flows      map[string]*flowRecord
	faults     Faults
	langs      map[uuid.UUID]string
}

type flowRecord struct {
	createdAt time.Time
	flowType  string // registration|login|verification|settings
	traits    map[string]interface{}
}

func NewFakeKratosService() *FakeKratosService {
	return &FakeKratosService{
		identities: make(map[uuid.UUID]map[string]*kratos.Identity),
		sessions:   make(map[string]*kratos.Session),
		flows:      make(map[string]*flowRecord),
		langs:      make(map[uuid.UUID]string),
	}
}

// Allow tests to configure faults.
func (f *FakeKratosService) SetFaults(faults Faults) {
	f.faults = faults
}

func (f *FakeKratosService) GetIdentities(ctx context.Context, tenantID uuid.UUID) (map[string]*kratos.Identity, error) {
	return f.identities[tenantID], nil
}

// --- Registration flow ---

func (f *FakeKratosService) InitializeRegistrationFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.RegistrationFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	id := uuid.NewString()
	f.flows[id] = &flowRecord{createdAt: time.Now(), flowType: "registration"}
	return &kratos.RegistrationFlow{Id: id}, nil
}

func (f *FakeKratosService) SubmitRegistrationFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.RegistrationFlow,
	method string,
	traits map[string]interface{},
) (*kratos.SuccessfulNativeRegistration, error) {
	if f.faults.NetworkError || f.faults.FailRegistration {
		return nil, errors.New("registration failed")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(f.flows[flow.Id].createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	id := uuid.NewString()
	identity := &kratos.Identity{Id: id, Traits: traits}

	if f.identities[tenantID] == nil {
		f.identities[tenantID] = make(map[string]*kratos.Identity)
	}
	for k, v := range traits {
		if k == constants.IdentifierEmail.String() || k == constants.IdentifierPhone.String() {
			val := v.(string)
			if f.faults.SimulateDuplicateError {
				if _, exists := f.identities[tenantID][val]; exists {
					return nil, fmt.Errorf("duplicate identifier: %s", val)
				}
			}
			f.identities[tenantID][val] = identity
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
	if r, ok := f.flows[flow.Id]; ok {
		r.traits = traits
	}

	return &kratos.SuccessfulNativeRegistration{Session: session, SessionToken: &token}, nil
}

// --- Login flow ---

func (f *FakeKratosService) InitializeLoginFlow(ctx context.Context, tenantID uuid.UUID) (*kratos.LoginFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	id := uuid.NewString()
	f.flows[id] = &flowRecord{createdAt: time.Now(), flowType: "login"}
	return &kratos.LoginFlow{Id: id}, nil
}

func (f *FakeKratosService) SubmitLoginFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flow *kratos.LoginFlow,
	method string,
	identifier, password, code *string,
) (*kratos.SuccessfulNativeLogin, error) {
	if f.faults.NetworkError || f.faults.FailLogin {
		return nil, errors.New("login failed")
	}
	if identifier == nil {
		return nil, fmt.Errorf("identifier required")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(f.flows[flow.Id].createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	idMap := f.identities[tenantID]
	if idMap == nil {
		return nil, fmt.Errorf("tenant not found")
	}
	identity, ok := idMap[*identifier]
	if !ok {
		if utils.IsEmail(*identifier) {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("phone number not found")
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

// --- Verification flow ---

func (f *FakeKratosService) SubmitVerificationFlow(
	ctx context.Context,
	tenantID uuid.UUID,
	flowID string,
	identifier *string,
	identifierType constants.IdentifierType,
	code *string,
) (*kratos.VerificationFlow, error) {
	if f.faults.NetworkError || f.faults.FailVerification {
		return nil, errors.New("verification failed")
	}
	if rec, ok := f.flows[flowID]; ok {
		if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
			return nil, errors.New("flow expired")
		}
	}
	if f.faults.RejectOTP {
		return &kratos.VerificationFlow{Id: flowID, State: "failed_challenge"}, nil
	}
	return &kratos.VerificationFlow{Id: flowID, State: "passed_challenge"}, nil
}

// --- Admin operations with faults ---

func (f *FakeKratosService) UpdateIdentifierTraitAdmin(ctx context.Context, tenantID, identityID uuid.UUID, traits map[string]interface{}) error {
	if f.faults.NetworkError || f.faults.FailUpdate {
		return errors.New("update failed")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, idMap := range f.identities {
		for _, ident := range idMap {
			if ident.Id == identityID.String() {
				ident.Traits = traits
				return nil
			}
		}
	}
	return fmt.Errorf("identifier not found")
}

func (f *FakeKratosService) DeleteIdentifierAdmin(ctx context.Context, tenantID, identityID uuid.UUID) error {
	if f.faults.NetworkError || f.faults.FailDelete {
		return errors.New("delete failed")
	}

	fmt.Println("identityID", identityID.String())
	fmt.Println("tenantID", tenantID.String())
	f.mu.Lock()
	defer f.mu.Unlock()

	idMap, ok := f.identities[tenantID]
	if !ok {
		return fmt.Errorf("tenant not found")
	}
	fmt.Println("idMap", idMap)

	// Remove every identifier string that points to this identityID
	for key, ident := range idMap {
		fmt.Println("ident", ident.Id)

		if ident.Id == identityID.String() {
			delete(idMap, key)
		}
	}

	fmt.Println("idMap", idMap)

	// Optional: clean up if tenant has no identities left
	if len(idMap) == 0 {
		delete(f.identities, tenantID)
	}

	return nil
}

// --- Additional interface methods to satisfy KratosService ---

func (f *FakeKratosService) GetRegistrationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.RegistrationFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	rec, ok := f.flows[flowID]
	if !ok || rec.flowType != "registration" {
		return nil, fmt.Errorf("flow not found")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}
	return &kratos.RegistrationFlow{Id: flowID}, nil
}

func (f *FakeKratosService) SubmitRegistrationFlowWithCode(ctx context.Context, tenantID uuid.UUID, flow *kratos.RegistrationFlow, code string) (*kratos.SuccessfulNativeRegistration, error) {
	if f.faults.NetworkError || f.faults.FailRegistration {
		return nil, errors.New("registration failed")
	}
	rec, ok := f.flows[flow.Id]
	if !ok || rec.flowType != "registration" {
		return nil, fmt.Errorf("flow not found")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	id := uuid.NewString()
	identity := &kratos.Identity{Id: id, Traits: rec.traits}
	if f.identities[tenantID] == nil {
		f.identities[tenantID] = make(map[string]*kratos.Identity)
	}
	for k, v := range rec.traits {
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

func (f *FakeKratosService) GetLoginFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.LoginFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	rec, ok := f.flows[flowID]
	if !ok || rec.flowType != "login" {
		return nil, fmt.Errorf("flow not found")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}
	return &kratos.LoginFlow{Id: flowID}, nil
}

func (f *FakeKratosService) InitializeVerificationFlow(ctx context.Context, tenantID uuid.UUID) (string, error) {
	if f.faults.NetworkError {
		return "", errors.New("network error")
	}
	id := uuid.NewString()
	f.flows[id] = &flowRecord{createdAt: time.Now(), flowType: "verification"}
	return id, nil
}

func (f *FakeKratosService) GetVerificationFlow(ctx context.Context, tenantID uuid.UUID, flowID string) (*kratos.VerificationFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	rec, ok := f.flows[flowID]
	if !ok || rec.flowType != "verification" {
		return nil, fmt.Errorf("flow not found")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}
	return &kratos.VerificationFlow{Id: flowID, State: "sent_challenge"}, nil
}

func (f *FakeKratosService) Logout(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	if f.faults.NetworkError {
		return errors.New("network error")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, sessionToken)
	return nil
}

func (f *FakeKratosService) GetSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[sessionToken]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return s, nil
}

func (f *FakeKratosService) RevokeSession(ctx context.Context, tenantID uuid.UUID, sessionToken string) error {
	return f.Logout(ctx, tenantID, sessionToken)
}

func (f *FakeKratosService) WhoAmI(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.Session, error) {
	return f.GetSession(ctx, tenantID, sessionToken)
}

func (f *FakeKratosService) InitializeSettingsFlow(ctx context.Context, tenantID uuid.UUID, sessionToken string) (*kratos.SettingsFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	id := uuid.NewString()
	f.flows[id] = &flowRecord{createdAt: time.Now(), flowType: "settings"}
	return &kratos.SettingsFlow{Id: id}, nil
}

func (f *FakeKratosService) SubmitSettingsFlow(ctx context.Context, tenantID uuid.UUID, flow *kratos.SettingsFlow, sessionToken, method string, traits map[string]interface{}) (*kratos.SettingsFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	if rec, ok := f.flows[flow.Id]; ok {
		rec.traits = traits
	}
	return &kratos.SettingsFlow{Id: flow.Id}, nil
}

func (f *FakeKratosService) GetSettingsFlow(ctx context.Context, tenantID uuid.UUID, flowID, sessionToken string) (*kratos.SettingsFlow, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	rec, ok := f.flows[flowID]
	if !ok || rec.flowType != "settings" {
		return nil, fmt.Errorf("flow not found")
	}
	if f.faults.ExpireFlowsAfter > 0 && time.Since(rec.createdAt) > f.faults.ExpireFlowsAfter {
		return nil, errors.New("flow expired")
	}
	return &kratos.SettingsFlow{Id: flowID}, nil
}

func (f *FakeKratosService) GetIdentity(ctx context.Context, tenantID, identityID uuid.UUID) (*kratos.Identity, error) {
	if f.faults.NetworkError {
		return nil, errors.New("network error")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, idMap := range f.identities {
		for _, ident := range idMap {
			if ident.Id == identityID.String() {
				return ident, nil
			}
		}
	}
	return nil, fmt.Errorf("identity not found")
}

func (f *FakeKratosService) UpdateLangAdmin(ctx context.Context, tenantID, identityID uuid.UUID, newLang string) error {
	if f.faults.NetworkError {
		return errors.New("network error")
	}
	f.langs[tenantID] = newLang
	return nil
}

// helpers
func ptr[T any](v T) *T { return &v }
