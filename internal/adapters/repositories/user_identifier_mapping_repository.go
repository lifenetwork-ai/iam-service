package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

type userIdentifierMappingRepository struct {
	db *gorm.DB
}

func NewUserIdentifierMappingRepository(db *gorm.DB) domainrepo.UserIdentifierMappingRepository {
	return &userIdentifierMappingRepository{db: db}
}

func (r *userIdentifierMappingRepository) GetByGlobalUserID(
	ctx context.Context,
	globalUserID string,
) (*domain.UserIdentifierMapping, error) {
	var mapping domain.UserIdentifierMapping
	if err := r.db.WithContext(ctx).
		Where("global_user_id = ?", globalUserID).
		First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *userIdentifierMappingRepository) Create(ctx context.Context, tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(mapping).Error
}

// Upsert creates or updates the entire mapping row by global_user_id.
func (r *userIdentifierMappingRepository) Upsert(
	ctx context.Context,
	tx *gorm.DB,
	mapping *domain.UserIdentifierMapping,
) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "global_user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"lang": mapping.Lang,
			}),
		}).
		Create(mapping).Error
}

func (r *userIdentifierMappingRepository) Update(tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&domain.UserIdentifierMapping{}).Where("id = ?", mapping.ID).Updates(mapping).Error
}
