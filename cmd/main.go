package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/daily"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/lifecycle"
	"github.com/alexperezortuno/youtube-tracker/internal/source"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
	"github.com/alexperezortuno/youtube-tracker/internal/youtube"
)

var (
	channelIDs    []string
	mu            sync.RWMutex
	enableMetrics = flag.Bool("metrics", true, "enable streaming metrics collector")
	enableDaily   = flag.Bool("daily", false, "enable daily snapshot collector")
	logLevel      = flag.String("log-level", "info", "log level: debug, info, warn, error")
)

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := setupLogger(*logLevel)
	slog.SetDefault(logger)

	if os.Getenv("LOG_LEVEL") != "" {
		logger = setupLogger(os.Getenv("LOG_LEVEL"))
		slog.SetDefault(logger)
	}

	cfg := config.Load()

	if len(cfg.YouTubeAPIKeys) == 0 {
		logger.Error("missing required config: YOUTUBE_API_KEY")
		os.Exit(1)
	}

	if cfg.PostgresURL == "" {
		logger.Error("missing required config: POSTGRES_URL")
		os.Exit(1)
	}

	watcher := source.NewChannelWatcher(cfg.ChannelFilePath)
	redisClient := cache.NewRedis(cfg.RedisAddr)

	if redisClient == nil {
		logger.Error("redis client initialization failed", "addr", cfg.RedisAddr)
		panic("redis client is nil")
	}

	lifecycleManager := lifecycle.NewManager(redisClient, 3)
	keyManager := youtube.NewKeyManager(cfg.YouTubeAPIKeys)

	store, err := storage.NewStore(cfg.PostgresURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	src := &source.StaticSource{
		Config: cfg,
	}

	channelIDs, err := src.GetChannelIDs()
	if err != nil {
		logger.Error("failed to get channel IDs", "error", err)
		os.Exit(1)
	}

	if len(channelIDs) == 0 {
		logger.Error("no channel IDs provided")
		os.Exit(1)
	}

	config.ValidateChannelIDs(cfg.ChannelIDs)
	logger.Info("configuration loaded", "channels", len(channelIDs), "redis", cfg.RedisAddr)

	discoverySvc := discovery.Discovery{
		KeyManager: keyManager,
		Redis:      redisClient,
	}

	collectorSvc := collector.NewCollector(
		keyManager,
		2,
		5,
	)

	dailyService := &daily.DailyService{
		Collector: collectorSvc,
		Store:     store,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *enableMetrics {
		go func() {
			for {
				if watcher.HasChanged() {
					newChannels := watcher.Reload()
					if len(newChannels) == 0 {
						logger.Warn("watcher ignored empty channel list")
						time.Sleep(10 * time.Second)
						continue
					}

					mu.Lock()
					channelIDs = newChannels
					mu.Unlock()

					logger.Info("channels updated", "count", len(channelIDs))
				}
				time.Sleep(10 * time.Second)
			}
		}()

		go func() {
			for {
				logger.Debug("discovery cycle started")

				mu.RLock()
				currentChannels := make([]string, len(channelIDs))
				copy(currentChannels, channelIDs)
				mu.RUnlock()

				for _, ch := range currentChannels {
					err := discoverySvc.FindLiveStreams(ctx, ch)
					if err != nil {
						logger.Error("discovery failed", "channel", ch, "error", err)
					}
				}

				logger.Info("discovery cycle completed", "channels", len(currentChannels))
				time.Sleep(50 * time.Minute)
			}
		}()

		for {
			select {
			case <-sigChan:
				logger.Info("shutdown signal received, stopping metrics loop")
				cancel()
				return
			default:
			}

			logger.Info("metrics cycle starting")

			streams, err := redisClient.GetStreams(ctx)
			if err != nil {
				logger.Error("redis get streams failed", "error", err)
				time.Sleep(3 * time.Minute)
				continue
			}

			logger.Info("active streams found", "count", len(streams))

			if len(streams) == 0 {
				logger.Info("no active streams, waiting")
				time.Sleep(3 * time.Minute)
				continue
			}

			streamsData, metrics, err := collectorSvc.Fetch(ctx, streams)
			if err != nil {
				logger.Error("collector fetch failed", "error", err)
				time.Sleep(3 * time.Minute)
				continue
			}

			if len(metrics) == 0 {
				logger.Warn("no metrics collected")
				time.Sleep(3 * time.Minute)
				continue
			}

			lifecycleManager.Process(ctx, streams, metrics)

			if err := store.SaveStreams(ctx, streamsData); err != nil {
				logger.Error("save streams failed", "error", err)
			}

			if err := store.SaveMetrics(ctx, metrics); err != nil {
				logger.Error("save metrics failed", "error", err)
			}

			logger.Info("metrics cycle completed", "metrics", len(metrics), "streams", len(streamsData))

			time.Sleep(3 * time.Minute)
		}
	}

	if *enableDaily {
		go func() {
			for {
				logger.Info("daily snapshot starting")

				videoIDs, err := store.GetAllVideoIDs(ctx)
				if err != nil {
					logger.Error("daily get video IDs failed", "error", err)
					time.Sleep(1 * time.Hour)
					continue
				}

				if len(videoIDs) == 0 {
					logger.Info("no videos to process")
					time.Sleep(1 * time.Hour)
					continue
				}

				err = dailyService.Run(ctx, videoIDs)
				if err != nil {
					logger.Error("daily service failed", "error", err)
				}

				logger.Info("daily snapshot completed", "videos", len(videoIDs))
				time.Sleep(24 * time.Hour)
			}
		}()
	}

	<-sigChan
	logger.Info("shutdown signal received")
	cancel()
}
