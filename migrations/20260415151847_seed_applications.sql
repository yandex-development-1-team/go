-- +goose Up
-- +goose StatementBegin
INSERT INTO applications (status, manager_id, form_answer_id, customer_name, contact_info, description) VALUES
('pending',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1),
    'form-001', 'Алексей Морозов', '@alex_morozov',
    'Хотим организовать корпоратив на 50 человек с квестом и банкетом'),
('confirmed',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1),
    'form-002', 'Мария Петрова', '@maria_petrova',
    'Нужен тимбилдинг для команды разработчиков, около 20 человек'),
('pending',
    (SELECT id FROM staff WHERE role = 'manager_3' LIMIT 1),
    'form-003', 'Дмитрий Захаров', '@dmitry_z',
    'Детский день рождения на 15 детей, возраст 7-10 лет'),
('cancelled',
    (SELECT id FROM staff WHERE role = 'manager_1' LIMIT 1),
    'form-004', 'Екатерина Смирнова', '@kate_smirnova',
    'Свадебный банкет на 100 человек, нужно оформление зала'),
('pending',
    (SELECT id FROM staff WHERE role = 'manager_2' LIMIT 1),
    'form-005', 'Николай Фёдоров', '@nikolay_fed',
    'Конференция для партнёров, около 80 участников, нужна техническая поддержка');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM applications WHERE form_answer_id IN ('form-001','form-002','form-003','form-004','form-005');
-- +goose StatementEnd