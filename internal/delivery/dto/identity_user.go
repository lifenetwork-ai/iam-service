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
	Type   string `json:"type" binding:"required,oneof=register login verify" description:"The type of the challenge, can be register, login or verify"`
}

type IdentityUserRegisterDTO struct {
	Lang  string `json:"lang" binding:"required,oneof=en vi" description:"The language for the user registration"`
	Email string `json:"email" binding:"omitempty,email"`
	Phone string `json:"phone" binding:"omitempty"`
}

// IdentityUserLoginDTO represents the request for a user login.
type IdentityUserLoginDTO struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// IdentityUserAddIdentifierDTO represents the request for adding a new identifier.
type IdentityUserAddIdentifierDTO struct {
	NewIdentifier string `json:"new_identifier" binding:"required"` // email address or phone number
}

// IdentityUserChangeIdentifierDTO represents the request for changing an identifier.
type IdentityUserChangeIdentifierDTO struct {
	NewIdentifier string `json:"new_identifier" binding:"required"`
}

// IdentityUserDeleteIdentifierDTO represents the request for deleting an identifier.
type IdentityUserDeleteIdentifierDTO struct {
	IdentifierType string `json:"identifier_type" binding:"required,oneof=email phone_number" description:"The type of the identifier, can be email or phone_number"`
}

// CheckIdentifierDTO represents the request for checking if an identifier exists.
type CheckIdentifierDTO struct {
	Identifier string `json:"identifier" binding:"required"`
}

// IdentityVerificationChallengeDTO represents the request for initiating a verification challenge.
type IdentityVerificationChallengeDTO struct {
	Identifier string `json:"identifier" binding:"required" description:"Email or phone number to verify"`
}

// IdentityUserUpdateLangDTO represents the request for updating user's language preference.
type IdentityUserUpdateLangDTO struct {
	Lang string `json:"lang" binding:"required"`
}
