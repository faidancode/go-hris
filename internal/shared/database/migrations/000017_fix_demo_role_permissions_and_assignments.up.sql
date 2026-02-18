-- Fix and bootstrap default RBAC role mappings for demo flow.
-- This migration is idempotent.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Ensure all required permissions exist.
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

-- Ensure default roles exist for each company.
INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT gen_random_uuid(), c.id, 'Owner', 'Akses penuh ke semua modul', now(), now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT gen_random_uuid(), c.id, 'HR', 'Mengelola karyawan, cuti, dan draft payroll', now(), now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT gen_random_uuid(), c.id, 'Finance', 'Mengelola gaji, persetujuan payroll, dan pembayaran', now(), now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT gen_random_uuid(), c.id, 'Employee', 'Hanya akses baca profil, slip gaji, dan pengajuan cuti', now(), now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

-- Owner: all permissions.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'Owner'
ON CONFLICT DO NOTHING;

-- HR: employee/department/position + attendance + leave(read/approve/manage) + payroll(read/create/delete)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    p.resource IN ('employee', 'department', 'position', 'attendance')
) OR (
    p.resource = 'leave' AND p.action IN ('read', 'approve', 'manage')
) OR (
    p.resource = 'payroll' AND p.action IN ('read', 'create', 'delete')
)
WHERE r.name = 'HR'
ON CONFLICT DO NOTHING;

-- Finance: salary + payroll(read/approve/pay)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    p.resource = 'salary'
) OR (
    p.resource = 'payroll' AND p.action IN ('read', 'approve', 'pay')
)
WHERE r.name = 'Finance'
ON CONFLICT DO NOTHING;

-- Employee: employee(read) + payroll(read) + leave(read/create)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON (
    p.resource = 'employee' AND p.action = 'read'
) OR (
    p.resource = 'payroll' AND p.action = 'read'
) OR (
    p.resource = 'leave' AND p.action IN ('read', 'create')
)
WHERE r.name = 'Employee'
ON CONFLICT DO NOTHING;

-- Remove SUPERADMIN assignment from demo users so role-based flow can be tested clearly.
DELETE FROM employee_roles er
USING roles r, employees e, companies c
WHERE er.role_id = r.id
  AND er.employee_id = e.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email IN ('owner@demo.com', 'hr@demo.com', 'employee@demo.com')
  AND r.name = 'SUPERADMIN';

-- Assign demo users to intended default roles.
INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT e.id, r.id, now()
FROM employees e
JOIN companies c ON c.id = e.company_id
JOIN roles r ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'owner@demo.com'
  AND r.name = 'Owner'
ON CONFLICT DO NOTHING;

INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT e.id, r.id, now()
FROM employees e
JOIN companies c ON c.id = e.company_id
JOIN roles r ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'hr@demo.com'
  AND r.name = 'HR'
ON CONFLICT DO NOTHING;

INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT e.id, r.id, now()
FROM employees e
JOIN companies c ON c.id = e.company_id
JOIN roles r ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'employee@demo.com'
  AND r.name = 'Employee'
ON CONFLICT DO NOTHING;
