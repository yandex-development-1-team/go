-- +goose Up

CREATE TABLE IF NOT EXISTS settings_messages (
                                        id SERIAL PRIMARY KEY,
                                        key VARCHAR(100) NOT NULL,
                                        value TEXT NOT NULL DEFAULT '',
                                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                        UNIQUE(key)
);

-- Вставляем дефолтные значения
INSERT INTO settings_messages (key, value) VALUES
                                                ('welcome_message', ''),
                                                ('record_confirmation', ''),
                                                ('event_reminder_for_week', ''),
                                                ('event_reminder_for_24_hours', ''),
                                                ('cancellation_message', ''),
                                                ('thanks_message', ''),
                                                ('system_err_message', '')
ON CONFLICT (key)
    DO UPDATE SET
                  value = EXCLUDED.value,
                  updated_at = NOW();

-- +goose Down

DROP TABLE IF EXISTS settings_messages;
