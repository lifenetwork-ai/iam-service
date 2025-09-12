package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

// SQLite-backed lightweight repositories intended for tests

// SQLiteTenantRepository implements TenantRepository against a provided *gorm.DB (sqlite)
type SQLiteTenantRepository struct{ db *gorm.DB }

func NewSQLiteTenantRepository(db *gorm.DB) domainrepo.TenantRepository {
	return &SQLiteTenantRepository{db: db}
}

func (r *SQLiteTenantRepository) Create(t *domain.Tenant) error { return r.db.Create(t).Error }
func (r *SQLiteTenantRepository) Update(t *domain.Tenant) error { return r.db.Save(t).Error }
func (r *SQLiteTenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Tenant{}, "id = ?", id).Error
}
func (r *SQLiteTenantRepository) GetByID(id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
func (r *SQLiteTenantRepository) List() ([]*domain.Tenant, error) {
	var ts []*domain.Tenant
	return ts, r.db.Find(&ts).Error
}
func (r *SQLiteTenantRepository) GetByName(name string) (*domain.Tenant, error) {
	var t domain.Tenant
	if err := r.db.First(&t, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// SQLiteGlobalUserRepository implements GlobalUserRepository for sqlite
type SQLiteGlobalUserRepository struct{ db *gorm.DB }

func NewSQLiteGlobalUserRepository(db *gorm.DB) domainrepo.GlobalUserRepository {
	return &SQLiteGlobalUserRepository{db: db}
}

func (r *SQLiteGlobalUserRepository) GetByID(ctx context.Context, id string) (*domain.GlobalUser, error) {
	var g domain.GlobalUser
	if err := r.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}
func (r *SQLiteGlobalUserRepository) Create(tx *gorm.DB, g *domain.GlobalUser) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(g).Error
}

// SQLiteUserIdentifierMappingRepository implements UserIdentifierMappingRepository for sqlite
type SQLiteUserIdentifierMappingRepository struct{ db *gorm.DB }

func NewSQLiteUserIdentifierMappingRepository(db *gorm.DB) domainrepo.UserIdentifierMappingRepository {
	return &SQLiteUserIdentifierMappingRepository{db: db}
}

func (r *SQLiteUserIdentifierMappingRepository) GetByGlobalUserID(ctx context.Context, globalUserID string) (*domain.UserIdentifierMapping, error) {
	var m domain.UserIdentifierMapping
	if err := r.db.First(&m, "global_user_id = ?", globalUserID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SQLiteUserIdentifierMappingRepository) ExistsByTenantAndKratosUserID(ctx context.Context, tx *gorm.DB, tenantID, kratosUserID string) (bool, error) {
	if tx == nil {
		tx = r.db
	}
	var c int64
	if err := tx.Model(&domain.UserIdentifierMapping{}).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND ui.kratos_user_id = ?", tenantID, kratosUserID).
		Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (r *SQLiteUserIdentifierMappingRepository) Create(ctx context.Context, tx *gorm.DB, m *domain.UserIdentifierMapping) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(m).Error
}

func (r *SQLiteUserIdentifierMappingRepository) GetByTenantIDAndIdentifier(ctx context.Context, tenantID, identifierType, identifierValue string) (string, error) {
	// get identity first
	var i domain.UserIdentity

	if err := r.db.First(&i, "tenant_id = ? AND type = ? AND value = ?", tenantID, identifierType, identifierValue).Error; err != nil {
		return "", err
	}

	var m domain.UserIdentifierMapping
	// Prefer the latest mapping for this global user within the tenant
	if err := r.db.
		Order("created_at DESC").
		First(&m, "tenant_id = ? AND global_user_id = ?", tenantID, i.GlobalUserID).Error; err != nil {
		return "", fmt.Errorf("identity with id: %s not found: %w", i.ID, err)
	}
	if m.ID == "" {
		return "", fmt.Errorf("mapping with global user id: %s not found", i.GlobalUserID)
	}
	return m.GlobalUserID, nil
}
func (r *SQLiteUserIdentifierMappingRepository) GetByGlobalUserIDAndTenantID(ctx context.Context, globalUserID, tenantID string) ([]*domain.UserIdentifierMapping, error) {
	var m []*domain.UserIdentifierMapping
	if err := r.db.Find(&m, "global_user_id = ? AND tenant_id = ?", globalUserID, tenantID).Error; err != nil {
		return nil, err
	}
	return m, nil
}
func (r *SQLiteUserIdentifierMappingRepository) GetByTenantIDAndTenantUserID(ctx context.Context, tenantID, tenantUserID string) (*domain.UserIdentifierMapping, error) {
	var m domain.UserIdentifierMapping
	if err := r.db.First(&m, "tenant_id = ? AND tenant_user_id = ?", tenantID, tenantUserID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *SQLiteUserIdentifierMappingRepository) Update(tx *gorm.DB, m *domain.UserIdentifierMapping) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(m).Error
}

func (r *SQLiteUserIdentifierMappingRepository) ExistsMapping(ctx context.Context, tenantID, globalUserID string) (bool, error) {
	var c int64
	if err := r.db.Model(&domain.UserIdentifierMapping{}).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND user_identifier_mapping.global_user_id = ?", tenantID, globalUserID).
		Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (r *SQLiteUserIdentifierMappingRepository) Upsert(ctx context.Context, tx *gorm.DB, mapping *domain.UserIdentifierMapping) error {
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

// SQLiteUserIdentityRepository implements UserIdentityRepository for sqlite
type SQLiteUserIdentityRepository struct{ db *gorm.DB }

func NewSQLiteUserIdentityRepository(db *gorm.DB) domainrepo.UserIdentityRepository {
	return &SQLiteUserIdentityRepository{db: db}
}

func (r *SQLiteUserIdentityRepository) GetByID(ctx context.Context, tx *gorm.DB, identityID string) (*domain.UserIdentity, error) {
	if tx == nil {
		tx = r.db
	}
	var u domain.UserIdentity
	if err := tx.First(&u, "id = ?", identityID).Error; err != nil {
		return nil, fmt.Errorf("identity with id: %s not found: %w", identityID, err)
	}
	return &u, nil
}

func (r *SQLiteUserIdentityRepository) GetByTypeAndValue(ctx context.Context, tx *gorm.DB, tenantID, identityType, value string) (*domain.UserIdentity, error) {
	if tx == nil {
		tx = r.db
	}
	var u domain.UserIdentity
	if err := tx.First(&u, "tenant_id = ? AND type = ? AND value = ?", tenantID, identityType, value).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *SQLiteUserIdentityRepository) FindGlobalUserIDByIdentity(ctx context.Context, tenantID, identityType, value string) (string, error) {
	var u domain.UserIdentity
	if err := r.db.First(&u, "tenant_id = ? AND type = ? AND value = ?", tenantID, identityType, value).Error; err != nil {
		return "", err
	}
	return u.GlobalUserID, nil
}
func (r *SQLiteUserIdentityRepository) InsertOnceByKratosUserAndType(ctx context.Context, tx *gorm.DB, tenantID, kratosUserID, globalUserID, idType, value string) (bool, error) {
	if tx == nil {
		tx = r.db
	}
	exists, err := r.ExistsWithinTenant(ctx, tenantID, idType, value)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	return true, tx.WithContext(ctx).Create(&domain.UserIdentity{
		TenantID:     tenantID,
		KratosUserID: kratosUserID,
		GlobalUserID: globalUserID,
		Type:         idType,
		Value:        value,
	}).Error
}
func (r *SQLiteUserIdentityRepository) Update(tx *gorm.DB, identity *domain.UserIdentity) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&domain.UserIdentity{}).Where("id = ?", identity.ID).Updates(map[string]interface{}{"global_user_id": identity.GlobalUserID, "tenant_id": identity.TenantID, "type": identity.Type, "value": identity.Value}).Error
}
func (r *SQLiteUserIdentityRepository) ExistsWithinTenant(ctx context.Context, tenantID, identityType, value string) (bool, error) {
	var c int64
	if err := r.db.Model(&domain.UserIdentity{}).Where("tenant_id = ? AND type = ? AND value = ?", tenantID, identityType, value).Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}
func (r *SQLiteUserIdentityRepository) GetByTenantAndTenantUserID(ctx context.Context, tx *gorm.DB, tenantID, tenantUserID string) (*domain.UserIdentity, error) {
	if tx == nil {
		tx = r.db
	}
	var m domain.UserIdentifierMapping
	if err := tx.First(&m, "tenant_id = ? AND tenant_user_id = ?", tenantID, tenantUserID).Error; err != nil {
		return nil, err
	}
	var u domain.UserIdentity
	if err := tx.First(&u, "tenant_id = ? AND global_user_id = ?", tenantID, m.GlobalUserID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteUserIdentityRepository) ExistsByTenantGlobalUserIDAndType(ctx context.Context, tenantID, globalUserID, identityType string) (bool, error) {
	var c int64
	if err := r.db.Model(&domain.UserIdentity{}).Where("tenant_id = ? AND global_user_id = ? AND type = ?", tenantID, globalUserID, identityType).Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}
func (r *SQLiteUserIdentityRepository) GetByGlobalUserIDAndTenantID(ctx context.Context, tx *gorm.DB, globalUserID, tenantID string) ([]*domain.UserIdentity, error) {
	if tx == nil {
		tx = r.db
	}
	var out []*domain.UserIdentity
	if err := tx.Where("global_user_id = ? AND tenant_id = ?", globalUserID, tenantID).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteUserIdentityRepository) Delete(tx *gorm.DB, identityID string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Delete(&domain.UserIdentity{}, "id = ?", identityID).Error
}

func (r *SQLiteUserIdentityRepository) GetByTenantAndKratosUserID(ctx context.Context, tx *gorm.DB, tenantID, kratosUserID string) (*domain.UserIdentity, error) {
	if tx == nil {
		tx = r.db
	}
	var out *domain.UserIdentity
	if err := tx.WithContext(ctx).
		Where(`
			tenant_id = ? 
			AND global_user_id = (
				SELECT global_user_id 
				FROM user_identities 
				WHERE tenant_id = ? AND kratos_user_id = ? 
				LIMIT 1
			)
		`, tenantID, tenantID, kratosUserID).
		First(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteUserIdentifierMappingRepository) GetByTenantIDAndKratosUserID(ctx context.Context, tenantID, kratosUserID string) (*domain.UserIdentifierMapping, error) {
	var m domain.UserIdentifierMapping
	err := r.db.WithContext(ctx).
		Model(&domain.UserIdentifierMapping{}).
		Joins("JOIN user_identities ui ON ui.global_user_id = user_identifier_mapping.global_user_id").
		Where("ui.tenant_id = ? AND ui.kratos_user_id = ?", tenantID, kratosUserID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
