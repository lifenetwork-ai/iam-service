package dto

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
	Type   string `json:"type" binding:"required,oneof=register login" description:"The type of the challenge, can be register or login"`
}

type IdentityUserRegisterDTO struct {
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone" validate:"omitempty"`
}

// IdentityUserLoginDTO represents the request for a user login.
type IdentityUserLoginDTO struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
