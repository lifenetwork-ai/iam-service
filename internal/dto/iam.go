package dto

// PolicyWithPermissionsDTO represents a policy along with its associated permissions.
type PolicyWithPermissionsDTO struct {
	Policy      PolicyDTO       `json:"policy"`
	Permissions []PermissionDTO `json:"permissions"`
}
