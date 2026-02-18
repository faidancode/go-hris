-- Add attendance:create for clock-in/clock-out flow.
INSERT INTO permissions (id, resource, action)
VALUES (gen_random_uuid(), 'attendance', 'create')
ON CONFLICT (resource, action) DO NOTHING;

-- Owner can create attendance records.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON p.resource = 'attendance' AND p.action = 'create'
WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- SUPERADMIN can create attendance records.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON p.resource = 'attendance' AND p.action = 'create'
WHERE r.name = 'SUPERADMIN'
ON CONFLICT DO NOTHING;

-- HR can create attendance records.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON p.resource = 'attendance' AND p.action = 'create'
WHERE r.name = 'HR'
ON CONFLICT DO NOTHING;

-- Employee can clock-in/clock-out (self-service).
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON p.resource = 'attendance' AND p.action = 'create'
WHERE r.name = 'Employee'
ON CONFLICT DO NOTHING;
