package repositories

import (
	"errors"

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

func (r *policyRepository) PolicyExists(policyID string) (bool, error) {
	var policy domain.Policy
	err := r.db.Select("id").Where("id = ?", policyID).First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
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

func (r *policyRepository) GetPolicyByName(name string) (*domain.Policy, error) {
	var policy domain.Policy
	err := r.db.Where("name = ?", name).First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDataNotFound
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *policyRepository) GetAllPolicies() ([]domain.Policy, error) {
	var policies []domain.Policy
	err := r.db.Find(&policies).Error
	return policies, err
}

// GetPolicyByID retrieves a policy by its ID.
func (r *policyRepository) GetPolicyByID(id string) (*domain.Policy, error) {
	var policy domain.Policy
	err := r.db.Where("id = ?", id).First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDataNotFound
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}
