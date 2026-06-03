SELECT
    m.video_id,
    MAX(m.likes)::float / NULLIF(MAX(m.viewers), 0) AS engagement_rate
FROM livestream_metrics m
GROUP BY m.video_id
ORDER BY engagement_rate DESC;
