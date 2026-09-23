-- Top streams per viewers
SELECT
    s.video_title,
    s.channel_title,
    MAX(m.viewers) AS max_viewers
FROM metrics_db.livestream_metrics m
         JOIN metrics_db.streams s ON s.video_id = m.video_id
WHERE m.time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
GROUP BY s.video_title, s.channel_title
ORDER BY max_viewers DESC
LIMIT 3;
