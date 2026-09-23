-- Top videos por viewers en vivo (pico)
WITH peak_viewers AS (
    SELECT
        video_id,
        MAX(viewers) as peak_viewers,
        AVG(viewers) as avg_viewers,
        COUNT(*) as data_points
    FROM metrics_db.livestream_metrics
    WHERE time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
    GROUP BY video_id
)
SELECT
    s.video_title,
    pv.video_id,
    ch.category,
    ch.language,
    ch.country,
    pv.peak_viewers,
    pv.avg_viewers,
    pv.data_points
FROM peak_viewers pv
         JOIN metrics_db.streams s ON s.video_id = pv.video_id
         LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
ORDER BY pv.peak_viewers DESC
LIMIT 20;
