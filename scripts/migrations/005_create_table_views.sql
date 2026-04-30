CREATE TABLE IF NOT EXISTS metrics_db.video_daily_stats
(
    date     DATE NOT NULL,
    video_id TEXT NOT NULL,
    views    BIGINT,
    likes    BIGINT,
    PRIMARY KEY (date, video_id)
    );

CREATE INDEX IF NOT EXISTS video_daily_stats_date_index ON metrics_db.video_daily_stats (date);
CREATE INDEX IF NOT EXISTS video_daily_stats_video_id_index ON metrics_db.video_daily_stats (video_id);

COMMENT ON TABLE metrics_db.video_daily_stats IS 'Daily statistics for each video, including views and likes.';
COMMENT ON COLUMN metrics_db.video_daily_stats.date IS 'The date of the statistics.';
COMMENT ON COLUMN metrics_db.video_daily_stats.video_id IS 'The ID of the video.';
COMMENT ON COLUMN metrics_db.video_daily_stats.views IS 'The number of views for the video on the given date.';
COMMENT ON COLUMN metrics_db.video_daily_stats.likes IS 'The number of likes for the video on the given date.';
