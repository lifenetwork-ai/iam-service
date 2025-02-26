package services

type LifeAIProfile struct {
	ID      string `json:"user_profile_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Avatar  string `json:"avatar"`
	DoB     string `json:"dob"`
}
