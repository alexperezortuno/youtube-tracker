CREATE TABLE IF NOT EXISTS metrics_db.livestream_metrics (
    time TIMESTAMPTZ NOT NULL,
    video_id TEXT,
    video_title TEXT,
    channel_title TEXT,
    viewers INT,
    likes INT
);

comment on table metrics_db.livestream_metrics is 'Metrics for livestream videos, including viewer count and likes at specific timestamps.';
comment on column metrics_db.livestream_metrics.time is 'Timestamp of the metric.';
comment on column metrics_db.livestream_metrics.video_id is 'ID of the livestream video.';
comment on column metrics_db.livestream_metrics.viewers is 'Number of viewers at the given timestamp.';
comment on column metrics_db.livestream_metrics.likes is 'Number of likes at the given timestamp.';
