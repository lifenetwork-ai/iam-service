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
	Lang  string `json:"lang" binding:"required,oneof=en vi" description:"The language for the user registration"`
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone" validate:"omitempty"`
}

// IdentityUserLoginDTO represents the request for a user login.
type IdentityUserLoginDTO struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// IdentityUserChangeIdentifierDTO represents the request for changing an identifier.
type IdentityUserAddIdentifierDTO struct {
	NewIdentifier string `json:"new_identifier" binding:"required"` // email address or phone number
}

// IdentityUserUpdateIdentifierDTO represents the request for updating an identifier.
type IdentityUserUpdateIdentifierDTO struct {
	NewIdentifier  string `json:"new_identifier" binding:"required"`
	IdentifierType string `json:"identifier_type" binding:"required,oneof=email phone_number" description:"The type of the identifier, can be email or phone_number"`
}
