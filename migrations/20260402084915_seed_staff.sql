-- +goose Up
-- +goose StatementBegin
INSERT INTO staff (
    telegram_nick, first_name, last_name, second_name, email, phone_number,
    password_hash, role, status, department, position, invite_token,
    permissions, created_at, updated_at
) VALUES (
    'admin', 'Admin', 'Adminov', '', 'admin@example.com', '543244',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'admin', 'active', '', '', '', '{}',
    NOW(), NOW()
) ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO staff (
    telegram_nick, first_name, last_name, second_name, email, phone_number,
    password_hash, role, status, department, position, invite_token,
    permissions, created_at, updated_at
) VALUES (
    'Manager_1', 'Manager_1', 'Managerov', '', 'manager1@example.com', '232323',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'manager_1', 'active', '', '', '', '{}',
    NOW(), NOW()
) ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO staff (
    telegram_nick, first_name, last_name, second_name, email, phone_number,
    password_hash, role, status, department, position, invite_token,
    permissions, created_at, updated_at
) VALUES (
    'Manager_2', 'Manager_2', 'Managerov', '', 'manager2@example.com', '532523432656',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'manager_2', 'active', '', '', '', '{}',
    NOW(), NOW()
) ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM staff WHERE email = 'admin@example.com';
-- +goose StatementEnd