package dto

// IdentityRoleDTO represents an IdentityRole.
type IdentityRoleDTO struct {
	ID          string    `json:"id"`
}

// CreateIdentityRolePayloadDTO defines the payload for the create group request
type CreateIdentityRolePayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateIdentityRolePayloadDTO defines the payload for the update group request
type UpdateIdentityRolePayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
