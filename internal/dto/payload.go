package dto

// LoginPayloadDTO defines the payload for the login request
type LoginPayloadDTO struct {
	Email    string `json:"email" validate:"required,email"` // User email
	Password string `json:"password" validate:"required"`    // User password
}
