package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type FileInfoRepository interface {
	CreateFileInfo(fileInfo *domain.FileInfo) error
}

type FileInfoUCase interface {
	CreateFileInfo(payload dto.FileInfoPayloadDTO) error
}
