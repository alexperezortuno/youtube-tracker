-- Top videos por crecimiento de views
WITH latest AS (
    SELECT video_id, MAX(views) AS views
    FROM metrics_db.video_daily_stats
    WHERE date BETWEEN '$dateFrom' AND '$dateTo'
    GROUP BY video_id
),
     yesterday AS (
         SELECT video_id, MAX(views) AS views
         FROM metrics_db.video_daily_stats
         WHERE date >= ('$dateTo'::timestamp - INTERVAL '1 day')
           AND date <  '$dateTo'::timestamp
         GROUP BY video_id
     )
SELECT s.video_title,
       s.video_id,
       ch.category,
       ch.language,
       l.views                        AS total_views,
       l.views - COALESCE(y.views, 0) AS views_growth_24h
FROM latest l
         JOIN metrics_db.streams s ON s.video_id = l.video_id
         LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
         LEFT JOIN yesterday y ON y.video_id = l.video_id
ORDER BY views_growth_24h DESC
LIMIT 20;
