package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationRepository struct {
	db *gorm.DB
}

func NewIdentityOrganizationRepository(db *gorm.DB) interfaces.IdentityOrganizationRepository {
	return &organizationRepository{db: db}
}

// Get retrieves a list of organizations based on the provided filters
func (r *organizationRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword string,
) ([]domain.IdentityOrganization, error) {
	var organizations []domain.IdentityOrganization

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Execute query
	if err := query.Find(&organizations).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve organizations: %w", err)
	}

	return organizations, nil
}

// GetByID retrieves an organization based on the provided ID
func (r *organizationRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.IdentityOrganization, error) {
	var organization domain.IdentityOrganization

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	return &organization, nil
}

// GetByCode retrieves an organization based on the provided code
func (r *organizationRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.IdentityOrganization, error) {
	var organization domain.IdentityOrganization

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	return &organization, nil
}

// Create creates a new organization
func (r *organizationRepository) Create(
	ctx context.Context,
	entity domain.IdentityOrganization,
) (*domain.IdentityOrganization, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &entity, nil
}

// Update updates an existing organization
func (r *organizationRepository) Update(
	ctx context.Context,
	entity domain.IdentityOrganization,
) (*domain.IdentityOrganization, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing organization
func (r *organizationRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.IdentityOrganization, error) {
	var organization domain.IdentityOrganization

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to delete organization: %w", err)
	}

	return &organization, nil
}

// SoftDelete soft-deletes an existing organization
func (r *organizationRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.IdentityOrganization, error) {
	var organization domain.IdentityOrganization

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.IdentityOrganization{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete organization: %w", err)
	}

	return &organization, nil
}
