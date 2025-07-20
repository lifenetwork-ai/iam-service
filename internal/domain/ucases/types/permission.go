package types

import (
	"errors"

	keto "github.com/ory/keto-client-go"
)

// CheckPermissionRequest represents a request to check permission
type CheckPermissionRequest struct {
	Namespace      string         // Area of the application (e.g., "document")
	Relation       string         // Relation between the subject and the object (e.g., "read", "write", "delete")
	Object         string         // What they want to do it to (e.g., "document:123")
	TenantRelation TenantRelation // Tenant and user relation, used for cross-tenant permission check
}

type TenantRelation struct {
	TenantID string
	UserID   string
}

var TenantMemberRelation = "member"

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
	if r.TenantRelation.TenantID == "" || r.TenantRelation.UserID == "" {
		return errors.New("tenant_relation is required")
	}
	return nil
}

// ToKetoPostCheckPermissionBody converts the CheckPermissionRequest to a keto.PostCheckPermissionBody
// the request should be validated before calling this method
func (r *CheckPermissionRequest) ToKetoPostCheckPermissionBody() keto.PostCheckPermissionBody {
	body := keto.PostCheckPermissionBody{
		Namespace: &r.Namespace,
		Relation:  &r.Relation,
		Object:    &r.Object,
	}

	body.SubjectSet = &keto.SubjectSet{
		Namespace: r.TenantRelation.TenantID,
		Relation:  TenantMemberRelation,
		Object:    r.TenantRelation.UserID,
	}

	return body
}

// =============================================================================
// BatchCheckPermissionRequest represents a request to batch check permission
// =============================================================================

type BatchCheckPermissionRequest struct {
	Tuples []CheckPermissionRequest
}

func (r *BatchCheckPermissionRequest) Validate() error {
	if len(r.Tuples) == 0 {
		return errors.New("tuples is required")
	}
	for _, tuple := range r.Tuples {
		if err := tuple.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// CreateRelationTupleRequest represents a request to create a relation tuple
// =============================================================================

type CreateRelationTupleRequest struct {
	Namespace  string
	Relation   string
	Object     string
	SubjectSet TenantRelation
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
	if r.SubjectSet.TenantID == "" || r.SubjectSet.UserID == "" {
		return errors.New("subject_set is required")
	}
	return nil
}

// ToKetoCreateRelationshipBody converts the CreateRelationTupleRequest to a keto.CreateRelationshipBody
// the request should be validated before calling this method
func (r *CreateRelationTupleRequest) ToKetoCreateRelationshipBody() keto.CreateRelationshipBody {
	body := keto.CreateRelationshipBody{
		Namespace: &r.Namespace,
		Relation:  &r.Relation,
		Object:    &r.Object,
	}

	body.SubjectSet = &keto.SubjectSet{
		Namespace: r.SubjectSet.TenantID,
		Relation:  TenantMemberRelation,
		Object:    r.SubjectSet.UserID,
	}

	return body
}
