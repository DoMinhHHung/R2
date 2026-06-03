
CREATE TYPE user_status AS ENUM ('ACTIVE', 'BANNED', 'DELETED');
CREATE TYPE user_gender AS ENUM ('MALE', 'FEMALE', 'OTHER');

CREATE TABLE users (
    id               UUID         PRIMARY KEY,

    email            VARCHAR(255) NOT NULL UNIQUE,

    full_name        VARCHAR(255),

    phone_number     VARCHAR(20)  UNIQUE,

    gender           user_gender,

    date_of_birth    DATE,

    avatar_url       TEXT,


    avatar_public_id TEXT,

    profile_completed BOOLEAN     NOT NULL DEFAULT FALSE,

    status           user_status  NOT NULL DEFAULT 'ACTIVE',

    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email  ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status != 'DELETED';

CREATE TABLE user_roles (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by UUID,
    PRIMARY KEY (user_id, role)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();