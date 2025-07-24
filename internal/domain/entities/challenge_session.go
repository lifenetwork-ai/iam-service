package domain

type ChallengeSession struct {
	GlobalUserID string `json:"global_user_id"`
	Type         string `json:"type"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	OTP          string `json:"otp"`
	Flow         string `json:"flow"`
}
