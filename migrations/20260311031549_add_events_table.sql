-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_status') THEN
        CREATE TYPE event_status AS ENUM ('active', 'cancelled', 'completed');
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS events (
    id             BIGSERIAL PRIMARY KEY,
    box_id         BIGINT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    event_date     DATE NOT NULL,
    event_time     TIME NOT NULL,
    total_slots    INTEGER NOT NULL DEFAULT 1,
    occupied_slots INTEGER NOT NULL DEFAULT 0,
    status         event_status NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Связываем старую таблицу bookings с новыми событиями
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS event_id BIGINT REFERENCES events(id) ON DELETE SET NULL;

-- Индексы для быстрой фильтрации по ТЗ (box_id + диапазон дат)
CREATE INDEX IF NOT EXISTS idx_events_box_date ON events(box_id, event_date);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);

-- +goose Down
ALTER TABLE bookings DROP COLUMN IF EXISTS event_id;
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_status;