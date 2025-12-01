package types

type ChooseChannelResponse struct {
	Channel   string `json:"channel"`
	ExpiresAt int64  `json:"expires_at"`
}
