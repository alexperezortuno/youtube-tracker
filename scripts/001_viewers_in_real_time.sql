SELECT DISTINCT ON (video_id)
    video_id,
    viewers,
    time
FROM livestream_metrics
ORDER BY video_id, time DESC;
