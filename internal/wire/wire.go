package wire

import (
	"gorm.io/gorm"

	infrainterfaces "github.com/lifenetwork-ai/iam-service/infrastructures/interfaces"
	repositories "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	repositories_interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services"
	ucases "github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	ucases_interfaces "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
)

// Struct to hold all repositories
type repos struct {
	IdentityOrganizationRepo repositories_interfaces.IdentityOrganizationRepository
	IdentityUserRepo         repositories_interfaces.IdentityUserRepository

	AccessSessionRepo    repositories_interfaces.AccessSessionRepository
	ChallengeSessionRepo repositories_interfaces.ChallengeSessionRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		IdentityOrganizationRepo: repositories.NewIdentityOrganizationRepository(db, cacheRepo),
		IdentityUserRepo:         repositories.NewIdentityUserRepository(db, cacheRepo),

		AccessSessionRepo:    repositories.NewAccessSessionRepository(db, cacheRepo),
		ChallengeSessionRepo: repositories.NewChallengeSessionRepository(cacheRepo),
	}
}

// Struct to hold all use cases
type UseCases struct {
	IdentityOrganizationUCase ucases_interfaces.IdentityOrganizationUseCase
	IdentityUserUCase         ucases_interfaces.IdentityUserUseCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *UseCases {
	repos := initializeRepos(db, cacheRepo)

	// Return all use cases
	return &UseCases{
		IdentityOrganizationUCase: ucases.NewIdentityOrganizationUseCase(repos.IdentityOrganizationRepo),
		IdentityUserUCase: ucases.NewIdentityUserUseCase(
			repos.ChallengeSessionRepo,
			services.NewKratosService(),
		),
	}
}
