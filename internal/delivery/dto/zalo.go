package dto

type CreateZaloTokenRequestDTO struct {
	AppID        string `json:"app_id" binding:"required"`
	SecretKey    string `json:"secret_key" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
	AccessToken  string `json:"access_token"` // Optional
}

type ZaloTokenResponseDTO struct {
	AppID        string `json:"app_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	UpdatedAt    string `json:"updated_at"`
}

type RefreshZaloTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshZaloTokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	UpdatedAt    string `json:"updated_at"`
}
