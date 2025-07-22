package dto

import (
	"errors"
)

// =============================================================================
// Base Types
// =============================================================================

// TenantMemberDTO represents a tenant member in the context of a permission check
type TenantMemberDTO struct {
	TenantID string `json:"tenant_id" binding:"required"`
	// Identifier is the identifier of the tenant member (can be email or phone number)
	Identifier string `json:"identifier" binding:"required"`
}

// =============================================================================
// Permission Check DTOs
// =============================================================================

// SelfCheckPermissionRequestDTO represents a permission check request initiated by the user themselves
// This is used to check if the user has permission to perform an action on a resource
// User id will be the subject of the check, and will be based on user's session token
type SelfCheckPermissionRequestDTO struct {
	Namespace string `json:"namespace" binding:"required"` // Name of the resource's group (e.g., "document", "user")
	Relation  string `json:"relation" binding:"required"`  // The relation between the subject and the object (e.g., "read", "write", "delete")
	Object    string `json:"object" binding:"required"`    // The specific resource (e.g., "document:123")
}

// CheckPermissionResponseDTO represents a permission check response
type CheckPermissionResponseDTO struct {
	Allowed bool   `json:"allowed"`          // Whether the action is allowed
	Reason  string `json:"reason,omitempty"` // Optional explanation for why permission was denied
}

// CheckPermissionRequestDTO represents a permission check request initiated by another service
type CheckPermissionRequestDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
	Object    string `json:"object" binding:"required"`

	// SubjectSet defines the subject that is being checked permission for
	TenantMember TenantMemberDTO `json:"tenant_member" binding:"required"`
}

type DelegateAccessRequestDTO struct {
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceID   string `json:"resource_id" binding:"required"`
	Permission   string `json:"permission" binding:"required"`
	TenantID     string `json:"tenant_id" binding:"required"`
	Identifier   string `json:"identifier" binding:"required"`
}

// =============================================================================
// Relation Tuple DTOs
// =============================================================================

// CreateRelationTupleRequestDTO represents a request to create a relation tuple
type CreateRelationTupleRequestDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
	Object    string `json:"object" binding:"required"`

	Identifier string `json:"identifier" binding:"required"`
}

// =============================================================================
// Validation Methods
// =============================================================================

// Validate ensures the SelfCheckPermissionRequestDTO has required fields
func (r *SelfCheckPermissionRequestDTO) Validate() error {
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Relation == "" {
		return errors.New("relation is required")
	}
	if r.Object == "" {
		return errors.New("object is required")
	}
	return nil
}

// Validate ensures the CheckPermissionRequestDTO has required fields
func (r *CheckPermissionRequestDTO) Validate() error {
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Relation == "" {
		return errors.New("relation is required")
	}
	if r.Object == "" {
		return errors.New("object is required")
	}
	if r.TenantMember.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if r.TenantMember.Identifier == "" {
		return errors.New("identifier is required")
	}
	return nil
}

// Validate ensures the CreateRelationTupleRequestDTO has required fields
func (r *CreateRelationTupleRequestDTO) Validate() error {
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Relation == "" {
		return errors.New("relation is required")
	}
	if r.Object == "" {
		return errors.New("object is required")
	}
	return nil
}

func (r *DelegateAccessRequestDTO) Validate() error {
	if r.ResourceType == "" {
		return errors.New("resource_type is required")
	}
	if r.ResourceID == "" {
		return errors.New("resource_id is required")
	}
	if r.Permission == "" {
		return errors.New("permission is required")
	}
	if r.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if r.Identifier == "" {
		return errors.New("identifier is required")
	}
	return nil
}
