-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE event_status AS ENUM ('active', 'cancelled', 'completed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS events (
    id             BIGSERIAL PRIMARY KEY,
    box_id         BIGINT NOT NULL REFERENCES boxes(id) ON DELETE CASCADE,
    event_date     DATE NOT NULL,
    event_time     TIME NOT NULL,
    total_slots    INTEGER NOT NULL DEFAULT 1,
    occupied_slots INTEGER NOT NULL DEFAULT 0,
    status         event_status NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_box_date ON events(box_id, event_date);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_status;