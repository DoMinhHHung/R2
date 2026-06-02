CREATE TYPE user_role AS ENUM ('ADMIN', 'LANDLORD', 'TENANT');

CREATE TABLE IF NOT EXISTS auth_users (
    id            UUID         PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    role          user_role    NOT NULL DEFAULT 'TENANT',
    is_verified   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_users_email ON auth_users(email);

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_auth_users_updated_at
    BEFORE UPDATE ON auth_users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();