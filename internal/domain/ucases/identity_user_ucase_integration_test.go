package ucases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
}) {
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

	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String(), data.newIdentifier, constants.IdentifierEmail.String())
	assert.Nil(t, changeErr)
	assert.NotNil(t, changeResp)

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
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)
	assert.Equal(t, newTenantUserID, verifyResp.User.ID)
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

	changeResp, changeErr := ucase.ChangeIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String(), data.newIdentifier, constants.IdentifierEmail.String())
	assert.Nil(t, changeErr)
	assert.NotNil(t, changeResp)

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
	assert.Error(t, verifyErr)
	assert.Nil(t, verifyResp)
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

	// AddNewIdentifier expectations
	deps.userIdentityRepo.EXPECT().ExistsWithinTenant(gomock.Any(), data.tenantID.String(), constants.IdentifierEmail.String(), data.newIdentifier).Return(false, nil)
	deps.userIdentityRepo.EXPECT().ExistsByTenantGlobalUserIDAndType(gomock.Any(), data.tenantID.String(), data.globalUserID, constants.IdentifierEmail.String()).Return(false, nil)
	regFlow := &kratos.RegistrationFlow{Id: data.flowID}
	deps.kratosService.EXPECT().InitializeRegistrationFlow(gomock.Any(), data.tenantID).Return(regFlow, nil)
	deps.tenantRepo.EXPECT().GetByID(data.tenantID).Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	deps.kratosService.EXPECT().SubmitRegistrationFlow(gomock.Any(), data.tenantID, regFlow, gomock.Eq("code"), gomock.Any()).Return(&kratos.SuccessfulNativeRegistration{}, nil)
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), data.flowID, gomock.Any(), gomock.Any()).Return(nil)

	addResp, addErr := ucase.AddNewIdentifier(ctx, data.tenantID, data.globalUserID, data.newIdentifier, constants.IdentifierEmail.String())
	assert.Nil(t, addErr)
	assert.NotNil(t, addResp)

	// VerifyRegister expectations for AddIdentifier path
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), data.flowID).Return(&domain.ChallengeSession{
		GlobalUserID:   data.globalUserID,
		Identifier:     data.newIdentifier,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeAddIdentifier,
	}, nil)
	deps.kratosService.EXPECT().GetRegistrationFlow(gomock.Any(), data.tenantID, data.flowID).Return(regFlow, nil)
	verifyResult := &kratos.SuccessfulNativeRegistration{
		Session: &kratos.Session{
			Id:                    "session-id-add",
			Active:                ptr(true),
			ExpiresAt:             ptr(time.Now().Add(30 * time.Minute)),
			IssuedAt:              ptr(time.Now()),
			AuthenticatedAt:       ptr(time.Now()),
			Identity:              &kratos.Identity{Id: "tenant-user-add", Traits: map[string]interface{}{"tenant": "tenant-name", string(constants.IdentifierEmail): data.newIdentifier}},
			AuthenticationMethods: []kratos.SessionAuthenticationMethod{{Method: ptr("code")}},
		},
		SessionToken: ptr("token-add"),
	}
	deps.kratosService.EXPECT().SubmitRegistrationFlowWithCode(gomock.Any(), data.tenantID, regFlow, gomock.Any()).Return(verifyResult, nil)
	deps.tenantRepo.EXPECT().GetByName("tenant-name").Return(&domain.Tenant{ID: data.tenantID, Name: "tenant-name"}, nil)
	deps.userIdentityRepo.EXPECT().InsertOnceByTenantUserAndType(gomock.Any(), gomock.Any(), data.tenantID.String(), data.globalUserID, constants.IdentifierEmail.String(), data.newIdentifier).Return(true, nil)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), data.flowID).Return(nil)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, data.tenantID, data.flowID, "654321")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)
	assert.Equal(t, "tenant-user-add", verifyResp.User.ID)
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
	assert.NotNil(t, derr)

	// Case 2: delete when there are two identifiers
	deps.userIdentityRepo.EXPECT().ListByTenantAndTenantUserID(gomock.Any(), gomock.Any(), data.tenantID.String(), data.tenantUserID).Return([]*domain.UserIdentity{
		{ID: "id-email", Type: constants.IdentifierEmail.String()},
		{ID: "id-phone", Type: constants.IdentifierPhone.String()},
	}, nil)
	deps.userIdentityRepo.EXPECT().Delete(gomock.Any(), "id-email").Return(nil)
	deps.kratosService.EXPECT().DeleteIdentifierAdmin(gomock.Any(), data.tenantID, gomock.Any()).Return(nil)
	derr = ucase.DeleteIdentifier(ctx, data.globalUserID, data.tenantID, data.tenantUserID, constants.IdentifierEmail.String())
	assert.Nil(t, derr)
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
	assert.Nil(t, regErr)
	assert.NotNil(t, regResp)
	assert.True(t, regResp.VerificationNeeded)
	assert.Equal(t, data.flowID, regResp.VerificationFlow.FlowID)

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
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)
	assert.Equal(t, "tenant-user-reg", verifyResp.User.ID)
}
