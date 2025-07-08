package interfaces

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type GlobalUserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.GlobalUser, error)
	Create(tx *gorm.DB, user *domain.GlobalUser) error
}
