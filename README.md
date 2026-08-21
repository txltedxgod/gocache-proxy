# gocache-proxy

[![Go CI](https://github.com/txltedxgod/gocache-proxy/actions/workflows/ci.yml/badge.svg)](https://github.com/txltedxgod/gocache-proxy/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go 1.22](https://img.shields.io/badge/go-1.22-00ADD8.svg?logo=go&logoColor=white)](https://golang.org/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C.svg?logo=prometheus&logoColor=white)](https://prometheus.io/)


> High-throughput **HTTP Caching Reverse Proxy** featuring concurrent in-memory LRU eviction, single-flight stampede protection, and native **Prometheus metrics exporter** built in **Go 1.22**.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-E6522C?style=flat-square&logo=prometheus)](https://prometheus.io)
[![CI](https://img.shields.io/badge/CI-Passing-238636?style=flat-square&logo=githubactions)](https://github.com/txltedxgod/gocache-proxy/actions)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

`#golang` `#reverse-proxy` `#caching` `#lru-cache` `#prometheus` `#http-proxy` `#performance`

---

## 🏛️ Request Lifecycle & Cache Resolution

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client Request
    participant Proxy as GoCache Reverse Proxy
    participant Cache as LRU Memory Cache
    participant Upstream as Origin / Upstream Server

    Client->>Proxy: GET /api/v1/resource
    Proxy->>Cache: Lookup SHA256 Key
    alt Cache HIT (Valid TTL)
        Cache-->>Proxy: Return Cached Response & Headers
        Proxy->>Proxy: Increment cache_hits_total Counter
        Proxy-->>Client: 200 OK (X-Cache: HIT)
    else Cache MISS / Expired
        Proxy->>Proxy: Increment cache_misses_total Counter
        Proxy->>Upstream: SingleFlight Fetch Upstream
        Upstream-->>Proxy: Upstream Response Body
        Proxy->>Cache: Store in LRU Cache (Set TTL)
        Proxy-->>Client: 200 OK (X-Cache: MISS)
    end
```

---

## Features

- **Concurrent LRU Cache:** Lock-striped in-memory cache with configurable item capacity and default TTL.
- **Cache-Control Header Compliance:** Parses `max-age`, `no-cache`, and `private` upstream directives.
- **Single-Flight Deduplication:** Multiple concurrent requests for the same uncached URL collapse into a single upstream request, preventing cache stampedes.
- **Prometheus Metrics:** Exposes `gocache_requests_total`, `gocache_hits_total`, and `gocache_latency_seconds` on `/metrics`.

## Quick Start

```bash
# Run reverse proxy targeting upstream API
go run cmd/proxy/main.go \
  -listen=:8080 \
  -upstream=https://api.github.com \
  -cache-size=1000 \
  -ttl=60s
```