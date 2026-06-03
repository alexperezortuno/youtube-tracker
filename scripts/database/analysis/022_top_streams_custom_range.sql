-- Parámetros para Grafana (usar variables defecha):
-- {{from:date:iso}} {{to:date:iso}}
-- O en SQL directo: BETWEEN '2024-01-01' AND '2024-01-07'

-- Top streams por viewers (rango custom)
SELECT
    s.video_title,
    s.video_id,
    ch.category,
    ch.language,
    ch.country,
    MAX(lm.viewers) as peak_viewers,
    ROUND(AVG(lm.viewers), 0) as avg_viewers,
    COUNT(*) as data_points,
    MIN(lm.time) as first_seen,
    MAX(lm.time) as last_seen
FROM metrics_db.livestream_metrics lm
JOIN metrics_db.streams s ON s.video_id = lm.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE lm.time BETWEEN '$dateFrom' AND '$dateTo'
GROUP BY s.video_title, s.video_id, ch.category, ch.language, ch.country
ORDER BY peak_viewers DESC
LIMIT 20;
