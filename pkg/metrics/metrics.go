package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CacheHitsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gocache_proxy_cache_hits_total",
		Help: "The total number of cached HTTP responses served",
	})

	CacheMissesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gocache_proxy_cache_misses_total",
		Help: "The total number of cache misses requiring upstream fetch",
	})

	UpstreamLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "gocache_proxy_upstream_latency_seconds",
		Help:    "Latency of requests proxied to the upstream target in seconds",
		Buckets: prometheus.DefBuckets,
	})
)
