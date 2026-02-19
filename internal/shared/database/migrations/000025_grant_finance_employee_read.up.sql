INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r
JOIN permissions p ON p.resource = 'employee' AND p.action = 'read'
WHERE LOWER(r.name) = 'finance'
ON CONFLICT DO NOTHING;
