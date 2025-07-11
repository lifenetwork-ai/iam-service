package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
)

// AdminAccount represents an admin account in the system
type AdminAccount struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);not null;unique"`
	Name         string    `gorm:"type:varchar(255);not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:varchar(50);not null;default:'ADMIN'"`
	Status       string    `gorm:"type:varchar(50);not null;default:'active'"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// FromCreateDTO creates a new AdminAccount from a CreateAdminAccountPayloadDTO
func (a *AdminAccount) FromCreateDTO(payload dto.CreateAdminAccountPayloadDTO) error {
	a.ID = uuid.New()
	a.Email = payload.Email
	a.Name = payload.Name
	a.Role = "ADMIN"
	a.Status = "active"
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.PasswordHash = string(hashedPassword)

	return nil
}

// ToDTO converts an AdminAccount to an AdminAccountDTO
func (a *AdminAccount) ToDTO() dto.AdminAccountDTO {
	return dto.AdminAccountDTO{
		ID:        a.ID.String(),
		Email:     a.Email,
		Name:      a.Name,
		Role:      a.Role,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
