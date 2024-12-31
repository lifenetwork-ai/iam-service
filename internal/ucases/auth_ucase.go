package ucases

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/pkg/utils"
)

type authUCase struct {
	accountRepository interfaces.AccountRepository
	authRepository    interfaces.AuthRepository
}

func NewAuthUCase(
	accountRepository interfaces.AccountRepository,
	authRepository interfaces.AuthRepository,
) interfaces.AuthUCase {
	return &authUCase{
		accountRepository: accountRepository,
		authRepository:    authRepository,
	}
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

// Login authenticates the user and returns a token pair (Access + Refresh)
func (u *authUCase) Login(email, password string) (*dto.TokenPairDTO, error) {
	// Validate input
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email and password are required")
	}

	// Find account by email
	account, err := u.accountRepository.FindAccountByEmail(email)
	if err != nil {
		return nil, errors.New("failed to fetch account")
	}
	if account == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Compare password
	if account.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate Access Token
	accessToken, err := utils.GenerateToken(account.ID, account.Email, account.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate Refresh Token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Save hashed refresh token
	hashedToken := utils.HashToken(refreshToken)
	if err := u.authRepository.CreateRefreshToken(&domain.RefreshToken{
		AccountID:   account.ID,
		HashedToken: hashedToken,
		ExpiresAt:   time.Now().Add(constants.RefreshTokenExpiry),
	}); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	// Return Token Pair
	return &dto.TokenPairDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokens generates a new token pair using the provided refresh token
func (u *authUCase) RefreshTokens(refreshToken string) (*dto.TokenPairDTO, error) {
	// Hash incoming token
	hashedToken := utils.HashToken(refreshToken)

	// Validate refresh token
	storedToken, err := u.authRepository.FindRefreshToken(hashedToken)
	if err != nil || storedToken == nil || storedToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Generate new Access Token
	account, err := u.accountRepository.FindAccountByID(storedToken.AccountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}
	accessToken, err := utils.GenerateToken(account.ID, account.Email, account.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate new Refresh Token
	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	newHashedToken := utils.HashToken(newRefreshToken)

	// Replace the old refresh token
	err = u.authRepository.DeleteRefreshToken(hashedToken)
	if err != nil {
		return nil, errors.New("failed to delete old refresh token")
	}
	err = u.authRepository.CreateRefreshToken(&domain.RefreshToken{
		AccountID:   account.ID,
		HashedToken: newHashedToken,
		ExpiresAt:   time.Now().Add(constants.RefreshTokenExpiry),
	})
	if err != nil {
		return nil, errors.New("failed to store new refresh token")
	}

	return &dto.TokenPairDTO{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// ValidateToken validates an access token and retrieves account details
func (u *authUCase) ValidateToken(token string) (*dto.AccountDTO, error) {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil, err
	}

	// Fetch account details
	account, err := u.accountRepository.FindAccountByID(claims.AccountID)
	if err != nil || account == nil {
		return nil, errors.New("account not found")
	}

	return account.ToDTO(), nil
}
