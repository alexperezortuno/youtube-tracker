-- Top videos por crecimiento de views (últimas 24h vs día anterior)
SELECT
    s.video_title,
    s.video_id,
    ch.category,
    ch.language,
    MAX(sd.views) AS total_views,
    MAX(sd.views) - COALESCE(
            (
                SELECT MAX(d2.views)
                FROM metrics_db.video_daily_stats d2
                WHERE d2.video_id = s.video_id
                  AND d2.date < CURRENT_DATE - INTERVAL '1 day'
                AND d2.date >= CURRENT_DATE - INTERVAL '2 days'
        ),
            0
                    ) AS views_growth_24h
FROM metrics_db.video_daily_stats sd
         JOIN metrics_db.streams s ON s.video_id = sd.video_id
         LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE sd.date >= CURRENT_DATE - INTERVAL '1 day'
GROUP BY s.video_title, s.video_id, ch.category, ch.language
ORDER BY views_growth_24h DESC
    LIMIT 20;
