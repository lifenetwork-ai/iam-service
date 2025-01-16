package dto

// RequestValidatorDTO combines validator account info with their validation status
type RequestValidatorDTO struct {
	AccountDTO               // Embed the account information
	ValidationStatus  string `json:"validation_status"`
	ValidationMessage string `json:"validation_message,omitempty"`
}
