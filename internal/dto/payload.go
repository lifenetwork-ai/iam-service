package dto

import (
	"github.com/genefriendway/human-network-auth/constants"
)

// LoginPayloadDTO defines the payload for the login request
type LoginPayloadDTO struct {
	Identifier     string                   `json:"identifier" validate:"required"`      // Identifier (email, username, or phone number)
	Password       string                   `json:"password" validate:"required"`        // User password
	IdentifierType constants.IdentifierType `json:"identifier_type" validate:"required"` // Type of identifier: "email", "username", or "phone"
}

// RegisterPayloadtDTO defines the payload for the register request
type RegisterPayloadDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenPayloadDTO defines the payload for refreshing tokens request
type RefreshTokenPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutPayloadDTO defines the payload for the logout request
type LogoutPayloadDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateRolePayloadDTO defines the payload for updating the role of an account
type UpdateRolePayloadDTO struct {
	Role        string                `json:"role" validate:"required"` // USER, PARTNER, CUSTOMER, VALIDATOR
	RoleDetails RoleDetailsPayloadDTO `json:"role_details,omitempty"`   // Role-specific details
}

// RoleDetailsPayloadDTO defines the payload for the role-specific details
type RoleDetailsPayloadDTO struct {
	// Common fields
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`

	// Partner fields
	CompanyName string `json:"company_name,omitempty"`
	ContactName string `json:"contact_name,omitempty"`

	// Customer fields
	OrganizationName string `json:"organization_name,omitempty"`
	Industry         string `json:"industry,omitempty"`

	// Validator fields
	ValidationOrganization string `json:"validation_organization,omitempty"`
}

// DataAccessRequestPayloadDTO defines the payload for creating a data access request
type DataAccessRequestPayloadDTO struct {
	RequestAccountID string `json:"request_account_id" validate:"required,uuid"` // Account whose data is being requested
	ReasonForRequest string `json:"reason_for_request" validate:"required"`      // Reason for the data access request
	FileID           string `json:"file_id" validate:"required,uuid"`            // File ID
}

// RejectRequestPayloadDTO defines the payload for rejecting a data access request
type RejectRequestPayloadDTO struct {
	Reason string `json:"reason" validate:"required"` // Reason for rejecting the data access request
}

// PolicyPayloadDTO defines the payload for creating a policy
type PolicyPayloadDTO struct {
	Name        string `json:"name" validate:"required"` // Name of the policy
	Description string `json:"description,omitempty"`    // Optional description
}

// AssignPolicyPayloadDTO defines the payload for assigning a policy to an account
type AssignPolicyPayloadDTO struct {
	PolicyID string `json:"policy_id" validate:"required,uuid"`
}

// CheckPermissionPayloadDTO defines the payload for checking if an account has permission to perform an action on a resource
type CheckPermissionPayloadDTO struct {
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// PermissionPayloadDTO defines the payload for creating a permission
type PermissionPayloadDTO struct {
	PolicyID    string `json:"policy_id" binding:"required_without=PolicyName"` // Either PolicyID or PolicyName is required
	PolicyName  string `json:"policy_name" binding:"required_without=PolicyID"` // Either PolicyID or PolicyName is required
	Resource    string `json:"resource" binding:"required"`                     // The resource the permission applies to
	Action      string `json:"action" binding:"required"`                       // The action the permission allows
	Description string `json:"description,omitempty"`                           // Optional description of the permission
}

// DataUploadNotificationPayloadDTO defines the payload for data upload notifications
type DataUploadNotificationPayloadDTO struct {
	DataID           string   `json:"data_id" validate:"required"`                      // Unique identifier of the uploaded data
	AccessAccountIDs []string `json:"access_account_ids" validate:"required,dive,uuid"` // List of account IDs with access to the data
}

// FileAccessMappingPayloadDTO defines the payload for creating a file access mapping
type FileAccessMappingPayloadDTO struct {
	FileID    string `json:"file_id" validate:"required,uuid"`    // File ID being accessed
	AccountID string `json:"account_id" validate:"required,uuid"` // Account requesting access
}

// FileInfoPayloadDTO defines the payload for creating a file info
type FileInfoPayloadDTO struct {
	ID         string `json:"id" validate:"required,uuid"`           // File ID
	Name       string `json:"name" validate:"required"`              // File name
	ShareCount int    `json:"share_count" validate:"required,min=0"` // Number of shares
	OwnerID    string `json:"owner_id" validate:"required,uuid"`     // Owner ID
}
