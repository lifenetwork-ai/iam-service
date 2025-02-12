package access_permission

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewAccessPermissionRepository(db *gorm.DB) interfaces.AccessPermissionRepository {
	return &permissionRepository{db: db}
}

// Get retrieves a list of permissions based on the provided filters
func (r *permissionRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword *string,
) ([]domain.AccessPermission, error) {
	var entities []domain.AccessPermission

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != nil {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
	}

	return entities, nil
}

// GetByID retrieves an permission based on the provided ID
func (r *permissionRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.AccessPermission, error) {
	var entity domain.AccessPermission

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve permission: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an permission based on the provided code
func (r *permissionRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.AccessPermission, error) {
	var entity domain.AccessPermission

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve permission: %w", err)
	}

	return &entity, nil
}

// Create creates a new permission
func (r *permissionRepository) Create(
	ctx context.Context,
	entity domain.AccessPermission,
) (*domain.AccessPermission, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return &entity, nil
}

// Update updates an existing permission
func (r *permissionRepository) Update(
	ctx context.Context,
	entity domain.AccessPermission,
) (*domain.AccessPermission, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing permission
func (r *permissionRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.AccessPermission, error) {
	var entity domain.AccessPermission

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete permission: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing permission
func (r *permissionRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.AccessPermission, error) {
	var entity domain.AccessPermission

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.AccessPermission{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete permission: %w", err)
	}

	return &entity, nil
}
