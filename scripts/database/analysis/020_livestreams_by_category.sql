-- Top livestreams por categoría
WITH category_stats AS (
    SELECT
        ch.category,
        lm.video_id,
        MAX(lm.viewers) as peak_viewers,
        AVG(lm.viewers) as avg_viewers,
        COUNT(*) as total_minutes
    FROM metrics_db.livestream_metrics lm
    JOIN metrics_db.streams s ON s.video_id = lm.video_id
    LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
    WHERE lm.time >= NOW() - INTERVAL '7 days'
      AND ch.category IS NOT NULL
    GROUP BY ch.category, lm.video_id
)
SELECT
    category,
    video_id,
    peak_viewers,
    avg_viewers,
    total_minutes,
    RANK() OVER (PARTITION BY category ORDER BY peak_viewers DESC) as rank_in_category
FROM category_stats
ORDER BY category, peak_viewers DESC;