package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	interfaces "github.com/genefriendway/human-network-iam/internal/adapters/repositories/types"
	entities "github.com/genefriendway/human-network-iam/internal/domain/entities"
	"github.com/genefriendway/human-network-iam/packages/utils"
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
	keyword string,
) ([]entities.AccessSession, error) {
	var entities []entities.AccessSession

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// If a status filter is provided, apply it to the query
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	return entities, nil
}

// FindByID retrieves an session based on the provided ID
func (r *sessionRepository) FindByID(
	ctx context.Context,
	id string,
) (*entities.AccessSession, error) {
	var entity entities.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &entity, nil
}

// FindByAccessToken retrieves an session based on the provided code
func (r *sessionRepository) FindByAccessToken(
	ctx context.Context,
	accessToken string,
) (*entities.AccessSession, error) {
	var entity entities.AccessSession

	tokenHashed := utils.HashToken(accessToken)
	// Execute query
	if err := r.db.WithContext(ctx).Where("access_token = ?", tokenHashed).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &entity, nil
}

// FindByRefreshToken retrieves an session based on the provided code
func (r *sessionRepository) FindByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*entities.AccessSession, error) {
	var entity entities.AccessSession

	tokenHashed := utils.HashToken(refreshToken)
	// Execute query
	if err := r.db.WithContext(ctx).Where("refresh_token = ?", tokenHashed).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &entity, nil
}

// Create creates a new session
func (r *sessionRepository) Create(
	ctx context.Context,
	entity *entities.AccessSession,
) (*entities.AccessSession, error) {
	if entity.AccessToken != "" {
		entity.AccessToken = utils.HashToken(entity.AccessToken)
	}

	if entity.RefreshToken != "" {
		entity.RefreshToken = utils.HashToken(entity.RefreshToken)
	}

	// Execute query
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return entity, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(
	ctx context.Context,
	entity *entities.AccessSession,
) (*entities.AccessSession, error) {
	// Execute query
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return entity, nil
}

// Delete deletes an existing session
func (r *sessionRepository) Delete(
	ctx context.Context,
	id string,
) (*entities.AccessSession, error) {
	var entity entities.AccessSession

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
) (*entities.AccessSession, error) {
	var entity entities.AccessSession

	// Execute query
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(entities.AccessSession{DeletedAt: gorm.DeletedAt{}}).Error; err != nil {
		return nil, fmt.Errorf("failed to soft-delete session: %w", err)
	}

	return &entity, nil
}
