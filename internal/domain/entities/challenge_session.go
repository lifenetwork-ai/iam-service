package domain

type ChallengeSession struct {
	GlobalUserID   string `json:"global_user_id"`
	TenantUserID   string `json:"tenant_user_id"`
	IdentifierType string `json:"identifier_type"`
	Identifier     string `json:"identifier"`
	ChallengeType  string `json:"challenge_type"`
	OTP            string `json:"otp"`
	FlowID         string `json:"flow_id"`
}
