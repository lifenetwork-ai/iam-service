package dto

// IdentityUserDTO represents an User.
type IdentityUserDTO struct {
	ID        string `json:"id"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    bool   `json:"status"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
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
	AccessExpiresAt  int64           `json:"access_expires_at"`
	RefreshExpiresAt int64           `json:"refresh_expires_at"`
	LastLoginAt      int64           `json:"last_login_at"`
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

type IdentityChallengeVerifyDTO struct {
	SessionID string `json:"session_id"`
	Code      string `json:"code"`
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
