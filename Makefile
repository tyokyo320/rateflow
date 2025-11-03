.PHONY: help docker-up docker-down docker-logs docker-restart build test clean swagger run dev stop

# Default target
.DEFAULT_GOAL := help

# Color definitions
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## Show help information
	@echo "$(CYAN)RateFlow Available Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ==================== Docker Commands ====================

docker-up: swagger ## 🚀 Start all services (PostgreSQL + Redis + API)
	@echo "$(CYAN)Starting Docker services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started!$(NC)"
	@echo "$(YELLOW)Access:$(NC)"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  Health:     http://localhost:8080/health"

docker-down: ## 🛑 Stop all services
	@echo "$(CYAN)Stopping Docker services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

docker-logs: ## 📋 View API logs
	docker-compose logs -f api

docker-logs-all: ## 📋 View all service logs
	docker-compose logs -f

docker-restart: ## 🔄 Restart all services
	@echo "$(CYAN)Restarting services...$(NC)"
	docker-compose restart
	@echo "$(GREEN)✓ Services restarted$(NC)"

docker-rebuild: ## 🔨 Rebuild and start
	@echo "$(CYAN)Rebuilding images...$(NC)"
	docker-compose build --no-cache
	docker-compose up -d
	@echo "$(GREEN)✓ Rebuild complete$(NC)"

docker-ps: ## 📊 View service status
	docker-compose ps

docker-clean: ## 🧹 Clean all data (⚠️ Deletes database data)
	@echo "$(YELLOW)⚠️  This will delete all data!$(NC)"
	@read -p "Continue? (y/N) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "$(GREEN)✓ Cleanup complete$(NC)"; \
	fi

# ==================== Local Development Commands ====================

run: ## 🏃 Run API service (local development)
	go run cmd/api/main.go

dev: ## 💻 Start development environment (database only, run API locally)
	@echo "$(CYAN)Starting database services...$(NC)"
	docker-compose up -d postgres redis
	@echo "$(GREEN)✓ Database services started$(NC)"
	@echo ""
	@echo "$(YELLOW)Now you can run the API:$(NC)"
	@echo "  make run"

build: ## 🔨 Build API binary
	@echo "$(CYAN)Building API...$(NC)"
	go build -o rateflow-api cmd/api/main.go
	@echo "$(GREEN)✓ Build complete: rateflow-api$(NC)"

build-worker: ## 🔨 Build Worker binary
	@echo "$(CYAN)Building Worker...$(NC)"
	go build -o rateflow-worker cmd/worker/main.go
	@echo "$(GREEN)✓ Build complete: rateflow-worker$(NC)"

# ==================== Frontend Commands ====================

web-install: ## 📦 Install frontend dependencies
	@echo "$(CYAN)Installing frontend dependencies...$(NC)"
	cd web && npm install
	@echo "$(GREEN)✓ Frontend dependencies installed$(NC)"

web-dev: ## 🎨 Start frontend development server
	@echo "$(CYAN)Starting frontend development server...$(NC)"
	cd web && npm run dev

web-build: ## 🏗️ Build frontend for production
	@echo "$(CYAN)Building frontend...$(NC)"
	cd web && npm run build
	@echo "$(GREEN)✓ Frontend build complete$(NC)"

web-preview: ## 👀 Preview frontend production build
	cd web && npm run preview

web-lint: ## ✨ Lint frontend code
	cd web && npm run lint

web-type-check: ## 🔍 Frontend type check
	cd web && npm run type-check

# ==================== Test Commands ====================

test: ## 🧪 Run all tests
	@echo "$(CYAN)Running tests...$(NC)"
	go test ./... -v

test-short: ## 🧪 Run quick tests
	go test ./... -short

test-cover: ## 📊 Run tests with coverage report
	@echo "$(CYAN)Generating coverage report...$(NC)"
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

# ==================== Code Quality ====================

fmt: ## 🎨 Format code
	@echo "$(CYAN)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Formatting complete$(NC)"

vet: ## 🔍 Static analysis
	@echo "$(CYAN)Running static analysis...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ Static analysis passed$(NC)"

lint: fmt vet ## ✨ Run all code checks
	@echo "$(GREEN)✓ Code checks complete$(NC)"

# ==================== Swagger Documentation ====================

swagger: ## 📚 Generate Swagger documentation
	@echo "$(CYAN)Generating Swagger documentation...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	elif [ -f ~/go/1.25.3/bin/swag ]; then \
		~/go/1.25.3/bin/swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	else \
		echo "$(YELLOW)⚠️  swag not installed, installing...$(NC)"; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	fi
	@echo "$(GREEN)✓ Swagger documentation generated$(NC)"

# ==================== Cleanup Commands ====================

clean: ## 🧹 Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	rm -f rateflow-api rateflow-worker
	rm -f coverage.out coverage.html
	rm -rf web/dist web/node_modules
	go clean -cache -testcache
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

# ==================== Dependency Management ====================

deps: ## 📦 Download dependencies
	@echo "$(CYAN)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

tidy: ## 📦 Tidy dependencies
	@echo "$(CYAN)Tidying dependencies...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

# ==================== Shortcuts ====================

start: docker-up ## 🚀 Start (alias for docker-up)

stop: docker-down ## 🛑 Stop (alias for docker-down)

restart: docker-restart ## 🔄 Restart (alias for docker-restart)

logs: docker-logs ## 📋 View logs (alias for docker-logs)

# ==================== Health Check ====================

health: ## 🏥 Health check
	@echo "$(CYAN)Checking service health...$(NC)"
	@echo ""
	@echo "1. Docker service status:"
	@docker-compose ps || echo "$(YELLOW)Docker services not started$(NC)"
	@echo ""
	@echo "2. API health check:"
	@curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "$(YELLOW)API not responding$(NC)"
	@echo ""

# ==================== Quick Start ====================

quickstart: docker-up web-install ## 🎯 Quick start (one-command startup)
	@echo ""
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)✓ Services started!$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(CYAN)📚 Access documentation:$(NC)"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  Health:     http://localhost:8080/health"
	@echo ""
	@echo "$(CYAN)🎨 Start frontend:$(NC)"
	@echo "  make web-dev"
	@echo "  然后Access: http://localhost:5173"
	@echo ""
	@echo "$(CYAN)🧪 Test API:$(NC)"
	@echo "  curl http://localhost:8080/health"
	@echo "  curl \"http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY\""
	@echo ""
	@echo "$(CYAN)📋 View logs:$(NC)"
	@echo "  make logs"
	@echo ""
	@echo "$(CYAN)🛑 Stop services:$(NC)"
	@echo "  make stop"
	@echo ""

# ==================== Full Check ====================

check: fmt vet test ## ✅ Full check (format + static analysis + tests)
	@echo "$(GREEN)✓ All checks passed!$(NC)"

# ==================== Development Workflow ====================

dev-full: deps build test swagger web-install ## 🎓 Full development workflow
	@echo "$(GREEN)✓ Development environment ready!$(NC)"

# ==================== Fullstack Development ====================

fullstack: dev ## 🚀 Start fullstack development environment
	@echo "$(CYAN)Starting backend...$(NC)"
	@make run &
	@sleep 3
	@echo "$(CYAN)Starting frontend...$(NC)"
	@make web-dev
