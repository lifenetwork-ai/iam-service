package ucases

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	"github.com/stretchr/testify/assert"
)

// Builder using SQLite-backed repos (Kratos and RateLimiter remain mocks)
func buildUseCaseWithSQLiteRepos(ctrl *gomock.Controller) (*userUseCase, struct {
	tenantRepo                domainrepo.TenantRepository
	globalUserRepo            domainrepo.GlobalUserRepository
	userIdentityRepo          domainrepo.UserIdentityRepository
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	challengeSessionRepo      *mock_repositories.MockChallengeSessionRepository
	kratosService             domainservice.KratosService
	rateLimiter               *mock_types.MockRateLimiter
}, *gorm.DB) {
	deps := struct {
		tenantRepo                domainrepo.TenantRepository
		globalUserRepo            domainrepo.GlobalUserRepository
		userIdentityRepo          domainrepo.UserIdentityRepository
		userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
		challengeSessionRepo      *mock_repositories.MockChallengeSessionRepository
		kratosService             domainservice.KratosService
		rateLimiter               *mock_types.MockRateLimiter
	}{}

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	// Create SQLite-compatible schemas
	_ = db.Exec("CREATE TABLE IF NOT EXISTS tenants (id TEXT PRIMARY KEY, name TEXT, public_url TEXT, admin_url TEXT, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS global_users (id TEXT PRIMARY KEY, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identities (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, type TEXT NOT NULL, value TEXT NOT NULL, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identifier_mapping (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, tenant_user_id TEXT NOT NULL, created_at datetime, updated_at datetime)").Error

	deps.tenantRepo = adaptersrepo.NewSQLiteTenantRepository(db)
	deps.globalUserRepo = adaptersrepo.NewSQLiteGlobalUserRepository(db)
	deps.userIdentityRepo = adaptersrepo.NewSQLiteUserIdentityRepository(db)
	deps.userIdentifierMappingRepo = adaptersrepo.NewSQLiteUserIdentifierMappingRepository(db)
	deps.challengeSessionRepo = mock_repositories.NewMockChallengeSessionRepository(ctrl)
	deps.kratosService = kratos_service.NewFakeKratosService()
	deps.rateLimiter = mock_types.NewMockRateLimiter(ctrl)

	ucase := &userUseCase{
		db:                        db,
		rateLimiter:               deps.rateLimiter,
		tenantRepo:                deps.tenantRepo,
		globalUserRepo:            deps.globalUserRepo,
		userIdentityRepo:          deps.userIdentityRepo,
		userIdentifierMappingRepo: deps.userIdentifierMappingRepo,
		challengeSessionRepo:      deps.challengeSessionRepo,
		kratosService:             deps.kratosService,
	}

	return ucase, deps, db
}

func TestIntegration_ChangeIdentifier_EmailToEmail_LoginChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps, db := buildUseCaseWithSQLiteRepos(ctrl)
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	phoneIdentityID := uuid.MustParse("12223222-2222-2222-2222-222222222222").String()
	emailIdentityID := uuid.MustParse("12222222-2222-2222-2222-222222222222").String()

	kratosPhoneTenantUserID, _ := uuid.NewRandom()
	kratosEmailTenantUserID, _ := uuid.NewRandom()

	globalUserID := uuid.MustParse("11223344-1122-3344-5566-778899001122").String()
	oldEmail := "hong.vu+5c_13@genefriendway.com"
	newEmail := "hong.vu+5c_14@genefriendway.com"
	phone := "+84345381013"

	// Seed DB state
	_ = db.Create(&domain.Tenant{ID: tenantID, Name: "tenant-name"}).Error
	_ = db.Create(&domain.GlobalUser{ID: globalUserID}).Error
	_ = db.Create(&domain.UserIdentity{ID: emailIdentityID, GlobalUserID: globalUserID, TenantID: tenantID.String(), Type: constants.IdentifierEmail.String(), Value: oldEmail}).Error
	_ = db.Create(&domain.UserIdentity{ID: phoneIdentityID, GlobalUserID: globalUserID, TenantID: tenantID.String(), Type: constants.IdentifierPhone.String(), Value: phone}).Error
	_ = db.Create(&domain.UserIdentifierMapping{ID: "map-phone", GlobalUserID: globalUserID, TenantID: tenantID.String(), TenantUserID: kratosPhoneTenantUserID.String()}).Error
	_ = db.Create(&domain.UserIdentifierMapping{ID: "map-email", GlobalUserID: globalUserID, TenantID: tenantID.String(), TenantUserID: kratosEmailTenantUserID.String()}).Error

	// Seed Kratos fake with existing phone identity so phone login works
	flowSeed, _ := deps.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	_, _ = deps.kratosService.SubmitRegistrationFlow(ctx, tenantID, flowSeed, constants.MethodTypeCode.String(), map[string]interface{}{
		"tenant":                           "tenant-name",
		constants.IdentifierPhone.String(): phone,
	})

	// Find phone from user identifier
	phoneIdentity, err := ucase.userIdentityRepo.GetByTypeAndValue(ctx, nil, tenantID.String(), constants.IdentifierPhone.String(), phone)
	assert.Nil(t, err)
	assert.NotNil(t, phoneIdentity)

	// --- ChangeIdentifier setup ---
	deps.challengeSessionRepo.EXPECT().SaveChallenge(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, globalUserID, newEmail)
	assert.Nil(t, changeErr)
	assert.NotNil(t, changeResp)

	// --- VerifyRegister ---
	deps.challengeSessionRepo.EXPECT().GetChallenge(gomock.Any(), gomock.Any()).Return(&domain.ChallengeSession{
		GlobalUserID:   globalUserID,
		TenantUserID:   kratosEmailTenantUserID.String(),
		Identifier:     newEmail,
		IdentifierType: constants.IdentifierEmail.String(),
		ChallengeType:  constants.ChallengeTypeChangeIdentifier,
		IdentityID:     emailIdentityID,
	}, nil)
	deps.challengeSessionRepo.EXPECT().DeleteChallenge(gomock.Any(), gomock.Any()).Return(nil)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)

	// query the identites
	identities, err := ucase.userIdentityRepo.GetByGlobalUserID(ctx, nil, tenantID.String(), globalUserID)
	assert.Nil(t, err)
	assert.NotNil(t, identities)
	assert.Equal(t, 2, len(identities))
	assert.Equal(t, constants.IdentifierEmail.String(), identities[0].Type)
	assert.Equal(t, newEmail, identities[0].Value)
	assert.Equal(t, constants.IdentifierPhone.String(), identities[1].Type)
	assert.Equal(t, phone, identities[1].Value)
	// --- Extra: Login checks ---
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)
	})

	t.Run("login with old email fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, oldEmail)
		assert.NotNil(t, err)
		assert.Nil(t, loginResp)
		assert.Contains(t, err.Error(), "identity not found")
	})

	t.Run("login with phone still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, phone)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)
	})
}
