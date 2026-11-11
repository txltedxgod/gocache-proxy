package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/txltedxgod/gocache-proxy/pkg/cache"
)

func TestNewProxyHandler_InvalidURL(t *testing.T) {
	_, err := NewProxyHandler("://bad-url", cache.NewLRUCache(10), time.Minute)
	if err == nil {
		t.Fatal("expected error for invalid upstream URL")
	}
}

func TestProxyHandler_CacheMissThenHit(t *testing.T) {
	var upstreamHits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHits.Add(1)
		w.Header().Set("X-Upstream", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	}))
	defer upstream.Close()

	lru := cache.NewLRUCache(16)
	handler, err := NewProxyHandler(upstream.URL, lru, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error creating proxy handler: %v", err)
	}

	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	resp1, err := http.Get(proxyServer.URL + "/items")
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on MISS, got %d", resp1.StatusCode)
	}
	if string(body1) != "hello world" {
		t.Fatalf("unexpected body on MISS: %q", string(body1))
	}
	if got := resp1.Header.Get("X-Cache"); got != "MISS" {
		t.Fatalf("expected X-Cache=MISS, got %q", got)
	}
	if upstreamHits.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", upstreamHits.Load())
	}

	resp2, err := http.Get(proxyServer.URL + "/items")
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on HIT, got %d", resp2.StatusCode)
	}
	if string(body2) != "hello world" {
		t.Fatalf("unexpected body on HIT: %q", string(body2))
	}
	xCacheVals := resp2.Header.Values("X-Cache")
	if len(xCacheVals) != 1 || xCacheVals[0] != "HIT" {
		t.Fatalf("expected single X-Cache=HIT header, got %v", xCacheVals)
	}
	if upstreamHits.Load() != 1 {
		t.Fatalf("expected cached HIT not to call upstream, got %d hits", upstreamHits.Load())
	}
}

func TestProxyHandler_SkipsCacheForPost(t *testing.T) {
	var upstreamHits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHits.Add(1)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	defer upstream.Close()

	handler, err := NewProxyHandler(upstream.URL, cache.NewLRUCache(16), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error creating proxy handler: %v", err)
	}

	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/submit", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Header().Get("X-Cache") != "" {
			t.Fatalf("did not expect X-Cache header on POST, got %q", rec.Header().Get("X-Cache"))
		}
	}

	if upstreamHits.Load() != 2 {
		t.Fatalf("expected POST requests to bypass cache (2 hits), got %d", upstreamHits.Load())
	}
}
