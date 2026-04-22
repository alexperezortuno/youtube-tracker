package lifecycle

import (
	"context"
	"log"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
)

type Manager struct {
	Redis         *cache.RedisClient
	MaxDeadCycles int
}

func NewManager(redis *cache.RedisClient, maxDead int) *Manager {
	return &Manager{
		Redis:         redis,
		MaxDeadCycles: maxDead,
	}
}

func (m *Manager) Process(ctx context.Context, activeIDs []string, metrics []models.Metric) {

	activeMap := make(map[string]bool)

	for _, metric := range metrics {

		videoID := metric.VideoID
		activeMap[videoID] = true

		if metric.Viewers > 0 {
			_ = m.Redis.ResetDeadCounter(ctx, videoID)
			continue
		}

		count, _ := m.Redis.IncrementDeadCounter(ctx, videoID)

		if int(count) >= m.MaxDeadCycles {
			log.Printf("[LIFECYCLE] removing dead stream (0 viewers): %s", videoID)
			_ = m.Redis.RemoveStream(ctx, videoID)
			_ = m.Redis.ResetDeadCounter(ctx, videoID)
		}
	}

	// 🔥 NUEVO: detectar streams que desaparecieron
	for _, id := range activeIDs {

		if _, exists := activeMap[id]; !exists {

			count, _ := m.Redis.IncrementDeadCounter(ctx, id)

			log.Printf("[LIFECYCLE] missing stream %s (count %d)", id, count)

			if int(count) >= m.MaxDeadCycles {
				log.Printf("[LIFECYCLE] removing missing stream: %s", id)
				_ = m.Redis.RemoveStream(ctx, id)
				_ = m.Redis.ResetDeadCounter(ctx, id)
			}
		}
	}
}
