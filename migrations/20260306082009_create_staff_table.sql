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
    manager_id BIGSERIAL,
    invite_token TEXT,
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
 );

-- обновление updated_at для всех таблиц
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $function$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$function$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- триггер обновления updated_at для таблицы staff
-- +goose StatementBegin
CREATE TRIGGER staff_updated_at
    BEFORE UPDATE ON staff
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_staff_email_unique;
DROP TYPE IF EXISTS user_status_type;
DROP TYPE IF EXISTS user_role_type;
DROP TABLE IF EXISTS staff;
