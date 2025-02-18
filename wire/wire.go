package wire

import (
	infrainterfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	"github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	ucases "github.com/genefriendway/human-network-iam/internal/ucases"
	"gorm.io/gorm"
)

// Struct to hold all repositories
type repos struct {
	IdentityOrganizationRepo interfaces.IdentityOrganizationRepository
	IdentityUserRepo         interfaces.IdentityUserRepository
}

// Initialize repositories (only using cache where needed)
func initializeRepos(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *repos {
	// Return all repositories
	return &repos{
		IdentityOrganizationRepo: repositories.NewIdentityOrganizationRepository(db, cacheRepo),
		IdentityUserRepo:         repositories.NewIdentityUserRepository(db, cacheRepo),
	}
}

// Struct to hold all use cases
type UseCases struct {
	IdentityOrganizationUCase interfaces.IdentityOrganizationUseCase
	IdentityUserUCase         interfaces.IdentityUserUseCase
}

// Initialize use cases
func InitializeUseCases(db *gorm.DB, cacheRepo infrainterfaces.CacheRepository) *UseCases {
	repos := initializeRepos(db, cacheRepo)

	// Return all use cases
	return &UseCases{
		IdentityOrganizationUCase: ucases.NewIdentityOrganizationUseCase(repos.IdentityOrganizationRepo),
		IdentityUserUCase:         ucases.NewIdentityUserUseCase(repos.IdentityUserRepo),
	}
}
