package access_policy

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type policyRepository struct {
	db *gorm.DB
}

func NewAccessPolicyRepository(db *gorm.DB) interfaces.AccessPolicyRepository {
	return &policyRepository{db: db}
}

// Get retrieves a list of policies based on the provided filters
func (r *policyRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword *string,
) ([]domain.AccessPolicy, error) {
	var entities []domain.AccessPolicy

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != nil {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve policys: %w", err)
	}

	return entities, nil
}

// GetByID retrieves an policy based on the provided ID
func (r *policyRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.AccessPolicy, error) {
	var entity domain.AccessPolicy

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve policy: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an policy based on the provided code
func (r *policyRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.AccessPolicy, error) {
	var entity domain.AccessPolicy

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve policy: %w", err)
	}

	return &entity, nil
}

// Create creates a new policy
func (r *policyRepository) Create(
	ctx context.Context,
	entity domain.AccessPolicy,
) (*domain.AccessPolicy, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return &entity, nil
}

// Update updates an existing policy
func (r *policyRepository) Update(
	ctx context.Context,
	entity domain.AccessPolicy,
) (*domain.AccessPolicy, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing policy
func (r *policyRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.AccessPolicy, error) {
	var entity domain.AccessPolicy

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete policy: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing policy
func (r *policyRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.AccessPolicy, error) {
	var entity domain.AccessPolicy

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.AccessPolicy{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete policy: %w", err)
	}

	return &entity, nil
}
