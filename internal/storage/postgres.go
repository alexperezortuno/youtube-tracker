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

type DBSource struct {
	Pool *pgxpool.Pool
}

func NewStore(conn string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}
	return &Store{DB: pool}, nil
}

func NewDBSource(pool *pgxpool.Pool) *DBSource {
	return &DBSource{Pool: pool}
}

func (s *Store) SaveMetrics(ctx context.Context, metrics []models.Metric) error {
	for _, m := range metrics {
		_, err := s.DB.Exec(ctx,
			`INSERT INTO livestream_metrics
				(time, video_id, video_title, channel_title, channel_id, viewers, likes)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			time.Now(),
			m.VideoID,
			m.VideoTitle,
			m.ChannelTitle,
			m.ChannelID,
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
			`INSERT INTO streams (video_id, video_title, channel_title, channel_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (video_id) DO NOTHING`,
			st.VideoID,
			st.VideoTitle,
			st.ChannelTitle,
			st.ChannelID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBSource) GetChannelIDs() ([]string, error) {
	rows, err := d.Pool.Query(context.Background(), "SELECT id FROM channels WHERE active = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	return result, nil
}

func (d *DBSource) GetChannels(activeOnly bool) ([]models.Channel, error) {
	query := "SELECT id, name, active, category, language, country, followed_at, created_at, updated_at FROM channels"
	if activeOnly {
		query += " WHERE active = true"
	}

	rows, err := d.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var ch models.Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Active, &ch.Category, &ch.Language, &ch.Country, &ch.FollowedAt, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}

	return channels, nil
}

func (d *DBSource) AddChannel(ctx context.Context, ch models.Channel) error {
	_, err := d.Pool.Exec(ctx,
		`INSERT INTO channels (id, name, active, category, language, country, followed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE SET
		   name = EXCLUDED.name,
		   active = EXCLUDED.active,
		   category = EXCLUDED.category,
		   language = EXCLUDED.language,
		   country = EXCLUDED.country,
		   updated_at = NOW()`,
		ch.ID, ch.Name, ch.Active, ch.Category, ch.Language, ch.Country, ch.FollowedAt,
	)
	return err
}

func (d *DBSource) UpdateChannel(ctx context.Context, ch models.Channel) error {
	_, err := d.Pool.Exec(ctx,
		`UPDATE channels SET
		   name = $2,
		   active = $3,
		   category = $4,
		   language = $5,
		   country = $6,
		   updated_at = NOW()
		 WHERE id = $1`,
		ch.ID, ch.Name, ch.Active, ch.Category, ch.Language, ch.Country,
	)
	return err
}

func (d *DBSource) DeleteChannel(ctx context.Context, id string) error {
	_, err := d.Pool.Exec(ctx, "DELETE FROM channels WHERE id = $1", id)
	return err
}

func (d *DBSource) GetChannelByID(ctx context.Context, id string) (*models.Channel, error) {
	var ch models.Channel
	err := d.Pool.QueryRow(ctx,
		"SELECT id, name, active, category, language, country, followed_at, created_at, updated_at FROM channels WHERE id = $1",
		id,
	).Scan(&ch.ID, &ch.Name, &ch.Active, &ch.Category, &ch.Language, &ch.Country, &ch.FollowedAt, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *Store) SaveDailyStats(ctx context.Context, stats []models.VideoDailyStat) error {
	for _, st := range stats {
		_, err := s.DB.Exec(ctx,
			`INSERT INTO video_daily_stats (date, video_id, views, likes, favorites, comments, channel_id, published_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (date, video_id)
			 DO UPDATE SET
			   views = EXCLUDED.views,
			   likes = EXCLUDED.likes,
			   favorites = EXCLUDED.favorites,
			   comments = EXCLUDED.comments,
			   channel_id = EXCLUDED.channel_id,
			   published_at = EXCLUDED.published_at`,
			st.Date,
			st.VideoID,
			st.Views,
			st.Likes,
			st.Favorites,
			st.Comments,
			st.ChannelID,
			st.PublishedAt,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetAllVideoIDs(ctx context.Context) ([]string, error) {
	rows, err := s.DB.Query(ctx, "SELECT video_id FROM metrics_db.livestream_metrics GROUP BY video_id")
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
