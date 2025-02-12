package dto

// AccessActionDTO represents an AccessAction.
type AccessActionDTO struct {
	ID          string    `json:"id"`
}

// CreateAccessActionPayloadDTO defines the payload for the create group request
type CreateAccessActionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateAccessActionPayloadDTO defines the payload for the update group request
type UpdateAccessActionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
