SELECT
    time_bucket('1 minute', time) AS bucket,
    avg(viewers) AS avg_viewers
FROM livestream_metrics
WHERE video_id = $1
GROUP BY bucket
ORDER BY bucket;
