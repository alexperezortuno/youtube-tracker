ALTER TABLE metrics_db.streams
    ADD COLUMN IF NOT EXISTS  channel_id TEXT;

ALTER TABLE metrics_db.streams
    ADD CONSTRAINT fk_stream_channel
        FOREIGN KEY (channel_id)
            REFERENCES metrics_db.channels (id);
