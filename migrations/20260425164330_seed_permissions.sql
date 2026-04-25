-- +goose Up
INSERT INTO role_permissions (role, permissions) VALUES
                                                     ('admin', ARRAY['bookings:view', 'bookings:edit', 'bookings:delete', 'boxes:create', 'boxes:edit', 'boxes:delete',
                                                         'specproject:view', 'specproject:edit', 'specproject:delete', 'presentation:view', 'presentation:edit',
                                                         'presentation:delete', 'analytics:view', 'analytics:download', 'poster:yes', 'aboutus:yes', 'faq:yes']),

                                                     ('manager_1', ARRAY['bookings:view', 'bookings:edit', 'boxes:create', 'boxes:edit',
                                                         'specproject:view', 'specproject:edit', 'presentation:view', 'presentation:edit',
                                                         'analytics:view', 'poster:yes', 'aboutus:yes', 'faq:yes']),

                                                     ('manager_2', ARRAY['bookings:view', 'bookings:edit', 'boxes:create', 'boxes:edit',
                                                         'specproject:view', 'specproject:edit', 'presentation:view', 'presentation:edit',
                                                         'poster:yes', 'faq:yes']),

                                                     ('manager_3', ARRAY['bookings:view', 'boxes:create', 'specproject:view', 'presentation:view'])
ON CONFLICT (role)
    DO UPDATE SET permissions = EXCLUDED.permissions;

