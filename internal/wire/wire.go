package wire

import (
	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

// Struct to hold all repositories
type repos struct {
	ChallengeSessionRepo      repotypes.ChallengeSessionRepository
	GlobalUserRepo            repotypes.GlobalUserRepository
	UserIdentityRepo          repotypes.UserIdentityRepository
	UserIdentifierMappingRepo repotypes.UserIdentifierMappingRepository
	TenantRepo                repotypes.TenantRepository
	AdminAccountRepo          repotypes.AdminAccountRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, cacheRepo types.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		ChallengeSessionRepo:      repositories.NewChallengeSessionRepository(cacheRepo),
		GlobalUserRepo:            repositories.NewGlobalUserRepository(db),
		UserIdentityRepo:          repositories.NewUserIdentityRepository(db),
		UserIdentifierMappingRepo: repositories.NewUserIdentifierMappingRepository(db),
		TenantRepo: repositories.NewTenantRepositoryCache(
			repositories.NewTenantRepository(db), cacheRepo,
		),
		AdminAccountRepo: repositories.NewAdminAccountRepository(db),
	}
}

// Struct to hold all use cases
type UseCases struct {
	IdentityUserUCase ucasetypes.IdentityUserUseCase
	AdminUCase        ucasetypes.AdminUseCase
	TenantUCase       ucasetypes.TenantUseCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, cacheRepo types.CacheRepository) *UseCases {
	repos := initializeRepos(db, cacheRepo)

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
			kratos.NewKratosService(repos.TenantRepo),
		),
		AdminUCase:  ucases.NewAdminUseCase(repos.TenantRepo, repos.AdminAccountRepo),
		TenantUCase: ucases.NewTenantUseCase(repos.TenantRepo),
	}
}
