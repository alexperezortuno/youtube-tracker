CREATE TABLE metrics_db.streams (
 video_id TEXT PRIMARY KEY,
 video_title TEXT,
 channel_title TEXT,
 created_at TIMESTAMPTZ DEFAULT NOW()
);
