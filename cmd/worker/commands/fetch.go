package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tyokyo320/rateflow/internal/application/command"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/postgres"
	redisCache "github.com/tyokyo320/rateflow/internal/infrastructure/persistence/redis"
	"github.com/tyokyo320/rateflow/internal/infrastructure/provider/unionpay"
)

var (
	fetchPair      string
	fetchDate      string
	fetchStartDate string
	fetchEndDate   string
	fetchProvider  string
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch exchange rates from external providers",
	Long: `Fetch exchange rates from external providers and store them in the database.

You can fetch rates for a specific date or a date range. If no date is specified,
it will fetch the latest available rate.

Examples:
  # Fetch latest CNY/JPY rate
  worker fetch --pair CNY/JPY

  # Fetch rate for a specific date
  worker fetch --pair CNY/JPY --date 2024-01-15

  # Fetch rates for a date range
  worker fetch --pair CNY/JPY --start 2024-01-01 --end 2024-01-31

  # Use a specific provider
  worker fetch --pair CNY/JPY --provider unionpay`,
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVar(&fetchPair, "pair", "CNY/JPY", "currency pair to fetch (e.g., CNY/JPY)")
	fetchCmd.Flags().StringVar(&fetchDate, "date", "", "specific date to fetch (YYYY-MM-DD)")
	fetchCmd.Flags().StringVar(&fetchStartDate, "start", "", "start date for range fetch (YYYY-MM-DD)")
	fetchCmd.Flags().StringVar(&fetchEndDate, "end", "", "end date for range fetch (YYYY-MM-DD)")
	fetchCmd.Flags().StringVar(&fetchProvider, "provider", "unionpay", "provider to use (unionpay)")
}

func runFetch(cmd *cobra.Command, args []string) error {
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
	log = logger.WithContext(log, "rateflow-worker", "1.4.0")

	log.Info("starting fetch command",
		slog.String("pair", fetchPair),
		slog.String("provider", fetchProvider),
	)

	// Parse currency pair
	pair, err := currency.ParsePair(fetchPair)
	if err != nil {
		return fmt.Errorf("invalid currency pair: %w", err)
	}

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

	// Test Redis connection
	ctx := context.Background()
	if err := cache.Ping(ctx); err != nil {
		log.Warn("redis connection failed, continuing without cache", "error", err)
	}

	// Initialize repository
	rateRepo := postgres.NewRateRepository(db, log)

	// Initialize provider
	var provider any
	switch fetchProvider {
	case "unionpay":
		provider = unionpay.NewClient(log)
	default:
		return fmt.Errorf("unknown provider: %s", fetchProvider)
	}

	// Initialize command handler
	fetchHandler := command.NewFetchRateHandler(
		rateRepo,
		provider.(interface {
			Name() string
			FetchRate(ctx context.Context, pair currency.Pair, date time.Time) (float64, error)
			FetchLatest(ctx context.Context, pair currency.Pair) (float64, error)
			SupportedPairs() []currency.Pair
			SupportsMulti() bool
			FetchMulti(ctx context.Context, pairs []currency.Pair, date time.Time) (map[string]float64, error)
		}),
		cache,
		log,
	)

	// Determine which dates to fetch
	var dates []time.Time

	if fetchStartDate != "" && fetchEndDate != "" {
		// Fetch range
		start, err := time.Parse("2006-01-02", fetchStartDate)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}

		end, err := time.Parse("2006-01-02", fetchEndDate)
		if err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}

		if end.Before(start) {
			return fmt.Errorf("end date must be after start date")
		}

		// Generate date range
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
	} else if fetchDate != "" {
		// Fetch specific date
		date, err := time.Parse("2006-01-02", fetchDate)
		if err != nil {
			return fmt.Errorf("invalid date: %w", err)
		}
		dates = []time.Time{date}
	} else {
		// Fetch today
		dates = []time.Time{time.Now()}
	}

	// Fetch rates for all dates
	successCount := 0
	errorCount := 0

	for _, date := range dates {
		log.Info("fetching rate", "date", date.Format("2006-01-02"))

		err := fetchHandler.Handle(ctx, command.FetchRateCommand{
			Pair: pair,
			Date: date,
		})

		if err != nil {
			log.Error("failed to fetch rate",
				"date", date.Format("2006-01-02"),
				"error", err,
			)
			errorCount++
		} else {
			successCount++
		}
	}

	// Summary
	log.Info("fetch completed",
		slog.Int("total", len(dates)),
		slog.Int("success", successCount),
		slog.Int("errors", errorCount),
	)

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors", errorCount)
	}

	return nil
}
