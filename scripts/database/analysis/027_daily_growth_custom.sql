-- Crecimiento de viewers día a día (para graficar tendencia)
WITH daily AS (
    SELECT
        DATE_TRUNC('day', lm.time) as day,
        SUM(lm.viewers) as total_viewers,
        MAX(lm.viewers) as peak_of_day,
        AVG(lm.viewers) as avg_of_day
    FROM metrics_db.livestream_metrics lm
    WHERE lm.time BETWEEN '$dateFrom' AND '$dateTo'
    GROUP BY DATE_TRUNC('day', lm.time)
)
SELECT
    day,
    total_viewers,
    peak_of_day,
    avg_of_day,
    LAG(total_viewers) OVER (ORDER BY day) as prev_day_viewers,
    ROUND((total_viewers - LAG(total_viewers) OVER (ORDER BY day))::numeric /
          NULLIF(LAG(total_viewers) OVER (ORDER BY day), 0) * 100, 1) as growth_pct
FROM daily
ORDER BY day;
