package dto

// AccessPermissionDTO represents an AccessPermission.
type AccessPermissionDTO struct {
	ID          string    `json:"id"`
}

// CreateAccessPermissionPayloadDTO defines the payload for the create group request
type CreateAccessPermissionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateAccessPermissionPayloadDTO defines the payload for the update group request
type UpdateAccessPermissionPayloadDTO struct {
	Name        string `json:"name" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}
