package ucases

import (
	"errors"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/pkg/crypto"
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
func (u *accountUCase) FindDetailByAccountID(accountID string, role constants.AccountRole) (*dto.AccountDetailDTO, error) {
	switch role {
	case constants.User:
		detail, err := u.accountRepository.FindUserDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("user detail not found")
		}

		// Generate public key
		publicKey, _, err := crypto.GenerateAccount("mnemonic", "passphrase", "salt", string(constants.User), accountID)
		if err != nil {
			return nil, err
		}
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}

		// Map UserDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:        detail.Account.ID,
				Email:     detail.Account.Email,
				Role:      detail.Account.Role,
				PublicKey: &publicKeyHex,
			},
			FirstName:   detail.FirstName,
			LastName:    detail.LastName,
			PhoneNumber: detail.PhoneNumber,
		}, nil

	case constants.Partner:
		detail, err := u.accountRepository.FindPartnerDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("partner detail not found")
		}

		// Generate public key
		publicKey, _, err := crypto.GenerateAccount("mnemonic", "passphrase", "salt", string(constants.User), accountID)
		if err != nil {
			return nil, err
		}
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}

		// Map PartnerDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:        detail.Account.ID,
				Email:     detail.Account.Email,
				Role:      detail.Account.Role,
				PublicKey: &publicKeyHex,
			},
			CompanyName: detail.CompanyName,
			ContactName: detail.ContactName,
			PhoneNumber: detail.PhoneNumber,
		}, nil

	case constants.Customer:
		detail, err := u.accountRepository.FindCustomerDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("customer detail not found")
		}

		// Generate public key
		publicKey, _, err := crypto.GenerateAccount("mnemonic", "passphrase", "salt", string(constants.User), accountID)
		if err != nil {
			return nil, err
		}
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}

		// Map CustomerDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:        detail.Account.ID,
				Email:     detail.Account.Email,
				Role:      detail.Account.Role,
				PublicKey: &publicKeyHex,
			},
			OrganizationName: detail.OrganizationName,
			Industry:         detail.Industry,
			ContactName:      detail.ContactName,
			PhoneNumber:      detail.PhoneNumber,
		}, nil

	case constants.Validator:
		detail, err := u.accountRepository.FindValidatorDetailByAccountID(accountID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("validator detail not found")
		}

		// Generate public key
		publicKey, _, err := crypto.GenerateAccount("mnemonic", "passphrase", "salt", string(constants.User), accountID)
		if err != nil {
			return nil, err
		}
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}

		// Map ValidatorDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			ID: detail.ID,
			Account: dto.AccountDTO{
				ID:        detail.Account.ID,
				Email:     detail.Account.Email,
				Role:      detail.Account.Role,
				PublicKey: &publicKeyHex,
			},
			ValidationOrganization: detail.ValidationOrganization,
			ContactName:            detail.ContactPerson,
			PhoneNumber:            detail.PhoneNumber,
		}, nil

	default:
		return nil, errors.New("invalid role provided")
	}
}
