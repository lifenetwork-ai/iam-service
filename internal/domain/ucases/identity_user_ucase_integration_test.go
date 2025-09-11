package ucases

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// build isolated usecase with local mocks
func buildIsolatedUseCase(ctrl *gomock.Controller) (*userUseCase, struct {
	tenantRepo                *mock_repositories.MockTenantRepository
	globalUserRepo            *mock_repositories.MockGlobalUserRepository
	userIdentityRepo          *mock_repositories.MockUserIdentityRepository
	userIdentifierMappingRepo *mock_repositories.MockUserIdentifierMappingRepository
	challengeSessionRepo      *mock_repositories.MockChallengeSessionRepository
	kratosService             *mock_services.MockKratosService
	rateLimiter               *mock_types.MockRateLimiter
},
) {
	deps := struct {
		tenantRepo                *mock_repositories.MockTenantRepository
		globalUserRepo            *mock_repositories.MockGlobalUserRepository
		userIdentityRepo          *mock_repositories.MockUserIdentityRepository
		userIdentifierMappingRepo *mock_repositories.MockUserIdentifierMappingRepository
		challengeSessionRepo      *mock_repositories.MockChallengeSessionRepository
		kratosService             *mock_services.MockKratosService
		rateLimiter               *mock_types.MockRateLimiter
	}{
		tenantRepo:                mock_repositories.NewMockTenantRepository(ctrl),
		globalUserRepo:            mock_repositories.NewMockGlobalUserRepository(ctrl),
		userIdentityRepo:          mock_repositories.NewMockUserIdentityRepository(ctrl),
		userIdentifierMappingRepo: mock_repositories.NewMockUserIdentifierMappingRepository(ctrl),
		challengeSessionRepo:      mock_repositories.NewMockChallengeSessionRepository(ctrl),
		kratosService:             mock_services.NewMockKratosService(ctrl),
		rateLimiter:               mock_types.NewMockRateLimiter(ctrl),
	}

	ucase := &userUseCase{
		db:                        &gorm.DB{},
		rateLimiter:               deps.rateLimiter,
		tenantRepo:                deps.tenantRepo,
		globalUserRepo:            deps.globalUserRepo,
		userIdentityRepo:          deps.userIdentityRepo,
		userIdentifierMappingRepo: deps.userIdentifierMappingRepo,
		challengeSessionRepo:      deps.challengeSessionRepo,
		kratosService:             deps.kratosService,
	}

	return ucase, deps
}

func TestIntegration_ChangeIdentifierThenVerifyRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	// in-memory DB for transactions
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	data := struct {
		globalUserID  string
		tenantID      uuid.UUID
		tenantUserID  string
		flowID        string
		newIdentifier string
	}{
		globalUserID:  "global-user-1",
		tenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		tenantUserID:  "00000000-0000-0000-0000-00000000aaaa",
		flowID:        "flow-int-1",
		newIdentifier: "newemail@example.com",
	}

	// ChangeIdentifier expectations
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).Return([]*domain.UserIdentity{{ID: "identity-email-id", Type: constants.IdentifierEmail.String()}}, nil)
	deps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, data.newIdentifier)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	// VerifyRegister expectations
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		GlobalUserID:   data.globalUserID,
		TenantUserID:   data.tenantUserID,
		Identifier:     data.newIdentifier,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     "identity-email-id",
	}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	newTenantUserID := "00000000-0000-0000-0000-00000000bbbb"
	verifyResult := &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:                    "session-id",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: newTenantUserID, Traits: map[string]interface{}{"tenant": "tenant-name", string(constants.IdentifierEmail): data.newIdentifier}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token"),
	}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(verifyResult, nil)
	deps.userIdentifierMappingRepo.EXPECT().GetByTenantIDAndTenantUserID(gomock.Any(), data.tenantID.String(), data.tenantUserID).Return(&domain.UserIdentifierMapping{ID: "map-id", GlobalUserID: data.globalUserID, TenantID: data.tenantID.String(), TenantUserID: data.tenantUserID}, nil)
	deps.userIdentityRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	deps.userIdentifierMappingRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), data.tenantID, gomock.Any()).Return(nil)
	deps.tenantRepo.EXPECT().GetByName("tenant-name").Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Return(nil)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "123456")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyResp)
	require.Equal(t, newTenantUserID, verifyResp.User.ID)
}

func TestIntegration_ChangeIdentifierThenVerifyRegister_Failure_NoMutations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	data := struct {
		globalUserID  string
		tenantID      uuid.UUID
		tenantUserID  string
		flowID        string
		newIdentifier string
	}{
		globalUserID:  "global-user-1",
		tenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		tenantUserID:  "00000000-0000-0000-0000-00000000cccc",
		flowID:        "flow-int-2",
		newIdentifier: "newemail@example.com",
	}

	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).Return([]*domain.UserIdentity{{ID: "identity-email-id", Type: constants.IdentifierEmail.String()}}, nil)
	deps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, data.newIdentifier)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		GlobalUserID:   data.globalUserID,
		TenantUserID:   data.tenantUserID,
		Identifier:     data.newIdentifier,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     "identity-email-id",
	}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(nil, assert.AnError)

	deps.userIdentityRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
	deps.userIdentifierMappingRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
	deps.userIdentifierMappingRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Times(0)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "000000")
	require.Error(t, verifyErr)
	require.Nil(t, verifyResp)
}

// =========================
// bind update failure & cleanup
// =========================
func TestIntegration_BindUpdateIdentifier_UpdateMappingFails_CleanupCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenant := &domain.Tenant{ID: tenantID, Name: "tenant-name"}
	oldTenantUserID := "00000000-0000-0000-0000-00000000aaaa"
	newTenantUserID := "00000000-0000-0000-0000-00000000eeef"

	// prepare changeIdentifier flow
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "new@example.com").Return(false, nil)
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), tenantID.String(), oldTenantUserID).Return([]*domain.UserIdentity{{ID: "identity-email-id", Type: constants.IdentifierEmail.String()}}, nil)
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(tenant, nil)
	regFlow := &kratos.RegistrationFlow{Id: "flow-bind-fail"}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(regFlow, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), regFlow.Id, gomock.Any(), gomock.Any()).Return(nil)

	_, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, oldTenantUserID, "new@example.com")
	require.Nil(t, derr)

	// verify with mapping update fail to trigger rollbackKratosUpdateIdentifier
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), regFlow.Id).Return(&domain.ChallengeSession{
		GlobalUserID:   "g1",
		TenantUserID:   oldTenantUserID,
		Identifier:     "new@example.com",
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     "identity-email-id",
	}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, regFlow.Id).Return(regFlow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{Session: &kratos.Session{Identity: &kratos.Identity{Id: newTenantUserID, Traits: map[string]interface{}{"tenant": tenant.Name, string(constants.IdentifierEmail): "new@example.com"}}}, SessionToken: ptr("tok")}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, regFlow, gomock.Any()).Return(verifyResult, nil)
	deps.userIdentifierMappingRepo.EXPECT().GetByTenantIDAndTenantUserID(gomock.Any(), tenantID.String(), oldTenantUserID).Return(&domain.UserIdentifierMapping{ID: "map-id", GlobalUserID: "g1", TenantID: tenantID.String(), TenantUserID: oldTenantUserID}, nil)
	deps.userIdentityRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	// fail mapping update
	deps.userIdentifierMappingRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(assert.AnError)
	// cleanup called to delete new identifier in kratos
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), tenantID, uuid.MustParse(newTenantUserID)).Return(nil)

	deps.tenantRepo.EXPECT().GetByName(tenant.Name).Return(tenant, nil)

	resp, vErr := ucase.VerifyRegister(ctx, tenantID, regFlow.Id, "000000")
	require.NotNil(t, vErr)
	require.Nil(t, resp)
}

func TestIntegration_AddIdentifierThenVerifyRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	data := struct {
		tenantID      uuid.UUID
		globalUserID  string
		flowID        string
		newIdentifier string
	}{
		tenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		globalUserID:  "global-user-add-1",
		flowID:        "flow-int-add-1",
		newIdentifier: "addemail@example.com",
	}

	// --- AddNewIdentifier expectations ---
	deps.userIdentityRepo.EXPECT().
		ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).
		Return(false, nil)
	deps.userIdentityRepo.EXPECT().
		ExistsByTenantGlobalUserIDAndType(gomock.Any(), data.tenantID.String(), data.globalUserID, constants.IdentifierEmail.String()).
		Return(false, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().
		InitializeRegistrationFlow(gomock.Any(), data.tenantID).
		Return(regFlow, nil)
	deps.tenantRepo.EXPECT().
		GetByID(data.tenantID).
		Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	deps.kratosService.EXPECT().
		SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).
		Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().
		SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).
		Return(nil)

	addResp, addErr := ucase.AddNewIdentifier(ctx, data.tenantID, data.globalUserID, data.newIdentifier, constants.IdentifierEmail.String())
	require.Nil(t, addErr)
	require.NotNil(t, addResp)

	// --- VerifyRegister expectations (AddIdentifier path) ---
	deps.challengeSessionRepo.EXPECT().
		GetChallenge(gomock.Any(), data.flowID).
		Return(&domain.ChallengeSession{
			GlobalUserID:   data.globalUserID,
			Identifier:     data.newIdentifier,
			IdentifierType: constants.IdentifierEmail.String(),
			ChallengeType:  constants.ChallengeTypeAddIdentifier,
		}, nil)

	deps.kratosService.EXPECT().
		GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).
		Return(regFlow, nil)

	verifyResult := &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:              "session-id-add",
			Active:          ptr(true),
			ExpiresAt:       ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:        ptr(time.Now()),
			AuthenticatedAt: ptr(time.Now()),
			Identity: &kratos.Identity{
				Id:     "tenant-user-add",
				Traits: map[string]interface{}{"tenant": "tenant-name", string(constants.IdentifierEmail): data.newIdentifier},
			},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token-add"),
	}
	deps.kratosService.EXPECT().
		SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).
		Return(verifyResult, nil)

	deps.tenantRepo.EXPECT().
		GetByName("tenant-name").
		Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)

	deps.userIdentityRepo.EXPECT().
		InsertOnceByTenantUserAndType(gomock.Any(), gomock.Any(), data.tenantID.String(), data.globalUserID, constants.IdentifierEmail.String(), data.newIdentifier).
		Return(true, nil)

	// 🔧 BỔ SUNG: expect Create(mapping) trong TX (dùng Any cho tx và mapping)
	deps.userIdentifierMappingRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	deps.challengeSessionRepo.EXPECT().
		DeleteChallenge(gomock.Any(), data.flowID).
		Return(nil)

	// --- Execute ---
	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "654321")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyResp)
	require.Equal(t, "tenant-user-add", verifyResp.User.ID)
}

func TestIntegration_DeleteIdentifier_Behavior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	data := struct {
		tenantID     uuid.UUID
		globalUserID string
		tenantUserID string
	}{
		tenantID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		globalUserID: "global-user-del-1",
		tenantUserID: "00000000-0000-0000-0000-00000000dddd",
	}

	// Case 1: cannot delete only identifier
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).Return([]*domain.UserIdentity{
		{ID: "id-email", Type: constants.IdentifierEmail.String()},
	}, nil)
	derr := ucase.DeleteIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String())
	require.NotNil(t, derr)

	// Case 2: delete when there are two identifiers
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).Return([]*domain.UserIdentity{
		{ID: "id-email", Type: constants.IdentifierEmail.String()},
		{ID: "id-phone", Type: constants.IdentifierPhone.String()},
	}, nil)
	deps.userIdentityRepo.EXPECT().Delete(gomock.Any(), "id-email").Return(nil)
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), data.tenantID, gomock.Any()).Return(nil)
	derr = ucase.DeleteIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String())
	require.Nil(t, derr)
}

func TestIntegration_RegisterThenVerifyRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	data := struct {
		tenantID uuid.UUID
		flowID   string
		email    string
		lang     string
	}{
		tenantID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		flowID:   "flow-int-reg-1",
		email:    "newreg@example.com",
		lang:     "en",
	}

	// Register expectations
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.email).Return(false, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	deps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	regResp, regErr := ucase.Register(ctx, data.tenantID, data.lang, data.email, "")
	require.Nil(t, regErr)
	require.NotNil(t, regResp)
	require.True(t, regResp.VerificationNeeded)
	require.Equal(t, data.flowID, regResp.VerificationFlow.FlowID)

	// VerifyRegister expectations for default register path
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		Identifier:     data.email,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeRegister,
	}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:                    "session-id-reg",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: "tenant-user-reg", Traits: map[string]interface{}{"tenant": "tenant-name", string(constants.IdentifierEmail): data.email}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token-reg"),
	}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(verifyResult, nil)
	deps.tenantRepo.EXPECT().GetByName("tenant-name").Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	// bindIAMToRegistration path (no existing identity): create global user, insert identity, create mapping
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.email).Return(nil, assert.AnError)
	deps.globalUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	deps.userIdentityRepo.EXPECT().InsertOnceByTenantUserAndType(gomock.Any(), gomock.Any(), data.tenantID.String(), gomock.Any(), constants.IdentifierEmail.String(), data.email).Return(true, nil)
	deps.userIdentifierMappingRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Return(nil)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "112233")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyResp)
	require.Equal(t, "tenant-user-reg", verifyResp.User.ID)
}

// =========================
// Additional success-path coverage to raise file coverage
// =========================
func TestIntegration_ChallengeWithPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	// provide non-normalized input to test normalization path
	inputPhone := "(202) 555-0123"
	flow := &kratos.LoginFlow{Id: "flow-login-phone"}

	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), tenantID.String(), constants.IdentifierPhone.String(), "+862025550123").Return(&domain.UserIdentity{ID: "uid"}, nil)
	deps.kratosService.EXPECT().InitializeLoginFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any(), gomock.Nil(), gomock.Nil()).Return(&kratos.SuccessfulNativeLogin{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)

	resp, derr := ucase.ChallengeWithPhone(ctx, tenantID, inputPhone)
	require.Nil(t, derr)
	require.NotNil(t, resp)
	require.Equal(t, flow.Id, resp.FlowID)
}

func TestIntegration_ChallengeWithEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	email := "user@example.com"
	flow := &kratos.LoginFlow{Id: "flow-login-email"}

	deps.kratosService.EXPECT().InitializeLoginFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any(), gomock.Nil(), gomock.Nil()).Return(&kratos.SuccessfulNativeLogin{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)

	resp, derr := ucase.ChallengeWithEmail(ctx, tenantID, email)
	require.Nil(t, derr)
	require.NotNil(t, resp)
	require.Equal(t, flow.Id, resp.FlowID)
}

func TestIntegration_ChallengeWithEmail_RateLimitAndSubmitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Rate limit error
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	resp, derr := ucase.ChallengeWithEmail(context.Background(), tenantID, "user@example.com")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// Submit error
	flow := &kratos.LoginFlow{Id: "flow-login-email-err"}
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.kratosService.EXPECT().InitializeLoginFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any(), gomock.Nil(), gomock.Nil()).Return(nil, assert.AnError)
	resp, derr = ucase.ChallengeWithEmail(context.Background(), tenantID, "user@example.com")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_VerifyLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	flowID := "flow-login-verify"
	loginFlow := &kratos.LoginFlow{Id: flowID}

	deps.kratosService.EXPECT().GetLoginFlow(gomock.Any(), tenantID, flowID).Return(loginFlow, nil)
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), flowID).Return(&domain.ChallengeSession{Identifier: "user@example.com"}, nil)

	// Build successful login result
	loginResult := &kratos.SuccessfulNativeLogin{
		Session: kratos.Session{
			Id:                    "sess-login",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: "tenant-user-login", Traits: map[string]interface{}{string(constants.IdentifierEmail): "user@example.com"}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token-login"),
	}

	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, loginFlow, gomock.Eq("code"), gomock.Any(), gomock.Nil(), gomock.Any()).Return(loginResult, nil)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), flowID).Return(nil)

	resp, derr := ucase.VerifyLogin(ctx, tenantID, flowID, "123456")
	require.Nil(t, derr)
	require.NotNil(t, resp)
	require.Equal(t, "tenant-user-login", resp.User.ID)
}

func TestIntegration_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	ctx := context.WithValue(context.Background(), constants.SessionTokenKey, "access-token")
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.kratosService.EXPECT().GetSession(gomock.Any(), tenantID, "access-token").Return(&kratos.Session{Active: ptr(true)}, nil)
	deps.kratosService.EXPECT().Logout(gomock.Any(), tenantID, "access-token").Return(nil)

	derr := ucase.Logout(ctx, tenantID)
	require.Nil(t, derr)
}

func TestIntegration_RefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	session := &kratos.Session{
		Id:                    "sess-refresh",
		Active:                ptr(true),
		ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
		IssuedAt:              ptr(time.Now()),
		AuthenticatedAt:       ptr(time.Now()),
		Identity:              &kratos.Identity{Id: "tenant-user-refresh", Traits: map[string]interface{}{string(constants.IdentifierEmail): "user@example.com"}},
		AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
	}
	deps.kratosService.EXPECT().GetSession(gomock.Any(), tenantID, "access-token").Return(session, nil)

	resp, derr := ucase.RefreshToken(context.Background(), tenantID, "access-token", "refresh-token")
	require.Nil(t, derr)
	require.NotNil(t, resp)
	require.Equal(t, "tenant-user-refresh", resp.User.ID)
}

func TestIntegration_Profile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ctx := context.WithValue(context.Background(), constants.SessionTokenKey, "token")

	traits := map[string]interface{}{string(constants.IdentifierEmail): "user@example.com", string(constants.IdentifierUsername): "u1"}
	session := &kratos.Session{
		Active:   ptr(true),
		Identity: &kratos.Identity{Id: "tenant-user-profile", Traits: traits},
	}
	deps.kratosService.EXPECT().WhoAmI(gomock.Any(), tenantID, "token").Return(session, nil)
	deps.userIdentityRepo.EXPECT().FindGlobalUserIDByIdentity(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user@example.com").Return("global-user", nil)
	deps.userIdentityRepo.
		EXPECT().
		GetByGlobalUserID(gomock.Any(), nil, tenantID.String(), "global-user").
		Return([]domain.UserIdentity{
			{Type: constants.IdentifierEmail.String(), Value: "user@example.com"},
		}, nil)

	user, derr := ucase.Profile(ctx, tenantID)
	require.Nil(t, derr)
	require.NotNil(t, user)
	require.Equal(t, "tenant-user-profile", user.ID)
	require.Equal(t, "global-user", user.GlobalUserID)
}

func TestIntegration_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	username := "user1"
	password := "pass1"
	flow := &kratos.LoginFlow{Id: "flow-login-password"}

	deps.kratosService.EXPECT().InitializeLoginFlow(gomock.Any(), tenantID).Return(flow, nil)
	loginResult := &kratos.SuccessfulNativeLogin{
		Session: kratos.Session{
			Id:                    "sess-login-pw",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: "tenant-user-login-pw", Traits: map[string]interface{}{string(constants.IdentifierUsername): username}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("password")}},
		},
		SessionToken: ptr("token-login-pw"),
	}
	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, flow, gomock.Eq("password"), gomock.Any(), gomock.Any(), gomock.Nil()).Return(loginResult, nil)

	resp, derr := ucase.Login(context.Background(), tenantID, username, password)
	require.Nil(t, derr)
	require.NotNil(t, resp)
	require.Equal(t, "tenant-user-login-pw", resp.User.ID)
}

// =========================
// CheckIdentifier coverage
// =========================
func TestIntegration_CheckIdentifier_Email_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user@example.com").Return(true, nil)

	ok, idType, derr := ucase.CheckIdentifier(context.Background(), tenantID, "user@example.com")
	require.Nil(t, derr)
	require.True(t, ok)
	require.Equal(t, constants.IdentifierEmail.String(), idType)
}

func TestIntegration_CheckIdentifier_Phone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// normalized E.164 expected by repo; default region normalization yields +862025550123 for this input
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierPhone.String(), "+862025550123").Return(false, nil)

	ok, idType, derr := ucase.CheckIdentifier(context.Background(), tenantID, "+862025550123")
	require.Nil(t, derr)
	require.False(t, ok)
	require.Equal(t, constants.IdentifierPhone.String(), idType)
}

func TestIntegration_CheckIdentifier_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user@example.com").Return(false, assert.AnError)

	ok, idType, derr := ucase.CheckIdentifier(context.Background(), tenantID, "user@example.com")
	require.NotNil(t, derr)
	require.False(t, ok)
	require.Equal(t, constants.IdentifierEmail.String(), idType)
}

// =========================
// ChangeIdentifier - validation failures
// =========================
func TestIntegration_ChangeIdentifier_ValidationFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, _ := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// invalid types
	resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, "t1", "bad")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// empty identifier
	resp, derr = ucase.ChangeIdentifier(ctx, "g1", tenantID, "t1", "")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// invalid email
	resp, derr = ucase.ChangeIdentifier(ctx, "g1", tenantID, "t1", "not-an-email")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// invalid phone
	resp, derr = ucase.ChangeIdentifier(ctx, "g1", tenantID, "t1", "12345")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

// =========================
// ChangeIdentifier - conflict and state failures
// =========================
func TestIntegration_ChangeIdentifier_ConflictFailures(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantUserID := "tenant-user-1"

	t.Run("identifier already exists in tenant", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ucase, deps := buildIsolatedUseCase(ctrl)
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		ucase.db = db
		ctx := context.Background()
		deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "dup@example.com").Return(true, nil)
		resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, tenantUserID, "dup@example.com")
		require.NotNil(t, derr)
		require.Nil(t, resp)
	})

	t.Run("no identifiers for user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ucase, deps := buildIsolatedUseCase(ctrl)
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		ucase.db = db
		ctx := context.Background()
		deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "new@example.com").Return(false, nil)
		deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), tenantID.String(), tenantUserID).Return([]*domain.UserIdentity{}, nil)
		resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, tenantUserID, "new@example.com")
		require.NotNil(t, derr)
		require.Nil(t, resp)
	})

	t.Run("cross-type change when multiple identifiers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ucase, deps := buildIsolatedUseCase(ctrl)
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		ucase.db = db
		ctx := context.Background()
		deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "another@example.com").Return(false, nil)
		deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), tenantID.String(), tenantUserID).Return([]*domain.UserIdentity{{ID: "id-email", Type: constants.IdentifierEmail.String()}, {ID: "id-phone", Type: constants.IdentifierPhone.String()}}, nil)
		// Expect Kratos flow initialization and submission since same-type replace is allowed
		flow := &kratos.RegistrationFlow{Id: "flow-cross-same"}
		deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
		deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
		deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
		deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)
		resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, tenantUserID, "another@example.com")
		require.Nil(t, derr)
		require.NotNil(t, resp)
	})

	t.Run("single identifier different type allows switch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ucase, deps := buildIsolatedUseCase(ctrl)
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		ucase.db = db
		ctx := context.Background()
		deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "brandnew@example.com").Return(false, nil)
		deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), tenantID.String(), tenantUserID).Return([]*domain.UserIdentity{{ID: "id-phone", Type: constants.IdentifierPhone.String()}}, nil)
		flow := &kratos.RegistrationFlow{Id: "flow-single-switch"}
		deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
		deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
		deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
		deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)
		resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, tenantUserID, "brandnew@example.com")
		require.Nil(t, derr)
		require.NotNil(t, resp)
	})

	// Add the opposite direction: only email exists -> switch to phone
	t.Run("single identifier different type allows switch (email->phone)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ucase, deps := buildIsolatedUseCase(ctrl)
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		ucase.db = db
		ctx := context.Background()
		deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		newPhone := "+862025550123"
		deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierPhone.String(), newPhone).Return(false, nil)
		deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), tenantID.String(), tenantUserID).Return([]*domain.UserIdentity{{ID: "id-email", Type: constants.IdentifierEmail.String()}}, nil)
		flow := &kratos.RegistrationFlow{Id: "flow-single-switch2"}
		deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
		deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
		deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
		deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)
		resp, derr := ucase.ChangeIdentifier(ctx, "g1", tenantID, tenantUserID, newPhone)
		require.Nil(t, derr)
		require.NotNil(t, resp)
	})
}

// =========================
// VerifyRegister(change) error paths
// =========================
func TestIntegration_VerifyRegister_Change_GetChallengeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), "flow-err").Return(nil, assert.AnError)

	resp, derr := ucase.VerifyRegister(context.Background(), tenantID, "flow-err", "000000")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_VerifyRegister_Change_GetFlowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), "flow-err2").Return(&domain.ChallengeSession{Identifier: "user@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeChangeIdentifier}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, "flow-err2").Return(nil, assert.AnError)

	resp, derr := ucase.VerifyRegister(context.Background(), tenantID, "flow-err2", "000000")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_VerifyRegister_Change_TenantByNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	flow := &kratos.RegistrationFlow{Id: "flow-err3"}
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), "flow-err3").Return(&domain.ChallengeSession{Identifier: "user@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeChangeIdentifier}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, "flow-err3").Return(flow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{Session: &kratos.Session{Identity: &kratos.Identity{Id: "tu", Traits: map[string]interface{}{"tenant": "tenant-name"}}}, SessionToken: ptr("tok")}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, flow, gomock.Any()).Return(verifyResult, nil)
	deps.tenantRepo.EXPECT().GetByName("tenant-name").Return(nil, assert.AnError)

	resp, derr := ucase.VerifyRegister(context.Background(), tenantID, "flow-err3", "000000")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_ChallengeWithPhone_Negatives(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Rate-limit error
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant"}, nil)
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	resp, derr := ucase.ChallengeWithPhone(context.Background(), tenantID, "+862025550123")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// Identity not found
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant"}, nil)
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), tenantID.String(), constants.IdentifierPhone.String(), "+862025550123").Return(nil, assert.AnError)
	resp, derr = ucase.ChallengeWithPhone(context.Background(), tenantID, "+862025550123")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_VerifyRegister_InvalidTraits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	flow := &kratos.RegistrationFlow{Id: "flow-bad-traits"}
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), flow.Id).Return(&domain.ChallengeSession{Identifier: "user@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeRegister}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, flow.Id).Return(flow, nil)
	// return traits as non-map to trigger invalid traits branch
	verifyResult := &kratos.SuccessfulNativeRegistration{Session: &kratos.Session{Identity: &kratos.Identity{Id: "tu-bad", Traits: "not-a-map"}}, SessionToken: ptr("tok")}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, flow, gomock.Any()).Return(verifyResult, nil)
	resp, derr := ucase.VerifyRegister(context.Background(), tenantID, flow.Id, "000000")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_Register_Negatives(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Exists conflict
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user@example.com").Return(true, nil)
	resp, derr := ucase.Register(context.Background(), tenantID, "en", "user@example.com", "")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// Init flow error
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user2@example.com").Return(false, nil)
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(nil, assert.AnError)
	resp, derr = ucase.Register(context.Background(), tenantID, "en", "user2@example.com", "")
	require.NotNil(t, derr)
	require.Nil(t, resp)

	// Save challenge error
	flow := &kratos.RegistrationFlow{Id: "flow-save-err"}
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "user3@example.com").Return(false, nil)
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(assert.AnError)
	resp, derr = ucase.Register(context.Background(), tenantID, "en", "user3@example.com", "")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_BindIAMToRegistration_Failures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ten := &domain.Tenant{ID: tenantID, Name: "tenant-name"}

	// Trigger registration then force identity insert failure
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "newuser@example.com").Return(false, nil)
	flow := &kratos.RegistrationFlow{Id: "flow-bindreg-fail"}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(ten, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)

	regResp, regErr := ucase.Register(ctx, tenantID, "en", "newuser@example.com", "")
	require.Nil(t, regErr)
	require.True(t, regResp.VerificationNeeded)

	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), flow.Id).Return(&domain.ChallengeSession{Identifier: "newuser@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeRegister}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, flow.Id).Return(flow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{Session: &kratos.Session{Identity: &kratos.Identity{Id: "tu-bindreg", Traits: map[string]interface{}{"tenant": ten.Name, string(constants.IdentifierEmail): "newuser@example.com"}}}, SessionToken: ptr("tok")}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, flow, gomock.Any()).Return(verifyResult, nil)
	deps.tenantRepo.EXPECT().GetByName(ten.Name).Return(ten, nil)
	// Fail identity insert
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "newuser@example.com").Return(nil, assert.AnError)
	deps.globalUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	deps.userIdentityRepo.EXPECT().InsertOnceByTenantUserAndType(gomock.Any(), gomock.Any(), tenantID.String(), gomock.Any(), constants.IdentifierEmail.String(), "newuser@example.com").Return(false, assert.AnError)

	authResp, authErr := ucase.VerifyRegister(ctx, tenantID, flow.Id, "000000")
	require.NotNil(t, authErr)
	require.Nil(t, authResp)
}

func TestIntegration_VerifyRegister_Register_SubmitCodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	flow := &kratos.RegistrationFlow{Id: "flow-reg-submit-err"}
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), flow.Id).Return(&domain.ChallengeSession{Identifier: "user@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeRegister}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, flow.Id).Return(flow, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, flow, gomock.Any()).Return(nil, assert.AnError)
	resp, derr := ucase.VerifyRegister(context.Background(), tenantID, flow.Id, "000000")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_Register_SubmitRegistrationFlow_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "submiterr@example.com").Return(false, nil)
	flow := &kratos.RegistrationFlow{Id: "flow-reg-submit"}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant-name"}, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(nil, assert.AnError)
	resp, derr := ucase.Register(context.Background(), tenantID, "en", "submiterr@example.com", "")
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_ChallengeWithPhone_SubmitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	phone := "+862025550123"
	flow := &kratos.LoginFlow{Id: "flow-login-phone-err"}

	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(&domain.Tenant{ID: tenantID, Name: "tenant"}, nil)
	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), tenantID.String(), constants.IdentifierPhone.String(), phone).Return(&domain.UserIdentity{ID: "uid"}, nil)
	deps.kratosService.EXPECT().InitializeLoginFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.kratosService.EXPECT().SubmitLoginFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any(), gomock.Nil(), gomock.Nil()).Return(nil, assert.AnError)
	resp, derr := ucase.ChallengeWithPhone(context.Background(), tenantID, phone)
	require.NotNil(t, derr)
	require.Nil(t, resp)
}

func TestIntegration_BindIAMToRegistration_MappingCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	ucase.db = db
	ctx := context.Background()
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ten := &domain.Tenant{ID: tenantID, Name: "tenant-name"}

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "merr@example.com").Return(false, nil)
	flow := &kratos.RegistrationFlow{Id: "flow-bind-map-err"}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), tenantID).Return(flow, nil)
	deps.tenantRepo.EXPECT().GetByID(tenantID).Return(ten, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), tenantID, flow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), flow.Id, gomock.Any(), gomock.Any()).Return(nil)
	_, derr := ucase.Register(ctx, tenantID, "en", "merr@example.com", "")
	require.Nil(t, derr)

	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), flow.Id).Return(&domain.ChallengeSession{Identifier: "merr@example.com", IdentifierType: constants.IdentifierEmail.String(), ChallengeType: constants.ChallengeTypeRegister}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), tenantID, flow.Id).Return(flow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{Session: &kratos.Session{Identity: &kratos.Identity{Id: "tu-map-err", Traits: map[string]interface{}{"tenant": ten.Name, string(constants.IdentifierEmail): "merr@example.com"}}}, SessionToken: ptr("tok")}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), tenantID, flow, gomock.Any()).Return(verifyResult, nil)
	deps.tenantRepo.EXPECT().GetByName(ten.Name).Return(ten, nil)
	deps.userIdentityRepo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), tenantID.String(), constants.IdentifierEmail.String(), "merr@example.com").Return(nil, assert.AnError)
	deps.globalUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	deps.userIdentityRepo.EXPECT().InsertOnceByTenantUserAndType(gomock.Any(), gomock.Any(), tenantID.String(), gomock.Any(), constants.IdentifierEmail.String(), "merr@example.com").Return(true, nil)
	deps.userIdentifierMappingRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(assert.AnError)
	authResp, authErr := ucase.VerifyRegister(ctx, tenantID, flow.Id, "000000")
	require.NotNil(t, authErr)
	require.Nil(t, authResp)
}

func TestIntegration_Logout_InactiveSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	ctx := context.WithValue(context.Background(), constants.SessionTokenKey, "access-token")
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	deps.kratosService.EXPECT().GetSession(gomock.Any(), tenantID, "access-token").Return(&kratos.Session{Active: ptr(false)}, nil)
	derr := ucase.Logout(ctx, tenantID)
	require.NotNil(t, derr)
}

func TestIntegration_Profile_WhoAmI_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ucase, deps := buildIsolatedUseCase(ctrl)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ctx := context.WithValue(context.Background(), constants.SessionTokenKey, "token")
	deps.kratosService.EXPECT().WhoAmI(gomock.Any(), tenantID, "token").Return(nil, assert.AnError)
	user, derr := ucase.Profile(ctx, tenantID)
	require.NotNil(t, derr)
	require.Nil(t, user)
}
