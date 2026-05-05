package lifecycle

import (
	"context"
	"log/slog"

	"github.com/alexperezortuno/youtube-tracker/internal/cache"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
)

type Manager struct {
	Redis         *cache.RedisClient
	MaxDeadCycles int
	Logger        *slog.Logger
}

func NewManager(redis *cache.RedisClient, maxDead int) *Manager {
	return NewManagerWithLogger(redis, maxDead, nil)
}

func NewManagerWithLogger(redis *cache.RedisClient, maxDead int, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		Redis:         redis,
		MaxDeadCycles: maxDead,
		Logger:        logger,
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
			m.Logger.Info("removing dead stream", "video_id", videoID)
			_ = m.Redis.RemoveStream(ctx, videoID)
			_ = m.Redis.ResetDeadCounter(ctx, videoID)
		}
	}

	for _, id := range activeIDs {

		if _, exists := activeMap[id]; !exists {

			count, _ := m.Redis.IncrementDeadCounter(ctx, id)

			m.Logger.Warn("stream missing", "video_id", id, "count", count)

			if int(count) >= m.MaxDeadCycles {
				m.Logger.Info("removing missing stream", "video_id", id)
				_ = m.Redis.RemoveStream(ctx, id)
				_ = m.Redis.ResetDeadCounter(ctx, id)
			}
		}
	}
}
