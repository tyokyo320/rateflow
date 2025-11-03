#!/bin/bash

# RateFlow 一键启动脚本
# 用于本地开发和测试

set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           RateFlow 本地开发环境启动脚本                   ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查依赖
check_dependency() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 未安装${NC}"
        echo "  请先安装 $1"
        return 1
    else
        echo -e "${GREEN}✓ $1 已安装${NC}"
        return 0
    fi
}

echo "1️⃣  检查依赖..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

all_deps_ok=true
check_dependency "go" || all_deps_ok=false
check_dependency "docker" || all_deps_ok=false
check_dependency "docker-compose" || all_deps_ok=false

if [ "$all_deps_ok" = false ]; then
    echo ""
    echo -e "${RED}请先安装缺失的依赖${NC}"
    exit 1
fi

echo ""
echo "2️⃣  创建配置文件..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 创建 docker-compose.yml
if [ ! -f "docker-compose.yml" ]; then
    echo "创建 docker-compose.yml..."
    cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: rateflow-postgres
    environment:
      POSTGRES_DB: rateflow
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: rateflow-redis
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  postgres_data:
EOF
    echo -e "${GREEN}✓ docker-compose.yml 创建成功${NC}"
else
    echo -e "${YELLOW}⚠ docker-compose.yml 已存在，跳过${NC}"
fi

# 创建 config.yml
if [ ! -f "config.yml" ]; then
    echo "创建 config.yml..."
    cat > config.yml << 'EOF'
server:
  host: "0.0.0.0"
  port: 8080
  environment: "development"
  read_timeout: 10s
  write_timeout: 10s

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "rateflow"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

logger:
  level: "info"
  format: "json"
  output: "stdout"
EOF
    echo -e "${GREEN}✓ config.yml 创建成功${NC}"
else
    echo -e "${YELLOW}⚠ config.yml 已存在，跳过${NC}"
fi

echo ""
echo "3️⃣  启动数据库服务..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 启动 Docker Compose
docker-compose up -d

echo "等待服务就绪..."
sleep 5

# 检查服务状态
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ PostgreSQL 和 Redis 已启动${NC}"
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    docker-compose logs
    exit 1
fi

echo ""
echo "4️⃣  验证服务连接..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 测试 PostgreSQL
if docker exec rateflow-postgres pg_isready -U postgres &> /dev/null; then
    echo -e "${GREEN}✓ PostgreSQL 连接正常${NC}"
else
    echo -e "${RED}✗ PostgreSQL 连接失败${NC}"
fi

# 测试 Redis
if docker exec rateflow-redis redis-cli ping | grep -q "PONG"; then
    echo -e "${GREEN}✓ Redis 连接正常${NC}"
else
    echo -e "${RED}✗ Redis 连接失败${NC}"
fi

echo ""
echo "5️⃣  下载 Go 依赖..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

go mod download
echo -e "${GREEN}✓ 依赖下载完成${NC}"

echo ""
echo "6️⃣  运行测试..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if go test ./internal/... -short > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 所有测试通过${NC}"
else
    echo -e "${YELLOW}⚠ 部分测试失败（可能需要数据）${NC}"
fi

echo ""
echo "7️⃣  构建 API 服务..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if go build -o rateflow-api cmd/api/main.go; then
    echo -e "${GREEN}✓ API 构建成功${NC}"
else
    echo -e "${RED}✗ API 构建失败${NC}"
    exit 1
fi

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                  🎉 启动成功！                            ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "📋 服务信息:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PostgreSQL:  localhost:5432"
echo "  Redis:       localhost:6379"
echo "  API 二进制:  ./rateflow-api"
echo ""
echo "🚀 启动 API 服务:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  # 方式 1: 直接运行"
echo "  ./rateflow-api"
echo ""
echo "  # 方式 2: 使用 go run"
echo "  go run cmd/api/main.go"
echo ""
echo "  # 方式 3: 后台运行"
echo "  nohup ./rateflow-api > api.log 2>&1 &"
echo ""
echo "📖 访问文档:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Swagger UI:  http://localhost:8080/swagger/index.html"
echo "  Health:      http://localhost:8080/health"
echo ""
echo "🧪 测试 API:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  curl http://localhost:8080/health"
echo "  curl \"http://localhost:8080/api/rates/latest?pair=CNY/JPY\""
echo ""
echo "🛑 停止服务:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Ctrl+C              # 停止 API"
echo "  docker-compose down # 停止数据库"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 询问是否启动 API
read -p "是否现在启动 API 服务？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "正在启动 API 服务..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    ./rateflow-api
fi
