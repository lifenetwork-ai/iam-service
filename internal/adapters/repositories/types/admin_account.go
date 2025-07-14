package interfaces

import (
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

// AdminAccountRepository defines the interface for admin account repository operations
type AdminAccountRepository interface {
	Create(account *domain.AdminAccount) error
	GetByUsername(username string) (*domain.AdminAccount, error)
	GetByID(id string) (*domain.AdminAccount, error)
	Update(account *domain.AdminAccount) error
	Delete(id string) error
}
