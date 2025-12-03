package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type zaloTokenRepository struct {
	db *gorm.DB
}

func NewZaloTokenRepository(db *gorm.DB) domainrepo.ZaloTokenRepository {
	return &zaloTokenRepository{db: db}
}

// Get retrieves the Zalo token for a specific tenant
func (r *zaloTokenRepository) Get(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	var token domain.ZaloToken
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// Save creates or updates the token for a tenant
func (r *zaloTokenRepository) Save(ctx context.Context, token *domain.ZaloToken) error {
	token.UpdatedAt = time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Save(token).Error
}

// GetAll retrieves all tokens (for the refresh worker)
func (r *zaloTokenRepository) GetAll(ctx context.Context) ([]*domain.ZaloToken, error) {
	var tokens []*domain.ZaloToken
	if err := r.db.WithContext(ctx).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// Delete removes a tenant's Zalo token configuration
func (r *zaloTokenRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&domain.ZaloToken{}).Error
}

// GetExpiringSoon retrieves tokens expiring within a duration (for worker)
func (r *zaloTokenRepository) GetExpiringSoon(ctx context.Context, within time.Duration) ([]*domain.ZaloToken, error) {
	var tokens []*domain.ZaloToken
	expiryThreshold := time.Now().Add(within)
	if err := r.db.WithContext(ctx).Where("expires_at <= ?", expiryThreshold).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}
