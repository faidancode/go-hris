-- Seed user module permissions (idempotent).
INSERT INTO permissions (id, resource, action, label, category)
VALUES
    (gen_random_uuid(), 'user', 'read', 'Melihat User', 'User Management'),
    (gen_random_uuid(), 'user', 'create', 'Tambah User', 'User Management'),
    (gen_random_uuid(), 'user', 'update', 'Edit User', 'User Management')
ON CONFLICT (resource, action) DO UPDATE
SET
    label = EXCLUDED.label,
    category = EXCLUDED.category;

-- Map user permissions to privileged tenant roles.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON p.resource = 'user' AND p.action IN ('read', 'create', 'update')
WHERE UPPER(r.name) IN ('SUPERADMIN', 'ADMIN', 'HR', 'OWNER')
ON CONFLICT DO NOTHING;
