SELECT
    s.video_title,
    s.channel_title,
    m.viewers,
    m.likes,
    m.time
FROM livestream_metrics m
         JOIN streams s ON s.video_id = m.video_id
WHERE m.time > NOW() - INTERVAL '1 minute'
ORDER BY m.viewers DESC
    LIMIT 10;
