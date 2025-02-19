package dto

// AccessSessionDTO represents an AccessSession.
type AccessSessionDTO struct {
	ID string `json:"id"`
}

// CreateAccessSessionPayloadDTO defines the payload for the create group request
type CreateAccessSessionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateAccessSessionPayloadDTO defines the payload for the update group request
type UpdateAccessSessionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
