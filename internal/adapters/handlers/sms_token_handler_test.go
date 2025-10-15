package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
	common "github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/common"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms/provider"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucases "github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/stretchr/testify/require"
)

type inMemoryZaloTokenRepo struct {
	token *domain.ZaloToken
}

func (r *inMemoryZaloTokenRepo) Get(ctx context.Context) (*domain.ZaloToken, error) {
	if r.token == nil {
		return nil, fmt.Errorf("no token")
	}
	// return a copy to avoid external mutation
	t := *r.token
	return &t, nil
}

func (r *inMemoryZaloTokenRepo) Save(ctx context.Context, token *domain.ZaloToken) error {
	// save a copy
	t := *token
	r.token = &t
	return nil
}

// roundTripFunc allows stubbing http.DefaultTransport
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestAdminRefreshZaloToken_WithInvalidDBToken_RefreshesViaAdminEndpoint(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// Configure Zalo with no seed tokens to ensure normal provider bootstrap cannot use config
	cfg := conf.GetConfiguration()
	cfg.Sms.Zalo.ZaloBaseURL = "https://business.openapi.zalo.me"
	cfg.Sms.Zalo.ZaloSecretKey = "test-secret"
	cfg.Sms.Zalo.ZaloAppID = "test-app"
	cfg.Sms.Zalo.ZaloAccessToken = ""
	cfg.Sms.Zalo.ZaloRefreshToken = ""

	// Seed repo with an invalid, undecryptable token (invalid base64 strings)
	repo := &inMemoryZaloTokenRepo{token: &domain.ZaloToken{ID: 1, AccessToken: "!!!", RefreshToken: "!!!", UpdatedAt: time.Now(), ExpiresAt: time.Now()}}

	// Assert that normal provider cannot bootstrap/refresh due to invalid DB token and empty config
	_, err := provider.NewZaloProvider(context.Background(), cfg.Sms.Zalo, repo)
	require.Error(t, err, "expected NewZaloProvider to fail when DB token invalid and config has no tokens")

	// Stub OAuth refresh endpoint to return a successful token refresh
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = mockZaloTokenRefresh()

	// Build handler with usecase and repo
	usecase := ucases.NewSmsTokenUseCase(repo)
	h := NewSmsTokenHandler(usecase, nil, repo)

	r := gin.New()
	r.POST("/api/v1/admin/sms/zalo/token/refresh", h.RefreshZaloToken)

	// Before refresh, getting token should fail to decrypt
	tok, derr := usecase.GetZaloToken(context.Background())
	require.Nil(t, tok)
	require.NotNil(t, derr, "expected GetZaloToken to fail before refresh due to invalid ciphertext")

	// Call admin refresh endpoint
	reqBody := bytes.NewBufferString(`{"refresh_token":"admin-provided-rt"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sms/zalo/token/refresh", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Ensure token updated and decryptable
	tok, derr = usecase.GetZaloToken(context.Background())
	require.Nilf(t, derr, "expected token decryptable after refresh, got error: %v", derr)
	require.Equal(t, "new-access", tok.AccessToken)
	require.Equal(t, "new-refresh", tok.RefreshToken)
}

func mockZaloTokenRefresh() http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "oauth.zaloapp.com" && req.URL.Path == "/v4/oa/access_token" && req.Method == http.MethodPost {
			body := map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    "3600",
				"error":         0,
				"message":       "",
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
}

func TestAdminRefreshZaloToken_WithExpiredDBToken_RefreshesViaAdminEndpoint(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	// Ensure a deterministic encryption key for test
	cfg := conf.GetConfiguration()
	cfg.DbEncryptionKey = "test-key"
	cfg.Sms.Zalo.ZaloBaseURL = "https://business.openapi.zalo.me"
	cfg.Sms.Zalo.ZaloSecretKey = "test-secret"
	cfg.Sms.Zalo.ZaloAppID = "test-app"
	cfg.Sms.Zalo.ZaloAccessToken = ""
	cfg.Sms.Zalo.ZaloRefreshToken = ""

	// Seed repo with a decryptable but expired token
	crypto := common.NewZaloTokenCrypto(cfg.DbEncryptionKey)
	expiredPlain := &domain.ZaloToken{ID: 1, AccessToken: "expired-access", RefreshToken: "expired-refresh", UpdatedAt: time.Now(), ExpiresAt: time.Now().Add(-1 * time.Hour)}
	encrypted, err := crypto.Encrypt(context.Background(), expiredPlain)
	require.NoError(t, err, "failed to encrypt seed token")
	repo := &inMemoryZaloTokenRepo{token: encrypted}

	// Stub OAuth: fail when using expired-refresh; succeed when using admin-provided-rt
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "oauth.zaloapp.com" && req.URL.Path == "/v4/oa/access_token" && req.Method == http.MethodPost {
			b, _ := io.ReadAll(req.Body)
			_ = req.Body.Close()
			vals, _ := url.ParseQuery(string(b))
			refresh := vals.Get("refresh_token")
			if refresh == "admin-provided-rt" {
				ok := map[string]any{"access_token": "new-access", "refresh_token": "new-refresh", "expires_in": "3600", "error": 0, "message": ""}
				jb, _ := json.Marshal(ok)
				resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(jb))}
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			}
			// simulate failure for expired-refresh
			errBody := map[string]any{"error": 1, "message": "invalid refresh token"}
			jb, _ := json.Marshal(errBody)
			resp := &http.Response{StatusCode: http.StatusBadRequest, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(jb))}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	// Normal provider bootstrap should fail because token is expired and refresh with expired-refresh fails
	_, err = provider.NewZaloProvider(context.Background(), cfg.Sms.Zalo, repo)
	require.Error(t, err, "expected NewZaloProvider to fail when token expired and refresh fails")

	// Now call admin refresh endpoint with a valid admin-provided refresh token
	usecase := ucases.NewSmsTokenUseCase(repo)
	h := NewSmsTokenHandler(usecase, nil, repo)
	r := gin.New()
	r.POST("/api/v1/admin/sms/zalo/token/refresh", h.RefreshZaloToken)

	reqBody := bytes.NewBufferString(`{"refresh_token":"admin-provided-rt"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sms/zalo/token/refresh", reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Verify token was updated and decryptable
	newTok, derr := usecase.GetZaloToken(context.Background())
	require.Nilf(t, derr, "expected token decryptable after admin refresh, got error: %v", derr)
	require.Equal(t, "new-access", newTok.AccessToken)
	require.Equal(t, "new-refresh", newTok.RefreshToken)
}

// fakeSmsTokenUseCase allows injecting errors for handler negative tests
type fakeSmsTokenUseCase struct {
	getTokenFunc func(ctx context.Context) (*domain.ZaloToken, error)
	refreshErr   error
}

func (f *fakeSmsTokenUseCase) GetZaloToken(ctx context.Context) (*domain.ZaloToken, *domainerrors.DomainError) {
	if f.getTokenFunc != nil {
		tok, err := f.getTokenFunc(ctx)
		if err != nil {
			return nil, domainerrors.NewInternalError("X", err.Error())
		}
		return tok, nil
	}
	return &domain.ZaloToken{}, nil
}

func (f *fakeSmsTokenUseCase) SetZaloToken(ctx context.Context, accessToken, refreshToken string) *domainerrors.DomainError {
	return nil
}

func (f *fakeSmsTokenUseCase) RefreshZaloToken(ctx context.Context, refreshToken string) *domainerrors.DomainError {
	if f.refreshErr != nil {
		return domainerrors.NewInternalError("MSG_REFRESH_TOKEN_FAILED", f.refreshErr.Error())
	}
	return nil
}

func TestRefreshZaloToken_InvalidBody_ReturnsBadRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := NewSmsTokenHandler(&fakeSmsTokenUseCase{}, nil, nil)
	r := gin.New()
	r.POST("/api/v1/admin/sms/zalo/token/refresh", h.RefreshZaloToken)

	// Missing refresh_token field
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sms/zalo/token/refresh", bytes.NewBufferString(`{"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusBadRequest, w.Code, "body: %s", w.Body.String())
}

func TestRefreshZaloToken_UseCaseRefreshError_ReturnsInternal(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := &fakeSmsTokenUseCase{refreshErr: fmt.Errorf("boom")}
	h := NewSmsTokenHandler(uc, nil, nil)
	r := gin.New()
	r.POST("/api/v1/admin/sms/zalo/token/refresh", h.RefreshZaloToken)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sms/zalo/token/refresh", bytes.NewBufferString(`{"refresh_token":"rt"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusInternalServerError, w.Code, "body: %s", w.Body.String())
}

func TestRefreshZaloToken_GetTokenError_ReturnsInternal(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := &fakeSmsTokenUseCase{
		getTokenFunc: func(ctx context.Context) (*domain.ZaloToken, error) { return nil, fmt.Errorf("cannot get token") },
	}
	h := NewSmsTokenHandler(uc, nil, nil)
	r := gin.New()
	r.POST("/api/v1/admin/sms/zalo/token/refresh", h.RefreshZaloToken)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sms/zalo/token/refresh", bytes.NewBufferString(`{"refresh_token":"rt"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusInternalServerError, w.Code, "body: %s", w.Body.String())
}

func TestGetZaloToken_Success_ReturnsOK(t *testing.T) {
	gintMode := gin.Mode()
	defer gin.SetMode(gintMode)
	gin.SetMode(gin.TestMode)

	// Fake usecase returns a token
	uc := &fakeSmsTokenUseCase{
		getTokenFunc: func(ctx context.Context) (*domain.ZaloToken, error) {
			return &domain.ZaloToken{
				AccessToken:  "tok-access",
				RefreshToken: "tok-refresh",
				UpdatedAt:    time.Now(),
				ExpiresAt:    time.Now().Add(10 * time.Minute),
			}, nil
		},
	}
	h := NewSmsTokenHandler(uc, nil, nil)
	r := gin.New()
	r.GET("/api/v1/admin/sms/zalo/token", h.GetZaloToken)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sms/zalo/token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var resp struct {
		Status int `json:"status"`
		Data   struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresAt    string `json:"expires_at"`
			UpdatedAt    string `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal response")
	require.Equal(t, "tok-access", resp.Data.AccessToken)
	require.Equal(t, "tok-refresh", resp.Data.RefreshToken)
	_, err = time.Parse(time.RFC3339, resp.Data.ExpiresAt)
	require.NoError(t, err, "expires_at not RFC3339")
	_, err = time.Parse(time.RFC3339, resp.Data.UpdatedAt)
	require.NoError(t, err, "updated_at not RFC3339")
}

func TestGetZaloToken_UsecaseError_ReturnsInternal(t *testing.T) {
	gintMode := gin.Mode()
	defer gin.SetMode(gintMode)
	gin.SetMode(gin.TestMode)

	uc := &fakeSmsTokenUseCase{
		getTokenFunc: func(ctx context.Context) (*domain.ZaloToken, error) { return nil, fmt.Errorf("fail") },
	}
	h := NewSmsTokenHandler(uc, nil, nil)
	r := gin.New()
	r.GET("/api/v1/admin/sms/zalo/token", h.GetZaloToken)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sms/zalo/token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusInternalServerError, w.Code, "body: %s", w.Body.String())
}

func TestGetZaloHealth_ProviderNotFound_ReturnsInternal(t *testing.T) {
	gintMode := gin.Mode()
	defer gin.SetMode(gintMode)
	gin.SetMode(gin.TestMode)

	cfg := conf.GetConfiguration()
	// Ensure Zalo is not configured so provider lookup fails
	cfg.Sms.Zalo.ZaloAppID = ""
	cfg.Sms.Zalo.ZaloAccessToken = ""
	cfg.Sms.Zalo.ZaloRefreshToken = ""

	svc, _ := sms.NewSMSService(&cfg.Sms, nil)
	h := NewSmsTokenHandler(&fakeSmsTokenUseCase{}, svc, nil)

	r := gin.New()
	r.GET("/api/v1/admin/sms/zalo/health", h.GetZaloHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sms/zalo/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusInternalServerError, w.Code, "body: %s", w.Body.String())
}

func TestGetZaloHealth_Healthy_ReturnsOK(t *testing.T) {
	gintMode := gin.Mode()
	defer gin.SetMode(gintMode)
	gin.SetMode(gin.TestMode)

	// Configure Zalo to initialize provider
	cfg := conf.GetConfiguration()
	cfg.Sms.Zalo.ZaloBaseURL = "https://business.openapi.zalo.me"
	cfg.Sms.Zalo.ZaloSecretKey = "test-secret"
	cfg.Sms.Zalo.ZaloAppID = "test-app"
	cfg.Sms.Zalo.ZaloAccessToken = "seed-access"
	cfg.Sms.Zalo.ZaloRefreshToken = "seed-refresh"

	// Mock template list endpoint
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			body := map[string]any{"error": 0, "message": "", "data": []any{}, "metadata": map[string]any{"total": 0}}
			b, _ := json.Marshal(body)
			resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(b))}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	repo := &inMemoryZaloTokenRepo{}
	svc, _ := sms.NewSMSService(&cfg.Sms, repo)
	h := NewSmsTokenHandler(&fakeSmsTokenUseCase{}, svc, repo)

	r := gin.New()
	r.GET("/api/v1/admin/sms/zalo/health", h.GetZaloHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sms/zalo/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var resp struct {
		Status int               `json:"status"`
		Data   map[string]string `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "failed to unmarshal response")
	require.Equal(t, "healthy", resp.Data["status"])
}

func TestGetZaloHealth_UpstreamError_ReturnsInternal(t *testing.T) {
	gintMode := gin.Mode()
	defer gin.SetMode(gintMode)
	gin.SetMode(gin.TestMode)

	// Configure Zalo to initialize provider
	cfg := conf.GetConfiguration()
	cfg.Sms.Zalo.ZaloBaseURL = "https://business.openapi.zalo.me"
	cfg.Sms.Zalo.ZaloSecretKey = "test-secret"
	cfg.Sms.Zalo.ZaloAppID = "test-app"
	cfg.Sms.Zalo.ZaloAccessToken = "seed-access"
	cfg.Sms.Zalo.ZaloRefreshToken = "seed-refresh"

	// Mock template list endpoint with upstream error
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "business.openapi.zalo.me" && req.URL.Path == "/template/all" && req.Method == http.MethodGet {
			body := map[string]any{"error": 1, "message": "upstream error", "data": []any{}, "metadata": map[string]any{"total": 0}}
			b, _ := json.Marshal(body)
			resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(b))}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
	})

	repo := &inMemoryZaloTokenRepo{}
	svc, _ := sms.NewSMSService(&cfg.Sms, repo)
	h := NewSmsTokenHandler(&fakeSmsTokenUseCase{}, svc, repo)

	r := gin.New()
	r.GET("/api/v1/admin/sms/zalo/health", h.GetZaloHealth)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sms/zalo/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equalf(t, http.StatusInternalServerError, w.Code, "body: %s", w.Body.String())
}
