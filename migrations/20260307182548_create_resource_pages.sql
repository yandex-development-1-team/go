-- +goose Up
CREATE TABLE IF NOT EXISTS resource_pages (
    slug VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    links JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_pages_updated_at ON resource_pages (updated_at);

INSERT INTO resource_pages (slug, title, content) VALUES 
('organizationInfo', 'Организационная информация', 'Компания полного цикла в организации мероприятий — от корпоративов и конференций до партнёрских программ и спецпроектов. Работает с крупными клиентами, имеет собственную команду менеджеров и отлаженные процессы под любой масштаб события.'),
('usefulLinks', 'Полезные ссылки', ''),
('faq', 'FAQ', 'https://disk.yandex.ru/i/nWbTdaWA9Xcgdg'),
('eventSchedule', 'Афиша Pertner Relations', 'https://afisha.yandex.ru');


-- +goose Down
DROP TABLE IF EXISTS resourse_pages;

