package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"harvest-cli/internal/api"
	"harvest-cli/internal/config"
)

var (
	cfg       *config.Config
	apiClient *api.Client
)

var rootCmd = &cobra.Command{
	Use:   "harvest",
	Short: "CLI tool for Harvest time tracking",
	Long: `A command-line interface for Harvest time tracking.

Log time, view entries, and manage timers directly from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for help commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		apiClient = api.NewClient(cfg)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}
