//go:build wireinject

package wire

import (
	"github.com/genefriendway/human-network-auth/internal/adapters/repositories"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/internal/ucases"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// UCase set
var authUCaseSet = wire.NewSet(
	repositories.NewAccountRepository,
	repositories.NewAuthRepository,
	ucases.NewAuthUCase,
)

// Init ucase
func InitializeAuthUCase(db *gorm.DB) interfaces.AuthUCase {
	wire.Build(authUCaseSet)
	return nil
}
