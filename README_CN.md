# 🌊 Rateflow

> 使用 Go 1.25+ 和 React 18+ 构建的现代化多货币汇率追踪平台

[English](README.md) | [中文](README_CN.md)

---

## ✨ 特性

- 🚀 **现代 Go**: 充分利用 Go 1.25+ 新特性（泛型、range over func、slog）
- 🎯 **领域驱动设计**: 清晰的架构分层
- 📊 **多货币支持**: 可扩展的数据源提供商系统
- ⚡ **高性能**: Redis 缓存 + 流式查询
- 🎨 **现代前端**: React 18 + Material-UI + TypeScript
- 🐳 **容器化**: 包含 Docker 和 Kubernetes 部署配置
- 🔧 **开发友好**: 使用 Cobra 的完整 CLI 工具

---

## 🏗️ 架构

### 系统架构

```
Frontend (React 18 + MUI)
         ↓
API Layer (Gin HTTP Server)
         ↓
Application Layer (CQRS)
    ↙          ↘
Query        Command
         ↓
Domain Layer (DDD)
    ↙     ↓      ↘
Entity  Repo  Provider
         ↓
Infrastructure Layer
    ↙     ↓      ↘
PostgreSQL Redis UnionPay
```

### 项目结构

```
rateflow/
├── cmd/                    # 入口程序
│   ├── api/               # API 服务
│   └── worker/            # CLI 工具
├── internal/              # 私有应用代码
│   ├── domain/           # 领域层（业务核心）
│   ├── application/      # 应用层（用例）
│   ├── infrastructure/   # 基础设施层
│   └── presentation/     # 表现层
├── pkg/                   # 公共可复用包
│   ├── result/           # Result 模式
│   ├── option/           # Option 模式
│   ├── stream/           # 流式处理
│   ├── genericrepo/      # 泛型仓储
│   └── ...
├── web/                   # React 前端
└── deploy/                # 部署配置
```

---

## 🚀 快速开始

### 前置要求

- Go 1.25+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 17+
- Redis 8+

### 本地开发

1. **克隆仓库**

```bash
cd /home/zhangqiang/work/repos/union-pay
```

2. **启动依赖服务**

```bash
make db-up
```

3. **配置环境**

```bash
cp .env.example .env
# 编辑 .env 文件
```

4. **运行 API 服务**

```bash
make run-api
```

5. **测试 API**

```bash
# 健康检查
curl http://localhost:8080/health

# 获取最新汇率
curl http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY
```

---


### 数据库初始化

数据库表结构会在 API 服务器启动时通过 GORM AutoMigrate 自动创建。但你需要手动获取初始汇率数据。

#### Docker 用户

```bash
# 1. 启动服务
docker-compose up -d

# 2. 数据库会在 API 首次启动时自动迁移

# 3. 获取初始汇率数据
docker-compose exec api ./rateflow-worker fetch --pair CNY/JPY

# 或者使用 docker run
docker run --rm --network rateflow_default \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=rateflow \
  -e DB_PASSWORD=rateflow_password \
  -e DB_NAME=rateflow \
  -e DB_SSLMODE=disable \
  tyokyo320/rateflow-worker:latest \
  fetch --pair CNY/JPY
```

#### Kubernetes 用户

```bash
# 1. 部署应用
kubectl apply -k deploy/k8s

# 2. 等待 Pod 就绪
kubectl wait --for=condition=ready pod -l app=rateflow-api -n rateflow --timeout=60s

# 3. 初始化汇率数据
kubectl run -it --rm rateflow-init \
  --image=tyokyo320/rateflow-worker:latest \
  --restart=Never \
  --namespace=rateflow \
  --env="DB_HOST=postgres" \
  --env="DB_PORT=5432" \
  --env="DB_USER=rateflow" \
  --env="DB_NAME=rateflow" \
  --env="DB_PASSWORD=your_password" \
  --env="DB_SSLMODE=disable" \
  -- fetch --pair CNY/JPY

# CronJob 会每小时自动获取新汇率
```

#### 本地开发（不使用 Docker）

```bash
# 1. 确保 PostgreSQL 和 Redis 正在运行
# PostgreSQL 17 在 localhost:5432
# Redis 8 在 localhost:6379

# 2. 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=rateflow
export DB_PASSWORD=your_password
export DB_NAME=rateflow
export DB_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export LOG_LEVEL=debug

# 3. 运行 API（自动迁移数据库）
go run cmd/api/main.go

# 4. 在另一个终端获取初始数据
go run cmd/worker/main.go fetch --pair CNY/JPY

# 5. 获取历史数据（可选）
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31
```

#### 验证数据库

```bash
# Docker
docker-compose exec postgres psql -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"

# Kubernetes
kubectl exec -it -n rateflow statefulset/postgres -- psql -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"

# 本地
psql -h localhost -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"
```

---

## 📖 核心概念

### 1. Go 1.23 新特性

#### Range Over Function（流式处理）
```go
// 内存高效的大数据集处理
for rate := range rateRepo.Stream(ctx) {
    process(rate)
}
```

#### 泛型
```go
// 通用仓储，支持任何实体类型
type Repository[T Entity] interface {
    Create(ctx context.Context, entity T) error
    FindByID(ctx context.Context, id string) (T, error)
    Stream(ctx context.Context) iter.Seq[T]
}
```

#### 结构化日志（slog）
```go
slog.Info("rate fetched",
    "pair", "CNY/JPY",
    "rate", 0.061234,
    slog.Group("metadata",
        "source", "unionpay",
    ),
)
```

### 2. Result 模式

```go
// 优雅的错误处理
result := GetLatestRate(ctx, pair)

finalResult := result.
    Map(func(r Rate) Rate { return r.WithDiscount() }).
    UnwrapOr(defaultRate)
```

### 3. 领域驱动设计

```go
// 值对象
pair, _ := currency.NewPair(currency.CNY, currency.JPY)

// 聚合根
rate, _ := rate.NewRate(pair, 0.061234, time.Now(), rate.SourceUnionPay)

// 领域验证
if err := rate.Validate(); err != nil {
    // 处理验证错误
}
```

---

## 🛠️ 开发命令

```bash
# 构建
make build

# 运行测试
make test

# 代码检查
make lint

# 格式化代码
make fmt

# 启动开发环境
make dev

# Docker 构建
make docker-build

# 启动所有服务
make docker-up
```

---

## 📚 文档

- [实施指南](IMPLEMENTATION_GUIDE.md) - 完整的实现步骤和示例代码
- [重构方案](REFACTOR_PLAN.md) - 详细的技术设计和架构决策
- [项目摘要](PROJECT_SUMMARY.md) - 当前进度和技术亮点
- [API 文档](README.md#api-documentation) - REST API 详细说明

---

## 🎯 技术栈

### 后端
- **语言**: Go 1.23
- **Web 框架**: Gin
- **ORM**: GORM
- **缓存**: Redis
- **CLI**: Cobra
- **日志**: slog (官方)
- **依赖注入**: Wire

### 前端
- **框架**: React 18
- **UI 库**: Material-UI (MUI)
- **状态管理**: TanStack Query + Zustand
- **构建工具**: Vite
- **语言**: TypeScript

### 基础设施
- **数据库**: PostgreSQL 17
- **缓存**: Redis 8
- **容器**: Docker
- **编排**: Kubernetes
- **CI/CD**: GitHub Actions

---

## 📊 性能指标

- **API 响应时间**: < 50ms（使用缓存）
- **缓存命中率**: > 90%（最新汇率）
- **吞吐量**: > 1000 req/s（单实例）
- **内存使用**: ~50MB（空闲）, ~200MB（峰值）

---

## 🤝 贡献

欢迎贡献！请遵循我们的开发工作流程：

### 分支策略

- `master` - 生产就绪代码，受保护分支
- `develop` - 开发分支，用于集成

### 开发工作流程

1. **Fork 仓库**

2. **克隆并从 develop 创建特性分支**
   ```bash
   git clone https://github.com/yourusername/rateflow.git
   cd rateflow
   git checkout develop
   git checkout -b feature/amazing-feature
   ```

3. **进行更改**
   - 遵循项目规范编写代码
   - 为新功能添加测试
   - 根据需要更新文档

4. **提交更改**
   ```bash
   git add .
   git commit -m 'feat: 添加某某功能'
   ```

   遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/):
   - `feat:` - 新功能
   - `fix:` - 错误修复
   - `docs:` - 文档更改
   - `refactor:` - 代码重构
   - `test:` - 添加测试
   - `chore:` - 维护任务

5. **推送到你的 fork**
   ```bash
   git push origin feature/amazing-feature
   ```

6. **创建 Pull Request**
   - 目标分支选择 `develop`
   - 填写 PR 模板
   - 等待 CI 检查通过
   - 请求维护者审查

7. **PR 批准后**
   - 维护者将合并到 `develop`
   - 定期将 `develop` 合并到 `master`

### 发布流程

从 `master` 分支创建发布：

1. **创建发布标签**
   ```bash
   git checkout master
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **自动化发布工作流**
   - GitHub Actions 自动构建二进制文件
   - 构建并推送 Docker 镜像（多架构：amd64/arm64）
   - 创建 GitHub Release 并生成更新日志
   - 镜像标记为 `v1.0.0` 和 `latest`

3. **可用的构建产物**
   - Docker 镜像：`tyokyo320/rateflow-api:v1.0.0`, `tyokyo320/rateflow-worker:v1.0.0`
   - Linux 二进制文件：`rateflow-api-linux-amd64`, `rateflow-worker-linux-amd64`
   - 校验和文件用于验证

---

## 📝 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 📧 联系方式

- **作者**: tyokyo320
- **网站**: https://rate.tyokyo320.com
- **GitHub**: [@tyokyo320](https://github.com/tyokyo320)

---

<div align="center">

**使用 Go 1.25+ 和 React 18+ 精心打造 ❤️**

[报告 Bug](https://github.com/tyokyo320/rateflow/issues) · [请求功能](https://github.com/tyokyo320/rateflow/issues)

</div>
