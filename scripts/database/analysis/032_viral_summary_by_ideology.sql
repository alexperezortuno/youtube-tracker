-- RESUMEN VIRAL por ideología (rango custom)
-- Para imagen de stats
SELECT
    ch.category as ideology,
    COUNT(DISTINCT lm.video_id) as total_streams,
    MAX(lm.viewers) as max_peak,
    ROUND(AVG(lm.viewers), 0) as avg_viewers,
    SUM(lm.viewers) / 1000000.0 as total_millions,
    ROUND(
        MAX(lm.viewers)::numeric / 
        NULLIF(AVG(lm.viewers), 0), 2
    ) as max_viral_ratio
FROM metrics_db.livestream_metrics lm
JOIN metrics_db.streams s ON s.video_id = lm.video_id
JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE lm.time BETWEEN '$from' AND '$to'
  AND ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
GROUP BY ch.category
ORDER BY max_peak DESC;