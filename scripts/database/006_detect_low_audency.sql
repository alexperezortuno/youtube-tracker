WITH lagged AS (
    SELECT
        video_id,
    time,
    viewers,
    LAG(viewers) OVER (PARTITION BY video_id ORDER BY time) AS prev_viewers
FROM livestream_metrics
    )
SELECT *
FROM lagged
WHERE prev_viewers IS NOT NULL
  AND viewers < prev_viewers * 0.7;
