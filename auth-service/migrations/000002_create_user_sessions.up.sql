-- migrations/000002_create_user_sessions.up.sql

CREATE TABLE IF NOT EXISTS user_sessions (
    id                 UUID         PRIMARY KEY,
    user_id            UUID         NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT         NOT NULL,
    device_name        VARCHAR(255),
    device_type        VARCHAR(50),
    browser            VARCHAR(100),
    os                 VARCHAR(100),
    ip_address         INET,
    user_agent         TEXT,
    last_activity      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at         TIMESTAMPTZ  NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id    ON user_sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
-- Partial index: chỉ index các session chưa expire — tiết kiệm space, query nhanh hơn
CREATE INDEX idx_sessions_active     ON user_sessions(user_id, last_activity)
    WHERE expires_at > NOW();