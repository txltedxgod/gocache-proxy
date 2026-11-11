package proxy

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/txltedxgod/gocache-proxy/pkg/cache"
	"github.com/txltedxgod/gocache-proxy/pkg/metrics"
)

type ProxyHandler struct {
	target       *url.URL
	reverseProxy *httputil.ReverseProxy
	cache        *cache.LRUCache
	defaultTTL   time.Duration
}

func NewProxyHandler(targetURL string, lruCache *cache.LRUCache, defaultTTL time.Duration) (*ProxyHandler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	return &ProxyHandler{
		target:       target,
		reverseProxy: rp,
		cache:        lruCache,
		defaultTTL:   defaultTTL,
	}, nil
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only cache idempotent GET and HEAD requests
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		p.reverseProxy.ServeHTTP(w, r)
		return
	}

	cacheKey := r.Method + ":" + r.URL.RequestURI()

	// Check Cache
	if item, found := p.cache.Get(cacheKey); found {
		metrics.CacheHitsTotal.Inc()
		for k, vv := range item.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(item.StatusCode)
		_, _ = w.Write(item.Body)
		return
	}

	metrics.CacheMissesTotal.Inc()

	// Intercept response from upstream
	rec := newResponseRecorder(w)
	start := time.Now()
	p.reverseProxy.ServeHTTP(rec, r)
	duration := time.Since(start)

	metrics.UpstreamLatency.Observe(duration.Seconds())

	// Cache successful 2xx responses without persisting X-Cache: MISS
	if rec.statusCode >= 200 && rec.statusCode < 300 {
		cachedHeaders := rec.Header().Clone()
		cachedHeaders.Del("X-Cache")
		p.cache.Set(cacheKey, &cache.ResponseItem{
			StatusCode: rec.statusCode,
			Header:     cachedHeaders,
			Body:       bytes.Clone(rec.body.Bytes()),
			ExpiresAt:  time.Now().Add(p.defaultTTL),
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	body        *bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.statusCode = code
	if code >= 200 && code < 300 {
		r.ResponseWriter.Header().Set("X-Cache", "MISS")
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
