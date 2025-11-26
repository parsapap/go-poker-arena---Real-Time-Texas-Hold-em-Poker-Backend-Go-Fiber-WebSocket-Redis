#!/bin/bash

echo "📦 Installing Test Dependencies..."
echo "=================================="

# Install miniredis for Redis mocking
go get github.com/alicebob/miniredis/v2

# Install SQLite driver for in-memory testing
go get gorm.io/driver/sqlite

# Tidy up dependencies
go mod tidy

echo ""
echo "✅ Test dependencies installed!"
echo ""
echo "Run tests with:"
echo "  go test ./... -v -cover"
echo ""
echo "Or use Docker:"
echo "  ./run-all-tests.sh"
