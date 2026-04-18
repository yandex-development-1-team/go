-- +goose Up

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE application_status AS ENUM ('pending', 'confirmed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS applications (
    id BIGSERIAL PRIMARY KEY,
    status application_status NOT NULL DEFAULT 'pending',
    manager_id BIGINT,
    form_answer_id VARCHAR(255) UNIQUE,
    customer_name VARCHAR(255) NOT NULL,
    contact_info VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL,

    CONSTRAINT fk_manager_id_applications
        FOREIGN KEY (manager_id)
            REFERENCES staff(id)
            ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_applications_status     ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_customer_name ON applications(customer_name);
CREATE INDEX IF NOT EXISTS idx_applications_manager_id ON applications(manager_id);

-- +goose Down
DROP INDEX IF EXISTS idx_applications_status;
DROP INDEX IF EXISTS idx_applications_customer_name;
DROP INDEX IF EXISTS idx_applications_manager_id;
DROP TYPE IF EXISTS application_status;
DROP TABLE IF EXISTS applications;
