-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE services_type AS ENUM ('active', 'inactive');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    rules TEXT,
    location VARCHAR(255),
    price INTEGER NOT NULL,
    image TEXT,
    status services_type,
    organizer VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS service_available_slots (
    service_id BIGINT NOT NULL,
    slot_date DATE NOT NULL,
	start_time TIME NULL,
	end_time TIME NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_available_slots_service
        FOREIGN KEY (service_id)
            REFERENCES services (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_available_slots_service_date
        UNIQUE (service_id, slot_date, start_time, end_time)
);

-- триггер обновления updated_at для таблицы services
-- +goose StatementBegin
CREATE TRIGGER services_updated_at
    BEFORE UPDATE ON services
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
-- +goose StatementEnd

-- триггер обновления updated_at для таблицы services
-- +goose StatementBegin
CREATE TRIGGER service_available_slots_updated_at
    BEFORE UPDATE ON service_available_slots
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_full_day_slots 
    ON service_available_slots (service_id, slot_date) 
    WHERE start_time IS NULL AND end_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_available_slots_service_date
    ON service_available_slots (service_id, slot_date);

-- +goose Down
DROP INDEX IF EXISTS idx_available_slots_service_date;
DROP INDEX IF EXISTS idx_unique_full_day_slots;
DROP TABLE IF EXISTS service_available_slots;
DROP TABLE IF EXISTS services;
