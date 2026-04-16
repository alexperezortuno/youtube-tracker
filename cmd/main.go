package main

import (
	"context"
	"log"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/collector"
	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/discovery"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
)

func main() {

	ctx := context.Background()
	cfg := config.Load()

	redis := cache.NewRedis(cfg.RedisAddr)
	store, err := storage.NewStore(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}

	discoverySvc := discovery.Discovery{
		APIKey: cfg.YouTubeAPIKey,
		Redis:  redis,
	}

	collectorSvc := collector.Collector{
		APIKey: cfg.YouTubeAPIKey,
	}

	channelIDs := []string{
		"UC_x5XG1OV2P6uZZ5FSM9Ttw", // ejemplo
	}

	// loop principal
	for {
		// 1. discovery
		for _, ch := range channelIDs {
			discoverySvc.FindLiveStreams(ctx, ch)
		}

		// 2. obtener streams activos
		streams, _ := redis.GetStreams(ctx)

		if len(streams) > 0 {
			// 3. colectar métricas
			metrics, err := collectorSvc.FetchMetrics(ctx, streams)
			if err == nil {
				// 4. guardar
				store.SaveMetrics(ctx, metrics)
			}
		}

		time.Sleep(30 * time.Second)
	}
}
