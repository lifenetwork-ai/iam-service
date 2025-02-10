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

// GetOrganizations retrieves a list of organizations based on the provided filters
func (r *organizationRepository) GetOrganizations(
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
