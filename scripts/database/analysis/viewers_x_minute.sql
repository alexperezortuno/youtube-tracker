SELECT
    video_id,
    time_bucket('1 minute', time) AS bucket,
    MAX(viewers) - MIN(viewers) AS growth
FROM livestream_metrics
GROUP BY video_id, bucket
ORDER BY growth DESC;
