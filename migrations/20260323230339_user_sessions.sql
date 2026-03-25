-- +goose Up
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    current_state VARCHAR(50),  -- 'main_menu', 'booking_form', 'service_detail', etc
    state_data JSONB,  -- сохранять любые данные формы
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);

-- +goose Down
DROP TABLE IF EXISTS user_sessions;