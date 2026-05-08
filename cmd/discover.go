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
var discoverByRSS bool
var discoverByAPI bool

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
				if discoverByAPI {
					err := discoverySvc.FindLiveStreams(ctx, ch, discoverInterval)
					logger.Error("%v", err)
				}
				if discoverByRSS {
					err := discoverySvc.FindLiveStreamsByRSS(ctx, ch, discoverInterval)
					logger.Error("%v", err)
				}
			}
			logger.Debug("[DISCOVER] sleeping for %d minutes", discoverInterval)
			time.Sleep(time.Duration(discoverInterval) * time.Minute)
		}
	},
}

func init() {
	discoverCmd.Flags().IntVar(&discoverInterval, "interval", 30, "minutes between discovery runs")
	discoverCmd.Flags().BoolVar(&discoverByRSS, "rss", true, "use RSS feed for discovery")
	discoverCmd.Flags().BoolVar(&discoverByAPI, "api", true, "use YouTube API for discovery")
}
