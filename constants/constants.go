package constants

import "time"

// Pagination
const (
	DEFAULT_PAGE_TEXT    = "page"
	DEFAULT_SIZE_TEXT    = "size"
	DEFAULT_PAGE         = "1"
	DEFAULT_PAGE_SIZE    = "10"
	DEFAULT_MIN_PAGESIZE = 5
	DEFAULT_MAX_PAGESIZE = 100
)

// Cache config
const (
	DefaultExpiration = 30 * time.Second
	CleanupInterval   = 1 * time.Minute
)

// Webhook constants
const (
	MaxWebhookWorkers = 10
	WebhookTimeout    = 5 * time.Second
)

// Order direction
type OrderDirection string

func (t OrderDirection) String() string {
	return string(t)
}

const (
	Asc  OrderDirection = "ASC"
	Desc OrderDirection = "DESC"
)

// Account role
type AccountRole string

func (t AccountRole) String() string {
	return string(t)
}

const (
	User         AccountRole = "USER"
	DataOwner    AccountRole = "DATA_OWNER"
	DataUtilizer AccountRole = "DATA_UTILIZER"
	Validator    AccountRole = "VALIDATOR"
	Admin        AccountRole = "ADMIN"
)

// IsValid checks if the AccountRole is one of the predefined roles.
func (r AccountRole) IsValid() bool {
	switch r {
	case DataOwner, DataUtilizer, Validator:
		return true
	default:
		return false
	}
}

// Token expiry
const (
	AccessTokenExpiry  = 24 * time.Hour     // 15 minutes
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

// IdentifierType represents the types of identifiers used for login
type IdentifierType string

func (t IdentifierType) String() string {
	return string(t)
}

const (
	IdentifierEmail    IdentifierType = "email"
	IdentifierUsername IdentifierType = "username"
	IdentifierPhone    IdentifierType = "phone"
)

// Refresh token renewal threshold
const RefreshTokenRenewalThreshold = 24 * time.Hour

// DataAccessRequestStatus represents the status of a data access request
type DataAccessRequestStatus string

func (t DataAccessRequestStatus) String() string {
	return string(t)
}

const (
	DataAccessRequestPending  DataAccessRequestStatus = "PENDING"
	DataAccessRequestApproved DataAccessRequestStatus = "APPROVED"
	DataAccessRequestRejected DataAccessRequestStatus = "REJECTED"
)

// DataAccessRequestStatus represents the status of a data access request
type RequesterRequestStatus string

func (t RequesterRequestStatus) String() string {
	return string(t)
}

const (
	RequestValidationPending RequesterRequestStatus = "PENDING"
	RequestValidationValid   RequesterRequestStatus = "VALID"
	RequestValidationInvalid RequesterRequestStatus = "INVALID"
)

// IAMResource represents a type for resource constants
type IAMResource string

func (t IAMResource) String() string {
	return string(t)
}

const (
	ResourceAccounts     IAMResource = "accounts"
	ResourceValidators   IAMResource = "validators"
	ResourceDataRequests IAMResource = "data_requests"
)

// IAMAction represents a type for action constants
type IAMAction string

func (t IAMAction) String() string {
	return string(t)
}

const (
	ActionRead    IAMAction = "read"
	ActionWrite   IAMAction = "write"
	ActionUpdate  IAMAction = "update"
	ActionDelete  IAMAction = "delete"
	ActionApprove IAMAction = "approve"
	ActionReject  IAMAction = "reject"
)

// IAMRolePolicy represents a type for IAM role policies
type IAMRolePolicy string

func (t IAMRolePolicy) String() string {
	return string(t)
}

const (
	AdminPolicy        IAMRolePolicy = "Admin Policy"
	UserPolicy         IAMRolePolicy = "User Policy"
	ValidatorPolicy    IAMRolePolicy = "Validator Policy"
	DataOwnerPolicy    IAMRolePolicy = "Data Owner Policy"
	DataUtilizerPolicy IAMRolePolicy = "Data Utilizer Policy"
)
