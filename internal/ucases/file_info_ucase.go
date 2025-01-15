package ucases

import (
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type fileInfoUCase struct {
	fileInfoRepo interfaces.FileInfoRepository
}

func NewFileInfoUCase(repo interfaces.FileInfoRepository) interfaces.FileInfoUCase {
	return &fileInfoUCase{fileInfoRepo: repo}
}

// CreateFileInfo creates a new file info record using a DTO
func (u *fileInfoUCase) CreateFileInfo(payload dto.FileInfoPayloadDTO) error {
	// Map the DTO to the domain model
	fileInfo := domain.FileInfo{
		ID:         payload.ID,
		Name:       payload.Name,
		ShareCount: payload.ShareCount,
		OwnerID:    payload.OwnerID,
	}

	// Save the file info using the repository
	return u.fileInfoRepo.CreateFileInfo(&fileInfo)
}
