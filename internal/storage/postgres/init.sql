CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
    slug TEXT PRIMARY KEY,
    percent SMALLINT NOT NULL CHECK (percent >= 0 AND percent <= 100)
);

CREATE TABLE IF NOT EXISTS users_segments (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_slug TEXT REFERENCES segments(slug) ON DELETE CASCADE,
    expire_at TIMESTAMP,
    PRIMARY KEY (user_id, segment_slug)
);

CREATE TABLE IF NOT EXISTS users_segments_history (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_slug TEXT REFERENCES segments(slug) ON DELETE CASCADE,
    operation TEXT NOT NULL CHECK (operation IN ('add', 'remove')),
    created_at TIMESTAMP DEFAULT NOW()
);
