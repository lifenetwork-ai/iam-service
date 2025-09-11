package domain

type ChallengeSession struct {
	GlobalUserID   string `json:"global_user_id"`
	KratosUserID   string `json:"kratos_user_id"`
	IdentifierType string `json:"identifier_type"`
	Identifier     string `json:"identifier"`
	ChallengeType  string `json:"challenge_type"`
	OTP            string `json:"otp"`

	// For update identifier challenge
	IdentityID string `json:"identity_id"`
}
