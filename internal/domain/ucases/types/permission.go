package types

import (
	"errors"
)

type PermissionRequest interface {
	Validate() error
	GetIdentifier() string
}

var (
	_ PermissionRequest = &CheckPermissionRequest{}
	_ PermissionRequest = &CreateRelationTupleRequest{}
	_ PermissionRequest = &DelegateAccessRequest{}
)

// CheckPermissionRequest represents a request to check permission
type CheckPermissionRequest struct {
	Namespace      string         // Area of the application (e.g., "document")
	Relation       string         // Relation between the subject and the object (e.g., "read", "write", "delete")
	Object         string         // What they want to do it to (e.g., "document:123")
	TenantRelation TenantRelation // Tenant and identifier relation, used for cross-tenant permission check

	// GlobalUserID is the user id of the user who is checking the permission
	GlobalUserID string
}

func (r *CheckPermissionRequest) GetIdentifier() string {
	return r.TenantRelation.Identifier
}

type TenantRelation struct {
	TenantID   string
	Identifier string // Identifier will be set by the ucase when the request is validated, so it is not required when passed by the handler
}

// Validate validates the CheckPermissionRequest
func (r *CheckPermissionRequest) Validate() error {
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Relation == "" {
		return errors.New("relation is required")
	}
	if r.Object == "" {
		return errors.New("object is required")
	}

	// TenantID is required for cross-tenant permission check
	if r.TenantRelation.TenantID == "" {
		return errors.New("tenant_relation is required")
	}
	if r.TenantRelation.Identifier == "" {
		return errors.New("identifier is required")
	}

	return nil
}

// =============================================================================
// CreateRelationTupleRequest represents a request to create a relation tuple
// =============================================================================

type CreateRelationTupleRequest struct {
	Namespace      string
	Relation       string
	Object         string
	TenantRelation TenantRelation

	// GlobalUserID is the user id of the user who is being added to the relation tuple
	GlobalUserID string
}

func (r *CreateRelationTupleRequest) GetIdentifier() string {
	return r.TenantRelation.Identifier
}

func (r *CreateRelationTupleRequest) Validate() error {
	if r.Namespace == "" {
		return errors.New("namespace is required")
	}
	if r.Relation == "" {
		return errors.New("relation is required")
	}
	if r.Object == "" {
		return errors.New("object is required")
	}
	if r.TenantRelation.TenantID == "" {
		return errors.New("tenant_relation is required")
	}
	if r.TenantRelation.Identifier == "" {
		return errors.New("identifier is required")
	}

	return nil
}

// =============================================================================
// DelegateAccessRequest represents a request to delegate access
// =============================================================================

type DelegateAccessRequest struct {
	ResourceType string
	ResourceID   string
	Permission   string
	TenantID     string
	Identifier   string // Identifier is the identifier of the user who is being added to the relation tuple
}

func (r *DelegateAccessRequest) Validate() error {
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

func (r *DelegateAccessRequest) GetIdentifier() string {
	return r.Identifier
}
