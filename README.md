# 🌊 Rateflow

> A modern, scalable multi-currency exchange rate tracking platform built with Go 1.25+ and React 18+

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md) | [中文](README_CN.md)

---

## ✨ Features

### Backend
- 🚀 **Modern Go**: Leverages Go 1.25+ features (generics, range over func, slog)
- 🎯 **Domain-Driven Design**: Clean architecture with clear separation of concerns
- 📊 **Multi-Currency Support**: Extensible provider system for various data sources
- ⚡ **High Performance**: Redis caching + streaming queries for efficiency
- 🐳 **Container-Ready**: Docker and Kubernetes deployment configs included
- 🔧 **Developer-Friendly**: Comprehensive CLI tools with Cobra

### Frontend
- 🎨 **Modern React 18**: Built with TypeScript and Material-UI v5
- 🌍 **Internationalization**: Full i18n support (English/Chinese)
- 🌓 **Dark Mode**: Persistent theme switching with localStorage
- 📈 **Rich Visualizations**:
  - 7-day mini sparkline charts with trend indicators
  - Interactive rate charts with smart Y-axis scaling
  - Custom date range selector (1-365 days)
  - Real-time data updates
- 💱 **Currency Tools**:
  - Dual-dropdown currency pair selector with swap button
  - Real-time currency converter
  - 30-day statistics (high/low/average/change)
- 📤 **Data Export**: Export historical data as CSV or JSON
- 📱 **Responsive Design**: Optimized for desktop, tablet, and mobile

---

## 🏗️ Architecture

### System Architecture

```mermaid
graph TB
    subgraph "Frontend"
        UI[React 18 + MUI]
    end

    subgraph "API Layer"
        API[Gin HTTP Server<br/>:8080]
    end

    subgraph "Application Layer"
        Query[Query Handlers<br/>CQRS Read]
        Command[Command Handlers<br/>CQRS Write]
    end

    subgraph "Domain Layer"
        Entity[Entities<br/>Rate, Currency]
        Repo[Repository Interfaces]
        Prov[Provider Interfaces]
    end

    subgraph "Infrastructure Layer"
        PG[(PostgreSQL<br/>Rate Storage)]
        Redis[(Redis<br/>Cache)]
        UP[UnionPay Provider]
        ECB[ECB Provider]
    end

    subgraph "CLI Tools"
        Worker[Cobra CLI<br/>Worker]
    end

    UI -->|REST API| API
    API --> Query
    API --> Command
    Query --> Repo
    Command --> Repo
    Command --> Prov
    Repo --> PG
    Repo --> Redis
    Prov --> UP
    Prov --> ECB
    Worker --> Command
    Worker --> Query
```

### Data Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Cache
    participant Handler
    participant Repository
    participant Database

    Client->>API: GET /api/v1/rates/latest?pair=CNY/JPY
    API->>Handler: GetLatestRateQuery
    Handler->>Cache: Check cache
    alt Cache Hit
        Cache-->>Handler: Return cached rate
    else Cache Miss
        Handler->>Repository: FindLatest(pair)
        Repository->>Database: SELECT * FROM rates WHERE...
        Database-->>Repository: Rate data
        Repository-->>Handler: Domain Rate entity
        Handler->>Cache: Store in cache (5min TTL)
    end
    Handler-->>API: GetLatestRateResult
    API-->>Client: JSON Response
```

### Project Structure

```
rateflow/
├── cmd/                          # Entry points
│   ├── api/                      # API server
│   └── worker/                   # CLI tools
├── internal/                     # Private application code
│   ├── domain/                   # Domain layer (business logic)
│   │   ├── currency/             # Currency value objects
│   │   ├── rate/                 # Rate aggregate root
│   │   └── provider/             # Provider interfaces
│   ├── application/              # Application layer (use cases)
│   │   ├── query/                # Query handlers (CQRS read)
│   │   ├── command/              # Command handlers (CQRS write)
│   │   └── dto/                  # Data transfer objects
│   ├── infrastructure/           # Infrastructure layer
│   │   ├── config/               # Configuration
│   │   ├── logger/               # Structured logging (slog)
│   │   ├── persistence/          # Data persistence
│   │   │   ├── postgres/         # PostgreSQL implementation
│   │   │   └── redis/            # Redis cache implementation
│   │   └── provider/             # External data sources
│   │       └── unionpay/         # UnionPay API client
│   └── presentation/             # Presentation layer
│       ├── http/                 # HTTP handlers & middleware
│       └── cli/                  # CLI commands
├── pkg/                          # Public reusable packages
│   ├── result/                   # Result type (error handling)
│   ├── option/                   # Option type (nullable values)
│   ├── stream/                   # Stream utilities (range over func)
│   ├── genericrepo/              # Generic repository pattern
│   ├── httputil/                 # HTTP client utilities
│   └── timeutil/                 # Time utilities
├── web/                          # React frontend
│   ├── src/
│   │   ├── features/             # Feature modules
│   │   ├── api/                  # API client
│   │   └── shared/               # Shared components
│   └── package.json
├── deploy/                       # Deployment configurations
│   ├── docker/                   # Dockerfiles
│   └── k8s/                      # Kubernetes manifests
└── docs/                         # Documentation
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16+ (or use Docker)
- Redis 7+ (or use Docker)

### Local Development

1. **Clone the repository**

```bash
git clone https://github.com/tyokyo320/rateflow.git
cd rateflow
```

2. **Start dependencies**

```bash
docker-compose up -d postgres redis
```

3. **Configure environment**

```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Run API server**

```bash
go run cmd/api/main.go
```

5. **Run worker to fetch rates**

```bash
go run cmd/worker/main.go fetch --pair CNY/JPY
```

6. **Test the API**

```bash
# Health check
curl http://localhost:8080/health

# Get latest rate
curl http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY
```

### Docker Deployment

```bash
# Build images
docker build -f deploy/docker/api.Dockerfile -t rateflow-api .
docker build -f deploy/docker/worker.Dockerfile -t rateflow-worker .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f api
```

### Frontend Development

```bash
cd web
npm install
npm run dev
```

Visit `http://localhost:5173` to see the dashboard.

**Frontend Features:**
- 📊 Real-time exchange rate monitoring
- 📈 Interactive charts with 7/14/30/60/90 day views + custom range
- 💱 Currency converter with bidirectional conversion
- 🌓 Dark/Light theme toggle
- 🌍 Language switcher (EN/中文)
- 📤 Export data as CSV or JSON
- 📱 Fully responsive design

---

## 📖 API Documentation

### Endpoints

#### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

#### Get Latest Rate

```http
GET /api/v1/rates/latest?pair=CNY/JPY
```

**Response:**
```json
{
  "success": true,
  "data": {
    "pair": "CNY/JPY",
    "rate": 0.061234,
    "effectiveDate": "2025-11-02T00:00:00Z",
    "source": "unionpay"
  },
  "timestamp": "2025-11-02T10:30:00Z"
}
```

#### Get Rate History

```http
GET /api/v1/rates/history?pair=CNY/JPY&days=7
```

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "pair": "CNY/JPY",
        "rate": 0.061234,
        "effectiveDate": "2025-11-02T00:00:00Z",
        "source": "unionpay"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 7,
      "total": 100,
      "totalPages": 15
    }
  }
}
```

---

## 🔧 CLI Usage

### Fetch Rates

```bash
# Fetch today's rate for CNY/JPY
./rateflow-worker fetch --pair CNY/JPY

# Fetch historical rate
./rateflow-worker fetch --pair CNY/JPY --date 2025-11-01

# Fetch multiple pairs
./rateflow-worker fetch --pair USD/JPY
./rateflow-worker fetch --pair EUR/JPY
```

### Consolidate Data

```bash
# Consolidate temporary rates into main table
./rateflow-worker consolidate
```

### Database Migration

```bash
# Run migrations
./rateflow-worker migrate up

# Rollback
./rateflow-worker migrate down
```

### Database Initialization

The database schema is automatically migrated using GORM AutoMigrate when the API server starts. However, you need to populate initial rate data.

#### For Docker Users

```bash
# 1. Start the services
docker-compose up -d

# 2. The database will be auto-migrated on first API start

# 3. Fetch initial rate data
docker-compose exec api ./rateflow-worker fetch --pair CNY/JPY

# Or use docker run
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

#### For Kubernetes Users

```bash
# 1. Deploy the application
kubectl apply -k deploy/k8s

# 2. Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=rateflow-api -n rateflow --timeout=60s

# 3. Initialize rate data
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

# The CronJob will automatically fetch new rates hourly
```

#### For Local Development (without Docker)

```bash
# 1. Make sure PostgreSQL and Redis are running
# PostgreSQL 17 on localhost:5432
# Redis 8 on localhost:6379

# 2. Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=rateflow
export DB_PASSWORD=your_password
export DB_NAME=rateflow
export DB_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export LOG_LEVEL=debug

# 3. Run the API (auto-migrates database)
go run cmd/api/main.go

# 4. In another terminal, fetch initial data
go run cmd/worker/main.go fetch --pair CNY/JPY

# 5. Fetch historical data (optional)
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31
```

#### Verify Database

```bash
# Docker
docker-compose exec postgres psql -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"

# Kubernetes
kubectl exec -it -n rateflow statefulset/postgres -- psql -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"

# Local
psql -h localhost -U rateflow -d rateflow -c "SELECT COUNT(*) FROM rates;"
```

---

## 🎯 Core Concepts

### 1. Generics (Go 1.25+)

```go
// Generic repository for any entity type
type Repository[T Entity] interface {
    Create(ctx context.Context, entity T) error
    FindByID(ctx context.Context, id string) (T, error)
    Stream(ctx context.Context) iter.Seq[T] // Go 1.23 range over func
}
```

### 2. Range Over Function

```go
// Memory-efficient streaming of large datasets
for rate := range rateRepo.Stream(ctx) {
    process(rate)
    // Early termination supported
    if someCondition {
        break
    }
}
```

### 3. Result Pattern

```go
// Elegant error handling without exceptions
result := GetLatestRate(ctx, pair)

// Chain operations
finalResult := result.
    Map(func(r Rate) Rate { return r.WithDiscount() }).
    UnwrapOr(defaultRate)
```

### 4. Domain-Driven Design

```go
// Value Object - Currency Pair
pair, _ := currency.NewPair(currency.CNY, currency.JPY)

// Aggregate Root - Rate Entity
rate, _ := rate.NewRate(pair, 0.061234, time.Now(), rate.SourceUnionPay)

// Domain validation
if err := rate.Validate(); err != nil {
    // Handle validation error
}
```

### 5. CQRS Pattern

```go
// Query (Read) - Optimized for reading
type GetLatestRateQuery struct {
    Pair currency.Pair
}

// Command (Write) - Optimized for writing
type FetchRateCommand struct {
    Pair currency.Pair
    Date time.Time
}
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain/rate/...

# Run with race detector
go test -race ./...
```

---

## 📊 Performance

- **API Response Time**: < 50ms (with cache)
- **Cache Hit Rate**: > 90% for latest rates
- **Throughput**: > 1000 req/s (single instance)
- **Database Connection Pool**: 25 connections
- **Memory Usage**: ~50MB (idle), ~200MB (peak)

---

## 🛠️ Configuration

### Environment Variables

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=production

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=rateflow
DB_PASSWORD=your_secure_password
DB_NAME=rateflow
DB_SSLMODE=require
DB_MAX_CONNS=25

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# Logging
LOG_LEVEL=info        # debug, info, warn, error
LOG_FORMAT=json       # json, text

# Config File (optional)
CONFIG_PATH=./config.json
```

### Config File

See [config.json.example](config.json.example) for a complete configuration file template.

---

## 🚢 Deployment

### Docker Compose

```bash
docker-compose up -d
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f deploy/k8s/

# Check status
kubectl get pods -l app=rateflow

# View logs
kubectl logs -f deployment/rateflow-api
```

### CronJob (Scheduled Tasks)

```yaml
# Fetch rates hourly
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rateflow-fetch
spec:
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            image: rateflow-worker:latest
            args: ["fetch", "--pair", "CNY/JPY"]
```

---

## 🤝 Contributing

Contributions are welcome! Please follow our development workflow:

### Branch Strategy

- `master` - Production-ready code, protected branch
- `develop` - Development branch for integration

### Development Workflow

1. **Fork the repository**

2. **Clone and create feature branch from develop**
   ```bash
   git clone https://github.com/yourusername/rateflow.git
   cd rateflow
   git checkout develop
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Write code following the project conventions
   - Add tests for new features
   - Update documentation as needed

4. **Commit your changes**
   ```bash
   git add .
   git commit -m 'feat: add amazing feature'
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `refactor:` - Code refactoring
   - `test:` - Adding tests
   - `chore:` - Maintenance tasks

5. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```

6. **Create Pull Request**
   - Target the `develop` branch
   - Fill in the PR template
   - Wait for CI checks to pass
   - Request review from maintainers

7. **After PR is approved**
   - Maintainers will merge to `develop`
   - Periodically, `develop` will be merged to `master`

### Release Process

Releases are created from the `master` branch:

1. **Create a release tag**
   ```bash
   git checkout master
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Automated release workflow**
   - GitHub Actions automatically builds binaries
   - Docker images are built and pushed (multi-arch: amd64/arm64)
   - GitHub Release is created with changelog
   - Images tagged as `v1.0.0` and `latest`

3. **Available artifacts**
   - Docker images: `tyokyo320/rateflow-api:v1.0.0`, `tyokyo320/rateflow-worker:v1.0.0`
   - Linux binaries: `rateflow-api-linux-amd64`, `rateflow-worker-linux-amd64`
   - Checksums for verification

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP framework
- [GORM](https://gorm.io/) - ORM library
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Material-UI](https://mui.com/) - React UI framework
- UnionPay International for providing exchange rate data

---

## 📧 Contact

- **Author**: tyokyo320
- **Website**: https://rate.tyokyo320.com
- **GitHub**: [@tyokyo320](https://github.com/tyokyo320)

---

<div align="center">

**Made with ❤️ using Go 1.25+ and React 18+**

[Report Bug](https://github.com/tyokyo320/rateflow/issues) · [Request Feature](https://github.com/tyokyo320/rateflow/issues)

</div>
