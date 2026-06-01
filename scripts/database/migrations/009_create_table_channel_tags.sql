CREATE TABLE IF NOT EXISTS metrics_db.channel_tags
(
    channel_id TEXT,
    tag_id     INT,
    PRIMARY KEY (channel_id, tag_id)
);
