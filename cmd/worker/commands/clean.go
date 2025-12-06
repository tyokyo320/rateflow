package commands

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/internal/infrastructure/persistence/postgres"
)

var (
	cleanPair   string
	cleanBefore string
	cleanAfter  string
	cleanDryRun bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean (delete) exchange rate data from database",
	Long: `Clean (delete) exchange rate data from database based on various criteria.

This is useful for removing incorrect data or data from specific date ranges.

Examples:
  # Delete all JPY/USD data (dry-run first)
  worker clean --pair JPY/USD --dry-run

  # Actually delete JPY/USD data
  worker clean --pair JPY/USD

  # Delete all data before 2024-01-01
  worker clean --before 2024-01-01

  # Delete CNY/JPY data between specific dates
  worker clean --pair CNY/JPY --after 2024-01-01 --before 2024-12-31

  # Delete all data for all pairs before 2024 (use with caution!)
  worker clean --before 2024-01-01`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVar(&cleanPair, "pair", "", "currency pair to clean (e.g., CNY/JPY), empty for all pairs")
	cleanCmd.Flags().StringVar(&cleanBefore, "before", "", "delete data before this date (YYYY-MM-DD)")
	cleanCmd.Flags().StringVar(&cleanAfter, "after", "", "delete data after this date (YYYY-MM-DD)")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "show what would be deleted without actually deleting")
}

func runClean(cmd *cobra.Command, args []string) error {
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
	log = logger.WithContext(log, "rateflow-worker", "1.5.3")

	if cleanDryRun {
		log.Warn("DRY RUN MODE - no data will be deleted")
	}

	log.Info("starting clean command",
		slog.String("pair", cleanPair),
		slog.String("before", cleanBefore),
		slog.String("after", cleanAfter),
		slog.Bool("dry_run", cleanDryRun),
	)

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

	// Build query
	query := db.Table("rates")

	// Apply filters
	if cleanPair != "" {
		// Parse pair
		parts := splitPair(cleanPair)
		if len(parts) != 2 {
			return fmt.Errorf("invalid pair format: %s", cleanPair)
		}
		query = query.Where("base_currency = ? AND quote_currency = ?", parts[0], parts[1])
		log.Info("filtering by pair", "base", parts[0], "quote", parts[1])
	}

	if cleanBefore != "" {
		beforeDate, err := time.Parse("2006-01-02", cleanBefore)
		if err != nil {
			return fmt.Errorf("invalid before date: %w", err)
		}
		query = query.Where("effective_date < ?", beforeDate)
		log.Info("filtering before date", "date", cleanBefore)
	}

	if cleanAfter != "" {
		afterDate, err := time.Parse("2006-01-02", cleanAfter)
		if err != nil {
			return fmt.Errorf("invalid after date: %w", err)
		}
		query = query.Where("effective_date > ?", afterDate)
		log.Info("filtering after date", "date", cleanAfter)
	}

	// Count affected rows first
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("count rows: %w", err)
	}

	log.Info("found rows to delete", "count", count)

	if count == 0 {
		log.Info("no rows to delete")
		return nil
	}

	if cleanDryRun {
		log.Warn("DRY RUN - would delete rows", "count", count)
		return nil
	}

	// Confirm deletion
	fmt.Printf("\n⚠️  WARNING: About to delete %d rows from database!\n", count)
	fmt.Printf("Filters:\n")
	if cleanPair != "" {
		fmt.Printf("  - Pair: %s\n", cleanPair)
	}
	if cleanBefore != "" {
		fmt.Printf("  - Before: %s\n", cleanBefore)
	}
	if cleanAfter != "" {
		fmt.Printf("  - After: %s\n", cleanAfter)
	}
	fmt.Printf("\nType 'yes' to confirm deletion: ")

	var confirmation string
	if _, err := fmt.Scanln(&confirmation); err != nil {
		log.Warn("failed to read confirmation", "error", err)
		return fmt.Errorf("deletion cancelled: %w", err)
	}

	if confirmation != "yes" {
		log.Info("deletion cancelled by user")
		return fmt.Errorf("deletion cancelled")
	}

	// Delete rows
	result := query.Delete(&postgres.RateModel{})
	if result.Error != nil {
		return fmt.Errorf("delete rows: %w", result.Error)
	}

	log.Info("rows deleted successfully", "count", result.RowsAffected)

	return nil
}

func splitPair(pair string) []string {
	// Try splitting by /
	if parts := splitString(pair, "/"); len(parts) == 2 {
		return parts
	}
	// Try splitting by -
	if parts := splitString(pair, "-"); len(parts) == 2 {
		return parts
	}
	return nil
}

func splitString(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
