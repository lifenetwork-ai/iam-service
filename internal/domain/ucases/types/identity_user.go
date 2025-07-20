package types

import "time"

// IdentityUserDTO represents an User.
type IdentityUserResponse struct {
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
	User IdentityUserResponse `json:"user,omitempty"`

	// Optional session metadata
	AuthenticationMethods []string `json:"authentication_methods,omitempty"`

	// Verification flow (for incomplete registrations)
	VerificationNeeded bool                           `json:"verification_needed,omitempty"`
	VerificationFlow   *IdentityUserChallengeResponse `json:"verification_flow,omitempty"`
}
