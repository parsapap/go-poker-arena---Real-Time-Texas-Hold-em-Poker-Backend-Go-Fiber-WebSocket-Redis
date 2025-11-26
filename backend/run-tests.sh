#!/bin/bash

# Run backend tests in Docker with Go environment

docker run --rm \
  -v $(pwd):/app \
  -w /app \
  -e POSTGRES_HOST=localhost \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=poker \
  -e POSTGRES_PASSWORD=poker123 \
  -e POSTGRES_DB=poker_arena_test \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  -e JWT_SECRET=test-secret \
  golang:1.21-alpine \
  sh -c "go test ./... -v -cover"
