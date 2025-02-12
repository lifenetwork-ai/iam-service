package dto

// AccessPolicyDTO represents an AccessPolicy.
type AccessPolicyDTO struct {
	ID          string    `json:"id"`
}

// CreateAccessPolicyPayloadDTO defines the payload for the create group request
type CreateAccessPolicyPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateAccessPolicyPayloadDTO defines the payload for the update group request
type UpdateAccessPolicyPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
