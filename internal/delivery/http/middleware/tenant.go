package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	interfaces "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type contextKey string

const (
	TenantIDKey     contextKey = "tenant_id"
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
func GetTenantFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok {
		logger.GetLogger().Errorf("tenant ID not found in context")
		return uuid.Nil, errors.New("tenant ID not found in context")
	}
	return tenantID, nil
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
