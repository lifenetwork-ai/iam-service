package dto

// IdentityGroupDTO represents an Group.
type IdentityGroupDTO struct {
	ID          string    `json:"id"`
}

// CreateIdentityGroupPayloadDTO defines the payload for the create group request
type CreateIdentityGroupPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateIdentityGroupPayloadDTO defines the payload for the update group request
type UpdateIdentityGroupPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
