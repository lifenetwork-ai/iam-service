package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) interfaces.OrganizationRepository {
	return &organizationRepository{db: db}
}

