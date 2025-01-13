package repositories

import (
	"strings"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type iamRepository struct {
	db *gorm.DB
}

func NewIAMRepository(db *gorm.DB) interfaces.IAMRepository {
	return &iamRepository{db: db}
}

func (r *iamRepository) AssignPolicyToAccount(accountID, policyID string) error {
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

func (r *iamRepository) GetAccountPermissions(accountID string) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Table("iam_permissions").
		Select("iam_permissions.*").
		Joins("INNER JOIN account_policies ON account_policies.policy_id = iam_permissions.policy_id").
		Where("account_policies.account_id = ?", accountID).
		Scan(&permissions).Error
	return permissions, err
}

func (r *iamRepository) GetPermissionsByPolicyID(policyID string) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Where("policy_id = ?", policyID).Find(&permissions).Error
	return permissions, err
}
