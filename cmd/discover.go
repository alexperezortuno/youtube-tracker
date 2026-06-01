package cmd

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

var (
	discoverInterval    int
	discoverByRSS       bool
	discoverByAPI       bool
	discoverByExtractor bool
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover live streams from channels",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.Load()
		ctx := context.Background()

		keyManager := youtube.NewKeyManager(cfg.YouTubeAPIKeys)
		redisClient := cache.NewRedis(cfg.RedisAddr)

		pool, err := pgxpool.New(ctx, cfg.PostgresURL)
		if err != nil {
			logger.Error("failed to connect to database: %v", err)
			return
		}
		defer pool.Close()

		dbSource := storage.NewDBSource(pool)

		discoverySvc := discovery.Discovery{
			KeyManager: keyManager,
			Redis:      redisClient,
		}

		for {
			logger.Info("[DISCOVER] running...")
			startTime := time.Now()

			if discoverByExtractor {
				func() {
					logger.Debug("[DISCOVER] starting extractor discovery")
					defer func() {
						if r := recover(); r != nil {
							logger.Error("extractor panic recovered: %v\n%s", r, debug.Stack())
						}
					}()

					l := launcher.New().
						Headless(true).
						NoSandbox(true).
						Set("disable-dev-shm-usage").
						Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

					browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
					defer browser.MustClose()

					channels, err := dbSource.GetChannels(true)
					if err != nil {
						logger.Error("failed to get channels from database: %v", err)
						return
					}

					var results []models.Result

					for _, ch := range channels {
						page := stealth.MustPage(browser)

						result, err := discoverySvc.GetLiveVideoID(page, ch.Name)
						if closeErr := page.Close(); closeErr != nil {
							logger.Error("error closing page: %v", closeErr)
						}

						if err != nil {
							logger.Error("error discovering streams for channel %s: %v", ch.Name, err)
							continue
						}

						if result != nil {
							result.Channel = ch.ID
							results = append(results, *result)
							logger.Info("[DISCOVER] found live stream: %s", result.VideoID)
						} else {
							logger.Info("[DISCOVER] no live stream found for %s", ch.Name)
						}
					}

					if len(results) > 0 {
						_ = discoverySvc.SaveLiveStreamsByExtractor(ctx, results, discoverInterval)
						logger.Debug("[DISCOVER] saved %d live streams discovered by extractor", len(results))
						for _, r := range results {
							logger.Debug("[DISCOVER] discovered stream: Channel: %s, VideoID: %s, URL: %s", r.Channel, r.VideoID, r.URL)
						}
					}
				}()
			}

			channelIDs, err := dbSource.GetChannelIDs()
			if err != nil {
				logger.Error("failed to get channel IDs: %v", err)
			} else {
				for _, chID := range channelIDs {
					if discoverByAPI {
						err := discoverySvc.FindLiveStreams(ctx, chID, discoverInterval)
						if err != nil {
							logger.Error("error discovering streams: %v", err)
						}
					}
					if discoverByRSS {
						err := discoverySvc.FindLiveStreamsByRSS(ctx, chID, discoverInterval)
						if err != nil {
							logger.Error("error discovering streams: %v", err)
						}
					}
				}
			}

			logger.Debug("[DISCOVER] execution time: %s", time.Since(startTime).String())
			logger.Debug("[DISCOVER] sleeping for %d minutes", discoverInterval)
			time.Sleep(time.Duration(discoverInterval) * time.Minute)
		}
	},
}

func init() {
	discoverCmd.Flags().IntVar(&discoverInterval, "interval", 30, "minutes between discovery runs")
	discoverCmd.Flags().BoolVar(&discoverByRSS, "rss", false, "use RSS feed for discovery")
	discoverCmd.Flags().BoolVar(&discoverByAPI, "api", false, "use YouTube API for discovery")
	discoverCmd.Flags().BoolVar(&discoverByExtractor, "extractor", false, "use extractor for discovery")
}
