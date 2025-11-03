package commands

import (
	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "RateFlow worker CLI",
	Long: `RateFlow worker is a command-line tool for managing exchange rate data.

It provides commands to fetch rates from external providers, manage historical data,
and perform maintenance tasks on the rate database.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: use environment variables)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
