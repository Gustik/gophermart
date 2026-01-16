.PHONY: build run test clean docker-up docker-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=gophermart
BINARY_PATH=cmd/gophermart

# Build the project
build:
	cd $(BINARY_PATH) && $(GOBUILD) -o $(BINARY_NAME) -v

# Run the project locally
run: build
	cd $(BINARY_PATH) && ./$(BINARY_NAME)

# Run with flags
run-flags:
	$(GOCMD) run $(BINARY_PATH)/main.go \
		-a "localhost:8080" \
		-d "postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" \
		-r "http://localhost:8081"

# Run tests
test:
	$(GOTEST) -v -cover ./...

# Test coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build files
clean:
	rm -f $(BINARY_PATH)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Docker compose up
docker-up:
	docker-compose up --build

# Docker compose down
docker-down:
	docker-compose down -v

# Database migration (placeholder)
migrate-up:
	@echo "Running migrations..."
	# TODO: add migration tool

migrate-down:
	@echo "Rolling back migrations..."
	# TODO: add migration tool

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Build and run locally"
	@echo "  run-flags     - Run with example flags"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-up     - Start with docker-compose"
	@echo "  docker-down   - Stop docker-compose"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"