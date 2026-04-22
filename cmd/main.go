package main

import (
	"context"
	"log"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/source"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
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

	config.ValidateChannelIDs(cfg.ChannelIDs)

	// =========================
	// INIT DEPENDENCIES
	// =========================
	redisClient := cache.NewRedis(cfg.RedisAddr)

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

	log.Printf("[INFO] loaded %d channel IDs", len(channelIDs))

	// =========================
	// SERVICES
	// =========================
	discoverySvc := discovery.Discovery{
		APIKey: cfg.YouTubeAPIKey,
		Redis:  redisClient,
	}

	collectorSvc := collector.NewCollector(
		cfg.YouTubeAPIKey,
		2,
		5,
	)

	// =========================
	// DISCOVERY WORKER (🔥 SOLO UNO)
	// =========================
	go func() {
		for {
			log.Println("[DISCOVERY] running...")

			for _, ch := range channelIDs {
				err := discoverySvc.FindLiveStreams(ctx, ch)
				if err != nil {
					log.Printf("[ERROR] discovery failed for channel %s: %v", ch, err)
				}
			}

			time.Sleep(10 * time.Minute) // 🔥 CONTROL DE COSTO
		}
	}()

	// =========================
	// METRICS LOOP
	// =========================
	for {
		log.Println("====================================")
		log.Println("[INFO] metrics cycle")

		streams, err := redisClient.GetStreams(ctx)
		if err != nil {
			log.Printf("[ERROR] redis get streams: %v", err)
			time.Sleep(40 * time.Second)
			continue
		}

		if len(streams) == 0 {
			log.Println("[INFO] no active streams found")
			time.Sleep(40 * time.Second)
			continue
		}

		log.Printf("[INFO] found %d active streams", len(streams))

		streamsData, metrics, err := collectorSvc.Fetch(ctx, streams)
		if err != nil {
			log.Printf("[ERROR] collector error: %v", err)
			time.Sleep(40 * time.Second)
			continue
		}

		if len(metrics) == 0 {
			log.Println("[WARN] no metrics returned")
			time.Sleep(40 * time.Second)
			continue
		}

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
