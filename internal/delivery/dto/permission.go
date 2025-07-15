package dto

import "errors"

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

type CreateRelationTupleRequestDTO struct {
	Namespace  string      `json:"namespace" binding:"required"`
	Relation   string      `json:"relation" binding:"required"`
	Object     string      `json:"object" binding:"required"`
	SubjectID  string      `json:"subject_id,omitempty"`
	SubjectSet *SubjectSet `json:"subject_set,omitempty"`
}

type BatchCheckPermissionRequestDTO struct {
	Tuples []CheckPermissionRequestDTO `json:"tuples" binding:"required"`
}

func (r *CheckPermissionRequestDTO) Validate() error {
	if r.SubjectID == "" && r.SubjectSet == nil {
		return errors.New("either subject_id or subject_set must be provided")
	}
	return nil
}

func (r *CreateRelationTupleRequestDTO) Validate() error {
	if r.SubjectID == "" && r.SubjectSet == nil {
		return errors.New("either subject_id or subject_set must be provided")
	}
	return nil
}

type SubjectSet struct {
	Namespace string `json:"namespace" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
	Object    string `json:"object" binding:"required"`
}
