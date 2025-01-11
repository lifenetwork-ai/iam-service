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

var iamUCaseSet = wire.NewSet(
	repositories.NewPolicyRepository,
	repositories.NewAccountRepository,
	ucases.NewIAMUCase,
)

// Init ucase
func GetAuthUCase(db *gorm.DB, config *conf.Configuration) interfaces.AuthUCase {
	wire.Build(authUCaseSet)
	return nil
}

func GetAccountUCase(db *gorm.DB, config *conf.Configuration) interfaces.AccountUCase {
	wire.Build(accountUCaseSet)
	return nil
}

func GetDataAccessUCase(db *gorm.DB, config *conf.Configuration) interfaces.DataAccessUCase {
	wire.Build(dataAccessUCaseSet)
	return nil
}

func GetIAMUCase(db *gorm.DB) interfaces.IAMUCase {
	wire.Build(iamUCaseSet)
	return nil
}
