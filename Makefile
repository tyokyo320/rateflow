.PHONY: help docker-up docker-down docker-logs docker-restart build test clean swagger run dev stop

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## 显示帮助信息
	@echo "$(CYAN)RateFlow 可用命令:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ==================== Docker 命令 ====================

docker-up: swagger ## 🚀 启动所有服务（PostgreSQL + Redis + API）
	@echo "$(CYAN)启动 Docker 服务...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ 服务已启动！$(NC)"
	@echo "$(YELLOW)访问:$(NC)"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  Health:     http://localhost:8080/health"

docker-down: ## 🛑 停止所有服务
	@echo "$(CYAN)停止 Docker 服务...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ 服务已停止$(NC)"

docker-logs: ## 📋 查看 API 日志
	docker-compose logs -f api

docker-logs-all: ## 📋 查看所有服务日志
	docker-compose logs -f

docker-restart: ## 🔄 重启所有服务
	@echo "$(CYAN)重启服务...$(NC)"
	docker-compose restart
	@echo "$(GREEN)✓ 服务已重启$(NC)"

docker-rebuild: ## 🔨 重新构建并启动
	@echo "$(CYAN)重新构建镜像...$(NC)"
	docker-compose build --no-cache
	docker-compose up -d
	@echo "$(GREEN)✓ 重新构建完成$(NC)"

docker-ps: ## 📊 查看服务状态
	docker-compose ps

docker-clean: ## 🧹 清理所有数据（⚠️ 会删除数据库数据）
	@echo "$(YELLOW)⚠️  这将删除所有数据！$(NC)"
	@read -p "确定继续？(y/N) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "$(GREEN)✓ 清理完成$(NC)"; \
	fi

# ==================== 本地开发命令 ====================

run: ## 🏃 运行 API 服务（本地开发）
	go run cmd/api/main.go

dev: ## 💻 启动开发环境（仅数据库，API 本地运行）
	@echo "$(CYAN)启动数据库服务...$(NC)"
	docker-compose up -d postgres redis
	@echo "$(GREEN)✓ 数据库服务已启动$(NC)"
	@echo ""
	@echo "$(YELLOW)现在可以运行 API:$(NC)"
	@echo "  make run"

build: ## 🔨 构建 API 二进制文件
	@echo "$(CYAN)构建 API...$(NC)"
	go build -o rateflow-api cmd/api/main.go
	@echo "$(GREEN)✓ 构建完成: rateflow-api$(NC)"

build-worker: ## 🔨 构建 Worker 二进制文件
	@echo "$(CYAN)构建 Worker...$(NC)"
	go build -o rateflow-worker cmd/worker/main.go
	@echo "$(GREEN)✓ 构建完成: rateflow-worker$(NC)"

# ==================== 前端命令 ====================

web-install: ## 📦 安装前端依赖
	@echo "$(CYAN)安装前端依赖...$(NC)"
	cd web && npm install
	@echo "$(GREEN)✓ 前端依赖安装完成$(NC)"

web-dev: ## 🎨 启动前端开发服务器
	@echo "$(CYAN)启动前端开发服务器...$(NC)"
	cd web && npm run dev

web-build: ## 🏗️ 构建前端生产版本
	@echo "$(CYAN)构建前端...$(NC)"
	cd web && npm run build
	@echo "$(GREEN)✓ 前端构建完成$(NC)"

web-preview: ## 👀 预览前端生产构建
	cd web && npm run preview

web-lint: ## ✨ 检查前端代码
	cd web && npm run lint

web-type-check: ## 🔍 前端类型检查
	cd web && npm run type-check

# ==================== 测试命令 ====================

test: ## 🧪 运行所有测试
	@echo "$(CYAN)运行测试...$(NC)"
	go test ./... -v

test-short: ## 🧪 运行快速测试
	go test ./... -short

test-cover: ## 📊 运行测试并生成覆盖率报告
	@echo "$(CYAN)生成覆盖率报告...$(NC)"
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ 覆盖率报告: coverage.html$(NC)"

# ==================== 代码质量 ====================

fmt: ## 🎨 格式化代码
	@echo "$(CYAN)格式化代码...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ 格式化完成$(NC)"

vet: ## 🔍 静态分析
	@echo "$(CYAN)运行静态分析...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ 静态分析通过$(NC)"

lint: fmt vet ## ✨ 运行所有代码检查
	@echo "$(GREEN)✓ 代码检查完成$(NC)"

# ==================== Swagger 文档 ====================

swagger: ## 📚 生成 Swagger 文档
	@echo "$(CYAN)生成 Swagger 文档...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	elif [ -f ~/go/1.25.3/bin/swag ]; then \
		~/go/1.25.3/bin/swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	else \
		echo "$(YELLOW)⚠️  swag 未安装，正在安装...$(NC)"; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal; \
	fi
	@echo "$(GREEN)✓ Swagger 文档已生成$(NC)"

# ==================== 清理命令 ====================

clean: ## 🧹 清理构建产物
	@echo "$(CYAN)清理构建产物...$(NC)"
	rm -f rateflow-api rateflow-worker
	rm -f coverage.out coverage.html
	rm -rf web/dist web/node_modules
	go clean -cache -testcache
	@echo "$(GREEN)✓ 清理完成$(NC)"

# ==================== 依赖管理 ====================

deps: ## 📦 下载依赖
	@echo "$(CYAN)下载依赖...$(NC)"
	go mod download
	@echo "$(GREEN)✓ 依赖下载完成$(NC)"

tidy: ## 📦 整理依赖
	@echo "$(CYAN)整理依赖...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ 依赖整理完成$(NC)"

# ==================== 快捷命令 ====================

start: docker-up ## 🚀 启动（docker-up 的别名）

stop: docker-down ## 🛑 停止（docker-down 的别名）

restart: docker-restart ## 🔄 重启（docker-restart 的别名）

logs: docker-logs ## 📋 查看日志（docker-logs 的别名）

# ==================== 健康检查 ====================

health: ## 🏥 健康检查
	@echo "$(CYAN)检查服务健康状态...$(NC)"
	@echo ""
	@echo "1. Docker 服务状态:"
	@docker-compose ps || echo "$(YELLOW)Docker 服务未启动$(NC)"
	@echo ""
	@echo "2. API 健康检查:"
	@curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "$(YELLOW)API 未响应$(NC)"
	@echo ""

# ==================== 快速开始 ====================

quickstart: docker-up web-install ## 🎯 快速开始（一键启动所有服务）
	@echo ""
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo "$(GREEN)✓ 服务已启动！$(NC)"
	@echo "$(GREEN)═══════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(CYAN)📚 访问文档:$(NC)"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  Health:     http://localhost:8080/health"
	@echo ""
	@echo "$(CYAN)🎨 启动前端:$(NC)"
	@echo "  make web-dev"
	@echo "  然后访问: http://localhost:5173"
	@echo ""
	@echo "$(CYAN)🧪 测试 API:$(NC)"
	@echo "  curl http://localhost:8080/health"
	@echo "  curl \"http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY\""
	@echo ""
	@echo "$(CYAN)📋 查看日志:$(NC)"
	@echo "  make logs"
	@echo ""
	@echo "$(CYAN)🛑 停止服务:$(NC)"
	@echo "  make stop"
	@echo ""

# ==================== 完整检查 ====================

check: fmt vet test ## ✅ 完整检查（格式化 + 静态分析 + 测试）
	@echo "$(GREEN)✓ 所有检查通过！$(NC)"

# ==================== 开发流程 ====================

dev-full: deps build test swagger web-install ## 🎓 完整开发流程
	@echo "$(GREEN)✓ 开发环境准备完成！$(NC)"

# ==================== 全栈开发 ====================

fullstack: dev ## 🚀 启动全栈开发环境
	@echo "$(CYAN)启动后端...$(NC)"
	@make run &
	@sleep 3
	@echo "$(CYAN)启动前端...$(NC)"
	@make web-dev
