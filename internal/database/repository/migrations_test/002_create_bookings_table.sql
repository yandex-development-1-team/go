-- Migration: 002_create_bookings_table
-- Создание таблицы бронирований для управления заказами услуг через телеграм-бот

-- +goose Up
CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled');

CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id SMALLINT NOT NULL,
    booking_date DATE NOT NULL,
    booking_time TIME,
    guest_name VARCHAR(255) NOT NULL,
    guest_organization VARCHAR(255),
    guest_position VARCHAR(255),
    visit_type VARCHAR(50),
    status booking_status DEFAULT 'pending',
    tracker_ticket_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_service_id ON bookings(service_id);
CREATE INDEX idx_bookings_status ON bookings(status);
