package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type AuthUCase interface {
	Register(account *dto.RegisterAccountDTO, roleSpecificDetails constants.AccountRole) error
	Login(email, password string) (*dto.AccountDTO, error)
}
