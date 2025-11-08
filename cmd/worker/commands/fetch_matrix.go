package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tyokyo320/rateflow/internal/application/command"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/provider"
	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/postgres"
	redisCache "github.com/tyokyo320/rateflow/internal/infrastructure/persistence/redis"
	"github.com/tyokyo320/rateflow/internal/infrastructure/provider/unionpay"
)

var (
	matrixCurrencies string
	matrixDate       string
	matrixStartDate  string
	matrixEndDate    string
	matrixProvider   string
	matrixForce      bool
)

// fetchMatrixCmd represents the fetch-matrix command
var fetchMatrixCmd = &cobra.Command{
	Use:   "fetch-matrix",
	Short: "Fetch exchange rates for all combinations of specified currencies",
	Long: `Fetch exchange rates for all combinations of specified currencies.

Given a list of currencies, this command will fetch rates for all possible pairs.
For example, with currencies CNY,JPY,USD it will fetch:
  - CNY/JPY, CNY/USD
  - JPY/CNY, JPY/USD
  - USD/CNY, USD/JPY

Examples:
  # Fetch latest rates for CNY, JPY, USD combinations
  worker fetch-matrix --currencies CNY,JPY,USD

  # Fetch for a specific date
  worker fetch-matrix --currencies CNY,JPY,USD,EUR --date 2024-11-08

  # Fetch for a date range
  worker fetch-matrix --currencies CNY,JPY,USD --start 2024-11-01 --end 2024-11-08

  # Force refetch even if data exists
  worker fetch-matrix --currencies CNY,JPY,USD --force`,
	RunE: runFetchMatrix,
}

func init() {
	rootCmd.AddCommand(fetchMatrixCmd)

	fetchMatrixCmd.Flags().StringVar(&matrixCurrencies, "currencies", "CNY,JPY,USD", "comma-separated list of currencies")
	fetchMatrixCmd.Flags().StringVar(&matrixDate, "date", "", "specific date to fetch (YYYY-MM-DD)")
	fetchMatrixCmd.Flags().StringVar(&matrixStartDate, "start", "", "start date for range fetch (YYYY-MM-DD)")
	fetchMatrixCmd.Flags().StringVar(&matrixEndDate, "end", "", "end date for range fetch (YYYY-MM-DD)")
	fetchMatrixCmd.Flags().StringVar(&matrixProvider, "provider", "unionpay", "provider to use (unionpay)")
	fetchMatrixCmd.Flags().BoolVar(&matrixForce, "force", false, "force refetch even if data exists")
}

func runFetchMatrix(cmd *cobra.Command, args []string) error {
	// Load configuration
	if configPath != "" {
		os.Setenv("CONFIG_PATH", configPath)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logger
	if verbose {
		cfg.Logger.Level = "debug"
	}
	log := logger.New(cfg.Logger)
	log = logger.WithContext(log, "rateflow-worker", "1.5.1")

	// Parse currencies
	currencyList := strings.Split(strings.ToUpper(strings.ReplaceAll(matrixCurrencies, " ", "")), ",")
	if len(currencyList) < 2 {
		return fmt.Errorf("need at least 2 currencies, got %d", len(currencyList))
	}

	log.Info("starting fetch-matrix command",
		slog.String("currencies", strings.Join(currencyList, ",")),
		slog.String("provider", matrixProvider),
		slog.Bool("force", matrixForce),
	)

	// Validate currencies
	var validCurrencies []currency.Code
	for _, cur := range currencyList {
		code, err := currency.NewCode(cur)
		if err != nil {
			log.Warn("invalid currency code, skipping", "currency", cur, "error", err)
			continue
		}
		validCurrencies = append(validCurrencies, code)
	}

	if len(validCurrencies) < 2 {
		return fmt.Errorf("need at least 2 valid currencies")
	}

	log.Info("validated currencies", "count", len(validCurrencies), "currencies", validCurrencies)

	// Generate all currency pairs (excluding same currency pairs)
	var pairs []currency.Pair
	for i, base := range validCurrencies {
		for j, quote := range validCurrencies {
			if i != j {
				pair := currency.MustNewPair(base, quote)
				pairs = append(pairs, pair)
			}
		}
	}

	log.Info("generated currency pairs", "count", len(pairs))

	// Initialize database
	db, err := postgres.NewConnection(cfg.Database, log)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	defer sqlDB.Close()

	// Initialize Redis cache
	cache := redisCache.NewCache(cfg.Redis, log)
	defer cache.Close()

	// Initialize provider
	var prov provider.Provider
	switch matrixProvider {
	case "unionpay":
		prov = unionpay.NewClient(log)
	default:
		return fmt.Errorf("unknown provider: %s", matrixProvider)
	}

	// Initialize repository and handler
	rateRepo := postgres.NewRateRepository(db, log)
	handler := command.NewFetchRateHandler(rateRepo, prov, cache, log)

	// Determine dates to fetch
	var dates []time.Time
	if matrixDate != "" {
		// Single date
		date, err := time.Parse("2006-01-02", matrixDate)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		dates = append(dates, date)
	} else if matrixStartDate != "" && matrixEndDate != "" {
		// Date range
		startDate, err := time.Parse("2006-01-02", matrixStartDate)
		if err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}
		endDate, err := time.Parse("2006-01-02", matrixEndDate)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
		if endDate.Before(startDate) {
			return fmt.Errorf("end date must be after start date")
		}

		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
	} else {
		// Default to today
		dates = append(dates, time.Now())
	}

	log.Info("fetching rates", "pairs", len(pairs), "dates", len(dates), "total_operations", len(pairs)*len(dates))

	// Fetch rates
	ctx := context.Background()
	successCount := 0
	errorCount := 0
	skippedCount := 0

	for _, date := range dates {
		for _, pair := range pairs {
			log.Info("fetching rate", "pair", pair.String(), "date", date.Format("2006-01-02"))

			cmd := command.FetchRateCommand{
				Pair: pair,
				Date: date,
			}

			if err := handler.Handle(ctx, cmd); err != nil {
				if strings.Contains(err.Error(), "already exists") && !matrixForce {
					log.Debug("rate already exists, skipping", "pair", pair.String(), "date", date.Format("2006-01-02"))
					skippedCount++
				} else {
					log.Error("failed to fetch rate", "pair", pair.String(), "date", date.Format("2006-01-02"), "error", err)
					errorCount++
				}
			} else {
				successCount++
			}
		}
	}

	log.Info("fetch-matrix completed",
		"total", len(pairs)*len(dates),
		"success", successCount,
		"errors", errorCount,
		"skipped", skippedCount,
	)

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors", errorCount)
	}

	return nil
}
