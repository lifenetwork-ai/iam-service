package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) interfaces.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) PermissionExists(policyID, resource, action string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Permission{}).
		Where("policy_id = ? AND resource = ? AND action = ?", policyID, resource, action).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *permissionRepository) CreatePermission(permission *domain.Permission) error {
	return r.db.Create(permission).Error
}
