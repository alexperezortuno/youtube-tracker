SELECT
    s.video_title,
    s.channel_title,
    MAX(m.viewers) AS peak
FROM
    metrics_db.livestream_metrics m
        JOIN metrics_db.streams s ON s.video_id = m.video_id
WHERE m.time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
    AND COALESCE(NULLIF('$channel', ''), '') = ''
   OR m.channel_title = '$channel'
GROUP BY
    s.video_title,
    s.channel_title
ORDER BY
    peak DESC
LIMIT
    15;
