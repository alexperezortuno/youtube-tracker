package cmd

import (
	"context"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

var discoverInterval int

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover live streams from channels",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.Load()
		ctx := context.Background()

		keyManager := youtube.NewKeyManager(cfg.YouTubeAPIKeys)
		redisClient := cache.NewRedis(cfg.RedisAddr)

		discoverySvc := discovery.Discovery{
			KeyManager: keyManager,
			Redis:      redisClient,
		}

		for {
			logger.Info("[DISCOVER] running...")

			for _, ch := range cfg.ChannelIDs {
				err := discoverySvc.FindLiveStreams(ctx, ch)
				if err != nil {
					logger.Error("[ERROR] %v", err)
				}
			}

			time.Sleep(time.Duration(discoverInterval) * time.Minute)
		}
	},
}

func init() {
	discoverCmd.Flags().IntVar(&discoverInterval, "interval", 30, "minutes between discovery runs")
}
