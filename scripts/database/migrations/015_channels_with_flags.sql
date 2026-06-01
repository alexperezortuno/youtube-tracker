CREATE EXTENSION IF NOT EXISTS timescaledb;

ALTER TABLE metrics_db.channels
    ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS category TEXT,
    ADD COLUMN IF NOT EXISTS language TEXT,
    ADD COLUMN IF NOT EXISTS country TEXT,
    ADD COLUMN IF NOT EXISTS followed_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_channels_active ON metrics_db.channels(active);
CREATE INDEX IF NOT EXISTS idx_channels_category ON metrics_db.channels(category);
CREATE INDEX IF NOT EXISTS idx_channels_language ON metrics_db.channels(language);
CREATE INDEX IF NOT EXISTS idx_channels_country ON metrics_db.channels(country);

ALTER TABLE metrics_db.streams DROP CONSTRAINT IF EXISTS fk_channels;
ALTER TABLE metrics_db.streams ADD CONSTRAINT fk_channels
    FOREIGN KEY (channel_id) REFERENCES metrics_db.channels(id) ON DELETE SET NULL;

ALTER TABLE metrics_db.livestream_metrics DROP CONSTRAINT IF EXISTS fk_channels_metrics;
ALTER TABLE metrics_db.livestream_metrics ADD CONSTRAINT fk_channels_metrics
    FOREIGN KEY (channel_id) REFERENCES metrics_db.channels(id) ON DELETE SET NULL;

ALTER TABLE metrics_db.video_daily_stats DROP CONSTRAINT IF EXISTS fk_channels_daily;
ALTER TABLE metrics_db.video_daily_stats ADD CONSTRAINT fk_channels_daily
    FOREIGN KEY (channel_id) REFERENCES metrics_db.channels(id) ON DELETE SET NULL;
