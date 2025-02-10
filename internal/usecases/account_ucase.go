package usecases

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/constants"
	"github.com/genefriendway/human-network-iam/internal/domain"
	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
	"github.com/genefriendway/human-network-iam/packages/crypto"
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
			mnemonic, passphrase, salt, constants.DataOwner.String(), account.ID,
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
			mnemonic, passphrase, salt, constants.DataUtilizer.String(), account.ID,
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
			mnemonic, passphrase, salt, constants.Validator.String(), account.ID,
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
			mnemonic, passphrase, salt, constants.Admin.String(), account.ID,
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

	case constants.User:
		return &dto.AccountDetailDTO{
			Account: dto.AccountDTO{
				ID:       account.ID,
				Email:    account.Email,
				Username: account.Username,
				Role:     account.Role,
			},
		}, nil

	default:
		return nil, errors.New("invalid role provided")
	}
}

func (u *accountUCase) GetActiveValidators(validatorIDs []string) ([]dto.AccountDetailDTO, error) {
	var validators []domain.Validator
	var err error

	// If validatorIDs is provided, filter by IDs; otherwise, fetch all active validators
	if len(validatorIDs) > 0 {
		validators, err = u.accountRepository.FindActiveValidatorsByAccountIDs(validatorIDs)
	} else {
		validators, err = u.accountRepository.FindActiveValidators()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch validators: %w", err)
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

func (u *accountUCase) FindAccountByAPIKey(apiKey string) (*dto.AccountDTO, error) {
	account, err := u.accountRepository.FindAccountByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account by API key: %w", err)
	}

	if account == nil {
		return nil, nil
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

func (u *accountUCase) GenerateAndAssignAPIKey(accountID string) (string, error) {
	// Fetch the account
	account, err := u.accountRepository.FindAccountByID(accountID)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", domain.ErrDataNotFound
	}

	// Generate a unique API key
	apiKey := uuid.NewString()

	// Update the account with the new API key
	account.APIKey = &apiKey
	if err := u.accountRepository.UpdateAccount(account); err != nil {
		return "", err
	}

	return apiKey, nil
}

func (u *accountUCase) RevokeAPIKey(accountID string) error {
	// Find the account by ID
	account, err := u.accountRepository.FindAccountByID(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrDataNotFound
	}

	// Set the API key to nil
	account.APIKey = nil // Ensure the API key is cleared by setting it to nil
	return u.accountRepository.UpdateAccount(account)
}
