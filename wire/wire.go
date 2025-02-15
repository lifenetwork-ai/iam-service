//go:build wireinject

package wire

import (
	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	ucases "github.com/genefriendway/human-network-iam/internal/ucases"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var organizationUseCaseSet = wire.NewSet(
	repositories.NewIdentityOrganizationRepository,
	ucases.NewIdentityOrganizationUseCase,
)

// Init ucase
func GetOrganizationUseCase(db *gorm.DB, config *conf.Configuration) interfaces.IdentityOrganizationUseCase {
	wire.Build(organizationUseCaseSet)
	return nil
}
