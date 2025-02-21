package wire

import (
	"gorm.io/gorm"

	infrainterfaces "github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	repositories "github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	repositories_interfaces "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	ucases "github.com/genefriendway/human-network-iam/internal/domain/ucases"
	ucases_interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
)

// Struct to hold all repositories
type repos struct {
	IdentityOrganizationRepo repositories_interfaces.IdentityOrganizationRepository
	IdentityUserRepo         repositories_interfaces.IdentityUserRepository
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
	IdentityOrganizationUCase ucases_interfaces.IdentityOrganizationUseCase
	IdentityUserUCase         ucases_interfaces.IdentityUserUseCase
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
