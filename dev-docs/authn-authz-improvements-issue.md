# Issue: Insufficient and Inflexible Authentication/Authorization for Permission Handler

## Issue Summary

The permission handler endpoints (`/api/v1/permissions/check` and `/api/v1/permissions/relation-tuples`) currently lack sufficient and flexible authentication and authorization mechanisms. The current implementation has several security gaps and limitations that need to be addressed.

## Current State Analysis

### Current Authentication Flow

Based on the code analysis in `internal/adapters/handlers/permission_handler.go`:

1. **Create Relation Tuple Endpoint** (`POST /api/v1/permissions/relation-tuples`):
   - Requires `X-Tenant-Id` header (validated by `XHeaderValidationMiddleware`)
   - Requires `Authorization: Bearer <token>` header
   - Extracts user from context via `middleware.GetUserFromContext(c)`
   - Creates relation tuple for the authenticated user in the specified tenant

2. **Check Permission Endpoint** (`POST /api/v1/permissions/check`):
   - Requires `X-Tenant-Id` header (validated by `XHeaderValidationMiddleware`)
   - **CRITICAL GAP**: No authentication middleware applied
   - Attempts to extract user from context but fails gracefully
   - Checks permission for the user in the specified tenant

### Current Middleware Stack

From `internal/delivery/http/route/v1.go`:
```go
permissionRouter := v1.Group("permissions")
permissionRouter.Use(middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware())
{
    permissionRouter.POST("/check", permissionHandler.CheckPermission)
    permissionRouter.POST("/relation-tuples", permissionHandler.CreateRelationTuple)
}
```

**Missing**: Authentication middleware (`authMiddleware.RequireAuth()`) is not applied to permission routes.

### Design Question: Who Should Call Check Permission?

The current implementation assumes the check permission endpoint is for **user self-checks** (users checking their own permissions). However, there are two distinct use cases that should be considered:

#### **Use Case 1: User Self-Check** (Current Design)
- **Purpose**: User wants to know if they can perform an action before attempting it
- **Example**: Frontend checking if user can edit a document before showing the edit button
- **Authentication**: Required (user checking their own permissions)
- **Authorization**: User can only check their own permissions

#### **Use Case 2: System Permission Check** (Missing)
- **Purpose**: A service needs to check if a specific user has permission
- **Example**: Document service checking if user:456 can read doc:123 before serving content
- **Authentication**: Required (service authentication)
- **Authorization**: Service needs permission to check other users' permissions

## Security Issues Identified

### 1. **Missing Authentication on Check Permission Endpoint**
- **Severity**: HIGH
- **Impact**: Anyone with a valid tenant ID can check permissions without authentication
- **Current Behavior**: The endpoint attempts to get user from context but continues execution even if no user is found
- **Risk**: Potential information disclosure and permission enumeration attacks

### 2. **Inconsistent Authentication Requirements**
- **Severity**: MEDIUM
- **Impact**: Different endpoints have different security levels
- **Issue**: Create relation tuple requires authentication, but check permission does not

### 3. **No Authorization Checks**
- **Severity**: HIGH
- **Impact**: No validation of whether the authenticated user has permission to perform the requested operations
- **Missing**:
  - Check if user can create relation tuples for the specified namespace/object
  - Check if user can query permissions for the specified namespace/object
  - Role-based access control for permission management operations

### 4. **No Rate Limiting**
- **Severity**: MEDIUM
- **Impact**: Potential for abuse and DoS attacks
- **Missing**: Rate limiting on permission check endpoints

### 5. **No Audit Logging**
- **Severity**: MEDIUM
- **Impact**: No visibility into permission-related operations
- **Missing**: Logging of permission checks and relation tuple creations

### 6. **Limited Use Case Support**
- **Severity**: MEDIUM
- **Impact**: Only supports user self-checks, not system-level permission checks
- **Missing**: Support for services to check permissions for any user

## Required Improvements

### 1. **Apply Consistent Authentication**

**Priority**: HIGH
**Effort**: LOW

Apply authentication middleware to all permission endpoints:

```go
permissionRouter := v1.Group("permissions")
permissionRouter.Use(
    middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
    authMiddleware.RequireAuth(), // Add this line
)
{
    permissionRouter.POST("/check", permissionHandler.CheckPermission)
    permissionRouter.POST("/relation-tuples", permissionHandler.CreateRelationTuple)
}
```

### 2. **Implement Authorization Checks**

**Priority**: HIGH
**Effort**: MEDIUM

Add authorization middleware or checks to validate:

- **For Create Relation Tuple**:
  - User has permission to create relation tuples in the specified namespace
  - User has permission to assign the specified relation to the object
  - User can only create relation tuples for themselves or users they have permission to manage

- **For Check Permission**:
  - User has permission to query permissions in the specified namespace
  - User can only check permissions for objects they have access to
  - Implement proper scope restrictions

### 3. **Add Role-Based Access Control (RBAC)**

**Priority**: MEDIUM
**Effort**: HIGH

Implement role-based permissions for permission management:

- **Permission Admin Role**: Can create/delete relation tuples for any user in their tenant
- **Permission Viewer Role**: Can only check permissions, cannot modify
- **Self-Service Role**: Can only manage their own permissions within allowed scopes

### 4. **Implement Rate Limiting**

**Priority**: MEDIUM
**Effort**: LOW

Add rate limiting middleware specifically for permission endpoints:

```go
permissionRouter.Use(
    middleware.NewXHeaderValidationMiddleware(repos.TenantRepo).Middleware(),
    authMiddleware.RequireAuth(),
    middleware.RateLimitMiddleware("permissions", 100, time.Minute), // 100 requests per minute
)
```

### 5. **Add Comprehensive Audit Logging**

**Priority**: MEDIUM
**Effort**: MEDIUM

Log all permission-related operations with structured data:

```go
// Example audit log entry
{
    "timestamp": "2024-01-15T10:30:00Z",
    "operation": "permission_check",
    "user_id": "user123",
    "tenant_id": "tenant456",
    "namespace": "documents",
    "relation": "read",
    "object": "doc:789",
    "result": "allowed",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0..."
}
```

### 6. **Add Input Validation and Sanitization**

**Priority**: MEDIUM
**Effort**: LOW

Enhance input validation for permission endpoints:

- Validate namespace format and allowed values
- Validate relation names against allowed relations
- Sanitize object identifiers
- Add maximum length limits for all inputs

### 7. **Implement Permission Caching**

**Priority**: LOW
**Effort**: MEDIUM

Add caching for frequently checked permissions to improve performance:

- Cache permission check results for a short duration (e.g., 30 seconds)
- Implement cache invalidation when relation tuples are created/modified
- Use tenant-scoped cache keys to prevent cross-tenant data leakage

### 8. **Support Multiple Use Cases** (NEW)

**Priority**: MEDIUM
**Effort**: MEDIUM

Consider supporting both user self-checks and system-level permission checks:

#### **Option A: Keep current + Add new endpoint**
```go
// Current: User self-check (requires user auth)
POST /api/v1/permissions/check
{
  "namespace": "documents", 
  "relation": "read",
  "object": "doc:123"
}
// Checks permissions for authenticated user

// New: System check (requires service auth)
POST /api/v1/permissions/check/user/{user_id}
{
  "namespace": "documents",
  "relation": "read", 
  "object": "doc:123"
}
// Checks permissions for specified user
```

#### **Option B: Modify current endpoint**
```go
POST /api/v1/permissions/check
{
  "namespace": "documents",
  "relation": "read",
  "object": "doc:123",
  "subject_user_id": "user:456"  // Optional - if not provided, use authenticated user
}
```

## Implementation Plan

### Phase 1: Critical Security Fixes (Week 1)
1. Apply authentication middleware to all permission endpoints
2. Add basic authorization checks
3. Implement rate limiting
4. Add audit logging

### Phase 2: Enhanced Authorization (Week 2-3)
1. Implement role-based access control
2. Add comprehensive permission validation
3. Enhance input validation
4. Add permission caching

### Phase 3: Use Case Expansion (Week 3-4)
1. Decide on approach for multiple use cases (Option A vs B)
2. Implement system-level permission checks
3. Add service authentication for permission checks
4. Update documentation

### Phase 4: Monitoring and Optimization (Week 4)
1. Add metrics and monitoring
2. Performance optimization
3. Security testing and validation
4. Documentation updates

## Testing Requirements

### Security Testing
- [ ] Test unauthenticated access attempts
- [ ] Test unauthorized access attempts
- [ ] Test permission enumeration attacks
- [ ] Test rate limiting effectiveness
- [ ] Test input validation and sanitization

### Functional Testing
- [ ] Test authentication flow for all endpoints
- [ ] Test authorization checks for different user roles
- [ ] Test audit logging functionality
- [ ] Test rate limiting behavior
- [ ] Test permission caching
- [ ] Test user self-check vs system check scenarios

### Performance Testing
- [ ] Test endpoint performance with authentication/authorization
- [ ] Test rate limiting impact on legitimate users
- [ ] Test cache effectiveness

## Success Criteria

1. **Security**: All permission endpoints require proper authentication and authorization
2. **Consistency**: All permission endpoints have consistent security requirements
3. **Auditability**: All permission operations are logged with sufficient detail
4. **Performance**: Endpoints maintain acceptable performance with security measures
5. **Usability**: Legitimate users can perform their required operations without issues
6. **Flexibility**: Support both user self-checks and system-level permission checks

## Related Files

- `internal/adapters/handlers/permission_handler.go` - Main handler implementation
- `internal/delivery/http/route/v1.go` - Route configuration
- `internal/delivery/http/middleware/authentication.go` - Authentication middleware
- `internal/delivery/http/middleware/auth.go` - Basic auth middleware
- `dev-docs/Authorization.md` - Current authorization documentation

## Notes

- This issue should be prioritized as HIGH due to the security implications
- Consider implementing these changes incrementally to minimize disruption
- Ensure backward compatibility during the transition period
- Update API documentation to reflect new authentication requirements
- **Key Decision Needed**: Whether to support system-level permission checks and how to implement them 