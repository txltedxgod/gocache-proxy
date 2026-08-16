# gocache-proxy

> High-performance HTTP caching reverse proxy with in-memory LRU storage and Prometheus metrics written in **Go (Golang)**.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-E6522C?style=flat-square&logo=prometheus)](https://prometheus.io)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

`#golang` `#reverse-proxy` `#http-cache` `#lru-cache` `#prometheus-metrics` `#devops` `#networking` `#performance`

---

## Features

- **Transparent HTTP Caching:** Intercepts and caches idempotent `GET` and `HEAD` requests in memory.
- **Thread-Safe Concurrent LRU:** O(1) cache lookups and evictions with granular mutex locks.
- **Cache Headers:** Injects `X-Cache: HIT` or `X-Cache: MISS` headers to every client response.
- **Prometheus Observability:** Native `/metrics` endpoint tracking cache hit/miss ratio and upstream request latencies.
- **Configurable TTL:** Automatic item expiration and memory recycling.

## Quick Start

### Build & Run Locally

```bash
# Run unit tests
go test ./...

# Build binary
go build -o gocache-proxy ./cmd/proxy

# Run proxy
./gocache-proxy -listen=:8080 -upstream=https://httpbin.org -ttl=60s
```

### With Docker Compose

```bash
docker compose up --build
```

Test cached endpoint:

```bash
# First request -> MISS
curl -i http://localhost:8080/json
# Second request -> HIT
curl -i http://localhost:8080/json
```

## CLI Configuration Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-listen` | `:8080` | Address and port to bind proxy |
| `-upstream` | `http://localhost:8000` | Target origin server URL |
| `-ttl` | `60s` | Default cache retention duration |
| `-capacity` | `5000` | Maximum number of items in LRU memory |

## Observability

- **Healthcheck:** `GET /healthz`
- **Prometheus Metrics:** `GET /metrics` (`gocache_proxy_cache_hits_total`, `gocache_proxy_cache_misses_total`, `gocache_proxy_upstream_latency_seconds`)
