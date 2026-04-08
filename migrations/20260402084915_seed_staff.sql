-- +goose Up
-- +goose StatementBegin
INSERT INTO staff (
    telegram_nick, first_name, last_name, second_name, email, phone_number,
    password_hash, role, status, department, position, manager_id, invite_token,
    permissions, created_at, updated_at
) VALUES (
    'admin', 'Admin', 'Adminov', '', 'admin@example.com', '',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'admin', 'active', '', '', 4, '', '{}',
    NOW(), NOW()
) ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM staff WHERE email = 'admin@example.com';
-- +goose StatementEnd