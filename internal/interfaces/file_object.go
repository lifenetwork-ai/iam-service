package interfaces

import (
	"context"

	"github.com/genefriendway/human-network-auth/internal/domain"
)

type FileObjectRepository interface {
	UploadFile(ctx context.Context) (domain.FileObject, error)
	GetDetail(ctx context.Context, objectID string) (domain.FileObject, error)
}

type FileObjectUCase interface {
	UploadFile(ctx context.Context) (domain.FileObject, error)
	GetDetail(ctx context.Context, objectID string) (domain.FileObject, error)
}
