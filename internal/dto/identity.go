package dto

// IdentityChallengeDTO represents a challenge for identity verification.
type IdentityChallengeDTO struct {
	SessionID   string `json:"session_id"`
	ChallengeAt string `json:"challenge_at"`
}
