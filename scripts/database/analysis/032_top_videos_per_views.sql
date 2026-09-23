WITH last_per_video AS (
    SELECT
        vds.video_id,
        MAX(vds.date) AS last_date
    FROM metrics_db.video_daily_stats vds
    WHERE vds.date BETWEEN ('$dateFrom'::timestamp  - INTERVAL '1 day') AND '$dateTo'::timestamp
    GROUP BY vds.video_id
)
SELECT
    s.video_title,
    s.channel_title,
    vds.video_id,
    vds.views,
    vds.likes
FROM metrics_db.video_daily_stats vds
         JOIN last_per_video lp
              ON lp.video_id = vds.video_id
                  AND lp.last_date = vds.date
         JOIN metrics_db.streams s
              ON s.video_id = vds.video_id
WHERE vds.date BETWEEN ('$dateFrom'::timestamp - INTERVAL '1 day') AND '$dateTo'::timestamp
ORDER BY vds.views DESC
LIMIT 10;
