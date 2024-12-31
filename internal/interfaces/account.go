package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type AccountRepository interface {
	// Accounts
	FindAccountByEmail(email string) (*domain.Account, error)
	FindAccountByID(id uint64) (*domain.Account, error)
	CreateAccount(account *domain.Account) error

	// UserDetails
	FindUserDetailByAccountID(accountID uint64) (*domain.UserDetail, error)
	CreateUserDetail(detail *domain.UserDetail) error

	// PartnerDetails
	FindPartnerDetailByAccountID(accountID uint64) (*domain.PartnerDetail, error)
	CreatePartnerDetail(detail *domain.PartnerDetail) error

	// CustomerDetails
	FindCustomerDetailByAccountID(accountID uint64) (*domain.CustomerDetail, error)
	CreateCustomerDetail(detail *domain.CustomerDetail) error

	// ValidatorDetails
	FindValidatorDetailByAccountID(accountID uint64) (*domain.ValidatorDetail, error)
	CreateValidatorDetail(detail *domain.ValidatorDetail) error
}

type AccountUCase interface {
	FindAccountByEmail(email string) (*dto.AccountDTO, error)
	FindDetailByAccountID(accountID uint64, role constants.AccountRole) (*dto.AccountDetailDTO, error)
}
