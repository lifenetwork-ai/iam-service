package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	repos "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucases "github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	"github.com/lifenetwork-ai/iam-service/internal/workers"
)

type rewriteTransport struct {
	// hostnames we want to capture and redirect to test server
	hosts []string
	// target test server URL
	target *url.URL
	// next underlying transport
	next http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, h := range rt.hosts {
		if strings.EqualFold(req.URL.Host, h) {
			// Clone the request to avoid mutating the original
			cloned := req.Clone(req.Context())
			cloned.URL.Scheme = rt.target.Scheme
			cloned.URL.Host = rt.target.Host
			// Keep original Path and Query; handler will route by Path
			return rt.next.RoundTrip(cloned)
		}
	}
	return rt.next.RoundTrip(req)
}

type oauthResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	Message      string `json:"message"`
	Error        int    `json:"error"`
}

type templatesResp struct {
	Error    int           `json:"error"`
	Message  string        `json:"message"`
	Data     []interface{} `json:"data"`
	Metadata struct {
		Total int `json:"total"`
	} `json:"metadata"`
}

func TestZaloTokenFlow_Retrieve_Refresh_Health_WorkerInterval(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// In-memory DB with GORM (SQLite)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.ZaloToken{}))

	repo := repos.NewZaloTokenRepository(db)
	uc := ucases.NewSmsTokenUseCase(repo, "integration-test-key")

	// Test HTTP server simulating Zalo OAuth and Business APIs
	var mu sync.Mutex
	var refreshCount int
	currentAccess := ""
	currentRefresh := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()

		switch {
		case strings.HasSuffix(r.URL.Path, "/v4/oa/access_token"):
			// Simulate token rotation on each call
			refreshCount++
			currentAccess = "acc-" + strings.Repeat("x", refreshCount)
			currentRefresh = "ref-" + strings.Repeat("y", refreshCount)
			_ = json.NewEncoder(w).Encode(oauthResp{
				AccessToken:  currentAccess,
				RefreshToken: currentRefresh,
				ExpiresIn:    "60",
				Message:      "ok",
				Error:        0,
			})
		case strings.HasPrefix(r.URL.Path, "/template/all"):
			// Health check path - ensure we get whatever token is set
			_ = json.NewEncoder(w).Encode(templatesResp{Error: 0, Message: "ok"})
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":404}`))
		}
	}))
	defer server.Close()

	// Override default transport to redirect Zalo hosts to our test server
	orig := http.DefaultTransport
	u, _ := url.Parse(server.URL)
	http.DefaultTransport = &rewriteTransport{
		hosts:  []string{"oauth.zaloapp.com", "business.openapi.zalo.me"},
		target: u,
		next:   orig,
	}
	defer func() { http.DefaultTransport = orig }()

	tenantID := uuid.New()

	// 1) Create without access token triggers refresh
	derr := uc.CreateOrUpdateZaloToken(ctx, tenantID, "app-id", "secret", "seed-refresh", "", "")
	require.Nil(t, derr, "Failed to create Zalo token")

	encTok, derr := uc.GetEncryptedZaloToken(ctx, tenantID)
	require.Nil(t, derr, "Failed to get encrypted Zalo token")
	require.NotEmpty(t, encTok.AccessToken)

	plainTok, derr := uc.GetZaloToken(ctx, tenantID)
	require.Nil(t, derr, "Failed to get plain Zalo token")
	require.Equal(t, currentAccess, plainTok.AccessToken)
	require.Equal(t, currentRefresh, plainTok.RefreshToken)
	require.WithinDuration(t, time.Now().Add(60*time.Second), plainTok.ExpiresAt, 2*time.Second)

	// 2) Health check attempts to list templates
	derr = uc.ZaloHealthCheck(ctx, tenantID)
	require.Nil(t, derr)

	// 3) Manual refresh updates tokens
	prevAccess := plainTok.AccessToken
	derr = uc.RefreshZaloToken(ctx, tenantID, plainTok.RefreshToken)
	require.Nil(t, derr)

	plainTok2, derr := uc.GetZaloToken(ctx, tenantID)
	require.Nil(t, derr)
	require.NotEqual(t, prevAccess, plainTok2.AccessToken)

	// 4) Worker refreshes per interval regardless of expiry
	w := workers.NewZaloRefreshTokenWorker(repo, "integration-test-key")
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Start with a short interval
	go w.Start(wctx, 100*time.Millisecond)

	// Wait until the token rotates compared to plainTok2.AccessToken
	require.Eventually(t, func() bool {
		pt, derr := uc.GetZaloToken(ctx, tenantID)
		return derr == nil && pt.AccessToken != plainTok2.AccessToken
	}, 2*time.Second, 50*time.Millisecond)
	cancel()

	plainTok3, derr := uc.GetZaloToken(ctx, tenantID)
	require.Nil(t, derr)
	// After worker ran, tokens should have rotated at least once more
	require.NotEqual(t, plainTok2.AccessToken, plainTok3.AccessToken)
}
