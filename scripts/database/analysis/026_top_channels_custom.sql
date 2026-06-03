-- Top channels por total viewers (rango custom)
SELECT
    ch.id as channel_id,
    ch.name as channel_name,
    ch.category,
    ch.language,
    COUNT(DISTINCT lm.video_id) as stream_count,
    MAX(lm.viewers) as peak_viewers,
    ROUND(AVG(lm.viewers), 0) as avg_viewers,
    SUM(lm.viewers) / 1000000.0 as total_viewers_millions
FROM metrics_db.livestream_metrics lm
         JOIN metrics_db.streams s ON s.video_id = lm.video_id
         JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE lm.time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
GROUP BY ch.id, ch.name, ch.category, ch.language
ORDER BY total_viewers_millions DESC
LIMIT 20;
