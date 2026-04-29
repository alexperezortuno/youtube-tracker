ALTER TABLE metrics_db.video_daily_stats ADD COLUMN "favorites" BIGINT DEFAULT 0;
ALTER TABLE metrics_db.video_daily_stats ADD COLUMN "comments" BIGINT DEFAULT 0;

COMMENT ON COLUMN metrics_db.video_daily_stats.favorites IS 'Number of favorites';
COMMENT ON COLUMN metrics_db.video_daily_stats.comments IS 'Number of comments';
