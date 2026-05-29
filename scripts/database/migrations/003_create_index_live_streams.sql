CREATE INDEX idx_metrics_video_time
    ON metrics_db.livestream_metrics (video_id, time DESC);

CREATE INDEX idx_metrics_time
    ON metrics_db.livestream_metrics (time DESC);
