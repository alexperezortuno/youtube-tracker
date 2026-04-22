CREATE TABLE IF NOT EXISTS metrics_db.livestream_metrics (
    time TIMESTAMPTZ NOT NULL,
    video_id TEXT,
    video_title TEXT,
    channel_title TEXT,
    viewers INT,
    likes INT
);
