package dto

// TokenPairDTO represents a pair of access and refresh tokens
type TokenPairDTO struct {
	AccessToken  string `json:"access_token"`  // The JWT Access Token
	RefreshToken string `json:"refresh_token"` // The plain Refresh Token
}
