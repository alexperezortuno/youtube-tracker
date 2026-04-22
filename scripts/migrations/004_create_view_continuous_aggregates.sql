CREATE MATERIALIZED VIEW viewers_per_minute AS
SELECT
    date_trunc('minute', time) AS bucket,
    video_id,
    AVG(viewers) AS avg_viewers
FROM livestream_metrics
GROUP BY 1, 2;
