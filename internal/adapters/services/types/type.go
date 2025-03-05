package services

type LifeAIProfile struct {
	ID      string `json:"user_profile_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	DoB     string `json:"dob"`
}
