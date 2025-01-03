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
	FindActiveRefreshToken(accountID string) (*domain.RefreshToken, error)
}

type AuthUCase interface {
	Register(input *dto.RegisterPayloadDTO) error
	Login(identifier, password string, identifierType constants.IdentifierType) (*dto.TokenPairDTO, error)
	Logout(refreshToken string) error
	RefreshTokens(refreshToken string) (*dto.TokenPairDTO, error)
	ValidateToken(token string) (*dto.AccountDTO, error)
}
