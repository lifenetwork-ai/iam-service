package dto

import "time"

// IdentityUserDTO represents an User.
type IdentityUserDTO struct {
	ID          string    `json:"id"`
	UserName    string    `json:"user_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Status      bool      `json:"status"`
	Name        string    `json:"name"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	FullName    string    `json:"full_name"`
	LastLoginAt time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IdentityUserChallengeDTO represents a challenge for identity verification.
type IdentityUserChallengeDTO struct {
	SessionID   string `json:"session_id"`
	Receiver    string `json:"receiver"`
	ChallengeAt int64  `json:"challenge_at"`
}

// IdentityUserAuthDTO represents the response for a successful login.
type IdentityUserAuthDTO struct {
	AccessToken      string          `json:"access_token"`
	RefreshToken     string          `json:"refresh_token"`
	AccessExpiresAt  time.Time       `json:"access_expires_at"`
	RefreshExpiresAt time.Time       `json:"refresh_expires_at"`
	LastLoginAt      time.Time       `json:"last_login_at"`
	User             IdentityUserDTO `json:"user"`
}

// IdentityChallengeWithPhoneDTO represents the request for a phone challenge.
type IdentityChallengeWithPhoneDTO struct {
	Phone string `json:"phone"`
}

// IdentityChallengeWithEmailDTO represents the request for a phone challenge.
type IdentityChallengeWithEmailDTO struct {
	Email string `json:"email"`
}

type IdentityUserRegisterDTO struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// IdentityUserLoginDTO represents the request for a user login.
type IdentityUserLoginDTO struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// IdentityRefreshTokenDTO represents the request for a refresh token.
type IdentityRefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
}
