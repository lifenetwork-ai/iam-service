package ucases

import (
	"fmt"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type fileAccessUCase struct {
	fileAccessRepository interfaces.FileAccessRepository
}

func NewFileAccessUCase(
	fileAccessRepository interfaces.FileAccessRepository,
) interfaces.FileAccessUCase {
	return &fileAccessUCase{
		fileAccessRepository: fileAccessRepository,
	}
}

func (u *fileAccessUCase) CreateFileAccessMapping(payload dto.FileAccessMappingPayloadDTO) error {
	// Validate input
	if payload.FileID == "" || payload.AccountID == "" {
		return fmt.Errorf("file_id and account_id are required")
	}

	// Map the payload DTO to the domain model
	mapping := domain.FileAccessMapping{
		FileID:    payload.FileID,
		AccountID: payload.AccountID,
		Granted:   true, // Default to granted
	}

	// Call the repository to create the mapping
	return u.fileAccessRepository.Create(mapping)
}
