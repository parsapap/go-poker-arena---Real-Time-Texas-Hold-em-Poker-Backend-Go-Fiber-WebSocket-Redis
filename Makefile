.PHONY: help build run docker-up docker-down test test-coverage load-test clean

help:
	@echo "Available commands:"
	@echo "  make build         - Build the Go application"
	@echo "  make run           - Run the application locally"
	@echo "  make docker-up     - Start Docker containers (postgres + redis)"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make load-test     - Run WebSocket load test with 100 connections"
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
	go test ./internal/poker/... -v

test-coverage:
	go test ./internal/poker/... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

load-test:
	@echo "Building load test..."
	cd test && go mod init test 2>/dev/null || true
	cd test && go get github.com/gorilla/websocket
	cd test && go run ws_load_test.go

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker-compose down -v
