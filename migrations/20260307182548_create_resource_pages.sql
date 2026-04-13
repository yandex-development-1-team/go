-- +goose Up
CREATE TABLE IF NOT EXISTS resource_pages (
    slug VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    links JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_pages_updated_at ON resource_pages (updated_at);

INSERT INTO resource_pages (slug, title, content) VALUES 
('org-info', 'Организационная информация', 'Компания полного цикла в организации мероприятий — от корпоративов и конференций до партнёрских программ и спецпроектов. Работает с крупными клиентами, имеет собственную команду менеджеров и отлаженные процессы под любой масштаб события.'),
('useful-links', 'Полезные ссылки', 'Ниже находятся ссылки на наши каналы/блоги и другие ресурсы'),
('faq', 'FAQ', 'https://disk.yandex.ru/i/nWbTdaWA9Xcgdg'),
('spec-projects', 'Примеры спецпроектов', 'https://disk.yandex.ru/i/nWbTdaWA9Xcgdg'),
('req-spec-projects', 'Запрос спецпроекта', 'https://forms.yandex.ru/u/69db858549af478aee1a0d15/?page=1'),
('event-schedule', 'Афиша Pertner Relations', '');


-- Добавляем ссылки для org-info
UPDATE resource_pages 
SET links = '[
    {"id": "org_1", "title": "О компании", "url": "https://example.com/about"},
    {"id": "org_2", "title": "Наши проекты", "url": "https://example.com/projects"},
    {"id": "org_3", "title": "Команда", "url": "https://example.com/team"}
]'::jsonb
WHERE slug = 'org-info';

-- Добавляем ссылки для useful-links
UPDATE resource_pages 
SET links = '[
    {"id": "link_1", "title": "Документация", "url": "https://example.com/docs"},
    {"id": "link_2", "title": "Блог", "url": "https://example.com/blog"},
    {"id": "link_3", "title": "YouTube канал", "url": "https://youtube.com/example"},
    {"id": "link_4", "title": "Telegram канал", "url": "https://t.me/example"}
]'::jsonb
WHERE slug = 'useful-links';

-- Добавляем ссылки для faq
UPDATE resource_pages 
SET links = '[
    {"id": "faq_1", "title": "Часто задаваемые вопросы", "url": "https://disk.yandex.ru/i/nWbTdaWA9Xcgdg"},
    {"id": "faq_2", "title": "Видеоинструкция", "url": "https://example.com/video"}
]'::jsonb
WHERE slug = 'faq';

-- Добавляем ссылки для spec-projects
UPDATE resource_pages 
SET links = '[
    {"id": "proj_1", "title": "Кейс: Конференция 2024", "url": "https://example.com/case1"},
    {"id": "proj_2", "title": "Кейс: Корпоратив", "url": "https://example.com/case2"},
    {"id": "proj_3", "title": "Портфолио", "url": "https://example.com/portfolio"},
    {"id": "proj_4", "title": "Примеры работ", "url": "https://disk.yandex.ru/i/nWbTdaWA9Xcgdg"}
]'::jsonb
WHERE slug = 'spec-projects';

-- Добавляем ссылки для event-schedule
UPDATE resource_pages 
SET links = '[
    {"id": "schedule_1", "title": "Расписание мероприятий", "url": "https://example.com/schedule"},
    {"id": "schedule_2", "title": "Календарь", "url": "https://example.com/calendar"},
    {"id": "schedule_3", "title": "Регистрация", "url": "https://example.com/register"}
]'::jsonb
WHERE slug = 'event-schedule';


-- +goose Down
DROP TABLE IF EXISTS resource_pages;

