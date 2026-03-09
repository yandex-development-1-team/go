-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE application_type AS ENUM ('box', 'special_project');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE application_source AS ENUM ('telegram_bot', 'manual');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE application_status AS ENUM ('queue', 'in_progress', 'done', 'cancelled');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS applications (
    id                 BIGSERIAL          PRIMARY KEY,
    type               application_type   NOT NULL,
    source             application_source NOT NULL,
    status             application_status NOT NULL DEFAULT 'queue',
    customer_name      VARCHAR(255)       NOT NULL,
    contact_info       VARCHAR(255)       NOT NULL,
    project_name       VARCHAR(255),
    box_id             BIGINT,
    special_project_id BIGINT,
    manager_id         BIGINT,
    created_at         TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_special_project
    FOREIGN KEY (special_project_id)
    REFERENCES special_projects (id)
    ON DELETE RESTRICT 
);

CREATE INDEX IF NOT EXISTS idx_applications_status     ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_type       ON applications(type);
CREATE INDEX IF NOT EXISTS idx_applications_manager_id ON applications(manager_id);
CREATE INDEX IF NOT EXISTS idx_applications_created_at ON applications(created_at);

-- +goose Down
DROP TABLE IF EXISTS applications;
DROP TYPE IF EXISTS application_status;
DROP TYPE IF EXISTS application_source;
DROP TYPE IF EXISTS application_type;
