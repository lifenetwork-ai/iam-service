package dto

// IdentityServiceDTO represents an IdentityService.
type IdentityServiceDTO struct {
	ID string `json:"id"`
}

// CreateIdentityServicePayloadDTO defines the payload for the create group request
type CreateIdentityServicePayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateIdentityServicePayloadDTO defines the payload for the update group request
type UpdateIdentityServicePayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
