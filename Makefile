.PHONY: all test run lint clean docker-build

all: test

test:
	go test -race ./...

run:
	go run ./cmd/proxy

lint:
	go vet ./...

clean:
	@echo "Cleaning artifacts..."

docker-build:
	docker build -t app:latest .
