SELECT
    s.video_title,
    s.channel_title,
    MAX(m.viewers) AS peak
FROM livestream_metrics m
         JOIN streams s ON s.video_id = m.video_id
GROUP BY s.video_title, s.channel_title
ORDER BY peak DESC;
