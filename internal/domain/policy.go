package domain

import "time"

// Policy represents the structure of a policy in the IAM system.
type Policy struct {
	ID          string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Unique ID for the policy
	Name        string    `json:"name" gorm:"not null;unique"`                               // Unique name of the policy
	Description string    `json:"description"`                                               // Optional description of the policy
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`                          // Timestamp of policy creation
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`                          // Timestamp of last update
}

// TableName overrides the default table name for GORM.
func (Policy) TableName() string {
	return "iam_policies"
}
