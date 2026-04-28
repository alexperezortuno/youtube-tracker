package daily

import (
	"context"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
)

type Collector interface {
	FetchDaily(ctx context.Context, videoIDs []string) ([]models.Stream, []models.Metric, error)
}

type DailyService struct {
	Collector Collector
	Store     Store
}

type Store interface {
	SaveDailyStats(ctx context.Context, stats []models.VideoDailyStat) error
}

func (d *DailyService) Run(ctx context.Context, videoIDs []string) error {
	_, metrics, err := d.Collector.FetchDaily(ctx, videoIDs)
	if err != nil {
		return err
	}

	now := time.Now().Truncate(24 * time.Hour)

	var stats []models.VideoDailyStat

	for _, m := range metrics {
		stats = append(stats, models.VideoDailyStat{
			VideoID: m.VideoID,
			Date:    now,
			Views:   int64(m.Viewers),
			Likes:   int64(m.Likes),
		})
	}

	return d.Store.SaveDailyStats(ctx, stats)
}
