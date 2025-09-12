package types

type AdminAddIdentifierDTO struct {
	ExistingIdentifier string `json:"existing_identifier" binding:"required"`
	NewIdentifier      string `json:"new_identifier" binding:"required"`
}

type AdminAddIdentifierResponse struct {
	GlobalUserID string `json:"global_user_id"`
	KratosUserID string `json:"kratos_user_id"`
	Identifier   string `json:"new_identifier"`
	Lang         string `json:"lang"`
	UpdatedAt    int64  `json:"updated_at"`
}
