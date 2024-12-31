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
	Register(input *dto.RegisterAccountDTO, role constants.AccountRole) error
	Login(email, password string) (*dto.TokenPairDTO, error)
	RefreshTokens(refreshToken string) (*dto.TokenPairDTO, error)
	ValidateToken(token string) (*dto.AccountDTO, error)
}
