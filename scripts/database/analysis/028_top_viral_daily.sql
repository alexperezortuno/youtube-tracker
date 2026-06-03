-- TOP VIRAL por día (viral_ratio = peak/avg)
-- Viral ratio >= 10 = 🔥 VIRAL, >= 5 = 📈 HIGH
WITH daily_stats AS (
    SELECT 
        video_id,
        DATE_TRUNC('day', time) as day,
        MAX(viewers) as peak,
        AVG(viewers) as avg_viewers,
        COUNT(*) as samples
    FROM metrics_db.livestream_metrics
    WHERE time >= NOW() - INTERVAL '7 days'
    GROUP BY video_id, DATE_TRUNC('day', time)
    HAVING COUNT(*) >= 10
)
SELECT
    s.video_title,
    s.video_id,
    ch.category,
    ch.language,
    ch.country,
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
LIMIT 20;