package ucases

import (
	"context"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type fileObjectUCase struct {
	fileObjectRepo interfaces.FileObjectRepository
}

func NewFileObjectUCase(fileObjectRepo interfaces.FileObjectRepository) interfaces.FileObjectUCase {
	return &fileObjectUCase{
		fileObjectRepo: fileObjectRepo,
	}
}

func (u *fileObjectUCase) UploadFile(ctx context.Context) (domain.FileObject, error) {
	return u.fileObjectRepo.UploadFile(ctx)
}

func (u *fileObjectUCase) GetDetail(ctx context.Context, objectID string) (domain.FileObject, error) {
	return u.fileObjectRepo.GetDetail(ctx, objectID)
}
