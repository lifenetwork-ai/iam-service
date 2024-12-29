//go:build wireinject

package wire

import (
	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/internal/adapters/repositories"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/internal/ucases"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var fileObjectUCaseSet = wire.NewSet(
	repositories.NewFileObjectRepository,
	ucases.NewFileObjectUCase,
)

// Init ucase
func InitializeFileObjectUCase(db *gorm.DB, config *conf.Configuration) interfaces.FileObjectUCase {
	wire.Build(fileObjectUCaseSet)
	return nil
}
