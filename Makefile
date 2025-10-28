.PHONY: help run build test clean docker-build docker-run fmt vet lint

# Default target
help:
	@echo "Available targets:"
	@echo "  run           - Run the application"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  fmt           - Format Go code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run linters"

# Run the application
run:
	go run main.go

# Build the application
build:
	go build -o metadata-api main.go

# Run tests (when tests are added)
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f metadata-api
	go clean

# Build Docker image
docker-build:
	docker build -t metadata-api:latest .

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 --name metadata-api metadata-api:latest

# Format Go code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run all linters
lint: fmt vet
	@echo "Linting complete"

# Install dependencies
deps:
	go mod download
	go mod verify

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy

