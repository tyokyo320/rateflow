# Historical Data Backfill Guide

[English](#english) | [中文](#中文)

---

## English

### Overview

RateFlow's CronJob only fetches **today's exchange rate** by default. To populate historical data, you need to manually run backfill operations.

### Why Historical Data Isn't Auto-Fetched

- **Performance**: Fetching years of data on startup would be slow
- **API Rate Limits**: External providers may have rate limits
- **Flexibility**: Different deployments may need different date ranges
- **Storage**: Allows control over data volume

### CronJob Behavior

The scheduled CronJob runs hourly and fetches only the current day:

```yaml
# deploy/k8s/worker/worker-cronjob-cny-jpy.yaml
spec:
  schedule: "0 * * * *"  # Every hour
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            args: ["fetch", "--pair", "CNY/JPY"]
            # No --date flag = fetches today only
```

### Backfill Methods

#### Method 1: Command Line (Local)

**Fetch a specific date:**
```bash
go run cmd/worker/main.go fetch --pair CNY/JPY --date 2024-01-15
```

**Fetch a date range (recommended):**
```bash
go run cmd/worker/main.go fetch --pair CNY/JPY \
  --start 2024-01-01 \
  --end 2024-12-31
```

**Multiple currency pairs:**
```bash
# CNY/JPY
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31

# JPY/USD
go run cmd/worker/main.go fetch --pair JPY/USD --start 2024-01-01 --end 2024-12-31
```

#### Method 2: Docker

```bash
docker run --rm \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=rateflow \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=rateflow \
  tyokyo320/rateflow-worker:latest \
  fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31
```

#### Method 3: Kubernetes One-Time Job

Create a file `backfill-job.yaml`:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: rateflow-backfill-2024
  namespace: rateflow
spec:
  template:
    metadata:
      labels:
        app: rateflow-backfill
    spec:
      containers:
      - name: worker
        image: tyokyo320/rateflow-worker:v1.1.3
        args:
          - "fetch"
          - "--pair"
          - "CNY/JPY"
          - "--start"
          - "2024-01-01"
          - "--end"
          - "2024-12-31"
        env:
        - name: DB_HOST
          value: "rateflow-postgres"
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: rateflow-secrets
              key: postgres-user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rateflow-secrets
              key: postgres-password
        - name: DB_NAME
          value: "rateflow"
        - name: DB_SSLMODE
          value: "disable"
        - name: REDIS_HOST
          value: "rateflow-redis"
        - name: REDIS_PORT
          value: "6379"
        - name: LOG_LEVEL
          value: "info"
      restartPolicy: OnFailure
  backoffLimit: 3
```

Apply the job:
```bash
kubectl apply -f backfill-job.yaml

# Monitor progress
kubectl logs -f job/rateflow-backfill-2024 -n rateflow

# Check status
kubectl get jobs -n rateflow

# Clean up after completion
kubectl delete job rateflow-backfill-2024 -n rateflow
```

#### Method 4: Multiple Currency Pairs in K8s

For backfilling multiple pairs, create separate jobs:

```bash
# CNY/JPY
kubectl create job rateflow-backfill-cny-jpy --from=cronjob/rateflow-fetch-cny-jpy -n rateflow

# JPY/USD
kubectl create job rateflow-backfill-jpy-usd --from=cronjob/rateflow-fetch-jpy-usd -n rateflow
```

Then edit each job to add date range parameters.

### Recommended Backfill Strategy

1. **Start with recent data** (last 30-90 days)
2. **Verify data quality** before fetching more
3. **Backfill in chunks** (e.g., 3-6 months at a time)
4. **Monitor database size** as data grows

### Example: Complete Backfill for 2024

```bash
# January to March
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-03-31

# April to June
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-04-01 --end 2024-06-30

# July to September
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-07-01 --end 2024-09-30

# October to December
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-10-01 --end 2024-12-31
```

### Troubleshooting

**Issue: Duplicate entries**
- RateFlow automatically skips duplicates based on unique constraint `(base_currency, quote_currency, effective_date)`
- Safe to re-run backfill commands

**Issue: Missing dates**
- Weekends and holidays may not have data from provider
- Check provider's data availability

**Issue: Slow backfill**
- Use smaller date ranges
- Consider running multiple jobs in parallel for different currency pairs
- Monitor API rate limits

### Performance Considerations

- **Average speed**: ~100-500 dates per minute (depends on provider)
- **Database impact**: Bulk inserts are optimized with GORM
- **Network**: Fetches data sequentially to respect rate limits

---

## 中文

### 概述

RateFlow 的 CronJob 默认只获取**今天的汇率数据**。要填充历史数据，需要手动运行回填操作。

### 为什么不自动获取历史数据

- **性能考虑**: 启动时获取多年数据会很慢
- **API 速率限制**: 外部提供商可能有速率限制
- **灵活性**: 不同部署可能需要不同的日期范围
- **存储控制**: 允许控制数据量

### CronJob 行为

定时 CronJob 每小时运行一次，仅获取当天数据：

```yaml
# deploy/k8s/worker/worker-cronjob-cny-jpy.yaml
spec:
  schedule: "0 * * * *"  # 每小时
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: worker
            args: ["fetch", "--pair", "CNY/JPY"]
            # 没有 --date 参数 = 仅获取今天
```

### 回填方法

#### 方法 1: 命令行（本地）

**获取特定日期：**
```bash
go run cmd/worker/main.go fetch --pair CNY/JPY --date 2024-01-15
```

**获取日期范围（推荐）：**
```bash
go run cmd/worker/main.go fetch --pair CNY/JPY \
  --start 2024-01-01 \
  --end 2024-12-31
```

**多货币对：**
```bash
# CNY/JPY
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31

# JPY/USD
go run cmd/worker/main.go fetch --pair JPY/USD --start 2024-01-01 --end 2024-12-31
```

#### 方法 2: Docker

```bash
docker run --rm \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=rateflow \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=rateflow \
  tyokyo320/rateflow-worker:latest \
  fetch --pair CNY/JPY --start 2024-01-01 --end 2024-12-31
```

#### 方法 3: Kubernetes 一次性 Job

创建文件 `backfill-job.yaml`:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: rateflow-backfill-2024
  namespace: rateflow
spec:
  template:
    metadata:
      labels:
        app: rateflow-backfill
    spec:
      containers:
      - name: worker
        image: tyokyo320/rateflow-worker:v1.1.3
        args:
          - "fetch"
          - "--pair"
          - "CNY/JPY"
          - "--start"
          - "2024-01-01"
          - "--end"
          - "2024-12-31"
        env:
        - name: DB_HOST
          value: "rateflow-postgres"
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: rateflow-secrets
              key: postgres-user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rateflow-secrets
              key: postgres-password
        - name: DB_NAME
          value: "rateflow"
        - name: DB_SSLMODE
          value: "disable"
        - name: REDIS_HOST
          value: "rateflow-redis"
        - name: REDIS_PORT
          value: "6379"
        - name: LOG_LEVEL
          value: "info"
      restartPolicy: OnFailure
  backoffLimit: 3
```

应用 Job:
```bash
kubectl apply -f backfill-job.yaml

# 监控进度
kubectl logs -f job/rateflow-backfill-2024 -n rateflow

# 检查状态
kubectl get jobs -n rateflow

# 完成后清理
kubectl delete job rateflow-backfill-2024 -n rateflow
```

#### 方法 4: K8s 中多货币对回填

为多个货币对回填，创建单独的 Job:

```bash
# CNY/JPY
kubectl create job rateflow-backfill-cny-jpy --from=cronjob/rateflow-fetch-cny-jpy -n rateflow

# JPY/USD
kubectl create job rateflow-backfill-jpy-usd --from=cronjob/rateflow-fetch-jpy-usd -n rateflow
```

然后编辑每个 Job 添加日期范围参数。

### 推荐的回填策略

1. **从最近数据开始**（最近 30-90 天）
2. **验证数据质量**后再获取更多
3. **分块回填**（例如每次 3-6 个月）
4. **监控数据库大小**随数据增长

### 示例：完整回填 2024 年数据

```bash
# 1月到3月
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-01-01 --end 2024-03-31

# 4月到6月
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-04-01 --end 2024-06-30

# 7月到9月
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-07-01 --end 2024-09-30

# 10月到12月
go run cmd/worker/main.go fetch --pair CNY/JPY --start 2024-10-01 --end 2024-12-31
```

### 故障排除

**问题：重复条目**
- RateFlow 基于唯一约束 `(base_currency, quote_currency, effective_date)` 自动跳过重复数据
- 可以安全地重新运行回填命令

**问题：缺少日期**
- 周末和节假日可能没有提供商数据
- 检查提供商的数据可用性

**问题：回填缓慢**
- 使用较小的日期范围
- 考虑为不同货币对并行运行多个 Job
- 监控 API 速率限制

### 性能考虑

- **平均速度**: 每分钟约 100-500 个日期（取决于提供商）
- **数据库影响**: 使用 GORM 优化批量插入
- **网络**: 顺序获取数据以遵守速率限制
