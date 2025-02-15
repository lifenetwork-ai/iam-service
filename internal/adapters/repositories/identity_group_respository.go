package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type groupRepository struct {
	db *gorm.DB
}

func NewIdentityGroupRepository(db *gorm.DB) interfaces.IdentityGroupRepository {
	return &groupRepository{db: db}
}

// Get retrieves a list of groups based on the provided filters
func (r *groupRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword string,
) ([]domain.IdentityGroup, error) {
	var entities []domain.IdentityGroup

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve groups: %w", err)
	}

	return entities, nil
}

// GetByID retrieves an group based on the provided ID
func (r *groupRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.IdentityGroup, error) {
	var entity domain.IdentityGroup

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve group: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an group based on the provided code
func (r *groupRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.IdentityGroup, error) {
	var entity domain.IdentityGroup

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve group: %w", err)
	}

	return &entity, nil
}

// Create creates a new group
func (r *groupRepository) Create(
	ctx context.Context,
	entity domain.IdentityGroup,
) (*domain.IdentityGroup, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return &entity, nil
}

// Update updates an existing group
func (r *groupRepository) Update(
	ctx context.Context,
	entity domain.IdentityGroup,
) (*domain.IdentityGroup, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing group
func (r *groupRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.IdentityGroup, error) {
	var entity domain.IdentityGroup

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete group: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing group
func (r *groupRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.IdentityGroup, error) {
	var entity domain.IdentityGroup

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.IdentityGroup{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete group: %w", err)
	}

	return &entity, nil
}
