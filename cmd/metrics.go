package cmd

import (
	"context"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

var metricsInterval int

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Collect live streaming metrics",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.Load()
		ctx := context.Background()

		keyManager := youtube.NewKeyManager(cfg.YouTubeAPIKeys)
		redisClient := cache.NewRedis(cfg.RedisAddr)
		store, _ := storage.NewStore(cfg.PostgresURL)

		collectorSvc := collector.NewCollector(keyManager, 2, 5)

		for {
			streams, _ := redisClient.GetStreams(ctx)

			if len(streams) == 0 {
				logger.Info("[METRICS] no streams")
				time.Sleep(time.Duration(metricsInterval) * time.Second)
				continue
			}

			streamsData, metrics, err := collectorSvc.Fetch(ctx, streams)
			if err != nil {
				logger.Error("[METRICS] error collecting data: %v", err)
				continue
			}

			_ = store.SaveStreams(ctx, streamsData)
			_ = store.SaveMetrics(ctx, metrics)

			logger.Info("[METRICS] saved %d metrics", len(metrics))

			time.Sleep(time.Duration(metricsInterval) * time.Second)
		}
	},
}

func init() {
	metricsCmd.Flags().IntVar(&metricsInterval, "interval", 30, "seconds between metrics collection")
}
