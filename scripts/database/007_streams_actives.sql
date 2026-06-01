SELECT DISTINCT video_id
FROM livestream_metrics
WHERE time > NOW() - INTERVAL '2 minutes';
