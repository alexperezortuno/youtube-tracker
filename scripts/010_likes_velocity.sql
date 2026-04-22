SELECT
    video_id,
    time_bucket('1 minute', time) AS bucket,
    MAX(likes) - MIN(likes) AS likes_growth
FROM livestream_metrics
GROUP BY video_id, bucket
ORDER BY likes_growth DESC;
