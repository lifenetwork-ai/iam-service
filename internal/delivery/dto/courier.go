package dto

type CourierWebhookRequestDTO struct {
	To   string `json:"To" binding:"required"`
	Body string `json:"Body" binding:"required"`
}

type CourierGetAvailableChannelsRequestDTO struct {
	Receiver string `form:"receiver" binding:"required"`
}

type CourierChooseChannelRequestDTO struct {
	Channel  string `json:"channel" binding:"required" description:"The channel to send OTP to the receiver, can be sms, whatsapp or zalo"`
	Receiver string `json:"receiver" binding:"required" description:"The phone number to send OTP to the receiver"`
}
