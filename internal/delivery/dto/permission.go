package dto

import (
	"errors"
)

// =============================================================================
// Base Types
// =============================================================================

// SubjectSet represents a subject set in Keto's permission model
type SubjectSet struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
	Object    string `json:"object" binding:"required"`
}

// =============================================================================
// Permission Check DTOs
// =============================================================================

// CheckPermissionRequestDTO represents a permission check request
type CheckPermissionRequestDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"` // What they want to do (e.g., "read", "write", "delete")
	Object    string `json:"object" binding:"required"`   // What they want to do it to (e.g., "document:123")
}

// CheckPermissionResponseDTO represents a permission check response
type CheckPermissionResponseDTO struct {
	Allowed bool   `json:"allowed"`          // Whether the action is allowed
	Reason  string `json:"reason,omitempty"` // Optional explanation for why permission was denied
}

// BatchCheckPermissionRequestDTO represents a batch permission check request
type BatchCheckPermissionRequestDTO struct {
	Tuples []CheckPermissionRequestDTO `json:"tuples" binding:"required"`
}

// =============================================================================
// Relation Tuple DTOs
// =============================================================================

// CreateRelationTupleRequestDTO represents a request to create a relation tuple
type CreateRelationTupleRequestDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
	Object    string `json:"object" binding:"required"`
}

// =============================================================================
// Validation Methods
// =============================================================================

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
