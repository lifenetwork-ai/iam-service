package dto

type CourierWebhookRequestDTO struct {
	To   string `json:"To" binding:"required"`
	Body string `json:"Body" binding:"required"`
}

type CourierGetAvailableChannelsRequestDTO struct {
	Receiver string `json:"receiver" binding:"required"`
}

type CourierChooseChannelRequestDTO struct {
	Channel  string `json:"channel" binding:"required"`
	Receiver string `json:"receiver" binding:"required"`
}
