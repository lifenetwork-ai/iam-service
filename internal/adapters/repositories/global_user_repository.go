package repositories

import (
	"context"

	"gorm.io/gorm"

	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type globalUserRepository struct {
	db *gorm.DB
}

func NewGlobalUserRepository(db *gorm.DB) interfaces.GlobalUserRepository {
	return &globalUserRepository{db: db}
}

func (r *globalUserRepository) GetByID(ctx context.Context, id string) (*domain.GlobalUser, error) {
	var user domain.GlobalUser
	if err := r.db.WithContext(ctx).Preload("Identities").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *globalUserRepository) Create(tx *gorm.DB, user *domain.GlobalUser) error {
	return tx.Create(user).Error
}
