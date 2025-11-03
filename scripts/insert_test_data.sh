#!/bin/bash

# Insert 30 days of test data for CNY/JPY and USD/JPY

# Database connection
DB_CONTAINER="rateflow-postgres"
DB_NAME="rateflow"
DB_USER="rateflow"

# Generate data for the past 30 days
for i in {0..30}; do
    # Calculate date (30 days ago to today)
    DATE=$(date -d "$i days ago" +%Y-%m-%d)

    # Generate random rates with realistic variations
    # CNY/JPY base rate around 20.50, variation ±0.20
    CNY_JPY_RATE=$(awk -v date="$i" 'BEGIN {
        srand(date + 1);
        base = 20.50;
        variation = (rand() - 0.5) * 0.40;
        printf "%.6f", base + variation;
    }')

    # USD/JPY base rate around 149.50, variation ±1.50
    USD_JPY_RATE=$(awk -v date="$i" 'BEGIN {
        srand(date + 2);
        base = 149.50;
        variation = (rand() - 0.5) * 3.00;
        printf "%.6f", base + variation;
    }')

    # Generate timestamps
    TIMESTAMP=$(date -d "$i days ago" +"%Y-%m-%d %H:%M:%S")

    echo "Inserting data for $DATE..."

    # Insert CNY/JPY data
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" <<EOF
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
EOF

    # Insert USD/JPY data
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" <<EOF
INSERT INTO exchange_rates (id, base_currency, quote_currency, value, effective_date, source, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'USD',
    'JPY',
    $USD_JPY_RATE,
    '$DATE',
    'unionpay',
    '$TIMESTAMP',
    '$TIMESTAMP'
)
ON CONFLICT (base_currency, quote_currency, effective_date, source) DO UPDATE
SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;
EOF
done

echo "Done! Inserted 30 days of test data."
