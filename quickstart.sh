#!/bin/bash

# metadata.party - Quick Start Script
# This script helps you get started with the metadata extraction API

set -e

echo "🎉 Welcome to metadata.party!"
echo "=============================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go $GO_VERSION detected"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod download
echo "✅ Dependencies installed"
echo ""

# Build the application
echo "🔨 Building the application..."
go build -o metadata-api main.go
echo "✅ Build complete"
echo ""

# Start the server in the background
echo "🚀 Starting the server..."
./metadata-api &
SERVER_PID=$!
echo "✅ Server started (PID: $SERVER_PID)"
echo ""

# Wait for server to be ready
echo "⏳ Waiting for server to be ready..."
sleep 2

# Test the API
echo "🧪 Testing the API..."
echo ""

# Health check
echo "1. Health Check:"
curl -s http://localhost:8080/health | jq .
echo ""

# Extract metadata
echo "2. Extracting metadata from example.com:"
curl -s -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' | jq .
echo ""

echo "✅ Everything is working!"
echo ""
echo "📝 Server is running at http://localhost:8080"
echo "   - API docs: http://localhost:8080/"
echo "   - Health check: http://localhost:8080/health"
echo "   - Extract metadata: POST http://localhost:8080/extract"
echo ""
echo "🛑 To stop the server, run: kill $SERVER_PID"
echo ""
echo "📚 For more examples, run: ./examples.sh"
echo "🐳 To run with Docker: docker-compose up"
echo ""
echo "Happy metadata extracting! 🎉"

