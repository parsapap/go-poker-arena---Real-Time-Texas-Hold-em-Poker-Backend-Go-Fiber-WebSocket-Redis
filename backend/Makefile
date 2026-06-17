.PHONY: help build run docker-up docker-down test test-coverage load-test stress-test clean

help:
	@echo "Available commands:"
	@echo "  make build         - Build the Go application"
	@echo "  make run           - Run the application locally"
	@echo "  make docker-up     - Start Docker containers (postgres + redis)"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make load-test     - Run WebSocket load test with 100 connections"
	@echo "  make stress-test   - Run stress test with 1000 connections"
	@echo "  make clean         - Clean build artifacts"

build:
	go build -o bin/poker-server ./cmd/server

run:
	go run ./cmd/server/main.go

docker-up:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5

docker-down:
	docker-compose down

test:
	go test ./... -v

test-coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	@echo "Requires a running server at localhost:8080 plus Postgres and Redis"
	go test -tags=integration ./test/integration/... -v

load-test:
	@echo "Running WebSocket load test (server must be running)..."
	go run ./test/loadtest

stress-test:
	@echo "Running stress test (server must be running)..."
	go run ./test/stress

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker-compose down -v
