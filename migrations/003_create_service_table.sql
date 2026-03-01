-- +goose Up
-- Создание таблицы хранения коробочных решений

CREATE TABLE IF NOT EXISTS services (
                                             id BIGSERIAL PRIMARY KEY,
                                             name VARCHAR(255) NOT NULL,
                                             description TEXT NULL,
                                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS available_slots (
                                               service_id BIGINT NOT NULL,
                                               slot_date DATE NOT NULL,
                                               time_slots TEXT[] NOT NULL DEFAULT '{}',
                                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                               CONSTRAINT fk_available_slots_service
                                                   FOREIGN KEY (service_id)
                                                       REFERENCES services (id)
                                                       ON DELETE CASCADE,

                                               CONSTRAINT uq_available_slots_service_date
                                                   UNIQUE (service_id, slot_date)
);

CREATE INDEX IF NOT EXISTS idx_available_slots_service_date
    ON available_slots (service_id, slot_date);

-- +goose Down
-- Удаление таблиц коробочных решений
DROP TABLE IF EXISTS available_slots;
DROP TABLE IF EXISTS services;
