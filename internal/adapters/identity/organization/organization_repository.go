package identity_organization

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) interfaces.OrganizationRepository {
	return &organizationRepository{db: db}
}

// Get retrieves a list of organizations based on the provided filters
func (r *organizationRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword *string,
) ([]domain.Organization, error) {
	var organizations []domain.Organization

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != nil {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
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
) (*domain.Organization, error) {
	var organization domain.Organization

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	return &organization, nil
}

// GetByCode retrieves an organization based on the provided code
func (r *organizationRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.Organization, error) {
	var organization domain.Organization

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	return &organization, nil
}

// Create creates a new organization
func (r *organizationRepository) Create(
	ctx context.Context,
	organization domain.Organization,
) (*domain.Organization, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &organization, nil
}

// Update updates an existing organization
func (r *organizationRepository) Update(
	ctx context.Context,
	organization domain.Organization,
) (*domain.Organization, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &organization, nil
}

// Delete deletes an existing organization
func (r *organizationRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.Organization, error) {
	var organization domain.Organization

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to delete organization: %w", err)
	}

	return &organization, nil
}
