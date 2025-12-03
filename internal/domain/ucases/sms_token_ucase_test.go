package ucases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

// roundTripFunc allows stubbing http.DefaultTransport
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// mockZaloRepository is a minimal in-memory repository for testing
type mockZaloRepository struct {
	tokens map[uuid.UUID]*domain.ZaloToken
	err    error
}

func newMockZaloRepository() *mockZaloRepository {
	return &mockZaloRepository{
		tokens: make(map[uuid.UUID]*domain.ZaloToken),
	}
}

func (r *mockZaloRepository) Get(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	if r.err != nil {
		return nil, r.err
	}
	token, exists := r.tokens[tenantID]
	if !exists {
		return nil, fmt.Errorf("no token for tenant")
	}
	t := *token
	return &t, nil
}

func (r *mockZaloRepository) Save(ctx context.Context, token *domain.ZaloToken) error {
	t := *token
	r.tokens[token.TenantID] = &t
	return nil
}

func (r *mockZaloRepository) GetAll(ctx context.Context) ([]*domain.ZaloToken, error) {
	var tokens []*domain.ZaloToken
	for _, token := range r.tokens {
		t := *token
		tokens = append(tokens, &t)
	}
	return tokens, nil
}

func (r *mockZaloRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	delete(r.tokens, tenantID)
	return nil
}

func TestZaloHealthCheck_Success(t *testing.T) {
	// Note: Cannot use t.Parallel() because we modify http.DefaultTransport

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Prepare encrypted token in repository
	crypto := common.NewZaloTokenCrypto(dbEncryptionKey)
	plainToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	encryptedToken, err := crypto.Encrypt(context.Background(), plainToken)
	require.NoError(t, err)

	repo := newMockZaloRepository()
	repo.tokens[tenantID] = encryptedToken

	// Mock successful GetAllTemplates response
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			// Verify that the access token is being sent
			require.Equal(t, "test-access-token", req.Header.Get("access_token"))

			body := map[string]any{
				"error":   0,
				"message": "",
				"data": []any{
					map[string]any{
						"templateId":      123,
						"templateName":    "Test Template",
						"createdTime":     1234567890,
						"status":          "ENABLE",
						"templateQuality": "HIGH",
					},
				},
				"metadata": map[string]any{"total": 1},
			}
			b, _ := json.Marshal(body)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(b)),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert success
	require.Nil(t, derr, "expected health check to succeed")
}

func TestZaloHealthCheck_NoTokenInRepository(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Empty repository
	repo := newMockZaloRepository()

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure - should fail to get token
	require.NotNil(t, derr, "expected health check to fail when no token in repository")
	require.Equal(t, "MSG_GET_TOKEN_FAILED", derr.Code)
}

func TestZaloHealthCheck_RepositoryError(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Repository that returns error
	repo := newMockZaloRepository()
	repo.err = fmt.Errorf("database connection failed")

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure
	require.NotNil(t, derr, "expected health check to fail on repository error")
	require.Equal(t, "MSG_GET_TOKEN_FAILED", derr.Code)
}

func TestZaloHealthCheck_InvalidEncryptedToken(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Repository with invalid encrypted token
	invalidToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "invalid-base64!!!",
		RefreshToken: "invalid-base64!!!",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	repo := newMockZaloRepository()
	repo.tokens[tenantID] = invalidToken

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure - should fail to decrypt token
	require.NotNil(t, derr, "expected health check to fail when token decryption fails")
	require.Equal(t, "MSG_DECRYPT_TOKEN_FAILED", derr.Code)
}

func TestZaloHealthCheck_GetAllTemplatesHTTPError(t *testing.T) {
	// Note: Cannot use t.Parallel() because we modify http.DefaultTransport

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Prepare encrypted token
	crypto := common.NewZaloTokenCrypto(dbEncryptionKey)
	plainToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	encryptedToken, err := crypto.Encrypt(context.Background(), plainToken)
	require.NoError(t, err)

	repo := newMockZaloRepository()
	repo.tokens[tenantID] = encryptedToken

	// Mock HTTP error (network failure)
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" {
			return nil, fmt.Errorf("connection timeout")
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure
	require.NotNil(t, derr, "expected health check to fail due to HTTP error")
	require.Equal(t, "MSG_GET_ALL_TEMPLATES_FAILED", derr.Code)
	require.Contains(t, derr.Message, "Failed to get all templates")
}

func TestZaloHealthCheck_GetAllTemplatesAPIError(t *testing.T) {
	// Note: Cannot use t.Parallel() because we modify http.DefaultTransport

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Prepare encrypted token
	crypto := common.NewZaloTokenCrypto(dbEncryptionKey)
	plainToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	encryptedToken, err := crypto.Encrypt(context.Background(), plainToken)
	require.NoError(t, err)

	repo := newMockZaloRepository()
	repo.tokens[tenantID] = encryptedToken

	// Mock API error response (error != 0)
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			body := map[string]any{
				"error":    -124, // Token invalid error
				"message":  "Invalid access token",
				"data":     []any{},
				"metadata": map[string]any{"total": 0},
			}
			b, _ := json.Marshal(body)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(b)),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure due to API error
	require.NotNil(t, derr, "expected health check to fail due to API error")
	require.Equal(t, "MSG_GET_ALL_TEMPLATES_FAILED", derr.Code)
	require.Contains(t, derr.Message, "Failed to get all templates")
}

func TestZaloHealthCheck_GetAllTemplatesInvalidJSON(t *testing.T) {
	// Note: Cannot use t.Parallel() because we modify http.DefaultTransport

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Prepare encrypted token
	crypto := common.NewZaloTokenCrypto(dbEncryptionKey)
	plainToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	encryptedToken, err := crypto.Encrypt(context.Background(), plainToken)
	require.NoError(t, err)

	repo := newMockZaloRepository()
	repo.tokens[tenantID] = encryptedToken

	// Mock invalid JSON response
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert failure due to JSON parsing error
	require.NotNil(t, derr, "expected health check to fail due to invalid JSON")
	require.Equal(t, "MSG_GET_ALL_TEMPLATES_FAILED", derr.Code)
}

func TestZaloHealthCheck_EmptyTemplateList(t *testing.T) {
	// Note: Cannot use t.Parallel() because we modify http.DefaultTransport

	tenantID := uuid.New()
	dbEncryptionKey := "test-key-123456"

	// Prepare encrypted token
	crypto := common.NewZaloTokenCrypto(dbEncryptionKey)
	plainToken := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "test-app",
		SecretKey:    "test-secret",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now(),
	}
	encryptedToken, err := crypto.Encrypt(context.Background(), plainToken)
	require.NoError(t, err)

	repo := newMockZaloRepository()
	repo.tokens[tenantID] = encryptedToken

	// Mock successful response with empty template list
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			body := map[string]any{
				"error":    0,
				"message":  "",
				"data":     []any{},
				"metadata": map[string]any{"total": 0},
			}
			b, _ := json.Marshal(body)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(b)),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Create use case
	useCase := NewSmsTokenUseCase(repo, dbEncryptionKey)

	// Execute health check
	derr := useCase.ZaloHealthCheck(context.Background(), tenantID)

	// Assert success - empty template list is still a valid response
	require.Nil(t, derr, "expected health check to succeed with empty template list")
}
