package domain

import (
	"time"

	"github.com/genefriendway/human-network-auth/internal/dto"
)

type FileObject struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ObjectID    string    `json:"object_id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	MineType    string    `json:"mine_type"`
	DownloadUrl string    `json:"download_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (m *FileObject) TableName() string {
	return "file_objects"
}

func (m *FileObject) ToDto() dto.FileObjectDTO {
	return dto.FileObjectDTO{
		ObjectID:    m.ObjectID,
		Name:        m.Name,
		Size:        m.Size,
		MineType:    m.MineType,
		DownloadUrl: m.DownloadUrl,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}
