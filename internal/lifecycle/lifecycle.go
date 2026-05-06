package lifecycle

import (
	"context"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
)

type Manager struct {
	Redis         *cache.RedisClient
	MaxDeadCycles int
}

func NewManager(redis *cache.RedisClient, maxDead int) *Manager {
	return NewManagerWithLogger(redis, maxDead)
}

func NewManagerWithLogger(redis *cache.RedisClient, maxDead int) *Manager {
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
			logger.Info("removing dead stream", "video_id", videoID)
			_ = m.Redis.RemoveStream(ctx, videoID)
			_ = m.Redis.ResetDeadCounter(ctx, videoID)
		}
	}

	for _, id := range activeIDs {

		if _, exists := activeMap[id]; !exists {

			count, _ := m.Redis.IncrementDeadCounter(ctx, id)

			logger.Warn("stream missing", "video_id", id, "count", count)

			if int(count) >= m.MaxDeadCycles {
				logger.Info("removing missing stream", "video_id", id)
				_ = m.Redis.RemoveStream(ctx, id)
				_ = m.Redis.ResetDeadCounter(ctx, id)
			}
		}
	}
}
