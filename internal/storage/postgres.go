package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

type DBSource struct {
	DB *sql.DB
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
			`INSERT INTO livestream_metrics 
				(time, video_id, video_title, channel_title, viewers, likes)
				VALUES ($1, $2, $3, $4, $5, $6)`,
			time.Now(),
			m.VideoID,
			m.VideoTitle,
			m.ChannelTitle,
			m.Viewers,
			m.Likes,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) SaveStreams(ctx context.Context, streams []models.Stream) error {

	for _, st := range streams {
		_, err := s.DB.Exec(ctx,
			`INSERT INTO streams (video_id, video_title, channel_title)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (video_id) DO NOTHING`,
			st.VideoID,
			st.VideoTitle,
			st.ChannelTitle,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBSource) GetChannelIDs() ([]string, error) {
	rows, err := d.DB.Query("SELECT channel_id FROM channels WHERE active = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		result = append(result, id)
	}

	return result, nil
}

func (s *Store) SaveDailyStats(ctx context.Context, stats []models.VideoDailyStat) error {
	for _, st := range stats {
		_, err := s.DB.Exec(ctx,
			`INSERT INTO video_daily_stats (date, video_id, views, likes)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (date, video_id)
			 DO UPDATE SET
			   views = EXCLUDED.views,
			   likes = EXCLUDED.likes`,
			st.Date,
			st.VideoID,
			st.Views,
			st.Likes,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetAllVideoIDs(ctx context.Context) ([]string, error) {
	rows, err := s.DB.Query(ctx, "SELECT video_id FROM metrics_db.streams GROUP BY video_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
