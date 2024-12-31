package ucases

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
)

type authUCase struct {
	accountRepository interfaces.AccountRepository
}

func NewAuthUCase(accountRepository interfaces.AccountRepository) interfaces.AuthUCase {
	return &authUCase{accountRepository: accountRepository}
}

// Register handles the creation of a new account and its role-specific details
func (u *authUCase) Register(input *dto.RegisterAccountDTO, role constants.AccountRole) error {
	// Validate input
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" || strings.TrimSpace(string(role)) == "" {
		return errors.New("email, password, and role are required")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Create domain account
	password := string(hashedPassword)
	domainAccount := &domain.Account{
		Email:        input.Email,
		Role:         string(role),
		PasswordHash: &password,
	}

	// Save account
	err = u.accountRepository.CreateAccount(domainAccount)
	if err != nil {
		return err
	}

	// Save role-specific details
	switch role {
	case constants.User:
		if input.FirstName == nil || input.LastName == nil || input.PhoneNumber == nil {
			return errors.New("first_name, last_name, and phone_number are required for USER role")
		}
		domainDetail := &domain.UserDetail{
			AccountID:   domainAccount.ID,
			FirstName:   *input.FirstName,
			LastName:    *input.LastName,
			DateOfBirth: *input.DateOfBirth,
			PhoneNumber: *input.PhoneNumber,
		}
		return u.accountRepository.CreateUserDetail(domainDetail)

	case constants.Partner:
		if input.CompanyName == nil || input.PhoneNumber == nil {
			return errors.New("company_name and phone_number are required for PARTNER role")
		}
		domainDetail := &domain.PartnerDetail{
			AccountID:   domainAccount.ID,
			CompanyName: *input.CompanyName,
			ContactName: *input.ContactName,
			PhoneNumber: *input.PhoneNumber,
		}
		return u.accountRepository.CreatePartnerDetail(domainDetail)

	case constants.Customer:
		if input.OrganizationName == nil || input.PhoneNumber == nil {
			return errors.New("organization_name and phone_number are required for CUSTOMER role")
		}
		domainDetail := &domain.CustomerDetail{
			AccountID:        domainAccount.ID,
			OrganizationName: *input.OrganizationName,
			Industry:         *input.Industry,
			ContactName:      *input.ContactName,
			PhoneNumber:      *input.PhoneNumber,
		}
		return u.accountRepository.CreateCustomerDetail(domainDetail)

	case constants.Validator:
		if input.ValidationOrganization == nil || input.PhoneNumber == nil {
			return errors.New("validation_organization and phone_number are required for VALIDATOR role")
		}
		domainDetail := &domain.ValidatorDetail{
			AccountID:              domainAccount.ID,
			ValidationOrganization: *input.ValidationOrganization,
			ContactPerson:          *input.ContactName,
			PhoneNumber:            *input.PhoneNumber,
		}
		return u.accountRepository.CreateValidatorDetail(domainDetail)

	default:
		return errors.New("invalid role")
	}
}

// Login authenticates an account by email and password
func (u *authUCase) Login(email, password string) (*dto.AccountDTO, error) {
	// Validate input
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email and password are required")
	}

	// Find account by email
	account, err := u.accountRepository.FindAccountByEmail(email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if account.PasswordHash == nil {
		return nil, errors.New("password authentication is not supported for this account")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Convert to DTO and return
	accountDTO := account.ToDTO()
	return accountDTO, nil
}
