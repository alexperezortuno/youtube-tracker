-- Livestreams por día (última semana)
WITH daily AS (
    SELECT
        video_id,
        DATE_TRUNC('day', time) as day,
        MAX(viewers) as peak_viewers,
        AVG(viewers) as avg_viewers
    FROM metrics_db.livestream_metrics
    WHERE time >= NOW() - INTERVAL '7 days'
    GROUP BY video_id, DATE_TRUNC('day', time)
)
SELECT
    s.video_title,
    d.video_id,
    ch.category,
    ch.language,
    d.day,
    d.peak_viewers,
    d.avg_viewers
FROM daily d
JOIN metrics_db.streams s ON s.video_id = d.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
ORDER BY d.peak_viewers DESC
LIMIT 50;