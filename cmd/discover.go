package cmd

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
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

					// Configure launcher to use headless mode
					l := launcher.New().
						Headless(true).
						Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

					// Launch browser
					browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
					defer browser.MustClose()

					var results []models.Result

					for _, ch := range cfg.ChannelNames {
						page := stealth.MustPage(browser)

						result, err := discoverySvc.GetLiveVideoID(page, ch)
						if closeErr := page.Close(); closeErr != nil {
							logger.Error("error closing page: %v", closeErr)
						}

						if err != nil {
							logger.Error("error discovering streams: %v", err)
							continue
						}

						if result != nil {
							results = append(results, *result)
							logger.Info("[DISCOVER] found live stream: %s", result.VideoID)
						} else {
							logger.Info("[DISCOVER] no live stream found for %s", ch)
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

			for _, ch := range cfg.ChannelIDs {
				if discoverByAPI {
					err := discoverySvc.FindLiveStreams(ctx, ch, discoverInterval)
					if err != nil {
						logger.Error("error discovering streams: %v", err)
					}
				}
				if discoverByRSS {
					err := discoverySvc.FindLiveStreamsByRSS(ctx, ch, discoverInterval)
					if err != nil {
						logger.Error("error discovering streams: %v", err)
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
