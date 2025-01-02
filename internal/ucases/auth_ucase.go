package ucases

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

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
func (u *authUCase) Register(input *dto.RegisterPayloadDTO, role constants.AccountRole) error {
	// Validate input
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" ||
		strings.TrimSpace(input.Username) == "" || strings.TrimSpace(string(role)) == "" {
		return errors.New("email, username, password, and role are required")
	}

	// Check if username already exists
	existingUsername, err := u.accountRepository.FindAccountByUsername(input.Username)
	if err != nil {
		return errors.New("failed to check if username exists")
	}
	if existingUsername != nil {
		return errors.New("username already taken")
	}

	// Check if account already exists by email
	existingAccount, err := u.accountRepository.FindAccountByEmail(input.Email)
	if err != nil {
		return errors.New("failed to check if email exists")
	}
	if existingAccount != nil {
		return domain.ErrAccountAlreadyExists
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
		Username:     input.Username,
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
		domainDetail := &domain.UserDetail{
			AccountID:   domainAccount.ID,
			FirstName:   &input.FirstName,
			LastName:    &input.LastName,
			PhoneNumber: &input.PhoneNumber,
		}
		return u.accountRepository.CreateUserDetail(domainDetail)

	case constants.Partner:
		domainDetail := &domain.PartnerDetail{
			AccountID:   domainAccount.ID,
			CompanyName: &input.CompanyName,
			ContactName: &input.ContactName,
			PhoneNumber: &input.PhoneNumber,
		}
		return u.accountRepository.CreatePartnerDetail(domainDetail)

	case constants.Customer:
		domainDetail := &domain.CustomerDetail{
			AccountID:        domainAccount.ID,
			OrganizationName: &input.OrganizationName,
			Industry:         &input.Industry,
			ContactName:      &input.ContactName,
			PhoneNumber:      &input.PhoneNumber,
		}
		return u.accountRepository.CreateCustomerDetail(domainDetail)

	case constants.Validator:
		domainDetail := &domain.ValidatorDetail{
			AccountID:              domainAccount.ID,
			ValidationOrganization: &input.ValidationOrganization,
			ContactPerson:          &input.ContactName,
			PhoneNumber:            &input.PhoneNumber,
		}
		return u.accountRepository.CreateValidatorDetail(domainDetail)

	default:
		return errors.New("invalid role")
	}
}

// Login authenticates the user and returns a token pair (Access + Refresh)
func (u *authUCase) Login(identifier, password string, identifierType constants.IdentifierType) (*dto.TokenPairDTO, error) {
	// Validate input
	if strings.TrimSpace(identifier) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("identifier and password are required")
	}

	// Find the account based on the identifier type
	var account *domain.Account
	var err error

	lookupMethods := map[constants.IdentifierType]func(string) (*domain.Account, error){
		constants.IdentifierEmail:    u.accountRepository.FindAccountByEmail,
		constants.IdentifierUsername: u.accountRepository.FindAccountByUsername,
		constants.IdentifierPhone:    u.accountRepository.FindAccountByPhoneNumber,
	}

	lookupMethod, exists := lookupMethods[identifierType]
	if !exists {
		return nil, fmt.Errorf("unsupported identifier type: %s", identifierType)
	}

	account, err = lookupMethod(identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account: %w", err)
	}
	if account == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Compare password
	if account.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*account.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check if the user already has an active session
	activeToken, err := u.authRepository.FindActiveRefreshToken(account.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}

	// If an active refresh token exists
	if activeToken != nil {
		if activeToken.ExpiresAt.After(time.Now()) {
			// Token is still valid, return an error indicating the user is already logged in
			return nil, errors.New("user already logged in with an active session")
		} else {
			// Token is expired, delete it
			if delErr := u.authRepository.DeleteRefreshToken(activeToken.HashedToken); delErr != nil {
				return nil, fmt.Errorf("failed to delete expired refresh token: %w", delErr)
			}
		}
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

// Logout invalidates the provided refresh token.
func (u *authUCase) Logout(refreshToken string) error {
	// Hash the incoming token
	hashedToken := utils.HashToken(refreshToken)

	// Validate refresh token existence
	storedToken, err := u.authRepository.FindRefreshToken(hashedToken)
	if err != nil {
		return domain.ErrInvalidToken
	}
	if storedToken == nil {
		return domain.ErrInvalidToken
	}

	// Delete the refresh token
	err = u.authRepository.DeleteRefreshToken(hashedToken)
	if err != nil {
		return errors.New("failed to delete refresh token")
	}

	return nil
}

// RefreshTokens generates a new token pair using the provided refresh token
func (u *authUCase) RefreshTokens(refreshToken string) (*dto.TokenPairDTO, error) {
	// Hash the incoming token
	hashedToken := utils.HashToken(refreshToken)

	// Validate refresh token
	storedToken, err := u.authRepository.FindRefreshToken(hashedToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if storedToken == nil || storedToken.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrExpiredToken
	}

	// Fetch the account associated with the token
	account, err := u.accountRepository.FindAccountByID(storedToken.AccountID)
	if err != nil || account == nil {
		return nil, domain.ErrInvalidToken
	}

	// Generate a new Access Token
	accessToken, err := utils.GenerateToken(account.ID, account.Email, account.Role)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate a new Refresh Token
	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	newHashedToken := utils.HashToken(newRefreshToken)

	// Replace the old refresh token in the repository
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

	// Return the new token pair
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
