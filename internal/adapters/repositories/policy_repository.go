package repositories

import (
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type policyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) interfaces.PolicyRepository {
	return &policyRepository{db: db}
}

func (r *policyRepository) CheckPolicyExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.Policy{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *policyRepository) CreatePolicy(policy *domain.Policy) error {
	return r.db.Create(policy).Error
}
