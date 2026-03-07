-- +goose Up

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE user_role_type AS ENUM ('admin', 'manager');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE user_status_type AS ENUM ('active', 'blocked', 'invited');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    grade SMALLINT DEFAULT 0,
    is_admin BOOLEAN DEFAULT FALSE,
    password_hash TEXT NOT NULL,
    role user_role_type  DEFAULT 'manager',
    status user_status_type  DEFAULT 'invited',
    invite_token TEXT,
    permissions TEXT[] DEFAULT '{}',
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status_type;
DROP TYPE IF EXISTS user_role_type;
