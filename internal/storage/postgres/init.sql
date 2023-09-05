CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
    slug TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users_segments (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_slug TEXT REFERENCES segments(slug) ON DELETE CASCADE,
    expire_at TIMESTAMP,
    PRIMARY KEY (user_id, segment_slug)
);

CREATE TYPE OPERATION AS ENUM ('add', 'remove');
CREATE TABLE IF NOT EXISTS users_segments_history (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_slug TEXT REFERENCES segments(slug) ON DELETE CASCADE,
    operation OPERATION NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
