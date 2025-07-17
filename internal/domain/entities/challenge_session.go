package domain

type ChallengeSession struct {
	ChallengeType string `json:"challenge_type"`
	Type          string `json:"type"`
	Email         string `json:"email,omitempty"`
	Phone         string `json:"phone,omitempty"`
	OTP           string `json:"otp"`
	Flow          string `json:"flow"`
	KratosSession string `json:"kratos_session,omitempty"`
}
