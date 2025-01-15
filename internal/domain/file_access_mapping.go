package domain

import "time"

type FileAccessMapping struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FileID    string    `gorm:"type:uuid;not null"`
	AccountID string    `gorm:"type:uuid;not null"`
	Granted   bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName overrides the default table name
func (m *FileAccessMapping) TableName() string {
	return "file_access_mappings"
}
