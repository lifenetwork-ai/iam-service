package ucases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	mock_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func buildUseCaseWithSQLiteRepos(ctrl *gomock.Controller) (*userUseCase, struct {
	tenantRepo                domainrepo.TenantRepository
	globalUserRepo            domainrepo.GlobalUserRepository
	userIdentityRepo          domainrepo.UserIdentityRepository
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	challengeSessionRepo      domainrepo.ChallengeSessionRepository
	kratosService             domainservice.KratosService
	rateLimiter               *mock_types.MockRateLimiter
}, *gorm.DB) {
	deps := struct {
		tenantRepo                domainrepo.TenantRepository
		globalUserRepo            domainrepo.GlobalUserRepository
		userIdentityRepo          domainrepo.UserIdentityRepository
		userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
		challengeSessionRepo      domainrepo.ChallengeSessionRepository
		kratosService             domainservice.KratosService
		rateLimiter               *mock_types.MockRateLimiter
	}{}

	// Use a shared in-memory SQLite database so all connections share the same schema
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	// Create SQLite-compatible schemas
	_ = db.Exec("CREATE TABLE IF NOT EXISTS tenants (id TEXT PRIMARY KEY, name TEXT, public_url TEXT, admin_url TEXT, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS global_users (id TEXT PRIMARY KEY, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identities (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, type TEXT NOT NULL, value TEXT NOT NULL, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identifier_mapping (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, tenant_user_id TEXT NOT NULL, created_at datetime, updated_at datetime)").Error
	inMemCache := caching.NewCachingRepository(context.Background(), caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)))
	deps.tenantRepo = adaptersrepo.NewSQLiteTenantRepository(db)
	deps.globalUserRepo = adaptersrepo.NewSQLiteGlobalUserRepository(db)
	deps.userIdentityRepo = adaptersrepo.NewSQLiteUserIdentityRepository(db)
	deps.userIdentifierMappingRepo = adaptersrepo.NewSQLiteUserIdentifierMappingRepository(db)
	deps.challengeSessionRepo = adaptersrepo.NewChallengeSessionRepository(inMemCache)
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

	var globalUserID string
	oldEmail := "oldemail@test.com"
	newEmail := "newemail@test.com"
	phone := "+84345381013"

	// Seed DB state
	_ = db.Create(&domain.Tenant{ID: tenantID, Name: "tenant-name"}).Error
	// Register a phone identity
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", phone)
	assert.Nil(t, registerErr)
	assert.NotNil(t, registerResp)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)
	globalUserID = verifyResp.User.GlobalUserID

	// Act: add an email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, oldEmail, constants.IdentifierEmail.String())
	assert.Nil(t, addEmailErr)
	assert.NotNil(t, addEmailResp)

	verifyResp, verifyErr = ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)

	// Act: change identifier phone→email
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, registerResp.User.ID, newEmail)
	fmt.Println("changeErr", changeErr)
	assert.Nil(t, changeErr)
	assert.NotNil(t, changeResp)

	verifyResp, verifyErr = ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	fmt.Println("verifyErr", verifyErr)
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)

	// Query identities in IAM
	identities, ucaseErr := ucase.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, tenantID.String(), globalUserID)
	assert.Nil(t, ucaseErr)
	assert.NotNil(t, identities)
	assert.Equal(t, 2, len(identities))
	assert.Equal(t, constants.IdentifierPhone.String(), identities[0].Type)
	assert.Equal(t, phone, identities[0].Value)
	assert.Equal(t, constants.IdentifierEmail.String(), identities[1].Type)
	assert.Equal(t, newEmail, identities[1].Value)

	//  query kratos
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	new, ok := ids[newEmail]
	assert.True(t, ok)
	assert.Equal(t, newEmail, new.Traits.(map[string]interface{})[constants.IdentifierEmail.String()])
	_, ok = ids[oldEmail]
	assert.False(t, ok)

	// Check: login with new email, old phone works, new email fails
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)

	})
	t.Run("login with old email fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, oldEmail)
		assert.NotNil(t, err)
		assert.Nil(t, loginResp)
		assert.Contains(t, err.Error(), "email not found")
	})
	t.Run("login with phone still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, phone)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)
	})
}

func TestIntegration_ChangeIdentifier_PhoneToPhone_LoginChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps, db := buildUseCaseWithSQLiteRepos(ctrl)
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	phoneIdentityID := uuid.MustParse("13333333-3333-3333-3333-333333333333").String()

	kratosPhoneTenantUserID, _ := uuid.NewRandom()

	globalUserID := uuid.MustParse("11223344-1122-3344-5566-778899001122").String()
	oldPhone := "+84345550001"
	newPhone := "+84345550002"
	email := "oldemail@test.com"

	// Seed DB state
	_ = db.Create(&domain.Tenant{ID: tenantID, Name: "tenant-name"}).Error
	_ = db.Create(&domain.GlobalUser{ID: globalUserID}).Error
	_ = db.Create(&domain.UserIdentity{ID: phoneIdentityID, GlobalUserID: globalUserID, TenantID: tenantID.String(), KratosUserID: kratosPhoneTenantUserID.String(), Type: constants.IdentifierPhone.String(), Value: oldPhone}).Error
	_ = db.Create(&domain.UserIdentifierMapping{ID: "map-phone", GlobalUserID: globalUserID}).Error

	// Seed Kratos fake with existing phone identity so old phone login works
	flowSeed, _ := deps.kratosService.InitializeRegistrationFlow(ctx, tenantID)
	_, _ = deps.kratosService.SubmitRegistrationFlow(ctx, tenantID, flowSeed, constants.MethodTypeCode.String(), map[string]interface{}{
		"tenant":                           "tenant-name",
		constants.IdentifierPhone.String(): oldPhone,
	})

	// Sanity check: old phone login works before change
	loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
	assert.Nil(t, err)
	assert.NotNil(t, loginResp)

	// Act: add new email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, email, constants.IdentifierEmail.String())
	assert.Nil(t, addEmailErr)
	assert.NotNil(t, addEmailResp)

	verifyResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)

	// Act: change identifier phone→phone
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, globalUserID, newPhone)
	assert.Nil(t, changeErr)
	assert.NotNil(t, changeResp)

	verifyResp, verifyErr = ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	assert.Nil(t, verifyErr)
	assert.NotNil(t, verifyResp)

	// Query identities
	identities, ucaseErr := ucase.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, tenantID.String(), globalUserID)
	assert.Nil(t, ucaseErr)
	assert.NotNil(t, identities)

	assert.Equal(t, 2, len(identities))
	// One must be phone with new value
	foundNewPhone := false
	foundEmail := false
	for _, id := range identities {
		if id.Type == constants.IdentifierPhone.String() {
			assert.Equal(t, newPhone, id.Value)
			foundNewPhone = true
		}
		if id.Type == constants.IdentifierEmail.String() {
			assert.Equal(t, email, id.Value)
			foundEmail = true
		}
	}
	assert.True(t, foundNewPhone)
	assert.True(t, foundEmail)

	// --- Extra login checks ---
	t.Run("login with new phone works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, newPhone)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)
	})

	t.Run("login with old phone fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
		assert.NotNil(t, err)
		assert.Nil(t, loginResp)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("login with email still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, email)
		assert.Nil(t, err)
		assert.NotNil(t, loginResp)
	})
}
