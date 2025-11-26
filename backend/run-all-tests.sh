#!/bin/bash

echo "🧪 Running All Backend Tests with Coverage"
echo "=========================================="

# Run tests in Docker with Go 1.21
docker run --rm \
  -v $(pwd):/app \
  -w /app \
  -e POSTGRES_HOST=localhost \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=test \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  -e JWT_SECRET=test-secret \
  -e ENV=test \
  -e LOG_LEVEL=error \
  golang:1.21-alpine \
  sh -c "
    apk add --no-cache git &&
    go test ./... -v -cover -coverprofile=coverage.out -covermode=atomic &&
    go tool cover -func=coverage.out | tail -1
  "

echo ""
echo "✅ Test run complete!"
echo ""
echo "To view detailed coverage:"
echo "  docker run --rm -v \$(pwd):/app -w /app golang:1.21-alpine go tool cover -html=coverage.out -o coverage.html"
