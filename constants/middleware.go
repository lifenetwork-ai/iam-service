package constants

type contextKey string

const (
	SessionTokenKey contextKey = "session_token"
	UserContextKey  contextKey = "user"
)

const (
	TenantKey   contextKey = "tenant"
	TenantIDKey contextKey = "tenant_id"
)

const TenantHeaderKey = "X-Tenant-ID"
