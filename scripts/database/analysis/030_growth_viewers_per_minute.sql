-- Growth of viewers per minute
SELECT
    date_trunc('minute', time) AS bucket,
    MAX(viewers) - MIN(viewers) AS growth
FROM metrics_db.livestream_metrics
WHERE
    time BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
        AND (COALESCE(NULLIF('$video_id', ''), '') = ''
        OR video_id = '$video_id')
   OR (COALESCE(NULLIF('$channel', ''), '') = ''
    OR channel_title = '$channel')
GROUP BY bucket
ORDER BY bucket;
