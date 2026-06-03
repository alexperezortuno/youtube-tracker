-- Comparativa por tags ideológicos (ejemplo: right, left, right+left, opinion, news)
-- Agrupa por categoría y suma viewers
WITH ideological_stats AS (
    SELECT
        ch.category as tag,
        DATE_TRUNC('hour', lm.time) as hour,
        SUM(lm.viewers) as total_viewers,
        AVG(lm.viewers) as avg_viewers,
        COUNT(DISTINCT lm.video_id) as unique_streams
    FROM metrics_db.livestream_metrics lm
    JOIN metrics_db.streams s ON s.video_id = lm.video_id
    LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
    WHERE lm.time >= NOW() - INTERVAL '24 hours'
      AND ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
    GROUP BY ch.category, DATE_TRUNC('hour', lm.time)
)
SELECT
    tag,
    hour,
    total_viewers,
    avg_viewers,
    unique_streams
FROM ideological_stats
ORDER BY tag, hour DESC;