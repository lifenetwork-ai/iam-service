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
	"github.com/stretchr/testify/require"
)

func buildUseCaseWithSQLiteRepos(ctrl *gomock.Controller) (*userUseCase, struct {
	tenantRepo                domainrepo.TenantRepository
	globalUserRepo            domainrepo.GlobalUserRepository
	userIdentityRepo          domainrepo.UserIdentityRepository
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	challengeSessionRepo      domainrepo.ChallengeSessionRepository
	kratosService             domainservice.KratosService
	rateLimiter               *mock_types.MockRateLimiter
}, *gorm.DB,
) {
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
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identities (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, kratos_user_id TEXT, type TEXT NOT NULL, value TEXT NOT NULL, created_at datetime, updated_at datetime)").Error
	_ = db.Exec("CREATE TABLE IF NOT EXISTS user_identifier_mapping (id TEXT PRIMARY KEY, global_user_id TEXT NOT NULL, lang TEXT DEFAULT '', created_at datetime, updated_at datetime)").Error
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
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	fmt.Println("verifyRegisterResp", verifyRegisterResp)
	fmt.Println("verifyErr", verifyErr)
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID = verifyRegisterResp.User.GlobalUserID

	// Act: add an email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, oldEmail, constants.IdentifierEmail.String())
	require.Nil(t, addEmailErr)
	require.NotNil(t, addEmailResp)

	verifyAddEmailResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyAddEmailResp)

	// Act: change identifier phone→email
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newEmail)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure the old email identity is deleted on IAM side
	identities, ucaseErr := ucase.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 2, len(identities))
	require.Equal(t, constants.IdentifierPhone.String(), identities[0].Type)
	require.Equal(t, phone, identities[0].Value)
	require.Equal(t, constants.IdentifierEmail.String(), identities[1].Type)
	require.Equal(t, newEmail, identities[1].Value)

	//  query kratos, ensure the old email identity is deleted on Kratos side
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	new, ok := ids[newEmail]
	require.True(t, ok)
	require.Equal(t, newEmail, new.Traits.(map[string]interface{})[constants.IdentifierEmail.String()])
	_, ok = ids[oldEmail]
	require.False(t, ok)

	// Check: login with new email, old phone works, new email fails
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})
	t.Run("login with old email fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, oldEmail)
		require.NotNil(t, err)
		require.Nil(t, loginResp)
		require.Contains(t, err.Error(), "email not found")
	})
	t.Run("login with phone still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, phone)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
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

	var globalUserID string
	oldPhone := "+84345381013"
	newPhone := "+84345381014"
	email := "oldemail@test.com"

	// Seed DB state
	_ = db.Create(&domain.Tenant{ID: tenantID, Name: "tenant-name"}).Error

	// Register a phone identity
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", oldPhone)
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID = verifyRegisterResp.User.GlobalUserID

	// Act: add an email identity
	addEmailResp, addEmailErr := ucase.AddNewIdentifier(ctx, tenantID, globalUserID, email, constants.IdentifierEmail.String())
	require.Nil(t, addEmailErr)
	require.NotNil(t, addEmailResp)

	verifyAddEmailResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, addEmailResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyAddEmailResp)

	// Act: change identifier phone→phone
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newPhone)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure the old phone identity is deleted on IAM side
	identities, ucaseErr := ucase.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 2, len(identities))
	require.Equal(t, constants.IdentifierPhone.String(), identities[0].Type)
	require.Equal(t, newPhone, identities[0].Value)
	require.Equal(t, constants.IdentifierEmail.String(), identities[1].Type)
	require.Equal(t, email, identities[1].Value)

	//  query kratos, ensure the old phone identity is deleted on Kratos side
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	new, ok := ids[newPhone]
	require.True(t, ok)
	require.Equal(t, newPhone, new.Traits.(map[string]interface{})[constants.IdentifierPhone.String()])
	_, ok = ids[oldPhone]
	require.False(t, ok)

	// Check: login with new phone, old phone fails, email still works
	t.Run("login with new phone works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, newPhone)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})

	t.Run("login with old phone fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
		require.NotNil(t, err)
		require.Nil(t, loginResp)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("login with email still works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, email)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})
}

func TestIntegration_ChangeIdentifier_PhoneToEmail_LoginChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ucase, deps, db := buildUseCaseWithSQLiteRepos(ctrl)
	ctx := context.Background()

	deps.rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	deps.rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	oldPhone := "+84345381013"
	newEmail := "newemail@test.com"

	// Seed DB state
	_ = db.Create(&domain.Tenant{ID: tenantID, Name: "tenant-name"}).Error

	// Register a phone identity (only phone exists initially)
	registerResp, registerErr := ucase.Register(ctx, tenantID, "en", "", oldPhone)
	require.Nil(t, registerErr)
	require.NotNil(t, registerResp)

	verifyRegisterResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, registerResp.VerificationFlow.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyRegisterResp)
	globalUserID := verifyRegisterResp.User.GlobalUserID

	// Act: change identifier phone→email (since only one identifier exists, it should replace it)
	changeResp, changeErr := ucase.ChangeIdentifier(ctx, globalUserID, tenantID, verifyRegisterResp.User.ID, newEmail)
	require.Nil(t, changeErr)
	require.NotNil(t, changeResp)

	verifyChangeResp, verifyErr := ucase.VerifyRegister(ctx, tenantID, changeResp.FlowID, "000000")
	require.Nil(t, verifyErr)
	require.NotNil(t, verifyChangeResp)

	// Query identities in IAM, ensure there is ONLY the new email identity (old phone removed)
	identities, ucaseErr := ucase.userIdentityRepo.GetByGlobalUserIDAndTenantID(ctx, nil, globalUserID, tenantID.String())
	require.Nil(t, ucaseErr)
	require.NotNil(t, identities)
	require.Equal(t, 1, len(identities))
	require.Equal(t, constants.IdentifierEmail.String(), identities[0].Type)
	require.Equal(t, newEmail, identities[0].Value)

	// Query Kratos, ensure old phone is deleted and new email exists
	servc := deps.kratosService.(*kratos_service.FakeKratosService)
	ids, _ := servc.GetIdentities(ctx, tenantID)
	newId, ok := ids[newEmail]
	require.True(t, ok)
	require.Equal(t, newEmail, newId.Traits.(map[string]interface{})[constants.IdentifierEmail.String()])
	_, ok = ids[oldPhone]
	require.False(t, ok)

	// Check: login with new email works, old phone fails
	t.Run("login with new email works", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithEmail(ctx, tenantID, newEmail)
		require.Nil(t, err)
		require.NotNil(t, loginResp)
	})

	t.Run("login with old phone fails", func(t *testing.T) {
		loginResp, err := ucase.ChallengeWithPhone(ctx, tenantID, oldPhone)
		require.NotNil(t, err)
		require.Nil(t, loginResp)
		require.Contains(t, err.Error(), "not found")
	})
}
