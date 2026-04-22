CREATE MATERIALIZED VIEW viewers_per_minute
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    video_id,
    AVG(viewers) AS avg_viewers
FROM livestream_metrics
GROUP BY bucket, video_id;
