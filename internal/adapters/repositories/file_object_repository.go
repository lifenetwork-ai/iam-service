package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type fileObjectRepository struct {
	db *gorm.DB
}

// NewFileObjectRepository creates a new fileObjectRepository
func NewFileObjectRepository(db *gorm.DB) interfaces.FileObjectRepository {
	return &fileObjectRepository{
		db: db,
	}
}

func (r *fileObjectRepository) UploadFile(ctx context.Context) (domain.FileObject, error) {
	return domain.FileObject{}, errors.New("not implemented")
}

func (r *fileObjectRepository) GetDetail(ctx context.Context, objectID string) (domain.FileObject, error) {
	return domain.FileObject{}, errors.New("not implemented")
}
