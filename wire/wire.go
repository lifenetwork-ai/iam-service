//go:build wireinject

package wire

import (
	"github.com/genefriendway/human-network-iam/conf"
	identity_organization "github.com/genefriendway/human-network-iam/internal/adapters/identity/organization"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var organizationUseCaseSet = wire.NewSet(
	identity_organization.NewOrganizationRepository,
	identity_organization.NewOrganizationUseCase,
)

// Init ucase
func GetOrganizationUseCase(db *gorm.DB, config *conf.Configuration) interfaces.OrganizationUseCase {
	wire.Build(organizationUseCaseSet)
	return nil
}
