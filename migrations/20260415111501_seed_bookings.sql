-- +goose Up
-- +goose StatementBegin
INSERT INTO bookings (user_id, service_id, booking_date, booking_time, guest_name, guest_organization, guest_position, status, manager_id) VALUES
(100001, 1, '2026-05-01', '10:00', 'Иван Петров', 'ООО Рога', 'Директор', 'pending',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1)),
(100002, 1, '2026-05-01', '14:00', 'Мария Сидорова', 'ИП Сидорова', 'Менеджер', 'confirmed',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1)),
(100003, 2, '2026-05-10', '09:00', 'Пётр Иванов', 'АО Копыта', 'Аналитик', 'pending',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1)),
(100004, 2, '2026-05-10', '11:00', 'Анна Кузнецова', 'ООО Луна', 'Бухгалтер', 'cancelled',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1)),
(100005, 1, '2026-05-15', '10:00', 'Сергей Новиков', 'ЗАО Звезда', 'CTO', 'confirmed',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1)),
(100006, 3, '2026-06-01', NULL, 'Елена Морозова', 'ООО Облако', 'HR', 'pending',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1)),
(100007, 1, '2026-06-05', '14:00', 'Дмитрий Волков', 'ИП Волков', 'CEO', 'pending',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1)),
(100008, 2, '2026-06-10', '09:00', 'Ольга Соколова', 'АО Солнце', 'CFO', 'confirmed',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1)),
(100009, 3, '2026-06-15', NULL, 'Алексей Попов', 'ООО Ветер', 'PM', 'pending',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1)),
(100010, 1, '2026-06-20', '11:00', 'Наталья Лебедева', 'ЗАО Море', 'CMO', 'cancelled',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1))
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM bookings WHERE user_id IN (100001,100002,100003,100004,100005,100006,100007,100008,100009,100010);
-- +goose StatementEnd