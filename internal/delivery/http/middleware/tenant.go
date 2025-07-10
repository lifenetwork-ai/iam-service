package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type contextKey string

const (
	TenantIDKey     contextKey = "tenant_id"
	TenantKey       contextKey = "tenant"
	TenantHeaderKey            = "X-Tenant-ID"
)

// TenantMiddleware handles tenant context in requests
type TenantMiddleware struct {
	tenantRepo interfaces.TenantRepository
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantRepo interfaces.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// GetTenantFromContext retrieves tenant ID from context
func GetTenantFromContext(ctx context.Context) (*domain.Tenant, error) {
	tenant, ok := ctx.Value(TenantKey).(*domain.Tenant)
	if !ok {
		logger.GetLogger().Errorf("tenant not found in context")
		return nil, errors.New("tenant not found in context")
	}
	return tenant, nil
}

// Middleware handles tenant context in requests
func (m *TenantMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get(TenantHeaderKey)
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
		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
