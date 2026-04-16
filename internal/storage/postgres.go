package storage

import (
	"context"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

func NewStore(conn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}
	return &Store{DB: pool}, nil
}

func (s *Store) SaveMetrics(ctx context.Context, metrics []models.Metric) error {

	for _, m := range metrics {
		_, err := s.DB.Exec(ctx,
			`INSERT INTO livestream_metrics (time, video_id, viewers, likes)
			 VALUES ($1, $2, $3, $4)`,
			time.Now(), m.VideoID, m.Viewers, m.Likes,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
