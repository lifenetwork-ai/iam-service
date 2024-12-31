package dto

type TokenTransferPayloadDTO struct {
	Network     string `json:"network"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TokenAmount string `json:"token_amount"`
	RequestID   string `json:"request_id"`
	Symbol      string `json:"symbol"`
}
