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

	// Validación de channels
	config.ValidateChannelIDs(cfg.ChannelIDs)

	// =========================
	// INIT DEPENDENCIES
	// =========================
	redisClient := cache.NewRedis(cfg.RedisAddr)

	store, err := storage.NewStore(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	// =========================
	// SOURCE (ENV / FILE)
	// =========================
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
		2, // requests per second
		5, // workers
	)

	// =========================
	// MAIN LOOP
	// =========================
	for {

		log.Println("====================================")
		log.Println("[INFO] starting cycle")

		// -------------------------
		// 1. DISCOVERY
		// -------------------------
		for _, ch := range channelIDs {
			err := discoverySvc.FindLiveStreams(ctx, ch)
			if err != nil {
				log.Printf("[ERROR] discovery failed for channel %s: %v", ch, err)
			}
		}

		// -------------------------
		// 2. GET ACTIVE STREAMS
		// -------------------------
		streamIDs, err := redisClient.GetStreams(ctx)
		if err != nil {
			log.Printf("[ERROR] redis get streams: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if len(streamIDs) == 0 {
			log.Println("[INFO] no active streams found")
			time.Sleep(40 * time.Second)
			continue
		}

		log.Printf("[INFO] found %d active streams", len(streamIDs))

		// -------------------------
		// 3. COLLECT METRICS
		// -------------------------
		streamsData, metrics, err := collectorSvc.Fetch(ctx, streamIDs)
		if err != nil {
			log.Printf("[ERROR] collector error: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if len(metrics) == 0 {
			log.Println("[WARN] no metrics returned")
			time.Sleep(30 * time.Second)
			continue
		}

		// -------------------------
		// 4. SAVE STREAMS (DIMENSION)
		// -------------------------
		err = store.SaveStreams(ctx, streamsData)
		if err != nil {
			log.Printf("[ERROR] saving streams: %v", err)
		}

		// -------------------------
		// 5. SAVE METRICS (FACTS)
		// -------------------------
		err = store.SaveMetrics(ctx, metrics)
		if err != nil {
			log.Printf("[ERROR] saving metrics: %v", err)
		}

		log.Printf("[INFO] saved %d metrics", len(metrics))

		// -------------------------
		// 6. SLEEP
		// -------------------------
		time.Sleep(30 * time.Second)
	}
}
