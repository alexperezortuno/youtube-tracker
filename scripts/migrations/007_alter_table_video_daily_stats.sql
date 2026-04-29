ALTER TABLE metrics_db.video_daily_stats ADD COLUMN "channel_id" VARCHAR(255) DEFAULT NULL;
ALTER TABLE metrics_db.video_daily_stats ADD COLUMN "published_at" VARCHAR(255) DEFAULT NULL;

COMMENT ON COLUMN metrics_db.video_daily_stats.published_at IS 'ISO 8601';
COMMENT ON COLUMN metrics_db.video_daily_stats.channel_id IS 'Channel ID';
