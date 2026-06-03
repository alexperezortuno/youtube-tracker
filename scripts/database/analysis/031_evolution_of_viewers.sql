-- Evolution of viewers
SELECT AVG(m.viewers) AS viewers
FROM metrics_db.livestream_metrics m
         JOIN metrics_db.streams s ON s.video_id = m.video_id
WHERE m.time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
  AND (
    COALESCE(NULLIF('$channel', ''), '') = ''
        OR s.channel_title = '$channel'
    );
