package ucases

import (
	"errors"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type accountUCase struct {
	accountRepository interfaces.AccountRepository
}

func NewAccountUCase(accountRepository interfaces.AccountRepository) interfaces.AccountUCase {
	return &accountUCase{accountRepository: accountRepository}
}

// FindAccountByEmail retrieves an account by email
func (u *accountUCase) FindAccountByEmail(email string) (*dto.AccountDTO, error) {
	account, err := u.accountRepository.FindAccountByEmail(email)
	return account.ToDTO(), err
}

// FindDetailByAccountID retrieves role-specific details by account ID
func (u *accountUCase) FindDetailByAccountID(accountID uint64, role constants.AccountRole) (*dto.AccountDetailDTO, error) {
	switch role {
	case constants.User:
		detail, err := u.accountRepository.FindUserDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("user detail not found")
		}
		// Map UserDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:    detail.Account.ID,
				Email: detail.Account.Email,
				Role:  detail.Account.Role,
			},
			FirstName:   &detail.FirstName,
			LastName:    &detail.LastName,
			DateOfBirth: &detail.DateOfBirth,
			PhoneNumber: &detail.PhoneNumber,
		}, nil

	case constants.Partner:
		detail, err := u.accountRepository.FindPartnerDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("partner detail not found")
		}
		// Map PartnerDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:    detail.Account.ID,
				Email: detail.Account.Email,
				Role:  detail.Account.Role,
			},
			CompanyName: &detail.CompanyName,
			ContactName: &detail.ContactName,
			PhoneNumber: &detail.PhoneNumber,
		}, nil

	case constants.Customer:
		detail, err := u.accountRepository.FindCustomerDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("customer detail not found")
		}
		// Map CustomerDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:    detail.Account.ID,
				Email: detail.Account.Email,
				Role:  detail.Account.Role,
			},
			OrganizationName: &detail.OrganizationName,
			Industry:         &detail.Industry,
			ContactName:      &detail.ContactName,
			PhoneNumber:      &detail.PhoneNumber,
		}, nil

	case constants.Validator:
		detail, err := u.accountRepository.FindValidatorDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("validator detail not found")
		}
		// Map ValidatorDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:    detail.Account.ID,
				Email: detail.Account.Email,
				Role:  detail.Account.Role,
			},
			ValidationOrganization: &detail.ValidationOrganization,
			ContactName:            &detail.ContactPerson,
			PhoneNumber:            &detail.PhoneNumber,
		}, nil

	default:
		return nil, errors.New("invalid role provided")
	}
}
