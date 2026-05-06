package cmd

import (
	"context"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/daily"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

var dailyInterval int

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Collect daily video statistics",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.Load()
		ctx := context.Background()

		keyManager := youtube.NewKeyManager(cfg.YouTubeAPIKeys)
		store, _ := storage.NewStore(cfg.PostgresURL)

		collectorSvc := collector.NewCollector(keyManager, 1, 2)

		dailySvc := &daily.DailyService{
			Collector: collectorSvc,
			Store:     store,
		}

		videoIDs, err := store.GetAllVideoIDs(ctx)
		if err != nil {
			logger.Error("daily get video IDs failed error: %v", err)
			time.Sleep(1 * time.Hour)
			return
		}

		if len(videoIDs) == 0 {
			logger.Info("no videos to process")
			time.Sleep(1 * time.Hour)
			return
		}

		for {
			logger.Info("[DAILY] running snapshot")

			err := dailySvc.Run(
				ctx,
				videoIDs,
			)
			if err != nil {
				logger.Error("failed: %v", err)
			}

			time.Sleep(time.Duration(dailyInterval) * time.Hour)
		}
	},
}

func init() {
	dailyCmd.Flags().IntVar(&dailyInterval, "interval", 12, "hours between daily runs")
}
