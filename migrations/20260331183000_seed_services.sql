-- +goose Up
-- +goose StatementBegin
-- вставка тестовых сервисов
INSERT INTO services (name, slug, description, rules, location, price, status, organizer) VALUES 
('Новогодний квест', 'new-year-quest', 'Новогоднее приключение', 'Правила новогоднего квеста', 'Москва', 1000, 'active', 'Организатор 1'),
('Летний квест', 'summer-quest', 'Летнее приключение', 'Правила летнего квеста', 'Сочи', 1500, 'active', 'Организатор 2'),
('Хэллоуин квест', 'halloween-quest', 'Страшное приключение', 'Правила хэллоуин квеста', 'СПб', 2000, 'inactive', 'Организатор 3');

-- вставка слотов
INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time) VALUES 
(1, '2024-06-01', '10:00', '12:00'),
(1, '2024-06-01', '14:00', '16:00'),
(1, '2024-06-02', '11:00', '13:00'),
(2, '2024-07-15', '09:00', '18:00')
ON CONFLICT (service_id, slot_date, start_time, end_time) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM service_available_slots WHERE service_id IN (1,2,3);
DELETE FROM services WHERE slug IN ('new-year-quest', 'summer-quest', 'halloween-quest');
-- +goose StatementEnd