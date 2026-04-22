CREATE INDEX idx_metrics_video_time
    ON livestream_metrics (video_id, time DESC);

CREATE INDEX idx_metrics_time
    ON livestream_metrics (time DESC);
