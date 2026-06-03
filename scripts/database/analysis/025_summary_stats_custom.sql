-- Resumen general (para imagen de stats)
SELECT
    COUNT(DISTINCT lm.video_id) as total_streams,
    COUNT(DISTINCT s.channel_id) as total_channels,
    MAX(lm.viewers) as overall_peak,
    ROUND(AVG(lm.viewers), 0) as overall_avg,
    SUM(lm.viewers) / 1000000.0 as total_viewers_millions
FROM metrics_db.livestream_metrics lm
         JOIN metrics_db.streams s ON s.video_id = lm.video_id
WHERE lm.time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp;
