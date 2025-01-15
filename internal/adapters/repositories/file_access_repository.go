package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type fileAccessRepository struct {
	db *gorm.DB
}

func NewFileAccessRepository(db *gorm.DB) interfaces.FileAccessRepository {
	return &fileAccessRepository{db: db}
}

func (r *fileAccessRepository) Create(mapping domain.FileAccessMapping) error {
	return r.db.Create(&mapping).Error
}
