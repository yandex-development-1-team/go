-- +goose Up
CREATE TABLE IF NOT EXISTS special_projects (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image TEXT,
    is_active_in_bot BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_special_projects_active ON special_projects (is_active_in_bot) WHERE is_active_in_bot = TRUE;
CREATE INDEX idx_special_projects_updated_at ON special_projects (updated_at DESC);
CREATE INDEX idx_special_projects_search ON special_projects USING GIN (to_tsvector('russian', title || ' ' || COALESCE(description, '')));

-- +goose Down
DROP TABLE IF EXISTS special_projects;