-- +goose Up

CREATE TABLE IF NOT EXISTS role_permissions  (
    role        user_role_type PRIMARY KEY,
    permissions TEXT[] NOT NULL DEFAULT '{}'
);

INSERT INTO role_permissions (role, permissions) VALUES
    ('manager_1', '{"bookings:view","bookings:edit","specproject:view","analytics:view"}'),
    ('manager_2', '{"bookings:view"}'),
    ('manager_3', '{"bookings:view"}')
ON CONFLICT (role) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS role_permissions;