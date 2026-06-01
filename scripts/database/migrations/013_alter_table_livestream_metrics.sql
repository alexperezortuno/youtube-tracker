ALTER TABLE metrics_db.livestream_metrics
    ADD COLUMN IF NOT EXISTS channel_id TEXT;
