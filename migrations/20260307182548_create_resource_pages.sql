-- +goose Up
CREATE TABLE IF NOT EXISTS resource_pages (
    slug VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    links JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE INDEX IF NOT EXISTS idx_resource_pages_updated_at ON resource_pages (updated_at);

INSERT INTO resource_pages (slug, title, content) VALUES 
('organizationInfo', 'Организационная информация', 'Текст о нас'),
('usefulLinks', 'Полезные ссылки', ''),
('faq', 'FAQ', '')
('eventSchedule', 'Афиша Pertner Relations', '');


-- +goose Down
DROP TABLE IF EXISTS resourse_pages

