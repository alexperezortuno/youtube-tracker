-- TOP VIRAL por día (rango custom)
-- Usar con Grafana: $from, $to
WITH daily_stats AS (
    SELECT 
        video_id,
        DATE_TRUNC('day', time) as day,
        MAX(viewers) as peak,
        AVG(viewers) as avg_viewers,
        COUNT(*) as samples
    FROM metrics_db.livestream_metrics
    WHERE time BETWEEN '$from' AND '$to'
    GROUP BY video_id, DATE_TRUNC('day', time)
    HAVING COUNT(*) >= 5
)
SELECT
    s.video_title,
    s.video_id,
    ch.category,
    ch.language,
    ds.day,
    ds.peak as peak_viewers,
    ROUND(ds.avg_viewers, 0) as avg_viewers,
    ROUND(ds.peak::numeric / NULLIF(ds.avg_viewers, 0), 2) as viral_ratio,
    CASE 
        WHEN ROUND(ds.peak::numeric / NULLIF(ds.avg_viewers, 0), 2) >= 10 THEN '🔥 VIRAL'
        WHEN ROUND(ds.peak::numeric / NULLIF(ds.avg_viewers, 0), 2) >= 5 THEN '📈 HIGH'
        ELSE '📊 NORMAL'
    END as status
FROM daily_stats ds
JOIN metrics_db.streams s ON s.video_id = ds.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
ORDER BY viral_ratio DESC
LIMIT 25;