-- +goose Up
CREATE TABLE IF NOT EXISTS user_favorites (
    user_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_favorites_user
        FOREIGN KEY (user_id)
            REFERENCES staff(id)
            ON DELETE CASCADE,

    CONSTRAINT fk_user_favorites_service
        FOREIGN KEY (service_id)
            REFERENCES services(id)
            ON DELETE CASCADE,

    CONSTRAINT uq_user_favorites_user_service
        UNIQUE (user_id, service_id)
);

CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_favorites_user_id;
DROP TABLE IF EXISTS user_favorites;
