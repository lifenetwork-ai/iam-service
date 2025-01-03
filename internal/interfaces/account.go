package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type AccountRepository interface {
	// Accounts
	FindAccountByUsername(username string) (*domain.Account, error)
	FindAccountByEmail(email string) (*domain.Account, error)
	FindAccountByPhoneNumber(phone string) (*domain.Account, error)
	FindAccountByID(id string) (*domain.Account, error)
	CreateAccount(account *domain.Account) error
	UpdateAccount(account *domain.Account) error

	// UserDetails
	FindUserDetailByAccountID(accountID string) (*domain.UserDetail, error)
	CreateOrUpdateUserDetail(detail *domain.UserDetail) error

	// PartnerDetails
	FindPartnerDetailByAccountID(accountID string) (*domain.PartnerDetail, error)
	CreateOrUpdatePartnerDetail(detail *domain.PartnerDetail) error

	// CustomerDetails
	FindCustomerDetailByAccountID(accountID string) (*domain.CustomerDetail, error)
	CreateOrUpdateCustomerDetail(detail *domain.CustomerDetail) error

	// ValidatorDetails
	FindValidatorDetailByAccountID(accountID string) (*domain.ValidatorDetail, error)
	CreateOrUpdateValidatorDetail(detail *domain.ValidatorDetail) error

	FindActiveValidators() ([]domain.ValidatorDetail, error)
}

type AccountUCase interface {
	FindAccountByEmail(email string) (*dto.AccountDTO, error)
	FindAccountByID(id string) (*dto.AccountDTO, error)
	FindDetailByAccountID(accountID string, role constants.AccountRole) (*dto.AccountDetailDTO, error)
	GetActiveValidators() ([]dto.AccountDetailDTO, error)
	UpdateAccount(accountDTO *dto.AccountDTO) error
}
