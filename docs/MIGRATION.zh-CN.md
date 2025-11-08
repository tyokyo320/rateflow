# v1.4.0 迁移指南

[English](../MIGRATION_GUIDE.md) | **简体中文**

## ⚠️ 重要提示:修复汇率数据错误

v1.4.0 版本修复了**关键的 bug**,该 bug 导致银联 API 响应被错误解析。**数据库中所有已存在的数据可能都是错误的**,需要重新获取。

### 🐛 问题说明

旧版本对银联 API 格式的理解有误,导致:
- ❌ **某些货币对的汇率被倒置** (例如:CNY/JPY 显示 0.046 而非 21.7)
- ❌ **许多有效货币对返回 404 错误** (例如:JPY/USD、USD/JPY)
- ❌ **数据不一致** 不同货币对的计算结果不一致

### 🔍 技术细节

**银联 API 响应格式**:
```json
{
  "transCur": "USD",
  "baseCur": "JPY",
  "rateData": 154.79
}
```

这个格式的含义是: **154.79 JPY = 1 USD** (即 1 美元兑换 154.79 日元)

**旧代码的问题**:
- 只检查一种匹配模式 (`transCur=BASE, baseCur=QUOTE`)
- 当查询 JPY/USD 时,找不到匹配而返回 404
- 当查询 CNY/JPY 时,匹配到倒置的数据,结果错误

**新代码的修复**:
- ✅ 尝试两种匹配模式
- ✅ 模式 1: `transCur=BASE, baseCur=QUOTE` → 直接使用 `rateData`
- ✅ 模式 2: `transCur=QUOTE, baseCur=BASE` → 使用 `1/rateData`
- ✅ 所有货币对现在都能正确工作

### 📊 影响范围

如果数据库中有以下数据,**必须清理并重新获取**:
- ❌ 所有 v1.3.1 及更早版本获取的数据
- ❌ 特别是 JPY/USD、USD/JPY 等货币对
- ❌ 显示异常汇率的货币对 (如 CNY/JPY = 0.046)

**正确的汇率参考值** (2024年11月):
- 1 CNY ≈ 21.7 JPY ✅ (不是 0.046 ❌)
- 1 JPY ≈ 0.0065 USD ✅ (不是 404 ❌)
- 1 USD ≈ 154 JPY ✅
- 1 USD ≈ 7.17 CNY ✅

## 📋 迁移步骤

### 第 1 步:检查当前数据

```bash
# 检查数据库中有哪些货币对
docker-compose exec postgres psql -U rateflow -d rateflow -c \
  "SELECT base_currency, quote_currency, COUNT(*), MIN(effective_date), MAX(effective_date)
   FROM exchange_rates
   GROUP BY base_currency, quote_currency
   ORDER BY count DESC;"
```

### 第 2 步:清理旧数据

#### 方案 A: 清理特定货币对

```bash
# 先预览要删除的内容(干运行)
docker-compose exec api /app/rateflow-worker clean --pair JPY/USD --dry-run

# 确认无误后执行删除
docker-compose exec api /app/rateflow-worker clean --pair JPY/USD

# 清理其他可能错误的货币对
docker-compose exec api /app/rateflow-worker clean --pair CNY/JPY
docker-compose exec api /app/rateflow-worker clean --pair USD/JPY
```

#### 方案 B: 清理所有数据重新开始

```bash
# ⚠️ 注意:这会删除所有数据!
# 先预览
docker-compose exec api /app/rateflow-worker clean --dry-run

# 确认后执行(会要求输入 'yes' 确认)
docker-compose exec api /app/rateflow-worker clean
```

### 第 3 步:使用修复后的代码重新获取数据

#### 使用 fetch-matrix (推荐 - 批量获取多个货币对)

```bash
# 获取 CNY、JPY、USD 之间所有组合的今日汇率(6个货币对)
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD

# 获取最近 30 天的历史数据
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --start 2024-10-08 \
  --end 2024-11-08

# 获取更多货币
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD,EUR,GBP \
  --start 2024-11-01 \
  --end 2024-11-08
```

**货币矩阵说明**:
使用 5 种货币会获取 20 个货币对 (5×4):
- CNY/JPY, CNY/USD, CNY/EUR, CNY/GBP
- JPY/CNY, JPY/USD, JPY/EUR, JPY/GBP
- USD/CNY, USD/JPY, USD/EUR, USD/GBP
- EUR/CNY, EUR/JPY, EUR/USD, EUR/GBP
- GBP/CNY, GBP/JPY, GBP/USD, GBP/EUR

#### 使用单个 fetch 命令

```bash
# 获取特定货币对的特定日期
docker-compose exec api /app/rateflow-worker fetch \
  --pair CNY/JPY \
  --date 2024-11-08

# 获取日期范围
docker-compose exec api /app/rateflow-worker fetch \
  --pair JPY/USD \
  --start 2024-11-01 \
  --end 2024-11-08
```

### 第 4 步:验证数据

```bash
# 检查 CNY/JPY 汇率(应该约为 21-22)
curl "http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY"

# 检查 JPY/USD 汇率(应该约为 0.0065)
curl "http://localhost:8080/api/v1/rates/latest?pair=JPY/USD"

# 检查 USD/JPY 汇率(应该约为 154)
curl "http://localhost:8080/api/v1/rates/latest?pair=USD/JPY"
```

**预期值** (大约,2024年11月):
- 1 CNY ≈ 21.7 JPY ✅ (不是 0.046!)
- 1 JPY ≈ 0.0065 USD ✅ (不是 404!)
- 1 USD ≈ 154 JPY ✅
- 1 USD ≈ 7.17 CNY ✅

## 🆕 新命令介绍

### fetch-matrix - 批量获取货币矩阵

批量获取指定货币列表之间的所有组合汇率。

```bash
docker-compose exec api /app/rateflow-worker fetch-matrix [flags]

标志:
  --currencies string   逗号分隔的货币列表 (默认 "CNY,JPY,USD")
  --date string         获取特定日期 (YYYY-MM-DD)
  --start string        起始日期 (YYYY-MM-DD)
  --end string          结束日期 (YYYY-MM-DD)
  --provider string     使用的提供者 (unionpay) (默认 "unionpay")
  --force               强制重新获取,覆盖已存在数据
```

**示例**:
```bash
# 获取 CNY、JPY、USD 所有组合的最新汇率
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD

# 获取历史数据
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD,EUR \
  --start 2024-11-01 \
  --end 2024-11-08

# 强制重新获取(覆盖已存在数据)
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --force
```

### clean - 清理数据

从数据库中删除汇率数据。

```bash
docker-compose exec api /app/rateflow-worker clean [flags]

标志:
  --pair string     要清理的货币对 (例如: CNY/JPY)
  --before string   删除此日期之前的数据 (YYYY-MM-DD)
  --after string    删除此日期之后的数据 (YYYY-MM-DD)
  --dry-run         显示将要删除的内容但不实际删除
```

**示例**:
```bash
# 预览将要删除的内容
docker-compose exec api /app/rateflow-worker clean \
  --pair JPY/USD \
  --dry-run

# 删除所有 JPY/USD 数据
docker-compose exec api /app/rateflow-worker clean \
  --pair JPY/USD

# 删除 2024 年之前的所有数据
docker-compose exec api /app/rateflow-worker clean \
  --before 2024-01-01

# 删除特定日期范围的 CNY/JPY 数据
docker-compose exec api /app/rateflow-worker clean \
  --pair CNY/JPY \
  --after 2024-01-01 \
  --before 2024-12-31
```

## 🐳 Docker 部署迁移

如果您使用 Docker 部署:

```bash
# 1. 重新构建镜像(包含新版本)
docker-compose build --no-cache

# 2. 清理旧数据
docker-compose run --rm api /app/rateflow-worker clean --dry-run
docker-compose run --rm api /app/rateflow-worker clean

# 3. 获取新数据
docker-compose run --rm api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD

# 4. 重启服务
docker-compose up -d
```

## ⚙️ 定时任务设置

在 cron 或计划任务中添加每日更新:

```bash
# 每天早上 9 点更新汇率
0 9 * * * docker-compose run --rm api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD,EUR,GBP
```

## ❓ 常见问题

### "rate already exists, skipping"

数据已存在于数据库中。使用 `--force` 标志或先清理:

```bash
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --force
```

### 历史日期返回 404

银联只提供 2024 年及之后的数据。更早的日期会返回 404。

### 汇率看起来仍然不对

1. 确保使用 v1.4.0 或更高版本
2. 完全清理旧数据
3. 使用新代码重新获取
4. 与外部数据源对比验证 (xe.com, Google Finance)

**验证方法**:
```bash
# 检查版本
docker-compose exec api /app/rateflow-api --version

# 查看 API 日志确认使用了新的解析逻辑
docker-compose logs api | grep "inverted"
```

### 数据不一致

如果发现数据不一致:

1. **检查数据源日期**:
```bash
docker-compose exec postgres psql -U rateflow -d rateflow -c \
  "SELECT base_currency, quote_currency, value, created_at
   FROM exchange_rates
   ORDER BY created_at DESC
   LIMIT 20;"
```

2. **清理旧数据**:
```bash
# 删除 2024-11-08 之前的数据
docker-compose exec api /app/rateflow-worker clean --before 2024-11-08
```

3. **重新获取**:
```bash
docker-compose exec api /app/rateflow-worker fetch-matrix \
  --currencies CNY,JPY,USD \
  --force
```

## 📞 需要帮助?

遇到问题请访问:
- GitHub Issues: https://github.com/tyokyo320/rateflow/issues
- 项目主页: https://github.com/tyokyo320/rateflow

## 🎯 迁移检查清单

- [ ] 备份现有数据库(如果需要)
- [ ] 更新到 v1.4.0 版本
- [ ] 检查数据库中现有的货币对
- [ ] 使用 `--dry-run` 预览清理操作
- [ ] 清理旧的错误数据
- [ ] 使用 `fetch-matrix` 重新获取数据
- [ ] 验证所有货币对的汇率正确
- [ ] 测试 API 端点返回正确数据
- [ ] 测试前端界面显示正确
- [ ] 设置定时任务自动更新
- [ ] 更新监控和告警(如果有)

---

**升级顺利!** 🚀

如有任何问题,请随时在 GitHub 上提 Issue。
