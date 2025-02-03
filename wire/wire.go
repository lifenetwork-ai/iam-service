//go:build wireinject

package wire

import (
	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/internal/adapters/repositories"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/internal/usecases"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var organizationUseCaseSet = wire.NewSet(
	repositories.NewOrganizationRepository,
	usecases.NewOrganizationUseCase,
)

// Init ucase
func GetOrganizationUseCase(db *gorm.DB, config *conf.Configuration) interfaces.OrganizationUseCase {
	wire.Build(organizationUseCaseSet)
	return nil
}
