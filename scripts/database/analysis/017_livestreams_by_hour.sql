-- Livestreams por hora (últimas 24h)
WITH hourly AS (
    SELECT
        video_id,
        DATE_TRUNC('hour', time) as hour,
        MAX(viewers) as peak_viewers,
        AVG(viewers) as avg_viewers
    FROM metrics_db.livestream_metrics
    WHERE time >= NOW() - INTERVAL '24 hours'
    GROUP BY video_id, DATE_TRUNC('hour', time)
)
SELECT
    s.video_title,
    h.video_id,
    ch.category,
    h.hour,
    h.peak_viewers,
    h.avg_viewers
FROM hourly h
JOIN metrics_db.streams s ON s.video_id = h.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
ORDER BY h.peak_viewers DESC
LIMIT 50;