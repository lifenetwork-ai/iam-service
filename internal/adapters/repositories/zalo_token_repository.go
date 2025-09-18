package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type zaloTokenRepository struct {
	db *gorm.DB
}

func NewZaloTokenRepository(db *gorm.DB) domainrepo.ZaloTokenRepository {
	return &zaloTokenRepository{db: db}
}

func (r *zaloTokenRepository) Get(ctx context.Context) (*domain.ZaloToken, error) {
	var token domain.ZaloToken
	if err := r.db.WithContext(ctx).Order(clause.OrderByColumn{
		Column: clause.Column{Name: "updated_at"},
		Desc:   true,
	}).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// Save creates or updates the token
func (r *zaloTokenRepository) Save(ctx context.Context, token *domain.ZaloToken) error {
	token.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(token).Error
}
