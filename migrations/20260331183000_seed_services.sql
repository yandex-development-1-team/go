-- +goose Up
-- +goose StatementBegin
INSERT INTO services (name, slug, description, rules, location, price, status, organizer) VALUES
('Новогодний квест', 'new-year-quest', 'Новогоднее приключение для всей семьи', 'Правила новогоднего квеста', 'Москва', 1000, 'active', 'Организатор 1'),
('Летний квест', 'summer-quest', 'Летнее приключение на свежем воздухе', 'Правила летнего квеста', 'Сочи', 1500, 'active', 'Организатор 2'),
('Хэллоуин квест', 'halloween-quest', 'Страшное приключение в ночь всех святых', 'Правила хэллоуин квеста', 'СПб', 2000, 'inactive', 'Организатор 3'),
('Корпоративный тимбилдинг', 'corp-teambuilding', 'Командообразующие мероприятия для бизнеса', 'Правила тимбилдинга', 'Москва', 5000, 'active', 'Организатор 1'),
('Детский день рождения', 'kids-birthday', 'Незабываемый праздник для детей', 'Правила детского праздника', 'Москва', 3000, 'active', 'Организатор 2'),
('Мастер-класс по кулинарии', 'cooking-masterclass', 'Научитесь готовить с профессиональным шеф-поваром', 'Правила мастер-класса', 'Москва', 2500, 'active', 'Организатор 4'),
('Фотосессия в студии', 'photo-studio', 'Профессиональная фотосессия в оборудованной студии', 'Правила фотосессии', 'СПб', 4000, 'inactive', 'Организатор 5'),
('Винный вечер', 'wine-evening', 'Дегустация вин с сомелье', 'Правила винного вечера', 'Москва', 3500, 'active', 'Организатор 3'),
('Йога на природе', 'yoga-nature', 'Утренняя йога на свежем воздухе', 'Правила йоги', 'Сочи', 800, 'active', 'Организатор 6'),
('Квиз для компаний', 'quiz-night', 'Интеллектуальная игра для команд', 'Правила квиза', 'Москва', 1200, 'active', 'Организатор 2'),
('Спа-день', 'spa-day', 'Полный день релаксации и ухода', 'Правила спа', 'Москва', 6000, 'inactive', 'Организатор 7'),
('Экскурсия по городу', 'city-tour', 'Авторская экскурсия с гидом', 'Правила экскурсии', 'СПб', 900, 'active', 'Организатор 8'),
('Кулинарный батл', 'cooking-battle', 'Соревнование команд в приготовлении блюд', 'Правила батла', 'Москва', 4500, 'active', 'Организатор 4'),
('Арт-мастерская', 'art-workshop', 'Рисование под руководством художника', 'Правила мастерской', 'Москва', 2000, 'active', 'Организатор 9'),
('Конференц-зал', 'conference-hall', 'Аренда зала для деловых мероприятий', 'Правила аренды', 'Москва', 10000, 'inactive', 'Организатор 1'),
('Пикник на природе', 'picnic', 'Организованный пикник с развлечениями', 'Правила пикника', 'Подмосковье', 2200, 'active', 'Организатор 6'),
('Ночная фотоохота', 'night-photo', 'Фотографирование ночного города', 'Правила фотоохоты', 'Москва', 1800, 'active', 'Организатор 5'),
('Гончарная мастерская', 'pottery', 'Лепка из глины с мастером', 'Правила гончарки', 'СПб', 2800, 'active', 'Организатор 9'),
('Скалодром для новичков', 'climbing', 'Введение в скалолазание', 'Правила скалодрома', 'Москва', 1600, 'inactive', 'Организатор 10'),
('Джазовый вечер', 'jazz-evening', 'Живая джазовая музыка и ужин', 'Правила вечера', 'Москва', 3800, 'active', 'Организатор 3'),
('Кинопоказ под открытым небом', 'open-air-cinema', 'Кино на свежем воздухе', 'Правила кинопоказа', 'Сочи', 700, 'active', 'Организатор 8'),
('Настольные игры', 'board-games', 'Вечер настольных игр для компании', 'Правила игр', 'Москва', 500, 'active', 'Организатор 2'),
('Танцевальный мастер-класс', 'dance-class', 'Латиноамериканские танцы для начинающих', 'Правила занятия', 'Москва', 1400, 'inactive', 'Организатор 11'),
('Поход выходного дня', 'weekend-hike', 'Однодневный поход в горы', 'Правила похода', 'Сочи', 2600, 'active', 'Организатор 6'),
('Кофейная дегустация', 'coffee-tasting', 'Знакомство с миром specialty кофе', 'Правила дегустации', 'Москва', 1100, 'active', 'Организатор 12');
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time)
SELECT s.id, v.slot_date::date, v.start_time::time, v.end_time::time
FROM services s
JOIN (VALUES
    ('new-year-quest',    '2026-06-01', '10:00', '12:00'),
    ('new-year-quest',    '2026-06-01', '14:00', '16:00'),
    ('new-year-quest',    '2026-06-02', '11:00', '13:00'),
    ('summer-quest',      '2026-07-15', '09:00', '18:00'),
    ('corp-teambuilding', '2026-05-20', '10:00', '18:00'),
    ('corp-teambuilding', '2026-05-21', '10:00', '18:00'),
    ('kids-birthday',     '2026-06-10', '12:00', '15:00'),
    ('kids-birthday',     '2026-06-15', '12:00', '15:00'),
    ('cooking-masterclass','2026-06-05', '18:00', '21:00'),
    ('wine-evening',      '2026-06-12', '19:00', '22:00'),
    ('wine-evening',      '2026-06-19', '19:00', '22:00'),
    ('yoga-nature',       '2026-06-01', '08:00', '09:30'),
    ('yoga-nature',       '2026-06-08', '08:00', '09:30'),
    ('quiz-night',        '2026-06-03', '19:00', '22:00'),
    ('city-tour',         '2026-06-07', '10:00', '13:00'),
    ('city-tour',         '2026-06-14', '10:00', '13:00'),
    ('cooking-battle',    '2026-06-20', '14:00', '18:00'),
    ('art-workshop',      '2026-06-06', '15:00', '18:00'),
    ('art-workshop',      '2026-06-13', '15:00', '18:00'),
    ('picnic',            '2026-06-08', '12:00', '17:00'),
    ('night-photo',       '2026-06-11', '22:00', '01:00'),
    ('pottery',           '2026-06-04', '11:00', '14:00'),
    ('jazz-evening',      '2026-06-18', '19:00', '23:00'),
    ('open-air-cinema',   '2026-06-25', '21:00', '23:30'),
    ('board-games',       '2026-06-05', '18:00', '22:00'),
    ('weekend-hike',      '2026-06-14', '08:00', '18:00'),
    ('coffee-tasting',    '2026-06-09', '11:00', '13:00')
) AS v(slug, slot_date, start_time, end_time) ON s.slug = v.slug
ON CONFLICT (service_id, slot_date, start_time, end_time) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM service_available_slots WHERE service_id IN (
    SELECT id FROM services WHERE slug IN (
        'new-year-quest','summer-quest','halloween-quest','corp-teambuilding',
        'kids-birthday','cooking-masterclass','photo-studio','wine-evening',
        'yoga-nature','quiz-night','spa-day','city-tour','cooking-battle',
        'art-workshop','conference-hall','picnic','night-photo','pottery',
        'climbing','jazz-evening','open-air-cinema','board-games',
        'dance-class','weekend-hike','coffee-tasting'
    )
);
DELETE FROM services WHERE slug IN (
    'new-year-quest','summer-quest','halloween-quest','corp-teambuilding',
    'kids-birthday','cooking-masterclass','photo-studio','wine-evening',
    'yoga-nature','quiz-night','spa-day','city-tour','cooking-battle',
    'art-workshop','conference-hall','picnic','night-photo','pottery',
    'climbing','jazz-evening','open-air-cinema','board-games',
    'dance-class','weekend-hike','coffee-tasting'
);
-- +goose StatementEnd