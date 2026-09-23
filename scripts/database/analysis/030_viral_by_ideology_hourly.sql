-- RANKING VIRAL por意识形态 (ideological comparison)
WITH viral_stats AS (
    SELECT 
        ch.category as ideology,
        DATE_TRUNC('hour', lm.time) as hour,
        SUM(lm.viewers) as total_viewers,
        MAX(lm.viewers) as peak_viewers,
        AVG(lm.viewers) as avg_viewers,
        COUNT(DISTINCT lm.video_id) as unique_streams,
        COUNT(*) as data_points
    FROM metrics_db.livestream_metrics lm
    JOIN metrics_db.streams s ON s.video_id = lm.video_id
    JOIN metrics_db.channels ch ON ch.id = s.channel_id
    WHERE lm.time >= NOW() - INTERVAL '24 hours'
      AND ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
    GROUP BY ch.category, DATE_TRUNC('hour', lm.time)
)
SELECT
    ideology,
    hour,
    total_viewers,
    peak_viewers,
    ROUND(avg_viewers, 0) as avg_viewers,
    unique_streams,
    RANK() OVER (PARTITION BY ideology ORDER BY total_viewers DESC) as hourly_rank
FROM viral_stats
ORDER BY ideology, total_viewers DESC;