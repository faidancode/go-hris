-- Bootstrap SUPERADMIN for demo flow.
-- Idempotent and safe to run multiple times.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Ensure base permissions exist (in case older seed did not insert completely).
INSERT INTO permissions (id, resource, action)
VALUES
    (gen_random_uuid(), 'employee', 'read'),
    (gen_random_uuid(), 'employee', 'create'),
    (gen_random_uuid(), 'employee', 'update'),
    (gen_random_uuid(), 'employee', 'delete'),
    (gen_random_uuid(), 'department', 'read'),
    (gen_random_uuid(), 'department', 'create'),
    (gen_random_uuid(), 'department', 'update'),
    (gen_random_uuid(), 'department', 'delete'),
    (gen_random_uuid(), 'position', 'read'),
    (gen_random_uuid(), 'position', 'create'),
    (gen_random_uuid(), 'position', 'update'),
    (gen_random_uuid(), 'position', 'delete'),
    (gen_random_uuid(), 'salary', 'read'),
    (gen_random_uuid(), 'salary', 'update'),
    (gen_random_uuid(), 'payroll', 'read'),
    (gen_random_uuid(), 'payroll', 'create'),
    (gen_random_uuid(), 'payroll', 'approve'),
    (gen_random_uuid(), 'payroll', 'pay'),
    (gen_random_uuid(), 'payroll', 'delete'),
    (gen_random_uuid(), 'leave', 'read'),
    (gen_random_uuid(), 'leave', 'create'),
    (gen_random_uuid(), 'leave', 'approve'),
    (gen_random_uuid(), 'leave', 'manage'),
    (gen_random_uuid(), 'role', 'read'),
    (gen_random_uuid(), 'role', 'manage'),
    (gen_random_uuid(), 'company', 'read'),
    (gen_random_uuid(), 'company', 'update'),
    (gen_random_uuid(), 'attendance', 'read'),
    (gen_random_uuid(), 'attendance', 'manage')
ON CONFLICT (resource, action) DO NOTHING;

-- Create SUPERADMIN role for each company.
INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT
    gen_random_uuid(),
    c.id,
    'SUPERADMIN',
    'Bootstrap full-access role for development/demo flow',
    now(),
    now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

-- Grant all permissions to SUPERADMIN.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    r.id,
    p.id,
    now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'SUPERADMIN'
ON CONFLICT DO NOTHING;

-- Assign SUPERADMIN to all demo-company employees so frontend flow can run end-to-end.
INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT
    e.id,
    r.id,
    now()
FROM employees e
JOIN companies c
    ON c.id = e.company_id
JOIN roles r
    ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND r.name = 'SUPERADMIN'
  AND e.deleted_at IS NULL
ON CONFLICT DO NOTHING;
