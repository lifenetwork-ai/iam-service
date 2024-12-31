package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type AuthRepository interface {
	CreateRefreshToken(token *domain.RefreshToken) error
	FindRefreshToken(hashedToken string) (*domain.RefreshToken, error)
	DeleteRefreshToken(hashedToken string) error
}

type AuthUCase interface {
	Register(account *dto.RegisterAccountDTO, roleSpecificDetails constants.AccountRole) error
	Login(email, password string) (*dto.AccountDTO, error)
}
