FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/gocache-proxy ./cmd/proxy

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/gocache-proxy /usr/local/bin/gocache-proxy

EXPOSE 8080
ENTRYPOINT ["gocache-proxy"]
CMD ["-listen=:8080", "-upstream=http://localhost:8000"]
