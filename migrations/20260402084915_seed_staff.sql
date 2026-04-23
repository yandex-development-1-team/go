-- +goose Up
-- +goose StatementBegin
INSERT INTO staff (telegram_nick, first_name, last_name, second_name, email, phone_number, password_hash, role, status, department, position, created_at, updated_at) VALUES
('admin', 'Admin', 'Adminov', '', 'admin@example.com', '543244', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', 'Руководство', 'Администратор', NOW(), NOW()),
('manager_1', 'Ольга', 'Дмитриева', 'Сергеевна', 'manager1@example.com', '79161234567', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_1', 'active', 'Продажи', 'Менеджер по работе с клиентами', NOW(), NOW()),
('manager_2', 'Иван', 'Соколов', 'Андреевич', 'manager2@example.com', '79162345678', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_2', 'active', 'Продажи', 'Старший менеджер', NOW(), NOW()),
('manager_3', 'Анна', 'Козлова', 'Викторовна', 'manager3@example.com', '79163456789', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_3', 'active', 'Продажи', 'Менеджер по спецпроектам', NOW(), NOW()),
('manager_4', 'Дмитрий', 'Волков', 'Павлович', 'manager4@example.com', '79164567890', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_1', 'active', 'Продажи', 'Менеджер по работе с клиентами', NOW(), NOW()),
('manager_5', 'Елена', 'Новикова', 'Игоревна', 'manager5@example.com', '79165678901', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_2', 'active', 'Продажи', 'Старший менеджер', NOW(), NOW()),
('manager_6', 'Сергей', 'Морозов', 'Алексеевич', 'manager6@example.com', '79166789012', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_3', 'active', 'Продажи', 'Менеджер по спецпроектам', NOW(), NOW()),
('manager_7', 'Наталья', 'Попова', 'Дмитриевна', 'manager7@example.com', '79167890123', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager_1', 'blocked', 'Продажи', 'Менеджер по работе с клиентами', NOW(), NOW()),
('user_1', 'Михаил', 'Лебедев', 'Романович', 'user1@example.com', '79168901234', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'active', 'Маркетинг', 'Специалист', NOW(), NOW()),
('user_2', 'Татьяна', 'Александрова', 'Юрьевна', 'user2@example.com', '79169012345', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'invited', 'HR', 'HR-специалист', NOW(), NOW()),
('my_admin', 'Sergey','Klo','sfdfd', 'kdfsdsfdsfdsfsdfs@yandex.com', '+7949', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', 'Маркетинг', 'Специалист', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM staff WHERE email IN (
    'admin@example.com', 'manager1@example.com', 'manager2@example.com',
    'manager3@example.com', 'manager4@example.com', 'manager5@example.com',
    'manager6@example.com', 'manager7@example.com', 'user1@example.com',
    'user2@example.com'
);
-- +goose StatementEnd