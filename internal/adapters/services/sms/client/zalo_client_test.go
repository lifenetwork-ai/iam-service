package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

type stubTransport struct {
	underlying http.RoundTripper
	baseURL    *url.URL
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "business.openapi.zalo.me" {
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = s.baseURL.Scheme
		cloned.URL.Host = s.baseURL.Host
		return s.underlying.RoundTrip(cloned)
	}
	return s.underlying.RoundTrip(req)
}

func TestZaloClient_UpdateTokensAffectsSubsequentRequests(t *testing.T) {
	var calls int32
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		seen = append(seen, r.Header.Get("access_token"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"error": 0})
	}))
	defer srv.Close()

	orig := http.DefaultTransport
	u, _ := url.Parse(srv.URL)
	http.DefaultTransport = &stubTransport{underlying: orig, baseURL: u}
	defer func() { http.DefaultTransport = orig }()

	cli, err := NewZaloClient(context.Background(), "", "sek", "app", "oldAT", "oldRT")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	// First request uses old token
	if _, err := cli.SendOTP(context.Background(), "+84123", "1234", 1); err != nil {
		t.Fatalf("send1: %v", err)
	}
	// Update tokens
	_ = cli.UpdateTokens(context.Background(), &ZaloTokenRefreshResponse{AccessToken: "newAT", RefreshToken: "newRT"})
	// Second request should use new token
	time.Sleep(10 * time.Millisecond)
	if _, err := cli.SendOTP(context.Background(), "+84123", "1234", 1); err != nil {
		t.Fatalf("send2: %v", err)
	}

	if len(seen) != 2 || seen[0] != "oldAT" || seen[1] != "newAT" {
		t.Fatalf("unexpected tokens seen: %#v", seen)
	}
}

