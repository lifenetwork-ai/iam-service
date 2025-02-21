package dto

import "time"

// IdentityOrganizationDTO represents an organization.
type IdentityOrganizationDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateIdentityOrganizationPayloadDTO defines the payload for the create organization request
type CreateIdentityOrganizationPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateIdentityOrganizationPayloadDTO defines the payload for the update organization request
type UpdateIdentityOrganizationPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
