WITH last_per_video AS (
    SELECT video_id, MAX(date) AS last_date
    FROM metrics_db.video_daily_stats
    WHERE date BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
    GROUP BY video_id
)
SELECT
    s.video_title,
    s.channel_title,
    vds.video_id,
    vds.likes,
    vds.views
FROM metrics_db.video_daily_stats vds
         JOIN last_per_video lpv
              ON vds.video_id = lpv.video_id
                  AND vds.date = lpv.last_date
         JOIN metrics_db.streams s
              ON s.video_id = vds.video_id
WHERE vds.date BETWEEN '$dateFrom'::timestamp AND '$dateTo'::timestamp
ORDER BY vds.likes DESC
LIMIT 10;
