package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type fileInfoRepository struct {
	db *gorm.DB
}

// NewFileInfoRepository creates a new instance of FileInfoRepository
func NewFileInfoRepository(db *gorm.DB) interfaces.FileInfoRepository {
	return &fileInfoRepository{db: db}
}

// CreateFileInfo inserts a new FileInfo record into the database
func (r *fileInfoRepository) CreateFileInfo(fileInfo *domain.FileInfo) error {
	if err := r.db.Create(fileInfo).Error; err != nil {
		return err
	}
	return nil
}
