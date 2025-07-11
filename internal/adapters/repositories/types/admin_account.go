package interfaces

import (
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"gorm.io/gorm"
)

// AdminAccountRepository defines the interface for admin account repository operations
type AdminAccountRepository interface {
	Create(db *gorm.DB, account *domain.AdminAccount) error
	GetByEmail(email string) (*domain.AdminAccount, error)
	GetByID(id string) (*domain.AdminAccount, error)
	Update(db *gorm.DB, account *domain.AdminAccount) error
	Delete(db *gorm.DB, id string) error
}
