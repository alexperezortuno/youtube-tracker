package cmd

import (
	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/spf13/cobra"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "yt-tracker",
	Short: "YouTube monitoring platform",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init(logLevel)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Error on init %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&logLevel,
		"log-level",
		"info",
		"log level (debug, info, warn, error)",
	)

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(dailyCmd)
}
