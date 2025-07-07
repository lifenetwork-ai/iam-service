package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	types "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{
		db: db,
	}
}

func (r *TenantRepository) Create(tenant *types.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) Update(tenant *types.Tenant) error {
	if tenant.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	return r.db.Save(tenant).Error
}

func (r *TenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&types.Tenant{}, id).Error
}

func (r *TenantRepository) GetByID(id uuid.UUID) (*types.Tenant, error) {
	var tenant types.Tenant
	if err := r.db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) List() ([]*types.Tenant, error) {
	var tenants []*types.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantRepository) GetByName(name string) (*types.Tenant, error) {
	var tenant types.Tenant
	if err := r.db.Where("name = ?", name).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}
