CREATE TABLE IF NOT EXISTS metrics_db.channels
(
    id         TEXT PRIMARY KEY, -- channel_id de YouTube
    name       TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
