package interfaces

import (
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
)

type AccountRepository interface {
	// Accounts
	FindAccountByEmail(email string) (*domain.Account, error)
	FindAccountByID(id string) (*domain.Account, error)
	CreateAccount(account *domain.Account) error

	// UserDetails
	FindUserDetailByAccountID(accountID string) (*domain.UserDetail, error)
	CreateUserDetail(detail *domain.UserDetail) error

	// PartnerDetails
	FindPartnerDetailByAccountID(accountID string) (*domain.PartnerDetail, error)
	CreatePartnerDetail(detail *domain.PartnerDetail) error

	// CustomerDetails
	FindCustomerDetailByAccountID(accountID string) (*domain.CustomerDetail, error)
	CreateCustomerDetail(detail *domain.CustomerDetail) error

	// ValidatorDetails
	FindValidatorDetailByAccountID(accountID string) (*domain.ValidatorDetail, error)
	CreateValidatorDetail(detail *domain.ValidatorDetail) error
}

type AccountUCase interface {
	FindAccountByEmail(email string) (*dto.AccountDTO, error)
	FindDetailByAccountID(accountID string, role constants.AccountRole) (*dto.AccountDetailDTO, error)
}
