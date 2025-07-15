package repositories

import (
	"errors"

	"gorm.io/gorm"

	"github.com/google/uuid"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

var _ repotypes.TenantRepository = &TenantRepository{}

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{
		db: db,
	}
}

func (r *TenantRepository) Create(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) Update(tenant *entities.Tenant) error {
	if tenant.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	return r.db.Save(tenant).Error
}

func (r *TenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.Tenant{}, id).Error
}

func (r *TenantRepository) GetByID(id uuid.UUID) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := r.db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) List() ([]*entities.Tenant, error) {
	var tenants []*entities.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantRepository) GetByName(name string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := r.db.Where("name = ?", name).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}
