# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Rateflow is a modern Go 1.25+ exchange rate tracking platform built with Domain-Driven Design (DDD) and CQRS principles. It tracks UnionPay JPY to CNY exchange rates, storing them in PostgreSQL with Redis caching, and provides REST APIs and CLI tools for data access.

**Key Technologies**: Go 1.25+, Gin, GORM, Redis, PostgreSQL, Cobra CLI, Swagger, React 18

**Website**: https://rateflow.tyokyo320.com

## Architecture

### Clean Architecture Layers

The project follows strict layered architecture with dependency inversion:

```
Presentation Layer (HTTP/CLI)
    ↓
Application Layer (CQRS - Commands & Queries)
    ↓
Domain Layer (Entities, Value Objects, Repository Interfaces)
    ↓
Infrastructure Layer (PostgreSQL, Redis, External Providers)
```

**Dependency Rule**: Inner layers never depend on outer layers. Domain layer is dependency-free.

### Key Architectural Patterns

1. **Domain-Driven Design (DDD)**
   - **Aggregate Root**: `rate.Rate` - exchange rate entity with business logic
   - **Value Objects**: `currency.Code`, `currency.Pair` - immutable, self-validating
   - **Repository Pattern**: `rate.Repository` interface in domain, implemented in infrastructure

2. **CQRS (Command Query Responsibility Segregation)**
   - **Commands** (`internal/application/command/`): Write operations (e.g., `FetchRateCommand`)
   - **Queries** (`internal/application/query/`): Read operations with caching (e.g., `GetLatestRateQuery`)

3. **Go 1.25+ Modern Features**
   - **Generics**: `genericrepo.Repository[T]` provides type-safe repository base
   - **Range over Function**: `iter.Seq[T]` for memory-efficient streaming queries
   - **Structured Logging**: `log/slog` for production-grade logging

4. **Functional Patterns**
   - **Result Type** (`pkg/result/`): Rust-inspired error handling without exceptions
   - **Option Type** (`pkg/option/`): Null-safe value handling
   - **Stream Utilities** (`pkg/stream/`): Functional operations on iterators

### Project Structure

```
rateflow/
├── cmd/
│   ├── api/main.go              # HTTP server entry point (port 8080)
│   └── worker/                  # Cobra CLI for rate fetching
├── internal/
│   ├── domain/                  # Domain layer (pure business logic)
│   │   ├── currency/            # Currency value objects (Code, Pair)
│   │   ├── rate/                # Rate aggregate root + Repository interface
│   │   └── provider/            # Provider interface
│   ├── application/             # Application layer (use cases)
│   │   ├── command/             # Write operations (CQRS)
│   │   ├── query/               # Read operations with caching (CQRS)
│   │   └── dto/                 # Data Transfer Objects
│   ├── infrastructure/          # Infrastructure implementations
│   │   ├── config/              # Configuration loading (env + JSON)
│   │   ├── logger/              # slog wrapper
│   │   ├── persistence/
│   │   │   ├── postgres/        # GORM repository implementation
│   │   │   └── redis/           # Cache implementation
│   │   └── provider/unionpay/   # UnionPay API client
│   └── presentation/            # Presentation layer
│       └── http/                # Gin router, handlers, middleware
└── pkg/                         # Reusable public packages
    ├── result/                  # Result[T] type
    ├── option/                  # Option[T] type
    ├── stream/                  # Iterator utilities
    ├── genericrepo/             # Generic repository pattern
    ├── httputil/                # HTTP client utilities
    └── timeutil/                # Time parsing/formatting
```

## Development Commands

### Using Make (Recommended)

```bash
# Show all available commands
make help

# Quick start (builds Swagger + starts Docker services)
make quickstart

# Local development (database only, run API locally)
make dev                    # Start PostgreSQL + Redis
make run                    # Run API server

# Testing
make test                   # Run all tests
make test-cover             # Generate coverage report

# Code quality
make fmt                    # Format code
make vet                    # Static analysis
make lint                   # Run all checks

# Docker operations
make docker-up              # Start all services
make docker-down            # Stop services
make docker-logs            # View API logs
make docker-rebuild         # Rebuild from scratch

# Build binaries
make build                  # Build API binary
make build-worker           # Build worker binary

# Documentation
make swagger                # Generate Swagger docs
```

### Direct Go Commands

```bash
# Run API server
go run cmd/api/main.go

# Worker commands (using Cobra CLI)
go run cmd/worker/main.go fetch --pair CNY/JPY
go run cmd/worker/main.go fetch --pair CNY/JPY --date 2024-01-15
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-01-31

# Build
go build -o rateflow-api cmd/api/main.go
go build -o rateflow-worker cmd/worker/main.go

# Test
go test ./...                              # All tests
go test ./internal/domain/rate/... -v     # Specific package
go test -race ./...                        # With race detector
```

### Docker Commands

```bash
# Full stack
docker-compose up -d
docker-compose logs -f api

# Database only (for local development)
docker-compose up -d postgres redis
```

## Key Implementation Patterns

### Adding a New Query Handler (CQRS Read)

1. Define query struct in `internal/application/query/`
2. Implement handler with caching support:
   ```go
   type GetFooQuery struct { /* fields */ }

   type GetFooHandler struct {
       repo rate.Repository
       cache redis.CacheInterface
       logger *slog.Logger
   }

   func (h *GetFooHandler) Handle(ctx context.Context, q GetFooQuery) (*dto.Response, error) {
       // Check cache → Query repo → Cache result → Return
   }
   ```
3. Wire in `cmd/api/main.go` and register in HTTP handler

### Adding a New Command Handler (CQRS Write)

1. Define command struct in `internal/application/command/`
2. Implement handler:
   ```go
   type DoSomethingCommand struct { /* fields */ }

   type DoSomethingHandler struct {
       repo rate.Repository
       provider provider.Interface
       logger *slog.Logger
   }

   func (h *DoSomethingHandler) Handle(ctx context.Context, cmd DoSomethingCommand) error {
       // Business logic → Write to repo → Invalidate cache
   }
   ```
3. Wire in `cmd/api/main.go` or `cmd/worker/`

### Using the Repository with Streaming

```go
// Memory-efficient iteration over large datasets
for rate := range rateRepo.Stream(ctx,
    genericrepo.WithFilter("base_currency", "CNY"),
    genericrepo.WithOrderBy("effective_date DESC"),
) {
    process(rate)
    if shouldStop {
        break  // Early termination supported
    }
}
```

### Using Result Type for Error Handling

```go
// Instead of returning (T, error)
func GetRate() result.Result[*rate.Rate] {
    rate, err := repo.FindLatest(ctx, pair)
    if err != nil {
        return result.Err[*rate.Rate](err)
    }
    return result.Ok(rate)
}

// Chain operations elegantly
result := GetRate().
    Map(func(r *rate.Rate) float64 { return r.Value() }).
    UnwrapOr(0.0)
```

### Domain Entity Validation

All domain entities self-validate. Never create invalid entities:

```go
// Creating a new rate - validation happens automatically
rate, err := rate.NewRate(pair, 0.061234, time.Now(), rate.SourceUnionPay)
if err != nil {
    // Handle domain validation error
}

// Reconstituting from database (skip validation for existing data)
rate := rate.Reconstitute(id, pair, value, date, source, created, updated)
```

### Configuration Loading

Configuration is loaded from:
1. Environment variables (highest priority)
2. JSON file at `CONFIG_PATH` env var (or `./config.json`)
3. Embedded defaults (fallback)

See `.env.example` and `config.json.example` for structure.

## Database Schema

Single main table (auto-migrated via GORM):

- **rates**: Stores all exchange rates
  - `id` (UUID primary key)
  - `base_currency`, `quote_currency` (e.g., "CNY", "JPY")
  - `value` (float64 - exchange rate)
  - `effective_date` (date - when rate is effective)
  - `source` (string - "unionpay", "ecb", etc.)
  - `created_at`, `updated_at` (timestamps)

**Note**: The old dual-table design (`temp_rates` + `update_rates`) has been refactored away.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/swagger/index.html` | Swagger UI |
| GET | `/api/v1/rates/latest?pair=CNY/JPY` | Get latest rate (cached 5min) |
| GET | `/api/v1/rates/history?pair=CNY/JPY&days=7` | Get historical rates |

## Testing Strategy

- **Domain layer**: Pure unit tests (no dependencies)
- **Application layer**: Test with mocked repositories
- **Infrastructure layer**: Integration tests (use test database)

Run tests with: `make test` or `go test ./...`

## Logging

Uses `log/slog` for structured logging:
- **Production**: JSON format to stdout
- **Development**: Text format with colors
- **Levels**: debug, info, warn, error

Configure via `LOG_LEVEL` and `LOG_FORMAT` env vars.

## Caching Strategy

Redis caching in query handlers:
- **Cache key format**: `latest:{pair}` (e.g., `latest:CNY/JPY`)
- **TTL**: 5 minutes for latest rates
- **Strategy**: Cache-aside (check cache → query DB → write cache)
- **Invalidation**: Automatic by TTL, or manual in command handlers

## Common Pitfalls

1. **Never bypass domain validation**: Use `rate.NewRate()` for new entities, not `rate.Reconstitute()`
2. **Dependency direction**: Domain never imports infrastructure; use interfaces
3. **CQRS separation**: Don't put caching in commands, only in queries
4. **Stream early termination**: Always support `break` in range loops over iterators
5. **Generic repository**: Use `genericrepo.QueryOption` functions for filters/pagination

## External Provider Integration

To add a new exchange rate provider:

1. Create new client in `internal/infrastructure/provider/{name}/`
2. Implement the provider interface from `internal/domain/provider/`
3. Register in `cmd/worker/commands/fetch.go` switch statement
4. Add provider name to `rate.Source` constants

## Swagger Documentation

Swagger docs are auto-generated from code annotations:

```bash
make swagger  # Regenerates docs/
```

View at: `http://localhost:8080/swagger/index.html`

Use `@Summary`, `@Description`, `@Param`, `@Success` annotations in handlers.
