[English](BACKEND_ARCHITECTURE.md) | [中文](BACKEND_ARCHITECTURE_CN.md)

# RateFlow Backend Architecture & Core Logic

## Table of Contents
1. [Overview](#overview)
2. [Architecture Layers](#architecture-layers)
3. [Core Logic Flow](#core-logic-flow)
4. [Caching Strategy](#caching-strategy)
5. [Data Fetching Strategy](#data-fetching-strategy)
6. [Multi-Currency Support](#multi-currency-support)

---

## Overview

RateFlow backend is built with **Domain-Driven Design (DDD)** and **CQRS** pattern using **Go 1.25+**.

### Key Design Principles

1. **Clean Architecture** - Clear separation of concerns with dependency inversion
2. **CQRS** - Separate read (Query) and write (Command) operations
3. **Domain-Driven Design** - Rich domain model with business logic encapsulated in entities
4. **Repository Pattern** - Abstract data access with interfaces
5. **Provider Pattern** - Pluggable external data sources

### Technology Stack

- **Language**: Go 1.25+ (generics, range over func, slog)
- **Database**: PostgreSQL 17 (single table design)
- **Cache**: Redis 8 (query results caching)
- **HTTP Framework**: Gin
- **ORM**: GORM with AutoMigrate
- **Logging**: structured logging with `log/slog`

---

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

The **core business logic** layer, dependency-free.

#### Key Components:

**A. Entities & Value Objects**

```go
// Currency Code (Value Object)
type Code string
const (
    CNY Code = "CNY"  // Chinese Yuan
    JPY Code = "JPY"  // Japanese Yen
    USD Code = "USD"  // US Dollar
    EUR Code = "EUR"  // Euro
    GBP Code = "GBP"  // British Pound
    HKD Code = "HKD"  // Hong Kong Dollar
    KRW Code = "KRW"  // South Korean Won
    SGD Code = "SGD"  // Singapore Dollar
)

// Currency Pair (Value Object)
type Pair struct {
    base  Code
    quote Code
}

// Rate (Aggregate Root)
type Rate struct {
    id            string
    pair          Pair
    value         float64
    effectiveDate time.Time
    source        Source
    createdAt     time.Time
    updatedAt     time.Time
}
```

**B. Repository Interface**

```go
type Repository interface {
    // Create saves a new rate
    Create(ctx context.Context, rate *Rate) error
    
    // FindLatest finds the most recent rate for a pair
    FindLatest(ctx context.Context, pair Pair) (*Rate, error)
    
    // FindByPairAndDateRange finds rates within a date range
    FindByPairAndDateRange(ctx context.Context, pair Pair, start, end time.Time) ([]*Rate, error)
    
    // ExistsByPairAndDate checks if a rate exists
    ExistsByPairAndDate(ctx context.Context, pair Pair, date time.Time) (bool, error)
    
    // Stream returns an iterator for memory-efficient querying
    Stream(ctx context.Context, opts ...genericrepo.QueryOption) iter.Seq[*Rate]
}
```

**C. Provider Interface**

```go
type Provider interface {
    // Name returns the provider identifier
    Name() string
    
    // FetchRate fetches exchange rate from external source
    FetchRate(ctx context.Context, pair Pair, date time.Time) (float64, error)
}
```

### 2. Application Layer (`internal/application/`)

Implements **use cases** with CQRS pattern.

#### A. Queries (Read Operations)

**GetLatestRateQuery** - Get most recent rate

```go
type GetLatestRateQuery struct {
    Pair currency.Pair
}

func (h *GetLatestRateHandler) Handle(ctx context.Context, query GetLatestRateQuery) (*dto.RateResponse, error) {
    // 1. Check Redis cache
    cacheKey := fmt.Sprintf("latest:%s", query.Pair.String())
    if cached, err := h.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil  // Cache HIT
    }
    
    // 2. Cache MISS - query database
    rate, err := h.rateRepo.FindLatest(ctx, query.Pair)
    
    // 3. Cache the result (TTL: 5 minutes)
    h.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    
    return result, nil
}
```

**GetHistoricalRatesQuery** - Get rates within date range

```go
type GetHistoricalRatesQuery struct {
    Pair     currency.Pair
    Days     int
    Page     int
    PageSize int
}

func (h *GetHistoricalRatesHandler) Handle(ctx context.Context, query GetHistoricalRatesQuery) (*dto.PaginatedRatesResponse, error) {
    // Calculate date range
    end := time.Now()
    start := end.AddDate(0, 0, -query.Days)
    
    // Query database (no cache for historical data - too many combinations)
    rates, err := h.rateRepo.FindByPairAndDateRange(ctx, query.Pair, start, end)
    
    // Apply pagination
    paginated := h.paginate(rates, query.Page, query.PageSize)
    
    return paginated, nil
}
```

#### B. Commands (Write Operations)

**FetchRateCommand** - Fetch and store new rate

```go
type FetchRateCommand struct {
    Pair currency.Pair
    Date time.Time
}

func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
    // 1. Check if rate already exists (avoid duplicates)
    exists, _ := h.rateRepo.ExistsByPairAndDate(ctx, cmd.Pair, cmd.Date)
    if exists {
        return nil  // Skip if already exists
    }
    
    // 2. Fetch rate from external provider
    rateValue, err := h.provider.FetchRate(ctx, cmd.Pair, cmd.Date)
    
    // 3. Create domain entity
    rate, err := rate.NewRate(cmd.Pair, rateValue, cmd.Date, rate.Source(h.provider.Name()))
    
    // 4. Save to database
    err = h.rateRepo.Create(ctx, rate)
    
    // 5. Invalidate cache (ensure fresh data on next query)
    cacheKey := fmt.Sprintf("latest:%s", cmd.Pair.String())
    h.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

### 3. Infrastructure Layer (`internal/infrastructure/`)

Implements **external dependencies**.

#### A. PostgreSQL Repository

**Single Table Design** (simplified from previous dual-table design):

```sql
CREATE TABLE rates (
    id              VARCHAR(36) PRIMARY KEY,
    base_currency   VARCHAR(3) NOT NULL,
    quote_currency  VARCHAR(3) NOT NULL,
    value           DOUBLE PRECISION NOT NULL,
    effective_date  DATE NOT NULL,
    source          VARCHAR(50) NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    
    CONSTRAINT unique_pair_date UNIQUE (base_currency, quote_currency, effective_date)
);

CREATE INDEX idx_rates_pair_date ON rates(base_currency, quote_currency, effective_date DESC);
```

**Key Features:**
- Auto-migration via GORM
- Unique constraint on (currency_pair, effective_date)
- Index optimized for latest rate queries
- Stores only ONE rate per day per pair (keeps most recent)

#### B. Redis Cache

**Cache Strategy:**
- **What to cache**: Latest rate queries only (high hit rate)
- **What NOT to cache**: Historical data (too many combinations)
- **TTL**: 5 minutes
- **Key format**: `latest:CNY/JPY`
- **Invalidation**: Deleted on new rate insert

#### C. UnionPay Provider

```go
func (c *Client) FetchRate(ctx context.Context, pair currency.Pair, date time.Time) (float64, error) {
    // UnionPay API URL format: https://m.unionpayintl.com/jfimg/YYYYMMDD.json
    dateStr := date.Format("20060102")
    url := fmt.Sprintf("%s/%s.json", baseURL, dateStr)
    
    // Fetch JSON
    data, _ := c.http.GetJSON(ctx, url)
    
    // Parse and find JPY/CNY rate
    for _, item := range resp.ExchangeRateJSON {
        if item.TransCur == "JPY" && item.BaseCur == "CNY" {
            return item.RateData, nil
        }
    }
}
```

### 4. Presentation Layer (`internal/presentation/`)

#### HTTP Handlers

```go
// GET /api/v1/rates/latest?pair=CNY/JPY
func (h *RateHandler) GetLatest(c *gin.Context) {
    pair, _ := currency.ParsePair(c.Query("pair"))
    
    query := query.GetLatestRateQuery{Pair: pair}
    result, err := h.getLatestHandler.Handle(c.Request.Context(), query)
    
    c.JSON(200, response.Success(result))
}

// GET /api/v1/rates/history?pair=CNY/JPY&days=30&page=1&page_size=10
func (h *RateHandler) GetHistory(c *gin.Context) {
    pair, _ := currency.ParsePair(c.Query("pair"))
    days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
    
    query := query.GetHistoricalRatesQuery{
        Pair:     pair,
        Days:     days,
        Page:     1,
        PageSize: 10,
    }
    result, err := h.getHistoryHandler.Handle(c.Request.Context(), query)
    
    c.JSON(200, response.Success(result))
}
```

---

## Core Logic Flow

### Query Flow (Read Path)

```
HTTP Request → Gin Handler → Query Handler → Cache → Repository → Database
                                                ↓
                                            Cache HIT
                                                ↓
                                           Return DTO
```

**Detailed Steps:**

1. **HTTP Request**: `GET /api/v1/rates/latest?pair=CNY/JPY`
2. **Handler**: Parse query parameter, create `GetLatestRateQuery`
3. **Query Handler**: 
   - Check Redis cache with key `latest:CNY/JPY`
   - **If cache HIT**: Return cached DTO (fast path ~1ms)
   - **If cache MISS**: Query database
4. **Repository**: Execute SQL `SELECT * FROM rates WHERE base_currency='CNY' AND quote_currency='JPY' ORDER BY effective_date DESC LIMIT 1`
5. **Cache Result**: Store in Redis with 5-minute TTL
6. **Return**: Convert domain entity to DTO, return JSON

### Command Flow (Write Path)

```
Worker CLI → Command Handler → Provider → Repository → Database
                                              ↓
                                      Invalidate Cache
```

**Detailed Steps:**

1. **Worker CLI**: `./rateflow-worker fetch --pair CNY/JPY --date 2025-11-03`
2. **Command Handler**:
   - Check if rate exists for this date (avoid duplicates)
   - If exists, skip
3. **Provider**: Fetch rate from UnionPay API
4. **Domain Entity**: Create `Rate` entity with validation
5. **Repository**: 
   - Insert into database
   - If duplicate (same pair+date), handle conflict (keep latest)
6. **Cache Invalidation**: Delete `latest:CNY/JPY` from Redis
7. **Return**: Log success

---

## Caching Strategy

### Why This Strategy?

**Problem**: 
- Bank/provider data updates **once per day** (around 6 PM China time)
- API queries can be **very frequent** (every page load, every chart interaction)
- Database queries are expensive

**Solution**: 
- Cache **latest rate only** (not historical data)
- **5-minute TTL** balances freshness and performance
- **Cache invalidation** on new data ensures consistency

### Cache Behavior

| Scenario | Cache Key | TTL | Behavior |
|----------|-----------|-----|----------|
| Latest rate query | `latest:CNY/JPY` | 5 min | HIT returns cached, MISS queries DB + cache |
| Historical data | (no cache) | N/A | Always query DB |
| New rate inserted | `latest:CNY/JPY` | Deleted | Force next query to be fresh |
| Rate updated | `latest:CNY/JPY` | Deleted | Same as insert |

### Cache Hit Rate Analysis

**Expected hit rate: > 90%**

- Users frequently check "current rate" (homepage, dashboard)
- Same currency pair queried multiple times in 5 minutes
- Historical charts don't use cache (acceptable - infrequent)

**Cache miss scenarios:**
- First query after 5 minutes
- First query after new rate inserted
- Different currency pair never queried before

---

## Data Fetching Strategy

### UnionPay Data Characteristics

**Update Schedule:**
- Bank updates data **once per day**
- Usually around **6:00 PM China time (UTC+8)**
- Reflects the **day's exchange rate**

**Special Cases:**

1. **Weekends & Holidays**: No new data
   - Provider may return previous day's rate
   - Or return error (404)

2. **Same-day Updates**: Rare but possible
   - If major event (currency crisis, policy change)
   - Provider may update multiple times per day

3. **Historical Data**: Available for past dates
   - Can backfill historical rates
   - Useful for charts and analysis

### Worker CronJob Strategy

**Current Setup:**
```yaml
# K8s CronJob runs every hour
schedule: "0 * * * *"
```

**Logic:**
```go
// Fetch today's rate every hour
func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
    // 1. Check if rate already exists for today
    exists, _ := h.rateRepo.ExistsByPairAndDate(ctx, cmd.Pair, time.Now())
    if exists {
        // Rate already fetched today, but we re-fetch anyway
        // in case provider updated (unlikely but handles special cases)
    }
    
    // 2. Fetch from provider
    newRate, _ := h.provider.FetchRate(ctx, cmd.Pair, time.Now())
    
    // 3. Save to database
    // Database constraint: UNIQUE (base_currency, quote_currency, effective_date)
    // If duplicate, PostgreSQL will:
    //   - Either reject (if exact same data)
    //   - Or update (if value changed - handle in repository)
}
```

**Why Run Every Hour?**

✅ **Pros:**
- Catches intra-day updates (rare but critical events)
- Resilient to temporary provider failures
- Simple logic (no complex scheduling)

✅ **Optimizations:**
- Duplicate check prevents unnecessary DB writes
- If rate unchanged, just log and skip
- Cache invalidation only if rate actually updated

### Database Deduplication

**Constraint:**
```sql
CONSTRAINT unique_pair_date UNIQUE (base_currency, quote_currency, effective_date)
```

**Behavior:**

| Scenario | Database Action |
|----------|----------------|
| First fetch for the day | INSERT new row |
| Re-fetch same rate | Skip (exists check prevents attempt) |
| Re-fetch updated rate | UPDATE existing row (if repository supports) |
| Weekend/holiday | Skip or insert previous day's rate |

**Repository Logic:**

```go
func (r *RateRepository) Create(ctx context.Context, rate *Rate) error {
    model := toModel(rate)
    
    // Use GORM's Upsert
    result := r.db.WithContext(ctx).
        Clauses(clause.OnConflict{
            Columns:   []clause.Column{{Name: "base_currency"}, {Name: "quote_currency"}, {Name: "effective_date"}},
            DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
        }).
        Create(&model)
        
    return result.Error
}
```

This handles:
- New rate → INSERT
- Updated rate → UPDATE value and updated_at
- Unchanged rate → No action (idempotent)

---

## Multi-Currency Support

### Supported Currency Pairs

**Currently Defined:**
```go
var CommonPairs = []Pair{
    CNY/JPY,  // Chinese Yuan to Japanese Yen
    USD/JPY,  // US Dollar to Japanese Yen
    EUR/JPY,  // Euro to Japanese Yen
    USD/CNY,  // US Dollar to Chinese Yuan
    EUR/USD,  // Euro to US Dollar
    GBP/USD,  // British Pound to US Dollar
}
```

### Adding New Currency Pairs

**1. Add Currency Code** (if not exists)

```go
// internal/domain/currency/code.go
const (
    AUD Code = "AUD" // Australian Dollar
)

var validCodes = map[Code]bool{
    AUD: true,  // Add to valid codes
}
```

**2. Create New Pair**

```go
pair := currency.MustNewPair(currency.AUD, currency.USD)
```

**3. Implement Provider** (if new source needed)

```go
// internal/infrastructure/provider/ecb/client.go
type ECBClient struct {}

func (c *ECBClient) FetchRate(ctx context.Context, pair currency.Pair, date time.Time) (float64, error) {
    // Fetch from European Central Bank API
}
```

**4. Register in Worker**

```go
// cmd/worker/commands/fetch.go
var provider provider.Provider
switch sourceName {
case "unionpay":
    provider = unionpay.NewClient(logger)
case "ecb":
    provider = ecb.NewClient(logger)  // Add new provider
}
```

**5. Create CronJob** (K8s)

```yaml
# New CronJob for EUR/USD
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rateflow-fetch-eur-usd
spec:
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            args: ["fetch", "--pair", "EUR/USD", "--provider", "ecb"]
```

### Frontend Integration

The React frontend already supports any currency pair:

```typescript
// web/src/types/index.ts
export type Currency = 'CNY' | 'JPY' | 'USD' | 'EUR' | 'GBP' | 'HKD' | 'KRW' | 'SGD'

// User selects from dropdown
<CurrencyPairSelector 
  baseCurrency="EUR" 
  quoteCurrency="USD" 
/>

// API call adapts automatically
GET /api/v1/rates/latest?pair=EUR/USD
```

---

## Performance Optimization

### Database Indexes

```sql
-- Optimized for latest rate queries
CREATE INDEX idx_rates_pair_date ON rates(
    base_currency, 
    quote_currency, 
    effective_date DESC
);

-- Query uses index:
SELECT * FROM rates 
WHERE base_currency = 'CNY' AND quote_currency = 'JPY'
ORDER BY effective_date DESC 
LIMIT 1;
-- → Index scan, ~1ms
```

### Connection Pooling

```go
// config.go
DB_MAX_CONNS=25
DB_MAX_IDLE=5
DB_CONN_MAX_LIFETIME=5m
```

### Redis Tuning

```go
// TTL chosen based on data update frequency
Latest rate: 5 minutes  // Good balance
Historical:  No cache   // Too many combinations
```

---

## Monitoring & Observability

### Structured Logging

```go
logger.Info("rate fetched",
    "pair", "CNY/JPY",
    "rate", 0.061234,
    "source", "unionpay",
    "cache_hit", false,
)
```

### Key Metrics to Track

- Cache hit rate (target: > 90%)
- Query latency (p50, p95, p99)
- Provider fetch success rate
- Database connection pool usage
- Rate update frequency

### Health Check

```bash
GET /health
{
  "status": "ok",
  "database": "connected",
  "redis": "connected"
}
```

---

## Summary

### Architecture Highlights

✅ **Clean Architecture** - Domain-driven, dependency-inverted
✅ **CQRS** - Separate read/write paths for optimization
✅ **Smart Caching** - 5-minute TTL for latest rates
✅ **Idempotent Writes** - Duplicate protection via DB constraints
✅ **Multi-Currency** - Extensible design for any currency pair
✅ **Production-Ready** - Structured logging, health checks, graceful shutdown

### Data Flow Summary

**Read Path (Query):**
```
API Request → Cache Check → (HIT) Return cached
                        ↓ (MISS)
                    Query DB → Cache result → Return
```

**Write Path (Command):**
```
CronJob/CLI → Fetch from Provider → Validate → Save to DB → Invalidate Cache
```

### Best Practices Implemented

1. **Domain Validation** - All business rules in entities
2. **Repository Pattern** - Data access abstraction
3. **Provider Pattern** - External service abstraction
4. **Cache-Aside** - Check cache, query DB, update cache
5. **Optimistic Concurrency** - Let DB handle duplicates
6. **Structured Logging** - Easy debugging and monitoring
7. **Health Checks** - Kubernetes readiness/liveness

---

