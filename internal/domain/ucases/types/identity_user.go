package types

import "time"

// IdentityUserDTO represents an User.
type IdentityUserResponse struct {
	GlobalUserID string `json:"global_user_id,omitempty"`
	ID           string `json:"id"`
	Seed         string `json:"seed,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Status       bool   `json:"status,omitempty"`
	Name         string `json:"name,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	FullName     string `json:"full_name,omitempty"`
	Tenant       string `json:"tenant"`
	Lang         string `json:"lang"`
	CreatedAt    int64  `json:"created_at,omitempty"`
	UpdatedAt    int64  `json:"updated_at,omitempty"`
}

// IdentityUserChallengeDTO represents a challenge for identity verification.
type IdentityUserChallengeResponse struct {
	FlowID      string `json:"flow_id" description:"The flow ID of the challenge"`
	Receiver    string `json:"receiver" description:"The receiver of the challenge"`
	ChallengeAt int64  `json:"challenge_at" description:"Time challenge was sent"`
}

// IdentityUserAuthDTO represents the response for a successful authentication with Kratos session
type IdentityUserAuthResponse struct {
	// Core session fields from Kratos
	SessionID       string     `json:"session_id,omitempty"`
	SessionToken    string     `json:"session_token,omitempty"` // Token used for authenticating subsequent requests
	Active          bool       `json:"active,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	AuthenticatedAt *time.Time `json:"authenticated_at,omitempty"`

	// User information
	User *IdentityUserResponse `json:"user,omitempty"`

	// Optional session metadata
	AuthenticationMethods []string `json:"authentication_methods,omitempty"`

	// Verification flow (for incomplete registrations)
	VerificationNeeded bool                           `json:"verification_needed,omitempty"`
	VerificationFlow   *IdentityUserChallengeResponse `json:"verification_flow,omitempty"`
}

// IdentityVerificationResponse represents the response for identity verification status
type IdentityVerificationResponse struct {
	FlowID         string `json:"flow_id"`
	Identifier     string `json:"identifier"`
	IdentifierType string `json:"identifier_type"`
	Verified       bool   `json:"verified"`
	VerifiedAt     int64  `json:"verified_at,omitempty"`
}
