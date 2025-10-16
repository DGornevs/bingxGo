# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Project parameters
PROJECT_NAME=bingxGo
MAIN_PATH=./cmd/app
BUILD_DIR=build
BINARY_NAME=$(PROJECT_NAME)
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: all build clean test deps fmt vet run help

# Default target
all: clean deps test build

# Build the application
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Linux
build-linux:
	@echo "Building $(PROJECT_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PATH)
	@echo "Linux build completed: $(BUILD_DIR)/$(BINARY_UNIX)"

# Build for Windows
build-windows:
	@echo "Building $(PROJECT_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) $(MAIN_PATH)
	@echo "Windows build completed: $(BUILD_DIR)/$(BINARY_WINDOWS)"

# Build for all platforms
build-all: build-linux build-windows build

# Run the application
run:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH) && ./$(BUILD_DIR)/$(BINARY_NAME)

# Run with live reload (requires 'air' to be installed)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running normally..."; \
		$(MAKE) run; \
	fi

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run examples
run-websocket-example:
	$(GOBUILD) -o $(BUILD_DIR)/websocket_example ./examples/websocket_example.go && ./$(BUILD_DIR)/websocket_example

run-binance-example:
	$(GOBUILD) -o $(BUILD_DIR)/binance_example ./examples/binance_pairs_example.go && ./$(BUILD_DIR)/binance_example

run-parser-example:
	$(GOBUILD) -o $(BUILD_DIR)/parser_example ./examples/parser_example.go && ./$(BUILD_DIR)/parser_example

run-telegram-example:
	$(GOBUILD) -o $(BUILD_DIR)/telegram_example ./examples/telegram_example.go && ./$(BUILD_DIR)/telegram_example

run-prices-example:
	$(GOBUILD) -o $(BUILD_DIR)/prices_example ./examples/fetch_prices_example.go && ./$(BUILD_DIR)/prices_example

run-wallet-example:
	$(GOBUILD) -o $(BUILD_DIR)/wallet_example ./examples/wallet_balance_example.go && ./$(BUILD_DIR)/wallet_example

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create .env template
env-template:
	@echo "Creating .env.example template..."
	@echo "# BingX API Configuration" > .env.example
	@echo "BINGX_API_KEY=your_api_key_here" >> .env.example
	@echo "BINGX_API_SECRET=your_api_secret_here" >> .env.example
	@echo "" >> .env.example
	@echo "# Telegram Bot Configuration" >> .env.example
	@echo "CHAT_BOT_TOKEN=your_bot_token_here" >> .env.example
	@echo "CHAT_ID=your_chat_id_here" >> .env.example
	@echo "" >> .env.example
	@echo "# Application Configuration" >> .env.example
	@echo "APP_NAME=BingX Go Application" >> .env.example
	@echo "MESSAGE=Hello from BingX Go!" >> .env.example
	@echo ".env.example created"

# Show help
help:
	@echo "Available commands:"
	@echo "  build              Build the application"
	@echo "  build-linux        Build for Linux"
	@echo "  build-windows      Build for Windows"
	@echo "  build-all          Build for all platforms"
	@echo "  run                Build and run the application"
	@echo "  dev                Run with live reload (requires air)"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Download and tidy dependencies"
	@echo "  fmt                Format code"
	@echo "  vet                Run go vet"
	@echo "  lint               Run golangci-lint"
	@echo "  install-tools      Install development tools"
	@echo "  env-template       Create .env.example template"
	@echo "  run-*-example      Run specific examples"
	@echo "  help               Show this help message"