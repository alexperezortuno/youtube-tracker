SELECT
    video_id,
    MAX(viewers) AS peak_viewers
FROM livestream_metrics
GROUP BY video_id
ORDER BY peak_viewers DESC
    LIMIT 10;
