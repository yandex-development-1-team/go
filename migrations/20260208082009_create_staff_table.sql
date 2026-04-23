-- +goose Up

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE user_role_type AS ENUM ('admin', 'manager_1', 'manager_2', 'manager_3', 'user');
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

CREATE TABLE IF NOT EXISTS staff (
    id BIGSERIAL PRIMARY KEY,
    telegram_nick VARCHAR(64) UNIQUE,
    first_name VARCHAR(255) NOT NULL DEFAULT '',
    last_name VARCHAR(255) NOT NULL DEFAULT '',
    second_name VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL UNIQUE,
    phone_number VARCHAR(64) UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    role user_role_type DEFAULT 'user',
    status user_status_type DEFAULT 'invited',
    department VARCHAR(255),
    position VARCHAR(255),
    supervisor VARCHAR(255),
    address VARCHAR(255),
    image TEXT,
    invite_token TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
 );


-- +goose Down
DROP INDEX IF EXISTS idx_staff_email_unique;
DROP TYPE IF EXISTS user_status_type;
DROP TYPE IF EXISTS user_role_type;
DROP TABLE IF EXISTS staff;