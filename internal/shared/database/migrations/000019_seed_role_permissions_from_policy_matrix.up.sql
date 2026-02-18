-- Align role_permissions to the RBAC policy matrix.
-- Applies to default role names across companies.

-- 1) Reset mappings for managed roles.
DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE name IN ('Owner', 'HR', 'Finance', 'Employee', 'SUPERADMIN')
);

-- 2) Owner: full access to all existing permissions.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- 3) SUPERADMIN: full access to all existing permissions.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'SUPERADMIN'
ON CONFLICT DO NOTHING;

-- 4) HR
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    -- Core people & structure
    (p.resource = 'employee'   AND p.action IN ('read', 'create', 'update', 'delete')) OR
    (p.resource = 'department' AND p.action IN ('read', 'create', 'update', 'delete')) OR
    (p.resource = 'position'   AND p.action IN ('read', 'create', 'update', 'delete')) OR
    -- Leave workflow
    (p.resource = 'leave'      AND p.action IN ('read', 'create', 'approve', 'manage')) OR
    -- Payroll prep
    (p.resource = 'payroll'    AND p.action IN ('read', 'create', 'delete')) OR
    -- Attendance
    (p.resource = 'attendance' AND p.action IN ('read', 'manage')) OR
    -- Company visibility
    (p.resource = 'company'    AND p.action IN ('read'))
)
WHERE r.name = 'HR'
ON CONFLICT DO NOTHING;

-- 5) Finance
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    (p.resource = 'salary'  AND p.action IN ('read', 'update')) OR
    (p.resource = 'payroll' AND p.action IN ('read', 'approve', 'pay')) OR
    (p.resource = 'company' AND p.action IN ('read'))
)
WHERE r.name = 'Finance'
ON CONFLICT DO NOTHING;

-- 6) Employee (self-service oriented; ownership checks remain in service layer).
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    (p.resource = 'employee'   AND p.action IN ('read')) OR
    (p.resource = 'payroll'    AND p.action IN ('read')) OR
    (p.resource = 'leave'      AND p.action IN ('read', 'create')) OR
    (p.resource = 'attendance' AND p.action IN ('read'))
)
WHERE r.name = 'Employee'
ON CONFLICT DO NOTHING;
