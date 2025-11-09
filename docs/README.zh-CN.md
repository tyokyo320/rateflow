# RateFlow - 汇率追踪平台

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/Version-1.5.2-blue.svg)](https://github.com/tyokyo320/rateflow/releases)

> 基于 Go 1.25+ 和银联数据的现代化汇率追踪系统

[English](../README.md) | **简体中文**

## 📖 简介

RateFlow 是一个采用领域驱动设计(DDD)和 CQRS 架构模式构建的现代化汇率追踪平台。系统从银联 API 获取多币种汇率数据,提供 REST API 和 Web 界面,支持历史数据查询和实时汇率追踪。

**官方网站**: https://rateflow.tyokyo320.com

### ✨ 核心特性

- 🌍 **多币种支持**: 支持 CNY、JPY、USD、EUR、GBP 等主流货币
- 📊 **实时数据**: 从银联 API 获取最新汇率数据
- 📈 **历史追踪**: 完整的历史汇率数据存储和查询
- ⚡ **高性能缓存**: 基于 Redis 的智能缓存策略
- 🔄 **批量获取**: 支持货币矩阵批量获取功能
- 🛠️ **CLI 工具**: 功能完善的命令行工具
- 📱 **现代 UI**: 基于 React 18 + Material-UI 的响应式界面
- 🐳 **容器化部署**: 完整的 Docker 和 Docker Compose 支持

### 🏗️ 技术栈

**后端**
- Go 1.25+ (泛型、迭代器、结构化日志)
- Gin (HTTP 框架)
- GORM (ORM)
- PostgreSQL 17 (数据库)
- Redis 8 (缓存)
- Cobra (CLI 框架)
- Swagger (API 文档)

**前端**
- React 18
- TypeScript
- Material-UI (MUI)
- Recharts / MUI X Charts (图表)
- React Query (数据管理)
- Vite (构建工具)

**架构模式**
- Domain-Driven Design (DDD)
- CQRS (命令查询职责分离)
- Clean Architecture (整洁架构)
- Repository Pattern (仓储模式)

## 🚀 快速开始

### 前置要求

- Go 1.25 或更高版本
- Docker & Docker Compose
- Node.js 18+ (用于前端开发)
- Make (可选,用于快捷命令)

### 使用 Make 快速启动

```bash
# 一键启动(推荐)
make quickstart

# 启动后端服务
docker-compose up -d

# 启动前端开发服务器
make web-dev
```

### 手动启动

```bash
# 1. 启动数据库服务
docker-compose up -d postgres redis

# 2. 启动 API 服务
go run cmd/api/main.go

# 3. 启动前端(新终端)
cd web
npm install
npm run dev
```

### 访问服务

- **前端界面**: http://localhost:5173
- **API 服务**: http://localhost:8080
- **Swagger 文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

## 📚 使用指南

### Worker CLI 命令

#### 1. 批量获取汇率矩阵

```bash
# 获取 CNY、JPY、USD 之间所有组合的汇率(6个货币对)
docker-compose exec api /app/rateflow-worker fetch-matrix --currencies CNY,JPY,USD

# 获取指定日期的数据
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD,EUR,GBP \
  --date 2024-11-08

# 获取日期范围数据
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --start 2024-11-01 \
  --end 2024-11-08

# 强制重新获取(覆盖已存在数据)
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --force
```

**货币矩阵说明**:
- 3个货币 → 6个货币对 (3×2)
- 4个货币 → 12个货币对 (4×3)
- 5个货币 → 20个货币对 (5×4)

#### 2. 获取单个货币对

```bash
# 获取最新汇率
docker-compose exec api /app/rateflow-worker fetch --pair CNY/JPY

# 获取指定日期汇率
docker-compose exec api /app/rateflow-worker fetch --pair JPY/USD --date 2024-11-08

# 获取日期范围汇率
docker-compose exec api /app/rateflow-worker fetch \
  --pair CNY/USD \
  --start 2024-11-01 \
  --end 2024-11-08
```

#### 3. 清理数据

```bash
# 预览要删除的数据(干运行)
docker-compose exec api /app/rateflow-worker clean --pair JPY/USD --dry-run

# 删除指定货币对的所有数据
docker-compose exec api /app/rateflow-worker clean --pair JPY/USD

# 删除指定日期之前的数据
docker-compose exec api /app/rateflow-worker clean --before 2024-01-01

# 删除指定日期范围的数据
docker-compose exec api /app/rateflow-worker clean \
  --pair CNY/JPY \
  --after 2024-01-01 \
  --before 2024-12-31
```

### REST API 示例

#### 获取最新汇率

```bash
# 获取 CNY/JPY 最新汇率
curl "http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY"

# 获取 USD/JPY 最新汇率
curl "http://localhost:8080/api/v1/rates/latest?pair=USD/JPY"
```

**响应示例**:
```json
{
  "data": {
    "id": "6d698d32-e67a-4335-ad5e-d95f56e22bca",
    "pair": "USD/JPY",
    "baseCurrency": "USD",
    "quoteCurrency": "JPY",
    "rate": 153.554285,
    "effectiveDate": "2024-11-08T00:00:00Z",
    "source": "unionpay",
    "createdAt": "2025-11-08T20:10:00.252788+08:00",
    "updatedAt": "2025-11-08T20:10:00.252812+08:00"
  },
  "success": true
}
```

#### 获取历史汇率

```bash
# 获取最近7天的历史汇率
curl "http://localhost:8080/api/v1/rates/history?pair=CNY/JPY&days=7"

# 获取最近30天的历史汇率
curl "http://localhost:8080/api/v1/rates/history?pair=USD/JPY&days=30"
```

## 🏛️ 架构设计

### 分层架构

```
┌─────────────────────────────────────────┐
│   表示层 (Presentation Layer)          │
│   - HTTP 处理器 (Gin)                   │
│   - CLI 命令 (Cobra)                    │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│   应用层 (Application Layer)           │
│   - 命令处理器 (CQRS Write)             │
│   - 查询处理器 (CQRS Read)              │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│   领域层 (Domain Layer)                 │
│   - 实体 (Entities)                     │
│   - 值对象 (Value Objects)              │
│   - 仓储接口 (Repository Interfaces)    │
└─────────────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│   基础设施层 (Infrastructure Layer)     │
│   - PostgreSQL (GORM)                  │
│   - Redis (缓存)                        │
│   - 银联 API 客户端                      │
└─────────────────────────────────────────┘
```

### 核心设计模式

1. **领域驱动设计 (DDD)**
   - 聚合根: `rate.Rate` - 汇率实体包含业务逻辑
   - 值对象: `currency.Code`, `currency.Pair` - 不可变、自验证
   - 仓储模式: 领域层定义接口,基础设施层实现

2. **CQRS (命令查询职责分离)**
   - 命令: 写操作 (例如 `FetchRateCommand`)
   - 查询: 读操作,带缓存 (例如 `GetLatestRateQuery`)

3. **Go 1.25+ 现代特性**
   - 泛型: 类型安全的仓储基类
   - 迭代器: 内存高效的数据流查询
   - 结构化日志: 生产级日志记录

### 项目结构

```
rateflow/
├── cmd/                      # 应用程序入口点
│   ├── api/                 # HTTP 服务器
│   └── worker/              # CLI 工具
├── internal/                # 内部应用代码
│   ├── domain/              # 领域层(纯业务逻辑)
│   │   ├── currency/        # 货币值对象
│   │   ├── rate/            # 汇率聚合根
│   │   └── provider/        # 提供者接口
│   ├── application/         # 应用层(用例)
│   │   ├── command/         # 写操作(CQRS)
│   │   ├── query/           # 读操作(CQRS)
│   │   └── dto/             # 数据传输对象
│   ├── infrastructure/      # 基础设施实现
│   │   ├── config/          # 配置加载
│   │   ├── logger/          # 日志封装
│   │   ├── persistence/     # 持久化
│   │   │   ├── postgres/    # GORM 实现
│   │   │   └── redis/       # 缓存实现
│   │   └── provider/        # 外部 API
│   │       └── unionpay/    # 银联客户端
│   └── presentation/        # 表示层
│       └── http/            # Gin 路由/处理器
├── pkg/                     # 可复用的公共包
│   ├── result/              # Result[T] 类型
│   ├── option/              # Option[T] 类型
│   ├── stream/              # 迭代器工具
│   └── genericrepo/         # 泛型仓储
├── web/                     # 前端应用
│   ├── src/
│   │   ├── api/             # API 客户端
│   │   ├── components/      # React 组件
│   │   ├── features/        # 功能模块
│   │   └── utils/           # 工具函数
│   └── package.json
└── docs/                    # 文档目录
    ├── README.zh-CN.md      # 中文文档
    ├── MIGRATION.zh-CN.md   # 中文迁移指南
    └── swagger/             # Swagger 文档
```

## 🔧 开发指南

### Make 命令

```bash
# 查看所有可用命令
make help

# 开发
make dev                     # 启动数据库
make run                     # 运行 API 服务器
make web-dev                 # 启动前端开发服务器

# Docker
make docker-up               # 启动所有服务
make docker-down             # 停止所有服务
make docker-rebuild          # 完全重建
make docker-logs             # 查看日志

# 测试和质量
make test                    # 运行测试
make test-cover              # 生成覆盖率报告
make fmt                     # 格式化代码
make vet                     # 静态分析
make lint                    # 运行所有检查

# 构建
make build                   # 构建 API 二进制
make build-worker            # 构建 Worker 二进制

# 文档
make swagger                 # 生成 Swagger 文档
```

### 配置

配置优先级(从高到低):
1. 环境变量
2. `CONFIG_PATH` 指定的 JSON 配置文件
3. 嵌入的默认配置

**环境变量示例** (`.env`):
```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=rateflow
DB_SSLMODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
```

**JSON 配置示例** (`config.json`):
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "readTimeout": 15000000000,
    "writeTimeout": 15000000000,
    "environment": "dev"
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "database": "rateflow",
    "sslMode": "disable",
    "timezone": "Asia/Shanghai",
    "maxConns": 25
  },
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  },
  "logger": {
    "level": "debug",
    "format": "text"
  }
}
```

### 缓存策略

- **缓存键格式**: `latest:{pair}` (例如: `latest:CNY/JPY`)
- **TTL**: 最新汇率缓存 5 分钟
- **策略**: Cache-Aside (检查缓存 → 查询数据库 → 写入缓存)
- **失效**: 自动过期或命令处理器手动清除

## 📦 部署

### Docker 部署

```bash
# 构建并启动
docker-compose build
docker-compose up -d

# 查看日志
docker-compose logs -f api

# 停止服务
docker-compose down

# 完全清理(包括数据卷)
docker-compose down -v
```

### 生产环境建议

1. **数据库优化**
   - 增加 `max_connections`
   - 配置连接池大小
   - 启用查询日志

2. **Redis 配置**
   - 启用持久化 (AOF/RDB)
   - 配置内存限制
   - 设置密码

3. **API 服务**
   - 使用生产级配置
   - 启用 JSON 日志
   - 配置适当的超时时间

4. **监控和日志**
   - 使用结构化日志
   - 集成 APM 工具
   - 设置健康检查

## 🔄 从 v1.3.1 迁移

v1.4.0 修复了银联 API 解析的严重 bug。如果您从 v1.3.1 升级,请参阅 [迁移指南](./MIGRATION.zh-CN.md)。

### 关键变更

- ✅ 修复了银联汇率解析逻辑,支持双向匹配
- ✅ 新增 `fetch-matrix` 批量获取命令
- ✅ 新增 `clean` 数据清理命令
- ✅ 所有旧数据需要清理并重新获取

## 🤝 贡献指南

我们欢迎任何形式的贡献!

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 运行 `make lint` 检查代码
- 添加适当的测试
- 更新相关文档

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../LICENSE) 文件。

## 📧 联系方式

- **项目主页**: https://github.com/tyokyo320/rateflow
- **问题反馈**: https://github.com/tyokyo320/rateflow/issues
- **官方网站**: https://rateflow.tyokyo320.com

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [React](https://react.dev/) - 前端框架
- [Material-UI](https://mui.com/) - UI 组件库
- 银联国际 - 汇率数据提供

---

**使用愉快!** 🎉

如果觉得这个项目有帮助,请给个 ⭐️ 吧!
