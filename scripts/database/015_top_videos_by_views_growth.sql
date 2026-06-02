-- Top videos por crecimiento de views (últimas 24h vs día anterior)
SELECT
    video_id,
    MAX(views) as total_views,
    MAX(views) - COALESCE(
        (SELECT MAX(views) FROM metrics_db.video_daily_stats d2
         WHERE d2.video_id = d1.video_id
         AND d2.date < CURRENT_DATE - INTERVAL '1 day'
         AND d2.date >= CURRENT_DATE - INTERVAL '2 days'),
        0
    ) as views_growth_24h,
    MAX(views) - COALESCE(
        (SELECT MAX(views) FROM metrics_db.video_daily_stats d3
         WHERE d3.video_id = d1.video_id
         AND d3.date < CURRENT_DATE - INTERVAL '7 days'
         AND d3.date >= CURRENT_DATE - INTERVAL '8 days'),
        0
    ) as views_growth_7d
FROM metrics_db.video_daily_stats d1
WHERE date >= CURRENT_DATE - INTERVAL '1 day'
GROUP BY video_id
ORDER BY views_growth_24h DESC
LIMIT 20;