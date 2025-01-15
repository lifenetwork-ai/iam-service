package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type FileInfo struct {
	ID         string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique identifier for the file
	Name       string    `json:"name" gorm:"type:varchar(255);not null"`                    // File name
	ShareCount int       `json:"share_count" gorm:"not null;check:share_count >= 0"`        // Number of shares, must be >= 0
	OwnerID    string    `json:"owner_id" gorm:"type:uuid;not null"`                        // Owner ID, references accounts table
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`                          // Automatically set creation timestamp
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`                          // Automatically update timestamp on changes
}

// TableName overrides the default table name for GORM
func (m *FileInfo) TableName() string {
	return "file_infos"
}

// ToDTO converts a FileInfo domain model to a FileInfoDTO.
func (m *FileInfo) ToDTO() *dto.FileInfoDTO {
	return &dto.FileInfoDTO{
		ID:         m.ID,
		Name:       m.Name,
		ShareCount: m.ShareCount,
		OwnerID:    m.OwnerID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
