SELECT
    time_bucket('1 minute', time) AS bucket,
    video_id,
    AVG(likes::float / NULLIF(viewers, 0)) AS engagement
FROM livestream_metrics
GROUP BY bucket, video_id
ORDER BY bucket;
