package repositories

import (
	"errors"
	"strings"

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

func (r *policyRepository) AssignPolicyToAccount(accountID, policyID string) error {
	accountPolicy := domain.AccountPolicy{
		AccountID: accountID,
		PolicyID:  policyID,
	}

	err := r.db.Create(&accountPolicy).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return domain.ErrAlreadyExists
		}
		return err
	}

	return nil
}
