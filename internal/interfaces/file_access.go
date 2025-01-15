package interfaces

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type FileAccessRepository interface {
	Create(mapping domain.FileAccessMapping) error
}

type FileAccessUCase interface {
	CreateFileAccessMapping(payload dto.FileAccessMappingPayloadDTO) error
}
