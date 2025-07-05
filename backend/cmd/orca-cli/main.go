package main

import (
	"fmt"
	"os"

	"github.com/orca-ng/orca/internal/cli"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// Set up logging
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	rootCmd := &cobra.Command{
		Use:   "orca-cli",
		Short: "ORCA CLI - Admin tool for ORCA (Orchestration for CyberArk)",
		Long: `ORCA CLI is an administrative command-line interface for managing
ORCA instances, users, and configurations.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add commands
	rootCmd.AddCommand(cli.NewLoginCmd())
	rootCmd.AddCommand(cli.NewLogoutCmd())
	rootCmd.AddCommand(cli.NewStatusCmd())
	rootCmd.AddCommand(cli.NewUserCmd())
	rootCmd.AddCommand(cli.NewConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}