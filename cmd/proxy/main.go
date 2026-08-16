package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/txltedxgod/gocache-proxy/pkg/cache"
	"github.com/txltedxgod/gocache-proxy/pkg/proxy"
)

var (
	listenAddr  = flag.String("listen", ":8080", "Server listen address")
	upstreamURL = flag.String("upstream", "http://localhost:8000", "Upstream target server URL")
	cacheTTL    = flag.Duration("ttl", 60*time.Second, "Default response cache TTL")
	maxCapacity = flag.Int("capacity", 5000, "Maximum number of cached entries in LRU")
)

func main() {
	flag.Parse()

	log.Printf("[gocache-proxy] Starting on %s -> Upstream: %s (TTL: %v, Capacity: %d)\n",
		*listenAddr, *upstreamURL, *cacheTTL, *maxCapacity)

	lruCache := cache.NewLRUCache(*maxCapacity)
	proxyHandler, err := proxy.NewProxyHandler(*upstreamURL, lruCache, *cacheTTL)
	if err != nil {
		log.Fatalf("Failed to initialize reverse proxy: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK (Cached items: %d)\n", lruCache.Len())
	})
	mux.Handle("/", proxyHandler)

	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[gocache-proxy] Shutting down gracefully...")
}
