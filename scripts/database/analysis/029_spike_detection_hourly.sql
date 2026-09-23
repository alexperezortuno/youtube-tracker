-- SPIKE DETECTION: Crecimientos súbitos por hora
WITH hourly_delta AS (
    SELECT 
        video_id,
        time,
        viewers,
        viewers - LAG(viewers) OVER (PARTITION BY video_id ORDER BY time) as growth
    FROM metrics_db.livestream_metrics
    WHERE time >= NOW() - INTERVAL '24 hours'
),
max_spike AS (
    SELECT video_id, MAX(growth) as max_growth
    FROM hourly_delta
    WHERE growth > 0
    GROUP BY video_id
)
SELECT
    s.video_title,
    ms.video_id,
    ch.category,
    ch.language,
    ch.country,
    ms.max_growth as max_spike_viewers,
    ROUND(ms.max_growth / 60.0, 1) as viewers_per_minute,
    CASE
        WHEN ms.max_growth >= 50000 THEN '🔥 VIRAL SPIKE'
        WHEN ms.max_growth >= 10000 THEN '📈 HIGH SPIKE'
        ELSE '📊 MODERATE'
    END as spike_status
FROM max_spike ms
JOIN metrics_db.streams s ON s.video_id = ms.video_id
LEFT JOIN metrics_db.channels ch ON ch.id = s.channel_id
WHERE ch.category IN ('right', 'left', 'right+left', 'opinion', 'news')
ORDER BY max_spike_viewers DESC
LIMIT 15;