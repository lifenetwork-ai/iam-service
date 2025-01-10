package ucases

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/pkg/crypto"
)

type accountUCase struct {
	config            *conf.Configuration
	accountRepository interfaces.AccountRepository
}

func NewAccountUCase(
	config *conf.Configuration,
	accountRepository interfaces.AccountRepository,
) interfaces.AccountUCase {
	return &accountUCase{
		config:            config,
		accountRepository: accountRepository,
	}
}

// FindAccountByEmail retrieves an account by email
func (u *accountUCase) FindAccountByEmail(email string) (*dto.AccountDTO, error) {
	account, err := u.accountRepository.FindAccountByEmail(email)
	return account.ToDTO(), err
}

// FindDetailByAccountID retrieves role-specific details by account ID
func (u *accountUCase) FindDetailByAccountID(account *dto.AccountDTO, role constants.AccountRole) (*dto.AccountDetailDTO, error) {
	// Retrieve secret values
	mnemonic := u.config.Secret.Mnemonic
	passphrase := u.config.Secret.Passphrase
	salt := u.config.Secret.Salt

	switch role {
	case constants.DataOwner:
		detail, err := u.accountRepository.FindDataOwnerByAccountID(account.ID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("data owner not found")
		}

		// Generate public key
		publicKey, privateKey, err := crypto.GenerateAccount(
			mnemonic, passphrase, salt, string(constants.DataOwner), account.ID,
		)
		if err != nil {
			return nil, err
		}

		// Convert public and private keys to hexadecimal strings
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}
		privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
		if err != nil {
			return nil, err
		}

		// Map UserDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:         detail.Account.ID,
				Email:      detail.Account.Email,
				Username:   detail.Account.Username,
				Role:       detail.Account.Role,
				PublicKey:  &publicKeyHex,
				PrivateKey: &privateKeyHex,
			},
			FirstName:   detail.FirstName,
			LastName:    detail.LastName,
			PhoneNumber: detail.PhoneNumber,
		}, nil

	case constants.DataUtilizer:
		detail, err := u.accountRepository.FindDataUtilizerByAccountID(account.ID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("data utilizer not found")
		}

		// Generate public key
		publicKey, privateKey, err := crypto.GenerateAccount(
			mnemonic, passphrase, salt, string(constants.DataUtilizer), account.ID,
		)
		if err != nil {
			return nil, err
		}

		// Convert public and private keys to hexadecimal strings
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}
		privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
		if err != nil {
			return nil, err
		}

		// Map CustomerDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:         detail.Account.ID,
				Email:      detail.Account.Email,
				Username:   detail.Account.Username,
				Role:       detail.Account.Role,
				PublicKey:  &publicKeyHex,
				PrivateKey: &privateKeyHex,
			},
			OrganizationName: detail.OrganizationName,
			Industry:         detail.Industry,
			ContactName:      detail.ContactName,
			PhoneNumber:      detail.PhoneNumber,
		}, nil

	case constants.Validator:
		detail, err := u.accountRepository.FindValidatorByAccountID(account.ID)
		if err != nil {
			return nil, err
		}
		if detail == nil {
			return nil, errors.New("validator detail not found")
		}

		// Generate public key
		publicKey, privateKey, err := crypto.GenerateAccount(
			mnemonic, passphrase, salt, string(constants.Validator), account.ID,
		)
		if err != nil {
			return nil, err
		}

		// Convert public and private keys to hexadecimal strings
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}
		privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
		if err != nil {
			return nil, err
		}

		// Map ValidatorDetail to AccountDetailDTO
		return &dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:         detail.Account.ID,
				Email:      detail.Account.Email,
				Username:   detail.Account.Username,
				Role:       detail.Account.Role,
				PublicKey:  &publicKeyHex,
				PrivateKey: &privateKeyHex,
			},
			ValidationOrganization: detail.ValidationOrganization,
			ContactName:            detail.ContactPerson,
			PhoneNumber:            detail.PhoneNumber,
		}, nil

	case constants.Admin:
		// Generate public key
		publicKey, privateKey, err := crypto.GenerateAccount(
			mnemonic, passphrase, salt, string(constants.Admin), account.ID,
		)
		if err != nil {
			return nil, err
		}

		// Convert public and private keys to hexadecimal strings
		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, err
		}
		privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
		if err != nil {
			return nil, err
		}

		return &dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:         account.ID,
				Email:      account.Email,
				Username:   account.Username,
				Role:       account.Role,
				PublicKey:  &publicKeyHex,
				PrivateKey: &privateKeyHex,
			},
		}, nil

	default:
		return nil, errors.New("invalid role provided")
	}
}

func (u *accountUCase) GetActiveValidators() ([]dto.AccountDetailDTO, error) {
	// Fetch active validators with preloaded account details
	validators, err := u.accountRepository.FindActiveValidators()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active validators: %w", err)
	}

	// Retrieve secret values for public key generation
	mnemonic := u.config.Secret.Mnemonic
	passphrase := u.config.Secret.Passphrase
	salt := u.config.Secret.Salt

	var result []dto.AccountDetailDTO
	for _, v := range validators {
		// Ensure the Account field is not nil
		if v.Account.ID == "" {
			return nil, fmt.Errorf("validator with ID %s has no associated account", *v.ID)
		}

		// Generate public key
		publicKey, _, err := crypto.GenerateAccount(mnemonic, passphrase, salt, v.Account.Role, v.Account.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public key for account %s: %w", v.Account.ID, err)
		}

		publicKeyHex, err := crypto.PublicKeyToHex(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert public key to hex for account %s: %w", v.Account.ID, err)
		}

		// Map ValidatorDetail to AccountDetailDTO
		result = append(result, dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:        v.Account.ID,
				Username:  v.Account.Username,
				Email:     v.Account.Email,
				Role:      v.Account.Role,
				PublicKey: &publicKeyHex,
			},
			ValidationOrganization: v.ValidationOrganization,
			ContactName:            v.ContactPerson,
			PhoneNumber:            v.PhoneNumber,
		})
	}

	return result, nil
}

func (u *accountUCase) FindAccountByID(id string) (*dto.AccountDTO, error) {
	account, err := u.accountRepository.FindAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account by ID: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("account not found")
	}

	return account.ToDTO(), nil
}

// UpdateAccount updates an existing account's details.
func (u *accountUCase) UpdateAccount(accountDTO *dto.AccountDTO) error {
	// Fetch the existing account to ensure it exists
	existingAccount, err := u.accountRepository.FindAccountByID(accountDTO.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch account: %w", err)
	}
	if existingAccount == nil {
		return errors.New("account not found")
	}

	// Map DTO data to the domain model
	existingAccount.Email = accountDTO.Email
	existingAccount.Username = accountDTO.Username
	existingAccount.Role = accountDTO.Role
	existingAccount.APIKey = accountDTO.APIKey
	existingAccount.OAuthProvider = accountDTO.OAuthProvider
	existingAccount.OAuthID = accountDTO.OAuthID

	// Update the account in the repository
	if err := u.accountRepository.UpdateAccount(existingAccount); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}
