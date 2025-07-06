package dto

import "time"

// IdentityUserDTO represents an User.
type IdentityUserDTO struct {
	ID        string `json:"id"`
	Seed      string `json:"seed"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    bool   `json:"status"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Tenant    string `json:"tenant"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// IdentityUserChallengeDTO represents a challenge for identity verification.
type IdentityUserChallengeDTO struct {
	FlowID      string `json:"flow_id" description:"The flow ID of the challenge"`
	Receiver    string `json:"receiver" description:"The receiver of the challenge"`
	ChallengeAt int64  `json:"challenge_at" description:"Time challenge was sent"`
}

// IdentityUserAuthDTO represents the response for a successful authentication with Kratos session
type IdentityUserAuthDTO struct {
	// Core session fields from Kratos
	SessionID       string     `json:"session_id"`
	SessionToken    string     `json:"session_token"` // Token used for authenticating subsequent requests
	Active          bool       `json:"active"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	AuthenticatedAt *time.Time `json:"authenticated_at,omitempty"`

	// User information
	User IdentityUserDTO `json:"user"`

	// Optional session metadata
	AuthenticationMethods []string `json:"authentication_methods,omitempty"`

	// Verification flow (for incomplete registrations)
	VerificationNeeded bool                      `json:"verification_needed,omitempty"`
	VerificationFlow   *IdentityUserChallengeDTO `json:"verification_flow,omitempty"`
}

// IdentityChallengeWithPhoneDTO represents the request for a phone challenge.
type IdentityChallengeWithPhoneDTO struct {
	Phone string `json:"phone"`
}

// IdentityChallengeWithEmailDTO represents the request for a email challenge.
type IdentityChallengeWithEmailDTO struct {
	Email string `json:"email"`
}

type IdentityChallengeVerifyDTO struct {
	FlowID string `json:"flow_id" binding:"required" description:"The flow ID of the challenge"`
	Code   string `json:"code" binding:"required" description:"The code of the challenge"`
	Type   string `json:"type" binding:"required,oneof=registration login" description:"The type of the challenge, can be registration or login"`
}

type IdentityUserRegisterDTO struct {
	Email  string `json:"email" validate:"omitempty,email"`
	Phone  string `json:"phone" validate:"omitempty"`
	Tenant string `json:"tenant" validate:"required"`
}

// IdentityUserLoginDTO represents the request for a user login.
type IdentityUserLoginDTO struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
