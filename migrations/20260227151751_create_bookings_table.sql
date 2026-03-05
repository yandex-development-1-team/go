-- +goose Up

-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    booking_date DATE NOT NULL,
    booking_time TIME,
    guest_name VARCHAR(255) NOT NULL,
    guest_organization VARCHAR(255),
    guest_position VARCHAR(255),
    visit_type VARCHAR(50),
    status booking_status DEFAULT 'pending',
    tracker_ticket_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_bookings_user
        FOREIGN KEY (user_id)
            REFERENCES users(id)
            ON DELETE CASCADE,

    CONSTRAINT fk_bookings_service
        FOREIGN KEY (service_id)
            REFERENCES services(id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_service_id ON bookings(service_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);

-- +goose Down
DROP TABLE IF EXISTS bookings;
DROP TYPE IF EXISTS booking_status;
