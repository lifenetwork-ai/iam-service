package access_session

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewAccessSessionRepository(db *gorm.DB) interfaces.AccessSessionRepository {
	return &sessionRepository{db: db}
}

// Get retrieves a list of sessions based on the provided filters
func (r *sessionRepository) Get(
	ctx context.Context,
	limit int,
	offset int,
	keyword *string,
) ([]domain.AccessSession, error) {
	var entities []domain.AccessSession

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != nil {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	return entities, nil
}

// GetByID retrieves an session based on the provided ID
func (r *sessionRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.AccessSession, error) {
	var entity domain.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &entity, nil
}

// GetByCode retrieves an session based on the provided code
func (r *sessionRepository) GetByCode(
	ctx context.Context,
	code string,
) (*domain.AccessSession, error) {
	var entity domain.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &entity, nil
}

// Create creates a new session
func (r *sessionRepository) Create(
	ctx context.Context,
	entity domain.AccessSession,
) (*domain.AccessSession, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &entity, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(
	ctx context.Context,
	entity domain.AccessSession,
) (*domain.AccessSession, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &entity, nil
}

// Delete deletes an existing session
func (r *sessionRepository) Delete(
	ctx context.Context,
	id string,
) (*domain.AccessSession, error) {
	var entity domain.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error; err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &entity, nil
}

// SoftDelete soft-deletes an existing session
func (r *sessionRepository) SoftDelete(
	ctx context.Context,
	id string,
) (*domain.AccessSession, error) {
	var entity domain.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(domain.AccessSession{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete session: %w", err)
	}

	return &entity, nil
}
