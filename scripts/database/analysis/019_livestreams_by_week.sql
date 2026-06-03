-- Livestreams por semana (últimas 4 semanas)
WITH weekly AS (
    SELECT
        video_id,
        DATE_TRUNC('week', time) as week,
        MAX(viewers) as peak_viewers,
        AVG(viewers) as avg_viewers
    FROM metrics_db.livestream_metrics
    WHERE time >= NOW() - INTERVAL '4 weeks'
    GROUP BY video_id, DATE_TRUNC('week', time)
)
SELECT
    s.video_title,
    w.video_id,
    ch.category,
    ch.country,
    w.week,
    w.peak_viewers,
    w.avg_viewers
FROM weekly w
JOIN metrics_db.streams s ON s.video_id = w.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
ORDER BY w.peak_viewers DESC
LIMIT 50;