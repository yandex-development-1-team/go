-- +goose Up
CREATE TABLE files (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    object_name TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    url TEXT NOT NULL,
    mime_type TEXT,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_is_active_false
    ON files (updated_at)
    WHERE is_active = false;

-- +goose Down
DROP TABLE IF EXISTS files;
