ALTER TABLE metrics_db.livestream_metrics
    ADD CONSTRAINT fk_metrics_channel
        FOREIGN KEY (channel_id)
            REFERENCES metrics_db.channels (id);
