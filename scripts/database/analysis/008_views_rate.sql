SELECT
    s.channel_title,
    AVG(m.viewers) AS avg_viewers
FROM livestream_metrics m
         JOIN streams s ON s.video_id = m.video_id
GROUP BY s.channel_title
ORDER BY avg_viewers DESC;
