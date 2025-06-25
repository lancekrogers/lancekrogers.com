# Blockhead Consulting Website - Makefile

.PHONY: help all test test-go test-js test-verbose install dev build clean serve lint dashboard dev-bg stop status logs docker-setup docker-dev docker-test docker-clean docker-logs docker-status dev-services stop-dev start-mailhog stop-mailhog check-docker

# Colors for output
GREEN = \033[0;32m
RED = \033[0;31m
YELLOW = \033[1;33m
BLUE = \033[0;34m
NC = \033[0m # No Color
BOLD = \033[1m

# Default target
all: build test-dashboard ## Run full build and test pipeline with dashboard

help: ## Show this help message
	@echo "$(BOLD)Blockhead Consulting Website - Available Commands:$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Dashboard target
dashboard: ## Run build and test pipeline with status dashboard
	@echo "$(BOLD)╔══════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BOLD)║               🏗️  BLOCKHEAD CONSULTING BUILD             ║$(NC)"
	@echo "$(BOLD)╚══════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Build & Test Dashboard$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@$(MAKE) --no-print-directory _dashboard-build
	@$(MAKE) --no-print-directory _dashboard-test-go
	@$(MAKE) --no-print-directory _dashboard-test-js
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) --no-print-directory _dashboard-summary

_dashboard-build:
	@printf "🔨 Build Status:        "
	@mkdir -p bin
	@if go build -o bin/blockhead-server main.go 2>/dev/null; then \
		echo "$(GREEN)✅ PASS$(NC)"; \
		echo "   └─ Binary: bin/blockhead-server"; \
	else \
		echo "$(RED)❌ FAIL$(NC)"; \
		echo "   └─ Go build failed"; \
	fi

_dashboard-test-go:
	@printf "🧪 Go Tests:           "
	@if go test ./... >/dev/null 2>&1; then \
		TESTS=$$(go test ./... -v 2>/dev/null | grep -E "^=== RUN|^--- PASS|^--- FAIL" | wc -l | tr -d ' '); \
		PASSED=$$(go test ./... -v 2>/dev/null | grep "^--- PASS" | wc -l | tr -d ' '); \
		echo "$(GREEN)✅ PASS$(NC) ($$PASSED tests)"; \
	else \
		echo "$(RED)❌ FAIL$(NC)"; \
		echo "   └─ Run 'make test-verbose' for details"; \
	fi

_dashboard-test-js:
	@printf "🌐 JS Tests:           "
	@if [ -d "node_modules" ]; then \
		if npm test >/dev/null 2>&1; then \
			echo "$(GREEN)✅ PASS$(NC)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
			echo "   └─ Run 'make test-verbose' for details"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC)"; \
		echo "   └─ Dependencies not installed (run 'make install')"; \
	fi

_dashboard-summary:
	@mkdir -p bin
	@BUILD_STATUS=$$(if go build -o bin/blockhead-server main.go 2>/dev/null; then echo "PASS"; else echo "FAIL"; fi); \
	TEST_GO_STATUS=$$(if go test ./... >/dev/null 2>&1; then echo "PASS"; else echo "FAIL"; fi); \
	TEST_JS_STATUS=$$(if [ -d "node_modules" ] && npm test >/dev/null 2>&1; then echo "PASS"; elif [ ! -d "node_modules" ]; then echo "SKIP"; else echo "FAIL"; fi); \
	echo "📋 Summary:"; \
	if [ "$$BUILD_STATUS" = "PASS" ] && [ "$$TEST_GO_STATUS" = "PASS" ] && ([ "$$TEST_JS_STATUS" = "PASS" ] || [ "$$TEST_JS_STATUS" = "SKIP" ]); then \
		echo "   $(GREEN)🎉 All systems operational! Ready for deployment.$(NC)"; \
	else \
		echo "   $(RED)⚠️  Issues detected. Review failed components above.$(NC)"; \
	fi; \
	echo ""; \
	echo "💡 Next steps:"; \
	if [ "$$BUILD_STATUS" = "FAIL" ]; then echo "   • Fix build errors with 'go build'"; fi; \
	if [ "$$TEST_GO_STATUS" = "FAIL" ]; then echo "   • Debug Go tests with 'make test-verbose'"; fi; \
	if [ "$$TEST_JS_STATUS" = "FAIL" ]; then echo "   • Debug JS tests with 'make test-verbose'"; fi; \
	if [ "$$TEST_JS_STATUS" = "SKIP" ]; then echo "   • Install JS dependencies with 'make install'"; fi; \
	if [ "$$BUILD_STATUS" = "PASS" ] && [ "$$TEST_GO_STATUS" = "PASS" ] && ([ "$$TEST_JS_STATUS" = "PASS" ] || [ "$$TEST_JS_STATUS" = "SKIP" ]); then \
		echo "   • Start development server with 'make dev'"; \
		echo "   • Deploy with './bin/blockhead-server'"; \
	fi

# Testing targets
test: test-go test-js ## Run all tests (Go + JavaScript)

test-quick: ## Run all tests with clean dashboard output (no verbose)
	@$(MAKE) --no-print-directory test-dashboard

test-verbose: ## Run all tests with verbose output
	@echo "$(BOLD)Running verbose tests...$(NC)"
	@echo ""
	@echo "$(YELLOW)Go Tests:$(NC)"
	@go test -v ./...
	@echo ""
	@echo "$(YELLOW)JavaScript Tests:$(NC)"
	@if [ -d "node_modules" ]; then \
		npm run test:js; \
	else \
		echo "JavaScript dependencies not installed. Run 'make install' first."; \
	fi

test-go: ## Run Go backend tests
	@echo "Running Go tests..."
	@go test -v ./...

test-js: ## Run JavaScript tests (requires npm install first)
	@echo "Running JavaScript tests..."
	@if [ -d "node_modules" ]; then \
		npm run test:js; \
	else \
		echo "JavaScript dependencies not installed. Run 'make install' first."; \
	fi

test-dashboard: ## Run all tests with comprehensive dashboard output
	@echo "$(BOLD)╔══════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BOLD)║              🧪 BLOCKHEAD CONSULTING TESTS               ║$(NC)"
	@echo "$(BOLD)╚══════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Test Results Dashboard$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@$(MAKE) --no-print-directory _test-dashboard-go
	@$(MAKE) --no-print-directory _test-dashboard-js-unit
	@$(MAKE) --no-print-directory _test-dashboard-js-navigation
	@$(MAKE) --no-print-directory _test-dashboard-js-mobile
	@$(MAKE) --no-print-directory _test-dashboard-js-animations
	@$(MAKE) --no-print-directory _test-dashboard-js-integration
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) --no-print-directory _test-dashboard-summary

_test-dashboard-go:
	@printf "🔹 Go Backend Tests:    "
	@if go test ./... >/dev/null 2>&1; then \
		PACKAGES=$$(go list ./... | wc -l | tr -d ' '); \
		TESTS=$$(go test ./... -v 2>/dev/null | grep -E "^--- PASS" | wc -l | tr -d ' '); \
		echo "$(GREEN)✅ PASS$(NC) ($$TESTS tests in $$PACKAGES packages)"; \
	else \
		FAILED=$$(go test ./... 2>&1 | grep -E "^--- FAIL" | wc -l | tr -d ' '); \
		echo "$(RED)❌ FAIL$(NC) ($$FAILED tests failed)"; \
		echo "   └─ Run 'make test-go' for details"; \
	fi

_test-dashboard-js-unit:
	@printf "🔹 JS Test Suite:       "
	@if [ -d "node_modules" ]; then \
		if npm run test:js -- --silent >/dev/null 2>&1; then \
			TESTS=$$(npm run test:js -- --silent 2>&1 | grep -E "Tests:.*passed" | sed -E 's/.*Tests:[[:space:]]+([0-9]+).*/\1/'); \
			echo "$(GREEN)✅ PASS$(NC) ($$TESTS total tests)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC) (dependencies not installed)"; \
	fi

_test-dashboard-js-navigation:
	@printf "🔹 Navigation Tests:    "
	@if [ -d "node_modules" ] && [ -f "tests/js/navigation.test.js" ]; then \
		if npm run test:navigation -- --silent >/dev/null 2>&1; then \
			TESTS=$$(npm run test:navigation -- --silent 2>&1 | grep -E "✓" | wc -l | tr -d ' '); \
			echo "$(GREEN)✅ PASS$(NC) ($$TESTS tests)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC)"; \
	fi

_test-dashboard-js-mobile:
	@printf "🔹 Mobile Tests:        "
	@if [ -d "node_modules" ] && [ -f "tests/js/mobile.test.js" ]; then \
		if npm run test:mobile -- --silent >/dev/null 2>&1; then \
			TESTS=$$(npm run test:mobile -- --silent 2>&1 | grep -E "✓" | wc -l | tr -d ' '); \
			echo "$(GREEN)✅ PASS$(NC) ($$TESTS tests)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC)"; \
	fi

_test-dashboard-js-animations:
	@printf "🔹 Animation Tests:     "
	@if [ -d "node_modules" ] && [ -f "tests/js/animations.test.js" ]; then \
		if npm run test:animations -- --silent >/dev/null 2>&1; then \
			TESTS=$$(npm run test:animations -- --silent 2>&1 | grep -E "✓" | wc -l | tr -d ' '); \
			echo "$(GREEN)✅ PASS$(NC) ($$TESTS tests)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC)"; \
	fi

_test-dashboard-js-integration:
	@printf "🔹 Integration Tests:   "
	@if [ -d "node_modules" ] && [ -f "tests/js/integration.test.js" ]; then \
		if npm run test:integration -- --silent >/dev/null 2>&1; then \
			TESTS=$$(npm run test:integration -- --silent 2>&1 | grep -E "✓" | wc -l | tr -d ' '); \
			echo "$(GREEN)✅ PASS$(NC) ($$TESTS tests)"; \
		else \
			echo "$(RED)❌ FAIL$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  SKIP$(NC)"; \
	fi

_test-dashboard-summary:
	@echo "📋 Test Summary:"
	@ALL_PASS=true; \
	if ! go test ./... >/dev/null 2>&1; then ALL_PASS=false; fi; \
	if [ -d "node_modules" ]; then \
		if ! npm run test:js -- --silent >/dev/null 2>&1; then ALL_PASS=false; fi; \
	fi; \
	if [ "$$ALL_PASS" = "true" ]; then \
		echo "   $(GREEN)🎉 All tests passing! Code is ready for deployment.$(NC)"; \
	else \
		echo "   $(RED)⚠️  Some tests are failing. Review details above.$(NC)"; \
		echo ""; \
		echo "💡 Debug commands:"; \
		echo "   • Go tests: make test-go"; \
		echo "   • All JS tests: make test-js"; \
		echo "   • Navigation: npm run test:navigation"; \
		echo "   • Mobile: npm run test:mobile"; \
		echo "   • Animations: npm run test:animations"; \
		echo "   • Integration: npm run test:integration"; \
	fi

test-watch: ## Run JavaScript tests in watch mode
	@if [ -d "node_modules" ]; then \
		npm run test:js:watch; \
	else \
		echo "$(RED)❌ JavaScript dependencies not installed. Run 'make install' first.$(NC)"; \
	fi

test-coverage: ## Run tests with coverage report
	@echo "$(BOLD)📊 Generating test coverage reports...$(NC)"
	@echo ""
	@echo "$(YELLOW)Go Coverage:$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage-go.html
	@echo "   └─ Report: coverage-go.html"
	@echo ""
	@echo "$(YELLOW)JavaScript Coverage:$(NC)"
	@if [ -d "node_modules" ]; then \
		npm run test:js:coverage; \
		echo "   └─ Report: coverage/lcov-report/index.html"; \
	else \
		echo "   └─ $(RED)Skipped (dependencies not installed)$(NC)"; \
	fi

test-e2e: ## Run end-to-end tests (requires server running)
	@echo "$(BOLD)🌐 Running end-to-end tests...$(NC)"
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	if ! lsof -ti :$$PORT >/dev/null 2>&1; then \
		echo "$(RED)❌ Server not running on port $$PORT$(NC)"; \
		echo "   └─ Start server first: make dev-bg"; \
		exit 1; \
	fi
	@if [ -d "node_modules" ]; then \
		npm run test:e2e; \
	else \
		echo "$(RED)❌ JavaScript dependencies not installed. Run 'make install' first.$(NC)"; \
	fi

# Development targets
install: ## Install JavaScript dependencies
	@echo "Installing JavaScript dependencies..."
	@npm install --silent 2>/dev/null || npm install

install-clean: ## Clean install JavaScript dependencies (removes warnings)
	@echo "Cleaning and installing JavaScript dependencies..."
	@rm -rf node_modules package-lock.json
	@npm install --silent 2>/dev/null || npm install

# Basic dev command (overridden below with enhanced version)
dev-simple: ## Start development server only (no services)
	@echo "Starting development server on port from .env file..."
	@go run main.go

serve: dev ## Alias for dev

# Build targets
build: ## Build the Go binary
	@echo "Building Go binary..."
	@mkdir -p bin
	@go build -o bin/blockhead-server main.go
	@chmod +x bin/blockhead-server
	@echo "Binary built: bin/blockhead-server"

run: build ## Build and run the server binary
	@echo "Starting server from binary..."
	@./bin/blockhead-server

build-css: ## Compile SCSS to CSS
	@echo "Compiling SCSS to CSS..."
	@mkdir -p static/css
	@if command -v sass >/dev/null 2>&1; then \
		sass static/scss/styles.scss static/css/styles.css --no-source-map --style=compressed; \
		echo "$(GREEN)✅ SCSS compiled to static/css/styles.css$(NC)"; \
	else \
		echo "$(RED)❌ sass command not found. Install with: npm install -g sass$(NC)"; \
		exit 1; \
	fi

# Utility targets
clean: ## Clean build artifacts and dependencies
	@echo "Cleaning build artifacts..."
	@rm -rf bin/         # Remove binary directory
	@rm -f server        # Remove legacy binary if it exists
	@rm -rf node_modules # Remove JS dependencies
	@rm -f data/server.log # Remove server logs
	@echo "Cleanup complete"

lint: ## Run linters (requires golangci-lint)
	@echo "Running Go linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# CI/Production targets
ci-test: ## Run tests in CI environment
	@echo "Running CI tests..."
	@go test -v -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Message Management Commands
generate-key: ## Generate a new encryption key for messages
	@echo "$(BOLD)🔐 Generating new encryption key...$(NC)"
	@go run cmd/generate-key/main.go

messages: ## List all contact messages
	@echo "$(BOLD)📬 Contact Messages:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@go run cmd/decrypt-messages/main.go || echo "$(RED)❌ No messages found or decryption failed$(NC)"

messages-new: ## List only new messages
	@echo "$(BOLD)📬 New Messages:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@go run cmd/decrypt-messages/main.go -status new || echo "$(YELLOW)⚠️  No new messages$(NC)"

message-read: ## Read a specific message (ID=msg_xxx)
	@if [ -z "$(ID)" ]; then \
		echo "$(RED)❌ Error: ID is required$(NC)"; \
		echo "   └─ Usage: make message-read ID=msg_xxx"; \
		exit 1; \
	fi
	@echo "$(BOLD)📧 Message Details:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@go run cmd/decrypt-messages/main.go -id $(ID)

message-status: ## Update message status (ID=msg_xxx STATUS=read|replied|closed)
	@if [ -z "$(ID)" ] || [ -z "$(STATUS)" ]; then \
		echo "$(RED)❌ Error: ID and STATUS are required$(NC)"; \
		echo "   └─ Usage: make message-status ID=msg_xxx STATUS=read"; \
		echo "   └─ Valid statuses: new, read, replied, closed"; \
		exit 1; \
	fi
	@echo "$(BOLD)📝 Updating message status...$(NC)"
	@go run cmd/message-status/main.go -id $(ID) -status $(STATUS)
	@echo "$(GREEN)✅ Message $(ID) marked as $(STATUS)$(NC)"

# Port management
kill-port: ## Kill any process using the configured port
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	echo "Killing processes on port $$PORT..."; \
	lsof -ti :$$PORT | xargs kill -9 2>/dev/null || echo "No processes found on port $$PORT"; \
	sleep 1

restart: kill-port dev ## Kill existing server and restart

# Legacy server management (replaced by enhanced version below)
dev-bg-simple: ## Start development server in background (no services)
	@echo "Starting development server in background..."
	@$(MAKE) --no-print-directory kill-port
	@mkdir -p data
	@nohup go run main.go > data/server.log 2>&1 &
	@echo "Waiting for server to start (loading templates and blog posts)..."
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
		sleep 1; \
		if lsof -ti :$$PORT >/dev/null 2>&1; then \
			echo ""; \
			echo "$(GREEN)✅ Server started successfully on port $$PORT$(NC)"; \
			echo "   └─ Logs: tail -f data/server.log"; \
			echo "   └─ Stop: make stop"; \
			echo "   └─ Status: make status"; \
			exit 0; \
		fi; \
		if [ $$i -eq 5 ]; then echo " (compiling Go code...)"; fi; \
		if [ $$i -eq 10 ]; then echo " (loading templates and blog posts...)"; fi; \
		printf "."; \
	done; \
	echo ""; \
	echo "$(RED)❌ Server failed to start within 15 seconds$(NC)"; \
	echo "   └─ Check logs: cat data/server.log"; \
	echo "   └─ Try manual start: make dev"

stop: ## Stop the development server
	@echo "Stopping development server..."
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	if lsof -ti :$$PORT >/dev/null 2>&1; then \
		PID=$$(lsof -ti :$$PORT); \
		echo "Sending graceful shutdown signal to process $$PID..."; \
		kill -TERM $$PID 2>/dev/null || true; \
		sleep 3; \
		if lsof -ti :$$PORT >/dev/null 2>&1; then \
			echo "$(YELLOW)⚠️  Graceful shutdown failed, force killing...$(NC)"; \
			$(MAKE) --no-print-directory kill-port; \
		else \
			echo "$(GREEN)✅ Server stopped gracefully$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  No server running on port $$PORT$(NC)"; \
	fi
	@echo "Cleaning up any orphaned Go processes..."
	@pkill -f "go run main.go" 2>/dev/null || true
	@pkill -f "main.go" 2>/dev/null || true

status: ## Check server status
	@echo "Server Status Check:"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━"
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	if lsof -ti :$$PORT >/dev/null 2>&1; then \
		PID=$$(lsof -ti :$$PORT); \
		BINARY=$$(ps -p $$PID -o comm= 2>/dev/null || echo "unknown"); \
		UPTIME=$$(ps -p $$PID -o etime= 2>/dev/null | tr -d ' ' || echo "unknown"); \
		echo "$(GREEN)🟢 Server RUNNING$(NC)"; \
		echo "   └─ PID: $$PID"; \
		echo "   └─ Binary: $$BINARY"; \
		echo "   └─ Uptime: $$UPTIME"; \
		echo "   └─ URL: http://localhost:$$PORT"; \
		if [ -f "data/server.log" ]; then \
			echo "   └─ Logs: tail -f data/server.log"; \
		fi \
	else \
		echo "$(RED)🔴 Server STOPPED$(NC)"; \
		echo "   └─ Start with: make dev or make dev-bg"; \
	fi

logs: ## Show server logs (if running in background)
	@if [ -f "data/server.log" ]; then \
		echo "$(BLUE)📋 Server Logs:$(NC)"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━"; \
		tail -f data/server.log; \
	else \
		echo "$(YELLOW)⚠️  No log file found (data/server.log)$(NC)"; \
		echo "   └─ Start server with 'make dev-bg' to generate logs"; \
	fi

# Docker targets
docker-setup: ## Set up Docker development environment
	@echo "$(BOLD)🐳 Setting up Docker development environment...$(NC)"
	@./scripts/dev-setup.sh

docker-dev: ## Start application in Docker development mode
	@echo "$(BLUE)🐳 Starting Docker development environment...$(NC)"
	@docker-compose up --build

docker-dev-bg: ## Start Docker development environment in background
	@echo "$(BLUE)🐳 Starting Docker development environment in background...$(NC)"
	@docker-compose up -d --build
	@echo "$(GREEN)✅ Docker services started$(NC)"
	@echo "   └─ App: http://localhost:8085"
	@echo "   └─ MailHog: http://localhost:8025"
	@echo "   └─ Gitea: http://localhost:3000"
	@echo "   └─ Logs: make docker-logs"
	@echo "   └─ Stop: make docker-stop"

docker-test: ## Run integration tests in Docker
	@echo "$(BLUE)🧪 Running Docker integration tests...$(NC)"
	@./scripts/integration-test.sh

docker-clean: ## Clean up Docker containers and volumes
	@echo "$(YELLOW)🧹 Cleaning up Docker environment...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker-compose -f docker-compose.test.yml down -v --remove-orphans
	@docker system prune -f
	@echo "$(GREEN)✅ Docker cleanup complete$(NC)"

docker-logs: ## Show Docker container logs
	@echo "$(BLUE)📋 Docker Container Logs:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@docker-compose logs -f

docker-status: ## Check Docker container status
	@echo "$(BLUE)🐳 Docker Container Status:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@docker-compose ps

docker-stop: ## Stop Docker development environment
	@echo "$(YELLOW)🛑 Stopping Docker development environment...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✅ Docker services stopped$(NC)"

docker-restart: docker-stop docker-dev-bg ## Restart Docker development environment

# Enhanced Development Service Management
check-docker: ## Check if Docker is running
	@if ! docker info >/dev/null 2>&1; then \
		echo "$(RED)❌ Docker daemon is not running$(NC)"; \
		echo "   └─ Please start Docker Desktop or run: open -a Docker"; \
		echo "   └─ On Linux: sudo systemctl start docker"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Docker daemon is running$(NC)"; \
	fi

start-mailhog: check-docker ## Start MailHog email testing service
	@echo "$(BLUE)📧 Starting MailHog email testing service...$(NC)"
	@if docker compose ps mailhog 2>/dev/null | grep -q "Up"; then \
		echo "$(YELLOW)⚠️  MailHog is already running$(NC)"; \
	else \
		docker compose up mailhog -d; \
		echo "$(GREEN)✅ MailHog started$(NC)"; \
		echo "   └─ SMTP: localhost:1025"; \
		echo "   └─ Web UI: http://localhost:8025"; \
	fi

stop-mailhog: ## Stop MailHog email testing service
	@echo "$(YELLOW)📧 Stopping MailHog email testing service...$(NC)"
	@docker compose stop mailhog 2>/dev/null || true
	@docker compose rm -f mailhog 2>/dev/null || true
	@echo "$(GREEN)✅ MailHog stopped$(NC)"

dev-services: start-mailhog ## Start all development services (MailHog, etc.)
	@echo "$(GREEN)🚀 All development services started!$(NC)"
	@echo ""
	@echo "$(BOLD)📋 Development Services Status:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📧 MailHog Email Testing:"
	@echo "   └─ SMTP: localhost:1025"
	@echo "   └─ Web UI: http://localhost:8025"
	@echo ""
	@echo "💡 Next steps:"
	@echo "   └─ Start website: make dev-bg"
	@echo "   └─ Stop all services: make stop-dev"

stop-dev: stop stop-mailhog ## Stop all development services (website + MailHog)
	@echo ""
	@echo "$(GREEN)🛑 All development services stopped$(NC)"

# Enhanced dev and dev-bg commands
dev: dev-services build-css ## Start development server with all services
	@echo ""
	@echo "$(BOLD)🚀 Starting complete development environment...$(NC)"
	@echo ""
	@go run main.go

dev-bg: dev-services build-css ## Start development server in background with all services
	@echo ""
	@echo "$(BOLD)🚀 Starting complete development environment in background...$(NC)"
	@echo ""
	@$(MAKE) --no-print-directory kill-port
	@mkdir -p data
	@nohup go run main.go > data/server.log 2>&1 &
	@echo "Waiting for server to start (loading templates and blog posts)..."
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
		sleep 1; \
		if lsof -ti :$$PORT >/dev/null 2>&1; then \
			echo ""; \
			echo "$(GREEN)✅ Complete development environment ready!$(NC)"; \
			echo ""; \
			echo "$(BOLD)📋 Development Environment Status:$(NC)"; \
			echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
			echo "🌐 Website: http://localhost:$$PORT"; \
			echo "📧 MailHog UI: http://localhost:8025"; \
			echo "📁 Server Logs: tail -f data/server.log"; \
			echo ""; \
			echo "💡 Commands:"; \
			echo "   └─ Stop all: make stop-dev"; \
			echo "   └─ Status: make status"; \
			echo "   └─ Logs: make logs"; \
			exit 0; \
		fi; \
		if [ $$i -eq 5 ]; then echo " (compiling Go code...)"; fi; \
		if [ $$i -eq 10 ]; then echo " (loading templates and blog posts...)"; fi; \
		printf "."; \
	done; \
	echo ""; \
	echo "$(RED)❌ Server failed to start within 15 seconds$(NC)"; \
	echo "   └─ Check logs: cat data/server.log"; \
	echo "   └─ Try manual start: make dev"

# Production server management
prod: ## Start production server in foreground
	@echo "$(BOLD)🚀 Starting production server...$(NC)"
	@echo "   └─ Environment: production"
	@echo "   └─ Port: $$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087")"
	@echo ""
	@ENVIRONMENT=production ./bin/blockhead-server

prod-bg: build ## Start production server in background
	@echo "$(BOLD)🚀 Starting production server in background...$(NC)"
	@$(MAKE) --no-print-directory kill-port
	@mkdir -p data
	@nohup env ENVIRONMENT=production ./bin/blockhead-server > data/production.log 2>&1 &
	@echo "Waiting for production server to start..."
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
		sleep 1; \
		if lsof -ti :$$PORT >/dev/null 2>&1; then \
			echo ""; \
			echo "$(GREEN)✅ Production server started successfully\!$(NC)"; \
			echo ""; \
			echo "$(BOLD)📋 Production Server Status:$(NC)"; \
			echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
			echo "🌐 Server: http://localhost:$$PORT"; \
			echo "📁 Logs: tail -f data/production.log"; \
			echo "🔧 Environment: production"; \
			echo ""; \
			echo "💡 Commands:"; \
			echo "   └─ Stop: make prod-stop"; \
			echo "   └─ Status: make prod-status"; \
			echo "   └─ Logs: make prod-logs"; \
			echo "   └─ Restart: make prod-restart"; \
			exit 0; \
		fi; \
		printf "."; \
	done; \
	echo ""; \
	echo "$(RED)❌ Production server failed to start within 10 seconds$(NC)"; \
	echo "   └─ Check logs: cat data/production.log"; \
	echo "   └─ Try manual start: make prod"

prod-stop: ## Stop production server
	@echo "Stopping production server..."
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	if lsof -ti :$$PORT >/dev/null 2>&1; then \
		PID=$$(lsof -ti :$$PORT); \
		echo "Sending graceful shutdown signal to process $$PID..."; \
		kill -TERM $$PID 2>/dev/null || true; \
		sleep 3; \
		if lsof -ti :$$PORT >/dev/null 2>&1; then \
			echo "$(YELLOW)⚠️  Graceful shutdown failed, force killing...$(NC)"; \
			$(MAKE) --no-print-directory kill-port; \
		else \
			echo "$(GREEN)✅ Production server stopped gracefully$(NC)"; \
		fi \
	else \
		echo "$(YELLOW)⚠️  No server running on port $$PORT$(NC)"; \
	fi

prod-status: ## Check production server status
	@echo "Production Server Status:"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@PORT=$$(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8087"); \
	if lsof -ti :$$PORT >/dev/null 2>&1; then \
		PID=$$(lsof -ti :$$PORT); \
		BINARY=$$(ps -p $$PID -o comm= 2>/dev/null || echo "unknown"); \
		UPTIME=$$(ps -p $$PID -o etime= 2>/dev/null | tr -d ' ' || echo "unknown"); \
		echo "$(GREEN)🟢 Production Server RUNNING$(NC)"; \
		echo "   └─ PID: $$PID"; \
		echo "   └─ Binary: $$BINARY"; \
		echo "   └─ Uptime: $$UPTIME"; \
		echo "   └─ URL: http://localhost:$$PORT"; \
		echo "   └─ Environment: production"; \
		if [ -f "data/production.log" ]; then \
			echo "   └─ Logs: tail -f data/production.log"; \
		fi \
	else \
		echo "$(RED)🔴 Production Server STOPPED$(NC)"; \
		echo "   └─ Start with: make prod-bg"; \
	fi

prod-logs: ## Show production server logs
	@if [ -f "data/production.log" ]; then \
		echo "$(BLUE)📋 Production Server Logs:$(NC)"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		tail -f data/production.log; \
	else \
		echo "$(YELLOW)⚠️  No production log file found (data/production.log)$(NC)"; \
		echo "   └─ Start server with 'make prod-bg' to generate logs"; \
	fi

prod-restart: prod-stop prod-bg ## Restart production server

prod-deploy: ## Deploy latest code (git pull + rebuild + restart)
	@echo "$(BOLD)🚀 Deploying latest code...$(NC)"
	@echo ""
	@echo "1. Pulling latest code from git..."
	@git pull origin main
	@echo ""
	@echo "2. Building application..."
	@$(MAKE) --no-print-directory build
	@echo ""
	@echo "3. Restarting production server..."
	@$(MAKE) --no-print-directory prod-restart
	@echo ""
	@echo "$(GREEN)✅ Deployment complete!$(NC)"

prod-install-service: build ## Install systemd service for auto-restart on boot
	@echo "$(BOLD)🔧 Installing systemd service...$(NC)"
	@echo "   └─ Stopping any running production server..."
	@$(MAKE) --no-print-directory prod-stop 2>/dev/null || true
	@echo "   └─ Copying service file..."
	@sudo cp blockhead.service /etc/systemd/system/
	@echo "   └─ Reloading systemd..."
	@sudo systemctl daemon-reload
	@echo "   └─ Enabling service for auto-start..."
	@sudo systemctl enable blockhead
	@echo "   └─ Starting service..."
	@sudo systemctl start blockhead
	@echo ""
	@echo "$(GREEN)✅ Systemd service installed and started!$(NC)"
	@echo ""
	@echo "$(BOLD)📋 Service Management Commands:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔍 Status:    sudo systemctl status blockhead"
	@echo "📋 Logs:      sudo journalctl -u blockhead -f"
	@echo "🔄 Restart:   sudo systemctl restart blockhead"
	@echo "🛑 Stop:      sudo systemctl stop blockhead"
	@echo "❌ Disable:   sudo systemctl disable blockhead"

prod-service-status: ## Check systemd service status
	@echo "$(BOLD)📊 Systemd Service Status:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@sudo systemctl status blockhead --no-pager || true

prod-service-logs: ## Show systemd service logs
	@echo "$(BOLD)📋 Systemd Service Logs:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@sudo journalctl -u blockhead -f

prod-service-restart: ## Restart systemd service
	@echo "$(BOLD)🔄 Restarting systemd service...$(NC)"
	@sudo systemctl restart blockhead
	@echo "$(GREEN)✅ Service restarted!$(NC)"

prod-service-remove: ## Remove systemd service
	@echo "$(BOLD)❌ Removing systemd service...$(NC)"
	@sudo systemctl stop blockhead 2>/dev/null || true
	@sudo systemctl disable blockhead 2>/dev/null || true
	@sudo rm -f /etc/systemd/system/blockhead.service
	@sudo systemctl daemon-reload
	@echo "$(GREEN)✅ Systemd service removed!$(NC)"