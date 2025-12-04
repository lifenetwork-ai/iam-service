package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	"github.com/stretchr/testify/require"
)

// fakeZaloTokenRepo is an in-memory implementation of ZaloTokenRepository used for testing
type fakeZaloTokenRepo struct {
	stored *domain.ZaloToken // stored is always encrypted at rest, like DB
}

func (f *fakeZaloTokenRepo) Get(ctx context.Context, tenantID uuid.UUID) (*domain.ZaloToken, error) {
	if f.stored == nil || f.stored.TenantID != tenantID {
		return nil, nil
	}
	return f.stored, nil
}

func (f *fakeZaloTokenRepo) Save(ctx context.Context, token *domain.ZaloToken) error {
	// Upsert by tenant ID semantics
	f.stored = token
	return nil
}

func (f *fakeZaloTokenRepo) GetAll(ctx context.Context) ([]*domain.ZaloToken, error) {
	if f.stored == nil {
		return []*domain.ZaloToken{}, nil
	}
	return []*domain.ZaloToken{f.stored}, nil
}

func (f *fakeZaloTokenRepo) Delete(ctx context.Context, tenantID uuid.UUID) error {
	if f.stored != nil && f.stored.TenantID == tenantID {
		f.stored = nil
	}
	return nil
}

// fakeTenantRepo maps a name to a fixed tenant
type fakeTenantRepo struct{ id uuid.UUID }

func (f *fakeTenantRepo) Create(tenant *domain.Tenant) error                     { return nil }
func (f *fakeTenantRepo) Update(tenant *domain.Tenant) error                     { return nil }
func (f *fakeTenantRepo) Delete(id uuid.UUID) error                              { return nil }
func (f *fakeTenantRepo) GetByID(id uuid.UUID) (*domain.Tenant, error)           { return &domain.Tenant{ID: id, Name: "tenant"}, nil }
func (f *fakeTenantRepo) List() ([]*domain.Tenant, error)                        { return nil, nil }
func (f *fakeTenantRepo) GetByName(name string) (*domain.Tenant, error)          { return &domain.Tenant{ID: f.id, Name: name}, nil }

// stubTransport rewrites requests to the Zalo OAuth host to a test server
type stubTransport struct {
	underlying http.RoundTripper
	oauthURL   *url.URL
	baseURL    *url.URL
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Host {
	case "oauth.zaloapp.com": // constants.ZaloOAuthBaseURL host
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = s.oauthURL.Scheme
		cloned.URL.Host = s.oauthURL.Host
		return s.underlying.RoundTrip(cloned)
	case "business.openapi.zalo.me": // constants.ZaloBaseURL host
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = s.baseURL.Scheme
		cloned.URL.Host = s.baseURL.Host
		return s.underlying.RoundTrip(cloned)
	}
	return s.underlying.RoundTrip(req)
}

func TestZaloProvider_RefreshOnSendOTP(t *testing.T) {
	// Ensure deterministic encryption key for tests
	conf.GetConfiguration().DbEncryptionKey = "unit-test-db-key"

	// Arrange tenant and initial token
	tenantID := uuid.New()
	tenantName := "tenant-a"

	// Start base API server to simulate Zalo business API
	var sendCalls int32
	var seenAccessTokens []string
	var refreshed int32
	baseSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/message/template" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// capture access token header used by client
		tok := r.Header.Get("access_token")
		seenAccessTokens = append(seenAccessTokens, tok)
		call := atomic.AddInt32(&sendCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if call == 1 {
			_ = enc.Encode(map[string]any{"error": -124, "message": "token invalid"})
			return
		}
		// After refresh, only succeed if the new token is used
		if atomic.LoadInt32(&refreshed) == 1 && tok == "newAT" {
			_ = enc.Encode(map[string]any{
				"error":   0,
				"message": "success",
				"data": map[string]any{
					"sent_time":    time.Now().Format(time.RFC3339),
					"sending_mode": "real",
					"quota": map[string]string{
						"remainingQuota": "99",
						"dailyQuota":     "100",
					},
					"msg_id": "abc",
				},
			})
			return
		}
		_ = enc.Encode(map[string]any{"error": -124, "message": "token invalid"})
	}))
	defer baseSrv.Close()

	// Start OAuth server to simulate token refresh endpoint
	oauthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v4/oa/access_token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		atomic.StoreInt32(&refreshed, 1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "newAT",
			"refresh_token": "newRT",
			"expires_in":    "3600",
			"message":       "ok",
			"error":         0,
		})
	}))
	defer oauthSrv.Close()

	// Replace global default transport so Zalo client constructed inside provider
	// uses our stub that rewrites OAuth host to the oauth test server
	origDefault := http.DefaultTransport
	oauthURL, _ := url.Parse(oauthSrv.URL)
	baseURL, _ := url.Parse(baseSrv.URL)
	http.DefaultTransport = &stubTransport{underlying: origDefault, oauthURL: oauthURL, baseURL: baseURL}
	defer func() { http.DefaultTransport = origDefault }()

	// Prepare encrypted token in repo
	repo := &fakeZaloTokenRepo{}
	crypto := common.NewZaloTokenCrypto(conf.GetConfiguration().DbEncryptionKey)
	plaintext := &domain.ZaloToken{
		ID:           1,
		TenantID:     tenantID,
		AppID:        "app-123",
		SecretKey:    "sek-xyz",
		AccessToken:  "oldAT",
		RefreshToken: "oldRT",
		ExpiresAt:    time.Now().Add(1 * time.Hour), // not near expiry to avoid proactive refresh
		UpdatedAt:    time.Now(),
	}
	encrypted, err := crypto.Encrypt(context.Background(), plaintext)
	require.NoError(t, err)
	repo.stored = encrypted

	// Provider configuration; base URL points to our test server
	cfg := conf.ZaloConfiguration{
		ZaloBaseURL:    baseSrv.URL,
		ZaloSecretKey:  "dummy",
		ZaloAppID:      "dummy",
		ZaloTemplateID: 123,
	}

	prov, err := NewZaloProvider(context.Background(), cfg, repo, &fakeTenantRepo{id: tenantID})
	require.NoError(t, err)

	// Act: send OTP which should trigger refresh on first -124, update in-memory client, then succeed
	err = prov.SendOTP(context.Background(), tenantName, "+84987654321", "123456", 5*time.Minute)
	require.NoError(t, err, "SendOTP should succeed after refreshing tokens and retrying")

	// Assert: two calls to send endpoint
	require.Equal(t, int32(2), atomic.LoadInt32(&sendCalls), "expected two send attempts: before and after refresh")
	// First call uses old access token, second uses new one from refresh
	require.Len(t, seenAccessTokens, 2)
	require.Equal(t, "oldAT", seenAccessTokens[0])
	require.Equal(t, "newAT", seenAccessTokens[1])

	// Assert repo has saved the new tokens (stored encrypted, verify by decrypting)
	savedEncrypted, err := repo.Get(context.Background(), tenantID)
	require.NoError(t, err)
	decrypted, err := crypto.Decrypt(context.Background(), savedEncrypted)
	require.NoError(t, err)
	require.Equal(t, "app-123", decrypted.AppID)
	require.Equal(t, "sek-xyz", decrypted.SecretKey)
	require.Equal(t, "newAT", decrypted.AccessToken)
	require.Equal(t, "newRT", decrypted.RefreshToken)
}

