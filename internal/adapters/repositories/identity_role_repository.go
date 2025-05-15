package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

type roleRepository struct {
	db *gorm.DB
}

func NewIdentityRoleRepository(db *gorm.DB) interfaces.IdentityRoleRepository {
	return &roleRepository{db: db}
}

// Get retrieves a list of roles based on the provided filters
func (r *roleRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword string,
) ([]entities.IdentityRole, error) {
	var entities []entities.IdentityRole

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve roles: %w", err)
	}

	return entities, nil
}

// GetByID retrieves an role based on the provided ID
func (r *roleRepository) GetByID(
	ctx context.Context,
	id string,
) (*entities.IdentityRole, error) {
	var entity entities.IdentityRole

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve role: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an role based on the provided code
func (r *roleRepository) GetByCode(
	ctx context.Context,
	code string,
) (*entities.IdentityRole, error) {
	var entity entities.IdentityRole

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve role: %w", err)
	}

	return &entity, nil
}

// Create creates a new role
func (r *roleRepository) Create(
	ctx context.Context,
	entity entities.IdentityRole,
) (*entities.IdentityRole, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &entity, nil
}

// Update updates an existing role
func (r *roleRepository) Update(
	ctx context.Context,
	entity entities.IdentityRole,
) (*entities.IdentityRole, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing role
func (r *roleRepository) Delete(
	ctx context.Context,
	id string,
) (*entities.IdentityRole, error) {
	var entity entities.IdentityRole

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete role: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing role
func (r *roleRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*entities.IdentityRole, error) {
	var entity entities.IdentityRole

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(entities.IdentityRole{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete role: %w", err)
	}

	return &entity, nil
}
