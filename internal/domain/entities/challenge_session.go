package domain

type ChallengeSession struct {
	Type  string `json:"type"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	OTP   string `json:"otp"`
}
