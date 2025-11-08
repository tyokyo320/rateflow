#!/bin/bash
set -e

# Configuration
DB_CONTAINER="rateflow-postgres"
DB_NAME="rateflow"
DB_USER="rateflow"

echo "Adding test data to Rateflow database..."
echo "==========================================="

# Generate data for the past 90 days
for i in {0..89}; do
    # Calculate date (90 days ago to today)
    DATE=$(date -d "$i days ago" +%Y-%m-%d 2>/dev/null || date -v-${i}d +%Y-%m-%d 2>/dev/null)

    # Generate random rates with realistic variations
    # CNY/JPY base rate around 20.50, variation ±0.30
    CNY_JPY_RATE=$(awk -v seed="$i" 'BEGIN {
        srand(seed + 1000);
        base = 20.50;
        variation = (rand() - 0.5) * 0.60;
        printf "%.6f", base + variation;
    }')

    # JPY/USD base rate around 0.0067 (inverse of ~149 USD/JPY), variation ±0.0001
    JPY_USD_RATE=$(awk -v seed="$i" 'BEGIN {
        srand(seed + 2000);
        base = 0.0067;
        variation = (rand() - 0.5) * 0.0002;
        printf "%.8f", base + variation;
    }')

    # USD/CNY base rate around 7.25, variation ±0.05
    USD_CNY_RATE=$(awk -v seed="$i" 'BEGIN {
        srand(seed + 3000);
        base = 7.25;
        variation = (rand() - 0.5) * 0.10;
        printf "%.6f", base + variation;
    }')

    # Generate timestamps
    TIMESTAMP=$(date -d "$i days ago" +"%Y-%m-%d %H:%M:%S" 2>/dev/null || date -v-${i}d +"%Y-%m-%d %H:%M:%S" 2>/dev/null)

    if [ $((i % 10)) -eq 0 ]; then
        echo "Inserting data for $DATE (day $((90-i))/90)..."
    fi

    # Insert all three currency pairs
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1 <<EOF
-- CNY/JPY
INSERT INTO exchange_rates (id, base_currency, quote_currency, value, effective_date, source, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'CNY',
    'JPY',
    $CNY_JPY_RATE,
    '$DATE',
    'unionpay',
    '$TIMESTAMP',
    '$TIMESTAMP'
)
ON CONFLICT (base_currency, quote_currency, effective_date, source) DO UPDATE
SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;

-- JPY/USD
INSERT INTO exchange_rates (id, base_currency, quote_currency, value, effective_date, source, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'JPY',
    'USD',
    $JPY_USD_RATE,
    '$DATE',
    'unionpay',
    '$TIMESTAMP',
    '$TIMESTAMP'
)
ON CONFLICT (base_currency, quote_currency, effective_date, source) DO UPDATE
SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;

-- USD/CNY
INSERT INTO exchange_rates (id, base_currency, quote_currency, value, effective_date, source, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'USD',
    'CNY',
    $USD_CNY_RATE,
    '$DATE',
    'unionpay',
    '$TIMESTAMP',
    '$TIMESTAMP'
)
ON CONFLICT (base_currency, quote_currency, effective_date, source) DO UPDATE
SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;
EOF
done

echo ""
echo "==========================================="
echo "✓ Successfully inserted 90 days of test data!"
echo ""

# Show summary
docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" <<EOF
SELECT
    base_currency || '/' || quote_currency as pair,
    COUNT(*) as count,
    MIN(effective_date) as earliest,
    MAX(effective_date) as latest
FROM exchange_rates
GROUP BY base_currency, quote_currency
ORDER BY pair;
EOF

echo ""
echo "You can now test the following currency pairs:"
echo "  - CNY/JPY and JPY/CNY"
echo "  - JPY/USD and USD/JPY"
echo "  - USD/CNY and CNY/USD"
