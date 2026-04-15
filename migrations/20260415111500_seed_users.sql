-- +goose Up
-- +goose StatementBegin
INSERT INTO users (telegram_id, username, first_name, last_name, grade, is_admin, password_hash, role, status, invite_token, permissions, email) VALUES
(100001, 'ivan_petrov', 'Иван', 'Петров', 0, false, '', 'user', 'active', null, '{}', 'ivan@example.com'),
(100002, 'maria_sidorova', 'Мария', 'Сидорова', 0, false, '', 'user', 'active', null, '{}', 'maria@example.com'),
(100003, 'petr_ivanov', 'Пётр', 'Иванов', 0, false, '', 'user', 'active', null, '{}', 'petr@example.com'),
(100004, 'anna_kuznetsova', 'Анна', 'Кузнецова', 0, false, '', 'user', 'active', null, '{}', 'anna@example.com'),
(100005, 'sergey_novikov', 'Сергей', 'Новиков', 0, false, '', 'user', 'active', null, '{}', 'sergey@example.com'),
(100006, 'elena_morozova', 'Елена', 'Морозова', 0, false, '', 'user', 'active', null, '{}', 'elena@example.com'),
(100007, 'dmitry_volkov', 'Дмитрий', 'Волков', 0, false, '', 'user', 'active', null, '{}', 'dmitry@example.com'),
(100008, 'olga_sokolova', 'Ольга', 'Соколова', 0, false, '', 'user', 'active', null, '{}', 'olga@example.com'),
(100009, 'alexey_popov', 'Алексей', 'Попов', 0, false, '', 'user', 'active', null, '{}', 'alexey@example.com'),
(100010, 'natalia_lebedev', 'Наталья', 'Лебедева', 0, false, '', 'user', 'active', null, '{}', 'natalia@example.com')
ON CONFLICT (telegram_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE telegram_id IN (100001,100002,100003,100004,100005,100006,100007,100008,100009,100010);
-- +goose StatementEnd