-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE spec_type AS ENUM ('active', 'inactive');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS special_projects (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image TEXT,
    status spec_type,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_special_projects_updated_at ON special_projects (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_special_projects_search ON special_projects USING GIN (to_tsvector('russian', title || ' ' || COALESCE(description, '')));

-- +goose Down
DROP TABLE IF EXISTS special_projects;