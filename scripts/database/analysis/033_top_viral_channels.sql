-- TOP CHANNELS VIRALES (suma de viral_ratio por canal)
WITH channel_viral AS (
    SELECT 
        ch.id as channel_id,
        ch.name as channel_name,
        ch.category as ideology,
        ch.language,
        SUM(
            lm.viewers::numeric / 
            NULLIF(AVG(lm.viewers) OVER (PARTITION BY lm.video_id), 0)
        ) as total_viral_score,
        MAX(lm.viewers) as max_peak,
        AVG(lm.viewers) as avg_viewers,
        COUNT(DISTINCT lm.video_id) as stream_count
    FROM metrics_db.livestream_metrics lm
    JOIN metrics_db.streams s ON s.video_id = lm.video_id
    JOIN metrics_db.channels ch ON ch.id = s.channel_id
    WHERE lm.time >= NOW() - INTERVAL '7 days'
      AND ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
    GROUP BY ch.id, ch.name, ch.category, ch.language
)
SELECT
    channel_id,
    channel_name,
    ideology,
    language,
    ROUND(total_viral_score, 2) as viral_score,
    max_peak as max_viewers,
    ROUND(avg_viewers, 0) as avg_viewers,
    stream_count
FROM channel_viral
ORDER BY viral_score DESC
LIMIT 20;