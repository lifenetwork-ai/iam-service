package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) interfaces.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateRefreshToken(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *authRepository) FindRefreshToken(hashedToken string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	if err := r.db.Where("hashed_token = ?", hashedToken).First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *authRepository) DeleteRefreshToken(hashedToken string) error {
	return r.db.Where("hashed_token = ?", hashedToken).Delete(&domain.RefreshToken{}).Error
}

func (r *authRepository) FindActiveRefreshToken(accountID string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.Where("account_id = ? AND expires_at > ?", accountID, time.Now()).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}
