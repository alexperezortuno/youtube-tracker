package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/lifecycle"
	"github.com/alexperezortuno/youtube-tracker/internal/source"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
)

var (
	channelIDs []string
	mu         sync.RWMutex
)

func main() {

	ctx := context.Background()

	// =========================
	// LOAD CONFIG
	// =========================
	cfg := config.Load()

	if cfg.YouTubeAPIKey == "" {
		log.Fatal("missing YOUTUBE_API_KEY")
	}

	if cfg.PostgresURL == "" {
		log.Fatal("missing POSTGRES_URL")
	}

	// INIT DEPENDENCIES
	watcher := source.NewChannelWatcher(cfg.ChannelFilePath)
	redisClient := cache.NewRedis(cfg.RedisAddr)
	lifecycleManager := lifecycle.NewManager(redisClient, 3)

	store, err := storage.NewStore(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	src := &source.StaticSource{
		Config: cfg,
	}

	channelIDs, err := src.GetChannelIDs()
	if err != nil {
		log.Fatalf("error getting channel ids: %v", err)
	}

	if len(channelIDs) == 0 {
		log.Fatal("no channel IDs provided")
	}

	config.ValidateChannelIDs(cfg.ChannelIDs)
	log.Printf("[INFO] loaded %d channel IDs", len(channelIDs))

	// SERVICES
	discoverySvc := discovery.Discovery{
		APIKey: cfg.YouTubeAPIKey,
		Redis:  redisClient,
	}

	collectorSvc := collector.NewCollector(
		cfg.YouTubeAPIKey,
		2,
		5,
	)

	go func() {
		for {
			if watcher.HasChanged() {

				newChannels := watcher.Reload()

				if len(newChannels) == 0 {
					log.Println("[WATCHER] ignored empty channel list")
					time.Sleep(10 * time.Second)
					continue
				}

				mu.Lock()
				channelIDs = newChannels
				mu.Unlock()

				log.Printf("[WATCHER] updated channels: %d", len(channelIDs))
			}

			time.Sleep(10 * time.Second)
		}
	}()

	// DISCOVERY WORKER
	go func() {
		for {
			log.Println("[DISCOVERY] running...")

			mu.RLock()
			currentChannels := make([]string, len(channelIDs))
			copy(currentChannels, channelIDs)
			mu.RUnlock()

			for _, ch := range currentChannels {
				err := discoverySvc.FindLiveStreams(ctx, ch)
				if err != nil {
					log.Printf("[ERROR] discovery failed: %v", err)
				}
			}

			time.Sleep(50 * time.Minute) // puedes ajustar
		}
	}()

	// METRICS LOOP
	for {
		log.Println("====================================")
		log.Println("[INFO] metrics cycle")

		streams, err := redisClient.GetStreams(ctx)
		if err != nil {
			log.Printf("[ERROR] redis get streams: %v", err)
			time.Sleep(5 * time.Minute)
			continue
		}

		log.Printf("[INFO] found %d active streams", len(streams))

		if len(streams) == 0 {
			log.Println("[INFO] no active streams found")
			time.Sleep(5 * time.Minute)
			continue
		}

		streamsData, metrics, err := collectorSvc.Fetch(ctx, streams)
		if err != nil {
			log.Printf("[ERROR] collector error: %v", err)
			time.Sleep(5 * time.Minute)
			continue
		}

		if len(metrics) == 0 {
			log.Println("[WARN] no metrics returned")
			time.Sleep(5 * time.Minute)
			continue
		}

		lifecycleManager.Process(ctx, streams, metrics)

		if err := store.SaveStreams(ctx, streamsData); err != nil {
			log.Printf("[ERROR] saving streams: %v", err)
		}

		if err := store.SaveMetrics(ctx, metrics); err != nil {
			log.Printf("[ERROR] saving metrics: %v", err)
		}

		log.Printf("[INFO] saved %d metrics", len(metrics))

		time.Sleep(40 * time.Second)
	}
}
