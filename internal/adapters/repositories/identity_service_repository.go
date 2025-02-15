package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type serviceRepository struct {
	db *gorm.DB
}

func NewIdentityServiceRepository(db *gorm.DB) interfaces.IdentityServiceRepository {
	return &serviceRepository{db: db}
}

// Get retrieves a list of roles based on the provided filters
func (r *serviceRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword string,
) ([]domain.IdentityService, error) {
	var entities []domain.IdentityService

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
func (r *serviceRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.IdentityService, error) {
	var entity domain.IdentityService

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve role: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an role based on the provided code
func (r *serviceRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.IdentityService, error) {
	var entity domain.IdentityService

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve role: %w", err)
	}

	return &entity, nil
}

// Create creates a new role
func (r *serviceRepository) Create(
	ctx context.Context,
	entity domain.IdentityService,
) (*domain.IdentityService, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &entity, nil
}

// Update updates an existing role
func (r *serviceRepository) Update(
	ctx context.Context,
	entity domain.IdentityService,
) (*domain.IdentityService, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing role
func (r *serviceRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.IdentityService, error) {
	var entity domain.IdentityService

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete role: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing role
func (r *serviceRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.IdentityService, error) {
	var entity domain.IdentityService

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.IdentityService{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete role: %w", err)
	}

	return &entity, nil
}
