package wire

import (
	"gorm.io/gorm"

	infrainterfaces "github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

// Struct to hold all repositories
type repos struct {
	IdentityOrganizationRepo repotypes.IdentityOrganizationRepository
	IdentityUserRepo         repotypes.IdentityUserRepository

	AccessSessionRepo    repotypes.AccessSessionRepository
	ChallengeSessionRepo repotypes.ChallengeSessionRepository

	GlobalUserRepo            repotypes.GlobalUserRepository
	UserIdentityRepo          repotypes.UserIdentityRepository
	UserIdentifierMappingRepo repotypes.UserIdentifierMappingRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		IdentityOrganizationRepo: repositories.NewIdentityOrganizationRepository(db, cacheRepo),
		IdentityUserRepo:         repositories.NewIdentityUserRepository(db, cacheRepo),

		AccessSessionRepo:    repositories.NewAccessSessionRepository(db, cacheRepo),
		ChallengeSessionRepo: repositories.NewChallengeSessionRepository(cacheRepo),

		GlobalUserRepo:            repositories.NewGlobalUserRepository(db),
		UserIdentityRepo:          repositories.NewUserIdentityRepository(db),
		UserIdentifierMappingRepo: repositories.NewUserIdentifierMappingRepository(db),
	}
}

// Struct to hold all use cases
type UseCases struct {
	IdentityOrganizationUCase ucasetypes.IdentityOrganizationUseCase
	IdentityUserUCase         ucasetypes.IdentityUserUseCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *UseCases {
	repos := initializeRepos(db, cacheRepo)

	// Return all use cases
	return &UseCases{
		IdentityOrganizationUCase: ucases.NewIdentityOrganizationUseCase(repos.IdentityOrganizationRepo),
		IdentityUserUCase: ucases.NewIdentityUserUseCase(
			db,
			repos.ChallengeSessionRepo,
			repos.GlobalUserRepo,
			repos.UserIdentityRepo,
			repos.UserIdentifierMappingRepo,
			services.NewKratosService(),
		),
	}
}
