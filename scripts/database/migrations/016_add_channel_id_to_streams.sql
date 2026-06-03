ALTER TABLE metrics_db.streams ADD COLUMN IF NOT EXISTS channel_id TEXT;

CREATE INDEX IF NOT EXISTS idx_streams_channel_id ON metrics_db.streams(channel_id);