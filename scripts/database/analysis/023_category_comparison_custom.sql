-- Comparativa por categoría (rango custom)
SELECT
    ch.category,
    COUNT(DISTINCT lm.video_id) as total_streams,
    MAX(lm.viewers) as peak_viewers,
    ROUND(AVG(lm.viewers), 2) as avg_viewers,
    SUM(lm.viewers) / 1000000.0 as total_viewers_millions
FROM metrics_db.livestream_metrics lm
JOIN metrics_db.streams s ON s.video_id = lm.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE lm.time BETWEEN '$dateFrom' AND '$dateTo'
  AND ch.category IS NOT NULL
GROUP BY ch.category
ORDER BY total_viewers_millions DESC;
