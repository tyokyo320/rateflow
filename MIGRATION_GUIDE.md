# Migration Guide for v1.3.1

## Critical: Fix for Incorrect Exchange Rate Data

Version 1.3.1 fixes a **critical bug** in how UnionPay API responses were interpreted. **All existing data in your database may be incorrect** and needs to be re-fetched.

### What Was Wrong?

The old code misunderstood the UnionPay API format, leading to:
- **Inverted rates** for some currency pairs
- **404 errors** for many valid pairs like JPY/USD

### Impact

If you have existing data in your database, it is likely **incorrect** and should be cleaned and re-fetched.

## Step-by-Step Migration

### 1. Check Your Current Data

```bash
# Check what pairs you have in database
docker-compose exec postgres psql -U postgres -d rateflow -c \
  "SELECT base_currency, quote_currency, COUNT(*), MIN(effective_date), MAX(effective_date) FROM rates GROUP BY base_currency, quote_currency ORDER BY count DESC;"
```

### 2. Clean Old Data

#### Option A: Clean Specific Pairs

```bash
# Clean JPY/USD data (was getting 404 errors before)
go run cmd/worker/main.go clean --pair JPY/USD --dry-run  # Check first
go run cmd/worker/main.go clean --pair JPY/USD             # Actually delete

# Clean CNY/JPY data (was inverted)
go run cmd/worker/main.go clean --pair CNY/JPY --dry-run
go run cmd/worker/main.go clean --pair CNY/JPY
```

#### Option B: Clean All Data and Start Fresh

```bash
# Delete all data (CAUTION!)
go run cmd/worker/main.go clean --dry-run                  # Check what will be deleted
go run cmd/worker/main.go clean                             # Delete everything
```

### 3. Re-fetch Data with Correct Logic

#### Using fetch-matrix (Recommended for Multiple Pairs)

```bash
# Fetch all combinations of CNY, JPY, USD for today
go run cmd/worker/main.go fetch-matrix --currencies CNY,JPY,USD

# Fetch historical data for the last 30 days
go run cmd/worker/main.go fetch-matrix \
  --currencies CNY,JPY,USD \
  --start 2024-10-08 \
  --end 2024-11-08

# Fetch more currencies
go run cmd/worker/main.go fetch-matrix \
  --currencies CNY,JPY,USD,EUR,GBP \
  --start 2024-11-01 \
  --end 2024-11-08
```

This will fetch:
- CNY/JPY, CNY/USD, CNY/EUR, CNY/GBP
- JPY/CNY, JPY/USD, JPY/EUR, JPY/GBP
- USD/CNY, USD/JPY, USD/EUR, USD/GBP
- EUR/CNY, EUR/JPY, EUR/USD, EUR/GBP
- GBP/CNY, GBP/JPY, GBP/USD, GBP/EUR

(Total: 20 pairs = 5 currencies × 4 other currencies)

#### Using Individual fetch Commands

```bash
# Fetch specific pair for specific date
go run cmd/worker/main.go fetch --pair CNY/JPY --date 2024-11-08

# Fetch date range
go run cmd/worker/main.go fetch --pair JPY/USD --start 2024-11-01 --end 2024-11-08
```

### 4. Verify the Data

```bash
# Check CNY/JPY rate (should be around 21-22)
curl "http://localhost:8080/api/v1/rates/latest?pair=CNY/JPY"

# Check JPY/USD rate (should be around 0.0065)
curl "http://localhost:8080/api/v1/rates/latest?pair=JPY/USD"
```

Expected values (approximate):
- 1 CNY ≈ 21.7 JPY (not 0.046!)
- 1 JPY ≈ 0.0065 USD (not 404!)
- 1 USD ≈ 154 JPY

## New Commands

### fetch-matrix

Fetch all combinations of specified currencies.

```bash
worker fetch-matrix [flags]

Flags:
      --currencies string   comma-separated list of currencies (default "CNY,JPY,USD")
      --date string         specific date to fetch (YYYY-MM-DD)
      --start string        start date for range fetch (YYYY-MM-DD)
      --end string          end date for range fetch (YYYY-MM-DD)
      --provider string     provider to use (unionpay) (default "unionpay")
      --force               force refetch even if data exists
```

Examples:
```bash
# Fetch latest rates for all CNY, JPY, USD combinations
worker fetch-matrix --currencies CNY,JPY,USD

# Fetch historical data
worker fetch-matrix --currencies CNY,JPY,USD,EUR --start 2024-11-01 --end 2024-11-08

# Force refetch (overwrite existing data)
worker fetch-matrix --currencies CNY,JPY,USD --force
```

### clean

Delete exchange rate data from database.

```bash
worker clean [flags]

Flags:
      --pair string     currency pair to clean (e.g., CNY/JPY)
      --before string   delete data before this date (YYYY-MM-DD)
      --after string    delete data after this date (YYYY-MM-DD)
      --dry-run         show what would be deleted without actually deleting
```

Examples:
```bash
# Preview what would be deleted
worker clean --pair JPY/USD --dry-run

# Delete all JPY/USD data
worker clean --pair JPY/USD

# Delete all data before 2024
worker clean --before 2024-01-01

# Delete CNY/JPY data in specific range
worker clean --pair CNY/JPY --after 2024-01-01 --before 2024-12-31
```

## Docker Deployment

If you're using Docker:

```bash
# Rebuild with new version
docker-compose build

# Clean old data
docker-compose run --rm api /rateflow-worker clean --dry-run
docker-compose run --rm api /rateflow-worker clean

# Fetch new data
docker-compose run --rm api /rateflow-worker fetch-matrix --currencies CNY,JPY,USD

# Restart services
docker-compose up -d
```

## Automated Daily Updates

Add to your cron or scheduled tasks:

```bash
# Fetch daily updates for your currency matrix
0 9 * * * docker-compose run --rm api /rateflow-worker fetch-matrix --currencies CNY,JPY,USD,EUR,GBP
```

## Troubleshooting

### "rate already exists, skipping"

The data exists in the database. Use `--force` flag or clean first:

```bash
worker fetch-matrix --currencies CNY,JPY,USD --force
```

### "404 for historical dates"

UnionPay only provides data from 2024 onwards. Dates before that will return 404.

### Rates still look wrong

1. Make sure you're using v1.3.1 or later
2. Clean the old data completely
3. Re-fetch with the new code
4. Compare with external sources (xe.com, Google Finance)

## Questions?

Open an issue at: https://github.com/tyokyo320/rateflow/issues
