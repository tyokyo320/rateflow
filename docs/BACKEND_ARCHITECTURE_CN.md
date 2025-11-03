[English](BACKEND_ARCHITECTURE.md) | [中文](BACKEND_ARCHITECTURE_CN.md)

# RateFlow 后端架构文档

## 目录

1. [架构概述](#架构概述)
2. [整洁架构层次](#整洁架构层次)
3. [核心架构模式](#核心架构模式)
4. [查询流程（读取操作）](#查询流程读取操作)
5. [命令流程（写入操作）](#命令流程写入操作)
6. [缓存策略](#缓存策略)
7. [数据库设计](#数据库设计)
8. [数据获取策略](#数据获取策略)
9. [多货币支持](#多货币支持)
10. [性能优化](#性能优化)

---

## 架构概述

RateFlow 采用**整洁架构**（Clean Architecture）和**领域驱动设计**（DDD）原则构建，使用 **CQRS**（命令查询职责分离）模式来分离读写操作。

### 核心设计原则

1. **依赖倒置**：内层永不依赖外层，所有依赖指向内部
2. **关注点分离**：每一层都有明确的职责
3. **可测试性**：领域逻辑与基础设施解耦
4. **可扩展性**：易于添加新的数据源、货币对和功能

### 技术栈

- **Go 1.25+**：使用泛型、range over func、slog 等现代特性
- **框架**：Gin（HTTP）、GORM（ORM）、Cobra（CLI）
- **数据库**：PostgreSQL 17（持久化）、Redis 8（缓存）
- **文档**：Swagger/OpenAPI
- **部署**：Docker、Kubernetes、Docker Compose

---

## 整洁架构层次

```
┌─────────────────────────────────────────────┐
│   表现层 (Presentation Layer)               │
│   - HTTP 处理器 (Gin)                       │
│   - CLI 命令 (Cobra)                        │
│   - 中间件、路由                             │
└─────────────────┬───────────────────────────┘
                  │ 依赖
┌─────────────────▼───────────────────────────┐
│   应用层 (Application Layer)                │
│   - 命令处理器 (写操作)                      │
│   - 查询处理器 (读操作 + 缓存)               │
│   - DTO (数据传输对象)                       │
└─────────────────┬───────────────────────────┘
                  │ 依赖
┌─────────────────▼───────────────────────────┐
│   领域层 (Domain Layer)                     │
│   - 实体 (Rate)                             │
│   - 值对象 (Currency, Pair)                 │
│   - 仓储接口 (Repository)                   │
│   - 提供者接口 (Provider)                   │
└─────────────────┬───────────────────────────┘
                  │ 实现
┌─────────────────▼───────────────────────────┐
│   基础设施层 (Infrastructure Layer)         │
│   - PostgreSQL (GORM)                       │
│   - Redis (缓存)                            │
│   - 外部 API (UnionPay)                     │
│   - 配置、日志                               │
└─────────────────────────────────────────────┘
```

### 各层职责

#### 1. 领域层 (`internal/domain/`)

**完全独立**，不依赖任何外部框架或库。

**文件结构：**
```
internal/domain/
├── currency/
│   ├── code.go          # 货币代码值对象 (CNY, JPY, USD...)
│   └── pair.go          # 货币对值对象 (CNY/JPY)
├── rate/
│   ├── rate.go          # Rate 聚合根实体
│   ├── repository.go    # Repository 接口定义
│   └── source.go        # 数据源常量
└── provider/
    └── interface.go     # 外部数据提供者接口
```

**核心概念：**

- **聚合根**：`Rate` - 封装汇率业务逻辑
- **值对象**：`currency.Code`, `currency.Pair` - 不可变、自验证
- **仓储接口**：`rate.Repository` - 定义数据访问契约

**示例 - Rate 实体：**
```go
// internal/domain/rate/rate.go
type Rate struct {
    id            uuid.UUID
    pair          currency.Pair
    value         float64
    effectiveDate time.Time
    source        Source
    createdAt     time.Time
    updatedAt     time.Time
}

// NewRate 创建新汇率（带验证）
func NewRate(pair currency.Pair, value float64, date time.Time, source Source) (*Rate, error) {
    if value <= 0 {
        return nil, ErrInvalidValue
    }
    // ... 更多验证
    return &Rate{
        id:            uuid.New(),
        pair:          pair,
        value:         value,
        effectiveDate: date,
        source:        source,
        createdAt:     time.Now(),
        updatedAt:     time.Now(),
    }, nil
}

// Reconstitute 从数据库重建（无验证）
func Reconstitute(id uuid.UUID, pair currency.Pair, value float64, ...) *Rate {
    return &Rate{id: id, pair: pair, value: value, ...}
}
```

#### 2. 应用层 (`internal/application/`)

**协调**业务用例的执行。

**文件结构：**
```
internal/application/
├── command/              # CQRS - 写操作
│   └── fetch_rate.go    # 获取并保存汇率
├── query/               # CQRS - 读操作
│   ├── get_latest_rate.go    # 获取最新汇率（带缓存）
│   └── get_history.go        # 获取历史汇率
└── dto/
    └── rate_response.go      # 响应数据结构
```

**CQRS 分离：**

- **命令（Command）**：修改状态，无返回值（或仅返回 ID）
- **查询（Query）**：读取数据，从不修改状态

#### 3. 基础设施层 (`internal/infrastructure/`)

**实现**领域层定义的接口。

**文件结构：**
```
internal/infrastructure/
├── config/              # 配置加载（env + JSON）
├── logger/              # slog 封装
├── persistence/
│   ├── postgres/        # GORM 仓储实现
│   └── redis/           # 缓存实现
└── provider/
    └── unionpay/        # UnionPay API 客户端
```

#### 4. 表现层 (`internal/presentation/`)

**暴露** API 和 CLI 接口。

**文件结构：**
```
internal/presentation/
└── http/
    ├── router.go        # Gin 路由配置
    ├── middleware/      # 日志、CORS、恢复
    └── handler/
        └── rate.go      # HTTP 处理器
```

---

## 核心架构模式

### 1. CQRS（命令查询职责分离）

#### 为什么使用 CQRS？

- **读写分离**：查询可以使用缓存，命令直接写数据库
- **性能优化**：读操作占 90%+，通过缓存大幅提升性能
- **可扩展性**：未来可以使用读写分离的数据库

#### 查询处理器示例

```go
// internal/application/query/get_latest_rate.go
type GetLatestRateQuery struct {
    Pair currency.Pair
}

type GetLatestRateHandler struct {
    rateRepo rate.Repository
    cache    redis.CacheInterface
    logger   *slog.Logger
}

func (h *GetLatestRateHandler) Handle(ctx context.Context, q GetLatestRateQuery) (*dto.RateResponse, error) {
    // 1. 尝试从缓存读取
    cacheKey := fmt.Sprintf("latest:%s", q.Pair.String()) // "latest:CNY/JPY"
    var cached dto.RateResponse

    if err := h.cache.Get(ctx, cacheKey, &cached); err == nil {
        h.logger.Debug("cache hit", "key", cacheKey)
        return &cached, nil
    }

    // 2. 缓存未命中 - 查询数据库
    h.logger.Debug("cache miss", "key", cacheKey)
    r, err := h.rateRepo.FindLatest(ctx, q.Pair)
    if err != nil {
        return nil, fmt.Errorf("find latest rate: %w", err)
    }

    // 3. 转换为 DTO
    result := &dto.RateResponse{
        ID:            r.ID().String(),
        BaseCurrency:  r.Pair().Base().String(),
        QuoteCurrency: r.Pair().Quote().String(),
        Value:         r.Value(),
        EffectiveDate: r.EffectiveDate().Format("2006-01-02"),
        Source:        string(r.Source()),
        CreatedAt:     r.CreatedAt(),
        UpdatedAt:     r.UpdatedAt(),
    }

    // 4. 写入缓存（TTL: 5 分钟）
    if err := h.cache.Set(ctx, cacheKey, result, 5*time.Minute); err != nil {
        h.logger.Warn("failed to cache result", "error", err)
        // 不返回错误 - 缓存失败不应影响请求
    }

    return result, nil
}
```

#### 命令处理器示例

```go
// internal/application/command/fetch_rate.go
type FetchRateCommand struct {
    Pair currency.Pair
    Date time.Time
}

type FetchRateHandler struct {
    rateRepo rate.Repository
    provider provider.Interface
    cache    redis.CacheInterface
    logger   *slog.Logger
}

func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
    // 1. 检查是否已存在（防止重复）
    exists, err := h.rateRepo.ExistsByPairAndDate(ctx, cmd.Pair, cmd.Date)
    if err != nil {
        return fmt.Errorf("check existence: %w", err)
    }
    if exists {
        h.logger.Info("rate already exists, skipping",
            "pair", cmd.Pair.String(),
            "date", cmd.Date.Format("2006-01-02"))
        return nil
    }

    // 2. 从外部提供者获取汇率
    h.logger.Info("fetching rate from provider",
        "pair", cmd.Pair.String(),
        "date", cmd.Date.Format("2006-01-02"))

    rateValue, err := h.provider.FetchRate(ctx, cmd.Pair, cmd.Date)
    if err != nil {
        return fmt.Errorf("fetch rate from provider: %w", err)
    }

    // 3. 创建领域实体（带验证）
    r, err := rate.NewRate(cmd.Pair, rateValue, cmd.Date, rate.Source(h.provider.Name()))
    if err != nil {
        return fmt.Errorf("create rate entity: %w", err)
    }

    // 4. 保存到仓储
    if err := h.rateRepo.Create(ctx, r); err != nil {
        return fmt.Errorf("save rate: %w", err)
    }

    h.logger.Info("rate saved successfully",
        "id", r.ID().String(),
        "pair", cmd.Pair.String(),
        "value", rateValue)

    // 5. 使此货币对的缓存失效
    cacheKey := fmt.Sprintf("latest:%s", cmd.Pair.String())
    if err := h.cache.Delete(ctx, cacheKey); err != nil {
        h.logger.Warn("failed to invalidate cache", "error", err, "key", cacheKey)
        // 不返回错误 - 缓存失效失败不应导致命令失败
    }

    return nil
}
```

### 2. 仓储模式

#### 仓储接口（领域层）

```go
// internal/domain/rate/repository.go
type Repository interface {
    // 创建
    Create(ctx context.Context, rate *Rate) error

    // 查询
    FindByID(ctx context.Context, id uuid.UUID) (*Rate, error)
    FindLatest(ctx context.Context, pair currency.Pair) (*Rate, error)
    FindByPairAndDateRange(ctx context.Context, pair currency.Pair, start, end time.Time) ([]*Rate, error)

    // 检查
    ExistsByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (bool, error)

    // 流式查询（Go 1.23+ iter.Seq）
    Stream(ctx context.Context, opts ...QueryOption) iter.Seq[*Rate]
}
```

#### 仓储实现（基础设施层）

```go
// internal/infrastructure/persistence/postgres/rate_repository.go
type RateRepository struct {
    *genericrepo.Repository[*rate.Rate, models.Rate]
}

func NewRateRepository(db *gorm.DB) *RateRepository {
    return &RateRepository{
        Repository: genericrepo.New[*rate.Rate, models.Rate](
            db,
            toModel,    // 领域实体 → GORM 模型
            toDomain,   // GORM 模型 → 领域实体
        ),
    }
}

func (r *RateRepository) FindLatest(ctx context.Context, pair currency.Pair) (*rate.Rate, error) {
    result, err := r.FindOne(ctx,
        genericrepo.WithFilter("base_currency", pair.Base().String()),
        genericrepo.WithFilter("quote_currency", pair.Quote().String()),
        genericrepo.WithOrderBy("effective_date DESC"),
    )
    if err != nil {
        return nil, err
    }
    return result, nil
}

func (r *RateRepository) ExistsByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (bool, error) {
    count, err := r.Count(ctx,
        genericrepo.WithFilter("base_currency", pair.Base().String()),
        genericrepo.WithFilter("quote_currency", pair.Quote().String()),
        genericrepo.WithFilter("effective_date", date.Format("2006-01-02")),
    )
    return count > 0, err
}
```

### 3. 泛型仓储

**Go 1.25 泛型**实现类型安全的基础仓储。

```go
// pkg/genericrepo/repository.go
type Repository[T any, M any] struct {
    db       *gorm.DB
    toModel  func(T) M      // 领域 → 数据库
    toDomain func(M) T      // 数据库 → 领域
}

// 通用查询
func (r *Repository[T, M]) FindOne(ctx context.Context, opts ...QueryOption) (T, error)
func (r *Repository[T, M]) FindAll(ctx context.Context, opts ...QueryOption) ([]T, error)
func (r *Repository[T, M]) Count(ctx context.Context, opts ...QueryOption) (int64, error)

// 流式查询（内存高效）
func (r *Repository[T, M]) Stream(ctx context.Context, opts ...QueryOption) iter.Seq[T]
```

**使用示例：**
```go
// 内存高效地处理大量数据
for rate := range rateRepo.Stream(ctx,
    genericrepo.WithFilter("base_currency", "CNY"),
    genericrepo.WithOrderBy("effective_date DESC"),
    genericrepo.WithLimit(1000),
) {
    process(rate)
    if shouldStop {
        break // 支持提前终止
    }
}
```

---

## 查询流程（读取操作）

### 典型场景：获取最新汇率

```
用户请求: GET /api/v1/rates/latest?pair=CNY/JPY

     │
     ▼
┌─────────────────────┐
│  HTTP 处理器        │  1. 解析请求参数
│  (Gin Handler)      │  2. 验证货币对
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  查询处理器         │  3. 构建 GetLatestRateQuery
│  (Query Handler)    │  4. 检查 Redis 缓存
└──────────┬──────────┘
           │
           ▼
    ┌──────────────┐
    │ 缓存命中？   │
    └──────┬───────┘
           │
      YES  │  NO
    ┌──────▼──────┐      ┌──────────────────┐
    │ 返回缓存数据 │      │ 查询 PostgreSQL   │
    └─────────────┘      │ FindLatest()      │
                         └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 写入 Redis 缓存   │
                         │ TTL: 5 分钟       │
                         └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 返回数据          │
                         └──────────────────┘
```

### 代码路径

1. **HTTP 处理器** (`internal/presentation/http/handler/rate.go:GetLatest`)
2. **查询处理器** (`internal/application/query/get_latest_rate.go:Handle`)
3. **缓存检查** (`internal/infrastructure/persistence/redis/cache.go:Get`)
4. **仓储查询** (`internal/infrastructure/persistence/postgres/rate_repository.go:FindLatest`)
5. **缓存写入** (`internal/infrastructure/persistence/redis/cache.go:Set`)

### 性能特征

- **缓存命中**：~1-2ms（Redis 响应时间）
- **缓存未命中**：~10-50ms（PostgreSQL 查询 + Redis 写入）
- **预期缓存命中率**：>90%（因为最新汇率是最常查询的）

---

## 命令流程（写入操作）

### 典型场景：从 UnionPay 获取汇率

```
CronJob: 每小时执行一次
命令: rateflow-worker fetch --pair CNY/JPY

     │
     ▼
┌─────────────────────┐
│  Cobra CLI 命令     │  1. 解析命令行参数
│  (cmd/worker)       │  2. 构建 FetchRateCommand
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  命令处理器         │  3. 检查数据库中是否已存在
│  (Command Handler)  │     (防止重复插入)
└──────────┬──────────┘
           │
           ▼
    ┌──────────────┐
    │ 已存在？     │
    └──────┬───────┘
           │
      YES  │  NO
    ┌──────▼──────┐      ┌──────────────────┐
    │ 跳过，记录   │      │ 调用 UnionPay API │
    │ INFO 日志    │      │ FetchRate()       │
    └─────────────┘      └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 创建 Rate 实体    │
                         │ (带领域验证)     │
                         └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 保存到 PostgreSQL │
                         │ (UPSERT 逻辑)     │
                         └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 删除 Redis 缓存   │
                         │ latest:CNY/JPY    │
                         └──────────┬─────────┘
                                    │
                                    ▼
                         ┌──────────────────┐
                         │ 记录成功日志      │
                         └──────────────────┘
```

### 代码路径

1. **CLI 命令** (`cmd/worker/commands/fetch.go`)
2. **命令处理器** (`internal/application/command/fetch_rate.go:Handle`)
3. **存在性检查** (`internal/infrastructure/persistence/postgres/rate_repository.go:ExistsByPairAndDate`)
4. **外部 API 调用** (`internal/infrastructure/provider/unionpay/client.go:FetchRate`)
5. **领域实体创建** (`internal/domain/rate/rate.go:NewRate`)
6. **仓储保存** (`internal/infrastructure/persistence/postgres/rate_repository.go:Create`)
7. **缓存失效** (`internal/infrastructure/persistence/redis/cache.go:Delete`)

### 幂等性保证

命令可以安全地多次执行：

1. **数据库级别**：UNIQUE 约束 `(base_currency, quote_currency, effective_date)`
2. **应用级别**：`ExistsByPairAndDate` 在插入前检查
3. **结果**：同一天多次运行 CronJob 不会创建重复数据

---

## 缓存策略

### Redis 缓存设计

#### 为什么只缓存最新汇率？

| 数据类型 | 是否缓存 | 原因 |
|---------|---------|------|
| 最新汇率 | ✅ 是 | 查询频率极高（>90%请求），数据量小（每个货币对1条） |
| 历史汇率 | ❌ 否 | 查询组合太多（货币对 × 日期范围），缓存命中率低 |

#### 缓存键设计

```
模式: latest:{base}/{quote}
示例:
  - latest:CNY/JPY
  - latest:USD/JPY
  - latest:EUR/USD
```

#### 缓存配置

```go
// TTL (生存时间)
const LatestRateCacheTTL = 5 * time.Minute

// 为什么是 5 分钟？
// 1. UnionPay 每天只更新一次（约晚上 6 点）
// 2. 5 分钟足够减少数据库负载
// 3. 足够短以保证数据新鲜度
```

#### 缓存实现

```go
// internal/infrastructure/persistence/redis/cache.go
type Cache struct {
    client *redis.Client
    logger *slog.Logger
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return ErrCacheMiss
    }
    if err != nil {
        return fmt.Errorf("redis get: %w", err)
    }

    if err := json.Unmarshal([]byte(val), dest); err != nil {
        return fmt.Errorf("unmarshal cached value: %w", err)
    }

    return nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("marshal value: %w", err)
    }

    if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
        return fmt.Errorf("redis set: %w", err)
    }

    return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
    if err := c.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("redis delete: %w", err)
    }
    return nil
}
```

#### 缓存失效策略

1. **时间失效（TTL）**：5 分钟后自动过期
2. **主动失效**：新数据插入时删除对应的缓存键
3. **惰性失效**：缓存未命中时重新加载

### Cache-Aside 模式

```
查询流程:
  1. 检查缓存
  2. 如果命中 → 返回
  3. 如果未命中 → 查询数据库
  4. 写入缓存
  5. 返回

更新流程:
  1. 写入数据库
  2. 删除缓存（而非更新）
  3. 下次查询时重新加载
```

**为什么删除而非更新缓存？**

- 避免竞态条件
- 更简单可靠
- 缓存未命中成本低（单次数据库查询）

---

## 数据库设计

### 表结构

#### `rates` 表

```sql
CREATE TABLE rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,      -- 'CNY'
    quote_currency VARCHAR(3) NOT NULL,     -- 'JPY'
    value DOUBLE PRECISION NOT NULL,        -- 18.5678
    effective_date DATE NOT NULL,           -- '2024-01-15'
    source VARCHAR(50) NOT NULL,            -- 'unionpay'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个货币对每天只有一条汇率
    CONSTRAINT uq_rates_pair_date UNIQUE (base_currency, quote_currency, effective_date)
);

-- 索引：优化最常见的查询（最新汇率）
CREATE INDEX idx_rates_pair_date
ON rates (base_currency, quote_currency, effective_date DESC);
```

### GORM 模型

```go
// internal/infrastructure/persistence/postgres/models/rate.go
type Rate struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    BaseCurrency  string    `gorm:"type:varchar(3);not null;index:idx_rates_pair_date,priority:1"`
    QuoteCurrency string    `gorm:"type:varchar(3);not null;index:idx_rates_pair_date,priority:2"`
    Value         float64   `gorm:"type:double precision;not null"`
    EffectiveDate time.Time `gorm:"type:date;not null;index:idx_rates_pair_date,priority:3"`
    Source        string    `gorm:"type:varchar(50);not null"`
    CreatedAt     time.Time `gorm:"not null;default:NOW()"`
    UpdatedAt     time.Time `gorm:"not null;default:NOW()"`
}

func (Rate) TableName() string {
    return "rates"
}
```

### 单表设计的原因

**为什么不用 `temp_rates` + `update_rates` 双表？**

旧设计存在的问题：
1. **复杂性高**：需要在两个表之间同步数据
2. **查询低效**：需要 JOIN 或多次查询
3. **维护困难**：两个表的一致性难以保证
4. **无必要**：单表 + UPSERT 可以满足所有需求

**单表优势：**
1. **简单**：一个表，一个 UNIQUE 约束
2. **高效**：单次查询即可获取最新汇率
3. **可靠**：数据库保证唯一性
4. **易维护**：迁移、备份、扩展都更简单

### UPSERT 逻辑

```go
// internal/infrastructure/persistence/postgres/rate_repository.go
func (r *RateRepository) Create(ctx context.Context, rate *rate.Rate) error {
    model := toModel(rate)

    // PostgreSQL UPSERT: INSERT ... ON CONFLICT ... DO UPDATE
    result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
        Columns: []clause.Column{
            {Name: "base_currency"},
            {Name: "quote_currency"},
            {Name: "effective_date"},
        },
        DoUpdates: clause.AssignmentColumns([]string{"value", "source", "updated_at"}),
    }).Create(&model)

    if result.Error != nil {
        return fmt.Errorf("insert or update rate: %w", result.Error)
    }

    return nil
}
```

**行为：**
- **新数据**：插入新行
- **重复日期**：更新现有行的 `value` 和 `source`
- **并发安全**：UNIQUE 约束保证原子性

---

## 数据获取策略

### UnionPay 特性

#### 更新时间

- **常规**：每天约 18:00 (UTC+8) 更新当日汇率
- **周末/节假日**：可能不更新，或复制前一天数据
- **特殊情况**：重大事件可能导致一天内多次更新

#### CronJob 配置

```yaml
# deploy/k8s/worker/worker-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rateflow-fetch-cny-jpy
spec:
  schedule: "0 * * * *"  # 每小时整点执行
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            image: tyokyo320/rateflow-worker:latest
            args: ["fetch", "--pair", "CNY/JPY"]
```

**为什么每小时执行？**

1. **及时性**：确保在 UnionPay 更新后 1 小时内获取到数据
2. **容错性**：如果某次失败，下一小时会重试
3. **成本低**：幂等操作，重复执行无副作用

### 重复处理机制

#### 场景 1：同一天多次运行

```
2024-01-15 12:00 - CronJob 运行
  → 检查数据库：CNY/JPY 2024-01-15 不存在
  → 调用 UnionPay API
  → 插入：value=18.5678

2024-01-15 13:00 - CronJob 再次运行
  → 检查数据库：CNY/JPY 2024-01-15 已存在
  → 跳过，记录日志："rate already exists, skipping"
  → 不调用 API，不修改数据库
```

#### 场景 2：UnionPay 当天更新汇率

```
2024-01-15 14:00 - UnionPay 发布初始汇率 18.5678
  → CronJob 运行
  → 插入：value=18.5678

2024-01-15 20:00 - UnionPay 更新汇率为 18.5912
  → CronJob 运行
  → 检查：数据已存在
  → 跳过 ❌ (当前逻辑不会更新)
```

**当前限制：** 如果 UnionPay 同一天更新多次，系统只保留第一次的值。

**改进方案（如果需要）：**
```go
// 修改命令处理器，始终获取并更新
func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
    // 总是调用 API
    rateValue, err := h.provider.FetchRate(ctx, cmd.Pair, cmd.Date)

    // 使用 UPSERT（如果值不同则更新）
    r, _ := rate.NewRate(cmd.Pair, rateValue, cmd.Date, rate.Source(h.provider.Name()))
    return h.rateRepo.Create(ctx, r) // UPSERT 逻辑
}
```

#### 场景 3：周末/节假日

```
2024-01-13 (周六) - UnionPay 无更新
  → CronJob 运行
  → API 返回：无数据或错误
  → 记录日志，不插入数据库
  → 最新汇率仍为 2024-01-12 的数据

2024-01-15 (周一) - UnionPay 恢复更新
  → CronJob 运行
  → API 返回：新汇率
  → 插入 2024-01-15 的数据
```

**行为：** 数据库中可能出现日期跳跃（周五 → 周一），这是正常的。

### 错误处理

```go
// internal/application/command/fetch_rate.go
func (h *FetchRateHandler) Handle(ctx context.Context, cmd FetchRateCommand) error {
    // 1. 检查是否存在
    exists, err := h.rateRepo.ExistsByPairAndDate(ctx, cmd.Pair, cmd.Date)
    if err != nil {
        // 数据库错误 - 返回错误，CronJob 标记失败
        return fmt.Errorf("check existence: %w", err)
    }
    if exists {
        // 数据已存在 - 正常情况，返回 nil（成功）
        h.logger.Info("rate already exists, skipping")
        return nil
    }

    // 2. 调用外部 API
    rateValue, err := h.provider.FetchRate(ctx, cmd.Pair, cmd.Date)
    if err != nil {
        // API 错误（网络、限流、无数据等）
        // 记录错误，返回错误，CronJob 标记失败
        return fmt.Errorf("fetch rate from provider: %w", err)
    }

    // 3. 保存
    r, _ := rate.NewRate(cmd.Pair, rateValue, cmd.Date, rate.Source(h.provider.Name()))
    if err := h.rateRepo.Create(ctx, r); err != nil {
        // 数据库写入错误 - 返回错误
        return fmt.Errorf("save rate: %w", err)
    }

    // 4. 缓存失效（非关键，失败不影响结果）
    cacheKey := fmt.Sprintf("latest:%s", cmd.Pair.String())
    _ = h.cache.Delete(ctx, cacheKey) // 忽略缓存失效错误

    return nil
}
```

**CronJob 行为：**
- **成功** (`return nil`)：Job 标记为成功
- **失败** (`return error`)：Job 标记为失败，根据 `backoffLimit` 重试

---

## 多货币支持

### 支持的货币

```go
// internal/domain/currency/code.go
const (
    CNY Code = "CNY" // 人民币
    JPY Code = "JPY" // 日元
    USD Code = "USD" // 美元
    EUR Code = "EUR" // 欧元
    GBP Code = "GBP" // 英镑
    HKD Code = "HKD" // 港币
    KRW Code = "KRW" // 韩元
    SGD Code = "SGD" // 新加坡元
)
```

### 常见货币对

系统支持任意货币组合，常见的包括：

| 货币对 | 描述 | UnionPay 支持 |
|-------|------|--------------|
| CNY/JPY | 人民币对日元 | ✅ |
| USD/JPY | 美元对日元 | ✅ |
| EUR/JPY | 欧元对日元 | ✅ |
| USD/CNY | 美元对人民币 | ✅ |
| EUR/USD | 欧元对美元 | ✅ |
| GBP/USD | 英镑对美元 | ✅ |

### 添加新货币对

#### 1. 确保货币代码存在

如果新货币未在 `currency.Code` 中定义，添加它：

```go
// internal/domain/currency/code.go
const (
    // ... 现有货币 ...
    AUD Code = "AUD" // 澳元
    CAD Code = "CAD" // 加元
)
```

#### 2. 创建新的 CronJob

```yaml
# deploy/k8s/worker/worker-cronjob-usd-jpy.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: rateflow-fetch-usd-jpy
  namespace: rateflow
  labels:
    currency-pair: USD-JPY
spec:
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            image: tyokyo320/rateflow-worker:latest
            args: ["fetch", "--pair", "USD/JPY"]
```

#### 3. 更新 Kustomize

```yaml
# deploy/k8s/worker/kustomization.yaml
resources:
  - worker-cronjob-cny-jpy.yaml
  - worker-cronjob-usd-jpy.yaml  # 新增
```

#### 4. 部署

```bash
kubectl apply -k deploy/k8s/
```

### 批量获取

Worker CLI 支持日期范围：

```bash
# 获取过去 30 天的 USD/JPY 汇率
rateflow-worker fetch \
  --pair USD/JPY \
  --start 2024-01-01 \
  --end 2024-01-30
```

---

## 性能优化

### 1. 数据库优化

#### 索引策略

```sql
-- 复合索引：支持最常见的查询模式
CREATE INDEX idx_rates_pair_date
ON rates (base_currency, quote_currency, effective_date DESC);

-- 查询示例：
SELECT * FROM rates
WHERE base_currency = 'CNY'
  AND quote_currency = 'JPY'
ORDER BY effective_date DESC
LIMIT 1;

-- 使用索引：idx_rates_pair_date
-- 执行计划：Index Scan (cost=0.15..8.17 rows=1)
```

#### 连接池配置

```go
// internal/infrastructure/persistence/postgres/connection.go
func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
    sqlDB, err := db.DB()

    // 连接池配置
    sqlDB.SetMaxOpenConns(25)           // 最大打开连接数
    sqlDB.SetMaxIdleConns(5)            // 最大空闲连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间

    return db, nil
}
```

#### 查询优化

```go
// ❌ N+1 查询问题
for _, pair := range pairs {
    rate, _ := repo.FindLatest(ctx, pair) // 每次都查询数据库
}

// ✅ 批量查询
rates := repo.FindLatestForPairs(ctx, pairs) // 单次查询
```

### 2. 缓存优化

#### 当前配置

```go
// TTL
LatestRateCacheTTL = 5 * time.Minute

// 内存使用估算
// - 单条缓存：~200 bytes (JSON)
// - 10 个货币对：~2 KB
// - 100 个货币对：~20 KB
// → Redis 内存占用极小
```

#### 预热策略（可选）

在应用启动时预加载热门货币对：

```go
// cmd/api/main.go
func warmupCache(ctx context.Context, handler *query.GetLatestRateHandler) {
    hotPairs := []currency.Pair{
        currency.MustNewPair("CNY", "JPY"),
        currency.MustNewPair("USD", "JPY"),
        currency.MustNewPair("EUR", "USD"),
    }

    for _, pair := range hotPairs {
        _, _ = handler.Handle(ctx, query.GetLatestRateQuery{Pair: pair})
    }
}
```

### 3. API 性能

#### 响应时间目标

| 端点 | 目标 | 典型值 |
|-----|------|-------|
| GET /health | <10ms | ~2ms |
| GET /api/v1/rates/latest (缓存命中) | <50ms | ~5-15ms |
| GET /api/v1/rates/latest (缓存未命中) | <200ms | ~50-100ms |
| GET /api/v1/rates/history (7 天) | <500ms | ~100-300ms |

#### 并发处理

```go
// 使用 Gin 的默认配置
router := gin.Default()

// Gin 为每个请求创建一个 goroutine
// 理论并发：受限于系统资源和数据库连接池
// 实际并发：~1000-5000 req/s（取决于硬件）
```

#### 超时控制

```go
// internal/presentation/http/middleware/timeout.go
func Timeout(duration time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// 使用
router.Use(middleware.Timeout(30 * time.Second))
```

### 4. Go 1.25+ 特性优化

#### 泛型减少内存分配

```go
// 旧方式：interface{} 需要类型断言
func FindOne(db *gorm.DB) (interface{}, error)

// 新方式：泛型，编译时类型安全
func FindOne[T any](db *gorm.DB) (T, error)
```

#### Range over Function (内存高效)

```go
// 旧方式：一次加载所有数据到内存
rates, _ := repo.FindAll(ctx) // 可能有 10 万条
for _, rate := range rates {
    process(rate) // 内存占用：10 万条 × ~200 bytes = 20 MB
}

// 新方式：流式处理，按需加载
for rate := range repo.Stream(ctx) { // 内存占用：~200 bytes (单条)
    process(rate)
}
```

#### 结构化日志 (slog)

```go
// 高性能、类型安全的日志
logger.Info("rate fetched",
    "pair", pair.String(),
    "value", rateValue,
    "date", date,
) // 零分配（zero-allocation）
```

---

## 总结

### 核心架构决策

1. **整洁架构 + DDD**：清晰的层次和关注点分离
2. **CQRS**：查询使用缓存，命令直接写数据库
3. **单表设计**：简单、高效、可靠
4. **Cache-Aside**：5 分钟 TTL，仅缓存最新汇率
5. **幂等 CronJob**：每小时运行，安全重试

### 数据流

```
UnionPay API (每天约 18:00 更新)
    ↓
CronJob (每小时运行)
    ↓
Command Handler (检查重复 → 获取 → 保存 → 使缓存失效)
    ↓
PostgreSQL (单表，UNIQUE 约束)
    ↓
Query Handler (缓存优先 → 数据库回退)
    ↓
Redis Cache (5 分钟 TTL)
    ↓
HTTP API (Gin)
    ↓
用户
```

### 关键指标

- **缓存命中率**：>90%
- **API 响应时间**：<50ms (缓存命中)
- **数据新鲜度**：最多 1 小时延迟
- **并发能力**：1000+ req/s
- **数据准确性**：100% (由数据库约束保证)

### 扩展点

1. **新货币对**：添加 CronJob YAML
2. **新数据源**：实现 `provider.Interface`
3. **历史数据缓存**：添加 Query Handler
4. **读写分离**：CQRS 已经准备好，只需配置副本数据库
5. **事件驱动**：在命令处理器中发布领域事件

---

## 相关文档

- [README.md](../README.md) - 快速开始和 API 文档
- [README_CN.md](../README_CN.md) - 中文版 README
- [CLAUDE.md](../CLAUDE.md) - Claude Code 开发指南
- [deploy/k8s/README.md](../deploy/k8s/README.md) - Kubernetes 部署指南
