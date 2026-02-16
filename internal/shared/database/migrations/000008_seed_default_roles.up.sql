-- =========================================
-- SEED DEFAULT ROLES PER COMPANY
-- Uses gen_random_uuid() (Postgres 18)
-- Idempotent
-- =========================================

-- 1️⃣ Insert Roles Per Company
INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT
    gen_random_uuid(),
    c.id,
    r.name,
    r.description,
    now(),
    now()
FROM companies c
CROSS JOIN (
    VALUES
        ('Owner', 'Full access to all resources'),
        ('HR', 'Manage employees and salary'),
        ('Employee', 'Read own profile only')
) AS r(name, description)
ON CONFLICT (company_id, name) DO NOTHING;



-- 2️⃣ OWNER → ALL PERMISSIONS
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now()
FROM roles ro
JOIN permissions p ON true
WHERE ro.name = 'Owner'
ON CONFLICT DO NOTHING;



-- 3️⃣ HR → employee + salary
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now()
FROM roles ro
JOIN permissions p
    ON p.resource IN ('employee', 'salary')
WHERE ro.name = 'HR'
ON CONFLICT DO NOTHING;



-- 4️⃣ Employee → employee.read only
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now()
FROM roles ro
JOIN permissions p
    ON p.resource = 'employee'
   AND p.action = 'read'
WHERE ro.name = 'Employee'
ON CONFLICT DO NOTHING;
