package dto

import (
	"errors"

	keto "github.com/ory/keto-client-go"
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
	Namespace  string      `json:"namespace" binding:"required"`
	Action     string      `json:"action" binding:"required"` // What they want to do (e.g., "read", "write", "delete")
	Object     string      `json:"object" binding:"required"` // What they want to do it to (e.g., "document:123")
	SubjectID  string      `json:"subject_id,omitempty"`      // Who wants to perform the action (user ID, role, etc.)
	SubjectSet *SubjectSet `json:"subject_set,omitempty"`
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
	Namespace  string      `json:"namespace" binding:"required"`
	Relation   string      `json:"relation" binding:"required"`
	Object     string      `json:"object" binding:"required"`
	SubjectID  string      `json:"subject_id,omitempty"`
	SubjectSet *SubjectSet `json:"subject_set,omitempty"`
}

// =============================================================================
// Validation Methods
// =============================================================================

// Validate ensures the CheckPermissionRequestDTO has required fields
func (r *CheckPermissionRequestDTO) Validate() error {
	if r.SubjectID == "" && r.SubjectSet == nil {
		return errors.New("either subject_id or subject_set must be provided")
	}
	return nil
}

// Validate ensures the CreateRelationTupleRequestDTO has required fields
func (r *CreateRelationTupleRequestDTO) Validate() error {
	if r.SubjectID == "" && r.SubjectSet == nil {
		return errors.New("either subject_id or subject_set must be provided")
	}
	return nil
}

// =============================================================================
// Conversion Methods
// =============================================================================

// ToKetoCreateRelationshipBody converts the CreateRelationTupleRequestDTO to a keto.CreateRelationshipBody
// This is used to create a relation tuple in Keto
func (r *CreateRelationTupleRequestDTO) ToKetoCreateRelationshipBody() keto.CreateRelationshipBody {
	body := keto.CreateRelationshipBody{
		Namespace: &r.Namespace,
		Relation:  &r.Relation,
		Object:    &r.Object,
	}

	if r.SubjectID != "" {
		body.SubjectId = &r.SubjectID
	} else if r.SubjectSet != nil {
		body.SubjectSet = &keto.SubjectSet{
			Namespace: r.SubjectSet.Namespace,
			Relation:  r.SubjectSet.Relation,
			Object:    r.SubjectSet.Object,
		}
	}

	return body
}

// ToKetoPostCheckPermissionBody converts the CheckPermissionRequestDTO to a keto.PostCheckPermissionBody
// This is used to check permission in Keto
// Note: The dto should be validated before calling this function
func (r *CheckPermissionRequestDTO) ToKetoPostCheckPermissionBody() keto.PostCheckPermissionBody {
	body := keto.PostCheckPermissionBody{
		Namespace: &r.Namespace,
		Relation:  &r.Action,
		Object:    &r.Object,
	}

	if r.SubjectID != "" {
		body.SubjectId = &r.SubjectID
	} else if r.SubjectSet != nil {
		body.SubjectSet = &keto.SubjectSet{
			Namespace: r.SubjectSet.Namespace,
			Relation:  r.SubjectSet.Relation,
			Object:    r.SubjectSet.Object,
		}
	}

	return body
}
