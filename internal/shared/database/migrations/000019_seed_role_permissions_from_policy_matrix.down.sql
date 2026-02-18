-- Roll back to previous baseline role mapping (before policy-matrix alignment).

DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE name IN ('Owner', 'HR', 'Finance', 'Employee', 'SUPERADMIN')
);

-- Owner: full access.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- HR baseline (legacy)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    (p.resource IN ('employee', 'department', 'position', 'attendance')) OR
    (p.resource = 'leave' AND p.action IN ('read', 'approve', 'manage')) OR
    (p.resource = 'payroll' AND p.action IN ('read', 'create', 'delete'))
)
WHERE r.name = 'HR'
ON CONFLICT DO NOTHING;

-- Finance baseline (legacy)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    (p.resource = 'salary') OR
    (p.resource = 'payroll' AND p.action IN ('read', 'approve', 'pay'))
)
WHERE r.name = 'Finance'
ON CONFLICT DO NOTHING;

-- Employee baseline (legacy)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    (p.resource = 'employee' AND p.action = 'read') OR
    (p.resource = 'payroll' AND p.action = 'read') OR
    (p.resource = 'leave' AND p.action IN ('read', 'create'))
)
WHERE r.name = 'Employee'
ON CONFLICT DO NOTHING;
