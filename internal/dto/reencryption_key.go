package dto

// ReencryptionKeyDTO defines the structure for a single re-encryption key entry
type ReencryptionKeyDTO struct {
	PubX               string `json:"pub_x" validate:"required"`
	RkKey              string `json:"rk_key" validate:"required"`
	ValidatorID        string `json:"validator_id" validate:"required,uuid"`
	ValidatorPublicKey string `json:"validator_public_key" validate:"required"`
}
