-- Tendencia horaria (rango custom) - ideal para generar imagen
SELECT
    DATE_TRUNC('hour', lm.time) as hour,
    SUM(lm.viewers) as total_viewers,
    AVG(lm.viewers) as avg_viewers,
    COUNT(DISTINCT lm.video_id) as concurrent_streams
FROM metrics_db.livestream_metrics lm
WHERE lm.time BETWEEN '$dateFrom' AND '$dateTo'
GROUP BY hour
ORDER BY hour;
