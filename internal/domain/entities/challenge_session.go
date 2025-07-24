package domain

type ChallengeSession struct {
	GlobalUserID   string `json:"global_user_id"`
	IdentifierType string `json:"identifier_type"`
	ChallengeType  string `json:"challenge_type"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	OTP            string `json:"otp"`
	FlowID         string `json:"flow_id"`
}
