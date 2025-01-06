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
var authUCaseSet = wire.NewSet(
	repositories.NewAccountRepository,
	repositories.NewAuthRepository,
	ucases.NewAuthUCase,
)

var accountUCaseSet = wire.NewSet(
	repositories.NewAccountRepository,
	ucases.NewAccountUCase,
)

var dataAccessUCaseSet = wire.NewSet(
	repositories.NewDataAccessRepository,
	repositories.NewAccountRepository,
	ucases.NewDataAccessUCase,
)

// Init ucase
func InitializeAuthUCase(db *gorm.DB) interfaces.AuthUCase {
	wire.Build(authUCaseSet)
	return nil
}

func InitializeAccountUCase(db *gorm.DB, config *conf.Configuration) interfaces.AccountUCase {
	wire.Build(accountUCaseSet)
	return nil
}

func InitializeDataAccessUCase(db *gorm.DB, config *conf.Configuration) interfaces.DataAccessUCase {
	wire.Build(dataAccessUCaseSet)
	return nil
}
