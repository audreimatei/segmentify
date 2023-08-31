CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users_segments (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_id BIGINT REFERENCES segments(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, segment_id)
);

CREATE TYPE OPERATION AS ENUM ('add', 'remove');
CREATE TABLE IF NOT EXISTS users_segments_history (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    segment_id BIGINT REFERENCES segments(id) ON DELETE CASCADE,
    operation OPERATION NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
