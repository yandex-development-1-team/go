-- +goose Up
CREATE TYPE setting_category_type AS ENUM ('general', 'booking', 'notifications');

CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    category setting_category_type NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(category, key)
);

-- Вставляем дефолтные значения
INSERT INTO settings (category, key, value) VALUES
                                                ('notifications', 'telegram_bot_token', ''),
                                                ('notifications', 'auto_reminders', true),
                                                ('notifications', 'reminder_hours_before', 24),
                                                ('booking', 'max_slots_per_event', 10),
                                                ('booking', 'allow_overbooking', false),
                                                ('booking', 'cancellation_allowed_hours', 24),
                                                ('general', 'site_name', ''),
                                                ('general', 'contact_email', ''),
                                                ('general', 'contact_phone', '');

-- +goose Down
DROP TABLE IF EXISTS settings;

DROP TYPE IF EXISTS setting_category_type;
