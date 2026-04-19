-- +goose Up
ALTER TABLE staff
    ADD COLUMN IF NOT EXISTS supervisor VARCHAR(255),
    ADD COLUMN IF NOT EXISTS address VARCHAR(255);

-- +goose Down
ALTER TABLE staff
    DROP COLUMN IF EXISTS supervisor,
    DROP COLUMN IF EXISTS address;