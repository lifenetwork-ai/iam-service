package wire

import (
	"context"

	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	keto "github.com/lifenetwork-ai/iam-service/internal/adapters/services/keto"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

// Struct to hold all repositories
type Repos struct {
	ChallengeSessionRepo      domainrepo.ChallengeSessionRepository
	GlobalUserRepo            domainrepo.GlobalUserRepository
	UserIdentityRepo          domainrepo.UserIdentityRepository
	UserIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	TenantRepo                domainrepo.TenantRepository
	AdminAccountRepo          domainrepo.AdminAccountRepository
	ZaloTokenRepo             domainrepo.ZaloTokenRepository
	CacheRepo                 types.CacheRepository
}

// Initialize repositories (only using cache where needed)
func InitializeRepos(db *gorm.DB, cacheRepo types.CacheRepository) *Repos {
	// Return all repositories
	return &Repos{
		CacheRepo:                 cacheRepo,
		ChallengeSessionRepo:      repositories.NewChallengeSessionRepository(cacheRepo),
		GlobalUserRepo:            repositories.NewGlobalUserRepository(db),
		UserIdentityRepo:          repositories.NewUserIdentityRepository(db),
		UserIdentifierMappingRepo: repositories.NewUserIdentifierMappingRepository(db),
		TenantRepo: repositories.NewTenantRepositoryCache(
			repositories.NewTenantRepository(db), cacheRepo,
		),
		AdminAccountRepo: repositories.NewAdminAccountRepository(db),
		ZaloTokenRepo:    repositories.NewZaloTokenRepositoryCache(repositories.NewZaloTokenRepository(db), cacheRepo),
	}
}

// Struct to hold all use cases
type UseCases struct {
	IdentityUserUCase interfaces.IdentityUserUseCase
	AdminUCase        interfaces.AdminUseCase
	TenantUCase       interfaces.TenantUseCase
	PermissionUCase   interfaces.PermissionUseCase
	CourierUCase      interfaces.CourierUseCase
	SmsTokenUCase     interfaces.SmsTokenUseCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, repos *Repos) *UseCases {
	// Return all use cases
	return &UseCases{
		IdentityUserUCase: ucases.NewIdentityUserUseCase(
			db,
			instances.RateLimiterInstance(),
			repos.ChallengeSessionRepo,
			repos.TenantRepo,
			repos.GlobalUserRepo,
			repos.UserIdentityRepo,
			repos.UserIdentifierMappingRepo,
			instances.KratosServiceInstance(repos.TenantRepo),
		),
		AdminUCase: ucases.NewAdminUseCase(
			repos.TenantRepo,
			repos.AdminAccountRepo,
			repos.UserIdentityRepo,
			repos.UserIdentifierMappingRepo,
			instances.KratosServiceInstance(repos.TenantRepo),
		),
		TenantUCase:     ucases.NewTenantUseCase(repos.TenantRepo),
		PermissionUCase: ucases.NewPermissionUseCase(keto.NewKetoService(repos.TenantRepo), repos.UserIdentityRepo),
		CourierUCase:    ucases.NewCourierUseCase(instances.OTPQueueRepositoryInstance(context.Background()), instances.SMSServiceInstance(repos.ZaloTokenRepo), repos.CacheRepo),
		SmsTokenUCase:   ucases.NewSmsTokenUseCase(repos.ZaloTokenRepo),
	}
}
