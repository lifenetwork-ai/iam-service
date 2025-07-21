package dto

type CourierWebhookRequestDTO struct {
	To   string `json:"To" binding:"required"`
	Body string `json:"Body" binding:"required"`
}
