package clients

type StoreReencryptionKeysRequest struct {
	ReencryptionKeyInfo []ReencryptionKeyInfo `json:"reencryption_key_info"`
}

type ReencryptionKeyInfo struct {
	PubX            string `json:"pub_x"`
	RkKey           string `json:"rk_key"`
	ValidatorID     string `json:"validator_id"`
	ValidatorPubKey string `json:"validator_public_key"`
}

type StoreReencryptionKeysResponse struct {
	Message string `json:"message"`
	OwnerID string `json:"owner_id"`
}
