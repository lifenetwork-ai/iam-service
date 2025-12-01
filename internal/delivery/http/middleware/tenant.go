package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
)

// TenantMiddleware handles tenant context in requests
type TenantMiddleware struct {
	tenantRepo domainrepo.TenantRepository
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantRepo domainrepo.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// getTenant extracts tenant from context
func GetTenantFromContext(ctx *gin.Context) (*domain.Tenant, error) {
	tenant, ok := ctx.Get(string(constants.TenantKey))
	if !ok {
		return nil, errors.New("tenant not found in context")
	}
	tenantObj, ok := tenant.(*domain.Tenant)
	if !ok {
		return nil, errors.New("invalid tenant type in context")
	}
	return tenantObj, nil
}

// Middleware handles tenant context in requests
func (m *TenantMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get(constants.TenantHeaderKey)
		if tenantIDStr == "" {
			http.Error(w, "Tenant ID header is required", http.StatusBadRequest)
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
			return
		}

		// Verify tenant exists
		tenant, err := m.tenantRepo.GetByID(tenantID)
		if err != nil {
			http.Error(w, "Error verifying tenant", http.StatusInternalServerError)
			return
		}
		if tenant == nil {
			http.Error(w, "Tenant not found", http.StatusNotFound)
			return
		}

		// Add tenant ID to context
		ctx := context.WithValue(r.Context(), constants.TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
